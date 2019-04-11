// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package olconsensus

import (
	"sync"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

type messageCheck struct {
	mu            sync.RWMutex
	curNumber     uint64 // 高度
	curLeaderTurn uint32 // leader轮次
	leaderCache   []*mc.LeaderChangeNotify
	capacity      int
	last          int
	blockHash     common.Hash
}

func newMessageCheck(capacity int) *messageCheck {
	return &messageCheck{
		curNumber:     0,
		curLeaderTurn: 0,
		leaderCache:   make([]*mc.LeaderChangeNotify, capacity),
		capacity:      capacity,
		last:          capacity - 1,
	}
}

func (chk *messageCheck) CheckRoleUpdateMsg(msg *mc.RoleUpdatedMsg) bool {
	if nil == msg {
		log.Error("共识节点状态", "处理CA通知消息", "消息为空")
		return false
	}

	if msg.Role != common.RoleValidator {
		log.Debug("共识节点状态", "身份不对", "不做处理")
		return false
	}

	chk.mu.Lock()
	defer chk.mu.Unlock()
	rst := chk.setCurNumber(msg.BlockNum + 1)

	if rst {
		chk.blockHash = msg.BlockHash
	}
	return rst
}

func (chk *messageCheck) CheckAndSaveLeaderChangeNotify(msg *mc.LeaderChangeNotify) bool {
	if nil == msg {
		return false
	}
	if msg.ConsensusState == false {
		return false
	}

	chk.mu.Lock()
	defer chk.mu.Unlock()
	switch cmpRound(chk.curNumber, chk.curLeaderTurn, msg.Number, msg.ConsensusTurn.TotalTurns()) {
	case 1: // cur > msg
		return false
	case -1: // cur < msg
		chk.setLeader(msg)
		return false
	case 0: // cur == msg
		return chk.setLeader(msg)
	default:
		return false
	}
}

func (chk *messageCheck) GetCurLeader() common.Address {
	chk.mu.Lock()
	defer chk.mu.Unlock()
	return chk.getLeader(chk.curNumber, chk.curLeaderTurn)
}

func (chk *messageCheck) GetRound() (number uint64, turn uint32) {
	chk.mu.Lock()
	defer chk.mu.Unlock()
	return chk.curNumber, chk.curLeaderTurn
}

// localRound > param, return 1
// localRound = param, return 0
// localRound < param, return -1
func (chk *messageCheck) CheckRound(number uint64, leaderTurn uint32) int {
	chk.mu.Lock()
	defer chk.mu.Unlock()
	return cmpRound(chk.curNumber, chk.curLeaderTurn, number, leaderTurn)
}

func (chk *messageCheck) setCurNumber(number uint64) bool {
	if number < chk.curNumber {
		log.Warn("共识节点状态", "区块高度不符", "     ", number, chk.curNumber)
		return false
	}
	if number > chk.curNumber {
		chk.curNumber = number
	}
	return true
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
func (chk *messageCheck) setLeader(msg *mc.LeaderChangeNotify) bool {
	// 检查重复
	for i, one := range chk.leaderCache {
		if one == nil {
			continue
		}
		if one.Number == msg.Number && one.ConsensusTurn == msg.ConsensusTurn {
			if one.Leader != msg.Leader {
				chk.leaderCache[i] = msg
				return true
			}
			return false
		}
	}
	chk.last = (chk.last + 1) % chk.capacity
	chk.leaderCache[chk.last] = msg
	return true
}

func (chk *messageCheck) getLeader(number uint64, turn uint32) common.Address {

	for _, msg := range chk.leaderCache {
		if msg == nil {
			continue
		}
		if msg.Number == number && msg.ConsensusTurn.TotalTurns() == turn {
			return msg.Leader
		}
	}
	return common.Address{}
}

// A > B, return 1
// A = B, return 0
// A < B, return -1
func cmpRound(ANumber uint64, ATurn uint32, BNumber uint64, BTurn uint32) int {
	if ANumber > BNumber {
		return 1
	} else if ANumber < BNumber {
		return -1
	} else {
		if ATurn > BTurn {
			return 1
		} else if ATurn < BTurn {
			return -1
		} else {
			return 0
		}
	}
}
