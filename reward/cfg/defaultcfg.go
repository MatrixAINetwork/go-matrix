package cfg

import (
	"math/big"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/reward/leaderreward"
	"github.com/matrix/go-matrix/reward/mineroutreward"
	"github.com/matrix/go-matrix/reward/selectedreward"

	"github.com/matrix/go-matrix/core/types"

	"github.com/matrix/go-matrix/common"
)

const (
	PackageName = "奖励"

	//todo: 分母10000， 加法做参数检查
	MinersBlockRewardRate     = uint64(4000) //矿工网络奖励40%
	ValidatorsBlockRewardRate = uint64(5000) //验证者网络奖励50%
	FoundationBlockRewardRate = uint64(1000) //基金会网络奖励10%

	MinerOutRewardRate     = uint64(4000) //出块矿工奖励40%
	ElectedMinerRewardRate = uint64(6000) //当选矿工奖励60%

	LeaderRewardRate            = uint64(4000) //出块验证者（leader）奖励40%
	ElectedValidatorsRewardRate = uint64(6000) //当选验证者奖励60%

	OriginElectOfflineRewardRate = uint64(5000) //初选下线验证者奖励50%
	BackupRate                   = uint64(5000) //当前替补验证者奖励50%
)

type RewardMountCfg struct {
	MinersRate     uint64 //矿工网络奖励
	ValidatorsRate uint64 //验证者网络奖励
	FoundationRate uint64 //基金会网络奖励

	MinerOutRate    uint64 //出块矿工奖励
	ElectedMineRate uint64 //当选矿工奖励

	LeaderRate            uint64 //出块验证者（leader）奖励
	ElectedValidatorsRate uint64 //当选验证者奖励

	OriginElectOfflineRate uint64 //初选下线验证者奖励
	BackupRewardRate       uint64 //当前替补验证者奖励
}
type RewardCfg struct {
	RewardMount *RewardMountCfg
	SetReward   SetRewardsExec
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
type SetRewardsExec interface {
	SetLeaderRewards(reward *big.Int, rewards map[common.Address]*big.Int, Leader common.Address, num *big.Int)
	SetMinerOutRewards(reward *big.Int, chain ChainReader, num *big.Int, rewards map[common.Address]*big.Int)
	SetSelectedRewards(reward *big.Int, chain ChainReader, topRewards map[common.Address]*big.Int, roleType common.RoleType, header *types.Header, rate uint64) //todo 金额
}
type DefaultSetRewards struct {
	leader   leaderreward.LeaderReward
	miner    mineroutreward.MinerOutReward
	selected selectedreward.SelectedReward
}

func DefaultSetRewardNew() *DefaultSetRewards {
	return &DefaultSetRewards{
		leader:   leaderreward.LeaderReward{},
		miner:    mineroutreward.MinerOutReward{},
		selected: selectedreward.SelectedReward{},
	}

}

func (str *DefaultSetRewards) SetLeaderRewards(reward *big.Int, rewards map[common.Address]*big.Int, Leader common.Address, num *big.Int) {
	if common.IsBroadcastNumber(num.Uint64()) {
		return
	}
	str.leader.SetLeaderRewards(reward, rewards, Leader, num)
}
func (str *DefaultSetRewards) SetSelectedRewards(reward *big.Int, chain ChainReader, topRewards map[common.Address]*big.Int, roleType common.RoleType, header *types.Header, rate uint64) {
	str.selected.SetSelectedRewards(reward, chain, topRewards, roleType, header, rate)
}
func (str *DefaultSetRewards) SetMinerOutRewards(reward *big.Int, chain ChainReader, num *big.Int, rewards map[common.Address]*big.Int) {
	if common.IsBroadcastNumber(num.Uint64()-1) {
		return
	}
	str.miner.SetMinerOutRewards(reward, chain, num, rewards)
}

func New(RewardMount *RewardMountCfg, SetReward SetRewardsExec) *RewardCfg {

	//默认配置
	if nil == RewardMount {
		RewardMount = &RewardMountCfg{
			MinersRate:     MinersBlockRewardRate,
			ValidatorsRate: ValidatorsBlockRewardRate,
			FoundationRate: FoundationBlockRewardRate,

			MinerOutRate:    MinerOutRewardRate,
			ElectedMineRate: ElectedMinerRewardRate,

			LeaderRate:            LeaderRewardRate,
			ElectedValidatorsRate: ElectedValidatorsRewardRate,

			OriginElectOfflineRate: OriginElectOfflineRewardRate,
			BackupRewardRate:       BackupRate,
		}

	}
	//默认配置
	if nil == SetReward {

		SetReward = DefaultSetRewardNew()
	}

	return &RewardCfg{
		RewardMount: RewardMount,
		SetReward:   SetReward,
	}
}
