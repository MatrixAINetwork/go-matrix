// Copyright (c) 2018 The MATRIX Authors 
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
		return "闲置状态"
	case waitingBlockReq:
		return "等待区块验证请求状态"
	case waitingDPOSResult:
		return "等待DPOS结果状态"
	case reelectLeader:
		return "重选状态"
	default:
		return "未知状态"
	}
}

var (
	waitingBlockReqTimer   = 40 * time.Second
	waitingDPOSResultTimer = 40 * time.Second
	reelectLeaderTimer     = 40 * time.Second

	ErrInvalidState = errors.New("不支持的状态")
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

func (self *controller) ReceiveMsg(msg interface{}) {
	self.msgCh <- msg
}

func (self *controller) Number() uint64 {
	return self.dc.number
}

func (self *controller) State() state {
	return self.dc.state
}

func (self *controller) ConsensusTurn() uint32 {
	return self.dc.curConsensusTurn
}

func (self *controller) run() {
	log.INFO(self.logInfo, "控制服务", "启动", "高度", self.dc.number)
	defer log.INFO(self.logInfo, "控制服务", "退出", "高度", self.dc.number)

	self.setTimer(0, self.timer)
	self.setTimer(0, self.reelectTimer)
	for {
		select {
		case msg := <-self.msgCh:
			self.handleMsg(msg)

		case <-self.timer.C:
			self.timeOutHandle()

		case <-self.reelectTimer.C:
			self.reelectTimeOutHandle()

		case <-self.quitCh:
			return
		}
	}
}

func (self *controller) sendLeaderMsg() {
	msg, err := self.dc.PrepareLeaderMsg()
	if err != nil {
		log.ERROR(self.logInfo, "公布leader身份", "准备消息错误", "err", err)
		return
	}
	log.INFO(self.logInfo, "公布leader身份, leader", msg.Leader.Hex(), "Next Leader", msg.NextLeader.Hex(), "高度", msg.Number,
		"共识状态", msg.ConsensusState, "共识轮次", msg.ConsensusTurn, "重选轮次", msg.ReelectTurn)
	mc.PublishEvent(mc.Leader_LeaderChangeNotify, msg)
}

func (self *controller) setTimer(outTime int64, timer *time.Timer) {
	var OK bool
	if outTime == 0 {
		OK = timer.Stop()
	} else {
		OK = timer.Reset(time.Duration(outTime) * time.Second)
	}
	if !OK {
		for {
			select {
			case <-timer.C:
				log.DEBUG(self.logInfo, "超时器处理", "释放无用超时")
			default:
				return
			}
		}
	}
}

func (self *controller) curTurnInfo() string {
	return "共识轮次(" + strconv.Itoa(int(self.dc.curConsensusTurn)) + ")&重选轮次(" + strconv.Itoa(int(self.dc.curReelectTurn)) + ")"
}
