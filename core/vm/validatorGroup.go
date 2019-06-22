// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package vm

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm/validatorGroup"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
	"math"
	"math/big"
	"errors"
	"github.com/MatrixAINetwork/go-matrix/reward/depositcfg"
)

func NewValidatorGroup() *ValidatorGroup {
	child := &ValidatorGroup{
		constate: *NewValidatorGroupState(),
	}
	//	child.constate.depInfo = NewValidatorGroupState
	child.methodMap = make(map[[4]byte]MethodInterface)
	child.AnonymousMethod()
	//	child.TransferOwnershipMethod()
	child.SetSignAccountMethod()
	child.WithdrawAllMethod()
	child.AddDepositMethod()
	child.WithdrawMethod()
	child.RefundMethod()
	child.GetRewardMethod()
	return child
}

type ValidatorGroup struct {
	BaseContract
	constate ValidatorGroupState
	//	reward RewardRate
	//	owner common.Address
	//	ownerAmount *big.Int
	//	validatorMap map[common.Address]validatorInfo

}

func (vg *ValidatorGroup) getDepositInfo() []validatorGroup.ValidatorInfo {
	return vg.constate.ValidatorMap
}
func (vg *ValidatorGroup) TransferCurrentInterests(amount *big.Int, time uint64, contractAddress common.Address, state StateDBManager) error {
	err := vg.constate.GetState(contractAddress, time, state)
	if err != nil {
		return err
	}
	err = vg.constate.DistributeCurrentInterests(amount)
	if err != nil {
		return err
	}
	err = vg.constate.SetState(contractAddress, state)
	if err != nil {
		return err
	}
	return nil
}
func (vg *ValidatorGroup) TransferRewards(amount *big.Int, time uint64, contractAddress common.Address, state StateDBManager) error {
	err := vg.constate.GetState(contractAddress, time, state)
	if err != nil {
		return err
	}
	err = vg.constate.DistributeRewards(amount)
	if err != nil {
		return err
	}
	err = vg.constate.SetState(contractAddress, state)
	if err != nil {
		return err
	}
	return nil
}
func (vg *ValidatorGroup) TransferMan(contract *Contract, evm *EVM) error {
	value := contract.Value()
	return vg.TransferRewards(value, evm.Time.Uint64(), contract.Address(), evm.StateDB)
}
func (vg *ValidatorGroup) Constructor(conAddr, signAddr, Owner common.Address, dType, OwnerRate,nodeRate *big.Int, lvlRate []*big.Int, contract *Contract, evm *EVM) error {
	if !evm.StateDB.Exist(evm.Cointyp, conAddr) {
		evm.StateDB.CreateAccount(evm.Cointyp, conAddr)
	}
	newCon := NewContract(contract, AccountRef(conAddr), contract.value, contract.Gas, evm.Cointyp)
	vg.constate.OwnerInfo.Owner = Owner
	err := vg.constate.SetRewardRate(OwnerRate,nodeRate, lvlRate)
	if err != nil {
		return err
	}
	valInfo := validatorGroup.NewValidatorInfo(Owner)
	_, err = vg.addDeposit(valInfo, signAddr, dType.Uint64(), newCon, evm)
	vg.constate.ValidatorMap.Insert(*valInfo)
	if err != nil {
		return err
	}
	err = vg.constate.SetState(newCon.Address(), evm.StateDB)
	if err != nil {
		return err
	}
	return nil
}

/*
//method Constructor
func (vg* ValidatorGroup)ConstructMethod(){
	bm := &BaseMethod{
		Name:"",
		Abi:&validatorGroup.ValidatorGroupAbi,
		GasUsed:params.TxGasContractCreation,
	}
	bm.run = func(input []byte, contract *Contract, evm *EVM) ([]byte, error) {
		return nil,vg.transferMan(contract,evm)
	}
	vg.anonymousMethod = bm
}
*/
//Anonymous Method
func (vg *ValidatorGroup) AnonymousMethod() {
	bm := &BaseMethod{
		Name:    "",
		Abi:     &validatorGroup.ValidatorGroupAbi,
		GasUsed: params.TxGasContractCreation,
	}
	bm.run = func(input []byte, contract *Contract, evm *EVM) ([]byte, error) {
		return nil, vg.TransferMan(contract, evm)
	}
	vg.anonymousMethod = bm
}

