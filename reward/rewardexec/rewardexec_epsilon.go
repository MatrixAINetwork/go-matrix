// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package rewardexec

import (
	"math/big"
	"sort"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/reward/cfg"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

type AIBlockReward struct {
	chain              util.ChainReader
	st                 util.StateDB
	rewardCfg          *cfg.AIRewardCfg
	foundationAccount  common.Address
	innerMinerAccounts []common.Address
	bcInterval         *mc.BCIntervalInfo
	topology           *mc.TopologyGraph
	elect              *mc.ElectGraph
}

func AIBlockNew(chain util.ChainReader, rewardCfg *cfg.AIRewardCfg, st util.StateDB, interval *mc.BCIntervalInfo, foundationAccount common.Address, top *mc.TopologyGraph, elect *mc.ElectGraph) *AIBlockReward {
	if util.RewardFullRate != rewardCfg.RewardMount.RewardRate.MinerOutRate+rewardCfg.RewardMount.RewardRate.ElectedMinerRate+rewardCfg.RewardMount.RewardRate.FoundationMinerRate+rewardCfg.RewardMount.RewardRate.AIMinerOutRate {
		log.Error(PackageName, "矿工固定区块奖励比例配置错误", "")
		return nil
	}
	if util.RewardFullRate != rewardCfg.RewardMount.RewardRate.LeaderRate+rewardCfg.RewardMount.RewardRate.ElectedValidatorsRate+rewardCfg.RewardMount.RewardRate.FoundationValidatorRate {
		log.Error(PackageName, "验证者固定区块奖励比例配置错误", "")
		return nil
	}

	if util.RewardFullRate != rewardCfg.RewardMount.RewardRate.OriginElectOfflineRate+rewardCfg.RewardMount.RewardRate.BackupRewardRate {
		log.Error(PackageName, "替补固定区块奖励比例配置错误", "")
		return nil
	}

	br := &AIBlockReward{
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
func (br *AIBlockReward) CalcValidatorRateMount(blockReward *big.Int) (*big.Int, *big.Int, *big.Int) {

	leaderBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.LeaderRate)
	electedReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.ElectedValidatorsRate)
	FoundationsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.FoundationValidatorRate)
	return leaderBlkReward, electedReward, FoundationsBlkReward
}

func (br *AIBlockReward) CalcMinerRateMount(blockReward *big.Int) (*big.Int, *big.Int, *big.Int, *big.Int) {

	minerOutReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.MinerOutRate)
	AIOutReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.AIMinerOutRate)
	electedReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.ElectedMinerRate)
	FoundationsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.FoundationMinerRate)
	return minerOutReward, electedReward, FoundationsBlkReward, AIOutReward
}

func (br *AIBlockReward) CalcValidatorRewards(Leader common.Address, num uint64, shouldPaySelectReward bool) map[common.Address]*big.Int {
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
		log.Warn(PackageName, "广播周期不处理", "")
		return nil
	}

	return br.getValidatorRewards(blockReward, Leader, num, params.MAN_COIN, shouldPaySelectReward)
}

func (br *AIBlockReward) getValidatorRewards(blockReward *big.Int, Leader common.Address, num uint64, coinType string, shouldPaySelectReward bool) map[common.Address]*big.Int {
	//广播区块不给矿工发钱
	rewards := make(map[common.Address]*big.Int, 0)
	leaderBlkMount, electedMount, FoundationsMount := br.CalcValidatorRateMount(blockReward)
	leaderReward := br.rewardCfg.SetReward.SetLeaderRewards(leaderBlkMount, Leader, num)
	electReward := br.selValidatorReward(electedMount, num, coinType, shouldPaySelectReward)
	foundationReward := br.calcFoundationRewards(FoundationsMount, num)
	util.MergeReward(rewards, leaderReward)
	util.MergeReward(rewards, electReward)
	util.MergeReward(rewards, foundationReward)
	return rewards
}
func (br *AIBlockReward) canPayBLKSelectValidatorReward(num uint64) bool {
	latestNum, err := matrixstate.GetBLKSelValidatorNum(br.st)
	if nil != err {
		log.Error(PackageName, "状态树获取前一发放参与验证者奖励高度错误", err)
		return false
	}
	if latestNum > br.bcInterval.GetLastReElectionNumber() {
		return false
	}

	return true
}

