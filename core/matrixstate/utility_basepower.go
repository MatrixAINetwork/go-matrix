// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixstate

import "github.com/MatrixAINetwork/go-matrix/mc"

func GetBasePowerStatsStatus(st StateDB) (*mc.BasePowerSlashStatsStatus, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBasePowerStatsStatus)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.BasePowerSlashStatsStatus), nil
}

func SetBasePowerStatsStatus(st StateDB, status *mc.BasePowerSlashStatsStatus) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBasePowerStatsStatus)
	if err != nil {
		return err
	}
	return opt.SetValue(st, status)
}

func GetBasePowerSlashCfg(st StateDB) (*mc.BasePowerSlashCfg, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBasePowerSlashCfg)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.BasePowerSlashCfg), nil
}

func SetBasePowerSlashCfg(st StateDB, cfg *mc.BasePowerSlashCfg) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBasePowerSlashCfg)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

func GetBasePowerStats(st StateDB) (*mc.BasePowerStats, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBasePowerStats)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.BasePowerStats), nil
}

func SetBasePowerStats(st StateDB, status *mc.BasePowerStats) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBasePowerStats)
	if err != nil {
		return err
	}
	return opt.SetValue(st, status)
}

func GetBasePowerBlackList(st StateDB) (*mc.BasePowerSlashBlackList, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBasePowerBlackList)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.BasePowerSlashBlackList), nil
}

func SetBasePowerBlackList(st StateDB, status *mc.BasePowerSlashBlackList) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBasePowerBlackList)
	if err != nil {
		return err
	}
	return opt.SetValue(st, status)
}
