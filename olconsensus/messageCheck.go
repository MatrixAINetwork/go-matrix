// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package olconsensus

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/log"
	"sync"
)

type messageCheck struct {
	mu       sync.RWMutex
	leader   common.Address
	curRound uint64
	blockHash common.Hash
}

func (chk *messageCheck) checkLeaderChangeNotify(msg *mc.LeaderChangeNotify) bool {
	if msg.ConsensusState {
		round := msg.Number*100 + uint64(msg.ReelectTurn)
		if chk.setRound(round) {
			chk.setLeader(msg.Leader)
			log.Info("topnodeOnline", "设置leader", msg.Leader, "设置Number", msg.Number)
			return true
		}
	}
	return false
}

func (chk *messageCheck) checkBlockHash(hash common.Hash) bool {
	if !hash.Equal(chk.getBlockHash()) {
		chk.setBlockHash(hash)
			return true
	}
	return false
}

func (chk *messageCheck) checkOnlineConsensusReq(msg *mc.OnlineConsensusReq) bool {
	return chk.setRound(msg.Seq)
}
func (chk *messageCheck) checkOnlineConsensusVote(msg *mc.HD_ConsensusVote) bool {
	return chk.setRound(msg.Round)
}
func (chk *messageCheck) setBlockHash(hash common.Hash) {
	chk.mu.Lock()
	defer chk.mu.Unlock()
	chk.blockHash = hash
}
func (chk *messageCheck) getBlockHash() common.Hash {
	chk.mu.RLock()
	defer chk.mu.RUnlock()
	return chk.blockHash
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
