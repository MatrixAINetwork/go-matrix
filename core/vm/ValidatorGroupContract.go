// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package vm

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"math/big"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/core/vm/validatorGroup"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/accounts/abi"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"errors"
)
var (
	manBase = big.NewInt(1e18)
	rateDecmal = 1.0e9
	rateDecmalBig = big.NewInt(1e9)

	developContractAddress = common.BytesToAddress([]byte{10})
	ValidatorGroupContractAddress = common.BytesToAddress([]byte{20})
	depLvel = []*big.Int{big.NewInt(0),
		new(big.Int).Mul(big.NewInt(10000),manBase),
		new(big.Int).Mul(big.NewInt(100000),manBase)}
	errArguments = errors.New("Input Arguments Error")
	errExpired = errors.New("Validator Group is Expired")
	errNotExist = errors.New("Validator Account is not exist")
	errOwnerInsufficient = errors.New("Owner Deposit Amount is insufficient")
	errNotOwner = errors.New("Tx sender is not Owner")
	errInsufficient = errors.New("insufficient balance for Refund")

)

type MethodInterface interface {
	PrecompiledContract
	MethodName()string
	MethodID()[4]byte
}
type BaseMethod struct {
	Name string
	Abi *abi.ABI
	GasUsed uint64
	run func (input []byte, contract *Contract, evm *EVM) ([]byte, error)
}
func (bm* BaseMethod)Inputs()abi.Arguments{
	return bm.Abi.Methods[bm.Name].Inputs
}
func (bm* BaseMethod)Outputs()abi.Arguments{
	return bm.Abi.Methods[bm.Name].Outputs
}
func (bm* BaseMethod)MethodName()string{
	return bm.Name
}
/*
func (bm* BaseMethod)getFuncID(abi *abi.ABI,funcName string)[4]byte{
	funId := abi.Methods[funcName].Id()
	var ID [4]byte
	copy(ID[:],funId[:4])
	return ID
}
*/
func (bm* BaseMethod)MethodID()[4]byte{
	if len(bm.Name)==0{
		return [4]byte{}
	}
	if method,exist := bm.Abi.Methods[bm.Name];exist{
		funId := method.Id()
		var ID [4]byte
		copy(ID[:],funId[:4])
		return ID
	}
	return [4]byte{}
}
func (bm* BaseMethod)RequiredGas(input []byte)uint64{
	return bm.GasUsed
}
func (bm* BaseMethod)Run(input []byte, contract *Contract, evm *EVM) ([]byte, error){
	return bm.run(input,contract,evm)
}
type BaseContract struct {
	anonymousMethod MethodInterface
//	constructMethod MethodInterface
	methodMap map[[4]byte]MethodInterface
}
func (bc* BaseContract)AddMethod(md MethodInterface) {
	bc.methodMap[md.MethodID()] = md
}
func (bc* BaseContract)GetMethod(input []byte)PrecompiledContract{
	if len(input) == 0 {
		return bc.anonymousMethod
	}
	if len(input)<4{
		return nil
	}
	var ID [4]byte
	copy(ID[:],input[:4])
	return bc.methodMap[ID]
}
func (bc* BaseContract)RequiredGas(input []byte) uint64 {
	method := bc.GetMethod(input)
	if method!= nil{
		return method.RequiredGas(input)
	}
	return params.TxGasContractCreation * 2
}
func (bc* BaseContract)Run(input []byte, contract *Contract, evm *EVM) ([]byte, error) {
	method := bc.GetMethod(input)
	if method!= nil{
		return method.Run(input,contract,evm)
	}
	return nil,errExecutionReverted

}