func (br *AIBlockReward) canPayTxsSelectValidatorReward(num uint64) bool {
	latestNum, err := matrixstate.GetTXSSelValidatorNum(br.st)
	if nil != err {
		log.Error(PackageName, "状态树获取前一发放参与验证者奖励高度错误", err)
		return false
	}
	if latestNum > br.bcInterval.GetLastReElectionNumber() {
		return false
	}
	/*if err := matrixstate.SetTXSSelValidatorNum(br.st, num); err != nil {
		log.Error(PackageName, "设置参与验证者奖励状态错误", err)
	}*/
	return true
}
func (br *AIBlockReward) CanPaySelectValidatorReward(num uint64) bool {
	if br.rewardCfg.RewardType == util.BlkReward {
		return br.canPayBLKSelectValidatorReward(num)
	} else {
		return br.canPayTxsSelectValidatorReward(num)
	}
}

func (br *AIBlockReward) SetPaySelectValidatorReward(num uint64) error {
	if br.rewardCfg.RewardType == util.BlkReward {

		return matrixstate.SetBLKSelValidatorNum(br.st, num)
	} else {
		return matrixstate.SetTXSSelValidatorNum(br.st, num)
	}
}

func (br *AIBlockReward) selPayValidatorReward(num uint64, coinType string, shouldPaySelectReward bool) map[common.Address]*big.Int {
	electReward := make(map[common.Address]*big.Int, 0)
	if !shouldPaySelectReward {
		return electReward
	}
	selReward, err := br.getSelRewardList(coinType)
	if nil != err {
		log.Error(PackageName, "获取验证者参与奖励错误", err)
		return electReward
	}
	if len(selReward.RewardList) == 0 {
		log.Error(PackageName, "无参与奖励", "")
		return electReward
	}
	for _, v := range selReward.RewardList {
		electReward[v.Address] = v.Amount
	}
	//清除累加的参与奖励
	br.setSelRewardList(mc.ValidatorSelReward{CoinType: coinType})
	return electReward
}

func findMultiCoinSelReward(coinType string, multiCoinSelReward []mc.ValidatorSelReward) mc.ValidatorSelReward {
	for _, v := range multiCoinSelReward {
		if v.CoinType == coinType {
			return v
		}
	}
	return mc.ValidatorSelReward{}
}

func setMultiCoinSelReward(SelReward mc.ValidatorSelReward, multiCoinSelRewardList []mc.ValidatorSelReward) []mc.ValidatorSelReward {
	if multiCoinSelRewardList == nil {
		multiCoinSelRewardList = make([]mc.ValidatorSelReward, 0)
	}

	for i, v := range multiCoinSelRewardList {
		if v.CoinType == SelReward.CoinType {
			multiCoinSelRewardList[i].RewardList = SelReward.RewardList
			return multiCoinSelRewardList
		}
	}

	multiCoinSelRewardList = append(multiCoinSelRewardList, SelReward)

	return multiCoinSelRewardList
}

func (br *AIBlockReward) getSelRewardList(coinType string) (mc.ValidatorSelReward, error) {
	if br.rewardCfg.RewardType == util.BlkReward {
		return matrixstate.GetBLKSelValidator(br.st)
	} else {
		selReward, err := matrixstate.GetTXSSelValidator(br.st)
		if nil != err {

			return mc.ValidatorSelReward{}, err
		}
		return findMultiCoinSelReward(coinType, selReward), nil
	}
}

