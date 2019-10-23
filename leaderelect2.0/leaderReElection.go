// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect2

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"
	"github.com/pkg/errors"
)

type LeaderIdentity struct {
	ctrlManager       *ControllerManager
	matrix            Matrix
	extraInfo         string
	newBlockReadyCh   chan *mc.NewBlockReadyMsg
	newBlockReadySub  event.Subscription
	roleUpdateCh      chan *mc.RoleUpdatedMsg
	roleUpdateSub     event.Subscription
	blkPOSNotifyCh    chan *mc.BlockPOSFinishedNotify
	blkPOSNotifySub   event.Subscription
	rlInquiryReqCh    chan *mc.HD_V2_ReelectInquiryReqMsg
	rlInquiryReqSub   event.Subscription
	rlInquiryRspCh    chan *mc.HD_V2_ReelectInquiryRspMsg
	rlInquiryRspSub   event.Subscription
	rlReqCh           chan *mc.HD_V2_ReelectLeaderReqMsg
	rlReqSub          event.Subscription
	rlVoteCh          chan *mc.HD_V2_ConsensusVote
	rlVoteSub         event.Subscription
	rlBroadcastCh     chan *mc.HD_V2_ReelectBroadcastMsg
	rlBroadcastSub    event.Subscription
	rlBroadcastRspCh  chan *mc.HD_V2_ReelectBroadcastRspMsg
	rlBroadcastRspSub event.Subscription
}

func NewLeaderIdentityService(matrix Matrix, extraInfo string) (*LeaderIdentity, error) {
	var server = &LeaderIdentity{
		ctrlManager:      NewControllerManager(matrix, extraInfo),
		matrix:           matrix,
		extraInfo:        extraInfo,
		newBlockReadyCh:  make(chan *mc.NewBlockReadyMsg, 1),
		roleUpdateCh:     make(chan *mc.RoleUpdatedMsg, 1),
		blkPOSNotifyCh:   make(chan *mc.BlockPOSFinishedNotify, 1),
		rlInquiryReqCh:   make(chan *mc.HD_V2_ReelectInquiryReqMsg, 1),
		rlInquiryRspCh:   make(chan *mc.HD_V2_ReelectInquiryRspMsg, 1),
		rlReqCh:          make(chan *mc.HD_V2_ReelectLeaderReqMsg, 1),
		rlVoteCh:         make(chan *mc.HD_V2_ConsensusVote, 1),
		rlBroadcastCh:    make(chan *mc.HD_V2_ReelectBroadcastMsg, 1),
		rlBroadcastRspCh: make(chan *mc.HD_V2_ReelectBroadcastRspMsg, 1),
	}

	if err := server.subEvents(); err != nil {
		log.Error(server.extraInfo, "服务创建失败", err)
		return nil, err
	}

	go server.run()

	log.Info(server.extraInfo, "服务创建", "成功")
	return server, nil
}

func (self *LeaderIdentity) subEvents() error {
	//订阅身份变更消息
	var err error
	if self.newBlockReadySub, err = mc.SubscribeEvent(mc.BlockGenor_NewBlockReady, self.newBlockReadyCh); err != nil {
		return errors.Errorf("订阅<new block ready>事件错误(%v)", err)
	}
	if self.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, self.roleUpdateCh); err != nil {
		return errors.Errorf("订阅<CA身份通知>事件错误(%v)", err)
	}
	if self.blkPOSNotifySub, err = mc.SubscribeEvent(mc.BlkVerify_POSFinishedNotify, self.blkPOSNotifyCh); err != nil {
		return errors.Errorf("订阅<POS验证完成>事件错误(%v)", err)
	}
	if self.rlInquiryReqSub, err = mc.SubscribeEvent(mc.HD_V2_LeaderReelectInquiryReq, self.rlInquiryReqCh); err != nil {
		return errors.Errorf("订阅<重选询问请求V2>事件错误(%v)", err)
	}
	if self.rlInquiryRspSub, err = mc.SubscribeEvent(mc.HD_V2_LeaderReelectInquiryRsp, self.rlInquiryRspCh); err != nil {
		return errors.Errorf("订阅<重选询问响应V2>事件错误(%v)", err)
	}
	if self.rlReqSub, err = mc.SubscribeEvent(mc.HD_V2_LeaderReelectReq, self.rlReqCh); err != nil {
		return errors.Errorf("订阅<leader重选请求V2>事件错误(%v)", err)
	}
	if self.rlVoteSub, err = mc.SubscribeEvent(mc.HD_V2_LeaderReelectVote, self.rlVoteCh); err != nil {
		return errors.Errorf("订阅<leader重选投票V2>事件错误(%v)", err)
	}
	if self.rlBroadcastSub, err = mc.SubscribeEvent(mc.HD_V2_LeaderReelectBroadcast, self.rlBroadcastCh); err != nil {
		return errors.Errorf("订阅<重选广播V2>事件错误(%v)", err)
	}
	if self.rlBroadcastRspSub, err = mc.SubscribeEvent(mc.HD_V2_LeaderReelectBroadcastRsp, self.rlBroadcastRspCh); err != nil {
		return errors.Errorf("订阅<重选广播响应V2>事件错误(%v)", err)
	}
	return nil
}

