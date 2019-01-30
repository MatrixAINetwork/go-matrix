package txsreward

import (
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/reward"
	"github.com/matrix/go-matrix/reward/rewardexec"

	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/reward/cfg"
)

const (
	PackageName = "交易奖励"
)

type TxsReward struct {
	blockReward *rewardexec.BlockReward
}

func New(chain util.ChainReader, st util.StateDB) reward.Reward {

	data, err := matrixstate.GetTxsCalc(st)
	if nil != err {
		log.ERROR(PackageName, "获取状态树配置错误")
		return nil
	}

	if data == util.Stop {
		log.ERROR(PackageName, "停止发放区块奖励", "")
		return nil
	}

	TC, err := matrixstate.GetTxsRewardCfg(st)
	if nil != err || nil == TC {
		log.ERROR(PackageName, "获取状态树配置错误", err)
		return nil
	}

	rate := TC.RewardRate

	if util.RewardFullRate != TC.ValidatorsRate+TC.MinersRate {
		log.ERROR(PackageName, "交易费奖励比例配置错误", "")
		return nil
	}
	cfg := cfg.New(&mc.BlkRewardCfg{RewardRate: rate}, nil)
	cfg.ValidatorsRate = TC.ValidatorsRate
	cfg.MinersRate = TC.MinersRate
	return rewardexec.New(chain, cfg, st)
}
