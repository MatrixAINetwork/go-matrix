// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package olconsensus

import (
	"errors"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"sync"
)

const (
	onlineNum  = 15
	offlineNum = 3
)

//读当前区块的状态，获得选举，获得在线，offline = elect - online
type topNodeState struct {
	mu                  sync.RWMutex
	curNumber           uint64
	curTopologyNodes    []*mc.TopologyNodeInfo                // 当前拓扑图
	curElectNodes       []*mc.ElectNodeInfo                   // 不在拓扑中选举节点的状态(链上的状态)
	consensusResultRing []*mc.HD_OnlineConsensusVoteResultMsg //  todo   使用环，增加数量限制
	capacity            int
	last                int
	extraInfo           string
}

func newTopNodeState(capacity int, info string) *topNodeState {
	return &topNodeState{
		curNumber:           0,
		curTopologyNodes:    nil,
		curElectNodes:       nil,
		consensusResultRing: make([]*mc.HD_OnlineConsensusVoteResultMsg, capacity),
		capacity:            capacity,
		last:                capacity - 1,
		extraInfo:           info,
	}
}

func (ts *topNodeState) SetCurStates(curNumber uint64, topologyGroup *mc.TopologyGraph, electStates *mc.ElectOnlineStatus) {
	log.Info("共识节点状态", "设置当前节点状态", "区块高度，拓扑图，选举节点")
	if topologyGroup == nil || electStates == nil {
		log.Error("共识节点状态", "拓扑或者选举节点在线信息异常", "为nil")
		return
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.curNumber = curNumber
	ts.curTopologyNodes = make([]*mc.TopologyNodeInfo, 0)
	ts.curElectNodes = make([]*mc.ElectNodeInfo, 0)
	for i := 0; i < len(topologyGroup.NodeList); i++ {
		node := topologyGroup.NodeList[i]
		if node.Type != common.RoleValidator && node.Type != common.RoleBackupValidator {
			// 只关注验证者和备选验证者
			continue
		}
		ts.curTopologyNodes = append(ts.curTopologyNodes, &node)
	}

	for i := 0; i < len(electStates.ElectOnline); i++ {
		electState := electStates.ElectOnline[i]
		if electState.Type != common.RoleValidator && electState.Type != common.RoleBackupValidator {
			// 只关注验证者和备选验证者
			continue
		}
		if isInsideList(electState.Account, ts.curTopologyNodes) {
			// 在当前拓扑图中的选举节点，状态均认为在线
			continue
		}
		ts.curElectNodes = append(ts.curElectNodes, &electState)
	}

	for i, info := range ts.curElectNodes {
		log.Debug("共识节点状态", "SetCurStates_elect index", i, "node", info.Account.Hex(), "pos", info.Position, "type", info.Type)
	}
	log.Info("共识节点状态", "当前节点状态设置完成", "     ", "区块高度", curNumber)
}

func (ts *topNodeState) SaveConsensusResult(result *mc.HD_OnlineConsensusVoteResultMsg) {
	if result == nil || result.Req == nil {
		return
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()

	old, index, err := ts.findConsensusResult(result.Req.Node)
	if err == nil {
		switch cmpRound(old.Req.Number, old.Req.LeaderTurn, result.Req.Number, result.Req.LeaderTurn) {
		case 0, 1: //old >= param
			return
		case -1: // old < param
			ts.consensusResultRing[index] = result
			log.Debug("共识节点状态", "共识请求发出时的高度", result.Req.Number, "轮次", result.Req.LeaderTurn, "被共识的节点", result.Req.Node.Hex(), "被共识的状态", result.Req.OnlineState.String())
		}
	} else {
		ts.insertConsensusResult(result)
		log.Debug("共识节点状态", "共识请求发出时的高度", result.Req.Number, "轮次", result.Req.LeaderTurn, "被共识的节点", result.Req.Node.Hex(), "被共识的状态", result.Req.OnlineState.String())
	}
	log.Info("共识节点状态", "保存共识结果完成", "   ", "被共识的节点", result.Req.Node.Hex(), "被共识的状态", result.Req.OnlineState.String())
}

func (ts *topNodeState) GetConsensusResults() (results []*mc.HD_OnlineConsensusVoteResultMsg) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for i := 0; i < ts.capacity; i++ {
		one := ts.consensusResultRing[i]
		if one == nil || one.Req == nil {
			continue
		}

		// 过期的删除
		if one.IsValidity(ts.curNumber, manparams.OnlineConsensusValidityTime) == false {
			ts.consensusResultRing[i] = nil
			continue
		}

		results = append(results, one)
	}

	return results
}

func (ts *topNodeState) findConsensusResult(node common.Address) (*mc.HD_OnlineConsensusVoteResultMsg, int, error) {
	for i := 0; i < ts.capacity; i++ {
		result := ts.consensusResultRing[i]
		if result == nil || result.Req == nil {
			continue
		}
		if result.Req.Node == node {
			return result, i, nil
		}
	}
	return nil, 0, errors.New("result not find")
}

func (ts *topNodeState) insertConsensusResult(result *mc.HD_OnlineConsensusVoteResultMsg) {
	ts.last = (ts.last + 1) % ts.capacity
	ts.consensusResultRing[ts.last] = result
}

func isInsideList(node common.Address, listNode []*mc.TopologyNodeInfo) bool {
	for _, item := range listNode {
		if item.Account == node {
			return true
		}
	}
	return false
}

func (ts *topNodeState) newTopNodeState(nodesOnlineInfo []NodeOnLineInfo, leader common.Address) (online, offline []common.Address) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	for _, value := range nodesOnlineInfo {
		if value.Address.Equal(leader) {
			continue
		}
		needCheck, checkState := ts.nodeIsNeedCheckState(value.Address)
		if needCheck == false {
			continue
		}

		switch checkState {
		case mc.OffLine:
			if isOffline(value.OnlineState) {
				offline = append(offline, value.Address)
				log.Info("共识节点状态", "检查节点在线状态", "离线", "当前高度", ts.curNumber, "节点", value.Address.String(), "offline", "需要共识")
			}
		case mc.OnLine:
			if isOnline(value.OnlineState) {
				online = append(online, value.Address)
				log.Info("共识节点状态", "检查节点在线状态", "在线", "当前高度", ts.curNumber, "节点", value.Address.String(), "online", "需要共识")
			}
		}
	}
	return
}

