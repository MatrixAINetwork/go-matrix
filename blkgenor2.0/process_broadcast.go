// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenorV2

import (
	"github.com/MatrixAINetwork/go-matrix/consensus/blkmanage"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

type bcBlockRspInfo struct {
	bcBlock       *mc.BlockData
	localVerified bool
}

func (p *Process) AddBroadcastBlockResult(result *mc.HD_BroadcastMiningRspMsg) {
	if result == nil || result.BlockMainData == nil || result.BlockMainData.Header == nil {
		log.Warn(p.logExtraInfo(), "广播区块挖矿结果", "消息为nil")
		return
	}
	if result.From != result.BlockMainData.Header.Leader {
		log.Info(p.logExtraInfo(), "广播区块挖矿结果", "消息from != 消息leader", "leader", result.BlockMainData.Header.Leader.Hex(), "from", result.From.Hex())
		return
	}
	if p.preVerifyBroadcastBlock(result.BlockMainData) == false {
		log.Warn(p.logExtraInfo(), "广播区块挖矿结果", "预验证失败, 抛弃该消息")
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 缓存广播区块挖矿结果
	log.Info(p.logExtraInfo(), "缓存广播区块挖矿结果成功，高度", p.number)
	p.broadcastRstCache[result.BlockMainData.Header.Leader] = &bcBlockRspInfo{result.BlockMainData, false}
	p.processBCBlockVerify()
}

func (p *Process) preVerifyBroadcastBlock(result *mc.BlockData) bool {
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
		log.Warn(p.logExtraInfo(), "预验证广播挖矿结果", "高度不是广播区块高度", "高度", result.Header.Number.Uint64())
		return false
	}
	return true
}

func (p *Process) processBCBlockVerify() {
	if p.checkState(StatePOSWaiting) == false {
		log.Warn(p.logExtraInfo(), "广播区块结果处理", "状态错误", "当前状态", p.state.String())
		return
	}

	log.Info(p.logExtraInfo(), "当前高度为广播区块, 进行广播区块验证, 高度", p.number)
	for _, result := range p.broadcastRstCache {
		bcBlock := result.bcBlock
		if err := p.blockChain().DPOSEngine(bcBlock.Header.Version).VerifyBlock(p.blockChain(), bcBlock.Header); err != nil {
			log.Warn(p.logExtraInfo(), "广播区块结果处理", "结果异常", "err", err)
			continue
		}

		state, retTxs, receipts, _, err := p.pm.manblk.VerifyTxsAndState(blkmanage.BroadcastBlk, string(bcBlock.Header.Version), bcBlock.Header, bcBlock.Txs)
		if nil != err {
			log.Warn(p.logExtraInfo(), "广播区块结果处理", "状态异常", "err", err)
			continue
		}
		p.blockPool.SaveCompleteBlock(&mc.BlockPOSFinishedV2{
			Header:      bcBlock.Header,
			BlockHash:   bcBlock.Header.HashNoSignsAndNonce(),
			OriginalTxs: bcBlock.Txs,
			FinalTxs:    retTxs,
			Receipts:    receipts,
			State:       state,
		})

		readyMsg := &mc.NewBlockReadyMsg{
			Header: bcBlock.Header,
			State:  state.Copy(),
		}
		log.Info(p.logExtraInfo(), "广播区块验证完成", "发送新区块准备完毕消息", "高度", p.number, "leader", bcBlock.Header.Leader.Hex())
		mc.PublishEvent(mc.BlockGenor_NewBlockReady, readyMsg)

		p.state = StateBlockInsert
		p.startBlockInsert(bcBlock.Header.Leader)
		return
	}
}
