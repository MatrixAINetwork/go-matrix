// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package leaderelect2

import (
	"time"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
)

func (self *controller) startReelect(reelectTurn uint32) {
	log.Info(self.logInfo, "重选流程", "开启处理", "重选轮次", reelectTurn, "共识轮次", self.dc.curConsensusTurn.String(), "高度", self.dc.number)
	if self.State() != stReelect {
		log.Trace(self.logInfo, "开启重选流程", "当前状态不是重选状态，不处理", "状态", self.State().String(), "重选轮次", reelectTurn, "共识轮次", self.dc.curConsensusTurn.String(), "高度", self.dc.number)
		return
	}
	if self.dc.curReelectTurn == reelectTurn {
		log.Trace(self.logInfo, "开启重选流程", "重选已启动，不处理", "重选轮次", reelectTurn, "共识轮次", self.dc.curConsensusTurn.String(), "高度", self.dc.number)
		return
	}

	if err := self.dc.SetReelectTurn(reelectTurn); err != nil {
		log.Error(self.logInfo, "开启重选流程", "设置重选轮次失败", "err", err, "高度", self.dc.number)
		return
	}
	beginTime, endTime := self.dc.turnTime.CalTurnTime(self.dc.curConsensusTurn.TotalTurns(), self.dc.curReelectTurn)
	master := self.dc.GetReelectMaster()
	if master == self.dc.selfAddr {
		log.Debug(self.logInfo, "(master)开启重选流程", master.Hex(), "轮次", self.curTurnInfo(), "高度", self.dc.number,
			"轮次开始时间", time.Unix(beginTime, 0).String(), "轮次结束时间", time.Unix(endTime, 0).String(), "self", self.dc.selfAddr.Hex())
		self.dc.isMaster = true
		self.setTimer(self.dc.turnTime.reelectHandleInterval, self.reelectTimer)
		self.sendInquiryReq()
	} else {
		log.Debug(self.logInfo, "(follower)开启重选流程", master.Hex(), "轮次", self.curTurnInfo(), "高度", self.dc.number,
			"轮次开始时间", time.Unix(beginTime, 0).String(), "轮次结束时间", time.Unix(endTime, 0).String(), "self", self.dc.selfAddr.Hex())
		self.dc.isMaster = false
		self.setTimer(0, self.reelectTimer)
	}

	self.publishLeaderMsg()
}

func (self *controller) finishReelectWithPOS(posResult *mc.HD_BlkConsensusReqMsg, from common.Address) {
	log.INFO(self.logInfo, "完成leader重选", "POS结果重置，恢复并开始挖矿等待", "共识轮次", self.ConsensusTurn().String(), "高度", self.Number())
	mc.PublishEvent(mc.Leader_RecoveryState, &mc.RecoveryStateMsg{Type: mc.RecoveryTypePOS, Header: posResult.Header, From: from})
	self.setTimer(0, self.timer)
	self.setTimer(0, self.reelectTimer)
	self.dc.state = stMining
	self.dc.SetReelectTurn(0)
	self.dc.isMaster = false
	self.selfCache.ClearSelfInquiryMsg()
	self.publishLeaderMsg()
}

func (self *controller) finishReelectWithRLConsensus(rlResult *mc.HD_V2_ReelectLeaderConsensus) {
	consensusTurn := calcNextConsensusTurn(rlResult.Req.InquiryReq.ConsensusTurn, rlResult.Req.InquiryReq.ReelectTurn)
	if err := self.dc.SetConsensusTurn(consensusTurn); err != nil {
		log.Error(self.logInfo, "完成leader重选", "leader重置, 设置共识轮次失败", "err", err)
		return
	}

	//缓存共识结果消息
	self.mp.SaveRLConsensusMsg(rlResult)

	self.setTimer(0, self.reelectTimer)
	self.selfCache.ClearSelfInquiryMsg()
	self.dc.isMaster = false
	curTime := time.Now().Unix()
	st, remainTime, reelectTurn := self.dc.turnTime.CalState(consensusTurn.TotalTurns(), curTime)
	log.INFO(self.logInfo, "完成leader重选", "leader重置", "重选轮次", reelectTurn, "旧共识轮次", self.ConsensusTurn().String(), "新共识轮次", consensusTurn.String(), "高度", self.Number(),
		"状态计算结果", st.String(), "下次超时时间", remainTime, "计算的重选轮次", reelectTurn, "轮次开始时间", self.dc.turnTime.GetBeginTime(self.ConsensusTurn().TotalTurns()))
	self.dc.state = st
	self.dc.curReelectTurn = 0
	self.setTimer(remainTime, self.timer)
	if st == stPos {
		self.processPOSState()
	} else if st == stReelect {
		self.startReelect(reelectTurn)
	}

	self.publishLeaderMsg()
}

