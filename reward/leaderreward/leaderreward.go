package leaderreward

import (
	"math/big"

	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
)

const (
	PackageName = "leader奖励"
)

type LeaderReward struct {
}

func (lr *LeaderReward) SetLeaderRewards(reward *big.Int, rewards map[common.Address]*big.Int, Leader common.Address, num *big.Int) {
	//广播区块不给验证者发钱
	if common.IsBroadcastNumber(num.Uint64()) {
		return
	}
	util.SetAccountRewards(rewards, Leader, reward)
	log.INFO(PackageName, "leader reward addr", Leader, "reward", reward)
}
