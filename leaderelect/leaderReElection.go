// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

type LeaderIdentity struct {
	ctrlManager      *ControllerManager
	matrix           Matrix
	extraInfo        string
	newBlockReadyCh  chan *mc.NewBlockReadyMsg
	newBlockReadySub event.Subscription
	roleUpdateCh     chan *mc.RoleUpdatedMsg
	roleUpdateSub    event.Subscription
	blkPOSNotifyCh   chan *mc.BlockPOSFinishedNotify
	blkPOSNotifySub  event.Subscription
	rlInquiryReqCh   chan *mc.HD_ReelectInquiryReqMsg
	rlInquiryReqSub  event.Subscription
	rlInquiryRspCh   chan *mc.HD_ReelectInquiryRspMsg
	rlInquiryRspSub  event.Subscription
	rlReqCh          chan *mc.HD_ReelectLeaderReqMsg
	rlReqSub         event.Subscription
	rlVoteCh         chan *mc.HD_ReelectLeaderVoteMsg
	rlVoteSub        event.Subscription
	rlResultBCCh     chan *mc.HD_ReelectResultBroadcastMsg
	rlResultBCSub    event.Subscription
	rlResultRspCh    chan *mc.HD_ReelectResultRspMsg
	rlResultRspSub   event.Subscription
}

func NewLeaderIdentityService(matrix Matrix, extraInfo string) (*LeaderIdentity, error) {
	var server = &LeaderIdentity{
		ctrlManager:     NewControllerManager(matrix, extraInfo),
		matrix:          matrix,
		extraInfo:       extraInfo,
		newBlockReadyCh: make(chan *mc.NewBlockReadyMsg, 1),
		roleUpdateCh:    make(chan *mc.RoleUpdatedMsg, 1),
		blkPOSNotifyCh:  make(chan *mc.BlockPOSFinishedNotify, 1),
		rlInquiryReqCh:  make(chan *mc.HD_ReelectInquiryReqMsg, 1),
		rlInquiryRspCh:  make(chan *mc.HD_ReelectInquiryRspMsg, 1),
		rlReqCh:         make(chan *mc.HD_ReelectLeaderReqMsg, 1),
		rlVoteCh:        make(chan *mc.HD_ReelectLeaderVoteMsg, 1),
		rlResultBCCh:    make(chan *mc.HD_ReelectResultBroadcastMsg, 1),
		rlResultRspCh:   make(chan *mc.HD_ReelectResultRspMsg, 1),
	}

	if err := server.subEvents(); err != nil {
		log.ERROR(server.extraInfo, "服务创建失败", err)
		return nil, err
	}

	go server.run()

	log.INFO(server.extraInfo, "服务创建", "成功")
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
	if self.rlInquiryReqSub, err = mc.SubscribeEvent(mc.HD_LeaderReelectInquiryReq, self.rlInquiryReqCh); err != nil {
		return errors.Errorf("订阅<重选询问请求>事件错误(%v)", err)
	}
	if self.rlInquiryRspSub, err = mc.SubscribeEvent(mc.HD_LeaderReelectInquiryRsp, self.rlInquiryRspCh); err != nil {
		return errors.Errorf("订阅<重选询问响应>事件错误(%v)", err)
	}
	if self.rlReqSub, err = mc.SubscribeEvent(mc.HD_LeaderReelectReq, self.rlReqCh); err != nil {
		return errors.Errorf("订阅<leader重选请求>事件错误(%v)", err)
	}
	if self.rlVoteSub, err = mc.SubscribeEvent(mc.HD_LeaderReelectVote, self.rlVoteCh); err != nil {
		return errors.Errorf("订阅<leader重选投票>事件错误(%v)", err)
	}
	if self.rlResultBCSub, err = mc.SubscribeEvent(mc.HD_LeaderReelectResultBroadcast, self.rlResultBCCh); err != nil {
		return errors.Errorf("订阅<重选结果广播>事件错误(%v)", err)
	}
	if self.rlResultRspSub, err = mc.SubscribeEvent(mc.HD_LeaderReelectResultBroadcastRsp, self.rlResultRspCh); err != nil {
		return errors.Errorf("订阅<重选结果响应>事件错误(%v)", err)
	}
	return nil
}

func (self *LeaderIdentity) run() {
	defer func() {
		self.rlResultRspSub.Unsubscribe()
		self.rlResultBCSub.Unsubscribe()
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
		case msg := <-self.rlResultBCCh:
			go self.rlResultBroadcastHandle(msg)
		case msg := <-self.rlResultRspCh:
			go self.rlResultRspHandle(msg)
		}
	}
}