func (self *LeaderIdentity) run() {
	defer func() {
		self.rlBroadcastRspSub.Unsubscribe()
		self.rlBroadcastSub.Unsubscribe()
		self.rlVoteSub.Unsubscribe()
		self.rlReqSub.Unsubscribe()
		self.rlInquiryRspSub.Unsubscribe()
		self.rlInquiryReqSub.Unsubscribe()
		self.blkPOSNotifySub.Unsubscribe()
		self.roleUpdateSub.Unsubscribe()
		self.newBlockReadySub.Unsubscribe()
	}()

	for {
		select {
		case msg := <-self.newBlockReadyCh:
			go self.newBlockReadyBCHandle(msg)
		case msg := <-self.roleUpdateCh:
			go self.roleUpdateMsgHandle(msg)
		case msg := <-self.blkPOSNotifyCh:
			go self.blockPOSFinishedMsgHandle(msg)
		case msg := <-self.rlInquiryReqCh:
			go self.rlInquiryReqHandle(msg)
		case msg := <-self.rlInquiryRspCh:
			go self.rlInquiryRspHandle(msg)
		case msg := <-self.rlReqCh:
			go self.rlReqMsgHandle(msg)
		case msg := <-self.rlVoteCh:
			go self.rlVoteMsgHandle(msg)
		case msg := <-self.rlBroadcastCh:
			go self.rlBroadcastHandle(msg)
		case msg := <-self.rlBroadcastRspCh:
			go self.rlBroadcastRspHandle(msg)
		}
	}
}

func (self *LeaderIdentity) newBlockReadyBCHandle(msg *mc.NewBlockReadyMsg) {
	if msg == nil || msg.Header == nil {
		log.Error(self.extraInfo, "NewBlockReady处理错误", ErrParamsIsNil)
		return
	}

	curNumber := msg.Header.Number.Uint64()
	log.Debug(self.extraInfo, "NewBlockReady消息处理", "开始", "高度", curNumber)

	// 获取超级区块序号
	supBlkState, err := matrixstate.GetSuperBlockCfg(msg.State)
	if err != nil {
		log.Error(self.extraInfo, "NewBlockReady消息处理", "获取超级区块序号失败", "err", err, "高度", curNumber)
		return
	}

	startMsg := &startControllerMsg{
		parentHeader:  msg.Header,
		parentStateDB: msg.State,
	}
	self.ctrlManager.StartController(curNumber+1, supBlkState.Seq, startMsg)
}

func (self *LeaderIdentity) roleUpdateMsgHandle(msg *mc.RoleUpdatedMsg) {
	if msg == nil {
		log.Error(self.extraInfo, "CA身份通知消息处理", ErrParamsIsNil)
		return
	}
	if (msg.Leader == common.Address{}) {
		log.Error(self.extraInfo, "CA身份通知消息处理", ErrMsgAccountIsNull)
		return
	}

	log.Debug(self.extraInfo, "CA身份通知消息处理", "开始", "高度", msg.BlockNum, "身份", msg.Role, "block hash", msg.BlockHash.TerminalString())
	header := self.matrix.BlockChain().GetHeaderByHash(msg.BlockHash)
	if nil == header {
		log.Error(self.extraInfo, "CA身份通知消息处理", "获取区块header失败", "block hash", msg.BlockHash.TerminalString())
		return
	}

	//获取状态树
	parentState, err := self.matrix.BlockChain().StateAt(header.Roots)
	if err != nil {
		log.Error(self.extraInfo, "CA身份通知消息处理", "获取区块状态树失败", "err", err, "高度", msg.BlockNum)
		return
	}

	startMsg := &startControllerMsg{
		parentHeader:  header,
		parentStateDB: parentState,
	}
	self.ctrlManager.StartController(msg.BlockNum+1, msg.SuperSeq, startMsg)
}

