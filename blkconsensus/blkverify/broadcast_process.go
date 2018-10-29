// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
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
		if err := p.blockChain().DPOSEngine().VerifyBlock(p.blockChain(), req.req.Header); err != nil {
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
		result := mc.BlockLocalVerifyOK{
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
