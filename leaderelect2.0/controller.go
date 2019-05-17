// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect2

import (
	"time"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"strconv"
)

type controller struct {
	timer        *time.Timer
	reelectTimer *time.Timer
	matrix       Matrix
	dc           *cdc
	mp           *msgPool
	selfCache    *masterCache
	msgCh        chan interface{}
	quitCh       chan struct{}
	logInfo      string
}

func newController(matrix Matrix, logInfo string, number uint64) *controller {
	if number < 1 {
		log.Crit(logInfo, "创建controller失败", "number < 1", "number", number)
	}
	ctrller := &controller{
		timer:        time.NewTimer(time.Minute),
		reelectTimer: time.NewTimer(time.Minute),
		matrix:       matrix,
		dc:           newCDC(number, matrix.BlockChain(), logInfo),
		mp:           newMsgPool(),
		selfCache:    newMasterCache(number),
		msgCh:        make(chan interface{}, 10),
		quitCh:       make(chan struct{}),
		logInfo:      logInfo,
	}

	go ctrller.run()
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

func (self *controller) State() stateDef {
	return self.dc.state
}

func (self *controller) ConsensusTurn() *mc.ConsensusTurnInfo {
	return &self.dc.curConsensusTurn
}

func (self *controller) ParentHash() common.Hash {
	return self.dc.leaderCal.preHash
}

func (self *controller) run() {
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

func (self *controller) publishLeaderMsg() {
	msg, err := self.dc.PrepareLeaderMsg()
	if err != nil {
		log.ERROR(self.logInfo, "公布leader身份消息", "准备消息失败", "err", err)
		return
	}
	log.Debug(self.logInfo, "公布leader身份消息, leader", msg.Leader.Hex(), "高度", msg.Number,
		"共识状态", msg.ConsensusState, "共识轮次", msg.ConsensusTurn.String(), "重选轮次", msg.ReelectTurn,
		"pre Leader", msg.PreLeader.Hex(), "Next Leader", msg.NextLeader.Hex())
	mc.PublishEvent(mc.Leader_LeaderChangeNotify, msg)
}

func (self *controller) setTimer(outTime int64, timer *time.Timer) {
	var OK bool
	if outTime <= 0 {
		OK = timer.Stop()
	} else {
		OK = timer.Reset(time.Duration(outTime) * time.Second)
	}
	if !OK {
		for {
			select {
			case <-timer.C:
				log.Trace(self.logInfo, "超时器处理", "释放无用超时")
			default:
				return
			}
		}
	}
}

func (self *controller) curTurnInfo() string {
	return "共识轮次(" + self.dc.curConsensusTurn.String() + ")&重选轮次(" + strconv.Itoa(int(self.dc.curReelectTurn)) + ")"
}
