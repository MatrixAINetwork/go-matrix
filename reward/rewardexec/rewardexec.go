// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package rewardexec

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/reward/cfg"
	"github.com/MatrixAINetwork/go-matrix/reward/util"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

const (
	PackageName = "奖励"
)

type BlockReward struct {
	chain              util.ChainReader
	st                 util.StateDB
	rewardCfg          *cfg.RewardCfg
	foundationAccount  common.Address
	innerMinerAccounts []common.Address
	bcInterval         *mc.BCIntervalInfo
	topology           *mc.TopologyGraph
	elect              *mc.ElectGraph
}

func New(chain util.ChainReader, rewardCfg *cfg.RewardCfg, st util.StateDB, interval *mc.BCIntervalInfo, foundationAccount common.Address, top *mc.TopologyGraph, elect *mc.ElectGraph) *BlockReward {
	if util.RewardFullRate != rewardCfg.RewardMount.RewardRate.MinerOutRate+rewardCfg.RewardMount.RewardRate.ElectedMinerRate+rewardCfg.RewardMount.RewardRate.FoundationMinerRate {
		log.ERROR(PackageName, "矿工固定区块奖励比例配置错误", "")
		return nil
	}
	if util.RewardFullRate != rewardCfg.RewardMount.RewardRate.LeaderRate+rewardCfg.RewardMount.RewardRate.ElectedValidatorsRate+rewardCfg.RewardMount.RewardRate.FoundationValidatorRate {
		log.ERROR(PackageName, "验证者固定区块奖励比例配置错误", "")
		return nil
	}

	if util.RewardFullRate != rewardCfg.RewardMount.RewardRate.OriginElectOfflineRate+rewardCfg.RewardMount.RewardRate.BackupRewardRate {
		log.ERROR(PackageName, "替补固定区块奖励比例配置错误", "")
		return nil
	}

	br := &BlockReward{
		chain:             chain,
		rewardCfg:         rewardCfg,
		st:                st,
		foundationAccount: foundationAccount,
		elect:             elect,
		topology:          top,
	}
	br.bcInterval = interval
	return br
}
func (br *BlockReward) CalcValidatorRateMount(blockReward *big.Int) (*big.Int, *big.Int, *big.Int) {

	leaderBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.LeaderRate)
	electedReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.ElectedValidatorsRate)
	FoundationsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.FoundationValidatorRate)
	return leaderBlkReward, electedReward, FoundationsBlkReward
}

func (br *BlockReward) CalcMinerRateMount(blockReward *big.Int) (*big.Int, *big.Int, *big.Int) {

	minerOutReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.MinerOutRate)
	electedReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.ElectedMinerRate)
	FoundationsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.FoundationMinerRate)
	return minerOutReward, electedReward, FoundationsBlkReward
}

func (br *BlockReward) CalcValidatorRewards(Leader common.Address, num uint64) map[common.Address]*big.Int {
	//广播区块不给矿工发钱
	RewardMan := new(big.Int).Mul(new(big.Int).SetUint64(br.rewardCfg.RewardMount.ValidatorMount), util.GetPrice(br.rewardCfg.Calc))
	halfNum := br.rewardCfg.RewardMount.ValidatorAttenuationNum
	attenuationRate := br.rewardCfg.RewardMount.ValidatorAttenuationRate
	blockReward := util.CalcRewardMountByNumber(br.st, RewardMan, num-1, halfNum, common.BlkValidatorRewardAddress, attenuationRate)
	if blockReward.Uint64() == 0 {
		log.Error(PackageName, "账户余额为0，不发放验证者奖励", "")
		return nil
	}

	if nil == br.rewardCfg {
		log.Error(PackageName, "奖励配置为空", "")
		return nil
	}

	if br.bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播周期不处理", "")
		return nil
	}

	return br.getValidatorRewards(blockReward, Leader, num)
}

func (br *BlockReward) getValidatorRewards(blockReward *big.Int, Leader common.Address, num uint64) map[common.Address]*big.Int {
	//广播区块不给矿工发钱
	rewards := make(map[common.Address]*big.Int, 0)
	leaderBlkMount, electedMount, FoundationsMount := br.CalcValidatorRateMount(blockReward)
	leaderReward := br.rewardCfg.SetReward.SetLeaderRewards(leaderBlkMount, Leader, num)
	electReward := br.rewardCfg.SetReward.GetSelectedRewards(electedMount, br.st, common.RoleValidator|common.RoleBackupValidator, num, br.rewardCfg.RewardMount.RewardRate.BackupRewardRate, br.topology, br.elect)
	foundationReward := br.calcFoundationRewards(FoundationsMount, num)
	util.MergeReward(rewards, leaderReward)
	util.MergeReward(rewards, electReward)
	util.MergeReward(rewards, foundationReward)
	return rewards
}

