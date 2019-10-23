// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect2

import (
	"time"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"
)

func (self *controller) handleMsg(data interface{}) {
	if nil == data {
		log.Warn(self.logInfo, "消息处理", "收到nil消息")
		return
	}

	switch data.(type) {
	case *startControllerMsg:
		msg, _ := data.(*startControllerMsg)
		self.handleStartMsg(msg)

	case *mc.BlockPOSFinishedNotify:
		msg, _ := data.(*mc.BlockPOSFinishedNotify)
		self.handleBlockPOSFinishedNotify(msg)

	case *mc.HD_V2_ReelectInquiryReqMsg:
		msg, _ := data.(*mc.HD_V2_ReelectInquiryReqMsg)
		self.handleInquiryReq(msg)

	case *mc.HD_V2_ReelectInquiryRspMsg:
		msg, _ := data.(*mc.HD_V2_ReelectInquiryRspMsg)
		self.handleInquiryRsp(msg)

	case *mc.HD_V2_ReelectLeaderReqMsg:
		msg, _ := data.(*mc.HD_V2_ReelectLeaderReqMsg)
		self.handleRLReq(msg)

	case *mc.HD_V2_ConsensusVote:
		msg, _ := data.(*mc.HD_V2_ConsensusVote)
		self.handleRLVote(msg)

	case *mc.HD_V2_ReelectBroadcastMsg:
		msg, _ := data.(*mc.HD_V2_ReelectBroadcastMsg)
		self.handleBroadcastMsg(msg)

	case *mc.HD_V2_ReelectBroadcastRspMsg:
		msg, _ := data.(*mc.HD_V2_ReelectBroadcastRspMsg)
		self.handleBroadcastRsp(msg)

	default:
		log.Warn(self.logInfo, "消息处理", "未知消息类型")
	}
}

func (self *controller) SetSelfAddress(addr common.Address, nodeAddr common.Address) {
	self.dc.selfAddr = addr
	self.dc.selfNodeAddr = nodeAddr
	self.selfCache.selfAddr = addr
	self.selfCache.selfNodeAddr = nodeAddr
}

func (self *controller) handleStartMsg(msg *startControllerMsg) {
	if nil == msg || nil == msg.parentHeader {
		log.Warn(self.logInfo, "开始消息处理", ErrParamsIsNil)
		return
	}

	if manversion.VersionCmp(string(msg.parentHeader.Version), manversion.VersionGamma) < 0 {
		log.Trace(self.logInfo, "开始消息处理", "版本号不匹配, 不处理消息", "header version", string(msg.parentHeader.Version))
		return
	}

	if self.mp.parentHeader != nil && self.mp.parentHeader.Time.Cmp(msg.parentHeader.Time) == 1 {
		log.Warn(self.logInfo, "开始消息处理", "parentHeader时间戳比现有时间戳小", "msg time", msg.parentHeader.Time, "cache time", self.mp.parentHeader.Time)
		return
	}

	a0Address := ca.GetDepositAddress()
	nodeAddress := ca.GetSignAddress()
	self.SetSelfAddress(a0Address, nodeAddress)

	log.Debug(self.logInfo, "开始消息处理", "start", "高度", self.dc.number, "preLeader", msg.parentHeader.Leader.Hex(), "header time", msg.parentHeader.Time.Int64())
	if err := self.dc.AnalysisState(msg.parentHeader, msg.parentStateDB); err != nil {
		log.Error(self.logInfo, "开始消息处理", "分析状态树信息错误", "err", err)
		return
	}

	if self.dc.role != common.RoleValidator {
		log.Debug(self.logInfo, "开始消息处理", "身份错误, 不是验证者", "高度", self.dc.number)
		self.mp.SaveParentHeader(msg.parentHeader)
		return
	}

	if self.dc.bcInterval.IsBroadcastNumber(self.dc.number) {
		log.Debug(self.logInfo, "开始消息处理", "区块为广播区块，不开启定时器")
		self.dc.state = stIdle
		self.publishLeaderMsg()
		self.mp.SaveParentHeader(msg.parentHeader)
		self.dc.state = stWaiting
		return
	}

	if self.dc.turnTime.SetBeginTime(msg.parentHeader.Time.Int64()) {
		self.mp.SaveParentHeader(msg.parentHeader)
		if isFirstConsensusTurn(self.ConsensusTurn()) {
			curTime := time.Now().Unix()
			st, remainTime, reelectTurn := self.dc.turnTime.CalState(0, curTime)
			log.Debug(self.logInfo, "开始消息处理", "完成", "状态计算结果", st.String(), "剩余时间", remainTime, "重选轮次", reelectTurn)
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
	self.publishLeaderMsg()
}

func (self *controller) handleBlockPOSFinishedNotify(msg *mc.BlockPOSFinishedNotify) {
	if nil == msg || nil == msg.Header {
		log.Warn(self.logInfo, "POS完成通知消息处理", ErrParamsIsNil)
		return
	}
	if err := self.mp.SavePOSNotifyMsg(msg); err == nil {
		log.Debug(self.logInfo, "POS完成通知消息处理", "缓存成功", "高度", msg.Number, "leader", msg.Header.Leader, "leader轮次", msg.ConsensusTurn.String())
	}
	self.processPOSState()
}

func (self *controller) timeOutHandle() {
	curTime := time.Now().Unix()
	st, remainTime, reelectTurn := self.dc.turnTime.CalState(self.dc.curConsensusTurn.TotalTurns(), curTime)
	switch self.State() {
	case stPos:
		log.Warn(self.logInfo, "超时事件", "POS未完成", "轮次", self.curTurnInfo(), "高度", self.Number(),
			"状态计算结果", st.String(), "下次超时时间", remainTime, "计算的重选轮次", reelectTurn,
			"轮次开始时间", self.dc.turnTime.GetBeginTime(self.ConsensusTurn().TotalTurns()), "leader", self.dc.GetConsensusLeader().Hex())
	case stReelect:
		log.Warn(self.logInfo, "超时事件", "重选未完成", "轮次", self.curTurnInfo(), "高度", self.Number(),
			"状态计算结果", st.String(), "下次超时时间", remainTime, "计算的重选轮次", reelectTurn,
			"轮次开始时间", self.dc.turnTime.GetBeginTime(self.ConsensusTurn().TotalTurns()), "master", self.dc.GetReelectMaster().Hex())
	default:
		log.Error(self.logInfo, "超时事件", "当前状态错误", "state", self.State().String(), "轮次", self.curTurnInfo(), "高度", self.Number(),
			"轮次开始时间", self.dc.turnTime.GetBeginTime(self.ConsensusTurn().TotalTurns()), "当前时间", curTime)
		return
	}

	self.setTimer(remainTime, self.timer)
	self.dc.state = st
	self.startReelect(reelectTurn)
}

func (self *controller) processPOSState() {
	if self.State() != stPos {
		log.Debug(self.logInfo, "执行检查POS状态", "状态不正常,不执行", "当前状态", self.State().String())
		return
	}

	if _, err := self.mp.GetPOSNotifyMsg(self.dc.GetConsensusLeader(), self.dc.curConsensusTurn); err != nil {
		log.Debug(self.logInfo, "执行检查POS状态", "获取POS完成消息失败", "err", err)
		return
	}

	log.Debug(self.logInfo, "POS完成", "状态切换为<挖矿结果等待阶段>")
	self.setTimer(0, self.timer)
	self.dc.state = stMining
}
