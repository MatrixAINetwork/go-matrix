// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package contractTest
import (
	"testing"
	"github.com/MatrixAINetwork/go-matrix/core/vm/validatorGroup"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/common"
	"math/big"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"encoding/json"
	"time"
	"github.com/MatrixAINetwork/go-matrix/core"
)

func TestValidatorGroupContract(t *testing.T){
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
	signAccount := common.Address{123,112,111}
	dType := big.NewInt(0)
	ownerRate := big.NewInt(0)
	lvlRate := make([]*big.Int,3)
	lvlRate[0] = big.NewInt(10)
	lvlRate[1] = big.NewInt(20)
	lvlRate[2] = big.NewInt(30)
	data,err := bm.Inputs().Pack(signAccount,dType,ownerRate,lvlRate)
	if err != nil {
		t.Error(err)
	}
	inputs = append(inputs,data...)
	contract.RequiredGas(inputs)
	value := big.NewInt(1e18)
	value.Mul(value,big.NewInt(1e6))
	conPtr := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")),
		vm.AccountRef(common.BytesToAddress([]byte{20})), value, uint64(1000000), params.MAN_COIN)
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{},mdb, state.NewDatabase(mdb))
	env := vm.NewEVM(vm.Context{BlockNumber:big.NewInt(2),Time:big.NewInt(12345)}, statedb, params.TestChainConfig, vm.Config{}, params.MAN_COIN)
	env.CanTransfer = func(db vm.StateDBManager, address common.Address, amount *big.Int, coin string) bool {
		return true
	}
	env.Transfer = func(vm.StateDBManager, common.Address, common.Address, *big.Int, string) { return }
	//	contract.Run(bm.methodID()[:])
	_,err = vm.RunPrecompiledContract(contract,inputs,conPtr,env)
	if err != nil {
		t.Error(err)
	}
	nonce := env.StateDB.GetNonce("MAN",conPtr.CallerAddress)
	env.StateDB.SetNonce("MAN",conPtr.CallerAddress,nonce+1)
	_,err = vm.RunPrecompiledContract(contract,inputs,conPtr,env)
	if err != nil {
		t.Log(err)
	}else {
		t.Error("Test Cannot passed")
	}
}
func TestValidatorGroupContractCurrent(t *testing.T){
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{},mdb, state.NewDatabase(mdb))
	signAccount := common.Address{123,112,111}
	A0Account := common.Address{223,212,211}
	for i:=0;i<100;i++{
		signAccount[4] = byte(i)
		A0Account[4] = byte(i)
		err := createValidatorGroupContract(statedb,A0Account,signAccount,big.NewInt(0),big.NewInt(10),[]*big.Int{big.NewInt(10),big.NewInt(20),big.NewInt(30)})
		if err != nil{
			t.Error(err)
		}
	}
	vcStates := &vm.ValidatorContractState{}
	valiMap,err := vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	if err != nil {
		t.Error(err)
	}
	if len(valiMap) != 100 {
		t.Error("ValidatorGroup create Error")
	}
	data,_ := json.Marshal(valiMap)
	t.Log(string(data))
}
func TestValidatorGroupContractAddCurrent(t *testing.T){
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{},mdb, state.NewDatabase(mdb))
	signAccount := common.Address{123,112,111}
	A0Account := common.Address{223,212,211}
	for i:=0;i<100;i++{
		signAccount[4] = byte(i)
//		A0Account[4] = byte(i)
		err := createValidatorGroupContract(statedb,A0Account,signAccount,big.NewInt(0),big.NewInt(10),[]*big.Int{big.NewInt(10),big.NewInt(20),big.NewInt(30)})
		if err != nil{
			t.Error(err)
		}
	}
	vcStates := &vm.ValidatorContractState{}
	valiMap,err := vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	if err != nil {
		t.Error(err)
	}
	if len(valiMap) != 100 {
		t.Error("ValidatorGroup create Error")
	}
	data,_ := json.Marshal(valiMap)
	t.Log(string(data))
}
func TestValidatorGroupContractPos(t *testing.T){
	mdb := mandb.NewMemDatabase()
	statedb, _ := state.NewStateDBManage([]common.CoinRoot{},mdb, state.NewDatabase(mdb))
	signAccount := common.Address{123,112,111}
	A0Account := common.Address{223,212,211}
	for i:=0;i<100;i++{
		signAccount[4] = byte(i)
		A0Account[4] = byte(i)
		err := createValidatorGroupContract(statedb,A0Account,signAccount,big.NewInt(1),big.NewInt(10),[]*big.Int{big.NewInt(10),big.NewInt(20),big.NewInt(30)})
		if err != nil{
			t.Error(err)
		}
	}
	vcStates := &vm.ValidatorContractState{}
	valiMap,err := vcStates.GetValidatorGroupInfo(uint64(time.Now().Unix()),statedb)
	if err != nil {
		t.Error(err)
	}
	if len(valiMap) != 100 {
		t.Error("ValidatorGroup create Error")
	}
	data,_ := json.Marshal(valiMap)
	t.Log(string(data))
}
func createValidatorGroupContract(stateDb vm.StateDBManager,A0,A1 common.Address,dType,ownerRate *big.Int,lvlRate []*big.Int)error{
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
	data,err := bm.Inputs().Pack(A1,dType,ownerRate,lvlRate)
	if err != nil {
		return err
	}
	inputs = append(inputs,data...)
	contract.RequiredGas(inputs)
	value := big.NewInt(1e18)
	value.Mul(value,big.NewInt(1e6))
	conPtr := vm.NewContract(vm.AccountRef(A0),
		vm.AccountRef(vm.ValidatorGroupContractAddress), value, uint64(1000000), params.MAN_COIN)
	env := vm.NewEVM(vm.Context{BlockNumber:big.NewInt(2),Time:big.NewInt(12345)}, stateDb, params.TestChainConfig, vm.Config{}, params.MAN_COIN)
	env.CanTransfer = core.CanTransfer
	env.Transfer = core.Transfer
	//	contract.Run(bm.methodID()[:])
	_,err = vm.RunPrecompiledContract(contract,inputs,conPtr,env)
	if err != nil {
		return err
	}
	nonce := env.StateDB.GetNonce("MAN",conPtr.CallerAddress)
	env.StateDB.SetNonce("MAN",conPtr.CallerAddress,nonce+1)
	return nil
}