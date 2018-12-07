// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package depoistInfo

import (
	"context"
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/hexutil"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/rpc"
)

type manBackend interface {
	StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error)
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

func AddOnlineTime(stateDB vm.StateDB, address common.Address, ot *big.Int) error {
	return depositInfo.MatrixDeposit.AddOnlineTime(depositInfo.Contract, stateDB, address, ot)
}

func GetOnlineTime(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
	return depositInfo.MatrixDeposit.GetOnlineTime(depositInfo.Contract, stateDB, address), nil
}

func SetOnlineTime(stateDB vm.StateDB, address common.Address, ot *big.Int) error {
	return depositInfo.MatrixDeposit.SetOnlineTime(depositInfo.Contract, stateDB, address, ot)
}

func GetDepositList(tm *big.Int, getDeposit common.RoleType) ([]vm.DepositDetail, error) {
	db, err := getDepositInfo(tm)
	if err != nil {
		return nil, err
	}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000)
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
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000)
	var depositList []vm.DepositDetail
	depositList = depositInfo.MatrixDeposit.GetAllDepositList(contract, db, true)
	return depositList, nil
}

func GetAllDeposit(tm *big.Int) ([]vm.DepositDetail, error) {
	db, err := getDepositInfo(tm)
	if err != nil {
		return nil, err
	}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000)
	var depositList []vm.DepositDetail
	depositList = depositInfo.MatrixDeposit.GetAllDepositList(contract, db, false)
	return depositList, nil
}

func getDepositInfo(tm *big.Int) (db vm.StateDB, err error) {
	depositInfo.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0)
	var c context.Context
	var h rpc.BlockNumber
	encode := hexutil.EncodeBig(tm)
	err = h.UnmarshalJSON([]byte(encode))
	if err != nil {
		return nil, err
	}
	db, _, err = depositInfo.manApi.StateAndHeaderByNumber(c, h)
	return db, err
}

func ResetSlash(stateDB vm.StateDB, address common.Address) error {
	return depositInfo.MatrixDeposit.ResetSlash(depositInfo.Contract, stateDB, address)
}

func GetSlash(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
	return depositInfo.MatrixDeposit.GetSlash(depositInfo.Contract, stateDB, address), nil
}

func GetAllSlash(stateDB vm.StateDB) map[common.Address]*big.Int {
	return depositInfo.MatrixDeposit.GetAllSlash(depositInfo.Contract, stateDB)
}

func AddSlash(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
	return depositInfo.MatrixDeposit.AddSlash(depositInfo.Contract, stateDB, address, slash)
}

func SetSlash(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
	return depositInfo.MatrixDeposit.SetSlash(depositInfo.Contract, stateDB, address, slash)
}

func ResetInterest(stateDB vm.StateDB, address common.Address) error {
	return depositInfo.MatrixDeposit.ResetInterest(depositInfo.Contract, stateDB, address)
}

func GetInterest(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
	return depositInfo.MatrixDeposit.GetInterest(depositInfo.Contract, stateDB, address), nil
}

func GetAllInterest(stateDB vm.StateDB) map[common.Address]*big.Int {
	return depositInfo.MatrixDeposit.GetAllInterest(depositInfo.Contract, stateDB)
}

func AddInterest(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
	return depositInfo.MatrixDeposit.AddInterest(depositInfo.Contract, stateDB, address, reward)
}

func SetInterest(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
	return depositInfo.MatrixDeposit.SetInterest(depositInfo.Contract, stateDB, address, reward)
}

func GetDeposit(stateDB vm.StateDB) *big.Int {
	return depositInfo.MatrixDeposit.GetDeposit(depositInfo.Contract, stateDB)
}

func SetDeposit(stateDB vm.StateDB, deposit *big.Int) error {
	return depositInfo.MatrixDeposit.SetDeposit(depositInfo.Contract, stateDB, deposit)
}
