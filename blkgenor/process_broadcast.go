// Copyright (c) 2018 The MATRIX Authors
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
	if err := p.dposEngine().VerifyBlock(p.blockChain(), result.Header); err != nil {
		log.WARN(p.logExtraInfo(), "验证广播挖矿结果", "结果异常", "err", err)
		return false
	}
	return true
}

func (p *Process) dealMinerResultVerifyBroadcast() {
	log.INFO(p.logExtraInfo(), "当前高度为广播区块, 进行广播挖矿结果验证, 高度", p.number)
	for _, result := range p.broadcastRstCache {
		// 运行广播区块交易
		parent := p.blockChain().GetBlockByHash(result.Header.ParentHash)
		if parent == nil {
			log.WARN(p.logExtraInfo(), "广播挖矿结果验证", "获取父区块错误!")
			continue
		}

		work, err := matrixwork.NewWork(p.blockChain().Config(), p.blockChain(), nil, result.Header, p.pm.random)
		if err != nil {
			log.WARN(p.logExtraInfo(), "广播挖矿结果验证, 创建worker错误", err)
			continue
		}

		// 运行版本更新检查
		if err := p.blockChain().ProcessStateVersion(result.Header.Version, work.State); err != nil {
			log.ERROR(p.logExtraInfo(), "广播挖矿结果验证, 版本更新检查失败", err)
			continue
		}
		//执行交易
		work.ProcessBroadcastTransactions(p.pm.matrix.EventMux(), result.Txs)
		retTxs := work.GetTxs()
		// 运行matrix状态树
		block := types.NewBlock(result.Header, retTxs, nil, work.Receipts)
		if err := p.blockChain().ProcessMatrixState(block, work.State); err != nil {
			log.ERROR(p.logExtraInfo(), "广播挖矿结果验证, matrix 状态树运行错误", err)
			continue
		}

		localBlock, err := p.blockChain().Engine().Finalize(p.blockChain(), block.Header(), work.State, retTxs, nil, work.Receipts)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "Failed to finalize block for sealing", err)
			continue
		}

		if localBlock.Root() != result.Header.Root {
			log.ERROR(p.logExtraInfo(), "广播挖矿结果验证", "root验证错误, 不匹配", "localRoot", localBlock.Root().TerminalString(), "remote root", result.Header.Root.TerminalString())
			continue
		}

		p.blockCache.SaveReadyBlock(&mc.BlockLocalVerifyOK{
			Header:      result.Header,
			BlockHash:   common.Hash{},
			OriginalTxs: retTxs,
			FinalTxs:    retTxs,
			Receipts:    work.Receipts,
			State:       work.State,
		})

		readyMsg := &mc.NewBlockReadyMsg{
			Header: result.Header,
			State:  work.State.Copy(),
		}
		log.INFO(p.logExtraInfo(), "广播区块验证完成", "发送新区块准备完毕消息", "高度", p.number, "leader", result.Header.Leader.Hex())
		mc.PublishEvent(mc.BlockGenor_NewBlockReady, readyMsg)

		p.changeState(StateBlockInsert)
		p.processBlockInsert(result.Header.Leader)
		return
	}
}
