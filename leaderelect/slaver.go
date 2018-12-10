// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

type ldreSlaver struct {
	voteReqMsgCh        chan *mc.HD_LeaderReelectVoteReqMsg
	voteReqMsgSub       event.Subscription
	voteConsensusMsgCh  chan *mc.HD_LeaderReelectConsensusBroadcastMsg
	voteConsensusMsgSub event.Subscription
	quitCh              chan bool
	resultCh            chan *mc.ReelectLeaderSuccMsg
	matrix              Matrix
	ce                  consensus.DPOSEngine
	signHelper          *signhelper.SignHelper
	extra               string
	consensusState      bool
}

var ErrSlaverSubVoteReqMsg = errors.New("订阅LDRE REQ事件错误")
var ErrSlaverSubConsensusMsg = errors.New("订阅LDRE 共识消息事件错误")
var ErrSlaverConsensusRep = errors.New("Slaver已经完成共识")
var ErrSlaverConsensusFrom = errors.New("LDRE 共识消息发起者错误")
var ErrSlaverConsensusTurns = errors.New("LDRE 共识请求轮次错误")
var ErrSlaverConsensusHeigth = errors.New("LDRE 共识消息发起者错误")
var ErrSlaverConsensusSign = errors.New("LDRE 共识请求验签错误")
var ErrSlaverMsgFromErr = errors.New("消息源不正确")

func newSlaver(matrix Matrix, extra string, msg *mc.FollowerReelectMsg, resultCh chan *mc.ReelectLeaderSuccMsg) (*ldreSlaver, error) {
	var err error

	slaver := &ldreSlaver{
		quitCh:             make(chan bool, 1),
		resultCh:           resultCh,
		voteReqMsgCh:       make(chan *mc.HD_LeaderReelectVoteReqMsg, 1),
		voteConsensusMsgCh: make(chan *mc.HD_LeaderReelectConsensusBroadcastMsg, 1),
		matrix:             matrix,
		ce:                 matrix.DPOSEngine(),
		signHelper:         matrix.SignHelper(),
		consensusState:     false,
		extra:              extra,
	}

	//订阅LDRE REQ消息事件
	if slaver.voteReqMsgSub, err = mc.SubscribeEvent(mc.HD_LeaderReelectVoteReq, slaver.voteReqMsgCh); err != nil {
		return nil, ErrSlaverSubVoteReqMsg
	}

	//订阅LDRE 共识消息事件
	if slaver.voteConsensusMsgSub, err = mc.SubscribeEvent(mc.HD_LeaderReelectConsensusBroadcast, slaver.voteConsensusMsgCh); err != nil {
		slaver.voteReqMsgSub.Unsubscribe()
		return nil, ErrSlaverSubConsensusMsg
	}

	go slaver.run(msg.Number, msg.ReelectTurn, msg.Leader)

	return slaver, nil
}

func (self *ldreSlaver) broadcastReelectLeaderVote(vote *mc.HD_ConsensusVote) {
	log.INFO(self.extra, "发出LDRE投票RSP, Self", ca.GetAddress(), "轮次", vote.Round)
	self.matrix.HD().SendNodeMsg(mc.HD_LeaderReelectVoteRsp, vote, common.RoleValidator, nil)

}

func (self *ldreSlaver) sendReelectLeaderSuccessMsg(masterAddress common.Address, turnNum uint8, height uint64) {
	msg := &mc.ReelectLeaderSuccMsg{
		Leader:      masterAddress,
		ReelectTurn: turnNum,
		Height:      height,
	}
	self.resultCh <- msg
	log.INFO(self.extra, "发布LDRE重选成功到LD控制模块", "")
}

