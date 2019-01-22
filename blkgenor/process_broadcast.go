// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/matrixwork"
	"github.com/matrix/go-matrix/mc"
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
		log.WARN(p.logExtraInfo(), "广播区块挖矿结果", "预验证事变, 抛弃该消息")
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
	if p.bcInterval == nil {
		log.WARN(p.logExtraInfo(), "验证广播挖矿结果", "广播周期信息为nil")
		return false
	}
	if false == p.bcInterval.IsBroadcastNumber(result.Header.Number.Uint64()) {
		log.WARN(p.logExtraInfo(), "验证广播挖矿结果", "高度不是广播区块高度", "高度", result.Header.Number.Uint64())
		return false
	}
	if err := p.blockChain().DPOSEngine(result.Header.Version).VerifyBlock(p.blockChain(), result.Header); err != nil {
		log.WARN(p.logExtraInfo(), "验证广播挖矿结果", "结果异常", "err", err)
		return false
	}
	return true
}

func (p *Process) dealMinerResultVerifyBroadcast() {
	log.INFO(p.logExtraInfo(), "当前高度为广播区块, 进行广播挖矿结果验证, 高度", p.number)
	for _, result := range p.broadcastRstCache {
		state, retTxs, receipsts, _, err := p.pm.manblk.VerifyTxsAndState(blkmanage.BroadcastBlk, string(result.Header.Version), result.Header, result.Txs)
		if nil != err {
			continue
		}
		p.blockCache.SaveReadyBlock(&mc.BlockLocalVerifyOK{
			Header:      result.Header,
			BlockHash:   common.Hash{},
			OriginalTxs: retTxs,
			FinalTxs:    retTxs,
			Receipts:    receipsts,
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
