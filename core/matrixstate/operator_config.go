// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixstate

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/rlp"
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

func (opt *operatorVersionInfo) KeyHash() common.Hash {
	return opt.key
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

func (opt *operatorInnerMinerAccounts) KeyHash() common.Hash {
	return opt.key
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

func (opt *operatorFoundationAccount) KeyHash() common.Hash {
	return opt.key
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

func (opt *operatorVersionSuperAccounts) KeyHash() common.Hash {
	return opt.key
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

func (opt *operatorBlockSuperAccounts) KeyHash() common.Hash {
	return opt.key
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
// 多币种签名账户
type operatorMultiCoinSuperAccounts struct {
	key common.Hash
}

func newMultiCoinSuperAccountsOpt() *operatorMultiCoinSuperAccounts {
	return &operatorMultiCoinSuperAccounts{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyAccountMultiCoinSupers),
	}
}

func (opt *operatorMultiCoinSuperAccounts) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorMultiCoinSuperAccounts) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return make([]common.Address, 0), nil
	}
	accounts, err := decodeAccounts(data)
	if err != nil {
		log.Error(logInfo, "MultiCoinSupers decode failed", err)
		return nil, err
	}
	return accounts, nil
}

func (opt *operatorMultiCoinSuperAccounts) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	accounts, OK := value.([]common.Address)
	if !OK {
		log.Error(logInfo, "input param(MultiCoinSupers) err", "reflect failed")
		return ErrParamReflect
	}
	if len(accounts) == 0 {
		log.Error(logInfo, "input param(MultiCoinSupers) err", "accounts is empty")
		return ErrParamReflect
	}
	data, err := encodeAccounts(accounts)
	if err != nil {
		log.Error(logInfo, "MultiCoinSupers encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 子链签名账户
type operatorSubChainSuperAccounts struct {
	key common.Hash
}

func newSubChainSuperAccountsOpt() *operatorSubChainSuperAccounts {
	return &operatorSubChainSuperAccounts{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyAccountSubChainSupers),
	}
}

func (opt *operatorSubChainSuperAccounts) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorSubChainSuperAccounts) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return make([]common.Address, 0), nil
	}
	accounts, err := decodeAccounts(data)
	if err != nil {
		log.Error(logInfo, "SubChainSupers decode failed", err)
		return nil, err
	}
	return accounts, nil
}

func (opt *operatorSubChainSuperAccounts) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	accounts, OK := value.([]common.Address)
	if !OK {
		log.Error(logInfo, "input param(SubChainSupers) err", "reflect failed")
		return ErrParamReflect
	}
	if len(accounts) == 0 {
		log.Error(logInfo, "input param(SubChainSupers) err", "accounts is empty")
		return ErrParamReflect
	}
	data, err := encodeAccounts(accounts)
	if err != nil {
		log.Error(logInfo, "SubChainSupers encode failed", err)
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

func (opt *operatorLeaderConfig) KeyHash() common.Hash {
	return opt.key
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
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "leaderConfig rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorLeaderConfig) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "leaderConfig rlp encode failed", err)
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

func (opt *operatorMinHash) KeyHash() common.Hash {
	return opt.key
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

	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "minHash rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorMinHash) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "minHash rlp encode failed", err)
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

func (opt *operatorSuperBlockCfg) KeyHash() common.Hash {
	return opt.key
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
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "superBlkCfg rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorSuperBlockCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "superBlkCfg rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 最小挖矿难度
type operatorMinDifficulty struct {
	key common.Hash
}

func newMinDiffcultyOpt() *operatorMinDifficulty {
	return &operatorMinDifficulty{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyMinimumDifficulty),
	}
}

func (opt *operatorMinDifficulty) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorMinDifficulty) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "minDifficulty data", "is empty")
		return nil, ErrDataEmpty
	}

	minDifficulty := big.NewInt(0).SetBytes(data)
	return minDifficulty, nil
}

func (opt *operatorMinDifficulty) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	minDifficulty, OK := value.(*big.Int)
	if !OK {
		log.Error(logInfo, "input param(minDifficulty) err", "reflect failed")
		return ErrParamReflect
	}
	data := minDifficulty.Bytes()
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 最大挖矿难度
type operatorMaxDifficulty struct {
	key common.Hash
}

func newMaxDiffcultyOpt() *operatorMaxDifficulty {
	return &operatorMaxDifficulty{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyMaximumDifficulty),
	}
}

func (opt *operatorMaxDifficulty) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorMaxDifficulty) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "MaxDifficulty data", "is empty")
		return nil, ErrDataEmpty
	}

	MaxDifficulty := big.NewInt(0).SetBytes(data)
	return MaxDifficulty, nil
}

func (opt *operatorMaxDifficulty) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	MaxDifficulty, OK := value.(*big.Int)
	if !OK {
		log.Error(logInfo, "input param(MaxDifficulty) err", "reflect failed")
		return ErrParamReflect
	}
	data := MaxDifficulty.Bytes()
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 换届初始挖矿难度
type operatorReelectionDifficulty struct {
	key common.Hash
}

func newReelectionDiffcultyOpt() *operatorReelectionDifficulty {
	return &operatorReelectionDifficulty{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyReelectionDifficulty),
	}
}

func (opt *operatorReelectionDifficulty) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorReelectionDifficulty) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "ReelectionDifficulty data", "is empty")
		return nil, ErrDataEmpty
	}

	ReelectionDifficulty := big.NewInt(0).SetBytes(data)
	return ReelectionDifficulty, nil
}

func (opt *operatorReelectionDifficulty) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	ReelectionDifficulty, OK := value.(*big.Int)
	if !OK {
		log.Error(logInfo, "input param(ReelectionDifficulty) err", "reflect failed")
		return ErrParamReflect
	}
	data := ReelectionDifficulty.Bytes()
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 多币种区块头
type operatorCurrencyHeaderCfg struct {
	key common.Hash
}

func newCurrencyHeaderCfgOpt() *operatorCurrencyHeaderCfg {
	return &operatorCurrencyHeaderCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSCurrencyHeader),
	}
}

func (opt *operatorCurrencyHeaderCfg) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorCurrencyHeaderCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.CurrencyHeader{Roots: []common.CoinRoot{}, Sharding: []common.Coinbyte{}}, nil
	}

	value := new(mc.CurrencyHeader)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "currency header rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorCurrencyHeaderCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "currency header rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 出块时间
type operatorBlockDuration struct {
	key common.Hash
}

func newBlockDurationOpt() *operatorBlockDuration {
	return &operatorBlockDuration{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBlockDurationStatus),
	}
}

func (opt *operatorBlockDuration) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBlockDuration) GetValue(st StateDB) (interface{}, error) {
	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.BlockDurationStatus{Status: []uint8{}}, nil
	}

	value := new(mc.BlockDurationStatus)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "BlockDuration rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBlockDuration) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "BlockDuration rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}
