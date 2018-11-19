// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

func (p *Process) startReqVerifyBC() {
	if p.checkState(StateStart) == false && p.checkState(StateEnd) == false {
		log.WARN(p.logExtraInfo(), "广播身份，开启启动阶段，状态错误", p.state.String(), "高度", p.number)
		return
	}

	reqList := p.reqCache.GetAllReq()
	log.INFO(p.logExtraInfo(), "广播身份，启动阶段，请求总数", len(reqList), "高度", p.number)
	for _, req := range reqList {
		if req.localVerifyResult != localVerifyResultProcessing {
			continue
		}
		//verify dpos
		if err := p.blockChain().DPOSEngine().VerifyBlock(req.req.Header); err != nil {
			log.WARN(p.logExtraInfo(), "广播身份，启动阶段, DPOS共识失败", err, "req leader", req.req.Header.Leader.Hex(), "高度", p.number)
			req.localVerifyResult = localVerifyResultStateFailed
			continue
		}

		p.curProcessReq = req
		p.curProcessReq.hash = p.curProcessReq.req.Header.HashNoSignsAndNonce()
		p.state = StateReqVerify
		p.bcProcessReqVerify()
		return
	}

	log.INFO(p.logExtraInfo(), "广播身份，启动阶段，未找到合适的请求", "继续等待请求", "高度", p.number)
}

func (p *Process) bcProcessReqVerify() {
	if p.checkState(StateReqVerify) == false {
		log.WARN(p.logExtraInfo(), "广播身份，准备开始请求验证阶段，状态错误", p.state.String(), "高度", p.number)
		return
	}

	// verify header
	if err := p.blockChain().VerifyHeader(p.curProcessReq.req.Header); err != nil {
		log.ERROR(p.logExtraInfo(), "广播身份，请求验证阶段, 通过DPOS共识的请求，但是预验证头信息错误", err)
		p.bcFinishedProcess(localVerifyResultStateFailed)
		return
	}

	p.startTxsVerify()
}

func (p *Process) bcFinishedProcess(lvResult uint8) {
	p.curProcessReq.localVerifyResult = lvResult
	if lvResult == localVerifyResultProcessing {
		log.ERROR(p.logExtraInfo(), "req is processing now, process can't finish!", "broadcast role")
		return
	}
	if lvResult == localVerifyResultStateFailed {
		log.ERROR(p.logExtraInfo(), "local verify header err, but dpos pass! please check your state!", "broadcast role")
		//todo 硬分叉了，以后加需要处理
		return
	}

	if lvResult == localVerifyResultSuccess {
		// notify block genor server the result
		result := mc.BlockVerifyConsensusOK{
			Header:    p.curProcessReq.req.Header,
			BlockHash: p.curProcessReq.hash,
			Txs:       p.curProcessReq.txs,
			Receipts:  p.curProcessReq.receipts,
			State:     p.curProcessReq.stateDB,
		}
		log.INFO(p.logExtraInfo(), "广播身份", "请求验证完成, 发出区块共识结果消息", "高度", p.number, "block hash", result.BlockHash.TerminalString())
		mc.PublishEvent(mc.BlkVerify_VerifyConsensusOK, &result)
	}

	p.state = StateEnd
}
