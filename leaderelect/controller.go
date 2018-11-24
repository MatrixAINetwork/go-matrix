//1543088871.023861
// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"errors"
	"time"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

type state uint8

const (
	idle state = iota
	waitingBlockReq
	waitingDPOSResult
	reelectLeader
)

func (s state) String() string {
	switch s {
	case idle:
		return "in idle"
	case waitingBlockReq:
		return "waitting for block validation request"
	case waitingDPOSResult:
		return "waitting for DPOS result"
	case reelectLeader:
		return "in re-election state"
	default:
		return "unknown state"
	}
}

var (
	waitingBlockReqTimer   = 40 * time.Second
	waitingDPOSResultTimer = 40 * time.Second
	reelectLeaderTimer     = 40 * time.Second

	ErrInvalidState = errors.New("unsupported state")
)

type controller struct {
	matrix                 Matrix
	number                 uint64
	state                  state
	curLeader              common.Address
	curTurns               uint8
	timer                  *time.Timer
	calServer              *leaderCalculator
	slaver                 *ldreSlaver
	master                 *ldreMaster
	blockVerifyStateCh     chan *mc.BlockVerifyStateNotify
	reelectLeaderSuccessCh chan *mc.ReelectLeaderSuccMsg
	quitCh                 chan struct{}
	extra                  string
}

func newController(matrix Matrix, extra string, calServer *leaderCalculator, number uint64) *controller {
	ctrller := &controller{
		matrix:                 matrix,
		number:                 number,
		state:                  idle,
		curLeader:              common.Address{},
		curTurns:               0,
		timer:                  nil,
		calServer:              calServer,
		slaver:                 nil,
		master:                 nil,
		blockVerifyStateCh:     make(chan *mc.BlockVerifyStateNotify, 2),
		reelectLeaderSuccessCh: make(chan *mc.ReelectLeaderSuccMsg, 1),
		quitCh:                 make(chan struct{}),
		extra:                  extra,
	}

	go ctrller.run()
	log.INFO(ctrller.extra, "创建控制", "成功", "高度", ctrller.number)
	return ctrller
}

func (self *controller) Close() {
	close(self.quitCh)
}

func (self *controller) SetBlockVerifyStateMsg(msg *mc.BlockVerifyStateNotify) {
	self.blockVerifyStateCh <- msg
}

func (self *controller) run() {
	log.INFO(self.extra, "控制服务", "启动", "高度", self.number)
	leaders, err := self.calServer.GetLeaderInfo()
	if err != nil {
		log.ERROR(self.extra, "服务启动失败", "通过calServer获取leader信息失败", "err", err)
		return
	}

	if leaders.number != self.number || leaders.turns != 0 {
		log.ERROR(self.extra, "服务启动失败", "leader信息高度或轮次不匹配",
			"leaders.number", leaders.number, "self.number", self.number,
			"leaders.turns", leaders.turns)
		return
	}

	self.curLeader.Set(leaders.leader)
	self.curTurns = leaders.turns
	self.state = waitingBlockReq
	if self.number == 1 && self.curTurns == 0 {
		self.timer = time.NewTimer(waitingBlockReqTimer + 60*time.Second)
	} else {
		self.timer = time.NewTimer(waitingBlockReqTimer)
	}

	for {
		select {
		case msg := <-self.blockVerifyStateCh:
			log.INFO(self.extra, "接收Block验证状态通知消息, 高度", msg.Number, "状态", msg.State, "leader", msg.Leader.Hex())
			self.blockVerifyStateMsgHandle(msg)

		case data := <-self.reelectLeaderSuccessCh:
			log.INFO(self.extra, "接收Leader重选成功消息, 高度", data.Height, "轮次", data.ReelectTurn, "leader", data.Leader.Hex())
			self.reelectLeaderSuccessHandle(data)

		case <-self.timer.C:
			log.INFO(self.extra, "超时事件, 当前状态", self.state.String(), "高度", self.number, "轮次", self.curTurns, "leader", self.curLeader.Hex())
			self.timeoutHandle()

		case <-self.quitCh:
			self.StopSlaver()
			self.StopMaster()
			log.INFO(self.extra, "控制服务", "退出", "高度", self.number)

			return
		}
	}
}

