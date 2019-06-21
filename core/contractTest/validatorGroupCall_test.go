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
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/core/vm/validatorGroup"
	"github.com/MatrixAINetwork/go-matrix/reward/depositcfg"
)

func TestCallCurrent(t *testing.T){
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{},mdb, state.NewDatabase(mdb))
	statedb.SetState(params.MAN_COIN, common.Address{}, common.BytesToHash([]byte(params.DepositVersionKey_1)), common.BytesToHash([]byte(params.DepositVersion_1)))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,testOwnerAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	callCreateValidatorGroup(statedb,t)
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
	balance := statedb.GetBalance(params.MAN_COIN,contractAddr)
	t.Log(balance)
	from := common.Address{33,33,33,33}
	for i:=0;i<100;i++{
		from[5] = byte(i)
		statedb.SetBalance(params.MAN_COIN, common.MainAccount,from,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
		value := new(big.Int).Mul(big.NewInt(100),big.NewInt(1e18))
		callAddDevalopMethod(from,contractAddr,value,big.NewInt(0),statedb,t)
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
		callAddDevalopMethod(from,contractAddr,value,big.NewInt(0),statedb,t)
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
		callAddDevalopMethod(from,contractAddr,value,big.NewInt(0),statedb,t)
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
	vm.NewValidatorGroup().TransferCurrentInterests(testValue,uint64(time.Now().Unix()),contractAddr,statedb)

	testValue = new(big.Int).Mul(big.NewInt(200),big.NewInt(1e18))
	for i:=0;i<100;i++{
		from[5] = byte(i)
		value := new(big.Int).Mul(big.NewInt(100),big.NewInt(1e18))
		callWithdrawMethod(from,contractAddr,value,big.NewInt(0),statedb,t)
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
	vm.NewValidatorGroup().TransferCurrentInterests(testValue,uint64(time.Now().Unix()),contractAddr,statedb)
//	valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()+vm.Days7Seconds),statedb)
//	data,_ := json.Marshal(valiMap)
//	t.Log(string(data))
//	return
	from[5] = byte(0)
	callRefundCurrentMethod(from,contractAddr,big.NewInt(0),statedb,t)
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
func TestCallOwnerWithdrawAll(t *testing.T){
	vcStates := &vm.ValidatorContractState{}
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{},mdb, state.NewDatabase(mdb))
	statedb.SetState(params.MAN_COIN, common.Address{}, common.BytesToHash([]byte(params.DepositVersionKey_1)), common.BytesToHash([]byte(params.DepositVersion_1)))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,testOwnerAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	callCreateValidatorGroup(statedb,t)
	valiMap,err := vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()+depositcfg.Days7Seconds),statedb)
	data,_ := json.Marshal(valiMap)
	//t.Log(string(data))
	callCreateValidatorGroup(statedb,t)
	valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	data,_ = json.Marshal(valiMap)
	//t.Log(string(data))
	info := MakeJsonInferface(valiMap)
	data,_ = json.Marshal(info)
	t.Log(string(data))
	if err != nil {
		t.Error(err)
	}
	var contractAddr common.Address
	for key,_ := range valiMap{
		contractAddr = key
		break
	}
	balance := statedb.GetBalance(params.MAN_COIN,contractAddr)
	t.Log(balance)
	from := testOwnerAddr
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,from,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	value := new(big.Int).Mul(big.NewInt(10000),big.NewInt(1e18))
	callAddDevalopMethod(from,contractAddr,value,big.NewInt(1),statedb,t)
	value = new(big.Int).Mul(big.NewInt(100),big.NewInt(1e18))
	callWithdrawAllMethod(from,contractAddr,value,big.NewInt(0),statedb,t)
	callRefundCurrentMethod(from,contractAddr,big.NewInt(0),statedb,t)
	callRefundCurrentMethod(from,contractAddr,big.NewInt(1),statedb,t)
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
	info = MakeJsonInferface(valiMap)
	data,_ = json.Marshal(info)
	t.Log(string(data))
}
func TestCallOwnerWithdraw(t *testing.T){
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{},mdb, state.NewDatabase(mdb))
	statedb.SetState(params.MAN_COIN, common.Address{}, common.BytesToHash([]byte(params.DepositVersionKey_1)), common.BytesToHash([]byte(params.DepositVersion_1)))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,testOwnerAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	callCreateValidatorGroup(statedb,t)
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
	balance := statedb.GetBalance(params.MAN_COIN,contractAddr)
	t.Log(balance)
	from := testOwnerAddr
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,from,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	value := new(big.Int).Mul(big.NewInt(10000),big.NewInt(1e18))
	callAddDevalopMethod(from,contractAddr,value,big.NewInt(1),statedb,t)
	value = new(big.Int).Mul(big.NewInt(100),big.NewInt(1e18))
	callWithdrawMethod(from,contractAddr,value,big.NewInt(0),statedb,t)
	callRefundCurrentMethod(from,contractAddr,big.NewInt(0),statedb,t)
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
func callCreateValidatorGroup(state vm.StateDBManager,t *testing.T){
	bm := &vm.BaseMethod{
		Name:"createValidatorGroup",
		Abi:&validatorGroup.ValidatorGroupContractAbi,
		GasUsed:params.TxGasContractCreation,
	}
	inputs := make([]byte,4)
	ID :=bm.MethodID()
	copy(inputs,ID[:])
	//function createValidatorGroup(address signAcount,uint amount,uint dType,uint ownerRate,uint[] lvlRate0)
	testA1Addr[5]++
	data,err := bm.Inputs().Pack(testA1Addr,big.NewInt(0),big.NewInt(10),[]*big.Int{big.NewInt(10),big.NewInt(20),big.NewInt(30)})
	if err != nil {
		t.Error(err)
		return
	}
	inputs = append(inputs,data...)
	value := big.NewInt(1e18)
	value.Mul(value,big.NewInt(1e5))
	CallMethodTest(testOwnerAddr,vm.ValidatorGroupContractAddress,value,inputs,big.NewInt(time.Now().Unix()),state,t)
	balance := state.GetBalance(params.MAN_COIN,testOwnerAddr)
	t.Log(balance)
	balance = state.GetBalance(params.MAN_COIN,vm.ValidatorGroupContractAddress)
	t.Log(balance)
	state.SetNonce(params.MAN_COIN,testOwnerAddr,state.GetNonce(params.MAN_COIN,testOwnerAddr)+1)
}
func callAddDevalopMethod(from,contractAddr common.Address,value,dType *big.Int,state vm.StateDBManager,t *testing.T){
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
	CallMethodTest(from,contractAddr,value,inputs,big.NewInt(time.Now().Unix()),state,t)
	if from[5] == 0{
		balance := state.GetBalance(params.MAN_COIN,from)
		t.Log(balance)
		balance = state.GetBalance(params.MAN_COIN,contractAddr)
		t.Log(balance)
	}
}
func callWithdrawMethod(from,contractAddr common.Address,value,position *big.Int,state vm.StateDBManager,t *testing.T){
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
	CallMethodTest(from,contractAddr,big.NewInt(0),inputs,big.NewInt(time.Now().Unix()),state,t)
	if from[5] == 0{
		balance := state.GetBalance(params.MAN_COIN,from)
		t.Log(balance)
		balance = state.GetBalance(params.MAN_COIN,contractAddr)
		t.Log(balance)
	}
}
func callWithdrawAllMethod(from,contractAddr common.Address,value,position *big.Int,state vm.StateDBManager,t *testing.T){
	bm := &vm.BaseMethod{
		Name:"withdrawAll",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.SstoreSetGas,
	}
	inputs := make([]byte,4)
	ID :=bm.MethodID()
	copy(inputs,ID[:])
	//function createValidatorGroup(address signAcount,uint amount,uint dType,uint ownerRate,uint[] lvlRate0)
	//	contract.RequiredGas(inputs)
	CallMethodTest(from,contractAddr,big.NewInt(0),inputs,big.NewInt(time.Now().Unix()),state,t)
}
func callRefundCurrentMethod(from,contractAddr common.Address,position *big.Int,state vm.StateDBManager,t *testing.T){
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

	CallMethodTest(from,contractAddr,big.NewInt(0),inputs,big.NewInt(time.Now().Unix()+depositcfg.SecondsPerMonth*2),state,t)
	if from[5] == 0{
		balance := state.GetBalance(params.MAN_COIN,from)
		t.Log(balance)
		balance = state.GetBalance(params.MAN_COIN,contractAddr)
		t.Log(balance)
	}
}
