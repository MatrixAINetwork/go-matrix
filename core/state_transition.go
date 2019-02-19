// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"encoding/json"
	"errors"
	"math"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/base58"
	"github.com/MatrixAINetwork/go-matrix/common"
	matrixstate "github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/supertxsstate"
	"github.com/MatrixAINetwork/go-matrix/core/txinterface"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
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
	gasprice, err := matrixstate.GetTxpoolGasLimit(evm.StateDB)
	if err != nil {
		//return errors.New("get txpool gasPrice err")
	}
	return &StateTransition{
		gp:       gp,
		evm:      evm,
		msg:      msg,
		gasPrice: gasprice,
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
	for _, tAccount := range st.state.GetBalance(st.msg.From()) {
		if tAccount.AccountType == common.MainAccount {
			if tAccount.Balance.Cmp(mgval) < 0 {
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
	st.state.SubBalance(common.MainAccount, st.msg.AmontFrom(), mgval)
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
		stsi = NewStateTransition(evm, tx, gp)
	}
	if stsi == nil {
		log.Error("state_transition", "AppleMessage", "interface is nil")
	}
	return stsi.TransitionDb()
}
func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, failed bool, err error) {
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	txtype := tx.GetMatrixType()
	if txtype != common.ExtraNormalTxType && txtype != common.ExtraAItxType {
		switch txtype {
		case common.ExtraRevocable:
			return st.CallRevocableNormalTx()
		case common.ExtraRevertTxType:
			return st.CallRevertNormalTx()
		case common.ExtraUnGasTxType:
			return st.CallUnGasNormalTx()
		case common.ExtraTimeTxType:
			return st.CallTimeNormalTx()
		case common.ExtraAuthTx:
			log.INFO("授权交易", "交易类型", txtype)
			return st.CallAuthTx()
		case common.ExtraCancelEntrust:
			log.INFO("取消委托", "交易类型", txtype)
			return st.CallCancelAuthTx()
		case common.ExtraSuperTxType:
			return st.CallSuperTx()
		//case common.ExtraCreatCurrency:
		//	return st.CallCreatCurrencyTx()
		default:
			log.Info("state transition unknown extra txtype")
			return nil, 0, false, ErrTXUnknownType
		}

	} else {
		return st.CallNormalTx()
	}
}
func (st *StateTransition) CallTimeNormalTx() (ret []byte, usedGas uint64, failed bool, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, errors.New("CallTimeNormalTx from is nil")
	}
	//usefrom := tx.AmontFrom()
	usefrom := tx.From()
	if usefrom == addr {
		return nil, 0, false, errors.New("CallTimeNormalTx usefrom is nil")
	}
	var (
		vmerr error
	)
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, err
	}
	mapTOAmonts := make([]common.AddrAmont, 0)
	//
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
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, err
	}
	st.state.SetNonce(from, st.state.GetNonce(from)+1)
	st.RefundGas()
	st.state.AddBalance(common.WithdrawAccount, usefrom, st.value)
	st.state.SubBalance(common.MainAccount, usefrom, st.value)
	mapTOAmont := common.AddrAmont{Addr: st.To(), Amont: st.value}
	mapTOAmonts = append(mapTOAmonts, mapTOAmont)
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			st.state.AddBalance(common.WithdrawAccount, usefrom, ex.Amount)
			st.state.SubBalance(common.MainAccount, usefrom, ex.Amount)
			mapTOAmont = common.AddrAmont{Addr: *ex.Recipient, Amont: ex.Amount}
			mapTOAmonts = append(mapTOAmonts, mapTOAmont)
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
	rt := new(common.RecorbleTx)
	rt.From = tx.From()
	rt.Tim = tx.GetCreateTime()
	rt.Typ = tx.GetMatrixType()
	rt.Adam = append(rt.Adam, mapTOAmonts...)
	b, marshalerr := json.Marshal(rt)
	if marshalerr != nil {
		return nil, 0, false, marshalerr
	}
	txHash := tx.Hash()
	mapHashamont := make(map[common.Hash][]byte)
	mapHashamont[txHash] = b
	st.state.SaveTx(tx.GetMatrixType(), rt.Tim, mapHashamont)
	st.state.AddBalance(common.MainAccount, common.TxGasRewardAddress, costGas)
	return ret, st.GasUsed(), vmerr != nil, err
}
func (st *StateTransition) CallRevertNormalTx() (ret []byte, usedGas uint64, failed bool, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	hashlist := make([]common.Hash, 0)
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, errors.New("CallRevertNormalTx from is nil")
	}
	var (
		vmerr error
	)
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, err
	}

	//
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
	hash = common.BytesToHash(tx.Data())
	hashlist = append(hashlist, hash)
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			hash = common.BytesToHash(ex.Payload)
			hashlist = append(hashlist, hash)
			if vmerr != nil {
				break
			}
		}
	}
	costGas := new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)
	st.RefundGas()
	st.state.AddBalance(common.MainAccount, common.TxGasRewardAddress, costGas)
	delval := make(map[uint32][]common.Hash)
	for _, tmphash := range hashlist {
		if common.EmptyHash(tmphash) {
			continue
		}
		b := st.state.GetMatrixData(tmphash)
		if b == nil {
			log.Error("CallRevertNormalTx not found tx hash,maybe the transaction has lasted more than 24 hours")
			continue
		}
		var rt common.RecorbleTx
		errRT := json.Unmarshal(b, &rt)
		if errRT != nil {
			log.Error("state_transition", "CallRevertNormalTx,Unmarshal err", errRT)
			continue
		}
		if rt.Typ != common.ExtraRevocable {
			log.Info("state_transition", "CallRevertNormalTx:err:type is ", rt.Typ, "Revert tx type should ", common.ExtraRevocable)
			continue
		}
		for _, vv := range rt.Adam { //一对多交易
			log.Info("state_transition", "CallRevertNormalTx:vv.Addr", vv.Addr, "vv.Amont", vv.Amont)
			log.Info("state_transition", "CallRevertNormalTx:from", rt.From, "vv.Amont", vv.Amont)
			st.state.AddBalance(common.MainAccount, rt.From, vv.Amont)
			st.state.SubBalance(common.WithdrawAccount, rt.From, vv.Amont)
		}
		if val, ok := delval[rt.Tim]; ok {
			val = append(val, hash)
			delval[rt.Tim] = val
		} else {
			delhashs := make([]common.Hash, 0)
			delhashs = append(delhashs, hash)
			delval[rt.Tim] = delhashs
		}
		st.state.DeleteMxData(tmphash, b)
	}
	for k, v := range delval {
		st.state.GetSaveTx(common.ExtraRevocable, k, v, true)
	}
	return ret, st.GasUsed(), vmerr != nil, err
}

