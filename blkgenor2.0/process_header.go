// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenorV2

import (
	"github.com/MatrixAINetwork/go-matrix/consensus/blkmanage"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/txpoolCache"
	"github.com/pkg/errors"
)

func (p *Process) processHeaderGen(AIResult *mc.HD_V2_AIMiningRspMsg) error {
	log.Info(p.logExtraInfo(), "processHeaderGen", "start")
	defer log.Info(p.logExtraInfo(), "processHeaderGen", "end")
	if p.bcInterval == nil {
		log.Error(p.logExtraInfo(), "区块生成阶段", "广播周期信息为空")
		return errors.New("广播周期信息为空")
	}
	if p.parentHeader == nil {
		log.Error(p.logExtraInfo(), "区块生成阶段", "父区块为nil")
		return errors.New("父区块为nil")
	}
	version, err := p.pm.manblk.ProduceBlockVersion(p.number, string(p.parentHeader.Version))
	if err != nil {
		return err
	}

	originHeader, extraData, err := p.pm.manblk.Prepare(blkmanage.CommonBlk, version, p.number, p.bcInterval, p.parentHash, AIResult)
	if err != nil {
		log.Error(p.logExtraInfo(), "准备阶段失败", err)
		return err
	}

	onlineConsensusResults, ok := extraData.([]*mc.HD_OnlineConsensusVoteResultMsg)
	if !ok {
		log.Error(p.logExtraInfo(), "反射在线状态失败", "")
		return errors.New("反射在线状态失败")
	}

	txsCode, stateDB, receipts, originalTxs, finalTxs, _, err := p.pm.manblk.ProcessState(blkmanage.CommonBlk, version, originHeader, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "运行交易和状态树失败", err)
		return err
	}
	//运行完matrix状态树后，生成root (p.blockChain(), header, stateDB, nil, tsBlock.Currencies())
	block, _, err := p.pm.manblk.Finalize(blkmanage.CommonBlk, version, originHeader, stateDB, finalTxs, nil, receipts, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "Finalize失败", err)
		return err
	}

	reqMsg := &mc.HD_BlkConsensusReqMsg{
		Header:                 block.Header(),
		TxsCode:                txsCode,
		ConsensusTurn:          p.consensusTurn,
		OnlineConsensusResults: onlineConsensusResults,
		From: ca.GetSignAddress(),
	}
	//send to local block verify module
	if len(originalTxs) > 0 {
		txpoolCache.MakeStruck(types.GetTX(originalTxs), reqMsg.Header.HashNoSignsAndNonce(), p.number)
	}
	p.sendHeaderVerifyReq(&mc.LocalBlockVerifyConsensusReq{BlkVerifyConsensusReq: reqMsg, OriginalTxs: originalTxs, FinalTxs: finalTxs, Receipts: receipts, State: stateDB})
	return nil
}

func (p *Process) sendHeaderVerifyReq(req *mc.LocalBlockVerifyConsensusReq) {
	log.Info(p.logExtraInfo(), "本地发送区块验证请求", req.BlkVerifyConsensusReq.Header.HashNoSignsAndNonce().TerminalString(), "高度", p.number)
	mc.PublishEvent(mc.BlockGenor_HeaderVerifyReq, req)
	p.startConsensusReqSender(req.BlkVerifyConsensusReq)
}

func (p *Process) startConsensusReqSender(req *mc.HD_BlkConsensusReqMsg) {
	p.closeMsgSender()
	sender, err := common.NewResendMsgCtrl(req, p.sendConsensusReqFunc, manparams.BlkPosReqSendInterval, manparams.BlkPosReqSendTimes)
	if err != nil {
		log.Error(p.logExtraInfo(), "创建req发送器", "失败", "err", err)
		return
	}
	p.msgSender = sender
}

func (p *Process) sendConsensusReqFunc(data interface{}, times uint32) {
	req, OK := data.(*mc.HD_BlkConsensusReqMsg)
	if !OK {
		log.Error(p.logExtraInfo(), "发出区块共识req", "反射消息失败", "次数", times)
		return
	}
	log.Info(p.logExtraInfo(), "!!!!网络发送区块验证请求, hash", req.Header.HashNoSignsAndNonce(), "tx数量", req.TxsCodeCount(), "次数", times)
	p.pm.hd.SendNodeMsg(mc.HD_BlkConsensusReq, req, common.RoleValidator, nil)
}

