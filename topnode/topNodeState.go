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
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/rlp"

	"sync"
)

const (
	onLine = iota + 1
	offLine

	onlineNum  = 15
	offlineNum = 3
)

type topNodeState struct {
	mu               sync.RWMutex
	electNode        map[common.Address]uint8
	onlineNode       map[common.Address]uint8
	offlineNode      map[common.Address]uint8
	finishedProposal *DPosVoteRing
}

func newTopNodeState(capacity int) *topNodeState {
	return &topNodeState{
		electNode:        make(map[common.Address]uint8),
		onlineNode:       make(map[common.Address]uint8),
		offlineNode:      make(map[common.Address]uint8),
		finishedProposal: NewDPosVoteRing(capacity),
	}
}
func (ts *topNodeState) setElectNodes(nodes []common.Address) {
	ts.electNode = make(map[common.Address]uint8)
	ts.onlineNode = make(map[common.Address]uint8)
	ts.offlineNode = make(map[common.Address]uint8)
	for _, item := range nodes {
		ts.electNode[item] = 1
		ts.onlineNode[item] = 1
	}
}

//modify
func (ts *topNodeState) modifyTopNodeState(online, offline []common.Address) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for _, item := range online {
		if _, exist := ts.electNode[item]; exist {
			ts.onlineNode[item] = 1
		}
		delete(ts.offlineNode, item)
	}
	for _, item := range offline {
		delete(ts.onlineNode, item)
		if _, exist := ts.electNode[item]; exist {
			ts.offlineNode[item] = 1
		}
	}
}
func (ts *topNodeState) newTopNodeState(nodesOnlineInfo []NodeOnLineInfo) (online, offline []common.Address) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	for _, value := range nodesOnlineInfo {
		if _, exist := ts.onlineNode[value.Address]; exist {
			if isOffline(value.OnlineState) && (!ts.isFinishedPropocal(value.Address, offLine)) {
				offline = append(offline, value.Address)
			}
		}
		if _, exist := ts.offlineNode[value.Address]; exist {
			if isOnline(value.OnlineState) && (!ts.isFinishedPropocal(value.Address, onLine)) {
				online = append(online, value.Address)
			}
		}
	}
	return
}
func getFinishedPropocalHash(node common.Address, onLine uint8) common.Hash {
	var hash common.Hash
	copy(hash[:20], node[:])
	hash[21] = onLine
	return hash
}
func (ts *topNodeState) isFinishedPropocal(node common.Address, onLine uint8) bool {
	propocal, _ := ts.finishedProposal.getVotes(getFinishedPropocalHash(node, onLine))
	return propocal != nil
}
func (ts *topNodeState) checkNodeOnline(node common.Address, nodesOnlineInfo []NodeOnLineInfo) bool {
	for _, item := range nodesOnlineInfo {
		if item.Address == node {
			return isOnline(item.OnlineState)
		}
	}
	return false
}
func (ts *topNodeState) checkNodeOffline(node common.Address, nodesOnlineInfo []NodeOnLineInfo) bool {
	for _, item := range nodesOnlineInfo {
		if item.Address == node {
			return isOffline(item.OnlineState)
		}
	}
	return false
}

func isOnline(state [30]uint8) bool {
	for i := 29; i > 29-onlineNum; i-- {
		if state[i] == 0 {
			return false
		}
	}
	return true
}
func isOffline(state [30]uint8) bool {
	for i := 29; i > 29-offlineNum; i-- {
		if state[i] != 0 {
			return false
		}
	}
	return true
}

type topNodeCheck struct {
	mu       sync.RWMutex
	curRound uint64
	caChan   chan struct{}
}

func (chk *topNodeCheck) checkMessage(aim mc.EventCode, value interface{}) (uint64, bool) {
	switch aim {
	case mc.CA_RoleUpdated:
		data := value.(mc.RoleUpdatedMsg)
		round := data.BlockNum * 100
		if chk.setRound(round) {
			return round, true
		}
	case mc.Leader_LeaderChangeNotify:
		data := value.(*mc.LeaderChangeNotify)
		round := data.Number*100 + uint64(data.ReelectTurn)
		if chk.setRound(round) {
			if data.Leader == ca.GetAddress() {
				chk.caChan <- struct{}{}
			}
			return round, true
		}
	case mc.HD_TopNodeConsensusReq:
		data := value.(*mc.HD_OnlineConsensusReqs)
		round := data.ReqList[0].Seq
		if chk.checkRound(round) {
			return round, true
		}
	case mc.HD_TopNodeConsensusVote:
		data := value.(*mc.HD_OnlineConsensusVotes)
		round := data.Votes[0].Round
		if chk.checkRound(round) {
			return round, true
		}
	}
	return 0, false
}
func (chk *topNodeCheck) getKeyBytes(value interface{}) []byte {
	val, _ := rlp.EncodeToBytes(value)
	return val
}
func (chk *topNodeCheck) checkState(state []byte, round uint64) bool {
	if chk.checkRound(round) {
		return (state[0] == 1 || state[1] == 1) && state[2] == 1 && state[3] == 1
	}
	return false
}
func (chk *topNodeCheck) checkRound(round uint64) bool {
	chk.mu.RLock()
	defer chk.mu.RUnlock()
	return round >= chk.curRound
}
func (chk *topNodeCheck) setRound(round uint64) bool {
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