func (st *StateTransition) CallRevocableNormalTx() (ret []byte, usedGas uint64, failed bool, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, errors.New("CallRevocableNormalTx from is nil")
	}
	usefrom := tx.From()
	if usefrom == addr {
		return nil, 0, false, errors.New("CallRevocableNormalTx usefrom is nil")
	}
	var (
		vmerr error
	)
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, err
	}
	mapTOAmonts := make([]common.AddrAmont, 0)
	//
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
	st.state.AddBalance(common.WithdrawAccount, usefrom, st.value)
	st.state.SubBalance(common.MainAccount, usefrom, st.value)
	mapTOAmont := common.AddrAmont{Addr: st.To(), Amont: st.value}
	mapTOAmonts = append(mapTOAmonts, mapTOAmont)
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			st.state.AddBalance(common.WithdrawAccount, usefrom, ex.Amount)
			st.state.SubBalance(common.MainAccount, usefrom, ex.Amount)
			mapTOAmont = common.AddrAmont{Addr: *ex.Recipient, Amont: ex.Amount}
			mapTOAmonts = append(mapTOAmonts, mapTOAmont)
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
	var rt common.RecorbleTx
	rt.From = tx.From()
	rt.Tim = tx.GetCreateTime()
	rt.Typ = tx.GetMatrixType()
	rt.Adam = append(rt.Adam, mapTOAmonts...)
	b, marshalerr := json.Marshal(&rt)
	if marshalerr != nil {
		return nil, 0, false, marshalerr
	}
	txHash := tx.Hash()
	//log.Info("file state_transition","func CallRevocableNormalTx:txHash",txHash)
	mapHashamont := make(map[common.Hash][]byte)
	mapHashamont[txHash] = b
	st.state.SaveTx(tx.GetMatrixType(), rt.Tim, mapHashamont)
	st.state.SetMatrixData(txHash, b)
	st.RefundGas()
	st.state.AddBalance(common.MainAccount, common.TxGasRewardAddress, costGas)
	return ret, st.GasUsed(), vmerr != nil, err
}
func (st *StateTransition) CallUnGasNormalTx() (ret []byte, usedGas uint64, failed bool, err error) {
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	toaddr := tx.To()
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, errors.New("CallUnGasNormalTx from is nil")
	}
	sender := vm.AccountRef(from)
	var (
		evm   = st.evm
		vmerr error
	)

	//
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, ErrTXCountOverflow
		}
	}
	st.gas = 0
	issendFromContract := false
	beforAmont := st.state.GetBalanceByType(common.ContractAddress, common.MainAccount)
	interestbefor := st.state.GetBalanceByType(common.InterestRewardAddress, common.MainAccount) // Test
	interset := big.NewInt(0)
	if toaddr == nil { //
		log.Error("state_transition callUnGasNormalTx to is nil")
		return nil, 0, false, ErrTXToNil
	} else {
		// Increment the nonce for the next transaction
		if st.To() == common.ContractAddress {
			interset = new(big.Int).Add(interset, st.value)
			issendFromContract = true
		}
		st.state.SetNonce(tx.From(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			if toaddr == nil {
				log.Error("state_transition callUnGasNormalTx Extro to is nil")
				return nil, 0, false, ErrTXToNil
			} else {
				if *ex.Recipient == common.ContractAddress {
					interset = new(big.Int).Add(interset, ex.Amount)
					issendFromContract = true
				}
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
	if issendFromContract {
		afterAmont := st.state.GetBalanceByType(common.ContractAddress, common.MainAccount)
		difAmont := new(big.Int).Sub(afterAmont, beforAmont)
		if difAmont.Cmp(interset) != 0 {
			log.Info("state_transition", "rewardTx", "ContractAddress 余额与增加的钱不一致")
			return nil, 0, false, ErrinterestAmont
		}
		interestafter := st.state.GetBalanceByType(common.InterestRewardAddress, common.MainAccount)
		dif := new(big.Int).Sub(interestbefor, interestafter)
		if difAmont.Cmp(dif) != 0 {
			log.Info("state_transition", "rewardTx", "InterestRewardAddress 余额与扣除的钱不一致")
			log.Error("state_transition", "difAmont", difAmont, "dif", dif, "afterAmont", afterAmont, "beforAmont", beforAmont, "interestafter", interestafter, "interestbefor", interestbefor)
			return nil, 0, false, ErrinterestAmont
		}
	}
	return ret, 0, vmerr != nil, err
}
func (st *StateTransition) CallNormalTx() (ret []byte, usedGas uint64, failed bool, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	toaddr := tx.To()
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, errors.New("CallNormalTx from is nil")
	}
	//usefrom := tx.AmontFrom()
	usefrom := from
	if usefrom == addr {
		return nil, 0, false, errors.New("CallNormalTx usefrom is nil")
	}
	sender := vm.AccountRef(usefrom)
	var (
		evm   = st.evm
		vmerr error
	)
	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, err
	}
	//
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
	if toaddr == nil { //
		ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(from, st.state.GetNonce(from)+1)
		ret, st.gas, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	//=========begin===============
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
	st.RefundGas()
	st.state.AddBalance(common.MainAccount, common.TxGasRewardAddress, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice))
	return ret, st.GasUsed(), vmerr != nil, err
}

//授权交易的from和to是同一个地址
func (st *StateTransition) CallAuthTx() (ret []byte, usedGas uint64, failed bool, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	toaddr := tx.To()
	sender := vm.AccountRef(tx.From())
	var (
		evm   = st.evm
		vmerr error
	)

	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, err
	}
	//
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
	if toaddr == nil { //
		log.Error("state_transition callAuthTx to is nil")
		return nil, 0, false, ErrTXToNil
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(tx.From(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			if toaddr == nil {
				log.Error("state_transition callAuthTx extro to is nil")
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
	st.RefundGas()
	st.state.AddBalance(common.MainAccount, common.TxGasRewardAddress, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice))

	var entrustOK bool = false
	Authfrom := tx.From()
	EntrustList := make([]common.EntrustType, 0)
	err = json.Unmarshal(tx.Data(), &EntrustList) //EntrustList为被委托人的EntrustType切片
	if err != nil {
		log.Error("CallAuthTx Unmarshal err")
		return nil, st.GasUsed(), true, ErrSpecialTxFailed
	}

	for _, EntrustData := range EntrustList {
		HeightAuthDataList := make([]common.AuthType, 0) //按高度存储授权数据列表
		TimeAuthDataList := make([]common.AuthType, 0)
		str_addres := EntrustData.EntrustAddres //被委托人地址
		addres ,err := base58.Base58DecodeToAddress(str_addres)
		if err != nil{
			return nil, st.GasUsed(), true, ErrSpecialTxFailed
		}
		//tmpAuthMarsha1Data := st.state.GetStateByteArray(addres, common.BytesToHash(addres[:])) //获取授权数据
		tmpAuthMarsha1Data := st.state.GetAuthStateByteArray(addres) //获取授权数据
		if len(tmpAuthMarsha1Data) != 0 {
			//AuthData := new(common.AuthType)
			AuthDataList := make([]common.AuthType, 0)
			err = json.Unmarshal(tmpAuthMarsha1Data, &AuthDataList)
			if err != nil {
				log.Error("CallAuthTx AuthDataList Unmarshal err")
				return nil, st.GasUsed(), true, ErrSpecialTxFailed
			}
			for _, AuthData := range AuthDataList {
				if AuthData.IsEntrustGas == false && AuthData.IsEntrustSign == false {
					continue
				}
				if AuthData.AuthAddres != (common.Address{}) && !(AuthData.AuthAddres.Equal(Authfrom)) {
					//如果不是同一个人授权，先判断之前的授权人权限是否失效，如果之前的授权权限没失效则不能被重复委托
					if AuthData.EnstrustSetType == params.EntrustByHeight{
						if st.evm.BlockNumber.Uint64() <= AuthData.EndHeight{
							//按高度委托未失效
							log.Error("该委托人已经被委托过了，不能重复委托", "from", tx.From(), "Nonce", tx.Nonce())
							return nil, st.GasUsed(), true, ErrSpecialTxFailed //如果一个不满足就返回，不continue
						}
					}else if EntrustData.EnstrustSetType == params.EntrustByTime {
						if st.evm.Time.Uint64() <= AuthData.EndTime{
							//按时间委托未失效
							log.Error("该委托人已经被委托过了，不能重复委托", "from", tx.From(), "Nonce", tx.Nonce())
							return nil, st.GasUsed(), true, ErrSpecialTxFailed //如果一个不满足就返回，不continue
						}
					}else{
						//该条件不可能发生
						log.Error("之前的授权人数据丢失")
					}
				}
				//如果是同一个人委托，委托的高度不能重合
				if AuthData.AuthAddres.Equal(Authfrom) {
					if EntrustData.EnstrustSetType == params.EntrustByHeight {
						//按高度委托
						if EntrustData.StartHeight <= AuthData.EndHeight {
							log.Error("同一个授权人的委托高度不能重合", "from", tx.From(), "Nonce", tx.Nonce())
							return nil, st.GasUsed(), true, ErrSpecialTxFailed
						}
						HeightAuthDataList = append(HeightAuthDataList, AuthData)
					} else if EntrustData.EnstrustSetType == params.EntrustByTime {
						//按时间委托
						if EntrustData.StartTime <= AuthData.EndTime {
							log.Error("同一个授权人的委托时间不能重合", "from", tx.From(), "Nonce", tx.Nonce())
							return nil, 0, true, ErrSpecialTxFailed
						}
						TimeAuthDataList = append(TimeAuthDataList, AuthData)
					} else {
						log.Error("未设置委托类型", "from", tx.From(), "Nonce", tx.Nonce())
						return nil, st.GasUsed(), true, ErrSpecialTxFailed
					}
				}
			}
		}
		entrustOK = true
		//反向存储AuthType结构，用来通过被委托人from和高度查找授权人from
		if EntrustData.EnstrustSetType == params.EntrustByHeight {
			//按块高存
			t_authData := new(common.AuthType)
			t_authData.EnstrustSetType = EntrustData.EnstrustSetType
			t_authData.StartHeight = EntrustData.StartHeight
			t_authData.EndHeight = EntrustData.EndHeight
			t_authData.IsEntrustSign = EntrustData.IsEntrustSign
			t_authData.IsEntrustGas = EntrustData.IsEntrustGas
			t_authData.AuthAddres = Authfrom
			HeightAuthDataList = append(HeightAuthDataList, *t_authData)
			marshalAuthData, err := json.Marshal(HeightAuthDataList)
			if err != nil {
				log.Error("Marshal err")
				return nil, st.GasUsed(), true, ErrSpecialTxFailed
			}
			//marsha1AuthData是authData的Marsha1编码
			st.state.SetAuthStateByteArray(addres, marshalAuthData) //设置授权数据
		}

		if EntrustData.EnstrustSetType == params.EntrustByTime {
			//按时间存
			t_authData := new(common.AuthType)
			t_authData.EnstrustSetType = EntrustData.EnstrustSetType
			t_authData.StartTime = EntrustData.StartTime
			t_authData.EndTime = EntrustData.EndTime
			t_authData.IsEntrustSign = EntrustData.IsEntrustSign
			t_authData.IsEntrustGas = EntrustData.IsEntrustGas
			t_authData.AuthAddres = Authfrom
			TimeAuthDataList = append(TimeAuthDataList, *t_authData)
			marshalAuthData, err := json.Marshal(TimeAuthDataList)
			if err != nil {
				log.Error("Marshal err")
				return nil, st.GasUsed(), true, ErrSpecialTxFailed
			}
			//marsha1AuthData是authData的Marsha1编码
			st.state.SetAuthStateByteArray(addres, marshalAuthData) //设置授权数据
		}
	}
	if entrustOK {
		//获取之前的委托数据(结构体切片经过marshal编码)
		AllEntrustList := make([]common.EntrustType, 0)
		oldEntrustList := st.state.GetEntrustStateByteArray(Authfrom) //获取委托数据
		if len(oldEntrustList) != 0 {
			err = json.Unmarshal(oldEntrustList, &AllEntrustList)
			if err != nil {
				log.Error("CallAuthTx Unmarshal err")
				return nil, st.GasUsed(), true, ErrSpecialTxFailed
			}
		}
		AllEntrustList = append(AllEntrustList, EntrustList...)
		allDataList, err := json.Marshal(AllEntrustList)
		if err != nil {
			log.Error("Marshal error")
		}
		st.state.SetEntrustStateByteArray(Authfrom, allDataList) //设置委托数据
		entrustOK = false
	} else {
		log.Error("委托条件不满足")
		return nil, st.GasUsed(), true, ErrSpecialTxFailed
	}

	return ret, st.GasUsed(), vmerr != nil, nil
}

func isContain(a uint32, list []uint32) bool {
	for _, data := range list {
		if data == a {
			return true
		}
	}
	return false
}

func (st *StateTransition) CallCancelAuthTx() (ret []byte, usedGas uint64, failed bool, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	toaddr := tx.To()
	sender := vm.AccountRef(tx.From())
	var (
		evm   = st.evm
		vmerr error
	)

	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, err
	}
	//
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
	if toaddr == nil { //
		log.Error("state transition callAuthTx to is nil")
		return nil, 0, false, ErrTXToNil
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(tx.From(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			if toaddr == nil {
				log.Error("state transition callAuthTx Extro to is nil")
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
	st.RefundGas()
	st.state.AddBalance(common.MainAccount, common.TxGasRewardAddress, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice))

	Authfrom := tx.From()
	delIndexList := make([]uint32, 0)
	err = json.Unmarshal(tx.Data(), &delIndexList) //EntrustList为被委托人的EntrustType切片
	if err != nil {
		log.Error("CallAuthTx Unmarshal err")
		return nil, st.GasUsed(), true, ErrSpecialTxFailed
	}
	EntrustMarsha1Data := st.state.GetEntrustStateByteArray(Authfrom) //获取委托数据
	if len(EntrustMarsha1Data) == 0 {
		log.Error("没有委托数据")
		return nil, st.GasUsed(), true, ErrSpecialTxFailed
	}
	entrustDataList := make([]common.EntrustType, 0)
	err = json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	if err != nil {
		log.Error("CallAuthTx Unmarshal err")
		return nil, st.GasUsed(), true, ErrSpecialTxFailed
	}
	newentrustDataList := make([]common.EntrustType, 0)
	for index, entrustFrom := range entrustDataList {
		if isContain(uint32(index), delIndexList) {
			//要删除的切片数据
			str_addres := entrustFrom.EntrustAddres //被委托人地址
			addres,err := base58.Base58DecodeToAddress(str_addres)
			if err != nil{
				return nil, st.GasUsed(), true, ErrSpecialTxFailed
			}
			marshaldata := st.state.GetAuthStateByteArray(addres) //获取之前的授权数据切片,marshal编码过的  //获取授权数据
			if len(marshaldata) > 0 {
				//oldAuthData := new(common.AuthType)   //oldAuthData的地址为0x地址
				oldAuthDataList := make([]common.AuthType, 0)
				err = json.Unmarshal(marshaldata, &oldAuthDataList) //oldAuthData的地址为0x地址
				if err != nil {
					return nil, st.GasUsed(), true, ErrSpecialTxFailed
				}
				newDelAuthDataList := make([]common.AuthType, 0)
				for _, oldAuthData := range oldAuthDataList {
					//只要起始高度或时间能对应上，就是要删除的切片
					if entrustFrom.StartHeight == oldAuthData.StartHeight || entrustFrom.StartTime == oldAuthData.StartTime {
						oldAuthData.IsEntrustGas = false
						oldAuthData.IsEntrustSign = false
						newDelAuthDataList = append(newDelAuthDataList, oldAuthData)
					}
				}
				newAuthDatalist, err := json.Marshal(newDelAuthDataList)
				if err != nil {
					return nil, st.GasUsed(), true, ErrSpecialTxFailed
				}
				st.state.SetAuthStateByteArray(addres, newAuthDatalist) //设置授权数据
			}
		} else {
			//新的切片数据
			newentrustDataList = append(newentrustDataList, entrustFrom)
		}
	}

	newEntrustList, err := json.Marshal(newentrustDataList)
	if err != nil {
		log.Error("CallAuthTx Marshal err")
	}
	st.state.SetEntrustStateByteArray(Authfrom, newEntrustList) //设置委托数据

	return ret, st.GasUsed(), vmerr != nil, nil
}
func (st *StateTransition) CallSuperTx() (ret []byte, usedGas uint64, failed bool, err error) {
	//if err = st.PreCheck(); err != nil {
	//	return
	//}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	toaddr := tx.To()
	sender := vm.AccountRef(tx.From())
	var (
		evm   = st.evm
		vmerr error
	)

	//
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, ErrTXCountOverflow
		}
	}
	st.gas = 0
	if toaddr == nil { //
		log.Error("state_transition callAuthTx to is nil")
		return nil, 0, false, ErrTXToNil
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(tx.From(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			if toaddr == nil {
				log.Error("state_transition callAuthTx Extro to is nil")
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

	configData := make(map[string]interface{})
	err = json.Unmarshal(tx.Data(), &configData)
	if err != nil {
		log.Error("CallSuperTx Unmarshal err")
		return nil, 0, true, nil
	}

	version := matrixstate.GetVersionInfo(st.state)
	mgr := matrixstate.GetManager(version)
	if mgr == nil {
		return nil, 0, true, nil
	}

	supMager := supertxsstate.GetManager(version)
	snp := st.state.Snapshot()
	for k, v := range configData {
		val, OK := supMager.Check(k, v)
		if OK {
			opt, err := mgr.FindOperator(k)
			if err != nil {
				log.Error("CallSuperTx:FindOperator failed", "key", k, "value", val, "err", err)
				st.state.RevertToSnapshot(snp)
				return nil, 0, true, nil
			}
			err = opt.SetValue(st.state, val)
			if err != nil {
				log.Error("CallSuperTx:SetValue failed", "key", k, "value", val, "err", err)
				st.state.RevertToSnapshot(snp)
				return nil, 0, true, nil
			}
		} else {
			log.Error("CallSuperTx:Check failed", "key", k, "value", val, "err", err)
			st.state.RevertToSnapshot(snp)
			return nil, 0, true, nil
		}
	}
	return ret, 0, vmerr != nil, nil
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
	st.state.AddBalance(common.MainAccount, st.msg.AmontFrom(), remaining)

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) GasUsed() uint64 {
	return st.initialGas - st.gas
}