//TransferOwnership
func (vg *ValidatorGroup) TransferOwnershipMethod() {
	bm := &BaseMethod{
		Name:    "transferOwnership",
		Abi:     &validatorGroup.ValidatorGroupAbi,
		GasUsed: params.SstoreSetGas,
	}
	bm.run = func(input []byte, contract *Contract, evm *EVM) ([]byte, error) {
		err := vg.constate.GetOwner(contract.Address(), evm.StateDB)
		if err != nil {
			return nil, err
		}
		if contract.CallerAddress != vg.constate.OwnerInfo.Owner {
			return nil, errNotOwner
		}
		var addr common.Address
		err = bm.Inputs().Unpack(&addr, input[4:])
		if err != nil || len(addr) != 20 {
			return nil, errArguments
		}
		vg.constate.OwnerInfo.Owner = addr
		err = vg.constate.SetOwner(contract.Address(), evm.StateDB)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	vg.AddMethod(bm)
}

//SetSignAccount
func (vg *ValidatorGroup) SetSignAccountMethod() {
	bm := &BaseMethod{
		Name:    "setSignAccount",
		Abi:     &validatorGroup.ValidatorGroupAbi,
		GasUsed: params.SstoreSetGas,
	}
	bm.run = func(input []byte, contract *Contract, evm *EVM) ([]byte, error) {
		err := vg.constate.GetState(contract.Address(), evm.Time.Uint64(), evm.StateDB)
		if err != nil {
			return nil, err
		}
		if contract.CallerAddress != vg.constate.OwnerInfo.Owner {
			return nil, errNotOwner
		}
		if vg.constate.OwnerInfo.WithdrawAllTime > 0 {
			return nil, errExpired
		}
		var addr common.Address
		err = bm.Inputs().Unpack(&addr, input[4:])
		if err != nil || len(addr) != 20 {
			return nil, err
		}
		return vg.addDeposit(&validatorGroup.ValidatorInfo{}, addr, 0, contract, evm)
	}
	vg.AddMethod(bm)
}
func (vg *ValidatorGroup) WithdrawAllMethod() {
	bm := &BaseMethod{
		Name:    "withdrawAll",
		Abi:     &validatorGroup.ValidatorGroupAbi,
		GasUsed: params.SstoreSetGas + params.CallValueTransferGas*2,
	}
	bm.run = func(input []byte, contract *Contract, evm *EVM) ([]byte, error) {
		err := vg.constate.GetState(contract.Address(), evm.Time.Uint64(), evm.StateDB)
		if err != nil {
			return nil, err
		}
		if vg.constate.OwnerInfo.Owner != contract.CallerAddress {
			return nil, errNotOwner
		}
		if vg.constate.OwnerInfo.WithdrawAllTime > 0 {
			return nil, errExpired
		}
		vg.constate.OwnerInfo.WithdrawAllTime = evm.Time.Uint64()

		for i := 0; i < len(vg.constate.ValidatorMap); i++ {
			info := &vg.constate.ValidatorMap[i]
			amount := new(big.Int).Add(info.Current.Amount,info.Current.Interest)
			if amount.Cmp(depositcfg.CruWithDrawAmountMin) >= 0 {
				ret, err := vg.withdrawCurrent(info, info.Current.Amount, contract, evm)
				if err != nil {
					return ret, err
				}
			}
			for j := 0; j < len(info.Positions); j++ {
				dep := &info.Positions[j]
				if dep.EndTime > 0 {
					continue
				}
				ret, err := vg.withdraw(dep.Amount, dep.Position, contract, evm)
				if err != nil {
					return ret, err
				}
			}
		}
		//		_,err = vg.withdraw(big.NewInt(0),uint64(0),contract,evm)
		if err != nil {
			return nil, err
		}
		err = vg.constate.SetState(contract.Address(), evm.StateDB)
		return nil, err
	}
	vg.AddMethod(bm)
}
func (vg *ValidatorGroup) GetRewardMethod() {
	bm := &BaseMethod{
		Name:    "getReward",
		Abi:     &validatorGroup.ValidatorGroupAbi,
		GasUsed: params.SstoreSetGas + params.CallValueTransferGas*2,
	}
	bm.run = func(input []byte, contract *Contract, evm *EVM) ([]byte, error) {
		err := vg.constate.GetState(contract.Address(), evm.Time.Uint64(), evm.StateDB)
		if err != nil {
			return nil, err
		}
		index, exist := vg.constate.ValidatorMap.Find(contract.CallerAddress)
		if !exist {
			if vg.constate.OwnerInfo.Owner == contract.CallerAddress && len(vg.constate.ValidatorMap) == 0 {
				balance := evm.StateDB.GetBalance(params.MAN_COIN, contract.Address())
				amount := balance[common.MainAccount].Balance
				if amount.Sign()>0{
					if !evm.CanTransfer(evm.StateDB, contract.Address(), amount, evm.Cointyp){
						return nil, errors.New("insufficient balance for getReward")
					}
					evm.Transfer(evm.StateDB, contract.Address(), contract.CallerAddress, amount, evm.Cointyp)
				}
				err = vg.constate.CheckValidatorInfo(nil, contract.Address(), evm.StateDB)
				if err != nil {
					return nil, err
				}
				return nil, vg.constate.SetState(contract.Address(), evm.StateDB)
			} else {
				return nil, errArguments
			}
		}
		valInfo := &vg.constate.ValidatorMap[index]
		amount := valInfo.Reward
		valInfo.Reward = big.NewInt(0)
		if amount.Sign()>0 {
			if !evm.CanTransfer(evm.StateDB, contract.Address(), amount, evm.Cointyp){
				return nil, errors.New("insufficient balance for getReward")
			}
			evm.Transfer(evm.StateDB, contract.Address(), contract.CallerAddress, amount, evm.Cointyp)
		}
		err = vg.constate.CheckValidatorInfo(valInfo, contract.Address(), evm.StateDB)
		if err != nil {
			return nil, err
		}
		return nil, vg.constate.SetState(contract.Address(), evm.StateDB)
	}
	vg.AddMethod(bm)
}
func (vg *ValidatorGroup) addDeposit(valiInfo *validatorGroup.ValidatorInfo, signAddr common.Address, dType uint64, contract *Contract, evm *EVM) ([]byte, error) {
	input10 := depositAbi_v2.Methods["valiDeposit"].Id()
	args, err := depositAbi_v2.Methods["valiDeposit"].Inputs.Pack(signAddr, new(big.Int).SetUint64(dType))
	if err != nil {
		return nil, err
	}
	input10 = append(input10, args...)
	ret, _, _, err := evm.Call(contract, developContractAddress, input10, contract.Gas, contract.value)
	if err == nil {
		vg.AddDepositLog(new(big.Int).SetUint64(dType), contract, evm)
	} else {
		log.ERROR("ValidatorGroup", "addDeposit evm.Call err", err)
		return ret, err
	}
	if dType == 0 {
		if valiInfo.Current.Amount == nil {
			valiInfo.Current.Amount = big.NewInt(0)
		}
		valiInfo.Current.Amount.Add(valiInfo.Current.Amount, contract.value)
	} else if contract.value.Sign() > 0 {
		allDepInfo := vg.constate.GetAllDepositInfo(contract.Address(), evm.StateDB)
		if len(allDepInfo.Dpstmsg) < 2 {
			return nil, errOverflow
		}
		latest := allDepInfo.Dpstmsg[len(allDepInfo.Dpstmsg)-1]
		valiInfo.Positions = append(valiInfo.Positions, validatorGroup.DepositPos{dType, latest.Position, latest.DepositAmount, 0})
	}
	return ret, err
}
func (vg* ValidatorGroup)withdrawCurrent(valiInfo *validatorGroup.ValidatorInfo,amount *big.Int,contract *Contract, evm *EVM)([]byte,error) {
	if valiInfo.Current.Amount.Cmp(amount)<0  {
		return nil, errors.New("insufficient balance for withdraw")
	}
	interest := big.NewInt(0)
	if valiInfo.Current.Amount.Sign()>0 {
		interest = new(big.Int).Mul(amount,valiInfo.Current.Interest)
		interest.Div(interest,valiInfo.Current.Amount)
	}else{
		if valiInfo.Current.Interest.Sign()<0{
			return nil, errors.New("insufficient balance for withdraw")
		}
		interest.Set(valiInfo.Current.Interest)
	}
	currentAmount := new(big.Int).Add(amount, interest)
	valiInfo.Current.Amount.Sub(valiInfo.Current.Amount, amount)
	valiInfo.Current.Interest.Sub(valiInfo.Current.Interest, interest)
	input10 := depositAbi_v2.Methods["withdraw"].Id()
	if currentAmount.Sign() == 0 {
		return nil, nil
	}
	//	currentAmount := vg.constate.CalCurrentWithDrawAmount(amount)
	args, err := depositAbi_v2.Methods["withdraw"].Inputs.Pack(big.NewInt(0), currentAmount)
	if err != nil {
		return nil, err
	}
	input10 = append(input10, args...)
	ret, _, _, err := evm.Call(contract, developContractAddress, input10, contract.Gas, contract.value)
	if err == nil {
		vg.AddWithdrawLog(currentAmount, big.NewInt(0), contract, evm)
	} else {
		log.ERROR("ValidatorGroup", "withdrawCurrent evm.Call err", err)
		return ret, err
	}
	allDepInfo := vg.constate.GetAllDepositInfo(contract.Address(), evm.StateDB)
	if len(allDepInfo.Dpstmsg) < 1 {
		return nil, errOverflow
	}
	current := allDepInfo.Dpstmsg[0]
	if len(current.WithDrawInfolist) == 0 {
		return nil, errOverflow
	}
	valiInfo.Current.WithdrawList = append(valiInfo.Current.WithdrawList, current.WithDrawInfolist[len(current.WithDrawInfolist)-1])
	return ret, err
}
func (vg *ValidatorGroup) withdraw(amount *big.Int, position uint64, contract *Contract, evm *EVM) ([]byte, error) {
	input10 := depositAbi_v2.Methods["withdraw"].Id()
	args, err := depositAbi_v2.Methods["withdraw"].Inputs.Pack(new(big.Int).SetUint64(position), amount)
	if err != nil {
		return nil, err
	}
	input10 = append(input10, args...)
	ret, _, _, err := evm.Call(contract, developContractAddress, input10, contract.Gas, contract.value)
	if err == nil {
		vg.AddWithdrawLog(amount, new(big.Int).SetUint64(position), contract, evm)
	}
	return ret, err
}
func (vg *ValidatorGroup) refundCurrent(contract *Contract, evm *EVM) ([]byte, error) {
	input10 := depositAbi_v2.Methods["refund"].Id()
	args, err := depositAbi_v2.Methods["refund"].Inputs.Pack(new(big.Int).SetUint64(0))
	if err != nil {
		return nil, err
	}
	input10 = append(input10, args...)
	ret, _, _, err := evm.Call(contract, developContractAddress, input10, contract.Gas, contract.value)
	if err == nil {
		vg.AddRefundLog(new(big.Int).SetUint64(0), contract, evm)
	} else {
		log.ERROR("ValidatorGroup", "refundCurrent evm.Call err", err)
		return ret, err
	}
	allDepInfo := vg.constate.GetAllDepositInfo(contract.Address(), evm.StateDB)
	reFundTime := uint64(math.MaxUint64)
	if allDepInfo != nil && len(allDepInfo.Dpstmsg) > 0 && len(allDepInfo.Dpstmsg[0].WithDrawInfolist) > 0 {
		reFundTime = allDepInfo.Dpstmsg[0].WithDrawInfolist[0].WithDrawTime
	}
	refundAmount := big.NewInt(0)
	index := -1
	for i := 0; i < len(vg.constate.ValidatorMap); i++ {
		valiInfo := &vg.constate.ValidatorMap[i]
		if valiInfo.Address == contract.CallerAddress {
			index = i
			for j := 0; j < len(valiInfo.Current.WithdrawList); j++ {
				if valiInfo.Current.WithdrawList[j].WithDrawTime < reFundTime {
					refundAmount.Add(refundAmount, valiInfo.Current.WithdrawList[j].WithDrawAmount)
					valiInfo.Current.WithdrawList = append(valiInfo.Current.WithdrawList[:j], valiInfo.Current.WithdrawList[j+1:]...)
					j--
				}
			}
		} else {
			for j := 0; j < len(valiInfo.Current.WithdrawList); j++ {
				if valiInfo.Current.WithdrawList[j].WithDrawTime < reFundTime {
					valiInfo.Reward.Add(valiInfo.Reward, valiInfo.Current.WithdrawList[j].WithDrawAmount)
					valiInfo.Current.WithdrawList = append(valiInfo.Current.WithdrawList[:j], valiInfo.Current.WithdrawList[j+1:]...)
					j--
				}
			}
		}
	}
	if index >= 0 {
		valiInfo := &vg.constate.ValidatorMap[index]
		if len(valiInfo.Current.WithdrawList) == 0 {
			err = vg.constate.CheckValidatorInfo(valiInfo, contract.Address(), evm.StateDB)
			if err != nil {
				return nil, err
			}
		}
	}
	if !evm.CanTransfer(evm.StateDB, contract.Address(), refundAmount, evm.Cointyp){
		return nil, errors.New("insufficient balance for withdraw")
	}
	evm.Transfer(evm.StateDB, contract.Address(), contract.CallerAddress, refundAmount, evm.Cointyp)
	return ret, err
}
func (vg *ValidatorGroup) refund(valiInfo *validatorGroup.ValidatorInfo, index int, contract *Contract, evm *EVM) ([]byte, error) {
	position := valiInfo.Positions[index].Position
	input10 := depositAbi_v2.Methods["refund"].Id()
	args, err := depositAbi_v2.Methods["refund"].Inputs.Pack(new(big.Int).SetUint64(position))
	if err != nil {
		return nil, err
	}
	input10 = append(input10, args...)
	refundAmount := valiInfo.Positions[index].Amount
	ret, _, _, err := evm.Call(contract,developContractAddress,input10,contract.Gas,contract.value)
	if err == nil{
		if !evm.CanTransfer(evm.StateDB, contract.Address(), refundAmount, evm.Cointyp){
			return nil, errors.New("insufficient balance for Refund")
		}
		evm.Transfer(evm.StateDB, contract.Address(), contract.CallerAddress, refundAmount, evm.Cointyp)
		valiInfo.Positions = append(valiInfo.Positions[:index], valiInfo.Positions[index+1:]...)
		err = vg.constate.CheckValidatorInfo(valiInfo, contract.Address(), evm.StateDB)
		if err != nil {
			return nil, err
		}
		vg.AddRefundLog(new(big.Int).SetUint64(position), contract, evm)
	}
	return ret, err
}
func (vg *ValidatorGroup) AddDepositMethod() {
	bm := &BaseMethod{
		Name:    "addDeposit",
		Abi:     &validatorGroup.ValidatorGroupAbi,
		GasUsed: params.SstoreSetGas + params.CallNewAccountGas + params.CallValueTransferGas*2,
	}
	bm.run = func(input []byte, contract *Contract, evm *EVM) ([]byte, error) {
		err := vg.constate.GetState(contract.Address(), evm.Time.Uint64(), evm.StateDB)
		if err != nil {
			return nil, err
		}
		if vg.constate.OwnerInfo.WithdrawAllTime > 0 {
			return nil, errExpired
		}
		data, err := bm.Inputs().UnpackValues(input[4:])
		if err != nil {
			return nil, err
		}
		index, exist := vg.constate.ValidatorMap.Find(contract.CallerAddress)
		if !exist {
			valInfo := validatorGroup.NewValidatorInfo(contract.CallerAddress)
			vg.constate.ValidatorMap.Insert(*valInfo)
		}
		ret, err := vg.addDeposit(&vg.constate.ValidatorMap[index], vg.constate.OwnerInfo.SignAddress, data[0].(*big.Int).Uint64(), contract, evm)
		if err != nil {
			return nil, err
		}
		return ret, vg.constate.SetState(contract.Address(), evm.StateDB)
	}
	vg.AddMethod(bm)
}
func (vg *ValidatorGroup) WithdrawMethod() {
	bm := &BaseMethod{
		Name:    "withdraw",
		Abi:     &validatorGroup.ValidatorGroupAbi,
		GasUsed: params.SstoreSetGas + params.CallValueTransferGas*2,
	}
	bm.run = func(input []byte, contract *Contract, evm *EVM) ([]byte, error) {
		err := vg.constate.GetState(contract.Address(), evm.Time.Uint64(), evm.StateDB)
		if err != nil {
			return nil, err
		}
		if vg.constate.OwnerInfo.WithdrawAllTime > 0 {
			return nil, errExpired
		}
		index, exist := vg.constate.ValidatorMap.Find(contract.CallerAddress)
		if !exist {
			return nil, errArguments
		}
		data, err := bm.Inputs().UnpackValues(input[4:])
		if err != nil {
			return nil, err
		}
		position := data[1].(*big.Int).Uint64()
		amount := data[0].(*big.Int)
		valInfo := &vg.constate.ValidatorMap[index]
		if contract.CallerAddress == vg.constate.OwnerInfo.Owner {
			allAmount := big.NewInt(0)
			allAmount.Add(allAmount, valInfo.Current.Amount)
			if position == 0 {
				allAmount.Sub(allAmount, amount)
			}
			for i:=0;i<len(valInfo.Positions);i++{
				if valInfo.Positions[i].EndTime == 0 && valInfo.Positions[i].Position != position {
					allAmount.Add(allAmount,valInfo.Positions[i].Amount)
				}
			}
			if allAmount.Cmp(validatorThreshold) < 0 {
				return nil, errOwnerInsufficient
			}
		}
		var ret []byte
		if position == 0 {
			ret, err = vg.withdrawCurrent(valInfo, amount, contract, evm)
		} else {
			success := false
			for i := 0; i < len(valInfo.Positions); i++ {
				if valInfo.Positions[i].Position == position {
					if valInfo.Positions[i].EndTime > 0 {
						return nil, errArguments
					}
					ret, err = vg.withdraw(big.NewInt(0), valInfo.Positions[i].Position, contract, evm)
					success = true
					break
				}
			}
			if !success {
				return nil, errArguments
			}
		}
		if err == nil {
			err = vg.constate.SetState(contract.Address(), evm.StateDB)
		}
		return ret, err
	}
	vg.AddMethod(bm)
}
func (vg *ValidatorGroup) RefundMethod() {
	bm := &BaseMethod{
		Name:    "refund",
		Abi:     &validatorGroup.ValidatorGroupAbi,
		GasUsed: params.SstoreSetGas + params.CallValueTransferGas*2,
	}
	bm.run = func(input []byte, contract *Contract, evm *EVM) ([]byte, error) {
		err := vg.constate.GetState(contract.Address(), evm.Time.Uint64(), evm.StateDB)
		if err != nil {
			return nil, err
		}
		index, exist := vg.constate.ValidatorMap.Find(contract.CallerAddress)
		if !exist {
			return nil, errNotExist
		}
		valInfo := &vg.constate.ValidatorMap[index]
		data, err := bm.Inputs().UnpackValues(input[4:])
		if err != nil {
			return nil, err
		}
		position := data[0].(*big.Int).Uint64()
		var ret []byte
		if position == 0 {
			ret, err = vg.refundCurrent(contract, evm)
			if err != nil {
				return ret, err
			}
		} else {

			success := false
			for i := 0; i < len(valInfo.Positions); i++ {
				if valInfo.Positions[i].Position == position {
					if valInfo.Positions[i].EndTime == 0 {
						return nil, errArguments
					}
					ret, err = vg.refund(valInfo, i, contract, evm)
					success = true
					break
				}
			}
			if !success {
				return nil, errArguments
			}
		}
		if err == nil {
			err = vg.constate.SetState(contract.Address(), evm.StateDB)
		}
		return ret, err
	}
	vg.AddMethod(bm)
}

/*
//view
func (vg* ValidatorGroup) GetDepositList(input []byte,contract *Contract, evm *EVM)error {
}
func (vg* ValidatorGroup) GetDepositInfo(input []byte,contract *Contract, evm *EVM)error {
}
*/
func (vg *ValidatorGroup) AddRefundLog(dtype *big.Int, contract *Contract, evm *EVM) error {
	topics := []common.Hash{
		validatorGroup.ValidatorGroupAbi.Events["Refund"].Id(),
		contract.CallerAddress.Hash(),
	}
	data, err := validatorGroup.ValidatorGroupAbi.Events["Refund"].Inputs.NonIndexed().Pack(dtype)
	if err != nil {
		return err
	}
	AddContractLog(topics, data, contract, evm)
	return nil
}
func (vg *ValidatorGroup) AddWithdrawLog(amount, dtype *big.Int, contract *Contract, evm *EVM) error {
	topics := []common.Hash{
		validatorGroup.ValidatorGroupAbi.Events["Withdraw"].Id(),
		contract.CallerAddress.Hash(),
	}
	data, err := validatorGroup.ValidatorGroupAbi.Events["Withdraw"].Inputs.NonIndexed().Pack(amount, dtype)
	if err != nil {
		return err
	}
	AddContractLog(topics, data, contract, evm)
	return nil
}
func (vg *ValidatorGroup) AddDepositLog(dType *big.Int, contract *Contract, evm *EVM) error {
	topics := []common.Hash{
		validatorGroup.ValidatorGroupAbi.Events["AddDeposit"].Id(),
		contract.CallerAddress.Hash(),
	}
	data, err := validatorGroup.ValidatorGroupAbi.Events["AddDeposit"].Inputs.NonIndexed().Pack(contract.value, dType)
	if err != nil {
		return err
	}
	AddContractLog(topics, data, contract, evm)
	return nil
}
func AddContractLog(topics []common.Hash, data []byte, contract *Contract, evm *EVM) {
	evm.StateDB.AddLog(evm.Cointyp, contract.CallerAddress, &types.Log{
		Address: contract.Address(),
		Topics:  topics,
		Data:    data,
		// This is a non-consensus field, but assigned here because
		// core/state doesn't know the current block number.
		BlockNumber: evm.BlockNumber.Uint64(),
	})
}
