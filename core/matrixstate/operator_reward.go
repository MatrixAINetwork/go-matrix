// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixstate

import (
	"math/big"

	"encoding/json"
	"reflect"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"github.com/pkg/errors"
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

func (opt *operatorBlkRewardCfg) KeyHash() common.Hash {
	return opt.key
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
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "blkRewardCfg rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBlkRewardCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "blkRewardCfg rlp encode failed", err)
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

func (opt *operatorTxsRewardCfg) KeyHash() common.Hash {
	return opt.key
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

	value := new(mc.TxsRewardCfg)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "txsRewardCfg rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorTxsRewardCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "txsRewardCfg rlp encode failed", err)
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

func (opt *operatorInterestCfg) KeyHash() common.Hash {
	return opt.key
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

	value := new(mc.InterestCfg)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "interestCfg rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorInterestCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "interestCfg rlp encode failed", err)
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

func (opt *operatorLotteryCfg) KeyHash() common.Hash {
	return opt.key
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

	value := new(mc.LotteryCfg)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "lotteryCfg rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorLotteryCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "lotteryCfg rlp encode failed", err)
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

func (opt *operatorSlashCfg) KeyHash() common.Hash {
	return opt.key
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

	value := new(mc.SlashCfg)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "slashCfg rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorSlashCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "slashCfg rlp encode failed", err)
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

func (opt *operatorPreMinerBlkReward) KeyHash() common.Hash {
	return opt.key
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
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "preMinerBlkReward rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorPreMinerBlkReward) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "preMinerBlkReward rlp encode failed", err)
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

func (opt *operatorPreMinerTxsReward) KeyHash() common.Hash {
	return opt.key
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
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "preMinerTxsReward rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorPreMinerTxsReward) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, err := rlp.EncodeToBytes(value)
	if err != nil {
		log.Error(logInfo, "preMinerTxsReward rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 上一矿工交易奖励金额
type operatorPreMinerMultiCoinTxsReward struct {
	key common.Hash
}

func newPreMinerMultiCoinTxsRewardOpt() *operatorPreMinerMultiCoinTxsReward {
	return &operatorPreMinerMultiCoinTxsReward{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyPreMinerTxsReward),
	}
}

func (opt *operatorPreMinerMultiCoinTxsReward) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorPreMinerMultiCoinTxsReward) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return make([]mc.MultiCoinMinerOutReward, 0), nil
	}

	value := make([]mc.MultiCoinMinerOutReward, 0)
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "operatorPreMinerMultiCoinTxsReward rlp decode failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorPreMinerMultiCoinTxsReward) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}
	preMiner, OK := value.([]mc.MultiCoinMinerOutReward)
	if !OK {
		log.Error(logInfo, "input param(MultiCoinMinerOutReward) err", "reflect failed")
		return ErrParamReflect
	}
	data, err := rlp.EncodeToBytes(preMiner)
	if err != nil {
		log.Error(logInfo, "operatorPreMinerMultiCoinTxsReward rlp encode failed", err)
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

func (opt *operatorUpTimeNum) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorUpTimeNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return uint64(0), nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "upTimeNum decode failed", err)
		return uint64(0), err
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

func (opt *operatorLotteryNum) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorLotteryNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return uint64(0), nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "lotteryNum decode failed", err)
		return uint64(0), err
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

func (opt *operatorLotteryAccount) KeyHash() common.Hash {
	return opt.key
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
	err := rlp.DecodeBytes(data, &value)
	if err != nil {
		log.Error(logInfo, "lotteryAccount rlp decode failed", err)
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
	data, err := rlp.EncodeToBytes(accounts)
	if err != nil {
		log.Error(logInfo, "lotteryAccount rlp encode failed", err)
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

func (opt *operatorInterestCalcNum) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorInterestCalcNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return uint64(0), nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "interestCalcNum decode failed", err)
		return uint64(0), err
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

func (opt *operatorInterestPayNum) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorInterestPayNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return uint64(0), nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "interestPayNum decode failed", err)
		return uint64(0), err
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

func (opt *operatorSlashNum) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorSlashNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return uint64(0), nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "slashNum decode failed", err)
		return uint64(0), err
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

/////////////////////////////////////////////////////////////////////////////////////////
// 固定区块算法配置
type operatorBlkCalc struct {
	key common.Hash
}

func newBlkCalcOpt() *operatorBlkCalc {
	return &operatorBlkCalc{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBlkCalc),
	}
}

func (opt *operatorBlkCalc) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBlkCalc) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}
	calc, err := decodeString(data)
	if err != nil {
		log.Error(logInfo, "BlkCalc decode failed", err)
		return nil, err
	}
	return calc, nil
}

func (opt *operatorBlkCalc) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(BlkCalc) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := encodeString(data)
	if err != nil {
		log.Error(logInfo, "BlkCalc encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 交易费算法配置
type operatorTxsCalc struct {
	key common.Hash
}

func newTxsCalcOpt() *operatorTxsCalc {
	return &operatorTxsCalc{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyTxsCalc),
	}
}

func (opt *operatorTxsCalc) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorTxsCalc) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}
	calc, err := decodeString(data)
	if err != nil {
		log.Error(logInfo, "TxsCalc decode failed", err)
		return nil, err
	}
	return calc, nil
}

