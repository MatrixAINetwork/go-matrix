// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
)

type leaderData struct {
	number         uint64
	turns          uint8
	consensusState bool
	leader         common.Address
	nextLeader     common.Address
}

func (self *leaderData) copyData() *leaderData {
	newData := &leaderData{
		number:         self.number,
		turns:          self.turns,
		consensusState: self.consensusState,
		leader:         common.Address{},
		nextLeader:     common.Address{},
	}

	newData.leader.Set(self.leader)
	newData.nextLeader.Set(self.nextLeader)
	return newData
}

type leaderCalculator struct {
	chain      *core.BlockChain
	curNumber  uint64
	preLeader  common.Address
	turns      uint8
	validators *mc.TopologyGraph
	leaders    leaderData
	mu         sync.Mutex
	extra      string
}

func newLeaderCal(matrix Matrix, extra string) *leaderCalculator {
	return &leaderCalculator{
		chain:      matrix.BlockChain(),
		curNumber:  0,
		preLeader:  common.Address{},
		turns:      0,
		validators: nil,
		extra:      extra,
		leaders:    leaderData{number: 0, turns: 0, leader: common.Address{}, nextLeader: common.Address{}, consensusState: true},
	}
}

func (self *leaderCalculator) GetCurNumber() uint64 {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.curNumber
}

func (self *leaderCalculator) UpdateCacheByHeader(height uint64, headerLeader common.Address, validators *mc.TopologyGraph) error {
	if headerLeader.Equal(common.Address{}) {
		return errors.New("header leader不合法")
	}

	if nil == validators {
		return errors.New("验证者列表不合法")
	}

	var preLeader common.Address
	if common.IsBroadcastNumber(height) && height != 0 {
		preLeader = self.chain.GetBlockByNumber(height - 1).Header().Leader
	} else {
		preLeader = headerLeader
	}

	self.mu.Lock()
	defer self.mu.Unlock()

	if height+1 != self.curNumber { //高度变化了，重选轮次清空
		self.turns = 0
	}
	self.curNumber = height + 1
	self.preLeader.Set(preLeader)
	self.validators = validators
	self.leaders.consensusState = true
	return nil
}

func (self *leaderCalculator) GetValidatorCount(height uint64) int {
	self.mu.Lock()
	defer self.mu.Unlock()
	if height != self.curNumber {
		return 0
	}

	return len(self.validators.NodeList)
}

func (self *leaderCalculator) UpdateCacheByConsensus(height uint64, turns uint8, consensusState bool) (*leaderData, error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if height != self.curNumber {
		return nil, errors.Errorf("高度不匹配, param height[%d] != cache height[%d]", height, self.curNumber)
	}

	log.INFO(self.extra, "修改轮次成功, 原轮次", self.turns, "现轮次", turns, "共识结果", consensusState, "高度", height)
	self.turns = turns

	if err := self.updateLeaders(); err != nil {
		return nil, err
	}

	self.leaders.consensusState = consensusState
	data := self.leaders.copyData()
	go self.sendLeaderChangeMsg(data)
	return data, nil
}

func (self *leaderCalculator) GetLeaderInfo() (*leaderData, error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if err := self.updateLeaders(); err != nil {
		return nil, err
	}

	return self.leaders.copyData(), nil
}

func (self *leaderCalculator) NotifyLeaderChange() error {
	self.mu.Lock()
	defer self.mu.Unlock()

	if err := self.updateLeaders(); err != nil {
		return err
	}

	go self.sendLeaderChangeMsg(self.leaders.copyData())
	return nil
}

func (self *leaderCalculator) updateLeaders() error {
	if self.leaders.number == self.curNumber && self.leaders.turns == self.turns {
		return nil
	}

	leader, err := self.calLeaderByHeader()
	if err != nil {
		return errors.Errorf("Leader计算错误, %s", err)
	}

	self.leaders.number = self.curNumber
	self.leaders.turns = self.turns
	self.leaders.leader.Set(leader[0])
	self.leaders.nextLeader.Set(leader[1])

	return nil
}

func (self *leaderCalculator) calLeaderByHeader() (leaders [2]common.Address, err error) {
	if self.validators == nil || len(self.validators.NodeList) == 0 {
		return leaders, errors.New("验证者列表为空")
	}

	if common.IsReElectionNumber(self.curNumber - 1) {
		if leaders[0], err = nextLeaderByNum(self.validators.NodeList, self.validators.NodeList[0].Account, self.turns); err != nil {
			return leaders, err
		}
		if leaders[1], err = nextLeaderByNum(self.validators.NodeList, leaders[0], self.turns+1); err != nil {
			return leaders, err
		}
		return leaders, nil
	}

	if common.IsReElectionNumber(self.curNumber) {
		leaders[0] = params.BroadCastNodes[0].Address
		if leaders[1], err = nextLeaderByNum(self.validators.NodeList, self.validators.NodeList[0].Account, self.turns); err != nil {
			return leaders, err
		}
		return leaders, nil
	}

	if common.IsBroadcastNumber(self.curNumber) {
		leaders[0] = params.BroadCastNodes[0].Address
		if leaders[1], err = nextLeaderByNum(self.validators.NodeList, self.preLeader, self.turns+1); err != nil {
			return leaders, err
		}
		return leaders, nil
	}

	if leaders[0], err = nextLeaderByNum(self.validators.NodeList, self.preLeader, self.turns+1); err != nil {
		return leaders, err
	}

	if common.IsBroadcastNumber(self.curNumber + 1) {
		leaders[1] = params.BroadCastNodes[0].Address
	} else {
		if leaders[1], err = nextLeaderByNum(self.validators.NodeList, self.preLeader, self.turns+2); err != nil {
			return leaders, err
		}
	}

	return leaders, nil
}

func (self *leaderCalculator) sendLeaderChangeMsg(leaders *leaderData) {
	msg := &mc.LeaderChangeNotify{
		Leader:         leaders.leader,
		NextLeader:     leaders.nextLeader,
		ReelectTurn:    leaders.turns,
		Number:         leaders.number,
		ConsensusState: leaders.consensusState,
	}

	mc.PublishEvent(mc.Leader_LeaderChangeNotify, msg)
	log.INFO(self.extra, "发布身份更新, Leader", common.Bytes2Hex(msg.Leader[:7]), "Next Leader", common.Bytes2Hex(msg.NextLeader[:7]), "Height", msg.Number, "Turns", msg.ReelectTurn)
}
