// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package verifier

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/man"
	"github.com/pkg/errors"
)

type leaderCalculator struct {
	preLeader  common.Address
	leaderList map[uint32]common.Address
	validators []mc.TopologyNodeInfo
	chain      *core.BlockChain
	cdc        *cdc
}

func newLeaderCalculator(chain *core.BlockChain, cdc *cdc) *leaderCalculator {
	return &leaderCalculator{
		preLeader:  common.Address{},
		leaderList: make(map[uint32]common.Address),
		validators: nil,
		chain:      chain,
		cdc:        cdc,
	}
}

func (self *leaderCalculator) SetValidators(preLeader common.Address, validators []mc.TopologyNodeInfo) error {
	if validators == nil {
		return ErrValidatorsIsNil
	}

	preNumber := self.cdc.number - 1
	if common.IsBroadcastNumber(preNumber) && preNumber != 0 {
		header := self.chain.GetHeaderByNumber(preNumber - 1)
		if nil == header {
			log.ERROR("")
			return errors.Errorf("error obtaining the previous block (%d) of broadcast block", preNumber-1)
		}
		preLeader = header.Leader
	}

	leaderList, err := calLeaderList(preLeader, validators, self.cdc.number)
	if err != nil {
		return err
	}
	self.leaderList = leaderList
	self.preLeader = preLeader
	self.validators = validators

	return nil
}

func (self *leaderCalculator) GetValidators() (*mc.TopologyGraph, error) {
	if len(self.validators) == 0 {
		return nil, errors.New("validator list is blank")
	}
	rlt := &mc.TopologyGraph{}
	for i := 0; i < len(self.validators); i++ {
		rlt.NodeList = append(rlt.NodeList, self.validators[i])
	}
	return rlt, nil
}

func (self *leaderCalculator) GetLeader(turn uint32) (*leaderData, error) {
	leaderCount := uint32(len(self.leaderList))
	if leaderCount == 0 {
		return nil, ErrValidatorsIsNil
	}

	leaders := &leaderData{}
	number := self.cdc.number
	if common.IsReElectionNumber(number) {
		leaders.leader.Set(man.BroadCastNodes[0].Address)
		leaders.nextLeader.Set(self.leaderList[turn%leaderCount])
		return leaders, nil
	}

	if common.IsBroadcastNumber(number) {
		leaders.leader.Set(man.BroadCastNodes[0].Address)
		leaders.nextLeader.Set(self.leaderList[(turn)%leaderCount])
		return leaders, nil
	}

	leaders.leader.Set(self.leaderList[turn%leaderCount])
	if common.IsBroadcastNumber(number + 1) {
		leaders.nextLeader.Set(man.BroadCastNodes[0].Address)
	} else {
		leaders.nextLeader.Set(self.leaderList[(turn+1)%leaderCount])
	}
	return leaders, nil
}

func calLeaderList(preLeader common.Address, validators []mc.TopologyNodeInfo, number uint64) (map[uint32]common.Address, error) {
	ValidatorNum := len(validators)
	startPos := 0
	if common.IsReElectionNumber(number) == false {
		preIndex, err := findLeaderIndex(preLeader, validators)
		if err != nil {
			return nil, err
		}
		startPos = preIndex + 1
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