func (self *controller) reelectTimeOutHandle() {
	if self.State() != stReelect {
		log.Info(self.logInfo, "重选处理定时器超时", "状态错误,当前状态不是重选阶段", "当前状态", self.State().String())
		return
	}
	switch self.selfCache.GetInquiryResult() {
	case mc.ReelectRSPTypeNone:
		self.sendInquiryReq()
	case mc.ReelectRSPTypeAgree:
		self.sendRLReq()
	case mc.ReelectRSPTypePOS, mc.ReelectRSPTypeAlreadyRL:
		self.sendResultBroadcastMsg()
	default:
		log.Warn(self.logInfo, "重选处理定时器超时", "当前询问结果错误", "inquiryResult", self.selfCache.GetInquiryResult())
	}
	self.setTimer(self.dc.turnTime.reelectHandleInterval, self.reelectTimer)
}

func (self *controller) handleInquiryReq(req *mc.HD_V2_ReelectInquiryReqMsg) {
	if nil == req {
		log.INFO(self.logInfo, "询问请求处理", "消息为nil")
		return
	}
	if self.State() == stIdle {
		log.INFO(self.logInfo, "询问请求处理", "当前状态为idle，忽略消息", "from", req.From.Hex(), "高度", self.dc.number)
		return
	}

	fromMaster, _, err := self.dc.GetA0AccountFromAnyAccount(req.From, self.dc.leaderCal.preHash)
	if err != nil {
		log.Info(self.logInfo, "询问请求处理", "根据from获取master的A0账户失败!", "from", req.From.Hex(), "err", err)
		return
	}

	if req.Master != fromMaster {
		log.INFO(self.logInfo, "询问请求处理", "消息master与from不匹配", "master", req.Master.Hex(), "from", fromMaster.Hex(), "高度", self.dc.number)
		return
	}
	log.Debug(self.logInfo, "询问消息处理", "开始", "高度", req.Number, "共识轮次", req.ConsensusTurn.String(), "重选轮次", req.ReelectTurn, "本地轮次信息", self.curTurnInfo(), "from", req.From.Hex())

	// 对比请求高度
	if req.Number < self.Number() {
		// 消息高度<本地高度: 请求方高度落后
		reqHash := types.RlpHash(req)
		log.Trace(self.logInfo, "询问请求处理", "请求高度<本地高度, 发送响应(新区块已准备完毕)", "请求高度", req.Number, "本地高度", self.Number(), "reqHash", reqHash.TerminalString())
		self.sendInquiryRspWithNewBlockReady(reqHash, req.From, req.Number)
		return
	} else if req.Number > self.Number() {
		// 消息高度>本地高度: 本地高度落后
		if self.State() == stReelect && self.dc.isMaster {
			log.Trace(self.logInfo, "询问请求处理", "高度落后，但当前是master，不另外发送询问", "请求高度", req.Number, "本地高度", self.Number(), "目标", req.From.Hex())
			return
		} else {
			log.Trace(self.logInfo, "询问请求处理", "请求高度>本地高度, 主动询问对方状态", "请求高度", req.Number, "本地高度", self.Number(), "目标", req.From.Hex())
			self.sendInquiryReqToSingle(req.From)
			return
		}
	}

	// 高度相同，对比header时间戳
	curHeaderTime := self.mp.parentHeader.Time.Uint64()
	if req.HeaderTime < curHeaderTime {
		// 消息HeaderTime<本地HeaderTime: 请求方HeaderTime落后
		reqHash := types.RlpHash(req)
		log.Trace(self.logInfo, "询问请求处理", "请求HeaderTime<本地HeaderTime, 发送响应(新区块已准备完毕)",
			"请求HeaderTime", req.HeaderTime, "本地HeaderTime", curHeaderTime, "reqHash", reqHash.TerminalString())
		self.sendInquiryRspWithNewBlockReady(reqHash, req.From, req.Number)
		return
	} else if req.HeaderTime > curHeaderTime {
		// 消息HeaderTime>本地HeaderTime: 本地HeaderTime落后
		if self.State() == stReelect && self.dc.isMaster {
			log.Trace(self.logInfo, "询问请求处理", "HeaderTime落后，但当前是master，不另外发送询问",
				"请求HeaderTime", req.HeaderTime, "本地HeaderTime", curHeaderTime, "req from", req.From.Hex())
			return
		} else {
			log.Trace(self.logInfo, "询问请求处理", "请求HeaderTime>本地HeaderTime, 主动询问对方状态",
				"请求HeaderTime", req.HeaderTime, "本地HeaderTime", curHeaderTime, "高度", req.Number, "目标", req.From.Hex())
			self.sendInquiryReqToSingle(req.From)
			return
		}
	}

	// 高度、headerTime相同，对比共识轮次
	if req.ConsensusTurn.Cmp(self.dc.curConsensusTurn) < 0 {
		// 消息轮次<本地轮次: 请求方共识轮次落后
		log.Trace(self.logInfo, "询问请求处理", "请求共识轮次<本地共识轮次, 发送响应(重选共识结果)", "消息共识轮次", req.ConsensusTurn.String(), "本地共识轮次", self.dc.curConsensusTurn.String(), "高度", self.dc.number)
		self.sendInquiryRspWithRLConsensus(types.RlpHash(req), req.From)
		return
	} else if req.ConsensusTurn.Cmp(self.dc.curConsensusTurn) > 0 {
		// 消息轮次>本地轮次: 本地共识轮次落后
		log.Trace(self.logInfo, "询问请求处理", "请求共识轮次>本地共识轮次, 主动询问对方状态", "消息共识轮次", req.ConsensusTurn.String(), "本地共识轮次", self.dc.curConsensusTurn.String(), "高度", self.dc.number)
		if self.State() == stReelect && self.dc.isMaster {
			log.Trace(self.logInfo, "询问请求处理", "共识轮次落后，但当前是master，不另外发送询问", "消息共识轮次", req.ConsensusTurn.String(), "本地共识轮次", self.dc.curConsensusTurn.String(), "高度", self.dc.number)
			return
		} else {
			self.sendInquiryReqToSingle(req.From)
			return
		}
	} else {
		master, err := self.dc.GetLeader(req.ConsensusTurn.TotalTurns()+req.ReelectTurn, self.dc.bcInterval)
		if err != nil {
			log.Info(self.logInfo, "询问请求处理", "验证消息合法性", "本地计算master失败", err)
			return
		}

		if master != req.Master {
			log.Info(self.logInfo, "询问请求处理", "验证消息合法性失败，master不匹配", "master", req.Master.Hex(), "local master", master.Hex())
			return
		}
		switch self.State() {
		case stMining:
			reqHash := types.RlpHash(req)
			log.Debug(self.logInfo, "询问消息处理", "本地状态为mining, 发送POS完成响应", "高度", req.Number, "from", req.From.Hex(), "req hash", reqHash.TerminalString())
			self.sendInquiryRspWithPOS(reqHash, req.From, req.Number)
			return

		case stReelect:
			if req.ReelectTurn != self.dc.curReelectTurn {
				log.Debug(self.logInfo, "询问请求处理", "重选轮次不匹配，忽略消息", "消息重选轮次", req.ReelectTurn, "本地重选轮次", self.dc.curReelectTurn, "高度", self.dc.number)
				return
			}
			if err := self.dc.turnTime.CheckTimeLegal(self.dc.curConsensusTurn.TotalTurns(), self.dc.curReelectTurn, int64(req.TimeStamp)); err != nil {
				log.Debug(self.logInfo, "询问请求处理", "请求时间搓检查", "异常", err, "轮次", self.curTurnInfo(), "高度", self.dc.number)
				return
			}
			self.sendInquiryRspWithAgree(types.RlpHash(req), req.From, req.Number)

		default:
			log.INFO(self.logInfo, "询问请求处理", "本地状态异常，不响应请求", "本地状态", self.State().String())
			return
		}
	}
}

