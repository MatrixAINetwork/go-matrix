// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	"strconv"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/pkg/errors"
	"sort"
)

func NewGenesisTopologyGraph(number uint64, netTopology common.NetTopology) (*TopologyGraph, error) {
	if number != 0 {
		return nil, errors.New("输入错误，创世区块高度不为0")
	}

	if netTopology.Type != common.NetTopoTypeAll {
		return nil, errors.New("输入错误，创世区块拓扑类型不是全拓扑")
	}

	newGraph := &TopologyGraph{
		NodeList:      make([]TopologyNodeInfo, 0),
		CurNodeNumber: 99,
	}
	for _, topNode := range netTopology.NetTopologyData {
		newGraph.NodeList = append(newGraph.NodeList, TopologyNodeInfo{
			Account:    topNode.Account,
			Position:   topNode.Position,
			Type:       common.GetRoleTypeFromPosition(topNode.Position),
			NodeNumber: newGraph.increaseNodeNumber(),
		})
	}
	newGraph.sort()
	return newGraph, nil
}

func (self *TopologyGraph) AccountIsInGraph(account common.Address) bool {
	if len(self.NodeList) == 0 {
		return false
	}
	for _, one := range self.NodeList {
		if account == one.Account {
			return true
		}
	}
	return false
}

func (self *TopologyGraph) CheckAccountRole(account common.Address, role common.RoleType) bool {
	if len(self.NodeList) == 0 {
		return false
	}
	for _, one := range self.NodeList {
		if account == one.Account {
			return one.Type == role
		}
	}
	return false
}

func (self *TopologyGraph) FindNextValidator(account common.Address) common.Address {
	validators := make([]common.Address, 0)
	for _, node := range self.NodeList {
		if node.Type == common.RoleValidator {
			validators = append(validators, node.Account)
		}
	}

	pos := -1
	size := len(validators)
	for i := 0; i < size; i++ {
		if account == validators[i] {
			pos = i
			break
		}
	}
	if pos == -1 {
		return common.Address{}
	}
	return validators[(pos+1)%size]
}

func (self *TopologyGraph) Transfer2NextGraph(number uint64, blockTopology *common.NetTopology) (*TopologyGraph, error) {

	newGraph := &TopologyGraph{
		NodeList:      make([]TopologyNodeInfo, 0),
		CurNodeNumber: self.CurNodeNumber,
	}

	switch blockTopology.Type {
	case common.NetTopoTypeAll:
		for _, topNode := range blockTopology.NetTopologyData {
			newGraph.NodeList = append(newGraph.NodeList, TopologyNodeInfo{
				Account:    topNode.Account,
				Position:   topNode.Position,
				Type:       common.GetRoleTypeFromPosition(topNode.Position),
				NodeNumber: newGraph.increaseNodeNumber(),
			})
		}
		newGraph.sort()
		return newGraph, nil

	case common.NetTopoTypeChange:
		newGraph.NodeList = append(newGraph.NodeList, self.NodeList...)
		for _, chgInfo := range blockTopology.NetTopologyData {
			newGraph.modifyGraphByChgInfo(&chgInfo)
		}
		return newGraph, nil

	default:
		return nil, errors.Errorf("生成验证者列表错误, 输入区块拓扑类型(%d)错误!", blockTopology.Type)
	}
}

func (self *TopologyGraph) modifyGraphByChgInfo(chgInfo *common.NetTopologyData) {
	// 上线节点，不处理
	if chgInfo.Position == common.PosOnline {
		return
	}

	size := len(self.NodeList)
	// 节点下线，从拓扑图中删除
	if chgInfo.Position == common.PosOffline {
		for i := 0; i < size; i++ {
			curNode := &self.NodeList[i]
			if chgInfo.Account == curNode.Account {
				self.NodeList = append(self.NodeList[:i], self.NodeList[i+1:]...)
				return
			}
		}
		return
	}

	// 位置替换信息处理
	if chgInfo.Position > self.NodeList[size-1].Position {
		// 变化位置，比当前最大位置还要大，将节点添加入队尾
		newNode := TopologyNodeInfo{
			Account:    chgInfo.Account,
			Position:   chgInfo.Position,
			Type:       common.GetRoleTypeFromPosition(chgInfo.Position),
			NodeNumber: self.increaseNodeNumber(),
		}
		self.NodeList = append(self.NodeList, newNode)
		return
	}
	for i := 0; i < size; i++ {
		curNode := &self.NodeList[i]
		if chgInfo.Position == curNode.Position {
			if (chgInfo.Account == common.Address{}) {
				self.NodeList = append(self.NodeList[:i], self.NodeList[i+1:]...)
			} else {
				curNode.Account.Set(chgInfo.Account)
				curNode.NodeNumber = self.increaseNodeNumber()
			}
			return
		} else if chgInfo.Position < curNode.Position {
			newNode := TopologyNodeInfo{
				Account:    chgInfo.Account,
				Position:   chgInfo.Position,
				Type:       common.GetRoleTypeFromPosition(chgInfo.Position),
				NodeNumber: self.increaseNodeNumber(),
			}
			//newNode插入切片I位置
			rear := append([]TopologyNodeInfo{}, self.NodeList[i:]...)
			self.NodeList = append(self.NodeList[:i], newNode)
			self.NodeList = append(self.NodeList, rear...)
			return
		}
	}
}

