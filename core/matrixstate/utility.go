// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package matrixstate

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/pkg/errors"
)

func ReaderVersionInfo(st StateDB) string {
	value, err := versionOpt.GetValue(st)
	if err != nil {
		log.Error(logInfo, "get version failed", err)
		return ""
	}

	version, _ := value.(string)
	if len(version) == 0 {
		// 第一版本state中没有版本信息
		log.Debug(logInfo, "not version in state", "使用Alpha版本")
		return manparams.VersionAlpha
	}
	return version
}

func GetBroadcastInterval(st StateDB) (*mc.BCIntervalInfo, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBroadcastInterval)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.BCIntervalInfo), nil
}

func SetBroadcastInterval(st StateDB, interval *mc.BCIntervalInfo) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBroadcastInterval)
	if err != nil {
		return err
	}
	return opt.SetValue(st, interval)
}

func GetBroadcastAccounts(st StateDB) ([]common.Address, error) {
	version := ReaderVersionInfo(st)
	mgr := GetManager(version)
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountBroadcasts)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}

	if version == manparams.VersionAlpha {
		// Alpha版，广播节点为单common.Address
		return []common.Address{value.(common.Address)}, nil
	} else {
		return value.([]common.Address), nil
	}
}

func SetBroadcastAccounts(st StateDB, accounts []common.Address) error {
	version := ReaderVersionInfo(st)
	mgr := GetManager(version)
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountBroadcasts)
	if err != nil {
		return err
	}
	if version == manparams.VersionAlpha {
		// Alpha版，广播节点为单common.Address
		if len(accounts) == 0 {
			return errors.New("account size is 0")
		}
		log.Info(logInfo, "Alpha版广播节点设置", "只保存第一个账户", "广播账户", accounts[0].Hex())
		return opt.SetValue(st, accounts[0])
	} else {
		return opt.SetValue(st, accounts)
	}
}

func GetInnerMinerAccounts(st StateDB) ([]common.Address, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountInnerMiners)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.([]common.Address), nil
}

func SetInnerMinerAccounts(st StateDB, accounts []common.Address) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountInnerMiners)
	if err != nil {
		return err
	}
	return opt.SetValue(st, accounts)
}

func GetFoundationAccount(st StateDB) (common.Address, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return common.Address{}, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountFoundation)
	if err != nil {
		return common.Address{}, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return common.Address{}, err
	}
	return value.(common.Address), nil
}

func SetFoundationAccount(st StateDB, account common.Address) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountFoundation)
	if err != nil {
		return err
	}
	return opt.SetValue(st, account)
}

func GetVersionSuperAccounts(st StateDB) ([]common.Address, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountVersionSupers)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.([]common.Address), nil
}

func SetVersionSuperAccounts(st StateDB, accounts []common.Address) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountVersionSupers)
	if err != nil {
		return err
	}
	return opt.SetValue(st, accounts)
}

func GetBlockSuperAccounts(st StateDB) ([]common.Address, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountBlockSupers)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.([]common.Address), nil
}

func SetBlockSuperAccounts(st StateDB, accounts []common.Address) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountBlockSupers)
	if err != nil {
		return err
	}
	return opt.SetValue(st, accounts)
}

func GetPreBroadcastRoot(st StateDB) (*mc.PreBroadStateRoot, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyPreBroadcastRoot)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.PreBroadStateRoot), nil
}

func GetLeaderConfig(st StateDB) (*mc.LeaderConfig, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyLeaderConfig)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.LeaderConfig), nil
}

func SetLeaderConfig(st StateDB, cfg *mc.LeaderConfig) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyLeaderConfig)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

func GetSuperBlockCfg(st StateDB) (*mc.SuperBlkCfg, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeySuperBlockCfg)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.SuperBlkCfg), nil
}

func SetSuperBlockCfg(st StateDB, cfg *mc.SuperBlkCfg) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeySuperBlockCfg)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

func GetBroadcastTxs(st StateDB) (map[string]map[common.Address][]byte, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBroadcastTx)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(map[string]map[common.Address][]byte), nil
}

func SetBroadcastTxs(st StateDB, txs map[string]map[common.Address][]byte) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBroadcastTx)
	if err != nil {
		return err
	}
	return opt.SetValue(st, txs)
}
