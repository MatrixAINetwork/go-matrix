// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderreward

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/reward/util"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

type LeaderEpsilionReward struct {
	Coinbase common.Address
}

const (
	RewardNoCoinbaseRate = uint64(5000)
)

func (lr *LeaderEpsilionReward) SetLeaderRewards(reward *big.Int, Leader common.Address, num uint64) map[common.Address]*big.Int {
	//广播区块不给验证者发钱
	log.Info(PackageName, "使用 Epsilion leader 计算引擎,矿工地址", lr.Coinbase, "leader地址", Leader)
	if manparams.IsBroadcastNumber(num, num-1) {
		return nil
	}
	if reward.Cmp(big.NewInt(0)) <= 0 {
		log.Warn(PackageName, "奖励金额不合法", reward)
		return nil
	}
	if Leader.Equal(common.Address{}) {
		log.Error(PackageName, "奖励的地址非法", Leader.Hex())
		return nil
	}
	if lr.Coinbase.Equal(common.Address{}) {
		reward = reward.Mul(reward, new(big.Int).SetUint64(RewardNoCoinbaseRate))
		reward = reward.Div(reward, new(big.Int).SetUint64(util.RewardFullRate))

		log.Info(PackageName, "无矿工结果，leader奖励衰减", reward.String())
	}
	rewards := make(map[common.Address]*big.Int)
	rewards[Leader] = reward
	//log.Debug(PackageName, "leader 奖励地址", Leader, "奖励金额", reward)
	return rewards
}
