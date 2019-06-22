// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package contractTest

import (
	"testing"
	"math/big"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"time"
	"github.com/MatrixAINetwork/go-matrix/core/vm/validatorGroup"
	"github.com/MatrixAINetwork/go-matrix/params"
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/reward/depositcfg"
)

func TestEmuliateCurrentReward(t *testing.T) {
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{}, mdb, state.NewDatabase(mdb))
	statedb.SetState(params.MAN_COIN, common.Address{}, common.BytesToHash([]byte(params.DepositVersionKey_1)), common.BytesToHash([]byte(params.DepositVersion_1)))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,vm.ValidatorGroupContractAddress,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,testOwnerAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	createValidatorGroup(statedb, t)
	vcStates := &vm.ValidatorContractState{}
	valiMap, err := vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	if err != nil {
		t.Error(err)
	}
	var contractAddr common.Address
	for key, _ := range valiMap {
		contractAddr = key
		break
	}
	from := common.Address{33, 33, 33, 33}
	for i := 0; i < 100; i++ {
		statedb.SetBalance(params.MAN_COIN, common.MainAccount,contractAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
		from[5] = byte(i)
		value := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
		addDevalopMethod(from, contractAddr, value, big.NewInt(0), statedb, t)
		valiMap, err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index, exist := valiInfo.ValidatorMap.Find(from)
		if !exist {
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if depInfo.Current.Amount.Cmp(value) != 0 {
			t.Error("Deposit Value Error")
		}
	}
	rewards := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))
	conGroup := vm.NewValidatorGroup()
	err = conGroup.TransferRewards(rewards,uint64(time.Now().Unix()),contractAddr,statedb)
	if err != nil {
		t.Error(err)
	}
	valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	if err != nil {
		t.Error(err)
	}
	/*
	allReward := big.NewInt(0)
	for _,info := range valiMap{
		for _,vali := range info.ValidatorMap{
			t.Log(vali.Reward)
			allReward.Add(allReward,vali.Reward)
		}
	}
	t.Log(rewards,allReward)
	allReward.Sub(rewards,allReward)
	t.Log(allReward)
	*/
	for i := 0; i < 100; i++ {
		from[5] = byte(i)
		value := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
		getCurrentReward(from, contractAddr,  statedb, t)
		valiMap, err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index, exist := valiInfo.ValidatorMap.Find(from)
		if !exist {
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if depInfo.Current.Amount.Cmp(value) != 0 {
			t.Error("Deposit Value Error")
		}
	}
	withdrawAllMethod(testOwnerAddr, contractAddr,  statedb, t)
	from[5] = byte(0)
	refundCurrentMethod(from,contractAddr,big.NewInt(0),statedb,t)
	for i := 1; i < 100; i++ {
		from[5] = byte(i)
//		value := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
		getCurrentReward(from, contractAddr,  statedb, t)
		valiMap, err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()+depositcfg.Days7Seconds),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		_, exist := valiInfo.ValidatorMap.Find(from)
		if exist {
			t.Error("Add Deposit Error")
		}
	}
	getCurrentReward(testOwnerAddr, contractAddr,  statedb, t)
	valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()+depositcfg.Days7Seconds),statedb)
	data,_ := json.Marshal(valiMap)
	t.Log(string(data))
}
func TestEmuliateTimeReward(t *testing.T) {
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{}, mdb, state.NewDatabase(mdb))
	statedb.SetState(params.MAN_COIN, common.Address{}, common.BytesToHash([]byte(params.DepositVersionKey_1)), common.BytesToHash([]byte(params.DepositVersion_1)))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,testOwnerAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,vm.ValidatorGroupContractAddress,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	createValidatorGroup(statedb, t)
	vcStates := &vm.ValidatorContractState{}
	valiMap, err := vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	if err != nil {
		t.Error(err)
	}
	var contractAddr common.Address
	for key, _ := range valiMap {
		contractAddr = key
		break
	}
	from := common.Address{33, 33, 33, 33}
	for i := 0; i < 100; i++ {
		statedb.SetBalance(params.MAN_COIN, common.MainAccount,contractAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
		statedb.SetBalance(params.MAN_COIN, common.MainAccount,from,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
		from[5] = byte(i)
		value := new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e18))
		addDevalopMethod(from, contractAddr, value, big.NewInt(1), statedb, t)
		valiMap, err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index, exist := valiInfo.ValidatorMap.Find(from)
		if !exist {
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if depInfo.Positions[0].Amount.Cmp(value) != 0 {
			t.Error("Deposit Value Error")
		}
	}
	from[5] = byte(0)
	rewards := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))
	transferReward(from,contractAddr,big.NewInt(0),big.NewInt(time.Now().Unix()),statedb,t)
	transferReward(from,contractAddr,rewards,big.NewInt(time.Now().Unix()),statedb,t)
	transferReward(from,contractAddr,big.NewInt(0),big.NewInt(time.Now().Unix()),statedb,t)
	valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	if err != nil {
		t.Error(err)
	}
	data,_ := json.Marshal(valiMap)
	t.Log(string(data))

	for i := 0; i < 100; i++ {
		from[5] = byte(i)
		getCurrentReward(from, contractAddr,  statedb, t)
		valiMap, err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		_, exist := valiInfo.ValidatorMap.Find(from)
		if !exist {
			t.Error("Add Deposit Error")
		}
	}

	withdrawAllMethod(testOwnerAddr, contractAddr,  statedb, t)

	from[5] = byte(0)
	for i := 0; i < 100; i++ {
		from[5] = byte(i)
		//		value := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
		refundCurrentMethod(from,contractAddr,big.NewInt(int64(i+1)),statedb,t)
		valiMap, err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()+depositcfg.Days7Seconds),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		_, exist := valiInfo.ValidatorMap.Find(from)
		if exist {
			t.Error("Add Deposit Error")
		}
	}
	/*
	for i := 0; i < 100; i++ {
		from[5] = byte(i)
		//		value := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
		getCurrentReward(from, contractAddr,  statedb, t)
		valiMap, err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()+vm.Days7Seconds),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		_, exist := valiInfo.ValidatorMap.Find(from)
		if exist {
			t.Error("Add Deposit Error")
		}
	}
*/
	refundCurrentMethod(testOwnerAddr,contractAddr,big.NewInt(0),statedb,t)
	getCurrentReward(testOwnerAddr, contractAddr,  statedb, t)
	valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()+depositcfg.Days7Seconds),statedb)
	data,_ = json.Marshal(valiMap)
	t.Log(string(data))
}
func TestEmuliateCurrentInterests(t *testing.T) {
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{}, mdb, state.NewDatabase(mdb))
	statedb.SetState(params.MAN_COIN, common.Address{}, common.BytesToHash([]byte(params.DepositVersionKey_1)), common.BytesToHash([]byte(params.DepositVersion_1)))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,testOwnerAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	statedb.SetBalance(params.MAN_COIN, common.MainAccount,vm.ValidatorGroupContractAddress,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
	createValidatorGroup(statedb, t)
	vcStates := &vm.ValidatorContractState{}
	valiMap, err := vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	if err != nil {
		t.Error(err)
	}
	var contractAddr common.Address
	for key, _ := range valiMap {
		contractAddr = key
		break
	}
	from := common.Address{33, 33, 33, 33}
	for i := 0; i < 100; i++ {
		statedb.SetBalance(params.MAN_COIN, common.MainAccount,contractAddr,new(big.Int).Mul(big.NewInt(1e10),big.NewInt(1e18)))
		from[5] = byte(i)
		value := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
		addDevalopMethod(from, contractAddr, value, big.NewInt(0), statedb, t)
		valiMap, err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index, exist := valiInfo.ValidatorMap.Find(from)
		if !exist {
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if depInfo.Current.Amount.Cmp(value) != 0 {
			t.Error("Deposit Value Error")
		}
	}
	interest := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))
	conGroup := vm.NewValidatorGroup()
	err = conGroup.TransferCurrentInterests(interest,uint64(time.Now().Unix()),contractAddr,statedb)
	if err != nil {
		t.Error(err)
	}
	valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	if err != nil {
		t.Error(err)
	}
	/*
	allReward := big.NewInt(0)
	for _,info := range valiMap{
		for _,vali := range info.ValidatorMap{
			t.Log(vali.Reward)
			allReward.Add(allReward,vali.Reward)
		}
	}
	t.Log(rewards,allReward)
	allReward.Sub(rewards,allReward)
	t.Log(allReward)
	*/
	for i := 0; i < 100; i++ {
		from[5] = byte(i)
		value := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
		getCurrentReward(from, contractAddr,  statedb, t)
		valiMap, err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		index, exist := valiInfo.ValidatorMap.Find(from)
		if !exist {
			t.Error("Add Deposit Error")
		}
		depInfo := valiInfo.ValidatorMap[index]
		if depInfo.Current.Amount.Cmp(value) != 0 {
			t.Error("Deposit Value Error")
		}
	}