func EmergeKey(address common.Address,surffix string)common.Hash{
	hashKey := common.Hash{}
	copy(hashKey[:],address[:])
	copy(hashKey[20:],[]byte(surffix))
	return hashKey
}
type ValidatorContractState struct {
	childGroup validatorGroup.AddressSlice
}
func (vc *ValidatorContractState)RemoveEmptyValidatorGroup(address common.Address,state StateDBManager)error{
	err := vc.GetState(ValidatorGroupContractAddress,state)
	if err != nil{
		return err
	}
	vc.childGroup.Remove(address)
	return vc.SetState(ValidatorGroupContractAddress,state)
//	return nil
}
func (vc *ValidatorContractState)Find(address common.Address)bool{
	_,exist := vc.childGroup.Find(address)
	return exist
}
func (vc *ValidatorContractState)Insert(address common.Address){
	vc.childGroup.Insert(address)
}
func (vc *ValidatorContractState)SetState(contractAddress common.Address,state StateDBManager)error {
	data,err := rlp.EncodeToBytes([]common.Address(vc.childGroup))
	if err != nil {
		return err
	}
	state.SetStateByteArray(params.MAN_COIN,contractAddress,EmergeKey(contractAddress,"ValiList"),data)
	return nil
}
func (vc *ValidatorContractState)GetState(contractAdress common.Address,state StateDBManager)error {
	data := state.GetStateByteArray(params.MAN_COIN,contractAdress,EmergeKey(contractAdress,"ValiList"))
	if len(data) == 0 {
		vc.childGroup = validatorGroup.AddressSlice{}
		return nil
	}
	err := rlp.DecodeBytes(data,&vc.childGroup)
	if err != nil {
		return err
	}
	return nil
}
func (vc *ValidatorContractState)GetValidatorGroupInfo(time uint64,state StateDBManager)(map[common.Address]*ValidatorGroupState,error){
	err := vc.GetState(ValidatorGroupContractAddress,state)
	if err != nil{
		return nil,err
	}
	validatorGroupMap := make(map[common.Address]*ValidatorGroupState)
	for _,addr := range vc.childGroup {
		states := NewValidatorGroupState()
		err := states.GetState(addr,time,state)
		if err != nil{
			return nil,err
		}
		validatorGroupMap[addr] = states
	}
	return validatorGroupMap,nil
}
//
type ValidatorGroupContract struct {
	BaseContract
	childs ValidatorContractState
}
func NewTransInterestsInterface()TransInterestsInterface{
	return NewValidatorGroupContract()
}
func NewValidatorGroupContract()*ValidatorGroupContract{
	vg := &ValidatorGroupContract{}
	vg.methodMap = make(map[[4]byte]MethodInterface)
	vg.CreateValidatorGroupMethod()
	return vg
}
func (vg* ValidatorGroupContract)IsPrecompiledContract(address common.Address) bool {
	return vg.childs.Find(address)
}
func (vg* ValidatorGroupContract)CreateValidatorGroupMethod(){
	bm := &BaseMethod{
		Name:"createValidatorGroup",
		Abi:&validatorGroup.ValidatorGroupContractAbi,
		GasUsed:params.TxGasContractCreation+params.CallValueTransferGas*2,
	}
	bm.run = func(input []byte, contract *Contract, evm *EVM) ([]byte, error) {
		data,err := bm.Inputs().UnpackValues(input[4:])
		if err!=nil{
			return nil,err
		}
		err = vg.childs.GetState(contract.Address(),evm.StateDB)
		if err!=nil{
			return nil,err
		}
		nonce := evm.StateDB.GetNonce(evm.Cointyp, contract.CallerAddress)-1
		contractAddr := crypto.CreateAddress(contract.CallerAddress, nonce)
		if vg.IsPrecompiledContract(contractAddr)|| len(evm.StateDB.GetCode(params.MAN_COIN,contractAddr))!=0 {
			return nil,ErrContractAddressCollision
		}
		vg.childs.Insert(contractAddr)
		err = vg.childs.SetState(contract.Address(),evm.StateDB)
		if err!=nil{
			return nil,err
		}
		if !evm.StateDB.Exist(evm.Cointyp, contractAddr) {
			evm.StateDB.CreateAccount(evm.Cointyp, contractAddr)
		}
		if !evm.CanTransfer(evm.StateDB, contract.Address(), contract.value, evm.Cointyp){
			return nil, errors.New("insufficient balance for createValidatorGroup")
		}
		evm.Transfer(evm.StateDB, contract.Address(), contractAddr, contract.value, evm.Cointyp)
		child := NewValidatorGroup()
		err = child.Constructor(contractAddr,
			data[0].(common.Address),
			contract.CallerAddress,
			data[1].(*big.Int),
			data[2].(*big.Int),
			data[3].(*big.Int),
			data[4].([]*big.Int),
			contract,evm)
		if err!=nil{
			return nil,err
		}
		dType := data[2].(*big.Int)
		vg.AddCreateValidatorGroupLog(data[0].(common.Address),contractAddr,dType,contract,evm)
		return contractAddr[:],nil
	}
	vg.AddMethod(bm)
}
func (vg* ValidatorGroupContract)AddCreateValidatorGroupLog(signAccount,newContract common.Address,dType *big.Int,contract *Contract, evm *EVM)error{
	topics := []common.Hash{
		validatorGroup.ValidatorGroupContractAbi.Events["CreateValidatorGroup"].Id(),
		contract.CallerAddress.Hash(),
		signAccount.Hash(),
		newContract.Hash(),
	}
	data,err := validatorGroup.ValidatorGroupContractAbi.Events["CreateValidatorGroup"].Inputs.NonIndexed().Pack(contract.value,dType)
	if err != nil {
		return err
	}
	AddContractLog(topics,data,contract, evm)
	return nil
}
//reward
func (vg* ValidatorGroupContract)TransferRewards(amount *big.Int,time uint64,contractAddress common.Address,state StateDBManager)error{
	err := vg.childs.GetState(ValidatorGroupContractAddress,state)
	if err != nil{
		return err
	}
	if vg.childs.Find(contractAddress){
		return NewValidatorGroup().TransferRewards(amount,time,contractAddress,state)
	}
	return nil
}
func (vg* ValidatorGroupContract)TransferInterests(amount *big.Int,position uint64,time uint64,address common.Address,state StateDBManager)error{
	if position == 0 {
		return vg.TransferCurrentInterests(amount,time,address,state)
	}
	return nil
}
//current Interests
func (vg* ValidatorGroupContract)TransferCurrentInterests(amount *big.Int,time uint64,contractAddress common.Address,state StateDBManager)error{
	err := vg.childs.GetState(ValidatorGroupContractAddress,state)
	if err != nil{
		return err
	}
	if vg.childs.Find(contractAddress){
		return NewValidatorGroup().TransferCurrentInterests(amount,time,contractAddress,state)
	}
	return nil

}
//view
func (vg* ValidatorGroupContract)ListValidatorGroup(){

}
