// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package depoistInfo

import (
	"context"
	"fmt"
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
	var depositList []vm.DepositDetail
	switch getDeposit {
	case common.RoleValidator:
		depositList = depositInfo.MatrixDeposit.GetValidatorDepositList(contract, db)
	case common.RoleMiner:
		depositList = depositInfo.MatrixDeposit.GetMinerDepositList(contract, db)
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

func getDepositListTest() {
	db, err := getDepositInfo(big.NewInt(0))
	if err != nil {
		return
	}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000)
	address := depositInfo.MatrixDeposit.GetValidatorDepositList(contract, db)
	fmt.Printf("get depositList:%v %d\n", address, len(address))
	address = depositInfo.MatrixDeposit.GetMinerDepositList(contract, db)
	fmt.Printf("get miner:%v %d\n", address, len(address))
}
