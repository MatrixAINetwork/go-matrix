// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package depoistInfo

import (
	"context"
	"math/big"

	"reflect"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rpc"
	"github.com/pkg/errors"
)

type manBackend interface {
	StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDBManage, *types.Header, error)
	StateAndHeaderByHash(ctx context.Context, hash common.Hash) (*state.StateDBManage, *types.Header, error)
}

func (v1 *DepositMnger_v1) AddOnlineTime(stateDB vm.StateDBManager, address common.Address, ot *big.Int) error {
	return depositInfo.MatrixDeposit.AddOnlineTime(depositInfo.Contract, stateDB, address, ot)
}

func (v1 *DepositMnger_v1) GetOnlineTime(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	return depositInfo.MatrixDeposit.GetOnlineTime(depositInfo.Contract, stateDB, address), nil
}

func (v1 *DepositMnger_v1) SetOnlineTime(stateDB vm.StateDBManager, address common.Address, ot *big.Int) error {
	return depositInfo.MatrixDeposit.SetOnlineTime(depositInfo.Contract, stateDB, address, ot)
}

func (v1 *DepositMnger_v1) GetDepositList(tm *big.Int, getDeposit common.RoleType, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfo(tm)
	//if err != nil {
	//	return nil, err
	//}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	//var depositList []vm.DepositDetail
	depositList := make([]vm.DepositDetail, 0)
	if common.RoleValidator == common.RoleValidator&getDeposit {

		depositList = append(depositList, depositInfo.MatrixDeposit.GetValidatorDepositList(contract, statedb)...)
	}

	if common.RoleMiner == common.RoleMiner&getDeposit {
		depositList = append(depositList, depositInfo.MatrixDeposit.GetMinerDepositList(contract, statedb)...)
	}
	return depositList, nil
}

func (v1 *DepositMnger_v1) GetDepositAndWithDrawList(tm *big.Int, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfo(tm)
	//if err != nil {
	//	return nil, err
	//}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	var depositList []vm.DepositDetail
	depositList = depositInfo.MatrixDeposit.GetAllDepositList(contract, statedb, true, headtime)
	return depositList, nil
}

func (v1 *DepositMnger_v1) GetAllDeposit(tm *big.Int, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfo(tm)
	//if err != nil {
	//	return nil, err
	//}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	var depositList []vm.DepositDetail
	depositList = depositInfo.MatrixDeposit.GetAllDepositList(contract, statedb, false, headtime)
	return depositList, nil
}

func (v1 *DepositMnger_v1) GetDepositListByHash(hash common.Hash, getDeposit common.RoleType, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfoByHash(hash)
	//if err != nil {
	//	return nil, err
	//}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	//var depositList []vm.DepositDetail
	depositList := make([]vm.DepositDetail, 0)
	if common.RoleValidator == common.RoleValidator&getDeposit {

		depositList = append(depositList, depositInfo.MatrixDeposit.GetValidatorDepositList(contract, statedb)...)
	}

	if common.RoleMiner == common.RoleMiner&getDeposit {
		depositList = append(depositList, depositInfo.MatrixDeposit.GetMinerDepositList(contract, statedb)...)
	}
	return depositList, nil
}

func (v1 *DepositMnger_v1) GetDepositAndWithDrawListByHash(hash common.Hash, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfoByHash(hash)
	//if err != nil {
	//	return nil, err
	//}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	var depositList []vm.DepositDetail
	depositList = depositInfo.MatrixDeposit.GetAllDepositList(contract, statedb, true, headtime)
	return depositList, nil
}

func (v1 *DepositMnger_v1) GetAllDepositByHash(hash common.Hash, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfoByHash(hash)
	//if err != nil {
	//	return nil, err
	//}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	var depositList []vm.DepositDetail
	depositList = depositInfo.MatrixDeposit.GetAllDepositList(contract, statedb, false, headtime)
	return depositList, nil
}

