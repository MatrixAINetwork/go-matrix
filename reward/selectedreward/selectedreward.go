// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package selectedreward

import (
	"errors"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/reward/util"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
)

const (
	PackageName = "参与奖励"
)

type SelectedReward struct {
}
type ChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeader retrieves a block header from the database by hash and number.
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	GetBlockByNumber(number uint64) *types.Block

	// GetBlock retrieves a block sfrom the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)
	State() (*state.StateDB, error)
	NewTopologyGraph(header *types.Header) (*mc.TopologyGraph, error)
}

func (sr *SelectedReward) GetTopAndDeposit(state util.StateDB, currentNum uint64, roleType common.RoleType, currentTop *mc.TopologyGraph, originElectNodes *mc.ElectGraph) ([]common.Address, map[common.Address]uint16, error) {

	if originElectNodes == nil || 0 == len(originElectNodes.ElectList) {
		log.Error(PackageName, "get获取初选列表为空", "")
		return nil, nil, errors.New("get获取初选列表为空")
	}

	if currentTop == nil || 0 == len(currentTop.NodeList) {
		log.Error(PackageName, "当前拓扑图是 空", "")
		return nil, nil, errors.New("当前拓扑图是 空")
	}

	topNodes := make([]common.Address, 0)
	for _, node := range currentTop.NodeList {
		if node.Type == node.Type&roleType {
			topNodes = append(topNodes, node.Account)
		}
	}

	electNodes := make(map[common.Address]uint16, 0)
	for _, node := range originElectNodes.ElectList {
		if node.Type == node.Type&roleType {
			electNodes[node.Account] = node.Stock
		}
	}

	return topNodes, electNodes, nil
}

func (sr *SelectedReward) GetSelectedRewards(reward *big.Int, state util.StateDB, roleType common.RoleType, currentNum uint64, rate uint64, topology *mc.TopologyGraph, elect *mc.ElectGraph) map[common.Address]*big.Int {

	//计算选举的拓扑图的高度
	if reward.Cmp(big.NewInt(0)) <= 0 {
		//log.WARN(PackageName, "奖励金额不合法", reward)
		return nil
	}
	//log.Debug(PackageName, "参与奖励大家共发放", reward)

	currentTop, originElectNodes, err := sr.GetTopAndDeposit(state, currentNum, roleType, topology, elect)
	if nil != err {
		return nil
	}

	selectedNodesDeposit := sr.CaclSelectedDeposit(currentTop, originElectNodes, rate)
	if 0 == len(selectedNodesDeposit) {
		log.Error(PackageName, "获取参与的抵押列表错误", "")
		return nil
	}

	return util.CalcStockRate(reward, selectedNodesDeposit)

}

func (sr *SelectedReward) CaclSelectedDeposit(newGraph []common.Address, originElectNodes map[common.Address]uint16, rewardRate uint64) map[common.Address]util.DepositInfo {
	NodesRewardMap := make(map[common.Address]uint64, 0)
	for _, nodelist := range newGraph {
		NodesRewardMap[nodelist] = rewardRate
		//log.Debug(PackageName, "当前节点", nodelist.Hex())
	}
	for electList := range originElectNodes {
		if _, ok := NodesRewardMap[electList]; ok {
			NodesRewardMap[electList] = util.RewardFullRate
		} else {
			NodesRewardMap[electList] = util.RewardFullRate - rewardRate
		}
		//log.Debug(PackageName, "初选节点", electList.Hex(), "比例", NodesRewardMap[electList])
	}

	selectedNodesDeposit := make(map[common.Address]util.DepositInfo, 0)

	for k, v := range NodesRewardMap {
		var finalStock uint64
		if stock, ok := originElectNodes[k]; ok {
			finalStock = uint64(stock) * v
		} else {
			//二级节点股权默认值为1
			finalStock = uint64(1) * v
		}
		selectedNodesDeposit[k] = util.DepositInfo{Deposit: big.NewInt(0), FixStock: finalStock}
		//log.Debug(PackageName, "计算抵押总额,账户", k.Hex(), "股权", finalStock)
	}
	return selectedNodesDeposit
}