func (self *ldreSlaver) run(height uint64, turnNum uint8, masterAddress common.Address) {
	log.INFO(self.extra, "服务", "启动", "高度", height, "轮次", turnNum)
	defer func() {
		self.voteReqMsgSub.Unsubscribe()
		self.voteConsensusMsgSub.Unsubscribe()
		log.INFO(self.extra, "服务", "退出", "高度", height, "轮次", turnNum)
	}()

	for {
		select {
		case msg := <-self.voteReqMsgCh:
			log.INFO(self.extra, "收到LDRE投票请求, Leader", msg.Leader.Hex(), "预期", masterAddress.Hex(), "当前高度", height, "当前轮次", turnNum)
			self.reelectLeaderRequestHandle(msg, masterAddress, height, turnNum)

		case msg := <-self.voteConsensusMsgCh:
			log.INFO(self.extra, "收到LDRE共识消息， Leader", msg.Req.Leader.Hex(), "当前高度", height, "当前轮次", turnNum)
			if err := self.reelectLeaderResultHandle(msg, masterAddress, height, turnNum); err != nil {
				log.ERROR(self.extra, "处理LDRE共识消息， 失败", err)
			} else {
				log.INFO(self.extra, "处理LDRE共识消息， 成功", "")
			}

		case <-self.quitCh:
			log.INFO(self.extra, "退出Slaver", "")
			return
		}
	}
}

func (self *ldreSlaver) reelectLeaderRequestHandle(msg *mc.HD_LeaderReelectVoteReqMsg, masterAddress common.Address, height uint64, turnNum uint8) error {
	var hash common.Hash
	var err error

	if msg.Leader != masterAddress {
		log.ERROR(self.extra, "LDRE投票请求 FROM ERR", "不符合预期")
		return ErrSlaverMsgFromErr
	}

	if msg.Height != height {
		log.ERROR(self.extra, "LDRE投票请求 FROM ERR", "高度不匹配")
		return ErrSlaverMsgFromErr
	}

	if msg.ReelectTurn != turnNum {
		log.ERROR(self.extra, "LDRE投票请求 FROM ERR", "轮次不匹配")
		return ErrSlaverMsgFromErr
	}

	hash = types.RlpHash(msg)
	sign, err := self.signHelper.SignHashWithValidate(hash.Bytes(), true)
	if err != nil {
		log.ERROR(self.extra, "投票签名错误", err)
		return err
	}

	vote := &mc.HD_ConsensusVote{
		SignHash: hash,
		Sign:     sign,
		Round:    uint64(msg.ReelectTurn),
	}
	self.broadcastReelectLeaderVote(vote)
	return nil
}

func (self *ldreSlaver) reelectLeaderResultHandle(msg *mc.HD_LeaderReelectConsensusBroadcastMsg, masterAddress common.Address, height uint64, turnNum uint8) error {
	if self.consensusState == true {
		return ErrSlaverConsensusRep
	}

	if !masterAddress.Equal(msg.Req.Leader) {
		log.ERROR(self.extra, "LDRE共识,消息发起者错误, From", msg.Req.Leader.Hex(), "期望", masterAddress.Hex())
		return ErrSlaverConsensusFrom
	}
	if msg.Req.Height != height {
		log.ERROR(self.extra, "LDRE共识,消息高度错误, msg Height", msg.Req.Height, "期望", height)
		return ErrSlaverConsensusHeigth
	}
	if msg.Req.ReelectTurn != turnNum {
		log.ERROR(self.extra, "LDRE共识请求轮次错误，msg Turn", msg.Req.ReelectTurn, "期望", turnNum)
		return ErrSlaverConsensusTurns
	}

	reqHash := types.RlpHash(msg.Req)
	if _, err := self.ce.VerifyHashWithNumber(reqHash, msg.Signatures, height-1); err != nil {
		log.ERROR(self.extra, "LDRE共识结果验签错误", err)
		return ErrSlaverConsensusSign
	}

	// 发送重选leader成功消息
	log.INFO(self.extra, "发布LDRE成功", "至LD控制模块")
	self.sendReelectLeaderSuccessMsg(masterAddress, turnNum, height)
	return nil
}
