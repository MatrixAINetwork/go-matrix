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
// 广播交易
type operatorBroadcastTx struct {
	key common.Hash
}

func newBroadcastTxOpt() *operatorBroadcastTx {
	return &operatorBroadcastTx{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBroadcastTx),
	}
}

func (opt *operatorBroadcastTx) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBroadcastTx) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	//value := make(map[string]map[common.Address][]byte)
	value := make(common.BroadTxSlice, 0)
	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return value, nil
	}
	if err := rlp.DecodeBytes(data, &value); err != nil {
		log.Error(logInfo, "broadcastTx rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBroadcastTx) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	//txs, OK := value.(map[string]map[common.Address][]byte)
	txs, OK := value.(common.BroadTxSlice)
	if !OK {
		log.Error(logInfo, "input param(broadcastTx) err", "reflect failed")
		return ErrParamReflect
	}
	data, err := rlp.EncodeToBytes(txs)
	if err != nil {
		log.Error(logInfo, "broadcastTx rlp encode failed", err)
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

func (opt *operatorBroadcastInterval) KeyHash() common.Hash {
	return opt.key
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
	if err := rlp.DecodeBytes(data, &value); err != nil {
		log.Error(logInfo, "broadcastInterval rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBroadcastInterval) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "broadcastInterval rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 广播账户
type operatorBroadcastAccounts struct {
	key common.Hash
}

func newBroadcastAccountsOpt() *operatorBroadcastAccounts {
	return &operatorBroadcastAccounts{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyAccountBroadcasts),
	}
}

func (opt *operatorBroadcastAccounts) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBroadcastAccounts) GetValue(st StateDB) (interface{}, error) {
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

func (opt *operatorBroadcastAccounts) SetValue(st StateDB, value interface{}) error {
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

func (opt *operatorPreBroadcastRoot) KeyHash() common.Hash {
	return opt.key
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

	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "preBroadcastRoot rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorPreBroadcastRoot) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "preBroadcastRoot rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}
