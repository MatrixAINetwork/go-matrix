// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderreward

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

const (
	PackageName = "leader奖励"
)

type LeaderReward struct {
}

func (lr *LeaderReward) SetLeaderRewards(reward *big.Int, Leader common.Address, num uint64) map[common.Address]*big.Int {
	//广播区块不给验证者发钱
	if manparams.IsBroadcastNumber(num, num-1) {
		return nil
	}
	if reward.Cmp(big.NewInt(0)) <= 0 {
		log.WARN(PackageName, "奖励金额不合法", reward)
		return nil
	}
	if Leader.Equal(common.Address{}) {
		log.ERROR(PackageName, "奖励的地址非法", Leader.Hex())
		return nil
	}
	rewards := make(map[common.Address]*big.Int)
	rewards[Leader] = reward
	//log.Debug(PackageName, "leader 奖励地址", Leader, "奖励金额", reward)
	return rewards
}
