// Copyright (c) 2018-2019 The MATRIX Authors
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
// 广播交易
type operatorBroadcastTx struct {
	key common.Hash
}

func newBroadcastTxOpt() *operatorBroadcastTx {
	return &operatorBroadcastTx{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBroadcastTx),
	}
}

func (opt *operatorBroadcastTx) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	value := make(map[string]map[common.Address][]byte)
	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return value, nil
	}
	if err := json.Unmarshal(data, &value); err != nil {
		log.Error(logInfo, "broadcastTx unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBroadcastTx) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	txs, OK := value.(map[string]map[common.Address][]byte)
	if !OK {
		log.Error(logInfo, "input param(broadcastTx) err", "reflect failed")
		return ErrParamReflect
	}
	data, err := json.Marshal(txs)
	if err != nil {
		log.Error(logInfo, "broadcastTx marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 广播区块周期
type operatorBroadcastInterval struct {
	key common.Hash
}

func newBroadcastIntervalOpt() *operatorBroadcastInterval {
	return &operatorBroadcastInterval{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBroadcastInterval),
	}
}

func (opt *operatorBroadcastInterval) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "broadcastInterval data", "is empty")
		return nil, ErrDataEmpty
	}
	value := new(mc.BCIntervalInfo)
	if err := json.Unmarshal(data, &value); err != nil {
		log.Error(logInfo, "broadcastInterval unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBroadcastInterval) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	interval, OK := value.(*mc.BCIntervalInfo)
	if !OK {
		log.Error(logInfo, "input param(broadcastInterval) err", "reflect failed")
		return ErrParamReflect
	}
	if interval == nil {
		log.Error(logInfo, "input param(broadcastInterval) err", "is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(interval)
	if err != nil {
		log.Error(logInfo, "broadcastInterval marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 广播账户
type operatorBroadcastAccount struct {
	key common.Hash
}

func newBroadcastAccountOpt() *operatorBroadcastAccount {
	return &operatorBroadcastAccount{
		key: types.RlpHash(matrixStatePrefix + mc.OldMSKeyAccountBroadcast),
	}
}

func (opt *operatorBroadcastAccount) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		// 广播账户数据不可为空
		log.Error(logInfo, "broadcastAccount data", "is empty")
		return nil, ErrDataEmpty
	}
	account, err := decodeAccount(data)
	if err != nil {
		log.Error(logInfo, "broadcastAccount decode failed", err)
		return nil, err
	}
	if account == (common.Address{}) {
		log.Error(logInfo, "broadcastAccount", "is empty account")
		return nil, ErrAccountNil
	}
	return account, nil
}

func (opt *operatorBroadcastAccount) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}
	account, OK := value.(common.Address)
	if !OK {
		log.Error(logInfo, "input param(broadcastAccount) err", "reflect failed")
		return ErrParamReflect
	}
	if account == (common.Address{}) {
		log.Error(logInfo, "input param(broadcastAccount) err", "account is empty account")
		return ErrAccountNil
	}

	data, err := encodeAccount(account)
	if err != nil {
		log.Error(logInfo, "broadcastAccount encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 广播账户新版
type operatorBroadcastAccountV2 struct {
	key common.Hash
}

func newBroadcastAccountOptV2() *operatorBroadcastAccountV2 {
	return &operatorBroadcastAccountV2{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyAccountBroadcasts),
	}
}

func (opt *operatorBroadcastAccountV2) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		// 广播账户数据不可为空
		log.Error(logInfo, "broadcastAccounts data", "is empty")
		return nil, ErrDataEmpty
	}
	accounts, err := decodeAccounts(data)
	if err != nil {
		log.Error(logInfo, "broadcastAccounts decode failed", err)
		return nil, err
	}
	if len(accounts) == 0 {
		log.Error(logInfo, "broadcastAccounts size", "is empty")
		return nil, ErrAccountNil
	}
	return accounts, nil
}

func (opt *operatorBroadcastAccountV2) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}
	accounts, OK := value.([]common.Address)
	if !OK {
		log.Error(logInfo, "input param(broadcastAccounts) err", "reflect failed")
		return ErrParamReflect
	}
	if len(accounts) == 0 {
		log.Error(logInfo, "input param(broadcastAccounts) err", "account is empty account")
		return ErrAccountNil
	}

	data, err := encodeAccounts(accounts)
	if err != nil {
		log.Error(logInfo, "broadcastAccounts encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 前广播区块root信息
type operatorPreBroadcastRoot struct {
	key common.Hash
}

func newPreBroadcastRootOpt() *operatorPreBroadcastRoot {
	return &operatorPreBroadcastRoot{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyPreBroadcastRoot),
	}
}

func (opt *operatorPreBroadcastRoot) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	value := new(mc.PreBroadStateRoot)
	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return value, nil
	}

	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "preBroadcastRoot unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorPreBroadcastRoot) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	roots, OK := value.(*mc.PreBroadStateRoot)
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
