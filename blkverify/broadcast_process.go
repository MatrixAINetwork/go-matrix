// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

func (p *Process) startReqVerifyBC() {
	if p.checkState(StateStart) == false {
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
		if err := p.blockChain().DPOSEngine(req.req.Header.Version).VerifyBlock(p.blockChain(), req.req.Header); err != nil {
			log.WARN(p.logExtraInfo(), "广播身份，启动阶段, DPOS共识失败", err, "req leader", req.req.Header.Leader.Hex(), "高度", p.number)
			req.localVerifyResult = localVerifyResultStateFailed
			continue
		}

		p.curProcessReq = req
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

func (p *Process) bcFinishedProcess(lvResult verifyResult) {
	p.curProcessReq.localVerifyResult = lvResult
	if lvResult == localVerifyResultProcessing {
		log.ERROR(p.logExtraInfo(), "req is processing now, process can't finish!", "broadcast role")
		return
	}

	if lvResult != localVerifyResultSuccess {
		log.Error(p.logExtraInfo(), "广播节点验证请求失败", lvResult.String(), "高度", p.number, "req hash", p.curProcessReq.hash.Hex(), "req from", p.curProcessReq.req.From.Hex())
		log.Info(p.logExtraInfo(), "广播节点", "重启process流程")
		p.curProcessReq = nil
		p.state = StateStart
		p.startReqVerifyBC()
		return
	}

	// notify block genor server the result
	result := mc.BlockLocalVerifyOK{
		Header:      p.curProcessReq.req.Header,
		BlockHash:   p.curProcessReq.hash,
		OriginalTxs: p.curProcessReq.originalTxs,
		FinalTxs:    p.curProcessReq.finalTxs,
		Receipts:    p.curProcessReq.receipts,
		State:       p.curProcessReq.stateDB,
	}
	log.INFO(p.logExtraInfo(), "广播身份", "请求验证完成, 发出区块共识结果消息", "高度", p.number, "block hash", result.BlockHash.TerminalString())
	mc.PublishEvent(mc.BlkVerify_VerifyConsensusOK, &result)

	p.state = StateEnd
}
