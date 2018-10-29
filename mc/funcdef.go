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