func getDepositInfo(tm *big.Int) (db vm.StateDBManager, headtime uint64, err error) {
	depositInfo.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	var c context.Context
	var h rpc.BlockNumber
	h = rpc.BlockNumber(tm.Int64())
	//encode := hexutil.EncodeBig(tm)
	//err = h.UnmarshalJSON([]byte(encode))
	//if err != nil {
	//	return nil, err
	//}
	db, header, err := depositInfo.manApi.StateAndHeaderByNumber(c, h)
	if err != nil {
		return nil, uint64(0), err
	}

	if db == nil {
		return nil, uint64(0), errors.New("db is nil")
	}
	dbValue := reflect.ValueOf(db)
	if dbValue.Kind() == reflect.Ptr && dbValue.IsNil() {
		return nil, uint64(0), errors.New("db is nil")
	}

	return db, header.Time.Uint64(), nil
}

func getDepositInfoByHash(hash common.Hash) (db vm.StateDBManager, headtime uint64, err error) {
	depositInfo.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	var c context.Context
	db, head, err := depositInfo.manApi.StateAndHeaderByHash(c, hash)
	if err != nil {
		return nil, uint64(0), err
	}

	if db == nil {
		return nil, 0, errors.New("db is nil")
	}
	dbValue := reflect.ValueOf(db)
	if dbValue.Kind() == reflect.Ptr && dbValue.IsNil() {
		return nil, 0, errors.New("db is nil")
	}

	return db, head.Time.Uint64(), nil
}

func (v1 *DepositMnger_v1) ResetSlash(stateDB vm.StateDBManager, address common.Address) error {
	return depositInfo.MatrixDeposit.ResetSlash(depositInfo.Contract, stateDB, address)
}

func (v1 *DepositMnger_v1) GetSlash(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	return depositInfo.MatrixDeposit.GetSlash(depositInfo.Contract, stateDB, address), nil
}

func (v1 *DepositMnger_v1) GetAllSlash(stateDB vm.StateDBManager) map[common.Address]*big.Int {
	return depositInfo.MatrixDeposit.GetAllSlash(depositInfo.Contract, stateDB)
}

func (v1 *DepositMnger_v1) AddSlash(stateDB vm.StateDBManager, address common.Address, slash *big.Int) error {
	return depositInfo.MatrixDeposit.AddSlash(depositInfo.Contract, stateDB, address, slash)
}

func (v1 *DepositMnger_v1) SetSlash(stateDB vm.StateDBManager, address common.Address, slash *big.Int) error {
	return depositInfo.MatrixDeposit.SetSlash(depositInfo.Contract, stateDB, address, slash)
}

func (v1 *DepositMnger_v1) ResetInterest(stateDB vm.StateDBManager, address common.Address) error {
	return depositInfo.MatrixDeposit.ResetInterest(depositInfo.Contract, stateDB, address)
}

func (v1 *DepositMnger_v1) GetInterest(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	return depositInfo.MatrixDeposit.GetInterest(depositInfo.Contract, stateDB, address), nil
}

func (v1 *DepositMnger_v1) GetAllInterest(stateDB vm.StateDBManager) map[common.Address]*big.Int {
	return depositInfo.MatrixDeposit.GetAllInterest(depositInfo.Contract, stateDB)
}

func (v1 *DepositMnger_v1) AddInterest(stateDB vm.StateDBManager, address common.Address, reward *big.Int) error {
	return depositInfo.MatrixDeposit.AddInterest(depositInfo.Contract, stateDB, address, reward)
}

func (v1 *DepositMnger_v1) SetInterest(stateDB vm.StateDBManager, address common.Address, reward *big.Int) error {
	return depositInfo.MatrixDeposit.SetInterest(depositInfo.Contract, stateDB, address, reward)
}
func (v1 *DepositMnger_v1) GetDeposit(stateDB vm.StateDBManager, address common.Address) *big.Int {
	return depositInfo.MatrixDeposit.GetDeposit(depositInfo.Contract, stateDB, address)
}

