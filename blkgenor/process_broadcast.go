// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/matrixwork"
	"github.com/matrix/go-matrix/mc"
)

func (p *Process) AddBroadcastMinerResult(result *mc.HD_BroadcastMiningRspMsg) {
	if p.preVerifyBroadcastMinerResult(result.BlockMainData) == false {
		log.WARN(p.logExtraInfo(), "预验证广播区块挖矿结果错误", "抛弃该消息")
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 缓存广播区块挖矿结果
	log.WARN(p.logExtraInfo(), "缓存广播区块挖矿结果成功，高度", p.number)
	p.broadcastRstCache = append(p.broadcastRstCache, result.BlockMainData)

	p.processMinerResultVerify(p.curLeader, true)
}

func (p *Process) preVerifyBroadcastMinerResult(result *mc.BlockData) bool {
	if false == common.IsBroadcastNumber(result.Header.Number.Uint64()) {
		log.ERROR(p.logExtraInfo(), "验证广播挖矿结果", "高度不是广播区块高度", "高度", result.Header.Number.Uint64())
		return false
	}
	if err := p.dposEngine().VerifyBlock(p.blockChain(), result.Header); err != nil {
		log.ERROR(p.logExtraInfo(), "验证广播挖矿结果", "结果异常", "err", err)
		return false
	}
	return true
}

func (p *Process) dealMinerResultVerifyBroadcast() {
	for _, result := range p.broadcastRstCache {
		// 运行广播区块交易
		parent := p.blockChain().GetBlockByHash(result.Header.ParentHash)
		if parent == nil {
			log.ERROR(p.logExtraInfo(), "广播挖矿结果验证", "获取父区块错误!")
			continue
		}

		work, err := matrixwork.NewWork(p.blockChain().Config(), p.blockChain(), nil, result.Header)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "广播挖矿结果验证, 创建worker错误", err)
			continue
		}
		//执行交易
		work.ProcessBroadcastTransactions(p.pm.matrix.EventMux(), result.Txs, p.pm.bc)
		retTxs := work.GetTxs()
		log.INFO("*********************", "len(result.Txs)", len(retTxs))
		for _, tx := range retTxs {
			log.INFO("==========", "Finalize:GasPrice", tx.GasPrice(), "amount", tx.Value())
		}
		_, err = p.blockChain().Engine().Finalize(p.blockChain(), result.Header, work.State, retTxs, nil, work.Receipts)

		if err != nil {
			log.ERROR(p.logExtraInfo(), "Failed to finalize block for sealing", err)
			continue
		}

		p.blockCache.SaveReadyBlock(&mc.BlockLocalVerifyOK{
			Header:    result.Header,
			BlockHash: common.Hash{},
			Txs:       retTxs,
			Receipts:  work.Receipts,
			State:     work.State,
		})

		readyMsg := &mc.NewBlockReadyMsg{
			Header: result.Header,
		}
		log.INFO(p.logExtraInfo(), "广播区块验证完成", "发送新区块准备完毕消息", "高度", p.number)
		mc.PublishEvent(mc.BlockGenor_NewBlockReady, readyMsg)

		p.changeState(StateBlockInsert)
		p.processBlockInsert(p.curLeader)
		return
	}
}