func (self *controller) handleInquiryRsp(rsp *mc.HD_V2_ReelectInquiryRspMsg) {
	log.Trace(self.logInfo, "询问响应处理", "开始", "req hash", rsp.ReqHash.TerminalString(), "类型", rsp.Type, "from", rsp.From.Hex())
	if self.selfCache.GetInquiryResult() != mc.ReelectRSPTypeNone {
		log.Trace(self.logInfo, "询问响应处理", "本地已有询问结果，忽略消息", "当前询问结果", self.selfCache.GetInquiryResult())
		return
	}

	if err := self.selfCache.IsMatchedInquiryRsp(rsp); err != nil {
		log.Info(self.logInfo, "询问响应处理", "响应匹配检查", "err", err, "高度", self.dc.number)
		return
	}
	switch rsp.Type {
	case mc.ReelectRSPTypeNewBlockReady:
		self.processNewBlockReadyRsp(rsp.NewBlock, rsp.From)
		return

	case mc.ReelectRSPTypeAlreadyRL:
		if err := self.checkRLResult(rsp.RLResult); err != nil {
			log.Info(self.logInfo, "询问响应处理(leader重选已完成)", "消息检查", "err", err)
			return
		}
		if err := self.selfCache.GenBroadcastMsgWithInquiryResult(rsp.Type, rsp); err != nil {
			log.Warn(self.logInfo, "询问响应处理(leader重选已完成)", "生成广播消息失败", "结果类型", rsp.Type, "err", err)
			return
		}
		self.sendResultBroadcastMsg()
		self.finishReelectWithRLConsensus(rsp.RLResult)

	case mc.ReelectRSPTypePOS:
		if err := self.checkPOSResult(rsp.POSResult); err != nil {
			log.Info(self.logInfo, "询问响应处理(POS完成响应)", "消息检查", "err", err)
			return
		}
		if err := self.selfCache.GenBroadcastMsgWithInquiryResult(rsp.Type, rsp); err != nil {
			log.Warn(self.logInfo, "询问响应处理(POS完成响应)", "生存广播消息失败", "结果类型", rsp.Type, "err", err)
			return
		}
		log.Trace(self.logInfo, "询问响应处理(POS完成响应)", "开始广播结果")
		self.sendResultBroadcastMsg()

	case mc.ReelectRSPTypeAgree:
		if err := self.selfCache.SaveInquiryVote(rsp.ReqHash, rsp.AgreeSign, rsp.From, self.dc, self.matrix.SignHelper()); err != nil {
			log.Info(self.logInfo, "询问响应处理(同意更换leader响应)", "保存同意签名失败", "err", err)
			return
		}

		signs := self.selfCache.GetInquiryVotes()
		log.Trace(self.logInfo, "询问响应处理(同意更换leader响应)", "保存签名成功", "签名总数", len(signs))
		rightSigns, err := self.matrix.DPOSEngine().VerifyHashWithVerifiedSignsAndBlock(self.dc, signs, self.ParentHash())
		if err != nil {
			log.Trace(self.logInfo, "询问响应处理(同意更换leader响应)", "同意的签名没有通过POS共识", "err", err)
			return
		}

		if err := self.selfCache.GenRLReqMsg(rightSigns); err != nil {
			log.Error(self.logInfo, "询问响应处理(同意更换leader响应)", "生存更换leader请求消息失败", "err", err)
			return
		}
		log.Trace(self.logInfo, "询问响应处理(同意更换leader响应)", "POS共识通过, 发送更换leader请求")
		self.sendRLReq()
	}
}

