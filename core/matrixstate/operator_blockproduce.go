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
type operatorBlockProduceStatsStatus struct {
	key common.Hash
}

func newBlockProduceStatsStatusOpt() *operatorBlockProduceStatsStatus {
	return &operatorBlockProduceStatsStatus{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBlockProduceStatsStatus),
	}
}

func (opt *operatorBlockProduceStatsStatus) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBlockProduceStatsStatus) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.BlockProduceSlashStatsStatus{Number: 0}, nil
	}

	value := new(mc.BlockProduceSlashStatsStatus)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "blockProduceStatsStatus rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBlockProduceStatsStatus) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "blockProduceStatsStatus rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//
type operatorBlockProduceSlashCfg struct {
	key common.Hash
}

func newBlockProduceSlashCfgOpt() *operatorBlockProduceSlashCfg {
	return &operatorBlockProduceSlashCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBlockProduceSlashCfg),
	}
}

func (opt *operatorBlockProduceSlashCfg) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBlockProduceSlashCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.BlockProduceSlashCfg{Switcher: false, LowTHR: 1, ProhibitCycleNum: 2}, nil
	}

	value := new(mc.BlockProduceSlashCfg)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "blockProduceSlashCfg rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBlockProduceSlashCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "blockProduceSlashCfg rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//
type operatorBlockProduceStats struct {
	key common.Hash
}

func newBlockProduceStatsOpt() *operatorBlockProduceStats {
	return &operatorBlockProduceStats{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBlockProduceStats),
	}
}

func (opt *operatorBlockProduceStats) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBlockProduceStats) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.BlockProduceStats{StatsList: make([]mc.UserBlockProduceNum, 0)}, nil
	}

	value := new(mc.BlockProduceStats)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "blockProduceStats rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBlockProduceStats) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "blockProduceStats rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//
type operatorBlockProduceBlackList struct {
	key common.Hash
}

func newBlockProduceBlackListOpt() *operatorBlockProduceBlackList {
	return &operatorBlockProduceBlackList{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBlockProduceBlackList),
	}
}

func (opt *operatorBlockProduceBlackList) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBlockProduceBlackList) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.BlockProduceSlashBlackList{BlackList: make([]mc.UserBlockProduceSlash, 0)}, nil
	}

	value := new(mc.BlockProduceSlashBlackList)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "blockProduceBlackList rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBlockProduceBlackList) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "blockProduceBlackList rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}
