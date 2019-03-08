package matrixstate

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
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
func GetCoinConfig(st StateDB) ([]common.CoinConfig, error) {
	version := GetVersionInfo(st)
	mgr := GetManager(version)
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSCurrencyConfig)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.([]common.CoinConfig), nil
}