func (self *controller) handleRLReq(req *mc.HD_V2_ReelectLeaderReqMsg) {
	if self.State() != stReelect {
		log.Info(self.logInfo, "leader重选请求处理", "本地状态错误", "本地状态", self.State().String())
		return
	}
	if err := self.checkRLReqMsg(req); err != nil {
		log.Info(self.logInfo, "leader重选请求处理", "消息异常", "err", err)
		return
	}

	hash := types.RlpHash(req)
	sign, err := self.matrix.SignHelper().SignHashWithValidateByReader(self.dc, hash.Bytes(), true, self.ParentHash())
	if err != nil {
		log.Error(self.logInfo, "leader重选请求处理", "签名失败", "err", err)
		return
	}
	rsp := &mc.HD_ConsensusVote{
		Number:   self.dc.number,
		SignHash: hash,
		Sign:     sign,
	}
	log.Trace(self.logInfo, "leader重选请求处理", "发送投票", "高度", self.dc.number, "req hash", rsp.SignHash.TerminalString(), "target", req.InquiryReq.From.Hex())
	self.matrix.HD().SendNodeMsg(mc.HD_V2_LeaderReelectVote, rsp, common.RoleNil, []common.Address{req.InquiryReq.From})
}

func (self *controller) handleRLVote(msg *mc.HD_V2_ConsensusVote) {
	if nil == msg {
		log.Info(self.logInfo, "处理leader重选响应", "消息为nil")
		return
	}
	if err := self.selfCache.SaveRLVote(msg.SignHash, msg.Sign, msg.From, self.dc, self.matrix.SignHelper()); err != nil {
		log.Info(self.logInfo, "处理leader重选响应", "保存签名错误", "err", err)
		return
	}
	signs := self.selfCache.GetRLVotes()
	rightSigns, err := self.matrix.DPOSEngine().VerifyHashWithVerifiedSignsAndBlock(self.dc, signs, self.ParentHash())
	if err != nil {
		log.Debug(self.logInfo, "处理leader重选响应", "签名没有通过POS共识", "总票数", len(signs), "err", err)
		return
	}

	log.Trace(self.logInfo, "处理leader重选响应", "POS共识通过, 准备发送重选结果广播")
	if err := self.selfCache.GenBroadcastMsgWithRLSuccess(rightSigns); err != nil {
		log.Warn(self.logInfo, "处理leader重选响应", "生成广播(leader重选成功)消息失败", "err", err)
		return
	}
	self.sendResultBroadcastMsg()
}

func (self *controller) handleBroadcastMsg(msg *mc.HD_V2_ReelectBroadcastMsg) {
	if nil == msg {
		log.Info(self.logInfo, "处理重选结果广播", "消息为nil")
		return
	}
	if err := self.processResultBroadcastMsg(msg); err != nil {
		log.Info(self.logInfo, "处理重选结果广播失败", err)
		return
	}
	self.sendResultBroadcastRsp(msg)
}

