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
	"math/big"
)

/////////////////////////////////////////////////////////////////////////////////////////
// 区块奖励配置
type operatorBlkRewardCfg struct {
	key common.Hash
}

func newBlkRewardCfgOpt() *operatorBlkRewardCfg {
	return &operatorBlkRewardCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBlkRewardCfg),
	}
}

func (opt *operatorBlkRewardCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "blkRewardCfg data", "is empty")
		return nil, ErrDataEmpty
	}

	value := new(mc.BlkRewardCfg)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "blkRewardCfg unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBlkRewardCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	cfg, OK := value.(*mc.BlkRewardCfg)
	if !OK {
		log.Error(logInfo, "input param(blkRewardCfg) err", "reflect failed")
		return ErrParamReflect
	}
	if cfg == nil {
		log.Error(logInfo, "input param(blkRewardCfg) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Error(logInfo, "blkRewardCfg marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 交易奖励配置
type operatorTxsRewardCfg struct {
	key common.Hash
}

func newTxsRewardCfgOpt() *operatorTxsRewardCfg {
	return &operatorTxsRewardCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyTxsRewardCfg),
	}
}

func (opt *operatorTxsRewardCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "txsRewardCfg data", "is empty")
		return nil, ErrDataEmpty
	}

	value := new(mc.TxsRewardCfgStruct)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "txsRewardCfg unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorTxsRewardCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	cfg, OK := value.(*mc.TxsRewardCfgStruct)
	if !OK {
		log.Error(logInfo, "input param(txsRewardCfg) err", "reflect failed")
		return ErrParamReflect
	}
	if cfg == nil {
		log.Error(logInfo, "input param(txsRewardCfg) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Error(logInfo, "txsRewardCfg marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 利息配置
type operatorInterestCfg struct {
	key common.Hash
}

func newInterestCfgOpt() *operatorInterestCfg {
	return &operatorInterestCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyInterestCfg),
	}
}

func (opt *operatorInterestCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "interestCfg data", "is empty")
		return nil, ErrDataEmpty
	}

	value := new(mc.InterestCfgStruct)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "interestCfg unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorInterestCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	cfg, OK := value.(*mc.InterestCfgStruct)
	if !OK {
		log.Error(logInfo, "input param(interestCfg) err", "reflect failed")
		return ErrParamReflect
	}
	if cfg == nil {
		log.Error(logInfo, "input param(interestCfg) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Error(logInfo, "interestCfg marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 彩票配置
type operatorLotteryCfg struct {
	key common.Hash
}

func newLotteryCfgOpt() *operatorLotteryCfg {
	return &operatorLotteryCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyLotteryCfg),
	}
}

func (opt *operatorLotteryCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "lotteryCfg data", "is empty")
		return nil, ErrDataEmpty
	}

	value := new(mc.LotteryCfgStruct)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "lotteryCfg unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorLotteryCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	cfg, OK := value.(*mc.LotteryCfgStruct)
	if !OK {
		log.Error(logInfo, "input param(lotteryCfg) err", "reflect failed")
		return ErrParamReflect
	}
	if cfg == nil {
		log.Error(logInfo, "input param(lotteryCfg) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Error(logInfo, "lotteryCfg marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 惩罚配置
type operatorSlashCfg struct {
	key common.Hash
}

func newSlashCfgOpt() *operatorSlashCfg {
	return &operatorSlashCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeySlashCfg),
	}
}

