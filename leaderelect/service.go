// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"sync"
	"time"

	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/msgsend"
	"github.com/matrix/go-matrix/p2p"
	"github.com/matrix/go-matrix/params"
	"github.com/pkg/errors"
)

type Matrix interface {
	BlockChain() *core.BlockChain
	SignHelper() *signhelper.SignHelper
	DPOSEngine() consensus.DPOSEngine
	HD() *msgsend.HD
}

type LeaderIdentity struct {
	serviceName            string
	calServer              *leaderCalculator //leader身份计算服务
	controller             *controller       //重选leader控制服务
	blkVerifyStateMsgCache []*mc.BlockVerifyStateNotify
	controllerLock         sync.Mutex
	extraInfo              string
	matrix                 Matrix
	newBlockReadyCh        chan *mc.NewBlockReady
	newBlockReadySub       event.Subscription
	roleUpdateCh           chan *mc.RoleUpdatedMsg
	roleUpdateSub          event.Subscription
	blkVerifyNotifyCh      chan *mc.BlockVerifyStateNotify
	blkVerifyNotifySub     event.Subscription
}

var ErrMsgPtrIsNull = errors.New("消息指针为空指针")
var ErrMsgAccountIsNull = errors.New("不合法的账户：空账户")

func NewLeaderIdentityService(matrix Matrix, extraInfo string) (*LeaderIdentity, error) {
	var err error

	var server = &LeaderIdentity{
		newBlockReadyCh:        make(chan *mc.NewBlockReady, 1),
		roleUpdateCh:           make(chan *mc.RoleUpdatedMsg, 1),
		blkVerifyNotifyCh:      make(chan *mc.BlockVerifyStateNotify, 1),
		calServer:              newLeaderCal(matrix, extraInfo+" LD计算"),
		controller:             nil,
		blkVerifyStateMsgCache: make([]*mc.BlockVerifyStateNotify, 0),
		matrix:                 matrix,
		extraInfo:              extraInfo,
	}

	//订阅身份变更消息
	if server.newBlockReadySub, err = mc.SubscribeEvent(mc.BlockGenor_NewBlockReady, server.newBlockReadyCh); err != nil {
		log.Error(server.extraInfo, "订阅new block ready 事件错误", err)
	}

	if server.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, server.roleUpdateCh); err != nil {
		log.Error(server.extraInfo, "订阅CA身份通知事件错误", err)
		server.newBlockReadySub.Unsubscribe()
		return nil, err
	}

	if server.blkVerifyNotifySub, err = mc.SubscribeEvent(mc.BlkVerify_VerifyStateNotify, server.blkVerifyNotifyCh); err != nil {
		log.Error(server.extraInfo, "订阅区块验证状态通知事件错误", err)
		server.roleUpdateSub.Unsubscribe()
		server.newBlockReadySub.Unsubscribe()
		return nil, err
	}

	go server.run()
	log.INFO(server.extraInfo, "服务创建成功", "")

	return server, nil
}

func (self *LeaderIdentity) run() {
	defer func() {
		self.newBlockReadySub.Unsubscribe()
		self.blkVerifyNotifySub.Unsubscribe()
		self.roleUpdateSub.Unsubscribe()
	}()

	for {
		select {
		case msg := <-self.newBlockReadyCh:
			go self.newBlockReadyMsgHandle(msg)

		case msg := <-self.roleUpdateCh:
			go self.roleUpdateMsgHandle(msg)

		case msg := <-self.blkVerifyNotifyCh:
			go self.blockVerifyMsgHandle(msg)
		}
	}
}

func (self *LeaderIdentity) newBlockReadyMsgHandle(msg *mc.NewBlockReady) {
	if msg == nil {
		log.ERROR(self.extraInfo, "NewBlockReady消息处理错误", ErrMsgPtrIsNull)
		return
	}
	if (msg.Leader == common.Address{}) {
		log.ERROR(self.extraInfo, "NewBlockReady消息处理错误", ErrMsgAccountIsNull)
		return
	}

	// 获取自己的身份
	role := self.getRoleFromTopology(msg.Validators)
	log.INFO(self.extraInfo, "NewBlockReady消息处理", "开始", "高度", msg.Number, "身份", role.String())

	for i, v := range msg.Validators.NodeList {
		log.INFO(self.extraInfo, "NewBlockReady消息处理", "消息[验证者列表]", "index", i, "addr", v.Account.Hex(), "pos", v.Position, "身份", v.Type.String())
	}
	// 维护controller状态
	self.maintainController(role, msg.Number, msg.Leader, msg.Validators)
	log.INFO(self.extraInfo, "NewBlockReady消息处理", "完成")
}