func (p *Process) processBroadcastBlockGen() error {
	log.Info(p.logExtraInfo(), "processBroadcastBlockGen", "start")
	defer log.Info(p.logExtraInfo(), "processBroadcastBlockGen", "end")
	if p.bcInterval == nil {
		log.Error(p.logExtraInfo(), "广播区块生成阶段", "广播周期信息为空")
		return errors.New("广播周期信息为空")
	}
	if p.parentHeader == nil {
		log.Error(p.logExtraInfo(), "广播区块生成阶段", "父区块为nil")
		return errors.New("父区块为nil")
	}
	version, err := p.pm.manblk.ProduceBlockVersion(p.number, string(p.parentHeader.Version))
	if err != nil {
		return err
	}

	originHeader, _, err := p.pm.manblk.Prepare(blkmanage.BroadcastBlk, version, p.number, p.bcInterval, p.parentHash)
	if err != nil {
		log.Error(p.logExtraInfo(), "广播区块生成阶段", "准备区块失败", err)
		return err
	}

	_, stateDB, receipts, _, finalTxs, _, err := p.pm.manblk.ProcessState(blkmanage.BroadcastBlk, version, originHeader, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "广播区块生成阶段, 运行交易和状态树失败", err)
		return err
	}

	//运行完matrix状态树后，生成root
	block, _, err := p.pm.manblk.Finalize(blkmanage.BroadcastBlk, version, originHeader, stateDB, finalTxs, nil, receipts, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "Finalize失败", err)
		return err
	}
	finalHeader := block.Header()
	err = p.setSignatures(finalHeader)
	if err != nil {
		return err
	}

	p.sendBroadcastRspMsg(&mc.BlockData{Header: finalHeader, Txs: finalTxs})
	return nil
}

func (p *Process) setSignatures(header *types.Header) error {
	signHash := header.HashNoSignsAndNonce()
	sign, err := p.signHelper().SignHashWithValidateByAccount(signHash.Bytes(), true, ca.GetDepositAddress())
	if err != nil {
		log.Error(p.logExtraInfo(), "广播区块生成，签名错误", err)
		return err
	}
	//log.Debug(p.logExtraInfo(), "test log", "广播区块签名成功", "sign hash", signHash.TerminalString(), "sign account", ca.GetDepositAddress().Hex(), "version", string(header.Version))
	header.Signatures = make([]common.Signature, 0, 1)
	header.Signatures = append(header.Signatures, sign)
	return nil
}

func (p *Process) sendBroadcastRspMsg(bcBlock *mc.BlockData) {
	log.Info(p.logExtraInfo(), "发送广播区块结果", bcBlock.Header.HashNoSigns().TerminalString(), "高度", p.number)
	p.startBroadcastRspSender(bcBlock)
}

func (p *Process) startBroadcastRspSender(bcBlock *mc.BlockData) {
	p.closeMsgSender()
	sender, err := common.NewResendMsgCtrl(bcBlock, p.sendBroadcastRspFunc, manparams.BlkPosReqSendInterval, manparams.BlkPosReqSendTimes)
	if err != nil {
		log.Error(p.logExtraInfo(), "创建广播区块结果发送器", "失败", "err", err)
		return
	}
	p.msgSender = sender
}

func (p *Process) sendBroadcastRspFunc(data interface{}, times uint32) {
	bcBlock, OK := data.(*mc.BlockData)
	if !OK {
		log.Error(p.logExtraInfo(), "发出广播区块结果", "反射消息失败", "次数", times)
		return
	}

	msg := &mc.HD_BroadcastMiningRspMsg{
		BlockMainData: bcBlock,
	}
	log.Trace(p.logExtraInfo(), "!!网络发送广播区块结果, hash", msg.BlockMainData.Header.HashNoSignsAndNonce(), "交易数量", len(types.GetTX(msg.BlockMainData.Txs)), "次数", times, "高度", msg.BlockMainData.Header.Number)
	p.pm.hd.SendNodeMsg(mc.HD_BroadcastMiningRsp, msg, common.RoleValidator, nil)
}
