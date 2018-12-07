package rewardexec

import (
	"github.com/matrix/go-matrix/params/manparams"
	"math/big"

	"github.com/matrix/go-matrix/reward/cfg"
	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
)

const (
	PackageName = "奖励"
)

type BlockReward struct {
	chain     util.ChainReader
	rewardCfg *cfg.RewardCfg
}

func New(chain util.ChainReader, rewardCfg *cfg.RewardCfg) *BlockReward {

	if util.RewardFullRate != rewardCfg.RewardMount.MinersRate+rewardCfg.RewardMount.ValidatorsRate+rewardCfg.RewardMount.FoundationRate {
		return nil
	}
	if util.RewardFullRate != rewardCfg.RewardMount.MinerOutRate+rewardCfg.RewardMount.ElectedMineRate {
		return nil
	}
	if util.RewardFullRate != rewardCfg.RewardMount.LeaderRate+rewardCfg.RewardMount.ElectedValidatorsRate {
		return nil
	}

	if util.RewardFullRate != rewardCfg.RewardMount.OriginElectOfflineRate+rewardCfg.RewardMount.BackupRewardRate {
		return nil
	}
	return &BlockReward{
		chain:     chain,
		rewardCfg: rewardCfg,
	}
}
func (br *BlockReward) calcValidatorRewards(blockReward *big.Int, rewards map[common.Address]*big.Int, Leader common.Address, header *types.Header) {
	leaderBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.LeaderRate)
	br.rewardCfg.SetReward.SetLeaderRewards(leaderBlkReward, rewards, Leader, header.Number)
	electedReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.ElectedValidatorsRate)
	br.rewardCfg.SetReward.SetSelectedRewards(electedReward, br.chain, rewards, common.RoleValidator|common.RoleBackupValidator, header, br.rewardCfg.RewardMount.BackupRewardRate)
	return
}

func (br *BlockReward) calcMinerRewards(blockReward *big.Int, rewards map[common.Address]*big.Int, header *types.Header) {
	//广播区块不给矿工发钱
	minerOutReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.MinerOutRate)
	br.rewardCfg.SetReward.SetMinerOutRewards(minerOutReward, br.chain, header.Number, rewards)
	electedReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.ElectedMineRate)
	br.rewardCfg.SetReward.SetSelectedRewards(electedReward, br.chain, rewards, common.RoleMiner|common.RoleBackupMiner, header, br.rewardCfg.RewardMount.BackupRewardRate)
	return
}

func (br *BlockReward) calcFoundationRewards(blockReward *big.Int, rewards map[common.Address]*big.Int, num *big.Int) {
	if common.IsBroadcastNumber(num.Uint64()) {
		return
	}
	foundationNum := int64(len(manparams.FoundationNodes))
	if foundationNum == 0 {
		return
	}

	oneFoundationReward := new(big.Int).Div(blockReward, big.NewInt(foundationNum))

	for _, v := range manparams.FoundationNodes {
		util.SetAccountRewards(rewards, v.Address, oneFoundationReward)
		log.Info(PackageName, "FoundationReward account", v.Address, "reward", oneFoundationReward.Uint64())
	}
}

//func (br *BlockReward) CalcRewards(state *state.StateDB, Leader common.Address, num *big.Int, txs types.Transactions, header *types.Header) map[common.Address]*big.Int {
//	rewards := make(map[common.Address]*big.Int, 0)
//
//	br.CalcBlockRewards(rewards, Leader, num, header)
//	br.CalcTransactionsRewards(rewards, Leader, num, txs, header)
//	for account, reward := range rewards {
//		depoistInfo.AddReward(state, account, reward)
//	}
//	return rewards
//}

func (br *BlockReward) CalcBlockRewards(blockReward *big.Int, Leader common.Address, header *types.Header) map[common.Address]*big.Int {

	rewards := make(map[common.Address]*big.Int, 0)
	if nil==br.rewardCfg{
		log.Error(PackageName,"奖励配置为空","")
		return nil
	}
	validatorsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.ValidatorsRate)
	br.calcValidatorRewards(validatorsBlkReward, rewards, Leader, header)
	minersBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.MinersRate)
	br.calcMinerRewards(minersBlkReward, rewards, header)
	FoundationsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.MinersRate)
	br.calcFoundationRewards(FoundationsBlkReward, rewards, header.Number)

	return rewards
}