func (br *AIBlockReward) setSelRewardList(data mc.ValidatorSelReward) error {
	if br.rewardCfg.RewardType == util.BlkReward {
		return matrixstate.SetBLKSelValidator(br.st, data)
	} else {
		selReward, err := matrixstate.GetTXSSelValidator(br.st)
		if nil != err {
			return err
		}
		return matrixstate.SetTXSSelValidator(br.st, setMultiCoinSelReward(data, selReward))
	}
}
func findValidatorSelAddress(addr common.Address, addrList []mc.SelReward) (int, bool) {
	for k, v := range addrList {
		if v.Address.Equal(addr) == true {
			return k, true
		}
	}
	return 0, false
}

func (br *AIBlockReward) selCalcValidatorReward(electedMount *big.Int, num uint64, coinType string) {
	if electedMount.Uint64() == 0 {
		return
	}
	selRewardList, err := br.getSelRewardList(coinType)
	if nil != err {
		log.Error(PackageName, "获取验证者参与奖励错误", err)

	}

	electReward := br.rewardCfg.SetReward.GetSelectedRewards(electedMount, br.st, common.RoleValidator|common.RoleBackupValidator, num, br.rewardCfg.RewardMount.RewardRate.BackupRewardRate, br.topology, br.elect)
	sorted_keys := make([]string, 0)

	for k, _ := range electReward {
		sorted_keys = append(sorted_keys, k.String())
	}
	sort.Strings(sorted_keys)

	for _, accountstring := range sorted_keys {
		account := common.HexToAddress(accountstring)
		if index, ok := findValidatorSelAddress(account, selRewardList.RewardList); ok {
			selRewardList.RewardList[index].Amount.Add(selRewardList.RewardList[index].Amount, electReward[account])
		} else {
			selRewardList.RewardList = append(selRewardList.RewardList, mc.SelReward{Address: account, Amount: electReward[account]})
		}

	}
	selRewardList.CoinType = coinType
	br.setSelRewardList(selRewardList)
}

func (br *AIBlockReward) selValidatorReward(electedMount *big.Int, num uint64, coinType string, shouldPaySelectReward bool) map[common.Address]*big.Int {
	electPayReward := br.selPayValidatorReward(num, coinType, shouldPaySelectReward)
	br.selCalcValidatorReward(electedMount, num, coinType)
	return electPayReward
}

func (br *AIBlockReward) getMinerRewards(blockReward *big.Int, num uint64, rewardType uint8, parentHash common.Hash, coinType string) map[common.Address]*big.Int {
	rewards := make(map[common.Address]*big.Int, 0)

	minerOutAmount, electedMount, FoundationsMount, AIMinerAMount := br.CalcMinerRateMount(blockReward)
	minerOutReward := br.rewardCfg.SetReward.SetMinerOutRewards(AIMinerAMount, minerOutAmount, br.st, br.chain, num, parentHash, coinType)
	electReward := br.rewardCfg.SetReward.GetSelectedRewards(electedMount, br.st, common.RoleMiner|common.RoleBackupMiner, num, br.rewardCfg.RewardMount.RewardRate.BackupRewardRate, br.topology, br.elect)
	foundationReward := br.calcFoundationRewards(FoundationsMount, num)
	util.MergeReward(rewards, minerOutReward)
	util.MergeReward(rewards, electReward)
	util.MergeReward(rewards, foundationReward)
	return rewards
}

func (br *AIBlockReward) CalcMinerRewards(num uint64, parentHash common.Hash) map[common.Address]*big.Int {

	return br.calcEpsilonMinerRewards(num, parentHash)

}

func (br *AIBlockReward) canCalcFoundationRewards(blockReward *big.Int, num uint64) bool {
	if br.bcInterval.IsBroadcastNumber(num) {
		return false
	}

	if blockReward.Cmp(big.NewInt(0)) <= 0 {
		//log.Error(PackageName, "奖励金额错误", blockReward)
		return false
	}
	return true

}
func (br *AIBlockReward) calcFoundationRewards(blockReward *big.Int, num uint64) map[common.Address]*big.Int {

	if false == br.canCalcFoundationRewards(blockReward, num) {
		return nil
	}
	accountRewards := make(map[common.Address]*big.Int)
	accountRewards[br.foundationAccount] = blockReward
	//log.Debug(PackageName, "基金会奖励,账户", br.foundationAccount.Hex(), "金额", blockReward)
	return accountRewards
}

