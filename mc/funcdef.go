// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	"strconv"

	"github.com/matrix/go-matrix/common"
	"github.com/pkg/errors"
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
	size := len(self.NodeList)
	for i := 0; i < size; i++ {
		topNode := &self.NodeList[i]
		if chgInfo.Position > topNode.Position {
			if chgInfo.Position == common.PosOffline && chgInfo.Account == topNode.Account {
				self.NodeList = append(self.NodeList[:i], self.NodeList[i+1:]...)
				return
			} else {
				continue
			}
		} else if chgInfo.Position == topNode.Position {
			if (chgInfo.Account == common.Address{}) {
				self.NodeList = append(self.NodeList[:i], self.NodeList[i+1:]...)
			} else {
				topNode.Account.Set(chgInfo.Account)
				topNode.NodeNumber = self.increaseNodeNumber()
			}
			return
		} else if chgInfo.Position < topNode.Position {
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
