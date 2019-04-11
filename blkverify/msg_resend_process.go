// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package blkverify

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

func (p *Process) stopSender() {
	p.closeMineReqMsgSender()
	p.closePosedReqSender()
	p.closeVoteMsgSender()
}

func (p *Process) startSendMineReq(req *mc.HD_MiningReqMsg) {
	p.closeMineReqMsgSender()
	sender, err := common.NewResendMsgCtrl(req, p.sendMineReqFunc, manparams.MinerReqSendInterval, 0)
	if err != nil {
		log.Error(p.logExtraInfo(), "创建挖矿请求发送器", "失败", "err", err)
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
		log.Error(p.logExtraInfo(), "发出挖矿请求", "反射消息失败")
		return
	}
	hash := req.Header.HashNoSignsAndNonce()
	//给矿工发送区块验证结果
	if times == 1 {
		log.Info(p.logExtraInfo(), "发出挖矿请求, Header hash with signs", hash, "高度", p.number)
	} else {
		log.Trace(p.logExtraInfo(), "发出挖矿请求, Header hash with signs", hash, "次数", times, "高度", p.number)
	}
	p.pm.hd.SendNodeMsg(mc.HD_MiningReq, req, common.RoleMiner|common.RoleInnerMiner, nil)
}

func (p *Process) startPosedReqSender(req *mc.HD_BlkConsensusReqMsg) {
	p.closePosedReqSender()
	sender, err := common.NewResendMsgCtrl(req, p.sendPosedReqFunc, manparams.PosedReqSendInterval, 0)
	if err != nil {
		log.Error(p.logExtraInfo(), "创建POS完成的req发送器", "失败", "err", err)
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
		log.Error(p.logExtraInfo(), "发出POS完成的req", "反射消息失败")
		return
	}
	//给广播节点发送区块验证请求(带签名列表)
	if times == 1 {
		log.Debug(p.logExtraInfo(), "发出POS完成的req(to broadcast) leader", req.Header.Leader.Hex(), "高度", p.number)
	} else {
		log.Trace(p.logExtraInfo(), "发出POS完成的req(to broadcast) leader", req.Header.Leader.Hex(), "次数", times, "高度", p.number)
	}
	p.pm.hd.SendNodeMsg(mc.HD_BlkConsensusReq, req, common.RoleBroadcast, nil)
}

func (p *Process) startVoteMsgSender(vote *mc.HD_ConsensusVote) {
	p.closeVoteMsgSender()
	sender, err := common.NewResendMsgCtrl(vote, p.sendVoteMsgFunc, manparams.BlkVoteSendInterval, manparams.BlkVoteSendTimes)
	if err != nil {
		log.Error(p.logExtraInfo(), "创建投票消息发送器", "失败", "err", err)
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
		log.Error(p.logExtraInfo(), "发出投票消息", "反射消息失败")
		return
	}
	//发送投票消息
	if times == 1 {
		log.Info(p.logExtraInfo(), "发出投票消息 signHash", vote.SignHash.TerminalString(), "高度", p.number)
	} else {
		log.Trace(p.logExtraInfo(), "发出投票消息 signHash", vote.SignHash.TerminalString(), "次数", times, "高度", p.number)
	}
	p.pm.hd.SendNodeMsg(mc.HD_BlkConsensusVote, vote, common.RoleValidator, nil)
}
