// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixstate

import (
	"math/big"

	"encoding/json"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

func GetVersionInfo(st StateDB) string {
	value, err := versionOpt.GetValue(st)
	if err != nil {
		log.Error(logInfo, "get version failed", err)
		return ""
	}
	return value.(string)
}

func SetVersionInfo(st StateDB, version string) error {
	return versionOpt.SetValue(st, version)
}

func GetBroadcastInterval(st StateDB) (*mc.BCIntervalInfo, error) {
	mgr := GetManager(GetVersionInfo(st))
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

func GetBroadcastIntervalByVersion(st StateDB, version string) (*mc.BCIntervalInfo, error) {
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
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
	version := GetVersionInfo(st)
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
	return value.([]common.Address), nil
}

func SetBroadcastAccounts(st StateDB, accounts []common.Address) error {
	version := GetVersionInfo(st)
	mgr := GetManager(version)
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountBroadcasts)
	if err != nil {
		return err
	}
	return opt.SetValue(st, accounts)
}

func GetInnerMinerAccounts(st StateDB) ([]common.Address, error) {
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountBlockSupers)
	if err != nil {
		return err
	}
	return opt.SetValue(st, accounts)
}

func GetMultiCoinSuperAccounts(st StateDB) ([]common.Address, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountMultiCoinSupers)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.([]common.Address), nil
}

func SetMultiCoinSuperAccounts(st StateDB, accounts []common.Address) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountMultiCoinSupers)
	if err != nil {
		return err
	}
	return opt.SetValue(st, accounts)
}

func GetSubChainSuperAccounts(st StateDB) ([]common.Address, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountSubChainSupers)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.([]common.Address), nil
}

func SetSubChainSuperAccounts(st StateDB, accounts []common.Address) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyAccountSubChainSupers)
	if err != nil {
		return err
	}
	return opt.SetValue(st, accounts)
}

func GetPreBroadcastRoot(st StateDB) (*mc.PreBroadStateRoot, error) {
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
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
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeySuperBlockCfg)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

func GetMinDifficulty(st StateDB) (*big.Int, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyMinimumDifficulty)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*big.Int), nil
}

func SetMinDifficulty(st StateDB, minDifficulty *big.Int) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyMinimumDifficulty)
	if err != nil {
		return err
	}
	return opt.SetValue(st, minDifficulty)
}

func GetMaxDifficulty(st StateDB) (*big.Int, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyMaximumDifficulty)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*big.Int), nil
}

func SetMaxDifficulty(st StateDB, maxDifficulty *big.Int) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyMaximumDifficulty)
	if err != nil {
		return err
	}
	return opt.SetValue(st, maxDifficulty)
}

func GetReelectionDifficulty(st StateDB) (*big.Int, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyReelectionDifficulty)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*big.Int), nil
}

func SetReelectionDifficulty(st StateDB, reelectionDifficulty *big.Int) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyReelectionDifficulty)
	if err != nil {
		return err
	}
	return opt.SetValue(st, reelectionDifficulty)
}

//func GetBroadcastTxs(st StateDB) (map[string]map[common.Address][]byte, error)
func GetBroadcastTxs(st StateDB) (common.BroadTxSlice, error) {
	mgr := GetManager(GetVersionInfo(st))
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
	return value.(common.BroadTxSlice), nil
}

func SetBroadcastTxs(st StateDB, txs common.BroadTxSlice) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBroadcastTx)
	if err != nil {
		return err
	}
	return opt.SetValue(st, txs)
}

func GetTxpoolGasLimit(st StateDB) (*big.Int, error) {
	version := GetVersionInfo(st)
	mgr := GetManager(version)
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSTxpoolGasLimitCfg)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil || value == "0" {
		return nil, err
	}
	return value.(*big.Int), nil
}
func GetAccountBlackList(st StateDB) ([]common.Address, error) {
	version := GetVersionInfo(st)
	mgr := GetManager(version)
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSAccountBlackList)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.([]common.Address), nil
}

//func GetCoinConfig(st StateDB) ([]common.CoinConfig, error) {
//	version := GetVersionInfo(st)
//	mgr := GetManager(version)
//	if mgr == nil {
//		return nil, ErrFindManager
//	}
//	opt, err := mgr.FindOperator(mc.MSCurrencyConfig)
//	if err != nil {
//		return nil, err
//	}
//	value, err := opt.GetValue(st)
//	if err != nil {
//		return nil, err
//	}
//	return value.([]common.CoinConfig), nil
//}
func GetCoinConfig(st StateDB) ([]common.CoinConfig, error) {
	coinconfig := st.GetMatrixData(types.RlpHash(common.COINPREFIX + mc.MSCurrencyConfig))
	var coincfglist []common.CoinConfig
	if len(coinconfig) > 0 {
		err := json.Unmarshal(coinconfig, &coincfglist)
		if err != nil {
			log.Trace("get coin config list", "unmarshal err", err)
			return nil, err
		}
	}
	return coincfglist, nil
}
func GetCurrenyHeader(st StateDB) (*mc.CurrencyHeader, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSCurrencyHeader)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.CurrencyHeader), nil
}

func SetCurrenyHeader(st StateDB, cfg *mc.CurrencyHeader) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSCurrencyHeader)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

func GetBlockDuration(st StateDB) (*mc.BlockDurationStatus, error) {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlockDurationStatus)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.BlockDurationStatus), nil
}

func SetBlockDuration(st StateDB, duration *mc.BlockDurationStatus) error {
	mgr := GetManager(GetVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyBlockDurationStatus)
	if err != nil {
		return err
	}
	return opt.SetValue(st, duration)
}