func (self *LeaderIdentity) blockPOSFinishedMsgHandle(msg *mc.BlockPOSFinishedNotify) {
	if msg == nil {
		log.Error(self.extraInfo, "区块POS完成消息处理", "错误", "消息不合法", ErrParamsIsNil)
		return
	}
	if (msg.Header.Leader == common.Address{}) {
		log.Error(self.extraInfo, "区块POS完成消息处理", "错误", "消息不合法", ErrMsgAccountIsNull)
		return
	}

	if manversion.VersionCmp(string(msg.Header.Version), manversion.VersionGamma) < 0 {
		log.Trace(self.extraInfo, "区块POS完成消息处理", "版本号不匹配, 不处理消息", "header version", string(msg.Header.Version), "number", msg.Header.Number)
		return
	}

	log.Debug(self.extraInfo, "区块POS完成消息处理", "开始", "高度", msg.Number)
	err := self.ctrlManager.ReceiveMsg(msg.Number, msg)
	if err != nil {
		log.Error(self.extraInfo, "区块POS完成消息处理", "controller接受消息失败", "err", err)
	}
}

func (self *LeaderIdentity) rlInquiryReqHandle(req *mc.HD_V2_ReelectInquiryReqMsg) {
	if req == nil {
		log.Info(self.extraInfo, "重选询问消息", "错误", "消息不合法", ErrParamsIsNil)
		return
	}
	self.ctrlManager.ReceiveMsgByCur(req)
}

func (self *LeaderIdentity) rlInquiryRspHandle(rsp *mc.HD_V2_ReelectInquiryRspMsg) {
	if rsp == nil {
		log.Info(self.extraInfo, "重选询问响应消息", "错误", "消息不合法", ErrParamsIsNil)
		return
	}
	err := self.ctrlManager.ReceiveMsg(rsp.Number, rsp)
	if err != nil {
		log.Info(self.extraInfo, "重选询问响应消息", "controller接受消息失败", "err", err)
	}
}

func (self *LeaderIdentity) rlReqMsgHandle(req *mc.HD_V2_ReelectLeaderReqMsg) {
	if req == nil {
		log.Info(self.extraInfo, "leader重选请求消息", "错误", "消息不合法", ErrParamsIsNil)
		return
	}
	err := self.ctrlManager.ReceiveMsg(req.InquiryReq.Number, req)
	if err != nil {
		log.Info(self.extraInfo, "leader重选请求消息", "controller接受消息失败", "err", err)
	}
}

func (self *LeaderIdentity) rlVoteMsgHandle(req *mc.HD_V2_ConsensusVote) {
	if req == nil {
		log.Info(self.extraInfo, "leader重选投票消息", "错误", "消息不合法", ErrParamsIsNil)
		return
	}
	err := self.ctrlManager.ReceiveMsg(req.Number, req)
	if err != nil {
		log.Info(self.extraInfo, "leader重选投票消息", "controller接受消息失败", "err", err)
	}
}

func (self *LeaderIdentity) rlBroadcastHandle(msg *mc.HD_V2_ReelectBroadcastMsg) {
	if msg == nil {
		log.Info(self.extraInfo, "重选广播消息", "错误", "消息不合法", ErrParamsIsNil)
		return
	}
	err := self.ctrlManager.ReceiveMsg(msg.Number, msg)
	if err != nil {
		log.Info(self.extraInfo, "重选广播消息", "controller接受消息失败", "err", err)
	}
}

func (self *LeaderIdentity) rlBroadcastRspHandle(rsp *mc.HD_V2_ReelectBroadcastRspMsg) {
	if rsp == nil {
		log.Info(self.extraInfo, "重选广播响应消息", "错误", "消息不合法", ErrParamsIsNil)
		return
	}
	err := self.ctrlManager.ReceiveMsg(rsp.Number, rsp)
	if err != nil {
		log.Info(self.extraInfo, "重选广播响应消息", "controller接受消息失败", "err", err)
	}
}