func (self *controller) handleBroadcastRsp(rsp *mc.HD_V2_ReelectBroadcastRspMsg) {
	if nil == rsp {
		log.Error(self.logInfo, "处理重选结果广播响应", "响应消息为nil")
		return
	}

	if err := self.selfCache.SaveBroadcastVote(rsp.ResultHash, rsp.Sign, rsp.From, self.dc, self.matrix.SignHelper()); err != nil {
		log.Error(self.logInfo, "处理重选结果广播响应", "保存响应失败", "err", err)
		return
	}
	signs := self.selfCache.GetBroadcastVotes()
	_, err := self.matrix.DPOSEngine().VerifyHashWithVerifiedSignsAndBlock(self.dc, signs, self.ParentHash())
	if err != nil {
		log.INFO(self.logInfo, "处理重选结果广播响应", "响应没有通过POS共识", "票总数", len(signs), "err", err)
		return
	}
	log.Trace(self.logInfo, "处理重选结果广播响应", "POS共识通过, 准备处理广播结果")
	resultMsg, err := self.selfCache.GetLocalBroadcastMsg()
	if err != nil {
		log.Error(self.logInfo, "处理本地重选结果广播", "获取本地重选结果广播错误", "err", err)
		return
	}
	if err := self.processResultBroadcastMsg(resultMsg); err != nil {
		log.Error(self.logInfo, "处理本地重选结果广播失败", err)
		return
	}
}

func (self *controller) processResultBroadcastMsg(msg *mc.HD_V2_ReelectBroadcastMsg) error {
	if msg == nil {
		return ErrParamsIsNil
	}
	switch msg.Type {
	case mc.ReelectRSPTypePOS:
		posResult := msg.POSResult
		if nil == posResult || nil == posResult.Header {
			return ErrPOSResultIsNil
		}
		if posResult.Header.Leader != self.dc.GetConsensusLeader() {
			return errors.Errorf("消息中headerLeader(%s) != 本地共识leader(%s)", posResult.Header.Leader.Hex(), self.dc.GetConsensusLeader().Hex())
		}
		if err := self.matrix.DPOSEngine().VerifyBlock(self.dc, posResult.Header); err != nil {
			return errors.Errorf("POS完成结果中的POS结果验证错误(%v)", err)
		}
		self.finishReelectWithPOS(posResult, msg.From)

	case mc.ReelectRSPTypeAgree, mc.ReelectRSPTypeAlreadyRL:
		if err := self.checkRLResult(msg.RLResult); err != nil {
			return err
		}
		self.finishReelectWithRLConsensus(msg.RLResult)
	default:
		return errors.Errorf("结果类型(%v)错误", msg.Type)
	}
	return nil
}

func (self *controller) sendInquiryReq() {
	req := &mc.HD_V2_ReelectInquiryReqMsg{
		Number:        self.Number(),
		HeaderTime:    self.mp.parentHeader.Time.Uint64(),
		ConsensusTurn: self.dc.curConsensusTurn,
		ReelectTurn:   self.dc.curReelectTurn,
		TimeStamp:     uint64(time.Now().Unix()),
		Master:        self.dc.selfAddr,
		From:          self.dc.selfNodeAddr,
	}
	reqHash := self.selfCache.SaveInquiryReq(req)
	selfSign, err := self.matrix.SignHelper().SignHashWithValidateByReader(self.dc, reqHash.Bytes(), true, self.ParentHash())
	if err != nil {
		log.Error(self.logInfo, "send<重选询问请求>", "自己的同意签名失败", "err", err, "高度", self.Number(), "轮次", self.curTurnInfo())
		return
	}
	if err := self.selfCache.SaveInquiryVote(reqHash, selfSign, self.dc.selfNodeAddr, self.dc, self.matrix.SignHelper()); err != nil {
		log.Error(self.logInfo, "send<重选询问请求>", "保存自己的同意签名错误", "err", err)
		return
	}

	log.Trace(self.logInfo, "send<重选询问请求>", "成功", "轮次", self.curTurnInfo(), "高度", self.Number(), "reqHash", reqHash.TerminalString())
	self.matrix.HD().SendNodeMsg(mc.HD_V2_LeaderReelectInquiryReq, req, common.RoleValidator, nil)
	return
}

