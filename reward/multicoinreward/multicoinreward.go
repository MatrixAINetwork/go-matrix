package multicoinreward

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/reward"
	"github.com/matrix/go-matrix/reward/cfg"
	"github.com/matrix/go-matrix/reward/rewardexec"
	"github.com/matrix/go-matrix/reward/util"
	"math/big"
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
BackupRate                   = uint64(0) //当前替补验证者奖励50%

)

type MultiCoinReward struct {
	chain util.ChainReader
	reward reward.Reward
}




func New(chain util.ChainReader) *MultiCoinReward {

	RewardMount := &cfg.RewardMountCfg{
		MinersRate:     MinerTxsRewardRate,
		ValidatorsRate: ValidatorsTxsRewardRate,

		MinerOutRate:     MinerOutRewardRate,
		ElectedMinerRate: ElectedMinerRewardRate,
		FoundationMinerRate: FoundationTxsRewardRate,

		LeaderRate:            LeaderRewardRate,
		ElectedValidatorsRate: ElectedValidatorsRewardRate,
		FoundationValidatorRate:FoundationTxsRewardRate,
		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}
	rewardCfg := cfg.New(RewardMount, nil)
	return &MultiCoinReward {chain,rewardexec.New(chain, rewardCfg)}
}

func (mcr *MultiCoinReward )CalcNodesReward(rewardMount *big.Int,currentnum uint64) map[common.Address]*big.Int{


	if !common.IsReElectionNumber(currentnum+1){
		return nil
	}
	var orignum uint64
	if currentnum < common.GetReElectionInterval() {
		orignum = 1
	} else {
		orignum = currentnum - common.GetReElectionInterval()
	}
	rewardMap := make(map[common.Address]*big.Int)
	for orignum < currentnum+1 {
		header := mcr.chain.GetBlockByNumber(orignum).Header()
		tempMap := mcr.reward.CalcNodesRewards(rewardMount, header.Leader, header)
		for i, v := range tempMap {
			util.SetAccountRewards(rewardMap, i, v)
		}
		orignum++
	}

	log.INFO(PackageName, "多币种", rewardMap)
	return rewardMap
}
