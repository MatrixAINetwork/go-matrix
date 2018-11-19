// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
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
		newGraph.ElectList = append(newGraph.ElectList, TopologyNodeInfo{
			Account:  topNode.Account,
			Position: topNode.Position,
			Type:     common.GetRoleTypeFromPosition(topNode.Position),
		})
	}
	return newGraph, nil
}
func checkInGraph(top []TopologyNodeInfo, pos uint16) common.Address {
	for _, v := range top {
		if v.Position == pos {
			return v.Account
		}
	}
	return common.Address{}
}
func checkInDiff(diff *common.NetTopology, add common.Address) bool {
	for _, v := range diff.NetTopologyData {
		if v.Account == add {
			return true
		}
	}
	return false
}

func ParseTopNodeOffline(topologyChg *common.NetTopology, prevTopology []TopologyNodeInfo) []common.Address {
	if topologyChg.Type != common.NetTopoTypeChange {
		return nil
	}

	offline := make([]common.Address, 0)

	for _, v := range topologyChg.NetTopologyData {
		if v.Position == common.PosOffline || v.Position == common.PosOnline {
			continue
		}
		log.Info("zzzzwww-验证者计算tiopnode阶段", "Positiom", v.Position, "account", v.Account)
		account := checkInGraph(prevTopology, v.Position)
		log.INFO("zzzzwww-验证者计算tiopnode阶段", "account", account)
		log.INFO("zzzzwww-验证者计算tiopnode阶段", "topologyChg", topologyChg)

		if checkInDiff(topologyChg, account) == false {
			offline = append(offline, account)
		}

	}
	return offline
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
			size := len(newGraph.NodeList)
			for i := 0; i < size; i++ {
				topNode := &newGraph.NodeList[i]
				if chgInfo.Position > topNode.Position {
					if chgInfo.Position == common.PosOffline && chgInfo.Account == topNode.Account {
						newGraph.NodeList = append(newGraph.NodeList[:i], newGraph.NodeList[i+1:]...)
						break
					}
					continue
				} else if chgInfo.Position == topNode.Position {
					if (chgInfo.Account == common.Address{}) {
						newGraph.NodeList = append(newGraph.NodeList[:i], newGraph.NodeList[i+1:]...)
						break
					}
					topNode.Account.Set(chgInfo.Account)
					topNode.Stock = getNodeStock(topNode.Account, electList)
					topNode.NodeNumber = newGraph.increaseNodeNumber()
					break
				} else if chgInfo.Position < topNode.Position {
					newNode := TopologyNodeInfo{
						Account:    chgInfo.Account,
						Position:   chgInfo.Position,
						Type:       common.GetRoleTypeFromPosition(chgInfo.Position),
						Stock:      getNodeStock(topNode.Account, electList),
						NodeNumber: newGraph.increaseNodeNumber(),
					}
					//todo test newNode插入切片I位置
					rear := append([]TopologyNodeInfo{}, newGraph.NodeList[i:]...)
					newGraph.NodeList = append(newGraph.NodeList[:i], newNode)
					newGraph.NodeList = append(newGraph.NodeList, rear...)
					break
				}
			}

			size = len(newGraph.ElectList)
			for i := 0; i < size; i++ {
				EleNode := &newGraph.ElectList[i]
				if chgInfo.Position == common.PosOffline && chgInfo.Account == EleNode.Account {
					newGraph.ElectList = append(newGraph.ElectList[:i], newGraph.ElectList[i+1:]...)
					break
				}
			}
			if chgInfo.Position == common.PosOnline {
				newNode := TopologyNodeInfo{
					Account:    chgInfo.Account,
					Position:   chgInfo.Position,
					Type:       common.GetRoleTypeFromPosition(chgInfo.Position),
					Stock:      getNodeStock(chgInfo.Account, electList),
					NodeNumber: newGraph.increaseNodeNumber(),
				}
				newGraph.ElectList = append(newGraph.ElectList, newNode)
			}

		}

		preTop := self.NodeList

		offline := ParseTopNodeOffline(blockTopology, preTop)
		log.INFO("funcdef.go", "offline节点", offline, "blockTopology", blockTopology, "preTop", preTop)
		for k, v := range newGraph.ElectList {
			flag := 0
			for _, vv := range offline {
				if vv == v.Account {
					flag = 1
				}
			}
			if flag == 1 {
				log.INFO("funcdef.go", "在ElectList中删除该节点", v.Account.String())
				newGraph.ElectList = append(newGraph.ElectList[:k], newGraph.ElectList[k+1:]...)
			}
		}

		log.Info("拓扑写进数据库", "number", newGraph.Number, "curNumber", newGraph.CurNodeNumber)
		for _, value := range newGraph.NodeList {
			log.Info("拓扑写进数据库", "NodeList", value.Account.String())
		}
		for _, value := range newGraph.ElectList {
			log.Info("拓扑写进数据库", "ElectList", value.Account.String())
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
