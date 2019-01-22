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
// 拓扑图
type operatorTopologyGraph struct {
	key common.Hash
}

func newTopologyGraphOpt() *operatorTopologyGraph {
	return &operatorTopologyGraph{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyTopologyGraph),
	}
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
	if err := json.Unmarshal(data, &value); err != nil {
		log.Error(logInfo, "topologyGraph unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorTopologyGraph) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	graph, OK := value.(*mc.TopologyGraph)
	if !OK {
		log.Error(logInfo, "input param(topologyGraph) err", "reflect failed")
		return ErrParamReflect
	}
	if graph == nil {
		log.Error(logInfo, "input param(topologyGraph) err", "is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(graph)
	if err != nil {
		log.Error(logInfo, "topologyGraph marshal failed", err)
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
	if err := json.Unmarshal(data, &value); err != nil {
		log.Error(logInfo, "electGraph unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorElectGraph) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	graph, OK := value.(*mc.ElectGraph)
	if !OK {
		log.Error(logInfo, "input param(electGraph) err", "reflect failed")
		return ErrParamReflect
	}
	if graph == nil {
		log.Error(logInfo, "input param(electGraph) err", "is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(graph)
	if err != nil {
		log.Error(logInfo, "electGraph marshal failed", err)
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
	if err := json.Unmarshal(data, &value); err != nil {
		log.Error(logInfo, "electOnlineStatus unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorElectOnlineState) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	status, OK := value.(*mc.ElectOnlineStatus)
	if !OK {
		log.Error(logInfo, "input param(electOnlineStatus) err", "reflect failed")
		return ErrParamReflect
	}
	if status == nil {
		log.Error(logInfo, "input param(electOnlineStatus) err", "is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(status)
	if err != nil {
		log.Error(logInfo, "electOnlineStatus marshal failed", err)
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
	if err := json.Unmarshal(data, &value); err != nil {
		log.Error(logInfo, "electGenTime unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorElectGenTime) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	genTime, OK := value.(*mc.ElectGenTimeStruct)
	if !OK {
		log.Error(logInfo, "input param(electGenTime) err", "reflect failed")
		return ErrParamReflect
	}
	if genTime == nil {
		log.Error(logInfo, "input param(electGenTime) err", "is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(genTime)
	if err != nil {
		log.Error(logInfo, "electGenTime marshal failed", err)
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
	if err := json.Unmarshal(data, &value); err != nil {
		log.Error(logInfo, "electMinerNum unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorElectMinerNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	info, OK := value.(*mc.ElectMinerNumStruct)
	if !OK {
		log.Error(logInfo, "input param(electMinerNum) err", "reflect failed")
		return ErrParamReflect
	}
	if info == nil {
		log.Error(logInfo, "input param(electMinerNum) err", "is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(info)
	if err != nil {
		log.Error(logInfo, "electMinerNum marshal failed", err)
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
	if err := json.Unmarshal(data, &value); err != nil {
		log.Error(logInfo, "electConfigInfo unmarshal failed", err)
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
	data, err := json.Marshal(info)
	if err != nil {
		log.Error(logInfo, "electConfigInfo marshal failed", err)
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
// VIP配置信息
type operatorVIPConfig struct {
	key common.Hash
}

func newVIPConfigOpt() *operatorVIPConfig {
	return &operatorVIPConfig{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyVIPConfig),
	}
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

	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "electGraph unmarshal failed", err)
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
	data, err := json.Marshal(config)
	if err != nil {
		log.Error(logInfo, "vipConfig marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}
