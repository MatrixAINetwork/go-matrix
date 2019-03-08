// Copyright (c) 2018-2019 The MATRIX Authors
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

type DepositInfo struct {
	MatrixDeposit vm.MatrixDeposit
	Contract      *vm.Contract
	manApi        manBackend
	p             vm.PrecompiledContract
}

var depositInfo *DepositInfo

func NewDepositInfo(manApi manBackend) {
	pre := vm.PrecompiledContractsByzantium[common.BytesToAddress([]byte{10})]
	depositInfo = &DepositInfo{manApi: manApi, p: pre}
	//getDepositListTest()
}

func AddOnlineTime(stateDB vm.StateDBManager, address common.Address, ot *big.Int) error {
	return depositInfo.MatrixDeposit.AddOnlineTime(depositInfo.Contract, stateDB, address, ot)
}

func GetOnlineTime(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	return depositInfo.MatrixDeposit.GetOnlineTime(depositInfo.Contract, stateDB, address), nil
}

func SetOnlineTime(stateDB vm.StateDBManager, address common.Address, ot *big.Int) error {
	return depositInfo.MatrixDeposit.SetOnlineTime(depositInfo.Contract, stateDB, address, ot)
}

func GetDepositList(tm *big.Int, getDeposit common.RoleType) ([]vm.DepositDetail, error) {
	db, err := getDepositInfo(tm)
	if err != nil {
		return nil, err
	}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	//var depositList []vm.DepositDetail
	depositList := make([]vm.DepositDetail, 0)
	if common.RoleValidator == common.RoleValidator&getDeposit {

		depositList = append(depositList, depositInfo.MatrixDeposit.GetValidatorDepositList(contract, db)...)
	}

	if common.RoleMiner == common.RoleMiner&getDeposit {
		depositList = append(depositList, depositInfo.MatrixDeposit.GetMinerDepositList(contract, db)...)
	}
	return depositList, nil
}

func GetDepositAndWithDrawList(tm *big.Int) ([]vm.DepositDetail, error) {
	db, err := getDepositInfo(tm)
	if err != nil {
		return nil, err
	}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	var depositList []vm.DepositDetail
	depositList = depositInfo.MatrixDeposit.GetAllDepositList(contract, db, true)
	return depositList, nil
}

func GetAllDeposit(tm *big.Int) ([]vm.DepositDetail, error) {
	db, err := getDepositInfo(tm)
	if err != nil {
		return nil, err
	}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	var depositList []vm.DepositDetail
	depositList = depositInfo.MatrixDeposit.GetAllDepositList(contract, db, false)
	return depositList, nil
}

func GetDepositListByHash(hash common.Hash, getDeposit common.RoleType) ([]vm.DepositDetail, error) {
	db, err := getDepositInfoByHash(hash)
	if err != nil {
		return nil, err
	}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	//var depositList []vm.DepositDetail
	depositList := make([]vm.DepositDetail, 0)
	if common.RoleValidator == common.RoleValidator&getDeposit {

		depositList = append(depositList, depositInfo.MatrixDeposit.GetValidatorDepositList(contract, db)...)
	}

	if common.RoleMiner == common.RoleMiner&getDeposit {
		depositList = append(depositList, depositInfo.MatrixDeposit.GetMinerDepositList(contract, db)...)
	}
	return depositList, nil
}

func GetDepositAndWithDrawListByHash(hash common.Hash) ([]vm.DepositDetail, error) {
	db, err := getDepositInfoByHash(hash)
	if err != nil {
		return nil, err
	}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	var depositList []vm.DepositDetail
	depositList = depositInfo.MatrixDeposit.GetAllDepositList(contract, db, true)
	return depositList, nil
}

func GetAllDepositByHash(hash common.Hash) ([]vm.DepositDetail, error) {
	db, err := getDepositInfoByHash(hash)
	if err != nil {
		return nil, err
	}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	var depositList []vm.DepositDetail
	depositList = depositInfo.MatrixDeposit.GetAllDepositList(contract, db, false)
	return depositList, nil
}

