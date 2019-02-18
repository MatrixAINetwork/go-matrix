package multicoinreward

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/mc"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/reward"
	"github.com/MatrixAINetwork/go-matrix/reward/cfg"
	"github.com/MatrixAINetwork/go-matrix/reward/rewardexec"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

const (
	PackageName = "多币种奖励"

	//todo: 分母10000， 加法做参数检查
	ValidatorsTxsRewardRate = uint64(util.RewardFullRate) //验证者交易奖励比例100%
	MinerTxsRewardRate      = uint64(0)                   //矿工交易奖励比例0%
	FoundationTxsRewardRate = uint64(0)                   //基金会交易奖励比例0%

	MinerOutRewardRate     = uint64(4000) //出块矿工奖励40%
	ElectedMinerRewardRate = uint64(6000) //当选矿工奖励60%

	LeaderRewardRate            = uint64(4000) //出块验证者（leader）奖励40%
	ElectedValidatorsRewardRate = uint64(6000) //当选验证者奖励60%

	OriginElectOfflineRewardRate = uint64(util.RewardFullRate) //初选下线验证者奖励50%
	BackupRate                   = uint64(0)                   //当前替补验证者奖励50%

)

type MultiCoinReward struct {
	chain  util.ChainReader
	reward reward.Reward
}

func New(chain util.ChainReader) *MultiCoinReward {

	RewardMount := &mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationTxsRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationTxsRewardRate,
		OriginElectOfflineRate:  OriginElectOfflineRewardRate,
		BackupRewardRate:        BackupRate,
	}
	rewardCfg := cfg.New(RewardMount, nil)
	return &MultiCoinReward{chain, rewardexec.New(chain, rewardCfg)}
}

func (mcr *MultiCoinReward) CalcNodesReward(rewardMount *big.Int, currentnum uint64) map[common.Address]*big.Int {

	if !common.IsReElectionNumber(currentnum - 1) {
		return nil
	}
	var originNum uint64
	if currentnum < common.GetReElectionInterval() {
		originNum = 1
	} else {
		originNum = currentnum - common.GetReElectionInterval() - 2
	}

	rewardMap := make(map[common.Address]*big.Int)
	for originNum < currentnum-1 {
		header := mcr.chain.GetBlockByNumber(originNum).Header()
		tempMap := mcr.reward.CalcNodesRewards(rewardMount, header.Leader, header.Number.Uint64())
		for i, v := range tempMap {
			util.SetAccountRewards(rewardMap, i, v)
		}
		originNum++
	}

	log.INFO(PackageName, "多币种", rewardMap)
	return rewardMap
}
