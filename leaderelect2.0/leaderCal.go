// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package leaderelect2

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
)

type leaderCalculator struct {
	number          uint64
	preLeader       common.Address
	preHash         common.Hash
	leaderList      map[uint32]common.Address
	validators      []mc.TopologyNodeInfo
	specialAccounts specialAccounts
	chain           *core.BlockChain
	logInfo         string
}

func newLeaderCalculator(chain *core.BlockChain, number uint64, logInfo string) *leaderCalculator {
	return &leaderCalculator{
		number:          number,
		preLeader:       common.Address{},
		preHash:         common.Hash{},
		leaderList:      make(map[uint32]common.Address),
		validators:      nil,
		specialAccounts: specialAccounts{},
		chain:           chain,
		logInfo:         logInfo,
	}
}

func (self *leaderCalculator) getRealPreLeader(preHeader *types.Header, bcInterval *mc.BCIntervalInfo) (common.Address, bool, error) {
	header := preHeader
	number := preHeader.Number.Uint64()
	preAppearSuper := header.IsSuperHeader()

	for header.IsSuperHeader() || bcInterval.IsBroadcastNumber(number) {
		if header.IsSuperHeader() {
			preAppearSuper = true
		}

		if number == 0 {
			return header.Leader, preAppearSuper, nil
		}
		header = self.chain.GetHeader(header.ParentHash, number-1)
		if header == nil {
			return common.Address{}, preAppearSuper, errors.Errorf("获取父区块(%d, %s)失败, header is nil! ", number-1, header.ParentHash.TerminalString())
		}
		number = header.Number.Uint64()
	}
	return header.Leader, preAppearSuper, nil
}

func (self *leaderCalculator) SetValidatorsAndSpecials(preHeader *types.Header, validators []mc.TopologyNodeInfo, specials *specialAccounts, bcInterval *mc.BCIntervalInfo) error {
	if preHeader == nil || validators == nil || specials == nil || bcInterval == nil {
		return ErrValidatorsIsNil
	}

	realPreLeader, preAppearSuper, err := self.getRealPreLeader(preHeader, bcInterval)
	if err != nil {
		log.Error(self.logInfo, "计算leader列表", "获取真实的preLeader失败", "err", err)
		return err
	}
	log.Trace(self.logInfo, "计算leader列表", "开始", "preLeader", realPreLeader.Hex(), "前区块中出现超级区块", preAppearSuper, "高度", self.number, "validators size", len(validators))
	leaderList, err := calLeaderList(realPreLeader, self.number, preAppearSuper, validators, bcInterval, self.logInfo)
	if err != nil {
		return err
	}
	self.leaderList = leaderList
	self.preLeader.Set(preHeader.Leader)
	self.preHash.Set(preHeader.Hash())
	self.validators = validators
	self.specialAccounts.broadcasts = specials.broadcasts
	self.specialAccounts.versionSupers = specials.versionSupers
	self.specialAccounts.blockSupers = specials.blockSupers

	return nil
}

func (self *leaderCalculator) GetValidators() (*mc.TopologyGraph, error) {
	if len(self.validators) == 0 {
		return nil, errors.New("验证者列表为空")
	}
	rlt := &mc.TopologyGraph{}
	for i := 0; i < len(self.validators); i++ {
		rlt.NodeList = append(rlt.NodeList, self.validators[i])
	}
	return rlt, nil
}

func (self *leaderCalculator) GetLeader(turn uint32, bcInterval *mc.BCIntervalInfo) (*leaderData, error) {
	if bcInterval == nil {
		return nil, errors.New("leader calculator: param bcInterval is nil")
	}
	leaderCount := uint32(len(self.leaderList))
	if leaderCount == 0 {
		return nil, ErrValidatorsIsNil
	}

	leaders := &leaderData{}
	number := self.number
	if bcInterval.IsReElectionNumber(number) {
		leaders.leader = common.Address{}
		leaders.nextLeader.Set(self.leaderList[turn%leaderCount])
		return leaders, nil
	}

	if bcInterval.IsBroadcastNumber(number) {
		leaders.leader = common.Address{}
		leaders.nextLeader.Set(self.leaderList[(turn)%leaderCount])
		return leaders, nil
	}

	leaders.leader.Set(self.leaderList[turn%leaderCount])
	if bcInterval.IsBroadcastNumber(number + 1) {
		leaders.nextLeader = common.Address{}
	} else {
		leaders.nextLeader.Set(self.leaderList[(turn+1)%leaderCount])
	}
	return leaders, nil
}

func calLeaderList(preLeader common.Address, curNumber uint64, preIsSupper bool, validators []mc.TopologyNodeInfo, bcInterval *mc.BCIntervalInfo, logInfo string) (map[uint32]common.Address, error) {
	ValidatorNum := len(validators)
	var startPos = 0
	if preIsSupper || bcInterval.IsReElectionNumber(curNumber-1) || bcInterval.IsReElectionNumber(curNumber) {
		startPos = 0
	} else {
		if preIndex, err := findLeaderIndex(preLeader, validators); err != nil {
			log.Info(logInfo, "未在验证者列表中未找到preLeader", preLeader.Hex(), "validators", validators)
			startPos = 0
		} else {
			startPos = preIndex + 1
		}
	}
	leaderList := make(map[uint32]common.Address)
	for i := 0; i < ValidatorNum; i++ {
		leaderList[uint32(i)] = validators[(startPos+int(i))%ValidatorNum].Account
	}
	return leaderList, nil
}

func findLeaderIndex(preLeader common.Address, validators []mc.TopologyNodeInfo) (int, error) {
	for index, v := range validators {
		if v.Account == preLeader {
			return index, nil
		}
	}
	return 0, ErrValidatorNotFound
}

func (self *leaderCalculator) dumpAllValidators(logInfo string) {
	size := len(self.validators)
	log.Debug(logInfo, "dump info", "验证者列表", "总数", size, "高度", self.number, "parentHash", self.preHash.TerminalString(), "parentLeader", self.preLeader.Hex())
	for i := 0; i < size; i++ {
		item := self.validators[i]
		log.Debug(logInfo, "dump info", "验证者列表", "index", i, "node", item.Account.Hex(), "pos", item.Position)
	}
}