func getDepositInfo(tm *big.Int) (db vm.StateDBManager, err error) {
	depositInfo.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	var c context.Context
	var h rpc.BlockNumber
	h = rpc.BlockNumber(tm.Int64())
	//encode := hexutil.EncodeBig(tm)
	//err = h.UnmarshalJSON([]byte(encode))
	//if err != nil {
	//	return nil, err
	//}
	db, _, err = depositInfo.manApi.StateAndHeaderByNumber(c, h)
	if err != nil {
		return nil, err
	}

	if db == nil {
		return nil, errors.New("db is nil")
	}
	dbValue := reflect.ValueOf(db)
	if dbValue.Kind() == reflect.Ptr && dbValue.IsNil() {
		return nil, errors.New("db is nil")
	}

	return db, nil
}

func getDepositInfoByHash(hash common.Hash) (db vm.StateDBManager, err error) {
	depositInfo.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	var c context.Context
	db, _, err = depositInfo.manApi.StateAndHeaderByHash(c, hash)
	if err != nil {
		return nil, err
	}

	if db == nil {
		return nil, errors.New("db is nil")
	}
	dbValue := reflect.ValueOf(db)
	if dbValue.Kind() == reflect.Ptr && dbValue.IsNil() {
		return nil, errors.New("db is nil")
	}

	return db, nil
}

func ResetSlash(stateDB vm.StateDBManager, address common.Address) error {
	return depositInfo.MatrixDeposit.ResetSlash(depositInfo.Contract, stateDB, address)
}

func GetSlash(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	return depositInfo.MatrixDeposit.GetSlash(depositInfo.Contract, stateDB, address), nil
}

func GetAllSlash(stateDB vm.StateDBManager) map[common.Address]*big.Int {
	return depositInfo.MatrixDeposit.GetAllSlash(depositInfo.Contract, stateDB)
}

func AddSlash(stateDB vm.StateDBManager, address common.Address, slash *big.Int) error {
	return depositInfo.MatrixDeposit.AddSlash(depositInfo.Contract, stateDB, address, slash)
}

func SetSlash(stateDB vm.StateDBManager, address common.Address, slash *big.Int) error {
	return depositInfo.MatrixDeposit.SetSlash(depositInfo.Contract, stateDB, address, slash)
}

func ResetInterest(stateDB vm.StateDBManager, address common.Address) error {
	return depositInfo.MatrixDeposit.ResetInterest(depositInfo.Contract, stateDB, address)
}

func GetInterest(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	return depositInfo.MatrixDeposit.GetInterest(depositInfo.Contract, stateDB, address), nil
}

func GetAllInterest(stateDB vm.StateDBManager) map[common.Address]*big.Int {
	return depositInfo.MatrixDeposit.GetAllInterest(depositInfo.Contract, stateDB)
}

func AddInterest(stateDB vm.StateDBManager, address common.Address, reward *big.Int) error {
	return depositInfo.MatrixDeposit.AddInterest(depositInfo.Contract, stateDB, address, reward)
}

func SetInterest(stateDB vm.StateDBManager, address common.Address, reward *big.Int) error {
	return depositInfo.MatrixDeposit.SetInterest(depositInfo.Contract, stateDB, address, reward)
}
func GetDeposit(stateDB vm.StateDBManager, address common.Address) *big.Int {
	return depositInfo.MatrixDeposit.GetDeposit(depositInfo.Contract, stateDB, address)
}

func SetDeposit(stateDB vm.StateDBManager, address common.Address) error {
	return depositInfo.MatrixDeposit.SetDeposit(depositInfo.Contract, stateDB, address)
}
func AddDeposit(stateDB vm.StateDBManager, address common.Address) error {
	return depositInfo.MatrixDeposit.AddDeposit(depositInfo.Contract, stateDB, address)
}

// 获取A0账户
func GetDepositAccount(stateDB vm.StateDBManager, authAccount common.Address) common.Address {
	if depositInfo.Contract == nil {
		depositInfo.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	}
	return depositInfo.MatrixDeposit.GetDepositAccount(depositInfo.Contract, stateDB, authAccount)
}

// 获取A1账户
func GetAuthAccount(stateDB vm.StateDBManager, depositAccount common.Address) common.Address {
	if depositInfo.Contract == nil {
		depositInfo.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	}
	return depositInfo.MatrixDeposit.GetAuthAccount(depositInfo.Contract, stateDB, depositAccount)
}