func (self *controller) blockVerifyStateMsgHandle(msg *mc.BlockVerifyStateNotify) {
	if msg.Number != self.number {
		log.ERROR(self.extra, "Block验证状态消息", "高度不匹配", "当前高度", self.number)
		return
	}

	if msg.Leader != self.curLeader {
		log.ERROR(self.extra, "Block验证状态消息", "Leader不匹配", "当前Leader", self.curLeader.Hex())
		return
	}

	// True: begin verify, False: end verify
	if msg.State {
		self.stateChangeToWaitingDPOSResult()
	} else {
		self.stateChangeToIdle()
	}
}

func (self *controller) reelectLeaderSuccessHandle(msg *mc.ReelectLeaderSuccMsg) {
	if msg.Height != self.number {
		log.ERROR(self.extra, "Leader重选成功消息处理错误", "高度不匹配", "当前高度", self.number)
		return
	}

	if msg.ReelectTurn != self.curTurns {
		log.ERROR(self.extra, "Leader重选成功消息处理错误", "轮次不匹配", "当前轮次", self.curTurns)
		return
	}

	if msg.Leader != self.curLeader {
		log.ERROR(self.extra, "Leader重选成功消息处理错误", "leader不匹配", "当前leader", self.curLeader.Hex())
		return
	}

	self.calServer.UpdateCacheByConsensus(msg.Height, msg.ReelectTurn, true)

	self.StopMaster()
	self.StopSlaver()

	self.state = waitingBlockReq
	self.timer = time.NewTimer(waitingBlockReqTimer)
}

func (self *controller) stateChangeToWaitingDPOSResult() {
	if self.state != waitingBlockReq {
		log.ERROR(self.extra, "切换等待DPOS状态错误", "当前状态不正确", "当前状态", self.state.String())
		return
	}

	if self.number == 1 && self.curTurns == 0 {
		self.timer = time.NewTimer(waitingDPOSResultTimer + 60*time.Second)
	} else {
		self.timer = time.NewTimer(waitingDPOSResultTimer)
	}
	self.state = waitingDPOSResult

	return
}

func (self *controller) stateChangeToIdle() {
	if self.state != waitingDPOSResult {
		log.ERROR(self.extra, "切换等待idle状态错误", "当前状态不正确", "当前状态", self.state.String())
		return
	}

	//关闭定时器，不可直接关闭，有可能定时器已经超时
	self.timer = time.NewTimer(waitingBlockReqTimer)
	self.timer.Stop()

	self.state = idle
}

func (self *controller) StartMaster() {
	self.StopMaster()

	msg := mc.LeaderReelectMsg{Leader: self.curLeader, ReelectTurn: self.curTurns, Number: self.number}
	if master, err := self.genMaster(&msg); err != nil {
		log.ERROR(self.extra, "创建Master失败", err)
	} else {
		self.master = master
	}
}

func (self *controller) StopMaster() {
	if self.master == nil {
		return
	}
	self.master.quitCh <- true
	self.master = nil
}

func (self *controller) genMaster(msg *mc.LeaderReelectMsg) (*ldreMaster, error) {
	master, err := newMaster(self.matrix, self.extra+" "+"Master", msg, self.reelectLeaderSuccessCh)
	return master, err
}

func (self *controller) StartSlaver() {
	self.StopSlaver()

	msg := mc.FollowerReelectMsg{Leader: self.curLeader, ReelectTurn: self.curTurns, Number: self.number}
	if slaver, err := self.genSlaver(&msg); err != nil {
		log.ERROR(self.extra, "创建Slaver失败", err)
	} else {
		self.slaver = slaver
	}
}

func (self *controller) StopSlaver() {
	if self.slaver == nil {
		return
	}
	self.slaver.quitCh <- true
	self.slaver = nil
}

func (self *controller) genSlaver(msg *mc.FollowerReelectMsg) (*ldreSlaver, error) {
	slaver, err := newSlaver(self.matrix, self.extra+" "+"Slaver", msg, self.reelectLeaderSuccessCh)
	return slaver, err
}
