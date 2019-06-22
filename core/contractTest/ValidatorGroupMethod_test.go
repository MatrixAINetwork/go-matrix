// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package contractTest

import (
	"testing"
	"math/big"
	"time"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/core/vm/validatorGroup"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/reward/depositcfg"
)
var (
	testOwnerAddr = common.Address{111,222,111,222}
	testA1Addr = common.Address{11,11,11,11,11,11}
)
func TestEmuliateCurrent(t *testing.T){
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{},mdb, state.NewDatabase(mdb))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,testOwnerAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,vm.ValidatorGroupContractAddress,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	createValidatorGroup(statedb,t)
	vcStates := &vm.ValidatorContractState{}
	valiMap,err := vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	if err != nil {
		t.Error(err)
	}
	var contractAddr common.Address
	for key,_ := range valiMap{
		contractAddr = key
		break
	}
	from := common.Address{33,33,33,33}
	for i:=0;i<100;i++{
		from[5] = byte(i)
		statedb.SetBalance(params.MAN_COIN, common.MainAccount,contractAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
		value := new(big.Int).Mul(big.NewInt(100),big.NewInt(1e18))
		addDevalopMethod(from,contractAddr,value,big.NewInt(0),statedb,t)
		valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index,exist := valiInfo.ValidatorMap.Find(from)
		if !exist{
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if depInfo.Current.Amount.Cmp(value) != 0{
			t.Error("Deposit Value Error")
		}
	}
	testValue := new(big.Int).Mul(big.NewInt(200),big.NewInt(1e18))
	for i:=0;i<100;i++{
		from[5] = byte(i)
		value := new(big.Int).Mul(big.NewInt(100),big.NewInt(1e18))
		addDevalopMethod(from,contractAddr,value,big.NewInt(0),statedb,t)
		valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index,exist := valiInfo.ValidatorMap.Find(from)
		if !exist{
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if depInfo.Current.Amount.Cmp(testValue) != 0{
			t.Error("Deposit Value Error")
		}
	}
	testValue = new(big.Int).Mul(big.NewInt(300),big.NewInt(1e18))
	for i:=0;i<100;i++{
		from[5] = byte(i)
		value := new(big.Int).Mul(big.NewInt(100),big.NewInt(1e18))
		addDevalopMethod(from,contractAddr,value,big.NewInt(0),statedb,t)
		valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index,exist := valiInfo.ValidatorMap.Find(from)
		if !exist{
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if depInfo.Current.Amount.Cmp(testValue) != 0{
			t.Error("Deposit Value Error")
		}
	}
	testValue = new(big.Int).Mul(big.NewInt(200),big.NewInt(1e18))
	for i:=0;i<100;i++{
		from[5] = byte(i)
		value := new(big.Int).Mul(big.NewInt(100),big.NewInt(1e18))
		withdrawMethod(from,contractAddr,value,big.NewInt(0),statedb,t)
		valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index,exist := valiInfo.ValidatorMap.Find(from)
		if !exist{
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if depInfo.Current.Amount.Cmp(testValue) != 0{
			t.Error("Deposit Value Error")
		}
		if len(depInfo.Current.WithdrawList) != 1{
			t.Error("WithdrawList length Error")
		}
	}
	from[5] = byte(0)
	refundCurrentMethod(from,contractAddr,big.NewInt(0),statedb,t)
	/*
	for i:=0;i<100;i++{
		from[5] = byte(i)
//		value := new(big.Int).Mul(big.NewInt(100),big.NewInt(1e18))
		refundCurrentMethod(from,contractAddr,big.NewInt(0),statedb,t)
		valiMap,err = vcStates.GetValidatorGroupInfo(statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index,exist := valiInfo.ValidatorMap.Find(from)
		if !exist{
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if depInfo.Current.Amount.Cmp(testValue) != 0{
			t.Error("Deposit Value Error")
		}
		if len(depInfo.Current.WithdrawList) != 0{
			t.Error("WithdrawList length Error")
		}
	}
	*/
	valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()+depositcfg.Days7Seconds),statedb)
	data,_ := json.Marshal(valiMap)
	t.Log(string(data))
}
func TestEmuliateDtype(t *testing.T){
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{},mdb, state.NewDatabase(mdb))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,testOwnerAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	createValidatorGroup(statedb,t)
	vcStates := &vm.ValidatorContractState{}
	valiMap,err := vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	if err != nil {
		t.Error(err)
	}
	var contractAddr common.Address
	for key,_ := range valiMap{
		contractAddr = key
		break
	}
	from := common.Address{33,33,33,33}
	for i:=0;i<100;i++{
		from[5] = byte(i)
		statedb.SetBalance(params.MAN_COIN, common.MainAccount,contractAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
		value := new(big.Int).Mul(big.NewInt(10000),big.NewInt(1e18))
		addDevalopMethod(from,contractAddr,value,big.NewInt(1),statedb,t)
		valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index,exist := valiInfo.ValidatorMap.Find(from)
		if !exist{
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if len(depInfo.Positions) == 1 && depInfo.Positions[len(depInfo.Positions)-1].DType == 1 &&
			depInfo.Positions[len(depInfo.Positions)-1].Amount.Cmp(value) == 0 {
		}else{
			t.Error("Deposit Value Error")
		}
	}
	for i:=0;i<100;i++{
		from[5] = byte(i)
		value := new(big.Int).Mul(big.NewInt(10000),big.NewInt(1e18))
		addDevalopMethod(from,contractAddr,value,big.NewInt(1),statedb,t)
		valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index,exist := valiInfo.ValidatorMap.Find(from)
		if !exist{
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if len(depInfo.Positions) == 2 && depInfo.Positions[len(depInfo.Positions)-1].DType == 1 &&
			depInfo.Positions[len(depInfo.Positions)-1].Amount.Cmp(value) == 0 {
		}else{
			t.Error("Deposit Value Error")
		}
	}
	for i:=0;i<100;i++{
		from[5] = byte(i)
		//value := new(big.Int).Mul(big.NewInt(100),big.NewInt(1e18))
		withdrawMethod(from,contractAddr,big.NewInt(0),big.NewInt(int64(i+1)),statedb,t)
		valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index,exist := valiInfo.ValidatorMap.Find(from)
		if !exist{
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if len(depInfo.Positions) == 0{
			t.Error("Deposit Value Error")
		}
	}
	valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()+depositcfg.Days7Seconds),statedb)
	data,_ := json.Marshal(valiMap)
	t.Log(string(data))
}
func addDevalopMethod(from,contractAddr common.Address,value,dType *big.Int,state vm.StateDBManager,t *testing.T){
	bm := &vm.BaseMethod{
		Name:"addDeposit",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.SstoreSetGas,
	}
	inputs := make([]byte,4)
	ID :=bm.MethodID()
	copy(inputs,ID[:])
	//function createValidatorGroup(address signAcount,uint amount,uint dType,uint ownerRate,uint[] lvlRate0)
	data,err := bm.Inputs().Pack(dType)
	if err != nil {
		t.Error(err)
		return
	}
	inputs = append(inputs,data...)
//	contract.RequiredGas(inputs)

	conPtr := vm.NewContract(vm.AccountRef(from),
		vm.AccountRef(contractAddr), value, uint64(1000000), params.MAN_COIN)
	GroupMethodTest(bm,big.NewInt(time.Now().Unix()),inputs,conPtr,state,t)
}
func withdrawMethod(from,contractAddr common.Address,value,position *big.Int,state vm.StateDBManager,t *testing.T){
	bm := &vm.BaseMethod{
		Name:"withdraw",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.SstoreSetGas,
	}
	inputs := make([]byte,4)
	ID :=bm.MethodID()
	copy(inputs,ID[:])
	//function createValidatorGroup(address signAcount,uint amount,uint dType,uint ownerRate,uint[] lvlRate0)
	data,err := bm.Inputs().Pack(value,position)
	if err != nil {
		t.Error(err)
		return
	}
	inputs = append(inputs,data...)
	//	contract.RequiredGas(inputs)

	conPtr := vm.NewContract(vm.AccountRef(from),
		vm.AccountRef(contractAddr), value, uint64(1000000), params.MAN_COIN)
	GroupMethodTest(bm,big.NewInt(time.Now().Unix()),inputs,conPtr,state,t)
}
func refundCurrentMethod(from,contractAddr common.Address,position *big.Int,state vm.StateDBManager,t *testing.T){
	bm := &vm.BaseMethod{
		Name:"refund",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.SstoreSetGas,
	}
	inputs := make([]byte,4)
	ID :=bm.MethodID()
	copy(inputs,ID[:])
	//function createValidatorGroup(address signAcount,uint amount,uint dType,uint ownerRate,uint[] lvlRate0)
	data,err := bm.Inputs().Pack(position)
	if err != nil {
		t.Error(err)
		return
	}
	inputs = append(inputs,data...)
	//	contract.RequiredGas(inputs)

	conPtr := vm.NewContract(vm.AccountRef(from),
		vm.AccountRef(contractAddr), big.NewInt(0), uint64(1000000), params.MAN_COIN)
	GroupMethodTest(bm,big.NewInt(time.Now().Unix()+depositcfg.Days7Seconds),inputs,conPtr,state,t)
}
func createValidatorGroup(state vm.StateDBManager,t *testing.T){
	contract := vm.NewValidatorGroupContract()
	bm := &vm.BaseMethod{
		Name:"createValidatorGroup",
		Abi:&validatorGroup.ValidatorGroupContractAbi,
		GasUsed:params.TxGasContractCreation,
	}
	inputs := make([]byte,4)
	ID :=bm.MethodID()
	copy(inputs,ID[:])
	//function createValidatorGroup(address signAcount,uint amount,uint dType,uint ownerRate,uint[] lvlRate0)
	data,err := bm.Inputs().Pack(testA1Addr,big.NewInt(0),big.NewInt(10),big.NewInt(2e8),[]*big.Int{big.NewInt(10),big.NewInt(20),big.NewInt(30)})
	if err != nil {
		t.Error(err)
		return
	}
	inputs = append(inputs,data...)
	contract.RequiredGas(inputs)
	value := big.NewInt(1e18)
	value.Mul(value,big.NewInt(1e6))
	conPtr := vm.NewContract(vm.AccountRef(testOwnerAddr),
		vm.AccountRef(vm.ValidatorGroupContractAddress), value, uint64(1000000), params.MAN_COIN)
	env := vm.NewEVM(vm.Context{BlockNumber:big.NewInt(2),Time:big.NewInt(time.Now().Unix())}, state, params.TestChainConfig, vm.Config{}, params.MAN_COIN)
	env.CanTransfer = core.CanTransfer
	env.Transfer = core.Transfer
	//	contract.Run(bm.methodID()[:])
	_,err = vm.RunPrecompiledContract(contract,inputs,conPtr,env)
	if err != nil {
		t.Error(err)
	}
	nonce := env.StateDB.GetNonce("MAN",conPtr.CallerAddress)
	env.StateDB.SetNonce("MAN",conPtr.CallerAddress,nonce+1)
}
func GroupMethodTest(bm *vm.BaseMethod,time *big.Int,data []byte,contract *vm.Contract,state vm.StateDBManager,t *testing.T){
	conGroup := vm.NewValidatorGroup()
//	contract.TransferOwnershipMethod()
	conGroup.RequiredGas(data)

	env := vm.NewEVM(vm.Context{BlockNumber:big.NewInt(2),Time:time}, state, params.TestChainConfig, vm.Config{}, params.MAN_COIN)
	env.CanTransfer = core.CanTransfer
	env.Transfer = core.Transfer
	//	contract.Run(bm.methodID()[:])
	_,err := vm.RunPrecompiledContract(conGroup,data,contract,env)
	if err!= nil{
		t.Error(err)
	}
	nonce := env.StateDB.GetNonce("MAN",contract.CallerAddress)
	env.StateDB.SetNonce("MAN",contract.CallerAddress,nonce+1)
}
func CallMethodTest(from,contractAddr common.Address,value *big.Int,data []byte,time *big.Int,state vm.StateDBManager,t *testing.T){
	env := vm.NewEVM(vm.Context{BlockNumber:big.NewInt(2),Time:time}, state, params.TestChainConfig, vm.Config{}, params.MAN_COIN)
	env.CanTransfer = core.CanTransfer
	env.Transfer = core.Transfer
	//	contract.Run(bm.methodID()[:])
	_,_,_,err := env.Call(vm.AccountRef(from),contractAddr,data,2e6,value)
	if err!= nil{
		t.Error(err)
	}
//	t.Log(ret)
//	nonce := env.StateDB.GetNonce("MAN",contract.CallerAddress)
//	env.StateDB.SetNonce("MAN",contract.CallerAddress,nonce+1)

}