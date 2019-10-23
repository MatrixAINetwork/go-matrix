// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixstate

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/rlp"
)

/////////////////////////////////////////////////////////////////////////////////////////
//
type operatorBasePowerStatsStatus struct {
	key common.Hash
}

func newBasePowerStatsStatusOpt() *operatorBasePowerStatsStatus {
	return &operatorBasePowerStatsStatus{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBasePowerStatsStatus),
	}
}

func (opt *operatorBasePowerStatsStatus) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBasePowerStatsStatus) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.BasePowerSlashStatsStatus{Number: 0}, nil
	}

	value := new(mc.BasePowerSlashStatsStatus)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "basePowerStatsStatus rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBasePowerStatsStatus) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "basePowerStatsStatus rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//
type operatorBasePowerSlashCfg struct {
	key common.Hash
}

func newBasePowerSlashCfgOpt() *operatorBasePowerSlashCfg {
	return &operatorBasePowerSlashCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBasePowerSlashCfg),
	}
}

func (opt *operatorBasePowerSlashCfg) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBasePowerSlashCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.BasePowerSlashCfg{Switcher: false, LowTHR: 1, ProhibitCycleNum: 2}, nil
	}

	value := new(mc.BasePowerSlashCfg)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "basePowerSlashCfg rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBasePowerSlashCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "basePowerSlashCfg rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//
type operatorBasePowerStats struct {
	key common.Hash
}

func newBasePowerStatsOpt() *operatorBasePowerStats {
	return &operatorBasePowerStats{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBasePowerStats),
	}
}

func (opt *operatorBasePowerStats) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBasePowerStats) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.BasePowerStats{StatsList: make([]mc.BasePowerNum, 0)}, nil
	}

	value := new(mc.BasePowerStats)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "basePowerStats rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBasePowerStats) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "basePowerStats rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//
type operatorBasePowerBlackList struct {
	key common.Hash
}

func newBasePowerBlackListOpt() *operatorBasePowerBlackList {
	return &operatorBasePowerBlackList{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBasePowerBlackList),
	}
}

func (opt *operatorBasePowerBlackList) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBasePowerBlackList) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.BasePowerSlashBlackList{BlackList: make([]mc.BasePowerSlash, 0)}, nil
	}

	value := new(mc.BasePowerSlashBlackList)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "basePowerBlackList rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBasePowerBlackList) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "basePowerBlackList rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}