func (v1 *DepositMnger_v1) SetDeposit(stateDB vm.StateDBManager, address common.Address) error {
	return depositInfo.MatrixDeposit.SetDeposit(depositInfo.Contract, stateDB, address)
}
func (v1 *DepositMnger_v1) AddDeposit(stateDB vm.StateDBManager, address common.Address) error {
	return depositInfo.MatrixDeposit.AddDeposit(depositInfo.Contract, stateDB, address)
}

// 获取A0账户
func (v1 *DepositMnger_v1) GetDepositAccount(stateDB vm.StateDBManager, authAccount common.Address) common.Address {
	if depositInfo.Contract == nil {
		depositInfo.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	}
	return depositInfo.MatrixDeposit.GetDepositAccount(depositInfo.Contract, stateDB, authAccount)
}

// 获取A1账户
func (v1 *DepositMnger_v1) GetAuthAccount(stateDB vm.StateDBManager, depositAccount common.Address) common.Address {
	if depositInfo.Contract == nil {
		depositInfo.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	}
	return depositInfo.MatrixDeposit.GetAuthAccount(depositInfo.Contract, stateDB, depositAccount)
}
func (v1 *DepositMnger_v1) AddSlash_v2(stateDB vm.StateDBManager, address common.Address, slash common.CalculateDeposit) error {
	return nil
}
func (v1 *DepositMnger_v1) GetSlash_v2(stateDB vm.StateDBManager, address common.Address) (common.CalculateDeposit, error) {
	ret, err := v1.GetSlash(stateDB, address)
	if err != nil {
		return common.CalculateDeposit{}, err
	}
	var numlist common.CalculateDeposit
	numlist.AddressA0 = address
	numlist.CalcDeposit = append(numlist.CalcDeposit, common.OperationalInterestSlash{OperAmount: ret})
	return numlist, nil
}
func (v1 *DepositMnger_v1) GetInterest_v2(stateDB vm.StateDBManager, address common.Address) (common.CalculateDeposit, error) {
	ret, err := v1.GetInterest(stateDB, address)
	if err != nil {
		return common.CalculateDeposit{}, err
	}
	var numlist common.CalculateDeposit
	numlist.AddressA0 = address
	numlist.CalcDeposit = append(numlist.CalcDeposit, common.OperationalInterestSlash{OperAmount: ret})
	return numlist, nil
}
func (v1 *DepositMnger_v1) GetAllInterest_v2(stateDB vm.StateDBManager, headTime uint64) map[common.Address]common.CalculateDeposit {
	ret := v1.GetAllInterest(stateDB)
	if ret == nil {
		return nil
	}
	retmap := make(map[common.Address]common.CalculateDeposit)
	for k, v := range ret {
		var numlist common.CalculateDeposit
		numlist.AddressA0 = k
		numlist.CalcDeposit = append(numlist.CalcDeposit, common.OperationalInterestSlash{OperAmount: v})
		retmap[k] = numlist
	}
	return retmap
}
func (v1 *DepositMnger_v1) AddInterest_v2(stateDB vm.StateDBManager, address common.Address, reward common.CalculateDeposit) error {
	return nil
}
func (v1 *DepositMnger_v1) GetAllDepositByHash_v2(hash common.Hash, stateDB vm.StateDBManager, headtime uint64) ([]common.DepositBase, error) {
	return nil, nil
}
func (v1 *DepositMnger_v1) PayInterest(stateDB vm.StateDBManager, addrA0 common.Address, position uint64, amount *big.Int) error {
	return nil
}
func (v1 *DepositMnger_v1) ConversionDeposit(statedb vm.StateDBManager, t uint64) map[common.Address]common.CheckDepositInfo {
	if depositInfo.Contract == nil {
		depositInfo.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	}
	return depositInfo.MatrixDeposit.ConversionDeposit(depositInfo.Contract, statedb, t)
}
