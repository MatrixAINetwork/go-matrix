// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixstate

import (
	"github.com/MatrixAINetwork/go-matrix/mc"
)

/////////////////////////////////////////////////////////////////////
// 区块奖励相关
func GetBlkRewardCfg(st StateDB) (*mc.BlkRewardCfg, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlkRewardCfg)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.BlkRewardCfg), nil
}

func SetBlkRewardCfg(st StateDB, cfg *mc.BlkRewardCfg) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlkRewardCfg)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

/////////////////////////////////////////////////////////////////////
// 交易奖励相关
func GetTxsRewardCfg(st StateDB) (*mc.TxsRewardCfg, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyTxsRewardCfg)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.TxsRewardCfg), nil
}

func SetTxsRewardCfg(st StateDB, cfg *mc.TxsRewardCfg) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyTxsRewardCfg)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

/////////////////////////////////////////////////////////////////////
// 利息相关
func GetInterestCfg(st StateDB) (*mc.InterestCfg, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyInterestCfg)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.InterestCfg), nil
}

func SetInterestCfg(st StateDB, cfg *mc.InterestCfg) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyInterestCfg)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

func GetInterestCalcNum(st StateDB) (uint64, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return uint64(0), ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyInterestCalcNum)
	if err != nil {
		return uint64(0), err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return uint64(0), err
	}
	return value.(uint64), nil
}

func SetInterestCalcNum(st StateDB, num uint64) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyInterestCalcNum)
	if err != nil {
		return err
	}
	return opt.SetValue(st, num)
}

func GetInterestPayNum(st StateDB) (uint64, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return uint64(0), ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyInterestPayNum)
	if err != nil {
		return uint64(0), err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return uint64(0), err
	}
	return value.(uint64), nil
}

func SetInterestPayNum(st StateDB, num uint64) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyInterestPayNum)
	if err != nil {
		return err
	}
	return opt.SetValue(st, num)
}

/////////////////////////////////////////////////////////////////////
// 彩票相关
func GetLotteryCfg(st StateDB) (*mc.LotteryCfg, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyLotteryCfg)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.LotteryCfg), nil
}

func SetLotteryCfg(st StateDB, cfg *mc.LotteryCfg) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyLotteryCfg)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

func GetLotteryAccount(st StateDB) (*mc.LotteryFrom, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyLotteryAccount)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.LotteryFrom), nil
}

func SetLotteryAccount(st StateDB, account *mc.LotteryFrom) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyLotteryAccount)
	if err != nil {
		return err
	}
	return opt.SetValue(st, account)
}

func GetLotteryNum(st StateDB) (uint64, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return uint64(0), ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyLotteryNum)
	if err != nil {
		return uint64(0), err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return uint64(0), err
	}
	return value.(uint64), nil
}

func SetLotteryNum(st StateDB, num uint64) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyLotteryNum)
	if err != nil {
		return err
	}
	return opt.SetValue(st, num)
}

/////////////////////////////////////////////////////////////////////
// 惩罚相关
func GetSlashCfg(st StateDB) (*mc.SlashCfg, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeySlashCfg)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.SlashCfg), nil
}

func SetSlashCfg(st StateDB, cfg *mc.SlashCfg) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeySlashCfg)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

func GetSlashNum(st StateDB) (uint64, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return uint64(0), ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeySlashNum)
	if err != nil {
		return uint64(0), err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return uint64(0), err
	}
	return value.(uint64), nil
}

func SetSlashNum(st StateDB, num uint64) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeySlashNum)
	if err != nil {
		return err
	}
	return opt.SetValue(st, num)
}

func GetUpTimeNum(st StateDB) (uint64, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return uint64(0), ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyUpTimeNum)
	if err != nil {
		return uint64(0), err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return uint64(0), err
	}
	return value.(uint64), nil
}

func SetUpTimeNum(st StateDB, num uint64) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyUpTimeNum)
	if err != nil {
		return err
	}
	return opt.SetValue(st, num)
}

/////////////////////////////////////////////////////////////////////
//
func GetPreMinerBlkReward(st StateDB) (*mc.MinerOutReward, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyPreMinerBlkReward)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.MinerOutReward), nil
}

func SetPreMinerBlkReward(st StateDB, reward *mc.MinerOutReward) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyPreMinerBlkReward)
	if err != nil {
		return err
	}
	return opt.SetValue(st, reward)
}

func GetPreMinerTxsReward(st StateDB) (*mc.MinerOutReward, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyPreMinerTxsReward)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.MinerOutReward), nil
}

func SetPreMinerTxsReward(st StateDB, reward *mc.MinerOutReward) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyPreMinerTxsReward)
	if err != nil {
		return err
	}
	return opt.SetValue(st, reward)
}

func GetBlkCalc(st StateDB) (string, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return "0", ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlkCalc)
	if err != nil {
		return "0", err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return "0", err
	}
	return value.(string), nil
}

func SetPreMinerMultiCoinTxsReward(st StateDB, reward []mc.MultiCoinMinerOutReward) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyPreMinerTxsReward)
	if err != nil {
		return err
	}
	return opt.SetValue(st, reward)
}

func GetPreMinerMultiCoinTxsReward(st StateDB) ([]mc.MultiCoinMinerOutReward, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyPreMinerTxsReward)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.([]mc.MultiCoinMinerOutReward), nil
}

func SetBlkCalc(st StateDB, Calc string) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlkCalc)
	if err != nil {
		return err
	}
	return opt.SetValue(st, Calc)
}

func GetTxsCalc(st StateDB) (string, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return "0", ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyTxsCalc)
	if err != nil {
		return "0", err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return "0", err
	}
	return value.(string), nil
}

func SetTxsCalc(st StateDB, Calc string) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyTxsCalc)
	if err != nil {
		return err
	}
	return opt.SetValue(st, Calc)
}

func GetInterestCalc(st StateDB) (string, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return "0", ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyInterestCalc)
	if err != nil {
		return "0", err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return "0", err
	}
	return value.(string), nil
}

func SetInterestCalc(st StateDB, Calc string) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyInterestCalc)
	if err != nil {
		return err
	}
	return opt.SetValue(st, Calc)
}

func GetLotteryCalc(st StateDB) (string, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return "0", ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyLotteryCalc)
	if err != nil {
		return "0", err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return "0", err
	}
	return value.(string), nil
}

func SetLotteryCalc(st StateDB, Calc string) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyLotteryCalc)
	if err != nil {
		return err
	}
	return opt.SetValue(st, Calc)
}

func GetSlashCalc(st StateDB) (string, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return "0", ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeySlashCalc)
	if err != nil {
		return "0", err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return "0", err
	}
	return value.(string), nil
}

func SetSlashCalc(st StateDB, Calc string) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeySlashCalc)
	if err != nil {
		return err
	}
	return opt.SetValue(st, Calc)
}