func (opt *operatorSlashCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "slashCfg data", "is empty")
		return nil, ErrDataEmpty
	}

	value := new(mc.SlashCfgStruct)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "slashCfg unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorSlashCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	cfg, OK := value.(*mc.SlashCfgStruct)
	if !OK {
		log.Error(logInfo, "input param(slashCfg) err", "reflect failed")
		return ErrParamReflect
	}
	if cfg == nil {
		log.Error(logInfo, "input param(slashCfg) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Error(logInfo, "slashCfg marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 上一矿工区块奖励金额
type operatorPreMinerBlkReward struct {
	key common.Hash
}

func newPreMinerBlkRewardOpt() *operatorPreMinerBlkReward {
	return &operatorPreMinerBlkReward{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyPreMinerBlkReward),
	}
}

func (opt *operatorPreMinerBlkReward) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.MinerOutReward{Reward: *big.NewInt(0)}, nil
	}

	value := new(mc.MinerOutReward)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "preMinerBlkReward unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorPreMinerBlkReward) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	reward, OK := value.(*mc.MinerOutReward)
	if !OK {
		log.Error(logInfo, "input param(preMinerBlkReward) err", "reflect failed")
		return ErrParamReflect
	}
	if reward == nil {
		log.Error(logInfo, "input param(preMinerBlkReward) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(reward)
	if err != nil {
		log.Error(logInfo, "preMinerBlkReward marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 上一矿工交易奖励金额
type operatorPreMinerTxsReward struct {
	key common.Hash
}

func newPreMinerTxsRewardOpt() *operatorPreMinerTxsReward {
	return &operatorPreMinerTxsReward{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyPreMinerTxsReward),
	}
}

func (opt *operatorPreMinerTxsReward) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.MinerOutReward{Reward: *big.NewInt(0)}, nil
	}

	value := new(mc.MinerOutReward)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "preMinerTxsReward unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorPreMinerTxsReward) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	reward, OK := value.(*mc.MinerOutReward)
	if !OK {
		log.Error(logInfo, "input param(preMinerTxsReward) err", "reflect failed")
		return ErrParamReflect
	}
	if reward == nil {
		log.Error(logInfo, "input param(preMinerTxsReward) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(reward)
	if err != nil {
		log.Error(logInfo, "preMinerTxsReward marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// upTime状态
type operatorUpTimeNum struct {
	key common.Hash
}

func newUpTimeNumOpt() *operatorUpTimeNum {
	return &operatorUpTimeNum{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyUpTimeNum),
	}
}

func (opt *operatorUpTimeNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return 0, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return 0, nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "upTimeNum decode failed", err)
		return 0, err
	}
	return num, nil
}

func (opt *operatorUpTimeNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	num, OK := value.(uint64)
	if !OK {
		log.Error(logInfo, "input param(upTimeNum) err", "reflect failed")
		return ErrParamReflect
	}
	st.SetMatrixData(opt.key, encodeUint64(num))
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 彩票状态
type operatorLotteryNum struct {
	key common.Hash
}

func newLotteryNumOpt() *operatorLotteryNum {
	return &operatorLotteryNum{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyLotteryNum),
	}
}

func (opt *operatorLotteryNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return 0, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return 0, nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "lotteryNum decode failed", err)
		return 0, err
	}
	return num, nil
}

func (opt *operatorLotteryNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	num, OK := value.(uint64)
	if !OK {
		log.Error(logInfo, "input param(lotteryNum) err", "reflect failed")
		return ErrParamReflect
	}
	st.SetMatrixData(opt.key, encodeUint64(num))
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 彩票候选账户
type operatorLotteryAccount struct {
	key common.Hash
}

func newLotteryAccountOpt() *operatorLotteryAccount {
	return &operatorLotteryAccount{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyLotteryAccount),
	}
}

func (opt *operatorLotteryAccount) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.LotteryFrom{From: make([]common.Address, 0)}, nil
	}

	value := new(mc.LotteryFrom)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "lotteryAccount unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorLotteryAccount) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	accounts, OK := value.(*mc.LotteryFrom)
	if !OK {
		log.Error(logInfo, "input param(lotteryAccount) err", "reflect failed")
		return ErrParamReflect
	}
	if accounts == nil {
		log.Error(logInfo, "input param(lotteryAccount) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(accounts)
	if err != nil {
		log.Error(logInfo, "lotteryAccount marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 利息计算状态
type operatorInterestCalcNum struct {
	key common.Hash
}

func newInterestCalcNumOpt() *operatorInterestCalcNum {
	return &operatorInterestCalcNum{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyInterestCalcNum),
	}
}

func (opt *operatorInterestCalcNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return 0, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return 0, nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "interestCalcNum decode failed", err)
		return 0, err
	}
	return num, nil
}

func (opt *operatorInterestCalcNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	num, OK := value.(uint64)
	if !OK {
		log.Error(logInfo, "input param(interestCalcNum) err", "reflect failed")
		return ErrParamReflect
	}
	st.SetMatrixData(opt.key, encodeUint64(num))
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 利息支付状态
type operatorInterestPayNum struct {
	key common.Hash
}

func newInterestPayNumOpt() *operatorInterestPayNum {
	return &operatorInterestPayNum{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyInterestPayNum),
	}
}

func (opt *operatorInterestPayNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return 0, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return 0, nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "interestPayNum decode failed", err)
		return 0, err
	}
	return num, nil
}

func (opt *operatorInterestPayNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	num, OK := value.(uint64)
	if !OK {
		log.Error(logInfo, "input param(interestPayNum) err", "reflect failed")
		return ErrParamReflect
	}
	st.SetMatrixData(opt.key, encodeUint64(num))
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 惩罚状态
type operatorSlashNum struct {
	key common.Hash
}

func newSlashNumOpt() *operatorSlashNum {
	return &operatorSlashNum{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeySlashNum),
	}
}

func (opt *operatorSlashNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return 0, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return 0, nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "slashNum decode failed", err)
		return 0, err
	}
	return num, nil
}

func (opt *operatorSlashNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	num, OK := value.(uint64)
	if !OK {
		log.Error(logInfo, "input param(slashNum) err", "reflect failed")
		return ErrParamReflect
	}
	st.SetMatrixData(opt.key, encodeUint64(num))
	return nil
}