func (br *BlockReward) getMinerRewards(blockReward *big.Int, num uint64, rewardType uint8, parentHash common.Hash, coinType string) map[common.Address]*big.Int {
	rewards := make(map[common.Address]*big.Int, 0)

	minerOutAmount, electedMount, FoundationsMount := br.CalcMinerRateMount(blockReward)
	minerOutReward := br.rewardCfg.SetReward.SetMinerOutRewards(minerOutAmount, br.st, br.chain, num, parentHash, coinType)
	electReward := br.rewardCfg.SetReward.GetSelectedRewards(electedMount, br.st, common.RoleMiner|common.RoleBackupMiner, num, br.rewardCfg.RewardMount.RewardRate.BackupRewardRate, br.topology, br.elect)
	foundationReward := br.calcFoundationRewards(FoundationsMount, num)
	util.MergeReward(rewards, minerOutReward)
	util.MergeReward(rewards, electReward)
	util.MergeReward(rewards, foundationReward)
	return rewards
}

func (br *BlockReward) CalcMinerRewards(num uint64, parentHash common.Hash) map[common.Address]*big.Int {
	//广播区块不给矿工发钱
	RewardMan := new(big.Int).Mul(new(big.Int).SetUint64(br.rewardCfg.RewardMount.MinerMount), util.GetPrice(br.rewardCfg.Calc))
	halfNum := br.rewardCfg.RewardMount.MinerAttenuationNum
	attenuationRate := br.rewardCfg.RewardMount.MinerAttenuationRate
	blockReward := util.CalcRewardMountByNumber(br.st, RewardMan, num-1, halfNum, common.BlkMinerRewardAddress, attenuationRate)
	if blockReward.Uint64() == 0 {
		log.Error(PackageName, "账户余额为0，不发放矿工奖励", "")
		return nil
	}
	if nil == br.rewardCfg {
		log.Error(PackageName, "奖励配置为空", "")
		return nil
	}

	if br.bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播周期不处理", "")
		return nil
	}
	return br.getMinerRewards(blockReward, num, util.BlkReward, parentHash, params.MAN_COIN)
}
func (br *BlockReward) canCalcFoundationRewards(blockReward *big.Int, num uint64) bool {
	if br.bcInterval.IsBroadcastNumber(num) {
		return false
	}

	if blockReward.Cmp(big.NewInt(0)) <= 0 {
		//log.ERROR(PackageName, "奖励金额错误", blockReward)
		return false
	}
	return true

}
func (br *BlockReward) calcFoundationRewards(blockReward *big.Int, num uint64) map[common.Address]*big.Int {

	if false == br.canCalcFoundationRewards(blockReward, num) {
		return nil
	}
	accountRewards := make(map[common.Address]*big.Int)
	accountRewards[br.foundationAccount] = blockReward
	//log.Debug(PackageName, "基金会奖励,账户", br.foundationAccount.Hex(), "金额", blockReward)
	return accountRewards
}

func (br *BlockReward) CalcNodesRewards(blockReward *big.Int, Leader common.Address, num uint64, parentHash common.Hash, coinType string) map[common.Address]*big.Int {

	if nil == br.rewardCfg {
		log.Error(PackageName, "奖励配置为空", "")
		return nil
	}

	if br.bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播周期不处理", "")
		return nil
	}

	rewards := make(map[common.Address]*big.Int, 0)
	//log.Debug(PackageName, "奖励金额", blockReward)
	minersBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.MinersRate)
	minerRewards := br.getMinerRewards(minersBlkReward, num, util.TxsReward, parentHash, coinType)
	if blockReward.Cmp(big.NewInt(0)) <= 0 {
		//	log.Warn(PackageName, "账户余额非法，不发放奖励", blockReward)
		return minerRewards
	}

	validatorsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.ValidatorsRate)
	validatorReward := br.getValidatorRewards(validatorsBlkReward, Leader, num)

	util.MergeReward(rewards, validatorReward)
	util.MergeReward(rewards, minerRewards)
	return rewards
}

func (br *BlockReward) GetRewardCfg() *cfg.RewardCfg {

	return br.rewardCfg
}
