// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package cfg

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/reward/leaderreward"
	"github.com/MatrixAINetwork/go-matrix/reward/mineroutreward"
	"github.com/MatrixAINetwork/go-matrix/reward/selectedreward"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

type EpsilionSetRewards struct {
	leader   leaderreward.LeaderReward
	miner    mineroutreward.MinerOutEpsilion
	selected selectedreward.SelectedReward
}

func EpsilionNew(RewardMount *mc.AIBlkRewardCfg, SetReward SetRewardsExec, preMiner []mc.MultiCoinMinerOutReward, innerMiners []common.Address, rewardType uint8, calc string) *AIRewardCfg {
	//默认配置
	if nil == SetReward {

		SetReward = DefaultSetRewardNew(preMiner, innerMiners, rewardType)
	}

	return &AIRewardCfg{
		RewardType:  rewardType,
		Calc:        calc,
		RewardMount: RewardMount,
		SetReward:   SetReward,
	}
}

func EpsilionSetRewardNew(preMiner []mc.MultiCoinMinerOutReward, innerMiners []common.Address, coinbase common.Address, rewardType uint8, rewardCfg *mc.AIBlkRewardCfg) *EpsilionSetRewards {
	return &EpsilionSetRewards{
		leader:   leaderreward.LeaderReward{},
		miner:    mineroutreward.MinerOutEpsilion{MR: mineroutreward.MinerOutReward{InnerMiners: innerMiners, RewardType: rewardType, PreReward: preMiner}, InnerMiners: innerMiners, RewardType: rewardType, Coinbase: coinbase, RewardCfg: rewardCfg},
		selected: selectedreward.SelectedReward{},
	}

}

func (str *EpsilionSetRewards) SetLeaderRewards(reward *big.Int, Leader common.Address, num uint64) map[common.Address]*big.Int {

	return str.leader.SetLeaderRewards(reward, Leader, num)
}
func (str *EpsilionSetRewards) GetSelectedRewards(reward *big.Int, state util.StateDB, roleType common.RoleType, number uint64, rate uint64, topology *mc.TopologyGraph, elect *mc.ElectGraph) map[common.Address]*big.Int {

	return str.selected.GetSelectedRewards(reward, state, roleType, number, rate, topology, elect)
}
func (str *EpsilionSetRewards) SetMinerOutRewards(airReward, reward *big.Int, state util.StateDB, chain util.ChainReader, num uint64, parentHash common.Hash, coinType string) map[common.Address]*big.Int {

	return str.miner.SetMinerOutRewards(airReward, reward, state, num, parentHash, chain, coinType)
}
