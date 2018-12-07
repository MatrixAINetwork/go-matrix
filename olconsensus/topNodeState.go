// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package olconsensus

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/math"
	"github.com/matrix/go-matrix/log"
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

//读当前区块的状态，获得选举，获得在线，offline = elect - online
type topNodeState struct {
	mu               sync.RWMutex
	electHeight      uint64
	electNode        map[common.Address]OnlineState //选举结果
	onlineNode       []common.Address               //当前在线
	offlineNode      []common.Address               //所有掉线，需要验证在线的
	consensusOn      []common.Address
	consensusOff     []common.Address
	finishedProposal *DPosVoteRing
	extraInfo        string
}

func newTopNodeState(capacity int, info string) *topNodeState {
	return &topNodeState{
		electHeight:      uint64(math.MaxUint64),
		electNode:        make(map[common.Address]OnlineState),
		finishedProposal: NewDPosVoteRing(capacity),
		extraInfo:        info,
	}
}
func (ts *topNodeState) setElectNodes(nodes []common.Address, height uint64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.electHeight != height {
		ts.electHeight = height
		ts.electNode = make(map[common.Address]OnlineState)
		ts.onlineNode = nodes
		ts.offlineNode = nil
		for _, item := range nodes {
			ts.electNode[item] = 1
		}
	}
}

//输入参数是差值，变化值
func (ts *topNodeState) setCurrentTopNodeState(onLineNode, onElectNode []common.Address) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.onlineNode = onLineNode
	for key, _ := range ts.electNode {
		ts.electNode[key] = offLine
	}
	for _, item := range onElectNode {
		if !isInsideList(item, ts.onlineNode) {
			ts.onlineNode = append(ts.onlineNode, item)
			log.Info(ts.extraInfo, "区块生成", "设置当前在线状态", "添加在线节点列表", item.String())

		}
		ts.electNode[item] = onLine
	}
	ts.offlineNode = nil
	for key, value := range ts.electNode {
		if value == offLine {
			ts.offlineNode = append(ts.offlineNode, key)
			log.Info(ts.extraInfo, "区块生成", "设置当前在线状态", "添加离线节点列表", key.String())
		}
	}
	for _, item := range ts.onlineNode {
		ts.consensusOn = removeFromList(item, ts.consensusOn)
	}
	for _, item := range ts.offlineNode {
		ts.consensusOff = removeFromList(item, ts.consensusOff)
	}
}
func (ts *topNodeState) getCurrentTopNodeChange() (ret_offLineNode, ret_onElectNode, ret_offElectNode []common.Address) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	ret_offLineNode = ts.consensusOff
	ret_onElectNode = ts.consensusOn
	log.Info(ts.extraInfo, "区块生成", "获取经过共识的在线状态", "consensusOff", len(ts.consensusOff), "consensusOn", len(ts.consensusOn))
	for _, item := range ts.consensusOff {
		if _, exist := ts.electNode[item]; exist {
			ret_offElectNode = append(ret_offElectNode, item)
		}
	}
	return
}

/*
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
*/
func (ts *topNodeState) saveConsensusNodeState(node common.Address, onlineState OnlineState) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	switch onlineState {
	case onLine:
		ts.saveOnlineNode(node)
	case offLine:
		ts.saveOfflineNode(node)
	default:
		log.Error(ts.extraInfo, "处理共识投票结果", "保存经过共识的节点", "无效的在线状态", onlineState)
	}
}
func removeFromList(node common.Address, listNode []common.Address) []common.Address {
	for i, item := range listNode {
		if item == node {
			len := len(listNode)
			listNode[i], listNode[len-1] = listNode[len-1], listNode[i]
			return listNode[:len-1]
		}
	}
	return listNode
}
func isInsideList(node common.Address, listNode []common.Address) bool {
	bHave := false
	for _, item := range listNode {
		if item == node {
			bHave = true
			break
		}
	}
	return bHave
}
func (ts *topNodeState) saveOnlineNode(node common.Address) {
	if _, exist := ts.electNode[node]; exist {
		if !isInsideList(node, ts.consensusOn) {
			ts.consensusOn = append(ts.consensusOn, node)
			log.Info(ts.extraInfo, "处理共识投票结果", "保存经过上线共识的节点", "node", node.String())
		} else {
			log.Info(ts.extraInfo, "处理共识投票结果", "保存经过上线共识的节点", "node", node.String(), "已经在共识在线节点列表中", "")
		}
	} else {
		log.Info(ts.extraInfo, "处理共识投票结果", "保存经过上线共识的节点", "node", node.String(), "不在 electNode 中", "不保存")
	}
}

func (ts *topNodeState) saveOfflineNode(node common.Address) {
	log.Info(ts.extraInfo, "保存掉线节点: node", node.String(), "ts.onlineNode", len(ts.onlineNode))
	if isInsideList(node, ts.onlineNode) {
		if !isInsideList(node, ts.consensusOff) {
			ts.consensusOff = append(ts.consensusOff, node)
			log.Info(ts.extraInfo, "处理共识投票结果", "保存经过离线共识的节点", "node", node.String())
		} else {
			log.Info(ts.extraInfo, "处理共识投票结果", "保存经过离线共识的节点", "node", node.String(), "已经在共识离线节点列表中", "")
		}
	} else {
		log.Info(ts.extraInfo, "处理共识投票结果", "保存经过离线共识的节点", "node", node.String(), "不在 electNode 中", "不保存")
	}
}