func (self *LeaderIdentity) newBlockReadyBCHandle(msg *mc.NewBlockReadyMsg) {
	if msg == nil || msg.Header == nil {
		log.ERROR(self.extraInfo, "NewBlockReady处理错误", ErrMsgIsNil)
		return
	}

	validators, err := self.genValidatorList(msg.Header)
	if err != nil {
		log.ERROR(self.extraInfo, "NewBlockReady处理", "获取验证者列表错误", "err", err)
		return
	}

	// 获取自己的身份
	curNumber := msg.Header.Number.Uint64()
	role := self.getRoleFromTopology(validators)
	log.INFO(self.extraInfo, "NewBlockReady消息处理", "开始", "高度", curNumber, "身份", role.String())

	for i, v := range validators.NodeList {
		log.INFO(self.extraInfo, "NewBlockReady消息处理", "消息[验证者列表]", "index", i, "addr", v.Account.Hex(), "pos", v.Position, "身份", v.Type.String())
	}

	startMsg := &startControllerMsg{
		role:         role,
		validators:   validators.NodeList,
		parentHeader: msg.Header,
	}
	self.ctrlManager.StartController(curNumber+1, startMsg)
	log.INFO(self.extraInfo, "NewBlockReady消息处理", "完成")
}

func (self *LeaderIdentity) roleUpdateMsgHandle(msg *mc.RoleUpdatedMsg) {
	if msg == nil {
		log.ERROR(self.extraInfo, "CA身份通知消息处理错误", ErrMsgIsNil)
		return
	}
	if (msg.Leader == common.Address{}) {
		log.ERROR(self.extraInfo, "CA身份通知消息处理错误", ErrMsgAccountIsNull)
		return
	}

	header := self.matrix.BlockChain().GetHeaderByHash(msg.BlockHash)
	if nil == header {
		log.ERROR(self.extraInfo, "CA身份通知消息处理错误", "获取header错误", "block hash", msg.BlockHash.TerminalString())
		return
	}

	//获取拓扑图
	validators, err := ca.GetTopologyByHash(common.RoleValidator, msg.BlockHash)
	if err != nil {
		log.ERROR(self.extraInfo, "CA身份通知消息处理错误", "获取验证者拓扑图错误", "err", err, "高度", msg.BlockNum)
		return
	}

	log.INFO(self.extraInfo, "CA身份通知消息处理", "开始", "高度", msg.BlockNum, "身份", msg.Role.String())
	defer log.INFO(self.extraInfo, "CA身份通知消息处理", "结束", "高度", msg.BlockNum, "身份", msg.Role.String())
	for i, v := range validators.NodeList {
		log.INFO(self.extraInfo, "CA身份通知消息处理", "拓扑图[验证者列表]", "高度", msg.BlockNum, "index", i, "pos", v.Position, "addr", v.Account.Hex(), "股权", v.Stock, "node number", v.NodeNumber)
	}

	startMsg := &startControllerMsg{
		role:         msg.Role,
		validators:   validators.NodeList,
		parentHeader: header,
	}
	self.ctrlManager.StartController(msg.BlockNum+1, startMsg)
}

func (self *LeaderIdentity) blockPOSFinishedMsgHandle(msg *mc.BlockPOSFinishedNotify) {
	if msg == nil {
		log.Error(self.extraInfo, "区块POS完成消息处理", "错误", "消息不合法", ErrMsgIsNil)
		return
	}
	if (msg.Header.Leader == common.Address{}) {
		log.ERROR(self.extraInfo, "区块POS完成消息处理", "错误", "消息不合法", ErrMsgAccountIsNull)
		return
	}

	log.INFO(self.extraInfo, "收到区块POS完成消息", "开始", "高度", msg.Number)
	ctrl, err := self.ctrlManager.GetController(msg.Number)
	if err != nil {
		log.ERROR(self.extraInfo, "区块POS完成消息处理", "获取controller失败", "err", err)
		return
	}
	ctrl.ReceiveMsg(msg)
}

func (self *LeaderIdentity) rlInquiryReqHandle(req *mc.HD_ReelectInquiryReqMsg) {
	if req == nil {
		log.Error(self.extraInfo, "重选询问消息", "错误", "消息不合法", ErrMsgIsNil)
		return
	}
	log.INFO(self.extraInfo, "收到重选询问消息", "开始", "高度", req.Number, "共识轮次", req.ConsensusTurn, "重选轮次", req.ReelectTurn, "from", req.From.Hex())

	curCtrl := self.ctrlManager.GetCurController()
	if curCtrl.Number() > req.Number {
		reqHash := types.RlpHash(req)
		log.INFO(self.extraInfo, "重选询问消息处理", "高度低于当前高度，发送新区块已准备完毕响应", "高度", req.Number, "本地高度", curCtrl.Number(), "reqHash", reqHash.TerminalString())
		curCtrl.ReceiveMsg(&sendNewBlockReadyRsp{repHash: reqHash, target: req.From, rspNumber: req.Number})
		return
	}

	ctrl, err := self.ctrlManager.GetController(req.Number)
	if err != nil {
		log.ERROR(self.extraInfo, "重选询问消息处理", "获取controller失败", "err", err)
		return
	}
	ctrl.ReceiveMsg(req)
}