// 判断是否需要检查状态
func (ts *topNodeState) nodeIsNeedCheckState(node common.Address) (needCheck bool, checkState mc.OnlineState) {
	if isInsideList(node, ts.curTopologyNodes) {
		// node在当前拓扑图中
		result, _, err := ts.findConsensusResult(node)
		if err == nil && result.IsValidity(ts.curNumber, manparams.OnlineConsensusValidityTime) && result.Req.OnlineState == mc.OffLine {
			// 存在有效的共识结果
			return false, mc.OffLine
		} else {
			return true, mc.OffLine
		}
	}

	for _, elect := range ts.curElectNodes {
		if elect.Account != node {
			continue
		}
		// node为不在拓扑中的elect节点
		if elect.Position == common.PosOnline {
			checkState = mc.OffLine
		} else if elect.Position == common.PosOffline {
			checkState = mc.OnLine
		} else {
			return false, mc.OffLine
		}

		result, _, err := ts.findConsensusResult(node)
		if err == nil && result.IsValidity(ts.curNumber, manparams.OnlineConsensusValidityTime) && result.Req.OnlineState == checkState {
			// 存在有效的共识结果
			return false, mc.OffLine
		} else {
			return true, checkState
		}
	}
	return false, mc.OffLine
}

func (ts *topNodeState) checkNodeState(node common.Address, nodesOnlineInfo []NodeOnLineInfo, checkState mc.OnlineState) bool {
	for _, item := range nodesOnlineInfo {
		if item.Address != node {
			continue
		}

		switch checkState {
		case mc.OnLine:
			return isOnline(item.OnlineState)
		case mc.OffLine:
			return isOffline(item.OnlineState)
		default:
			return false
		}
	}
	return false
}

func isOnline(state []uint8) bool {
	heartNumber := len(state) - 1
	for i := heartNumber; i > heartNumber-onlineNum; i-- {
		if state[i] == 0 {
			return false
		}
	}
	return true
}
func isOffline(state []uint8) bool {
	heartNumber := len(state) - 1
	for i := heartNumber; i > heartNumber-offlineNum; i-- {
		if state[i] != 0 {
			return false
		}
	}
	return true
}
