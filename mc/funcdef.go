// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/pkg/errors"
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
	}
	return newGraph, nil
}

func (self *TopologyGraph) Transfer2NextGraph(number uint64, blockTopology *common.NetTopology, electList []common.Elect) (*TopologyGraph, error) {
	if self.Number+1 != number {
		return nil, errors.Errorf("高度不匹配,current(%d) + 1 != target(%d)", self.Number, number)
	}

	newGraph := &TopologyGraph{
		Number:        number,
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
				Stock:      getNodeStock(topNode.Account, electList),
				NodeNumber: newGraph.increaseNodeNumber(),
			})
		}
		return newGraph, nil

	case common.NetTopoTypeChange:
		newGraph.NodeList = append(newGraph.NodeList, self.NodeList...)
		for _, chgInfo := range blockTopology.NetTopologyData {
			size := len(newGraph.NodeList)
			for i := 0; i < size; i++ {
				topNode := &newGraph.NodeList[i]
				if chgInfo.Position == topNode.Position {
					topNode.Account.Set(chgInfo.Account)
					topNode.Stock = getNodeStock(topNode.Account, electList)
					topNode.NodeNumber = newGraph.increaseNodeNumber()
					break
				}

				if chgInfo.Position == common.PosOffline && chgInfo.Account == topNode.Account {
					newGraph.NodeList = append(newGraph.NodeList[:i], newGraph.NodeList[i+1:]...)
					break
				}
			}
		}
		return newGraph, nil

	default:
		return nil, errors.Errorf("生成验证者列表错误, 输入区块拓扑类型(%d)错误!", blockTopology.Type)
	}
}

func getNodeStock(addr common.Address, electList []common.Elect) uint16 {
	for _, electInfo := range electList {
		if electInfo.Account == addr {
			return electInfo.Stock
		}
	}

	return 1
}

func (self *TopologyGraph) increaseNodeNumber() uint8 {
	if self.CurNodeNumber >= 99 {
		self.CurNodeNumber = 0
	} else {
		self.CurNodeNumber++
	}

	return self.CurNodeNumber
}