func (self *LeaderIdentity) rlInquiryRspHandle(rsp *mc.HD_ReelectInquiryRspMsg) {
	if rsp == nil {
		log.Error(self.extraInfo, "重选询问响应", "错误", "消息不合法", ErrMsgIsNil)
		return
	}
	ctrl, err := self.ctrlManager.GetController(rsp.Number)
	if err != nil {
		log.ERROR(self.extraInfo, "重选询问消息响应", "获取controller失败", "err", err)
		return
	}
	ctrl.ReceiveMsg(rsp)
}

func (self *LeaderIdentity) rlReqMsgHandle(req *mc.HD_ReelectLeaderReqMsg) {
	if req == nil {
		log.Error(self.extraInfo, "leader重选请求", "错误", "消息不合法", ErrMsgIsNil)
		return
	}
	ctrl, err := self.ctrlManager.GetController(req.InquiryReq.Number)
	if err != nil {
		log.ERROR(self.extraInfo, "leader重选请求处理", "获取controller失败", "err", err)
		return
	}
	ctrl.ReceiveMsg(req)
}

func (self *LeaderIdentity) rlVoteMsgHandle(req *mc.HD_ReelectLeaderVoteMsg) {
	if req == nil {
		log.Error(self.extraInfo, "leader重选投票", "错误", "消息不合法", ErrMsgIsNil)
		return
	}
	log.INFO(self.extraInfo, "收到leader重选投票", "开始", "高度", req.Number)

	ctrl, err := self.ctrlManager.GetController(req.Number)
	if err != nil {
		log.ERROR(self.extraInfo, "leader重选投票处理", "获取controller失败", "err", err)
		return
	}
	ctrl.ReceiveMsg(req)
}

func (self *LeaderIdentity) rlResultBroadcastHandle(msg *mc.HD_ReelectResultBroadcastMsg) {
	if msg == nil {
		log.Error(self.extraInfo, "重选结果广播", "错误", "消息不合法", ErrMsgIsNil)
		return
	}
	log.INFO(self.extraInfo, "收到重选结果广播", "开始", "高度", msg.Number, "结果类型", msg.Type, "from", msg.From.Hex())

	ctrl, err := self.ctrlManager.GetController(msg.Number)
	if err != nil {
		log.ERROR(self.extraInfo, "重选结果广播处理", "获取controller失败", "err", err)
		return
	}
	ctrl.ReceiveMsg(msg)
}

func (self *LeaderIdentity) rlResultRspHandle(rsp *mc.HD_ReelectResultRspMsg) {
	if rsp == nil {
		log.Error(self.extraInfo, "重选结果响应", "错误", "消息不合法", ErrMsgIsNil)
		return
	}
	log.INFO(self.extraInfo, "收到重选结果响应", "开始", "高度", rsp.Number)

	ctrl, err := self.ctrlManager.GetController(rsp.Number)
	if err != nil {
		log.ERROR(self.extraInfo, "重选结果响应处理", "获取controller失败", "err", err)
		return
	}
	ctrl.ReceiveMsg(rsp)
}

func (self *LeaderIdentity) genValidatorList(header *types.Header) (*mc.TopologyGraph, error) {
	newGraph, err := self.matrix.BlockChain().NewTopologyGraph(header)
	if err != nil {
		return nil, errors.Errorf("创建新拓扑图失败:%v", err)
	}

	validators := &mc.TopologyGraph{
		Number:   newGraph.Number,
		NodeList: make([]mc.TopologyNodeInfo, 0),
	}

	for _, node := range newGraph.NodeList {
		if node.Type == common.RoleValidator {
			validators.NodeList = append(validators.NodeList, node)
		}
	}
	return validators, nil
}

func (self *LeaderIdentity) getRoleFromTopology(validators *mc.TopologyGraph) common.RoleType {
	selfAccount := ca.GetAddress()
	for _, v := range validators.NodeList {
		if v.Account == selfAccount {
			return v.Type
		}
	}
	return common.RoleNil
}
