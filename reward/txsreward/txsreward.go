// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package txsreward

import (
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/reward"
	"github.com/MatrixAINetwork/go-matrix/reward/rewardexec"

	"github.com/MatrixAINetwork/go-matrix/reward/util"

	"github.com/MatrixAINetwork/go-matrix/reward/cfg"
)

const (
	PackageName = "交易奖励"
)

type TxsReward struct {
	blockReward *rewardexec.BlockReward
}

func New(chain util.ChainReader, st util.StateDB, preSt util.StateDB, ppreSt util.StateDB) reward.Reward {

	data, err := matrixstate.GetTxsCalc(preSt)
	if nil != err {
		log.Error(PackageName, "获取状态树配置错误")
		return nil
	}

	if data == util.Stop {
		log.Error(PackageName, "停止发放区块奖励", "")
		return nil
	}

	TC, err := matrixstate.GetTxsRewardCfg(preSt)
	if nil != err || nil == TC {
		log.Error(PackageName, "获取状态树配置错误", err)
		return nil
	}
	interval, err := matrixstate.GetBroadcastInterval(preSt)
	if err != nil {
		log.Error(PackageName, "获取广播周期失败", err)
		return nil
	}

	foundationAccount, err := matrixstate.GetFoundationAccount(preSt)
	if err != nil {
		log.Error(PackageName, "获取基金会账户数据失败", err)
		return nil
	}

	innerMinerAccounts, err := matrixstate.GetInnerMinerAccounts(ppreSt)
	if err != nil {
		log.Error(PackageName, "获取内部矿工账户数据失败", err)
		return nil
	}
	rate := TC.RewardRate

	if util.RewardFullRate != TC.ValidatorsRate+TC.MinersRate {
		log.Error(PackageName, "交易费奖励比例配置错误", "")
		return nil
	}
	currentTop, originElectNodes, err := chain.GetGraphByState(preSt)
	if err != nil {
		log.Error("PackageName", "获取拓扑图错误", err)
		return nil
	}
	preMiner, err := util.GetPreMinerReward(preSt, util.TxsReward)
	if err != nil {
		log.Error(PackageName, "获取前一个矿工奖励错误", err)
	}
	if data >= util.CalcEpsilon {
		cfg := cfg.EpsilionNew(&mc.AIBlkRewardCfg{RewardRate: mc.AIRewardRateCfg{
			MinerOutRate:        TC.RewardRate.MinerOutRate,
			AIMinerOutRate:      0,                                 //AI出块矿工奖励
			ElectedMinerRate:    TC.RewardRate.ElectedMinerRate,    //当选矿工奖励
			FoundationMinerRate: TC.RewardRate.FoundationMinerRate, //基金会网络奖励

			LeaderRate:              TC.RewardRate.LeaderRate,            //出块验证者（leader）奖励
			ElectedValidatorsRate:   TC.RewardRate.ElectedValidatorsRate, //当选验证者奖励
			FoundationValidatorRate: TC.RewardRate.FoundationMinerRate,   //基金会网络奖励

			OriginElectOfflineRate: TC.RewardRate.OriginElectOfflineRate, //初选下线验证者奖励
			BackupRewardRate:       TC.RewardRate.BackupRewardRate,       //当前替补验证者奖励

		}}, nil, preMiner, innerMinerAccounts, util.TxsReward, data)
		cfg.ValidatorsRate = TC.ValidatorsRate
		cfg.MinersRate = TC.MinersRate
		return rewardexec.AIBlockNew(chain, cfg, st, interval, foundationAccount, currentTop, originElectNodes)
	} else {
		cfg := cfg.New(&mc.BlkRewardCfg{RewardRate: rate}, nil, preMiner, innerMinerAccounts, util.TxsReward, data)
		cfg.ValidatorsRate = TC.ValidatorsRate
		cfg.MinersRate = TC.MinersRate
		return rewardexec.New(chain, cfg, st, interval, foundationAccount, currentTop, originElectNodes)
	}

}
