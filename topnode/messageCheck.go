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
package topnode

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mc"

	"sync"
)

type messageCheck struct {
	mu       sync.RWMutex
	leader   common.Address
	curRound uint64
}

func (chk *messageCheck) checkLeaderChangeNotify(msg *mc.LeaderChangeNotify) bool {
	if msg.ConsensusState {
		round := msg.Number*100 + uint64(msg.ReelectTurn)
		if chk.setRound(round) {
			chk.setLeader(msg.Leader)
			return true
		}
	}
	return false
}
func (chk *messageCheck) checkOnlineConsensusReq(msg *mc.OnlineConsensusReq) bool {
	return chk.setRound(msg.Seq)
}
func (chk *messageCheck) checkOnlineConsensusVote(msg *mc.HD_ConsensusVote) bool {
	return chk.setRound(msg.Round)
}

func (chk *messageCheck) setLeader(leader common.Address) {
	chk.mu.Lock()
	defer chk.mu.Unlock()
	chk.leader = leader
}
func (chk *messageCheck) getLeader() common.Address {
	chk.mu.RLock()
	defer chk.mu.RUnlock()
	return chk.leader
}
func (chk *messageCheck) getRound() uint64 {
	chk.mu.RLock()
	defer chk.mu.RUnlock()
	return chk.curRound
}
func (chk *messageCheck) checkRound(round uint64) bool {
	chk.mu.RLock()
	defer chk.mu.RUnlock()
	return round >= chk.curRound
}
func (chk *messageCheck) setRound(round uint64) bool {
	chk.mu.Lock()
	defer chk.mu.Unlock()
	if round < chk.curRound {
		return false
	}
	if round > chk.curRound {
		chk.curRound = round
	}
	return true
}
