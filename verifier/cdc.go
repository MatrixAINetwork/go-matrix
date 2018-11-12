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
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

type cdc struct {
	state            state
	number           uint64
	curConsensusTurn uint32
	consensusLeader  common.Address
	curReelectTurn   uint32
	reelectMaster    common.Address
	leaderCal        *leaderCalculator
	turnTime         *turnTimes
	chain            *core.BlockChain
	logInfo          string
}

func newCDC(number uint64, chain *core.BlockChain, logInfo string) *cdc {
	dc := &cdc{
		state:            stIdle,
		number:           number,
		curConsensusTurn: 0,
		consensusLeader:  common.Address{},
		curReelectTurn:   0,
		reelectMaster:    common.Address{},
		turnTime:         newTurnTimes(),
		chain:            chain,
	}

	dc.leaderCal = newLeaderCalculator(chain, dc)
	return dc
}

func (dc *cdc) SetValidators(preLeader common.Address, validators []mc.TopologyNodeInfo) error {
	if err := dc.leaderCal.SetValidators(preLeader, validators); err != nil {
		return err
	}

	consensusLeader, err := dc.GetLeader(dc.curConsensusTurn)
	if err != nil {
		return err
	}
	if dc.curReelectTurn != 0 {
		reelectLeader, err := dc.GetLeader(dc.curConsensusTurn + dc.curReelectTurn)
		if err != nil {
			return err
		}
		dc.reelectMaster.Set(reelectLeader)
	} else {
		dc.reelectMaster.Set(common.Address{})
	}
	dc.consensusLeader.Set(consensusLeader)
	return nil
}

func (dc *cdc) SetConsensusTurn(consensusTurn uint32) error {
	consensusLeader, err := dc.GetLeader(consensusTurn)
	if err != nil {
		return errors.Errorf("获取共识leader错误(%v), 共识轮次(%d)", err, consensusTurn)
	}

	dc.consensusLeader.Set(consensusLeader)
	dc.curConsensusTurn = consensusTurn
	dc.reelectMaster.Set(common.Address{})
	dc.curReelectTurn = 0
	return nil
}

func (dc *cdc) SetReelectTurn(reelectTurn uint32) error {
	if dc.curReelectTurn == reelectTurn {
		return nil
	}
	if reelectTurn == 0 {
		dc.reelectMaster.Set(common.Address{})
		dc.curReelectTurn = 0
		return nil
	}
	master, err := dc.GetLeader(dc.curConsensusTurn + reelectTurn)
	if err != nil {
		return errors.Errorf("获取master错误(%v), 重选轮次(%d), 共识轮次(%d)", err, reelectTurn, dc.curConsensusTurn)
	}
	dc.reelectMaster.Set(master)
	dc.curReelectTurn = reelectTurn
	return nil
}

func (dc *cdc) GetLeader(turn uint32) (common.Address, error) {
	leaders, err := dc.leaderCal.GetLeader(turn)
	if err != nil {
		return common.Address{}, err
	}
	return leaders.leader, nil
}

func (dc *cdc) GetConsensusLeader() common.Address {
	return dc.consensusLeader
}

func (dc *cdc) GetReelectMaster() common.Address {
	return dc.reelectMaster
}

func (dc *cdc) PrepareLeaderMsg() (*mc.LeaderChangeNotify, error) {
	leaders, err := dc.leaderCal.GetLeader(dc.curConsensusTurn + dc.curReelectTurn)
	if err != nil {
		return nil, err
	}

	return &mc.LeaderChangeNotify{
		Leader:         leaders.leader,
		NextLeader:     leaders.nextLeader,
		ConsensusTurn:  dc.curConsensusTurn,
		ReelectTurn:    dc.curReelectTurn,
		Number:         dc.number,
		ConsensusState: dc.state != stReelect,
		TurnBeginTime:  dc.turnTime.GetBeginTime(dc.curConsensusTurn),
		TurnEndTime:    dc.turnTime.GetPosEndTime(dc.curConsensusTurn),
	}, nil
}

func (dc *cdc) GetCurrentNumber() uint64 {
	return dc.number - 1
}
func (dc *cdc) GetValidatorByNumber(number uint64) (*mc.TopologyGraph, error) {
	if number >= dc.number {
		return nil, errors.Errorf("获取验证者列表错误,高度过高")
	}

	validators, err := dc.chain.GetValidatorByNumber(number)
	if err == nil {
		return validators, nil
	}

	if number == dc.number-1 {
		return dc.leaderCal.GetValidators()
	} else {
		return nil, err
	}
}
