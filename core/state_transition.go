// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php


package core

import (
	"errors"
	"math"
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/core/txinterface"
	"github.com/matrix/go-matrix/core/types"
	"sync"
	"encoding/json"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)
type StateTransition struct {
	gp         *GasPool
	msg        txinterface.Message
	gas        uint64
	gasPrice   *big.Int
	initialGas uint64
	value      *big.Int
	data       []byte
	state      vm.StateDB
	evm        *vm.EVM
}
type mapHashAmont struct {
	mapHashamont map[common.Hash][]byte
	mu sync.RWMutex
}
var saveMapHashAmont mapHashAmont = mapHashAmont{mapHashamont:make(map[common.Hash][]byte)}
type addrAmont struct {
	addr common.Address
	amont *big.Int
}
// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
//func IntrinsicGas(data []byte, contractCreation, homestead bool) (uint64, error) {
func IntrinsicGas(data []byte) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	gas = params.TxGas
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		if (math.MaxUint64-gas)/params.TxDataNonZeroGas < nz {
			return 0, vm.ErrOutOfGas
		}
		gas += nz * params.TxDataNonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, vm.ErrOutOfGas
		}
		gas += z * params.TxDataZeroGas
	}
	return gas, nil
}
// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(evm *vm.EVM, msg txinterface.Message, gp *GasPool) *StateTransition {
	return &StateTransition{
		gp:       gp,
		evm:      evm,
		msg:      msg,
		gasPrice: msg.GasPrice(),
		value:    msg.Value(),
		data:     msg.Data(),
		state:    evm.StateDB,
	}
}
// to returns the recipient of the message.
func (st *StateTransition) To() common.Address {
	if st.msg == nil || st.msg.To() == nil /* contract creation */ {
		return common.Address{}
	}
	return *st.msg.To()
}

func (st *StateTransition) UseGas(amount uint64) error {
	if st.gas < amount {
		return vm.ErrOutOfGas
	}
	st.gas -= amount

	return nil
}

func (st *StateTransition) BuyGas() error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Gas()), st.gasPrice)
	for _,tAccount := range st.state.GetBalance(st.msg.From()){
		if tAccount.AccountType == common.MainAccount{
			if tAccount.Balance.Cmp(mgval) < 0{
		return errInsufficientBalanceForGas
			}
			break
		}
	}
	if err := st.gp.SubGas(st.msg.Gas()); err != nil {
		return err
	}
	st.gas += st.msg.Gas()

	st.initialGas = st.msg.Gas()
	st.state.SubBalance(common.MainAccount,st.msg.AmontFrom(), mgval)
	return nil
}