/*
	withdrawAllMethod(testOwnerAddr, contractAddr,  statedb, t)
	from[5] = byte(0)
	refundCurrentMethod(from,contractAddr,big.NewInt(0),statedb,t)
	for i := 1; i < 100; i++ {
		from[5] = byte(i)
		//		value := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
		getCurrentReward(from, contractAddr,  statedb, t)
		valiMap, err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()+vm.Days7Seconds),statedb)
		if err != nil {
			t.Error(err)
		}
		valiInfo := valiMap[contractAddr]
		if valiInfo == nil {
			t.Error("Create Contract Error")
		}
		_, exist := valiInfo.ValidatorMap.Find(from)
		if exist {
			t.Error("Add Deposit Error")
		}
	}
	getCurrentReward(testOwnerAddr, contractAddr,  statedb, t)
	valiMap,err = vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()+vm.Days7Seconds),statedb)
	data,_ := json.Marshal(valiMap)
	t.Log(string(data))
*/
}
func transferReward(from,contractAddr common.Address,value,time *big.Int,state vm.StateDBManager,t *testing.T){
	CallMethodTest(from,contractAddr,value,nil,time,state,t)
	balance := state.GetBalance(params.MAN_COIN,from)
	t.Log(balance)
	balance = state.GetBalance(params.MAN_COIN,contractAddr)
	t.Log(balance)
}
func getCurrentReward(from,contractAddr common.Address,state vm.StateDBManager,t *testing.T)  {
	bm := &vm.BaseMethod{
		Name:"getReward",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.SstoreSetGas,
	}
	inputs := make([]byte,4)
	ID :=bm.MethodID()
	copy(inputs,ID[:])
	//function createValidatorGroup(address signAcount,uint amount,uint dType,uint ownerRate,uint[] lvlRate0)
//	data,err := bm.Inputs().Pack(position)
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	inputs = append(inputs,data...)
	//	contract.RequiredGas(inputs)

	conPtr := vm.NewContract(vm.AccountRef(from),
		vm.AccountRef(contractAddr), big.NewInt(0), uint64(1000000), params.MAN_COIN)
	GroupMethodTest(bm,big.NewInt(time.Now().Unix()),inputs,conPtr,state,t)
}
func withdrawAllMethod(from,contractAddr common.Address,state vm.StateDBManager,t *testing.T){
	bm := &vm.BaseMethod{
		Name:"withdrawAll",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.SstoreSetGas,
	}
	inputs := make([]byte,4)
	ID :=bm.MethodID()
	copy(inputs,ID[:])
	//function createValidatorGroup(address signAcount,uint amount,uint dType,uint ownerRate,uint[] lvlRate0)
//	data,err := bm.Inputs().Pack(value,position)
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	inputs = append(inputs,data...)
	//	contract.RequiredGas(inputs)

	conPtr := vm.NewContract(vm.AccountRef(from),
		vm.AccountRef(contractAddr), big.NewInt(0), uint64(1000000), params.MAN_COIN)
	GroupMethodTest(bm,big.NewInt(time.Now().Unix()),inputs,conPtr,state,t)
}