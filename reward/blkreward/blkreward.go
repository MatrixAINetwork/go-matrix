// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkreward

import (
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/reward"
	"github.com/MatrixAINetwork/go-matrix/reward/cfg"
	"github.com/MatrixAINetwork/go-matrix/reward/rewardexec"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

type blkreward struct {
	blockReward *rewardexec.BlockReward
	state       util.StateDB
}

func New(chain util.ChainReader, st util.StateDB, preSt util.StateDB, ppreSt util.StateDB) reward.Reward {
	data, err := matrixstate.GetBlkCalc(preSt)
	if nil != err {
		log.ERROR("固定区块奖励", "获取状态树配置错误")
		return nil
	}

	if data == util.Stop {
		log.ERROR("固定区块奖励", "停止发放区块奖励", "")
		return nil
	}
	RC, err := matrixstate.GetBlkRewardCfg(preSt)
	if nil != err || nil == RC {
		log.ERROR("固定区块奖励", "获取状态树配置错误", err)
		return nil
	}
	interval, err := matrixstate.GetBroadcastInterval(preSt)
	if err != nil {
		log.ERROR("固定区块奖励", "获取广播周期失败", err)
		return nil
	}

	foundationAccount, err := matrixstate.GetFoundationAccount(preSt)
	if err != nil {
		log.ERROR("固定区块奖励", "获取基金会账户数据失败", err)
		return nil
	}
	innerMinerAccounts, err := matrixstate.GetInnerMinerAccounts(ppreSt)
	if err != nil {
		log.ERROR("固定区块奖励", "获取内部矿工账户数据失败", err)
		return nil
	}

	currentTop, originElectNodes, err := chain.GetGraphByState(preSt)
	if err != nil {
		log.Error("固定区块奖励", "获取拓扑图错误", err)
		return nil
	}
	preMiner, err := util.GetPreMinerReward(preSt, util.BlkReward)
	if err != nil {
		log.Error("固定区块奖励", "获取前一个矿工奖励错误", err)
		return nil
	}
	rewardCfg := cfg.New(RC, nil, preMiner, innerMinerAccounts, util.BlkReward, data)
	return rewardexec.New(chain, rewardCfg, st, interval, foundationAccount, currentTop, originElectNodes)
}

//func (tr *blkreward) CalcNodesRewards(blockReward *big.Int, Leader common.Address, header *types.Header) map[common.Address]*big.Int {
//	return tr.blockReward.CalcNodesRewards(blockReward, Leader, header)
//}
