// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixstate

import "github.com/MatrixAINetwork/go-matrix/mc"

func GetBlockProduceStatsStatus(st StateDB) (*mc.BlockProduceSlashStatsStatus, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlockProduceStatsStatus)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.BlockProduceSlashStatsStatus), nil
}

func SetBlockProduceStatsStatus(st StateDB, status *mc.BlockProduceSlashStatsStatus) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlockProduceStatsStatus)
	if err != nil {
		return err
	}
	return opt.SetValue(st, status)
}

func GetBlockProduceSlashCfg(st StateDB) (*mc.BlockProduceSlashCfg, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlockProduceSlashCfg)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.BlockProduceSlashCfg), nil
}

func SetBlockProduceSlashCfg(st StateDB, cfg *mc.BlockProduceSlashCfg) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlockProduceSlashCfg)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

func GetBlockProduceStats(st StateDB) (*mc.BlockProduceStats, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlockProduceStats)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.BlockProduceStats), nil
}

func SetBlockProduceStats(st StateDB, status *mc.BlockProduceStats) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlockProduceStats)
	if err != nil {
		return err
	}
	return opt.SetValue(st, status)
}

func GetBlockProduceBlackList(st StateDB) (*mc.BlockProduceSlashBlackList, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlockProduceBlackList)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.BlockProduceSlashBlackList), nil
}

func SetBlockProduceBlackList(st StateDB, status *mc.BlockProduceSlashBlackList) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlockProduceBlackList)
	if err != nil {
		return err
	}
	return opt.SetValue(st, status)
}