func (self *controller) sendInquiryReqToSingle(target common.Address) {
	curTime := time.Now().Unix()
	if false == self.selfCache.CanSendSingleInquiryReq(curTime, self.dc.turnTime.reelectHandleInterval) {
		log.Trace(self.logInfo, "send<重选询问请求>single", "尚未达到发送间隔，不发送请求")
		return
	}
	req := &mc.HD_V2_ReelectInquiryReqMsg{
		Number:        self.Number(),
		HeaderTime:    self.mp.parentHeader.Time.Uint64(),
		ConsensusTurn: self.dc.curConsensusTurn,
		ReelectTurn:   self.dc.curReelectTurn,
		TimeStamp:     uint64(curTime),
		Master:        self.dc.selfAddr,
		From:          self.dc.selfNodeAddr,
	}
	reqHash := self.selfCache.SaveInquiryReq(req)
	log.Trace(self.logInfo, "send<重选询问请求>single", "成功", "轮次", self.curTurnInfo(), "高度", self.Number(), "reqHash", reqHash.TerminalString())
	self.matrix.HD().SendNodeMsg(mc.HD_V2_LeaderReelectInquiryReq, req, common.RoleNil, []common.Address{target})
	self.selfCache.SetLastSingleInquiryReqTime(curTime)
	return
}

func (self *controller) sendInquiryRspWithPOS(reqHash common.Hash, target common.Address, number uint64) {
	POSMsg, err := self.mp.GetPOSNotifyMsg(self.dc.GetConsensusLeader(), self.dc.curConsensusTurn)
	if err != nil {
		log.Warn(self.logInfo, "send<询问响应(POS完成响应)>", "获取POS通知消息错误", "err", err, "高度", number,
			"共识轮次", self.dc.curConsensusTurn, "共识leader", self.dc.GetConsensusLeader())
		return
	}
	rsp := &mc.HD_V2_ReelectInquiryRspMsg{
		Number:    number,
		ReqHash:   reqHash,
		Type:      mc.ReelectRSPTypePOS,
		AgreeSign: common.Signature{},
		POSResult: &mc.HD_BlkConsensusReqMsg{Header: POSMsg.Header, TxsCode: POSMsg.TxsCode, ConsensusTurn: self.dc.curConsensusTurn},
		RLResult:  nil,
		NewBlock:  nil,
	}
	self.matrix.HD().SendNodeMsg(mc.HD_V2_LeaderReelectInquiryRsp, rsp, common.RoleNil, []common.Address{target})
}

func (self *controller) sendInquiryRspWithAgree(reqHash common.Hash, target common.Address, number uint64) {
	sign, err := self.matrix.SignHelper().SignHashWithValidateByReader(self.dc, reqHash.Bytes(), true, self.ParentHash())
	if err != nil {
		log.Error(self.logInfo, "send<询问响应(同意更换leader响应)>", "签名失败", "err", err, "高度", number,
			"共识轮次", self.dc.curConsensusTurn.String(), "重选轮次", self.dc.curReelectTurn)
		return
	}
	rsp := &mc.HD_V2_ReelectInquiryRspMsg{
		Number:    number,
		ReqHash:   reqHash,
		Type:      mc.ReelectRSPTypeAgree,
		AgreeSign: sign,
		POSResult: nil,
		RLResult:  nil,
		NewBlock:  nil,
	}
	log.Trace(self.logInfo, "send<询问响应(同意更换leader响应)>", "成功", "reqHash", reqHash.TerminalString(), "高度", number,
		"轮次信息", self.curTurnInfo(), "目标", target.Hex())
	self.matrix.HD().SendNodeMsg(mc.HD_V2_LeaderReelectInquiryRsp, rsp, common.RoleNil, []common.Address{target})
}

func (self *controller) sendInquiryRspWithRLConsensus(reqHash common.Hash, target common.Address) {
	consensusMsg, err := self.mp.GetRLConsensusMsg(self.dc.curConsensusTurn)
	if err != nil {
		log.Error(self.logInfo, "send<询问响应(leader重选已完成)>", "获取leader重选共识消息错误", "err", err, "高度", self.Number(),
			"轮次信息", self.curTurnInfo(), "目标", target.Hex())
		return
	}
	rsp := &mc.HD_V2_ReelectInquiryRspMsg{
		Number:    self.Number(),
		ReqHash:   reqHash,
		Type:      mc.ReelectRSPTypeAlreadyRL,
		AgreeSign: common.Signature{},
		POSResult: nil,
		RLResult:  consensusMsg,
		NewBlock:  nil,
	}
	log.Trace(self.logInfo, "send<询问响应(leader重选已完成)>", "成功", "轮次", consensusMsg.Req.InquiryReq.ConsensusTurn, "高度", self.Number(),
		"轮次信息", self.curTurnInfo(), "目标", target.Hex())
	self.matrix.HD().SendNodeMsg(mc.HD_V2_LeaderReelectInquiryRsp, rsp, common.RoleNil, []common.Address{target})
}

