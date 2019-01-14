// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package matrixstate

import (
	"encoding/json"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

/////////////////////////////////////////////////////////////////////////////////////////
// 版本信息
type operatorVersionInfo struct {
	key common.Hash
}

func newVersionInfoOpt() *operatorVersionInfo {
	return &operatorVersionInfo{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyVersionInfo),
	}
}

func (opt *operatorVersionInfo) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	return string(data[:]), nil
}

func (opt *operatorVersionInfo) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	version, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(versionInfo) err", "reflect failed")
		return ErrParamReflect
	}
	st.SetMatrixData(opt.key, []byte(version))
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 基金会矿工
type operatorInnerMinerAccounts struct {
	key common.Hash
}

func newInnerMinerAccountsOpt() *operatorInnerMinerAccounts {
	return &operatorInnerMinerAccounts{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyAccountInnerMiners),
	}
}

func (opt *operatorInnerMinerAccounts) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return make([]common.Address, 0), nil
	}
	accounts, err := decodeAccounts(data)
	if err != nil {
		log.Error(logInfo, "innerMinerAccounts decode failed", err)
		return nil, err
	}
	return accounts, nil
}

func (opt *operatorInnerMinerAccounts) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	accounts, OK := value.([]common.Address)
	if !OK {
		log.Error(logInfo, "input param(innerMinerAccounts) err", "reflect failed")
		return ErrParamReflect
	}
	data, err := encodeAccounts(accounts)
	if err != nil {
		log.Error(logInfo, "innerMinerAccounts encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 基金会账户
type operatorFoundationAccount struct {
	key common.Hash
}

func newFoundationAccountOpt() *operatorFoundationAccount {
	return &operatorFoundationAccount{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyAccountFoundation),
	}
}

func (opt *operatorFoundationAccount) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return common.Address{}, nil
	}
	account, err := decodeAccount(data)
	if err != nil {
		log.Error(logInfo, "FoundationAccount decode failed", err)
		return nil, err
	}
	return account, nil
}

func (opt *operatorFoundationAccount) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	account, OK := value.(common.Address)
	if !OK {
		log.Error(logInfo, "input param(FoundationAccount) err", "reflect failed")
		return ErrParamReflect
	}
	data, err := encodeAccount(account)
	if err != nil {
		log.Error(logInfo, "FoundationAccount encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 版本签名账户
type operatorVersionSuperAccounts struct {
	key common.Hash
}

func newVersionSuperAccountsOpt() *operatorVersionSuperAccounts {
	return &operatorVersionSuperAccounts{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyAccountVersionSupers),
	}
}

func (opt *operatorVersionSuperAccounts) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return make([]common.Address, 0), nil
	}
	accounts, err := decodeAccounts(data)
	if err != nil {
		log.Error(logInfo, "versionSuperAccounts decode failed", err)
		return nil, err
	}
	return accounts, nil
}

func (opt *operatorVersionSuperAccounts) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	accounts, OK := value.([]common.Address)
	if !OK {
		log.Error(logInfo, "input param(versionSuperAccounts) err", "reflect failed")
		return ErrParamReflect
	}
	if len(accounts) == 0 {
		log.Error(logInfo, "input param(versionSuperAccounts) err", "accounts is empty")
		return ErrParamReflect
	}
	data, err := encodeAccounts(accounts)
	if err != nil {
		log.Error(logInfo, "versionSuperAccounts encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 超级区块签名账户
type operatorBlockSuperAccounts struct {
	key common.Hash
}

func newBlockSuperAccountsOpt() *operatorBlockSuperAccounts {
	return &operatorBlockSuperAccounts{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyAccountBlockSupers),
	}
}

func (opt *operatorBlockSuperAccounts) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return make([]common.Address, 0), nil
	}
	accounts, err := decodeAccounts(data)
	if err != nil {
		log.Error(logInfo, "blockSuperAccounts decode failed", err)
		return nil, err
	}
	return accounts, nil
}

func (opt *operatorBlockSuperAccounts) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	accounts, OK := value.([]common.Address)
	if !OK {
		log.Error(logInfo, "input param(blockSuperAccounts) err", "reflect failed")
		return ErrParamReflect
	}
	if len(accounts) == 0 {
		log.Error(logInfo, "input param(blockSuperAccounts) err", "accounts is empty")
		return ErrParamReflect
	}
	data, err := encodeAccounts(accounts)
	if err != nil {
		log.Error(logInfo, "blockSuperAccounts encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// leader服务配置信息
type operatorLeaderConfig struct {
	key common.Hash
}

func newLeaderConfigOpt() *operatorLeaderConfig {
	return &operatorLeaderConfig{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyLeaderConfig),
	}
}

func (opt *operatorLeaderConfig) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "leaderConfig data", "is empty")
		return nil, ErrDataEmpty
	}

	value := new(mc.LeaderConfig)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "leaderConfig unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorLeaderConfig) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	roots, OK := value.(*mc.LeaderConfig)
	if !OK {
		log.Error(logInfo, "input param(preBroadcastRoot) err", "reflect failed")
		return ErrParamReflect
	}
	data, err := json.Marshal(roots)
	if err != nil {
		log.Error(logInfo, "preBroadcastRoot marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 最小hash
type operatorMinHash struct {
	key common.Hash
}

func newMinHashOpt() *operatorMinHash {
	return &operatorMinHash{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyMinHash),
	}
}

func (opt *operatorMinHash) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	value := new(mc.RandomInfoStruct)
	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return value, nil
	}

	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "minHash unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorMinHash) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	info, OK := value.(*mc.RandomInfoStruct)
	if !OK {
		log.Error(logInfo, "input param(minHash) err", "reflect failed")
		return ErrParamReflect
	}
	data, err := json.Marshal(info)
	if err != nil {
		log.Error(logInfo, "minHash marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 超级区块配置
type operatorSuperBlockCfg struct {
	key common.Hash
}

func newSuperBlockCfgOpt() *operatorSuperBlockCfg {
	return &operatorSuperBlockCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeySuperBlockCfg),
	}
}

func (opt *operatorSuperBlockCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.SuperBlkCfg{Seq: 0, Num: 0}, nil
	}

	value := new(mc.SuperBlkCfg)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "superBlkCfg unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorSuperBlockCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	cfg, OK := value.(*mc.SuperBlkCfg)
	if !OK {
		log.Error(logInfo, "input param(superBlkCfg) err", "reflect failed")
		return ErrParamReflect
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Error(logInfo, "superBlkCfg marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}