func (br *AIBlockReward) CalcNodesRewards(blockReward *big.Int, Leader common.Address, num uint64, parentHash common.Hash, coinType string, shouldPaySelectReward bool) map[common.Address]*big.Int {

	if nil == br.rewardCfg {
		log.Error(PackageName, "奖励配置为空", "")
		return nil
	}

	if br.bcInterval.IsBroadcastNumber(num) {
		log.Warn(PackageName, "广播周期不处理", "")
		return nil
	}

	rewards := make(map[common.Address]*big.Int, 0)
	//log.Debug(PackageName, "奖励金额", blockReward)
	minersBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.MinersRate)
	minerRewards := br.getMinerRewards(minersBlkReward, num, util.TxsReward, parentHash, coinType)

	validatorsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.ValidatorsRate)
	validatorReward := br.getValidatorRewards(validatorsBlkReward, Leader, num, coinType, shouldPaySelectReward)

	util.MergeReward(rewards, validatorReward)
	util.MergeReward(rewards, minerRewards)
	return rewards
}

func (br *AIBlockReward) GetRewardCfg() *cfg.AIRewardCfg {

	return br.rewardCfg
}

func (br *AIBlockReward) calcEpsilonMinerRewards(num uint64, parentHash common.Hash) map[common.Address]*big.Int {
	//广播区块不给矿工发钱
	if nil == br.rewardCfg {
		log.Error(PackageName, "奖励配置为空", "")
		return nil
	}
	if br.bcInterval.IsBroadcastNumber(num) {
		log.Warn(PackageName, "广播周期不处理", "")
		return nil
	}

	rewards := br.getEpsilonMinerOutRewards(num, parentHash)
	if !br.canPaySelectMinerReward(num) {
		return rewards
	}
	br.getEpsilonMinerSelect(num, rewards, parentHash)
	return rewards
}

func (br *AIBlockReward) canPaySelectMinerReward(num uint64) bool {
	latestNum, err := matrixstate.GetSelMinerNum(br.st)
	if nil != err {
		log.Error(PackageName, "状态树获取前一发放参与矿工奖励高度错误", err)
		return false
	}
	if latestNum > br.bcInterval.GetLastReElectionNumber() {
		return false
	}
	if err := matrixstate.SetSelMinerNum(br.st, num); err != nil {
		log.Error(PackageName, "设置参与矿工奖励状态错误", err)
	}
	return true
}

func (br *AIBlockReward) getEpsilonMinerSelect(num uint64, rewards map[common.Address]*big.Int, parentHash common.Hash) {
	originBlockRewardMount, finalBlockRewardMount := br.getEpsilonSelectAttenuationMount(num)
	preAttenuationNum, afterAttenuationNum := br.getEpsilonSelectAttenuationNum(originBlockRewardMount, finalBlockRewardMount)
	preElectReward := br.getEpsilonSelectReward(originBlockRewardMount, preAttenuationNum, parentHash)
	util.MergeReward(rewards, preElectReward)
	afterElectReward := br.getEpsilonSelectReward(finalBlockRewardMount, afterAttenuationNum, parentHash)
	util.MergeReward(rewards, afterElectReward)
}

