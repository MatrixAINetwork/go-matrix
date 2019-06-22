// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package contractTest

import (
	"testing"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/core/vm/validatorGroup"
	"math/big"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/common"
)

func TestTransferOwnership(t *testing.T){
	bm := &vm.BaseMethod{
		Name:"transferOwnership",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.SstoreSetGas,
	}
	data,err := bm.Inputs().Pack(common.Address{123,112,222})
	if err != nil {
		t.Error(err)
	}

	ValidatorGroupTest(bm,data,t)
}
func TestSetSignAccount(t *testing.T){
	bm := &vm.BaseMethod{
		Name:"setSignAccount",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.SstoreSetGas,
	}
	data,err := bm.Inputs().Pack(common.Address{123,112,222})
	if err != nil {
		t.Error(err)
	}
	ValidatorGroupTest(bm,data,t)
}
func TestWithDrawAll(t *testing.T){
	bm := &vm.BaseMethod{
		Name:"withDrawAll",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.SstoreSetGas,
	}
	/*
	data,err := bm.Inputs().Pack(common.Address{123,112,222})
	if err != nil {
		t.Error(err)
	}
	*/
	ValidatorGroupTest(bm,[]byte{},t)
}
func TestWithDraw(t *testing.T){
	bm := &vm.BaseMethod{
		Name:"withDraw",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.SstoreSetGas,
	}

	data,err := bm.Inputs().Pack(common.Address{123,112,222})
	if err != nil {
		t.Error(err)
	}
	ValidatorGroupTest(bm,data,t)
}
/*
func TestRefund(t *testing.T){
	bm := &vm.BaseMethod{
		Name:"refund",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.SstoreSetGas,
	}
	data,err := bm.Inputs().Pack(big.NewInt(1))
	if err != nil {
		t.Error(err)
	}
	ValidatorGroupTest(bm,data,t)
}
*/
func ValidatorGroupTest(bm *vm.BaseMethod,data []byte,t *testing.T){
	contract := vm.NewValidatorGroup()
	contract.TransferOwnershipMethod()
	inputs := make([]byte,4)
	ID :=bm.MethodID()
	copy(inputs,ID[:])
	inputs = append(inputs,data...)
	contract.RequiredGas(inputs)
	value := big.NewInt(1e18)
	value.Mul(value,big.NewInt(1e6))
	ownerAccount := common.Address{123,112,111}
	conPtr := vm.NewContract(vm.AccountRef(ownerAccount),
		vm.AccountRef(common.BytesToAddress([]byte{20})), value, uint64(1000000), params.MAN_COIN)
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{},mdb, state.NewDatabase(mdb))
	statedb.SetState(params.MAN_COIN, common.Address{}, common.BytesToHash([]byte(params.DepositVersionKey_1)), common.BytesToHash([]byte(params.DepositVersion_1)))
	env := vm.NewEVM(vm.Context{BlockNumber:big.NewInt(2),Time:big.NewInt(12345)}, statedb, params.TestChainConfig, vm.Config{}, params.MAN_COIN)
	env.CanTransfer = func(db vm.StateDBManager, address common.Address, amount *big.Int, coin string) bool {
		return true
	}
	env.Transfer = func(vm.StateDBManager, common.Address, common.Address, *big.Int, string) { return }
	//	contract.Run(bm.methodID()[:])
	contract.Constructor(common.Address{11,11,11},common.Address{22,22,22},common.Address{123,112,111},
	big.NewInt(0),big.NewInt(0),big.NewInt(0),nil,conPtr,env)
	_,err := vm.RunPrecompiledContract(contract,inputs,conPtr,env)
	if err!= nil{
		t.Error(err)
	}
}