func (self *LeaderIdentity) roleUpdateMsgHandle(msg *mc.RoleUpdatedMsg) {
	if msg == nil {
		log.ERROR(self.extraInfo, "CA身份通知消息处理错误", ErrMsgPtrIsNull)
		return
	}
	if (msg.Leader == common.Address{}) {
		log.ERROR(self.extraInfo, "CA身份通知消息处理错误", ErrMsgAccountIsNull)
		return
	}

	//启动保护，暂定15秒，等待上层服务启动
	if msg.BlockNum == 0 && msg.Role == common.RoleValidator {
		targetCount := self.getBootPeerCount()
		for {
			peersCount := p2p.ServerP2p.PeerCount()
			if peersCount >= targetCount {
				log.INFO(self.extraInfo, "peer数量满足条件", "完成boot", "peer数量", peersCount, "目标数量", targetCount)
				break
			}
			if self.matrix.BlockChain().CurrentBlock().Number().Uint64() >= 1 {
				log.WARN(self.extraInfo, "当前高度已大于0", "退出boot流程")
				return
			}
			time.Sleep(5 * time.Second)
			log.INFO("Peer 数量", "len", peersCount, "目标", targetCount)
		}
		time.Sleep(60 * time.Second)
	}

	//获取拓扑图
	validators, err := ca.GetTopologyByNumber(common.RoleValidator, msg.BlockNum)
	if err != nil {
		log.ERROR(self.extraInfo, "CA身份通知消息处理错误", "获取验证者拓扑图错误", "err", err, "高度", msg.BlockNum)
		return
	}

	log.INFO(self.extraInfo, "CA身份通知消息处理", "开始", "高度", msg.BlockNum, "身份", msg.Role.String())
	defer log.INFO(self.extraInfo, "CA身份通知消息处理", "结束", "高度", msg.BlockNum, "身份", msg.Role.String())
	for i, v := range validators.NodeList {
		log.INFO(self.extraInfo, "CA身份通知消息处理", "拓扑图[验证者列表]", "index", i, "pos", v.Position, "addr", v.Account.Hex(), "股权", v.Stock)
	}

	// 维护controller状态
	self.maintainController(msg.Role, msg.BlockNum, msg.Leader, validators)
	return
}

func (self *LeaderIdentity) blockVerifyMsgHandle(msg *mc.BlockVerifyStateNotify) {
	if msg == nil {
		log.Error(self.extraInfo, "区块验证状态处理", "错误", "请求不合法", ErrMsgPtrIsNull)
		return
	}
	if (msg.Leader == common.Address{}) {
		log.ERROR(self.extraInfo, "区块验证状态处理", "错误", "请求不合法", ErrMsgAccountIsNull)
		return
	}

	log.INFO(self.extraInfo, "区块验证状态处理", "开始", "高度", msg.Number)

	self.controllerLock.Lock()
	defer self.controllerLock.Unlock()
	number := self.calServer.GetCurNumber()
	if msg.Number == number+1 {
		log.INFO(self.extraInfo, "区块验证状态处理", "缓存", "当前高度", number, "消息高度", msg.Number)
		self.blkVerifyStateMsgCache = append(self.blkVerifyStateMsgCache, msg)
		return
	} else if msg.Number == number {
		if self.controller == nil {
			log.ERROR(self.extraInfo, "区块验证状态处理", "完成", "controller服务未启动", "抛弃消息")
			return
		}
		self.controller.SetBlockVerifyStateMsg(msg)
	} else {
		log.ERROR(self.extraInfo, "区块验证状态处理", "完成", "高度不合法", "抛弃消息")
	}

	log.INFO(self.extraInfo, "区块验证状态处理", "完成", "高度", msg.Number)
}

func (self *LeaderIdentity) maintainController(role common.RoleType, number uint64, leader common.Address, validators *mc.TopologyGraph) {
	self.controllerLock.Lock()
	defer self.controllerLock.Unlock()
	curNumber := self.calServer.GetCurNumber()
	dealNumber := number + 1
	if curNumber > dealNumber {
		log.WARN(self.extraInfo, "controller服务", "不处理", "处理高度过低, 请求高度(输入高度+1)", dealNumber, "current number", curNumber)
		return
	}

	if role != common.RoleValidator {
		self.stopController()
		self.clearMsgCache()
		log.INFO(self.extraInfo, "controller服务", "关闭", "高度", number, "身份", role.String())
		return
	}

	self.calServer.UpdateCacheByHeader(number, leader, validators)
	self.calServer.NotifyLeaderChange()
	if self.controller == nil || self.controller.number != dealNumber {
		if common.IsBroadcastNumber(dealNumber) {
			self.stopController()
		} else {
			self.startController(dealNumber)
			self.dealCacheMsg()
		}
		self.clearMsgCache()
		log.INFO(self.extraInfo, "controller服务", "开启", "处理高度", dealNumber)
	} else {
		log.INFO(self.extraInfo, "controller服务", "不处理", "处理高度相同", dealNumber)
	}
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

func (self *LeaderIdentity) startController(height uint64) {
	if self.controller != nil {
		self.stopController()
	}

	self.controller = newController(self.matrix, self.extraInfo+"控制模块", self.calServer, height)
}

func (self *LeaderIdentity) stopController() {
	if self.controller != nil {
		self.controller.Close()
		self.controller = nil
	}
}

func (self *LeaderIdentity) clearMsgCache() {
	if len(self.blkVerifyStateMsgCache) > 0 {
		self.blkVerifyStateMsgCache = make([]*mc.BlockVerifyStateNotify, 0)
	}
}

func (self *LeaderIdentity) dealCacheMsg() {
	if len(self.blkVerifyStateMsgCache) <= 0 || self.controller != nil {
		return
	}

	for _, msg := range self.blkVerifyStateMsgCache {
		self.controller.SetBlockVerifyStateMsg(msg)
	}

	self.clearMsgCache()
}

func (self *LeaderIdentity) getBootPeerCount() int {
	var nodeCount, bCount int
	nodes, err := ca.GetTopologyByNumber(common.RoleValidator|common.RoleMiner, 0)
	if err != nil {
		nodeCount = 0
	} else {
		nodeCount = len(nodes.NodeList)
	}

	bCount = len(params.BroadCastNodes)
	return nodeCount + bCount - 1
}