func (self *controller) sendInquiryRspWithNewBlockReady(reqHash common.Hash, target common.Address, number uint64) {
	parentHeader := self.mp.GetParentHeader()
	if parentHeader == nil {
		log.Warn(self.logInfo, "send<询问响应(新区块已准备完毕响应)>", "获取区块失败")
		return
	}

	rsp := &mc.HD_V2_ReelectInquiryRspMsg{
		Number:    number,
		ReqHash:   reqHash,
		Type:      mc.ReelectRSPTypeNewBlockReady,
		AgreeSign: common.Signature{},
		POSResult: nil,
		RLResult:  nil,
		NewBlock:  parentHeader,
	}
	log.Trace(self.logInfo, "send<询问响应(新区块已准备完毕响应)>", "成功", "block hash", parentHeader.Hash().TerminalString(), "req hash", reqHash.TerminalString(), "to", target.Hex())
	self.matrix.HD().SendNodeMsg(mc.HD_V2_LeaderReelectInquiryRsp, rsp, common.RoleNil, []common.Address{target})
}

func (self *controller) sendRLReq() {
	req, reqHash, err := self.selfCache.GetRLReqMsg()
	if err != nil {
		log.Warn(self.logInfo, "send<leader重选请求>", "获取请求消息失败", "err", err)
		return
	}

	selfSign, err := self.matrix.SignHelper().SignHashWithValidateByReader(self.dc, reqHash.Bytes(), true, self.ParentHash())
	if err != nil {
		log.Error(self.logInfo, "send<leader重选请求>", "自己的签名失败", "err", err, "高度", self.Number(), "轮次", self.curTurnInfo())
		return
	}
	if err := self.selfCache.SaveRLVote(reqHash, selfSign, self.dc.selfNodeAddr, self.dc, self.matrix.SignHelper()); err != nil {
		log.Error(self.logInfo, "send<leader重选请求>", "保存自己签名错误", "err", err)
		return
	}

	log.Trace(self.logInfo, "send<Leader重选请求>, hash", reqHash.TerminalString(), "轮次", self.curTurnInfo(), "高度", self.Number())
	self.matrix.HD().SendNodeMsg(mc.HD_V2_LeaderReelectReq, req, common.RoleValidator, nil)
}

func (self *controller) sendResultBroadcastMsg() {
	msg, msgHash, err := self.selfCache.GetBroadcastMsg()
	if err != nil {
		log.Warn(self.logInfo, "send<重选结果广播>", "获取广播消息失败", "err", err)
		return
	}
	selfSign, err := self.matrix.SignHelper().SignHashWithValidateByReader(self.dc, msgHash.Bytes(), true, self.ParentHash())
	if err != nil {
		log.Error(self.logInfo, "send<重选结果广播>", "自己的响应签名失败", "err", err, "高度", self.Number(), "轮次", self.curTurnInfo())
		return
	}
	if err := self.selfCache.SaveBroadcastVote(msgHash, selfSign, self.dc.selfNodeAddr, self.dc, self.matrix.SignHelper()); err != nil {
		log.Error(self.logInfo, "send<重选结果广播>", "保存自己的响应失败", "err", err)
		return
	}
	log.Trace(self.logInfo, "send<重选结果广播>, hash", msgHash.TerminalString(), "轮次", self.curTurnInfo(), "高度", self.Number(), "类型", msg.Type)
	self.matrix.HD().SendNodeMsg(mc.HD_V2_LeaderReelectBroadcast, msg, common.RoleValidator, nil)
}

func (self *controller) sendResultBroadcastRsp(req *mc.HD_V2_ReelectBroadcastMsg) {
	resultHash := types.RlpHash(req)
	sign, err := self.matrix.SignHelper().SignHashWithValidateByReader(self.dc, resultHash.Bytes(), true, self.ParentHash())
	if err != nil {
		log.Error(self.logInfo, "响应结果广播消息", "签名失败", "err", err)
		return
	}
	rsp := mc.HD_V2_ReelectBroadcastRspMsg{
		Number:     self.Number(),
		ResultHash: resultHash,
		Sign:       sign,
	}
	self.matrix.HD().SendNodeMsg(mc.HD_V2_LeaderReelectBroadcastRsp, rsp, common.RoleNil, []common.Address{req.From})
}

