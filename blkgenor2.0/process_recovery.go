// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenorV2

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus/blkmanage"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"time"
)

func (p *Process) ProcessRecoveryMsg(msg *mc.RecoveryStateMsg) {
	log.Info(p.logExtraInfo(), "状态恢复消息处理", "开始", "类型", msg.Type, "高度", p.number, "leader", msg.Header.Leader.Hex())
	defer log.Debug(p.logExtraInfo(), "状态恢复消息处理", "结束", "类型", msg.Type, "高度", p.number, "leader", msg.Header.Leader.Hex())

	p.mu.Lock()
	defer p.mu.Unlock()
	header := msg.Header
	headerHash := header.HashNoSignsAndNonce()

	if blockData := p.blockPool.GetBlockDataByBlockHash(headerHash); blockData != nil && blockData.state == blockStateComplete {
		log.Debug("状态恢复消息处理", "已存在完整区块信息", "抛弃恢复消息")
		return
	}
	p.sendFullBlockReq(headerHash, header.Number.Uint64(), msg.From)
}

func (p *Process) sendFullBlockReq(hash common.Hash, number uint64, target common.Address) {
	if p.FullBlockReqCache.IsExistMsg(hash) {
		data, err := p.FullBlockReqCache.ReUseMsg(hash)
		if err != nil {
			return
		}
		reqMsg, _ := data.(*mc.HD_V2_FullBlockReqMsg)
		log.Debug(p.logExtraInfo(), "状态恢复消息处理", "发送完整区块获取请求消息", "to", target.Hex(), "高度", reqMsg.Number, "hash", reqMsg.HeaderHash.TerminalString())
		p.pm.hd.SendNodeMsg(mc.HD_V2_FullBlockReq, reqMsg, common.RoleNil, []common.Address{target})
	} else {
		reqMsg := &mc.HD_V2_FullBlockReqMsg{
			HeaderHash: hash,
			Number:     number,
		}
		p.FullBlockReqCache.AddMsg(hash, reqMsg, time.Now().Unix())
		log.Debug(p.logExtraInfo(), "状态恢复消息处理", "发送完整区块获取请求消息", "to", target.Hex(), "高度", reqMsg.Number, "hash", reqMsg.HeaderHash.TerminalString())
		p.pm.hd.SendNodeMsg(mc.HD_V2_FullBlockReq, reqMsg, common.RoleNil, []common.Address{target})
	}
}

func (p *Process) ProcessFullBlockReq(req *mc.HD_V2_FullBlockReqMsg) {
	p.mu.Lock()
	defer p.mu.Unlock()

	blockData := p.blockPool.GetBlockDataByBlockHash(req.HeaderHash)
	if blockData == nil {
		log.Error(p.logExtraInfo(), "处理完整区块请求", "区块信息未找到", "高度", p.number, "hash", req.HeaderHash.TerminalString())
		return
	}
	if blockData.state != blockStateComplete {
		log.Error(p.logExtraInfo(), "处理完整区块请求", "区块为非完整区块", "高度", p.number, "hash", req.HeaderHash.TerminalString())
		return
	}

	rspMsg := &mc.HD_V2_FullBlockRspMsg{
		Header: blockData.block.Header,
		Txs:    blockData.block.OriginalTxs,
	}
	log.Debug(p.logExtraInfo(), "处理完整区块请求", "发送响应消息", "to", req.From, "hash", rspMsg.Header.Hash(), "交易", rspMsg.Txs)
	p.pm.hd.SendNodeMsg(mc.HD_V2_FullBlockRsp, rspMsg, common.RoleNil, []common.Address{req.From})
}

func (p *Process) ProcessFullBlockRsp(rsp *mc.HD_V2_FullBlockRspMsg) {
	fullHash := rsp.Header.Hash()
	headerHash := rsp.Header.HashNoSignsAndNonce()
	log.Info(p.logExtraInfo(), "处理完整区块响应", "开始", "区块 hash", fullHash.TerminalString(), "交易", rsp.Txs, "root", rsp.Header.Roots, "高度", p.number)
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.bcInterval == nil {
		log.Error(p.logExtraInfo(), "处理完整区块响应", "广播周期为nil", "header Hash", headerHash.TerminalString(), "高度", p.number)
		return
	}
	if blockData := p.blockPool.GetBlockDataByBlockHash(headerHash); blockData != nil && blockData.state == blockStateComplete {
		log.Debug(p.logExtraInfo(), "处理完整区块响应", "已存在完整区块信息, 抛弃恢复消息")
		return
	}

	isBroadcast := p.bcInterval.IsBroadcastNumber(rsp.Header.Number.Uint64())
	if err := p.pm.bc.Engine(rsp.Header.Version).VerifyHeader(p.pm.bc, rsp.Header, !isBroadcast, false); err != nil {
		log.Error(p.logExtraInfo(), "处理完整区块响应", "POW验证未通过", "err", err, "高度", p.number)
		return
	}

	if err := p.pm.bc.DPOSEngine(rsp.Header.Version).VerifyBlock(p.pm.bc, rsp.Header); err != nil {
		log.Error(p.logExtraInfo(), "处理完整区块响应", "POS验证未通过", "err", err, "高度", p.number)
		return
	}

	// 没有本地运行的区块结果，运行交易
	blkType := blkmanage.CommonBlk
	if isBroadcast {
		blkType = blkmanage.BroadcastBlk
	}
	//运行交易
	stateDB, finalTxs, receipts, _, err := p.pm.manblk.VerifyTxsAndState(blkType, string(rsp.Header.Version), rsp.Header, rsp.Txs)
	if err != nil {
		log.Error(p.logExtraInfo(), "处理完整区块响应", "执行交易错误", "err", err, "高度", p.number)
		return
	}

	p.blockPool.SaveCompleteBlock(&mc.BlockPOSFinishedV2{
		Header:      rsp.Header,
		BlockHash:   rsp.Header.HashNoSignsAndNonce(),
		OriginalTxs: rsp.Txs,
		FinalTxs:    finalTxs,
		Receipts:    receipts,
		State:       stateDB,
	})

	readyMsg := &mc.NewBlockReadyMsg{
		Header: rsp.Header,
		State:  stateDB.Copy(),
	}
	mc.PublishEvent(mc.BlockGenor_NewBlockReady, readyMsg)

	p.state = StateBlockInsert
	p.startBlockInsert(rsp.Header.Leader)
}
