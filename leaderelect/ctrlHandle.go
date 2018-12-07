// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"time"
)

func (self *controller) handleMsg(data interface{}) {
	if nil == data {
		log.WARN(self.logInfo, "消息处理", "收到nil消息")
		return
	}

	switch data.(type) {
	case *startControllerMsg:
		msg, _ := data.(*startControllerMsg)
		self.handleStartMsg(msg)

	case *mc.BlockPOSFinishedNotify:
		msg, _ := data.(*mc.BlockPOSFinishedNotify)
		self.handleBlockPOSFinishedNotify(msg)

	case *mc.HD_ReelectInquiryReqMsg:
		msg, _ := data.(*mc.HD_ReelectInquiryReqMsg)
		self.handleInquiryReq(msg)

	case *mc.HD_ReelectInquiryRspMsg:
		msg, _ := data.(*mc.HD_ReelectInquiryRspMsg)
		self.handleInquiryRsp(msg)

	case *mc.HD_ReelectLeaderReqMsg:
		msg, _ := data.(*mc.HD_ReelectLeaderReqMsg)
		self.handleRLReq(msg)

	case *mc.HD_ReelectLeaderVoteMsg:
		msg, _ := data.(*mc.HD_ReelectLeaderVoteMsg)
		self.handleRLVote(msg)

	case *mc.HD_ReelectResultBroadcastMsg:
		msg, _ := data.(*mc.HD_ReelectResultBroadcastMsg)
		self.handleResultBroadcastMsg(msg)

	case *mc.HD_ReelectResultRspMsg:
		msg, _ := data.(*mc.HD_ReelectResultRspMsg)
		self.handleResultRsp(msg)

	default:
		log.WARN(self.logInfo, "消息处理", "未知消息类型")
	}
}

func (self *controller) handleStartMsg(msg *startControllerMsg) {
	if nil == msg || nil == msg.parentHeader {
		log.WARN(self.logInfo, "处理开始消息错误", "nil")
		return
	}

	log.INFO(self.logInfo, "处理开始消息", "start", "身份", msg.role, "高度", self.dc.number, "preLeader", msg.parentHeader.Leader, "header time", msg.parentHeader.Time.Int64())
	if msg.role != common.RoleValidator {
		log.INFO(self.logInfo, "处理开始消息", "身份不是验证者", "role", msg.role)
		//非验证者身份，保存父区块，以便收到 过低区块询问请求时，给出应答
		self.mp.SaveParentHeader(msg.parentHeader)
		return
	}

	preIsSupper := msg.parentHeader.IsSuperHeader()
	if err := self.dc.SetValidators(msg.parentHeader.Hash(), preIsSupper, msg.parentHeader.Leader, msg.validators); err != nil {
		log.ERROR(self.logInfo, "处理开始消息", "验证者列表设置错误", "err", err)
		return
	}

	if common.IsBroadcastNumber(self.dc.number) {
		log.INFO(self.logInfo, "处理开始消息", "区块为广播区块，不开启定时器", "role", msg.role)
		self.dc.state = stIdle
		self.sendLeaderMsg()
		self.mp.SaveParentHeader(msg.parentHeader)
		return
	}

	if self.dc.turnTime.SetBeginTime(0, msg.parentHeader.Time.Int64()) {
		log.INFO(self.logInfo, "处理开始消息", "更新轮次时间成功", "高度", self.dc.number)
		self.mp.SaveParentHeader(msg.parentHeader)
		if self.ConsensusTurn() == 0 {
			curTime := time.Now().Unix()
			st, remainTime, reelectTurn := self.dc.turnTime.CalState(0, curTime)
			log.INFO(self.logInfo, "处理开始消息", "计算状态结果", "状态", st, "剩余时间", remainTime, "重选轮次", reelectTurn)
			self.dc.state = st
			self.dc.curReelectTurn = 0
			self.setTimer(remainTime, self.timer)
			if st == stPos {
				self.processPOSState()
			} else if st == stReelect {
				self.startReelect(reelectTurn)
			}
		}
	}

	//公布leader身份
	self.sendLeaderMsg()
}

func (self *controller) handleBlockPOSFinishedNotify(msg *mc.BlockPOSFinishedNotify) {
	if nil == msg {
		log.WARN(self.logInfo, "处理POS完成通知消息错误", "nil")
		return
	}
	self.mp.SavePOSNotifyMsg(msg)
	self.processPOSState()
}

func (self *controller) timeOutHandle() {
	switch self.State() {
	case stPos:
		log.INFO(self.logInfo, "超时事件", "POS未完成", "轮次", self.curTurnInfo(), "高度", self.Number(),
			"轮次开始时间", self.dc.turnTime.GetBeginTime(self.ConsensusTurn()), "leader", self.dc.GetConsensusLeader().Hex())
		remainTime := self.dc.turnTime.CalRemainTime(self.dc.curConsensusTurn, 1, time.Now().Unix())
		//todo 负数怎么办
		self.setTimer(remainTime, self.timer)
		self.dc.state = stReelect
		self.startReelect(1)

	case stReelect:
		log.INFO(self.logInfo, "超时事件", "重选未完成", "轮次", self.curTurnInfo(), "高度", self.Number(),
			"轮次开始时间", self.dc.turnTime.GetBeginTime(self.ConsensusTurn()), "master", self.dc.GetReelectMaster().Hex())
		reelectTurn := self.dc.curReelectTurn + 1
		remainTime := self.dc.turnTime.CalRemainTime(self.dc.curConsensusTurn, reelectTurn, time.Now().Unix())
		//todo 负数怎么办
		self.setTimer(remainTime, self.timer)
		self.startReelect(reelectTurn)
	}
}

func (self *controller) processPOSState() {
	if self.State() != stPos {
		log.INFO(self.logInfo, "执行检查POS状态", "状态不正常,不执行", "当前状态", self.State().String())
		return
	}

	if _, err := self.mp.GetPOSNotifyMsg(self.dc.GetConsensusLeader(), self.dc.curConsensusTurn); err != nil {
		log.INFO(self.logInfo, "执行检查POS状态", "获取POS完成消息失败", "err", err)
		return
	}

	self.dc.state = stMining
}

func (self *controller) processNewBlockReadyRsp(header *types.Header, from common.Address) {
	if nil == header {
		log.ERROR(self.logInfo, "处理新区块响应", "区块header为nil")
		return
	}

	number := header.Number.Uint64()
	parentHeader := self.matrix.BlockChain().GetHeader(header.ParentHash, number-1)
	if parentHeader == nil {
		log.ERROR(self.logInfo, "处理新区块响应", "没有父区块，进行fetch", "parent number", number-1, "parent hash", header.ParentHash.TerminalString())
		self.matrix.FetcherNotify(header.ParentHash, number-1)
		return
	}

	//POW验证
	err := self.matrix.Engine().VerifyHeader(self.matrix.BlockChain(), header, true)
	if err != nil {
		log.ERROR(self.logInfo, "处理新区块响应", "POW验证失败", "高度", number, "hash", header.Hash().TerminalString(), "err", err)
		return
	}

	//POS验证
	err = self.matrix.DPOSEngine().VerifyBlock(self.dc, header)
	if err != nil {
		log.ERROR(self.logInfo, "处理新区块响应", "POS验证失败", "高度", number, "hash", header.Hash().TerminalString(), "err", err)
		return
	}

	//发送恢复状态消息
	log.INFO(self.logInfo, "处理新区块响应", "发送恢复状态消息")
	mc.PublishEvent(mc.Leader_RecoveryState, &mc.RecoveryStateMsg{Type: mc.RecoveryTypeFullHeader, Header: header, From: from})
}