func (br *AIBlockReward) getEpsilonSelectAttenuationMount(num uint64) (*big.Int, *big.Int) {
	RewardMan := new(big.Int).Mul(new(big.Int).SetUint64(br.rewardCfg.RewardMount.MinerMount), util.GetPrice(br.rewardCfg.Calc))
	attenuationRate := br.rewardCfg.RewardMount.MinerAttenuationRate
	//blockReward := util.CalcRewardMountByNumber(br.st, RewardMan, num-br.bcInterval.GetReElectionInterval(), halfNum, common.BlkMinerRewardAddress, attenuationRate)
	n := util.CalcN(br.rewardCfg.RewardMount.MinerAttenuationNum, num-br.bcInterval.GetReElectionInterval())
	originBlockRewardMount := util.CalcRewardMount(RewardMan, n, attenuationRate)
	n = util.CalcN(br.rewardCfg.RewardMount.MinerAttenuationNum, num-2)
	finalBlockRewardMount := util.CalcRewardMount(RewardMan, n, attenuationRate)
	return originBlockRewardMount, finalBlockRewardMount
}

func (br *AIBlockReward) getEpsilonSelectAttenuationNum(originBlockRewardMount *big.Int, finalBlockRewardMount *big.Int) (uint64, uint64) {
	var preAttenuationNum uint64
	var afterAttenuationNum uint64
	//衰减高度区间,300周期整数倍切换
	if originBlockRewardMount.Cmp(finalBlockRewardMount) != 0 {
		prebroadintervalTimes := br.rewardCfg.RewardMount.MinerAttenuationNum % br.bcInterval.GetReElectionInterval() / br.bcInterval.GetBroadcastInterval()
		preAttenuationNum = (br.rewardCfg.RewardMount.MinerAttenuationNum)%br.bcInterval.GetReElectionInterval() - prebroadintervalTimes - 1
		afterAttenuationNum = br.bcInterval.GetReElectionInterval() - mc.ReelectionTimes - preAttenuationNum
	} else {
		afterAttenuationNum = br.bcInterval.GetReElectionInterval() - mc.ReelectionTimes
		preAttenuationNum = 0
	}
	return preAttenuationNum, afterAttenuationNum
}

func (br *AIBlockReward) getEpsilonSelectReward(BlockRewardMount *big.Int, AttenuationNum uint64, parentHash common.Hash) map[common.Address]*big.Int {
	if AttenuationNum == 0 {
		return nil
	}
	electedMount := br.getElectMount(BlockRewardMount, AttenuationNum)
	//electReward := br.rewardCfg.SetReward.GetSelectedRewards(electedMount, br.st, common.RoleMiner, num, br.rewardCfg.RewardMount.RewardRate.BackupRewardRate, br.topology, br.elect)
	electReward := br.calcEpslionMinerSelReward(electedMount, parentHash)
	return electReward
}

func (br *AIBlockReward) getElectMount(BlockRewardMount *big.Int, AttenuationNum uint64) *big.Int {
	preAttenuationReWard := new(big.Int).Mul(BlockRewardMount, new(big.Int).SetUint64(AttenuationNum))
	_, electedMount, _, _ := br.CalcMinerRateMount(preAttenuationReWard)
	return electedMount
}

func (br *AIBlockReward) getEpsilonMinerOutRewards(num uint64, parentHash common.Hash) map[common.Address]*big.Int {
	RewardMan := new(big.Int).Mul(new(big.Int).SetUint64(br.rewardCfg.RewardMount.MinerMount), util.GetPrice(br.rewardCfg.Calc))
	halfNum := br.rewardCfg.RewardMount.MinerAttenuationNum
	attenuationRate := br.rewardCfg.RewardMount.MinerAttenuationRate
	blockReward := util.CalcRewardMountByNumber(br.st, RewardMan, num-1, halfNum, common.BlkMinerRewardAddress, attenuationRate)
	if blockReward.Uint64() == 0 {
		log.Error(PackageName, "账户余额为0，不发放矿工奖励", "")
		return nil
	}
	rewards := make(map[common.Address]*big.Int, 0)
	minerOutAmount, _, FoundationsMount, AIminerOUnt := br.CalcMinerRateMount(blockReward)
	minerOutReward := br.rewardCfg.SetReward.SetMinerOutRewards(AIminerOUnt, minerOutAmount, br.st, br.chain, num, parentHash, params.MAN_COIN)
	util.MergeReward(rewards, minerOutReward)
	foundationReward := br.calcFoundationRewards(FoundationsMount, num)
	util.MergeReward(rewards, foundationReward)
	return rewards
}