func (st *StateTransition) PreCheck() error {
	// Make sure this transaction's nonce is correct.
	if st.msg.CheckNonce() {
		nonce := st.state.GetNonce(st.msg.From())
		if nonce < st.msg.Nonce() {
			return ErrNonceTooHigh
		} else if nonce > st.msg.Nonce() {
			return ErrNonceTooLow
		}
	}
	return st.BuyGas()
}
// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ApplyMessage(evm *vm.EVM, tx txinterface.Message, gp *GasPool) ([]byte, uint64, bool, error) {
	var stsi txinterface.StateTransitioner
	switch tx.TxType() {
	case types.NormalTxIndex:
		stsi = NewStateTransition(evm,tx,gp)
	}
	if stsi == nil{
		log.Error("File state_transition","func AppleMessage","interface is nil")
	}
	return stsi.TransitionDb()
}
func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, failed bool, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	txtype := tx.GetMatrixType()
	if txtype != common.ExtraNormalTxType{
		switch txtype{
		case common.ExtraRevocable:
			return st.CallRevocableNormalTx()
		case common.ExtraRevertTxType:
			return st.CallRevertNormalTx()
		case common.ExtraUnGasTxType:
			return st.CallUnGasNormalTx()
		case common.ExtraTimeTxType:
			return st.CallTimeNormalTx()
		//case common.ExtraEntrustTx:
			//todo
			//tx.Data()

		default:
			log.Info("File state_transition","func Transitiondb","Unknown extra txtype")
			return nil,0,false,ErrTXUnknownType
		}

	}else{
		return st.CallNormalTx()
	}
}
func (st *StateTransition) CallTimeNormalTx()(ret []byte, usedGas uint64, failed bool, err error){
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, errors.New("file state_transition,func CallRevertNormalTx ,from is nil")
	}
	usefrom := tx.AmontFrom()
	if usefrom == addr {
		return nil, 0, false, errors.New("file state_transition,func CallRevertNormalTx ,usefrom is nil")
	}
	//sender := vm.AccountRef(usefrom)
	var (
		vmerr error
	)
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, err
	}
	mapTOAmonts := make([]*addrAmont,0)
	//YY
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, ErrTXCountOverflow
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return nil, 0, false, err
			}
			//0.7+0.3*pow(0.9,(num-1))
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, err
	}
	st.state.SetNonce(from, st.state.GetNonce(from)+1)
	st.state.AddBalance(common.WithdrawAccount,tx.From(), st.value)
	mapTOAmont := &addrAmont{addr:st.To(),amont:st.value}
	mapTOAmonts = append(mapTOAmonts,mapTOAmont)
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			st.state.AddBalance(common.WithdrawAccount,tx.From(), ex.Amount)
			mapTOAmont = &addrAmont{addr:*ex.Recipient,amont:ex.Amount}
			mapTOAmonts = append(mapTOAmonts,mapTOAmont)
			if vmerr != nil {
				break
			}
		}
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, vmerr
		}
	}
	b,marshalerr:=json.Marshal(mapTOAmonts)
	if marshalerr != nil{
		return nil, 0, false,marshalerr
	}
	saveMapHashAmont.mu.Lock()
	saveMapHashAmont.mapHashamont[tx.Hash()] = b
	saveMapHashAmont.mu.Unlock()

	st.state.AddBalance(common.MainAccount,common.TxGasRewardAddress, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice))
	return ret, st.GasUsed(), vmerr != nil, err
}
func (st *StateTransition) CallRevertNormalTx()(ret []byte, usedGas uint64, failed bool, err error){
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	hashlist := make([]common.Hash,0)
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, errors.New("file state_transition,func CallRevertNormalTx ,from is nil")
	}
	usefrom := tx.AmontFrom()
	if usefrom == addr {
		return nil, 0, false, errors.New("file state_transition,func CallRevertNormalTx ,usefrom is nil")
	}
	var (
		vmerr error
	)
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, err
	}

	//YY
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, ErrTXCountOverflow
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return nil, 0, false, err
			}
			//0.7+0.3*pow(0.9,(num-1))
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, err
	}
	st.state.SetNonce(from, st.state.GetNonce(from)+1)
	var hash common.Hash
	hash.SetBytes(tx.Data())
	hashlist = append(hashlist,hash)
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			hash.SetBytes(ex.Payload)
			hashlist = append(hashlist,hash)
			if vmerr != nil {
				break
			}
		}
	}
	costGas := new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)
	st.state.AddBalance(common.MainAccount,common.TxGasRewardAddress, costGas)
	saveMapHashAmont.mu.Lock()
	for _,tmphash := range hashlist{
		if common.EmptyHash(tmphash){
			continue
		}
		b,ok:=saveMapHashAmont.mapHashamont[tmphash]
		if !ok {
			continue
		}
		mapTOAmonts := make([]*addrAmont,0)
		Unmarshalerr:=json.Unmarshal(b,&mapTOAmonts)
		if Unmarshalerr != nil{
			saveMapHashAmont.mu.Unlock()
			return nil, 0, false,Unmarshalerr
		}
		for _,ada := range mapTOAmonts{
			st.state.AddBalance(common.MainAccount,usefrom, ada.amont)
			st.state.SubBalance(common.WithdrawAccount,usefrom, ada.amont)
		}
		delete(saveMapHashAmont.mapHashamont,tmphash)
	}
	saveMapHashAmont.mu.Unlock()
	return ret, st.GasUsed(), vmerr != nil, err
}
/*
 TODO
	1、可撤销交易中存储的数据格式map[hash][]byte 其中[]byte结构为结构体的切片，结构体由to和金额组成
	2、撤销交易（收gas费）会在交易的data中携带可撤销交易的hash，根据此hash找到对应的[]byte解析出结构体，并将每笔金额退回，不收取gas费用
	3、定时执行可撤销交易，同样从map中获取数据解析出结构体按照对应的to给其转账，此时不再收取交易费
*/
func (st *StateTransition) CallRevocableNormalTx()(ret []byte, usedGas uint64, failed bool, err error){
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, errors.New("file state_transition,func CallRevocableNormalTx ,from is nil")
	}
	usefrom := tx.AmontFrom()
	if usefrom == addr {
		return nil, 0, false, errors.New("file state_transition,func CallRevocableNormalTx ,usefrom is nil")
	}
	var (
		vmerr error
	)
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, err
	}
	mapTOAmonts := make([]*addrAmont,0)
	//YY
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, ErrTXCountOverflow
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return nil, 0, false, err
			}
			//0.7+0.3*pow(0.9,(num-1))
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, err
	}
	st.state.SetNonce(from, st.state.GetNonce(from)+1)
	st.state.AddBalance(common.WithdrawAccount,usefrom, st.value)
	st.state.SubBalance(common.MainAccount,usefrom, st.value)
	mapTOAmont := &addrAmont{addr:st.To(),amont:st.value}
	mapTOAmonts = append(mapTOAmonts,mapTOAmont)
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			st.state.AddBalance(common.WithdrawAccount,usefrom, ex.Amount)
			st.state.SubBalance(common.MainAccount,usefrom, ex.Amount)
			mapTOAmont = &addrAmont{addr:*ex.Recipient,amont:ex.Amount}
			mapTOAmonts = append(mapTOAmonts,mapTOAmont)
			if vmerr != nil {
				break
			}
		}
	}
	costGas := new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, vmerr
		}
	}
	b,marshalerr:=json.Marshal(mapTOAmonts)
	if marshalerr != nil{
		return nil, 0, false,marshalerr
	}
	saveMapHashAmont.mu.Lock()
	saveMapHashAmont.mapHashamont[tx.Hash()] = b
	saveMapHashAmont.mu.Unlock()

	st.state.AddBalance(common.MainAccount,common.TxGasRewardAddress, costGas)
	return ret, st.GasUsed(), vmerr != nil, err
}
func (st *StateTransition) CallUnGasNormalTx()(ret []byte, usedGas uint64, failed bool, err error){
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	toaddr := tx.To()
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, errors.New("file state_transition,func CallUnGasNormalTx ,from is nil")
	}
	sender := vm.AccountRef(from)
	var (
		evm = st.evm
		vmerr error
	)
	//YY
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, ErrTXCountOverflow
		}
	}
	st.gas = 0
	if toaddr == nil {//YY
		log.Error("file state_transition","func CallUnGasNormalTx()","to is nil")
		return nil, 0, false, ErrTXToNil
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(tx.From(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			if toaddr == nil {
				log.Error("file state_transition","func CallUnGasNormalTx()","Extro to is nil")
				return nil, 0, false, ErrTXToNil
			} else {
				// Increment the nonce for the next transaction
				ret, st.gas, vmerr = evm.Call(sender, *ex.Recipient, ex.Payload, st.gas, ex.Amount)
			}
			if vmerr != nil {
				break
			}
		}
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, vmerr
		}
	}
	return ret, 0, vmerr != nil, err
}
func (st *StateTransition) CallNormalTx()(ret []byte, usedGas uint64, failed bool, err error){
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	toaddr := tx.To()
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, errors.New("file state_transition,func CallRevocableNormalTx ,from is nil")
	}
	usefrom := tx.AmontFrom()
	if usefrom == addr {
		return nil, 0, false, errors.New("file state_transition,func CallRevocableNormalTx ,usefrom is nil")
	}
	sender := vm.AccountRef(usefrom)
	var (
		evm = st.evm
		vmerr error
	)
	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, err
	}
	//YY
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, ErrTXCountOverflow
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return nil, 0, false, err
			}
			//0.7+0.3*pow(0.9,(num-1))
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, err
	}
	if toaddr == nil {//YY
		ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(from, st.state.GetNonce(from)+1)
		ret, st.gas, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	//YY=========begin===============
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			if toaddr == nil {
				//ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
				ret, _, st.gas, vmerr = evm.Create(sender, ex.Payload, st.gas, ex.Amount)
			} else {
				// Increment the nonce for the next transaction
				ret, st.gas, vmerr = evm.Call(sender, *ex.Recipient, ex.Payload, st.gas, ex.Amount)
			}
			if vmerr != nil {
				break
			}
		}
	}
	//==============end============
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, vmerr
		}
	}
	//st.RefundGas()
	st.state.AddBalance(common.MainAccount,common.TxGasRewardAddress, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice))
	return ret, st.GasUsed(), vmerr != nil, err
}
func (st *StateTransition) RefundGas() {
	// Apply refund counter, capped to half of the used gas.
	refund := st.GasUsed() / 2
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	// Return ETH for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.state.AddBalance(common.MainAccount,st.msg.From(), remaining)

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) GasUsed() uint64 {
	return st.initialGas - st.gas
}
