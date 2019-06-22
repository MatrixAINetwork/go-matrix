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
	"reflect"
)

/////////////////////////////////////////////////////////////////////////////////////////
// 拓扑图
type operatorTopologyGraph struct {
	key common.Hash
}

func newTopologyGraphOpt() *operatorTopologyGraph {
	return &operatorTopologyGraph{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyTopologyGraph),
	}
}

func (opt *operatorTopologyGraph) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorTopologyGraph) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "topologyGraph data", "is empty")
		return nil, ErrDataEmpty
	}
	value := new(mc.TopologyGraph)
	if err := rlp.DecodeBytes(data, &value); err != nil {
		log.Error(logInfo, "topologyGraph rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorTopologyGraph) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "topologyGraph rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 选举图
type operatorElectGraph struct {
	key common.Hash
}

func newELectGraphOpt() *operatorElectGraph {
	return &operatorElectGraph{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyElectGraph),
	}
}

func (opt *operatorElectGraph) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorElectGraph) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "electGraph data", "is empty")
		return nil, ErrDataEmpty
	}
	value := new(mc.ElectGraph)
	if err := rlp.DecodeBytes(data, &value); err != nil {
		log.Error(logInfo, "electGraph rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorElectGraph) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "electGraph rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 选举节点在线信息
type operatorElectOnlineState struct {
	key common.Hash
}

func newELectOnlineStateOpt() *operatorElectOnlineState {
	return &operatorElectOnlineState{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyElectOnlineState),
	}
}

func (opt *operatorElectOnlineState) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorElectOnlineState) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "electOnlineStatus data", "is empty")
		return nil, ErrDataEmpty
	}
	value := new(mc.ElectOnlineStatus)
	if err := rlp.DecodeBytes(data, &value); err != nil {
		log.Error(logInfo, "electOnlineStatus rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorElectOnlineState) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}
	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "electOnlineStatus rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 选举生成时间
type operatorElectGenTime struct {
	key common.Hash
}

func newElectGenTimeOpt() *operatorElectGenTime {
	return &operatorElectGenTime{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyElectGenTime),
	}
}

func (opt *operatorElectGenTime) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorElectGenTime) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "electGenTime data", "is empty")
		return nil, ErrDataEmpty
	}
	value := new(mc.ElectGenTimeStruct)
	if err := rlp.DecodeBytes(data, &value); err != nil {
		log.Error(logInfo, "electGenTime rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorElectGenTime) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "electGenTime rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 选举矿工数量
type operatorElectMinerNum struct {
	key common.Hash
}

func newElectMinerNumOpt() *operatorElectMinerNum {
	return &operatorElectMinerNum{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyElectMinerNum),
	}
}

func (opt *operatorElectMinerNum) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorElectMinerNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "electMinerNum data", "is empty")
		return nil, ErrDataEmpty
	}
	value := new(mc.ElectMinerNumStruct)
	if err := rlp.DecodeBytes(data, &value); err != nil {
		log.Error(logInfo, "electMinerNum rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorElectMinerNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "electMinerNum rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 选举配置
type operatorElectConfigInfo struct {
	key common.Hash
}

func newElectConfigInfoOpt() *operatorElectConfigInfo {
	return &operatorElectConfigInfo{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyElectConfigInfo),
	}
}

func (opt *operatorElectConfigInfo) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorElectConfigInfo) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "electConfigInfo data", "is empty")
		return nil, ErrDataEmpty
	}
	value := new(mc.ElectConfigInfo)
	if err := rlp.DecodeBytes(data, &value); err != nil {
		log.Error(logInfo, "electConfigInfo rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorElectConfigInfo) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	info, OK := value.(*mc.ElectConfigInfo)
	if !OK {
		log.Error(logInfo, "input param(electConfigInfo) err", "reflect failed")
		return ErrParamReflect
	}
	if info == nil {
		log.Error(logInfo, "input param(electConfigInfo) err", "is nil")
		return ErrParamNil
	}
	data, err := rlp.EncodeToBytes(info)
	if err != nil {
		log.Error(logInfo, "electConfigInfo rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 选举黑名单
type operatorElectBlackList struct {
	key common.Hash
}

func newElectBlackListOpt() *operatorElectBlackList {
	return &operatorElectBlackList{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyElectBlackList),
	}
}

func (opt *operatorElectBlackList) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorElectBlackList) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return make([]common.Address, 0), nil
	}
	accounts, err := decodeAccounts(data)
	if err != nil {
		log.Error(logInfo, "electBlackList decode failed", err)
		return nil, err
	}
	return accounts, nil
}

func (opt *operatorElectBlackList) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}
	v1 := reflect.ValueOf(value)
	if v1.Kind() == reflect.Slice && v1.Len() == 0 {
		nilSlice := make([]byte, 0)
		st.SetMatrixData(opt.key, nilSlice)
		return nil
	}
	accounts, OK := value.([]common.Address)
	if !OK {
		log.Error(logInfo, "input param(electBlackList) err", "reflect failed")
		return ErrParamReflect
	}
	data, err := encodeAccounts(accounts)
	if err != nil {
		log.Error(logInfo, "electBlackList encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 选举白名单
type operatorElectWhiteList struct {
	key common.Hash
}

func newElectWhiteListOpt() *operatorElectWhiteList {
	return &operatorElectWhiteList{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyElectWhiteList),
	}
}

func (opt *operatorElectWhiteList) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorElectWhiteList) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return make([]common.Address, 0), nil
	}
	accounts, err := decodeAccounts(data)
	if err != nil {
		log.Error(logInfo, "electWhiteList decode failed", err)
		return nil, err
	}
	return accounts, nil
}

func (opt *operatorElectWhiteList) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}
	v1 := reflect.ValueOf(value)
	if v1.Kind() == reflect.Slice && v1.Len() == 0 {
		nilSlice := make([]byte, 0)
		st.SetMatrixData(opt.key, nilSlice)
		return nil
	}
	accounts, OK := value.([]common.Address)
	if !OK {
		log.Error(logInfo, "input param(electWhiteList) err", "reflect failed")
		return ErrParamReflect
	}
	data, err := encodeAccounts(accounts)
	if err != nil {
		log.Error(logInfo, "electWhiteList encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 选举白名单开关
type operatorElectWhiteListSwitcher struct {
	key common.Hash
}

func newElectWhiteListSwitcherOpt() *operatorElectWhiteListSwitcher {
	return &operatorElectWhiteListSwitcher{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyElectWhiteListSwitcher),
	}
}

func (opt *operatorElectWhiteListSwitcher) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorElectWhiteListSwitcher) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.ElectWhiteListSwitcher{Switcher: false}, ErrDataEmpty
	}
	value := new(mc.ElectWhiteListSwitcher)
	if err := rlp.DecodeBytes(data, &value); err != nil {
		log.Error(logInfo, "electWhiteListSwitcher rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorElectWhiteListSwitcher) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}
	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "electWhiteListSwitcher rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// VIP配置信息
type operatorVIPConfig struct {
	key common.Hash
}

func newVIPConfigOpt() *operatorVIPConfig {
	return &operatorVIPConfig{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyVIPConfig),
	}
}

func (opt *operatorVIPConfig) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorVIPConfig) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	value := make([]mc.VIPConfig, 0)
	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return value, nil
	}

	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "electGraph rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorVIPConfig) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	config, OK := value.([]mc.VIPConfig)
	if !OK {
		log.Error(logInfo, "input param(vipConfig) err", "reflect failed")
		return ErrParamReflect
	}
	data, err := rlp.EncodeToBytes(config)
	if err != nil {
		log.Error(logInfo, "vipConfig rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}