func (opt *operatorTxsCalc) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(TxsCalc) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := encodeString(data)
	if err != nil {
		log.Error(logInfo, "TxsCalc encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 利息算法配置
type operatorInterestCalc struct {
	key common.Hash
}

func newInterestCalcOpt() *operatorInterestCalc {
	return &operatorInterestCalc{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyInterestCalc),
	}
}

func (opt *operatorInterestCalc) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorInterestCalc) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}
	calc, err := decodeString(data)
	if err != nil {
		log.Error(logInfo, "InterestCalc decode failed", err)
		return nil, err
	}
	return calc, nil
}

func (opt *operatorInterestCalc) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(InterestCalc) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := encodeString(data)
	if err != nil {
		log.Error(logInfo, "InterestCalc encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 彩票算法配置
type operatorLotteryCalc struct {
	key common.Hash
}

func newLotteryCalcOpt() *operatorLotteryCalc {
	return &operatorLotteryCalc{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyLotteryCalc),
	}
}

func (opt *operatorLotteryCalc) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorLotteryCalc) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}
	calc, err := decodeString(data)
	if err != nil {
		log.Error(logInfo, "LotteryCalc decode failed", err)
		return nil, err
	}
	return calc, nil
}

func (opt *operatorLotteryCalc) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(LotteryCalc) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := encodeString(data)
	if err != nil {
		log.Error(logInfo, "LotteryCalc encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 惩罚算法配置
type operatorSlashCalc struct {
	key common.Hash
}

func newSlashCalcOpt() *operatorSlashCalc {
	return &operatorSlashCalc{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeySlashCalc),
	}
}

func (opt *operatorSlashCalc) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorSlashCalc) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}
	calc, err := decodeString(data)
	if err != nil {
		log.Error(logInfo, "SlashCalc decode failed", err)
		return nil, err
	}
	return calc, nil
}

func (opt *operatorSlashCalc) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(SlashCalc) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := encodeString(data)
	if err != nil {
		log.Error(logInfo, "SlashCalc encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//入池gas门限
type operatorTxpoolGasLimit struct {
	key common.Hash
}

func newTxpoolGasLimitOpt() *operatorTxpoolGasLimit {
	return &operatorTxpoolGasLimit{
		key: types.RlpHash(matrixStatePrefix + mc.MSTxpoolGasLimitCfg),
	}
}

func (opt *operatorTxpoolGasLimit) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorTxpoolGasLimit) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return big.NewInt(int64(params.TxGasPrice)), nil
	}

	msg := new(big.Int)
	err := rlp.DecodeBytes(data, &msg)
	if err != nil {
		return nil, errors.Errorf("json.rlp decode failed: %s", err)
	}

	return msg, nil
}

func (opt *operatorTxpoolGasLimit) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(*big.Int)
	if !OK {
		log.Error(logInfo, "input param(TxpoolGasLimit) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := rlp.EncodeToBytes(data)
	if err != nil {
		log.Error(logInfo, "TxpoolGasLimit encode failed", err)
		return err
	}

	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//币种打包限制
type operatorCurrencyConfig struct {
	key common.Hash
}

func newCurrencyPackOpt() *operatorCurrencyConfig {
	return &operatorCurrencyConfig{
		key: types.RlpHash(matrixStatePrefix + mc.MSCurrencyConfig),
	}
}

func (opt *operatorCurrencyConfig) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorCurrencyConfig) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return make([]common.CoinConfig, 0), nil
	}
	currencylist := make([]common.CoinConfig, 0)
	//err := rlp.DecodeBytes(data, &currencylist)
	err := json.Unmarshal(data, &currencylist)
	if err != nil {
		return nil, errors.Errorf("operatorCurrencyPack rlp decode  failed: %s", err)
	}

	return currencylist, nil
}

func (opt *operatorCurrencyConfig) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}
	//取消
	v1 := reflect.ValueOf(value)
	if v1.Kind() == reflect.Slice && v1.Len() == 0 {
		nilSlice := make([]byte, 0)
		st.SetMatrixData(opt.key, nilSlice)
		return nil
	}
	data, OK := value.([]common.CoinConfig)
	if !OK {
		log.Error(logInfo, "input param(CurrencyPack) err", "reflect failed")
		return ErrParamReflect
	}

	//encodeData, err := rlp.EncodeToBytes(data)
	encodeData, err := json.Marshal(data)
	if err != nil {
		log.Error(logInfo, "operatorCurrencyPack rlp encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 账户黑名单
type operatorAccountBlackList struct {
	key common.Hash
}

func newAccountBlackListOpt() *operatorAccountBlackList {
	return &operatorAccountBlackList{
		key: types.RlpHash(matrixStatePrefix + mc.MSAccountBlackList),
	}
}

func (opt *operatorAccountBlackList) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorAccountBlackList) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return make([]common.Address, 0), nil
	}
	accounts, err := decodeAccounts(data)
	if err != nil {
		log.Error(logInfo, "AccountBlackList decode failed", err)
		return nil, err
	}
	return accounts, nil
}

func IsInBlackList(addr common.Address, blacklist []common.Address) bool {
	for _, blackaddr := range blacklist {
		if addr.Equal(blackaddr) {
			return true
		}
	}
	return false
}
func (opt *operatorAccountBlackList) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}
	//取消
	v1 := reflect.ValueOf(value)
	if v1.Kind() == reflect.Slice && v1.Len() == 0 {
		nilSlice := make([]byte, 0)
		st.SetMatrixData(opt.key, nilSlice)
		return nil
	}

	accounts, OK := value.([]common.Address)
	if !OK {
		log.Error(logInfo, "input param(AccountBlackList) err", "reflect failed")
		return ErrParamReflect
	}

	data, err := encodeAccounts(accounts)
	if err != nil {
		log.Error(logInfo, "AccountBlackList encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}