func (br *AIBlockReward) getMinerElect(parentHash common.Hash) []common.Address {
	//计算选举的拓扑图的高度
	eleNum := br.bcInterval.GetLastReElectionNumber() - 2
	ancetorHash, err := br.chain.GetAncestorHash(parentHash, eleNum)
	if err != nil {
		log.Error(PackageName, "获取指定高度hash失败", err, "eleNum", eleNum)
		return nil
	}
	_, electGraph, err := br.chain.GetGraphByHash(ancetorHash)
	if err != nil {
		log.Error(PackageName, "获取选举失败", err, "ancetorHash", ancetorHash.TerminalString())
		return nil
	}
	electMiner := make([]common.Address, 0)
	for _, v := range electGraph.ElectList {
		if v.Type != common.RoleMiner {
			continue
		}
		electMiner = append(electMiner, v.Account)
	}
	return electMiner

}
func findAddress(addr common.Address, addrList []mc.BasePowerSlash) bool {
	for _, v := range addrList {
		if v.Address.Equal(addr) == true {
			return true
		}
	}
	return false
}
func (br *AIBlockReward) calcEpslionMinerSelReward(reward *big.Int, parentHash common.Hash) map[common.Address]*big.Int {
	elect := br.getMinerElect(parentHash)
	if elect == nil {
		log.Error(PackageName, "获取选举结果错误", "")
	}
	if len(elect) == 0 {
		log.Error(PackageName, "获取选举个数为", 0)
		return nil
	}
	bpBlackList, err := matrixstate.GetBasePowerBlackList(br.st)
	if nil != err {
		log.Error(PackageName, "获取算力黑名单错误", err)
	}
	nodelist := br.filterBlackList(elect, bpBlackList)
	// elect 按照选举人个数均分，黑名单不参与分成
	oneNodeReward := br.getOneNodeReward(reward, uint64(len(elect)))
	//log.Info(PackageName, "计算抵押总额,账户股权", totalStock)
	return br.setRewards(nodelist, oneNodeReward)

}

func (br *AIBlockReward) calcEpslionMinerSelRewardB(reward *big.Int, parentHash common.Hash) map[common.Address]*big.Int {
	elect := br.getMinerElect(parentHash)
	if elect == nil {
		log.Crit(PackageName, "获取选举结果错误", "")
	}
	bpBlackList, err := matrixstate.GetBasePowerBlackList(br.st)
	if nil != err {
		log.Crit(PackageName, "获取算力黑名单错误", err)
	}
	nodelist := br.filterBlackList(elect, bpBlackList)
	// elect 按照选举人个数均分，黑名单不参与分成
	oneNodeReward := br.getOneNodeReward(reward, uint64(len(nodelist)))
	//log.Info(PackageName, "计算抵押总额,账户股权", totalStock)
	return br.setRewards(nodelist, oneNodeReward)

}

func (br *AIBlockReward) getOneNodeReward(reward *big.Int, count uint64) *big.Int {
	oneNodeReward := new(big.Int).Div(reward, new(big.Int).SetUint64(count))
	return oneNodeReward
}

func (br *AIBlockReward) filterBlackList(elect []common.Address, bpBlackList *mc.BasePowerSlashBlackList) []common.Address {
	nodelist := make([]common.Address, 0)
	for _, v := range elect {
		if findAddress(v, bpBlackList.BlackList) {
			continue
		}
		nodelist = append(nodelist, v)
	}
	return nodelist
}

func (br *AIBlockReward) setRewards(nodelist []common.Address, oneNodeReward *big.Int) map[common.Address]*big.Int {
	rewards := make(map[common.Address]*big.Int)

	for _, k := range nodelist {
		rewards[k] = oneNodeReward
		//log.Debug(PackageName, "计算奖励金额,账户", k, "奖励金额", oneNodeReward)
	}
	return rewards
}