func (ts *topNodeState) getNodes(nodesOnlineStat []NodeOnLineInfo) []common.Address {
	nodes := make([]common.Address, 0)

	for _, value := range nodesOnlineStat {
		if value.Role == common.RoleValidator {
			nodes = append(nodes, value.Address)

		}
	}

	return nodes
}
func (ts *topNodeState) newTopNodeState(nodesOnlineInfo []NodeOnLineInfo, leader common.Address) (online, offline []common.Address) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	log.Debug(ts.extraInfo, "检查节点在线状态", "", "onlineNode Length", len(ts.onlineNode))
	for _, value := range nodesOnlineInfo {
		if value.Address.Equal(leader) {
			continue
		}
		if isInsideList(value.Address, ts.onlineNode) && (!isInsideList(value.Address, ts.consensusOff)) {
			log.Debug(ts.extraInfo, "检查节点在线状态", "在线", "节点", value.Address.String(), "onlineState", value.OnlineState)
			if isOffline(value.OnlineState) /*&& (!ts.isFinishedPropocal(value.Address, offLine))*/ {
				offline = append(offline, value.Address)
				log.Info(ts.extraInfo, "检查节点在线状态", "在线", "节点", value.Address.String(), "offline", "需要共识")

			} else {
				log.Debug(ts.extraInfo, "检查节点在线状态", "在线", "节点", value.Address.String(), "仍然online", "不需要共识")
			}
		} else {
			log.Debug(ts.extraInfo, "检查节点在线状态", "在线", "节点", value.Address.String(), "不在onlneNode中", "不检查在线状态")
		}
		if isInsideList(value.Address, ts.offlineNode) && (!isInsideList(value.Address, ts.consensusOn)) {
			if isOnline(value.OnlineState) /* && (!ts.isFinishedPropocal(value.Address, onLine))*/ {
				online = append(online, value.Address)
				log.Info(ts.extraInfo, "检查节点在线状态", "离线", "节点", value.Address.String(), "online", "需要共识")

			} else {
				log.Debug(ts.extraInfo, "检查节点在线状态", "离线", "节点", value.Address.String(), "仍然offline", "不需要共识")
			}
		} else {
			log.Debug(ts.extraInfo, "检查节点在线状态", "离线", "节点", value.Address.String(), "不在offlineNode中", "不检查在线状态")

		}
	}
	log.Info(ts.extraInfo, "检查节点在线状态", "", "online", online, "offline", offline)
	return
}
func (ts *topNodeState) checkAddressConsesusOnlineState(node common.Address, onlineState uint8) bool {
	propocaloff, _ := ts.finishedProposal.getVotes(getFinishedPropocalHash(node, offLine))
	propocalon, _ := ts.finishedProposal.getVotes(getFinishedPropocalHash(node, onLine))
	curState := uint8(offLine)
	curRound := uint64(math.MaxUint64)
	if propocaloff != nil {
		prop := propocaloff.(*mc.OnlineConsensusReq)
		curRound = prop.Seq
		curState = offLine
	}
	if propocalon != nil {
		prop := propocalon.(*mc.OnlineConsensusReq)
		if prop.Seq > curRound {
			curRound = prop.Seq
			curState = onLine
		}
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if curRound < uint64(math.MaxUint64) {
		log.Debug(ts.extraInfo, "区块验证", "", "curRount", curRound)
		return onlineState == curState
	} else if onlineState == offLine {
		if isInsideList(node, ts.consensusOff) {
			log.Debug(ts.extraInfo, "区块验证", "", "检查节点共识状态", "离线", "节点", node.String(), "是否在共识离线列表中",
				"true", "consensusOff", len(ts.consensusOff))

			return true
		}
		if isInsideList(node, ts.consensusOn) {
			log.Debug(ts.extraInfo, "区块验证", "", "检查节点共识状态", "离线", "节点", node.String(), "是否在共识离线列表中",
				"true", "consensusOn", len(ts.consensusOn))

			return false
		}
		log.Debug(ts.extraInfo, "区块验证", "", "检查节点共识状态", "离线", "节点", node.String(),
			"是否在离线列表中", isInsideList(node, ts.onlineNode), "offlineNode", len(ts.offlineNode))

		return isInsideList(node, ts.offlineNode)
	} else {
		if isInsideList(node, ts.consensusOn) {
			log.Debug(ts.extraInfo, "区块验证", "", "检查节点共识状态", "在线", "节点", node.String(), "是否在共识在线列表中",
				"true", "consensusOn", len(ts.consensusOn))

			return true
		} else {
			log.Debug(ts.extraInfo, "区块验证", "", "检查节点共识状态", "在线", "节点", node.String(), "是否在共识在线列表中",
				"false", "consensusOn", len(ts.consensusOn))

		}
		if isInsideList(node, ts.consensusOff) {
			return false
		}
		log.Debug(ts.extraInfo, "区块验证", "", "检查节点共识状态", "在线", "节点", node.String(), "是否在在线列表中",
			isInsideList(node, ts.onlineNode), "onlineNode", len(ts.onlineNode))

		return isInsideList(node, ts.onlineNode)
	}
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
		log.Info(ts.extraInfo, "item", item.Address.String())
		if item.Address == node {
			return isOffline(item.OnlineState)
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