func (self *TopologyGraph) sort() {
	sort.Slice(self.NodeList, func(i, j int) bool {
		return self.NodeList[i].Position < self.NodeList[j].Position
	})
}

func (self *TopologyGraph) increaseNodeNumber() uint8 {
	if self.CurNodeNumber >= 99 {
		self.CurNodeNumber = 0
	} else {
		self.CurNodeNumber++
	}

	return self.CurNodeNumber
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
func (eg *ElectGraph) TransferElect2CommonElect() []common.Elect {
	size := len(eg.ElectList)
	rst := make([]common.Elect, size, size)
	for i := 0; i < size; i++ {
		rst[i].Account = eg.ElectList[i].Account
		rst[i].Stock = eg.ElectList[i].Stock
		rst[i].Type = eg.ElectList[i].Type.Transfer2ElectRole()
	}
	return rst
}

func (eg *ElectGraph) TransferNextElect2CommonElect() []common.Elect {
	nextElect := []common.Elect{}
	lenM := len(eg.NextMinerElect)
	lenV := len(eg.NextValidatorElect)
	for index := 0; index < lenM; index++ {
		nextElect = append(nextElect, common.Elect{
			Account: eg.NextMinerElect[index].Account,
			Stock:   eg.NextMinerElect[index].Stock,
			Type:    eg.NextMinerElect[index].Type.Transfer2ElectRole(),
		})
	}
	for index := 0; index < lenV; index++ {
		nextElect = append(nextElect, common.Elect{
			Account: eg.NextValidatorElect[index].Account,
			Stock:   eg.NextValidatorElect[index].Stock,
			Type:    eg.NextValidatorElect[index].Type.Transfer2ElectRole(),
		})
	}
	return nextElect
}

func (eos *ElectOnlineStatus) FindNodeElectOnlineState(node common.Address) *ElectNodeInfo {
	for _, elect := range eos.ElectOnline {
		if elect.Account == node {
			return &elect
		}
	}
	return nil
}

func (msg *HD_OnlineConsensusVoteResultMsg) IsValidity(curNumber uint64, validityTime uint64) bool {
	if msg.Req == nil {
		return false
	}
	return curNumber-msg.Req.Number <= validityTime
}

func (os OnlineState) String() string {
	switch os {
	case OnLine:
		return "OnLine"
	case OffLine:
		return "OffLine"
	default:
		return strconv.Itoa(int(os))
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
func (info *ConsensusTurnInfo) String() string {
	return strconv.Itoa(int(info.TotalTurns())) + "[" + strconv.Itoa(int(info.PreConsensusTurn)) + "," + strconv.Itoa(int(info.UsedReelectTurn)) + "]"
}

func (info *ConsensusTurnInfo) TotalTurns() uint32 {
	return info.PreConsensusTurn + info.UsedReelectTurn
}

// if < target, return -1
// if == target, return 0
// if > target, return 1
func (info *ConsensusTurnInfo) Cmp(target ConsensusTurnInfo) int64 {
	if *info == target {
		return 0
	}

	if info.TotalTurns() < target.TotalTurns() {
		return -1
	} else if info.TotalTurns() > target.TotalTurns() {
		return 1
	} else {
		if info.PreConsensusTurn < target.PreConsensusTurn {
			return -1
		} else {
			return 1
		}
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
func (req *HD_BlkConsensusReqMsg) TxsCodeCount() int {
	txsCodeCount := 0
	for _, item := range req.TxsCode {
		txsCodeCount += len(item.ListN)
	}
	return txsCodeCount
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
// self  > param: return 1
// self == param: return 0
// self  < param: return -1
func (self *ChainState) Cmp(superSeq uint64, curNumber uint64) int {
	if self.superSeq > superSeq {
		return 1
	} else if self.superSeq < superSeq {
		return -1
	} else {
		if self.curNumber > curNumber {
			return 1
		} else if self.curNumber < curNumber {
			return -1
		} else {
			return 0
		}
	}
}

func (self *ChainState) CurNumber() uint64 {
	return self.curNumber
}

func (self *ChainState) SuperSeq() uint64 {
	return self.superSeq
}

func (self *ChainState) Reset(superSeq uint64, curNumber uint64) {
	self.superSeq = superSeq
	self.curNumber = curNumber
}