func (self *controller) checkRLReqMsg(req *mc.HD_V2_ReelectLeaderReqMsg) error {
	if nil == req || nil == req.InquiryReq {
		return ErrParamsIsNil
	}
	if req.InquiryReq.ConsensusTurn != self.dc.curConsensusTurn {
		return errors.Errorf("共识轮次不匹配, 消息(%d) != 本地(%d)", req.InquiryReq.ConsensusTurn, self.dc.curConsensusTurn)
	}
	if req.InquiryReq.ReelectTurn != self.dc.curReelectTurn {
		return errors.Errorf("重选轮次不匹配, 消息(%d) != 本地(%d)", req.InquiryReq.ReelectTurn, self.dc.curReelectTurn)
	}
	if req.InquiryReq.Master != self.dc.GetReelectMaster() {
		return errors.Errorf("master不匹配, master(%s) != 本地master(%s)", req.InquiryReq.Master.Hex(), self.dc.GetReelectMaster().Hex())
	}
	if int64(req.TimeStamp) < int64(req.InquiryReq.TimeStamp) {
		return errors.Errorf("请求时间戳(%d) < 询问时间戳(%d)", int64(req.TimeStamp), int64(req.InquiryReq.TimeStamp))
	}
	if err := self.dc.turnTime.CheckTimeLegal(self.dc.curConsensusTurn.TotalTurns(), self.dc.curReelectTurn, int64(req.TimeStamp)); err != nil {
		return err
	}
	if _, err := self.matrix.DPOSEngine().VerifyHashWithBlock(self.dc, types.RlpHash(req.InquiryReq), req.AgreeSigns, self.ParentHash()); err != nil {
		return errors.Errorf("请求中的询问同意签名POS未通过(%v)", err)
	}

	return nil
}

func (self *controller) checkRLResult(result *mc.HD_V2_ReelectLeaderConsensus) error {
	if nil == result {
		return ErrLeaderResultIsNil
	}
	turn := calcNextConsensusTurn(result.Req.InquiryReq.ConsensusTurn, result.Req.InquiryReq.ReelectTurn)
	if turn.Cmp(self.dc.curConsensusTurn) < 0 {
		return errors.Errorf("消息目标共识轮次(%d) < 本地共识轮次(%d)", turn.String(), self.dc.curConsensusTurn.String())
	}
	if _, err := self.matrix.DPOSEngine().VerifyHashWithBlock(self.dc, types.RlpHash(result.Req), result.Votes, self.ParentHash()); err != nil {
		return errors.Errorf("leader重选完成结果的POS验证失败(%v)", err)
	}
	return nil
}

func (self *controller) checkPOSResult(posResult *mc.HD_BlkConsensusReqMsg) error {
	if nil == posResult {
		return errors.New("POS结果为nil")
	}
	if posResult.Header.Number.Uint64() != self.Number() {
		return errors.Errorf("高度不匹配, pos number[%d] != local number[%d]", posResult.Header.Number.Uint64(), self.Number())
	}
	if posResult.ConsensusTurn != self.dc.curConsensusTurn || posResult.Header.Leader != self.dc.consensusLeader {
		return errors.Errorf("pos结果轮次不匹配, pos轮次[%s] leader[%s], local轮次[%s], leader[%s]",
			posResult.ConsensusTurn.String(), posResult.Header.Leader.Hex(), self.ConsensusTurn().String(), self.dc.GetConsensusLeader().Hex())
	}
	if err := self.matrix.DPOSEngine().VerifyBlock(self.dc, posResult.Header); err != nil {
		return errors.Errorf("POS验证失败(%v)", err)
	}
	return nil
}

func (self *controller) processNewBlockReadyRsp(header *types.Header, from common.Address) {
	if nil == header {
		log.Info(self.logInfo, "处理新区块响应", "区块header为nil")
		return
	}

	number := header.Number.Uint64()
	parentHeader := self.matrix.BlockChain().GetHeader(header.ParentHash, number-1)
	if parentHeader == nil {
		log.Warn(self.logInfo, "处理新区块响应", "没有父区块，进行fetch", "parent number", number-1, "parent hash", header.ParentHash.TerminalString(), "source", from.Hex())
		self.matrix.FetcherNotify(header.ParentHash, number-1, from)
		return
	}

	//POW验证
	bcInterval, err := self.matrix.BlockChain().GetBroadcastIntervalByHash(header.ParentHash)
	if err != nil {
		log.Warn(self.logInfo, "处理新区块响应", "获取广播周期信息失败", "err", err)
		return
	}

	isBroadcast := bcInterval.IsBroadcastNumber(number)
	seal := !isBroadcast
	err = self.matrix.Engine().VerifyHeader(self.matrix.BlockChain(), header, seal)
	if err != nil {
		log.Warn(self.logInfo, "处理新区块响应", "POW验证失败", "高度", number, "verify seal", seal, "block hash", header.Hash().TerminalString(), "err", err)
		return
	}

	//POS验证
	err = self.matrix.DPOSEngine().VerifyBlock(self.dc, header)
	if err != nil {
		log.Warn(self.logInfo, "处理新区块响应", "POS验证失败", "高度", number, "block hash", header.Hash().TerminalString(), "err", err)
		return
	}

	//发送恢复状态消息
	log.Debug(self.logInfo, "处理新区块响应", "发送恢复状态消息", "高度", number, "block hash", header.Hash().TerminalString())
	mc.PublishEvent(mc.Leader_RecoveryState, &mc.RecoveryStateMsg{Type: mc.RecoveryTypeFullHeader, Header: header, From: from, IsBroadcast: isBroadcast})
}
