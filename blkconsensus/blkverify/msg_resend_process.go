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
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/man"
)

func (p *Process) stopSender() {
	p.closeMineReqMsgSender()
	p.closePosedReqSender()
	p.closeVoteMsgSender()
}

func (p *Process) startSendMineReq(req *mc.HD_MiningReqMsg) {
	p.closeMineReqMsgSender()
	sender, err := common.NewResendMsgCtrl(req, p.sendMineReqFunc, man.MinerReqSendInterval, 0)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "创建挖矿请求发送器", "失败", "err", err)
		return
	}
	p.mineReqMsgSender = sender
}

func (p *Process) closeMineReqMsgSender() {
	if p.mineReqMsgSender == nil {
		return
	}
	p.mineReqMsgSender.Close()
	p.mineReqMsgSender = nil
}

func (p *Process) sendMineReqFunc(data interface{}, times uint32) {
	req, OK := data.(*mc.HD_MiningReqMsg)
	if !OK {
		log.ERROR(p.logExtraInfo(), "发出挖矿请求", "反射消息失败")
		return
	}
	hash := req.Header.HashNoSignsAndNonce()
	//给矿工发送区块验证结果
	log.INFO(p.logExtraInfo(), "发出挖矿请求, Header hash with signs", hash, "次数", times, "高度", p.number)
	p.pm.hd.SendNodeMsg(mc.HD_MiningReq, req, common.RoleMiner, nil)
}

func (p *Process) startPosedReqSender(req *mc.HD_BlkConsensusReqMsg) {
	p.closePosedReqSender()
	sender, err := common.NewResendMsgCtrl(req, p.sendPosedReqFunc, man.PosedReqSendInterval, 0)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "创建POS完成的req发送器", "失败", "err", err)
		return
	}
	p.posedReqSender = sender
}

func (p *Process) closePosedReqSender() {
	if p.posedReqSender == nil {
		return
	}
	p.posedReqSender.Close()
	p.posedReqSender = nil
}

func (p *Process) sendPosedReqFunc(data interface{}, times uint32) {
	req, OK := data.(*mc.HD_BlkConsensusReqMsg)
	if !OK {
		log.ERROR(p.logExtraInfo(), "发出POS完成的req", "反射消息失败")
		return
	}
	//给广播节点发送区块验证请求(带签名列表)
	log.INFO(p.logExtraInfo(), "发出POS完成的req(to broadcast) leader", req.Header.Leader.Hex(), "次数", times, "高度", p.number)
	p.pm.hd.SendNodeMsg(mc.HD_BlkConsensusReq, req, common.RoleBroadcast, nil)
}

func (p *Process) startVoteMsgSender(vote *mc.HD_ConsensusVote) {
	p.closeVoteMsgSender()
	sender, err := common.NewResendMsgCtrl(vote, p.sendVoteMsgFunc, man.BlkVoteSendInterval, man.BlkVoteSendTimes)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "创建投票消息发送器", "失败", "err", err)
		return
	}
	p.voteMsgSender = sender
}

func (p *Process) closeVoteMsgSender() {
	if p.voteMsgSender == nil {
		return
	}
	p.voteMsgSender.Close()
	p.voteMsgSender = nil
}

func (p *Process) sendVoteMsgFunc(data interface{}, times uint32) {
	vote, OK := data.(*mc.HD_ConsensusVote)
	if !OK {
		log.ERROR(p.logExtraInfo(), "发出投票消息", "反射消息失败")
		return
	}
	//给广播节点发送区块验证请求(带签名列表)
	log.INFO(p.logExtraInfo(), "发出投票消息 signHash", vote.SignHash.TerminalString(), "次数", times, "高度", p.number)
	p.pm.hd.SendNodeMsg(mc.HD_BlkConsensusVote, vote, common.RoleValidator, nil)
}
