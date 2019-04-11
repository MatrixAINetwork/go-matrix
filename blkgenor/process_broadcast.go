// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"github.com/MatrixAINetwork/go-matrix/consensus/blkmanage"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

func (p *Process) AddBroadcastMinerResult(result *mc.HD_BroadcastMiningRspMsg) {
	if result == nil || result.BlockMainData == nil || result.BlockMainData.Header == nil {
		log.Warn(p.logExtraInfo(), "广播区块挖矿结果", "消息为nil")
		return
	}
	if result.From != result.BlockMainData.Header.Leader {
		log.Info(p.logExtraInfo(), "广播区块挖矿结果", "消息from != 消息leader", "leader", result.BlockMainData.Header.Leader.Hex(), "from", result.From.Hex())
		return
	}
	if p.preVerifyBroadcastMinerResult(result.BlockMainData) == false {
		log.WARN(p.logExtraInfo(), "广播区块挖矿结果", "预验证失败, 抛弃该消息")
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 缓存广播区块挖矿结果
	log.INFO(p.logExtraInfo(), "缓存广播区块挖矿结果成功，高度", p.number)
	p.broadcastRstCache[result.From] = result.BlockMainData
	p.processMinerResultVerify(p.curLeader, true)
}

func (p *Process) preVerifyBroadcastMinerResult(result *mc.BlockData) bool {
	bcInterval := p.bcInterval
	if bcInterval == nil {
		var err error
		bcInterval, err = p.blockChain().GetBroadcastInterval()
		if err != nil {
			log.Error(p.logExtraInfo(), "预验证广播挖矿结果", "获取当前广播周期失败", "err", err)
			return false
		}
	}
	if false == bcInterval.IsBroadcastNumber(result.Header.Number.Uint64()) {
		log.WARN(p.logExtraInfo(), "预验证广播挖矿结果", "高度不是广播区块高度", "高度", result.Header.Number.Uint64())
		return false
	}
	return true
}

func (p *Process) dealMinerResultVerifyBroadcast() {
	log.INFO(p.logExtraInfo(), "当前高度为广播区块, 进行广播挖矿结果验证, 高度", p.number)
	for _, result := range p.broadcastRstCache {
		if err := p.blockChain().DPOSEngine(result.Header.Version).VerifyBlock(p.blockChain(), result.Header); err != nil {
			log.WARN(p.logExtraInfo(), "广播挖矿结果处理", "结果异常", "err", err)
			continue
		}

		state, retTxs, receipts, _, err := p.pm.manblk.VerifyTxsAndState(blkmanage.BroadcastBlk, string(result.Header.Version), result.Header, result.Txs)
		if nil != err {
			log.WARN(p.logExtraInfo(), "广播挖矿结果处理", "状态异常", "err", err)
			continue
		}
		p.blockCache.SaveReadyBlock(&mc.BlockLocalVerifyOK{
			Header:      result.Header,
			BlockHash:   result.Header.HashNoSignsAndNonce(),
			OriginalTxs: result.Txs,
			FinalTxs:    retTxs,
			Receipts:    receipts,
			State:       state,
		})

		readyMsg := &mc.NewBlockReadyMsg{
			Header: result.Header,
			State:  state.Copy(),
		}
		log.INFO(p.logExtraInfo(), "广播区块验证完成", "发送新区块准备完毕消息", "高度", p.number, "leader", result.Header.Leader.Hex())
		mc.PublishEvent(mc.BlockGenor_NewBlockReady, readyMsg)

		p.changeState(StateBlockInsert)
		p.processBlockInsert(result.Header.Leader)
		return
	}
}
