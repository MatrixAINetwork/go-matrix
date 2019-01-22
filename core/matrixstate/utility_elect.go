// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package matrixstate

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mc"
)

func GetTopologyGraph(st StateDB) (*mc.TopologyGraph, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyTopologyGraph)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.TopologyGraph), nil
}

func SetTopologyGraph(st StateDB, graph *mc.TopologyGraph) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyTopologyGraph)
	if err != nil {
		return err
	}
	return opt.SetValue(st, graph)
}

func GetElectGraph(st StateDB) (*mc.ElectGraph, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectGraph)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.ElectGraph), nil
}

func SetElectGraph(st StateDB, graph *mc.ElectGraph) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectGraph)
	if err != nil {
		return err
	}
	return opt.SetValue(st, graph)
}

func GetElectOnlineState(st StateDB) (*mc.ElectOnlineStatus, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectOnlineState)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.ElectOnlineStatus), nil
}

func SetElectOnlineState(st StateDB, onlineState *mc.ElectOnlineStatus) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectOnlineState)
	if err != nil {
		return err
	}
	return opt.SetValue(st, onlineState)
}

func GetElectGenTime(st StateDB) (*mc.ElectGenTimeStruct, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectGenTime)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.ElectGenTimeStruct), nil
}

func SetElectGenTime(st StateDB, genTime *mc.ElectGenTimeStruct) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectGraph)
	if err != nil {
		return err
	}
	return opt.SetValue(st, genTime)
}

func GetElectConfigInfo(st StateDB) (*mc.ElectConfigInfo, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectConfigInfo)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.ElectConfigInfo), nil
}

func SetElectConfigInfo(st StateDB, cfg *mc.ElectConfigInfo) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectConfigInfo)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfg)
}

func GetElectMinerNum(st StateDB) (*mc.ElectMinerNumStruct, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectMinerNum)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.ElectMinerNumStruct), nil
}

func SetElectMinerNum(st StateDB, num *mc.ElectMinerNumStruct) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectMinerNum)
	if err != nil {
		return err
	}
	return opt.SetValue(st, num)
}

func GetElectWhiteList(st StateDB) ([]common.Address, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectWhiteList)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.([]common.Address), nil
}

func SetElectWhiteList(st StateDB, accounts []common.Address) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectWhiteList)
	if err != nil {
		return err
	}
	return opt.SetValue(st, accounts)
}

func GetElectBlackList(st StateDB) ([]common.Address, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectBlackList)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.([]common.Address), nil
}

func SetElectBlackList(st StateDB, accounts []common.Address) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyElectBlackList)
	if err != nil {
		return err
	}
	return opt.SetValue(st, accounts)
}

func GetVIPConfig(st StateDB) ([]mc.VIPConfig, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyVIPConfig)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.([]mc.VIPConfig), nil
}

func SetVIPConfig(st StateDB, cfgs []mc.VIPConfig) error {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyVIPConfig)
	if err != nil {
		return err
	}
	return opt.SetValue(st, cfgs)
}

func GetMinHash(st StateDB) (*mc.RandomInfoStruct, error) {
	mgr := GetManager(ReaderVersionInfo(st))
	if mgr == nil {
		return nil, ErrFindManager
	}
	opt, err := mgr.FindOperator(mc.MSKeyMinHash)
	if err != nil {
		return nil, err
	}
	value, err := opt.GetValue(st)
	if err != nil {
		return nil, err
	}
	return value.(*mc.RandomInfoStruct), nil
}
