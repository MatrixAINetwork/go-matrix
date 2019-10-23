// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package mineroutreward

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
	"github.com/pkg/errors"
)

var (
	AiReward = uint64(2e18) //2man
)

type MinerOutEpsilion struct {
	InnerMiners []common.Address
	Coinbase    common.Address
	RewardType  uint8
	MR          MinerOutReward
	RewardCfg   *mc.AIBlkRewardCfg
}

func (moe *MinerOutEpsilion) canSetMinerOutRewards(num uint64, reward *big.Int, bcInterval *mc.BCIntervalInfo, innerMiners []common.Address) (common.Address, error) {
	if num < 2 {
		log.Debug(PackageName, "高度为小于2 不发放奖励：", "")
		return common.Address{}, errors.New("高度为小于2 不发放奖励：")
	}

	if reward.Cmp(big.NewInt(0)) <= 0 {
		return common.Address{}, errors.New("奖励金额不合法")
	}

	coinbase := moe.Coinbase
	if coinbase.Equal(common.Address{}) {
		return common.Address{}, errors.New("矿工奖励的地址非法")
	}
	for _, v := range innerMiners {
		if coinbase.Equal(v) {
			log.Warn(PackageName, "基金会矿工不发钱，账户为", coinbase)
			return common.Address{}, errors.New("基金会矿工")
		}
	}
	return coinbase, nil
}

func (moe *MinerOutEpsilion) SetMinerOutRewards(AIReward, curReward *big.Int, state util.StateDB, num uint64, parentHash common.Hash, reader util.ChainReader, coinType string) map[common.Address]*big.Int {
	//当前块给当前矿工发放奖励
	bcInterval, err := matrixstate.GetBroadcastInterval(state)
	if err != nil {
		log.Error(PackageName, "获取广播周期失败", err)
		return nil
	}
	if bcInterval.IsBroadcastNumber(num) {
		log.Warn(PackageName, "广播区块不发钱：", num)
		return nil
	}
	//3个区块一个矿工奖励，奖励乘以3倍
	outReward := moe.getReward(curReward)
	rewards := moe.MR.SetMinerOutRewards(AIReward, outReward, state, num, parentHash, reader, coinType)
	//交易费不发放ai出矿工奖励

	outReward = moe.getReward(AIReward)
	coinBase, err := moe.canSetMinerOutRewards(num, outReward, bcInterval, moe.InnerMiners)
	if nil != err {
		return rewards
	}
	util.SetAccountRewards(rewards, coinBase, outReward)
	log.Debug(PackageName, "ai矿工账户", coinBase.String(), "发放奖励高度", num, "奖励金额", outReward)
	return rewards
}

func (moe *MinerOutEpsilion) getReward(reward *big.Int) *big.Int {
	if moe.RewardType == util.TxsReward {
		return reward
	}
	//3个区块一个矿工奖励，奖励乘以3倍
	outreward := new(big.Int).Mul(reward, new(big.Int).SetUint64(3))

	return outreward
}
