// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/pkg/errors"
	"strconv"
)

func NewGenesisTopologyGraph(genesisHeader *types.Header) (*TopologyGraph, error) {
	if genesisHeader.Number.Uint64() != 0 {
		return nil, errors.New("输入错误，创世区块高度不为0")
	}

	if genesisHeader.NetTopology.Type != common.NetTopoTypeAll {
		return nil, errors.New("输入错误，创世区块拓扑类型不是全拓扑")
	}

	newGraph := &TopologyGraph{
		Number:        0,
		NodeList:      make([]TopologyNodeInfo, 0),
		ElectList:     make([]TopologyNodeInfo, 0),
		CurNodeNumber: 99,
	}
	for _, topNode := range genesisHeader.NetTopology.NetTopologyData {
		newGraph.NodeList = append(newGraph.NodeList, TopologyNodeInfo{
			Account:    topNode.Account,
			Position:   topNode.Position,
			Type:       common.GetRoleTypeFromPosition(topNode.Position),
			Stock:      getNodeStock(topNode.Account, genesisHeader.Elect),
			NodeNumber: newGraph.increaseNodeNumber(),
		})
		newGraph.ElectList = append(newGraph.ElectList, TopologyNodeInfo{
			Account:  topNode.Account,
			Position: topNode.Position,
			Type:     common.GetRoleTypeFromPosition(topNode.Position),
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

func (self *TopologyGraph) GetAccountElectInfo(account common.Address) *TopologyNodeInfo {
	if len(self.ElectList) == 0 {
		return nil
	}
	for i := 0; i < len(self.ElectList); i++ {
		info := self.ElectList[i]
		if account == info.Account {
			return &info
		}
	}
	return nil
}

func (self *TopologyGraph) Transfer2NextGraph(number uint64, blockTopology *common.NetTopology, electList []common.Elect) (*TopologyGraph, error) {
	if self.Number+1 != number {
		return nil, errors.Errorf("高度不匹配,current(%d) + 1 != target(%d)", self.Number, number)
	}

	newGraph := &TopologyGraph{
		Number:        number,
		NodeList:      make([]TopologyNodeInfo, 0),
		ElectList:     make([]TopologyNodeInfo, 0),
		CurNodeNumber: self.CurNodeNumber,
	}

	switch blockTopology.Type {
	case common.NetTopoTypeAll:
		for _, topNode := range blockTopology.NetTopologyData {
			newGraph.NodeList = append(newGraph.NodeList, TopologyNodeInfo{
				Account:    topNode.Account,
				Position:   topNode.Position,
				Type:       common.GetRoleTypeFromPosition(topNode.Position),
				Stock:      getNodeStock(topNode.Account, electList),
				NodeNumber: newGraph.increaseNodeNumber(),
			})
		}
		newGraph.ElectList = append(newGraph.ElectList, newGraph.NodeList...)
		return newGraph, nil

	case common.NetTopoTypeChange:
		newGraph.NodeList = append(newGraph.NodeList, self.NodeList...)
		newGraph.ElectList = append(newGraph.ElectList, self.ElectList...)
		for _, chgInfo := range blockTopology.NetTopologyData {
			newGraph.modifyGraphByChgInfo(&chgInfo, electList)
			newGraph.modifyElectStateByChgInfo(&chgInfo)
		}
		return newGraph, nil

	default:
		return nil, errors.Errorf("生成验证者列表错误, 输入区块拓扑类型(%d)错误!", blockTopology.Type)
	}
}

func (self *TopologyGraph) modifyGraphByChgInfo(chgInfo *common.NetTopologyData, electList []common.Elect) {
	size := len(self.NodeList)
	for i := 0; i < size; i++ {
		topNode := &self.NodeList[i]
		if chgInfo.Position > topNode.Position {
			if chgInfo.Position == common.PosOffline && chgInfo.Account == topNode.Account {
				self.NodeList = append(self.NodeList[:i], self.NodeList[i+1:]...)
				return
			}
		} else if chgInfo.Position == topNode.Position {
			if (chgInfo.Account == common.Address{}) {
				self.NodeList = append(self.NodeList[:i], self.NodeList[i+1:]...)
			} else {
				topNode.Account.Set(chgInfo.Account)
				topNode.Stock = getNodeStock(topNode.Account, electList)
				topNode.NodeNumber = self.increaseNodeNumber()
			}
			return
		} else if chgInfo.Position < topNode.Position {
			newNode := TopologyNodeInfo{
				Account:    chgInfo.Account,
				Position:   chgInfo.Position,
				Type:       common.GetRoleTypeFromPosition(chgInfo.Position),
				Stock:      getNodeStock(topNode.Account, electList),
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

func (self *TopologyGraph) modifyElectStateByChgInfo(chgInfo *common.NetTopologyData) {
	size := len(self.ElectList)
	for i := 0; i < size; i++ {
		eleNode := &self.ElectList[i]
		if chgInfo.Position == common.PosOffline || chgInfo.Position == common.PosOnline {
			if chgInfo.Account == eleNode.Account {
				eleNode.Position = chgInfo.Position
				return
			}
		} else {
			if chgInfo.Position == eleNode.Position && chgInfo.Account != eleNode.Account {
				// 说明该位置的选举节点掉线了，别的节点顶替了
				eleNode.Position = common.PosOffline
				return
			}
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

func getNodeStock(addr common.Address, electList []common.Elect) uint16 {
	for _, electInfo := range electList {
		if electInfo.Account == addr {
			return electInfo.Stock
		}
	}

	return 1
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
