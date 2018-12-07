package rewardexec

import (
	"github.com/matrix/go-matrix/core/state"
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

	if util.RewardFullRate != rewardCfg.RewardMount.MinersRate+rewardCfg.RewardMount.ValidatorsRate {
		log.ERROR(PackageName, "固定区块奖励比例配置错误", "")
		return nil
	}
	if util.RewardFullRate != rewardCfg.RewardMount.MinerOutRate+rewardCfg.RewardMount.ElectedMinerRate+rewardCfg.RewardMount.FoundationMinerRate {
		log.ERROR(PackageName, "矿工固定区块奖励比例配置错误", "")
		return nil
	}
	if util.RewardFullRate != rewardCfg.RewardMount.LeaderRate+rewardCfg.RewardMount.ElectedValidatorsRate+rewardCfg.RewardMount.FoundationValidatorRate {
		log.ERROR(PackageName, "验证者固定区块奖励比例配置错误", "")
		return nil
	}

	if util.RewardFullRate != rewardCfg.RewardMount.OriginElectOfflineRate+rewardCfg.RewardMount.BackupRewardRate {
		log.ERROR(PackageName, "替补固定区块奖励比例配置错误", "")
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
	FoundationsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.FoundationValidatorRate)
	br.calcFoundationRewards(FoundationsBlkReward, rewards, header.Number)

	return
}

func (br *BlockReward) calcMinerRewards(blockReward *big.Int, rewards map[common.Address]*big.Int, header *types.Header) {
	//广播区块不给矿工发钱
	minerOutReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.MinerOutRate)
	br.rewardCfg.SetReward.SetMinerOutRewards(minerOutReward, br.chain, header.Number, rewards)
	electedReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.ElectedMinerRate)
	br.rewardCfg.SetReward.SetSelectedRewards(electedReward, br.chain, rewards, common.RoleMiner|common.RoleBackupMiner, header, br.rewardCfg.RewardMount.BackupRewardRate)
	FoundationsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.FoundationMinerRate)
	br.calcFoundationRewards(FoundationsBlkReward, rewards, header.Number)
	return
}

func (br *BlockReward) CalcValidatorRewards(blockReward *big.Int, Leader common.Address, header *types.Header) map[common.Address]*big.Int {
	//广播区块不给矿工发钱

	if blockReward.Uint64() == 0 {
		log.Error(PackageName, "账户余额为0，不发放验证者奖励", "")
		return nil
	}
	rewards := make(map[common.Address]*big.Int, 0)
	if nil == br.rewardCfg {
		log.Error(PackageName, "奖励配置为空", "")
		return nil
	}
	br.calcValidatorRewards(blockReward, rewards, Leader, header)
	return rewards
}

func (br *BlockReward) CalcMinerRewards(blockReward *big.Int, header *types.Header) map[common.Address]*big.Int {
	//广播区块不给矿工发钱
	if blockReward.Uint64() == 0 {
		log.Error(PackageName, "账户余额为0，不发放矿工奖励", "")
		return nil
	}
	rewards := make(map[common.Address]*big.Int, 0)
	if nil == br.rewardCfg {
		log.Error(PackageName, "奖励配置为空", "")
		return nil
	}
	br.calcMinerRewards(blockReward, rewards, header)
	return rewards
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
		log.Info(PackageName, "基金会 账户", v.Address, "奖励", oneFoundationReward.Uint64())
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

func (br *BlockReward) CalcNodesRewards(blockReward *big.Int, Leader common.Address, header *types.Header) map[common.Address]*big.Int {

	if blockReward.Uint64() == 0 {
		log.Error(PackageName, "账户余额为0，不发放奖励", "")
		return nil
	}
	rewards := make(map[common.Address]*big.Int, 0)
	if nil == br.rewardCfg {
		log.Error(PackageName, "奖励配置为空", "")
		return nil
	}
	validatorsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.ValidatorsRate)
	br.calcValidatorRewards(validatorsBlkReward, rewards, Leader, header)
	minersBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.MinersRate)
	br.calcMinerRewards(minersBlkReward, rewards, header)
	return rewards
}

func (br *BlockReward) CalcRewardMount(state *state.StateDB, blockReward *big.Int, address common.Address) *big.Int {
	//todo:后续从状态树读取对应币种减半金额,现在每个100个区块余额减半，如果减半值为0则不减半
	halfBalance := new(big.Int).Exp(big.NewInt(10), big.NewInt(21), big.NewInt(0))
	balance := state.GetBalance(address)
	genesisState, _ := br.chain.StateAt(br.chain.Genesis().Root())
	genesisBalance := genesisState.GetBalance(address)
	log.INFO(PackageName, "计算区块奖励参数 衰减金额:", halfBalance.String(),
		"初始账户", address.String(), "初始金额", genesisBalance[common.MainAccount].Balance.String(), "当前金额", balance[common.MainAccount].Balance.String())
	var reward *big.Int
	if balance[common.MainAccount].Balance.Cmp(genesisBalance[common.MainAccount].Balance) >= 0 {
		reward = blockReward
	}

	subBalance := new(big.Int).Sub(genesisBalance[common.MainAccount].Balance, balance[common.MainAccount].Balance)
	n := int64(0)
	if 0 != halfBalance.Int64() {
		n = new(big.Int).Div(subBalance, halfBalance).Int64()
	}

	if 0 == n {
		reward = blockReward
	} else {
		reward = new(big.Int).Div(blockReward, new(big.Int).Exp(big.NewInt(2), big.NewInt(n), big.NewInt(0)))
	}
	log.INFO(PackageName, "计算区块奖励金额:", reward.String())
	if balance[common.MainAccount].Balance.Cmp(reward) < 0 {
		log.ERROR(PackageName, "账户余额不足，余额为", balance[common.MainAccount].Balance.String())
		return big.NewInt(0)
	} else {
		return reward
	}

}
