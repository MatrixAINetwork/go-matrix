// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package cfg

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/mc"

	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/reward/leaderreward"
	"github.com/MatrixAINetwork/go-matrix/reward/mineroutreward"
	"github.com/MatrixAINetwork/go-matrix/reward/selectedreward"

	"github.com/MatrixAINetwork/go-matrix/core/types"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

const (
	PackageName = "奖励"

	MinersBlockRewardRate     = uint64(5000) //矿工网络奖励50%
	ValidatorsBlockRewardRate = uint64(5000) //验证者网络奖励50%

	MinerOutRewardRate        = uint64(4000) //出块矿工奖励40%
	ElectedMinerRewardRate    = uint64(5000) //当选矿工奖励50%
	FoundationMinerRewardRate = uint64(1000) //基金会网络奖励10%

	LeaderRewardRate               = uint64(4000) //出块验证者（leader）奖励40%
	ElectedValidatorsRewardRate    = uint64(5000) //当选验证者奖励60%
	FoundationValidatorsRewardRate = uint64(1000) //基金会网络奖励10%

	OriginElectOfflineRewardRate = uint64(5000) //初选下线验证者奖励50%
	BackupRate                   = uint64(5000) //当前替补验证者奖励50%
)

type RewardStateCfg struct {
	MinersRate     uint64 //矿工网络奖励
	ValidatorsRate uint64 //验证者网络奖励

	MinerOutRate        uint64 //出块矿工奖励
	ElectedMinerRate    uint64 //当选矿工奖励
	FoundationMinerRate uint64 //基金会网络奖励

	LeaderRate              uint64 //出块验证者（leader）奖励
	ElectedValidatorsRate   uint64 //当选验证者奖励
	FoundationValidatorRate uint64 //基金会网络奖励

	OriginElectOfflineRate uint64 //初选下线验证者奖励
	BackupRewardRate       uint64 //当前替补验证者奖励
}

type RewardCfg struct {
	Calc           string
	MinersRate     uint64 //矿工网络奖励
	ValidatorsRate uint64 //验证者网络奖励
	RewardMount    *mc.BlkRewardCfg
	SetReward      SetRewardsExec
}
type ChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeader retrieves a block header from the database by hash and number.
	GetHeader(hash []common.CoinRoot, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash []common.CoinRoot) *types.Header

	GetBlockByNumber(number uint64) *types.Block

	// GetBlock retrieves a block sfrom the database by hash and number.
	GetBlock(hash []common.CoinRoot, number uint64) *types.Block
	StateAt(root []common.CoinRoot) (*state.StateDBManage, error)
	State() (*state.StateDBManage, error)
}
type SetRewardsExec interface {
	SetLeaderRewards(reward *big.Int, Leader common.Address, num uint64) map[common.Address]*big.Int
	SetMinerOutRewards(reward *big.Int, state util.StateDB, chain util.ChainReader, num uint64, parentHash common.Hash, coinType string) map[common.Address]*big.Int
	GetSelectedRewards(reward *big.Int, state util.StateDB, roleType common.RoleType, number uint64, rate uint64, topology *mc.TopologyGraph, elect *mc.ElectGraph) map[common.Address]*big.Int //todo 金额
}
type DefaultSetRewards struct {
	leader   leaderreward.LeaderReward
	miner    mineroutreward.MinerOutReward
	selected selectedreward.SelectedReward
}

func DefaultSetRewardNew(preMiner []mc.MultiCoinMinerOutReward, innerMiners []common.Address, rewardType uint8) *DefaultSetRewards {
	return &DefaultSetRewards{
		leader:   leaderreward.LeaderReward{},
		miner:    mineroutreward.MinerOutReward{PreReward: preMiner, InnerMiners: innerMiners, RewardType: rewardType},
		selected: selectedreward.SelectedReward{},
	}

}

func (str *DefaultSetRewards) SetLeaderRewards(reward *big.Int, Leader common.Address, num uint64) map[common.Address]*big.Int {

	return str.leader.SetLeaderRewards(reward, Leader, num)
}
func (str *DefaultSetRewards) GetSelectedRewards(reward *big.Int, state util.StateDB, roleType common.RoleType, number uint64, rate uint64, topology *mc.TopologyGraph, elect *mc.ElectGraph) map[common.Address]*big.Int {

	return str.selected.GetSelectedRewards(reward, state, roleType, number, rate, topology, elect)
}
func (str *DefaultSetRewards) SetMinerOutRewards(reward *big.Int, state util.StateDB, chain util.ChainReader, num uint64, parentHash common.Hash, coinType string) map[common.Address]*big.Int {

	return str.miner.SetMinerOutRewards(reward, state, num, parentHash, chain, coinType)
}

func New(RewardMount *mc.BlkRewardCfg, SetReward SetRewardsExec, preMiner []mc.MultiCoinMinerOutReward, innerMiners []common.Address, rewardType uint8, calc string) *RewardCfg {
	//默认配置
	if nil == SetReward {

		SetReward = DefaultSetRewardNew(preMiner, innerMiners, rewardType)
	}

	return &RewardCfg{
		Calc:        calc,
		RewardMount: RewardMount,
		SetReward:   SetReward,
	}
}
