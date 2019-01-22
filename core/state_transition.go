// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"encoding/json"
	"errors"
	"math"
	"math/big"

	"github.com/matrix/go-matrix/base58"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/txinterface"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
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
	return &StateTransition{
		gp:       gp,
		evm:      evm,
		msg:      msg,
		gasPrice: big.NewInt(int64(params.TxGasPrice)),
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
		log.Error("File state_transition", "func AppleMessage", "interface is nil")
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
		//case common.ExtraEntrustTx:
		//todo
		//tx.Data()
		case common.ExtraAuthTx:
			log.INFO("====ZH: 授权交易", "txtype", txtype)
			return st.CallAuthTx()
		case common.ExtraCancelEntrust:
			log.INFO("====ZH: 取消委托", "txtype", txtype)
			return st.CallCancelAuthTx()
		default:
			log.Info("File state_transition", "func Transitiondb", "Unknown extra txtype")
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
		return nil, 0, false, errors.New("file state_transition,func CallTimeNormalTx ,from is nil")
	}
	//usefrom := tx.AmontFrom()
	usefrom := tx.From()
	if usefrom == addr {
		return nil, 0, false, errors.New("file state_transition,func CallTimeNormalTx ,usefrom is nil")
	}
	var (
		vmerr error
	)
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, err
	}
	mapTOAmonts := make([]common.AddrAmont, 0)
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
		return nil, 0, false, errors.New("file state_transition,func CallRevertNormalTx ,from is nil")
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
	hashlist = append(hashlist, hash)
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			hash.SetBytes(ex.Payload)
			hashlist = append(hashlist, hash)
			if vmerr != nil {
				break
			}
		}
	}
	costGas := new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)
	st.state.AddBalance(common.MainAccount, common.TxGasRewardAddress, costGas)
	delval := make(map[uint32][]common.Hash)
	for _, tmphash := range hashlist {
		if common.EmptyHash(tmphash) {
			continue
		}
		b := st.state.GetMatrixData(tmphash)
		if b == nil {
			log.Error("file state_transition", "func CallRevertNormalTx,err", "not found tx hash,maybe the transaction has lasted more than 24 hours")
			continue
		}
		var rt common.RecorbleTx
		errRT := json.Unmarshal(b, &rt)
		if errRT != nil {
			log.Error("file state_transition", "func CallRevertNormalTx,Unmarshal err", errRT)
			continue
		}
		if rt.Typ != common.ExtraRevocable {
			log.Info("file state_transition", "func CallRevertNormalTx:err:type is ", rt.Typ, "Revert tx type should ", common.ExtraRevocable)
			continue
		}
		for _, vv := range rt.Adam { //一对多交易
			log.Info("file state_transition", "func CallRevertNormalTx:vv.Addr", vv.Addr, "vv.Amont", vv.Amont)
			log.Info("file state_transition", "func CallRevertNormalTx:from", rt.From, "vv.Amont", vv.Amont)
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
		return nil, 0, false, errors.New("file state_transition,func CallRevocableNormalTx ,from is nil")
	}
	usefrom := tx.From()
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
	mapTOAmonts := make([]common.AddrAmont, 0)
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
	st.state.AddBalance(common.MainAccount, common.TxGasRewardAddress, costGas)
	return ret, st.GasUsed(), vmerr != nil, err
}
func (st *StateTransition) CallUnGasNormalTx() (ret []byte, usedGas uint64, failed bool, err error) {
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	toaddr := tx.To()
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, errors.New("file state_transition,func CallUnGasNormalTx ,from is nil")
	}
	sender := vm.AccountRef(from)
	var (
		evm   = st.evm
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
	issendFromContract := false
	beforAmont := st.state.GetBalanceByType(common.ContractAddress, common.MainAccount)
	interestbefor := st.state.GetBalanceByType(common.InterestRewardAddress, common.MainAccount) // Test
	interset := big.NewInt(0)
	if toaddr == nil { //YY
		log.Error("file state_transition", "func CallUnGasNormalTx()", "to is nil")
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
				log.Error("file state_transition", "func CallUnGasNormalTx()", "Extro to is nil")
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
			log.Info("file state_transition", "func rewardTx", "ContractAddress 余额与增加的钱不一致")
			return nil, 0, false, ErrinterestAmont
		}
		interestafter := st.state.GetBalanceByType(common.InterestRewardAddress, common.MainAccount)
		dif := new(big.Int).Sub(interestbefor, interestafter)
		if difAmont.Cmp(dif) != 0 {
			log.Info("file state_transition", "func rewardTx", "InterestRewardAddress 余额与扣除的钱不一致")
			log.Error("ZH:state_transition", "difAmont", difAmont, "dif", dif, "afterAmont", afterAmont, "beforAmont", beforAmont, "interestafter", interestafter, "interestbefor", interestbefor)
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
		return nil, 0, false, errors.New("file state_transition,func CallNormalTx ,from is nil")
	}
	//usefrom := tx.AmontFrom()
	usefrom := from
	if usefrom == addr {
		return nil, 0, false, errors.New("file state_transition,func CallNormalTx ,usefrom is nil")
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
	if toaddr == nil { //YY
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

	var entrustOK bool = false
	Authfrom := tx.From()
	EntrustList := make([]common.EntrustType, 0)
	err = json.Unmarshal(tx.Data(), &EntrustList) //EntrustList为被委托人的EntrustType切片
	if err != nil {
		log.Error("CallAuthTx Unmarshal err")
		return nil, 0, false, err
	}

	for _, EntrustData := range EntrustList {
		HeightAuthDataList := make([]common.AuthType, 0) //按高度存储授权数据列表
		TimeAuthDataList := make([]common.AuthType, 0)
		str_addres := EntrustData.EntrustAddres //被委托人地址
		addres := base58.Base58DecodeToAddress(str_addres)
		tmpAuthMarsha1Data := st.state.GetStateByteArray(addres, common.BytesToHash(addres[:]))
		if len(tmpAuthMarsha1Data) != 0 {
			//AuthData := new(common.AuthType)
			AuthDataList := make([]common.AuthType, 0)
			err = json.Unmarshal(tmpAuthMarsha1Data, &AuthDataList)
			if err != nil {
				log.Error("CallAuthTx AuthDataList Unmarshal err")
				return nil, 0, false, err
			}
			for _, AuthData := range AuthDataList {
				if AuthData.IsEntrustGas == false && AuthData.IsEntrustSign == false {
					continue
				}
				if AuthData.AuthAddres != (common.Address{}) && !(AuthData.AuthAddres.Equal(Authfrom)) {
					log.Error("该委托人已经被委托过了，不能重复委托", "from", tx.From(), "Nonce", tx.Nonce())
					return nil, 0, false, ErrRepeatEntrust //如果一个不满足就返回，不continue
				}
				//如果是同一个人委托，委托的高度不能重合
				if AuthData.AuthAddres.Equal(Authfrom) {
					if EntrustData.EnstrustSetType == params.EntrustByHeight {
						//按高度委托
						if EntrustData.StartHeight <= AuthData.EndHeight {
							log.Error("同一个授权人的委托高度不能重合", "from", tx.From(), "Nonce", tx.Nonce())
							return nil, 0, false, ErrRepeatEntrust
						}
						HeightAuthDataList = append(HeightAuthDataList, AuthData)
					} else if EntrustData.EnstrustSetType == params.EntrustByTime {
						//按时间委托
						if EntrustData.StartTime <= AuthData.EndTime {
							log.Error("同一个授权人的委托时间不能重合", "from", tx.From(), "Nonce", tx.Nonce())
							return nil, 0, false, ErrRepeatEntrust
						}
						TimeAuthDataList = append(TimeAuthDataList, AuthData)
					} else {
						log.Error("未设置委托类型", "from", tx.From(), "Nonce", tx.Nonce())
						return nil, 0, false, errors.New("without set entrust type")
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
				return nil, 0, false, err
			}
			//marsha1AuthData是authData的Marsha1编码
			st.state.SetStateByteArray(addres, common.BytesToHash(addres[:]), marshalAuthData)
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
				return nil, 0, false, err
			}
			//marsha1AuthData是authData的Marsha1编码
			st.state.SetStateByteArray(addres, common.BytesToHash(addres[:]), marshalAuthData)
		}
	}
	if entrustOK {
		//获取之前的委托数据(结构体切片经过marshal编码)
		AllEntrustList := make([]common.EntrustType, 0)
		oldEntrustList := st.state.GetStateByteArray(Authfrom, common.BytesToHash(Authfrom[:]))
		if len(oldEntrustList) != 0 {
			err = json.Unmarshal(oldEntrustList, &AllEntrustList)
			if err != nil {
				log.Error("CallAuthTx Unmarshal err")
				return nil, 0, false, err
			}
		}
		AllEntrustList = append(AllEntrustList, EntrustList...)
		allDataList, err := json.Marshal(AllEntrustList)
		if err != nil {
			log.Error("Marshal error")
		}
		st.state.SetStateByteArray(Authfrom, common.BytesToHash(Authfrom[:]), allDataList)
		entrustOK = false
	} else {
		log.Error("委托条件不满足")
	}

	//YY
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, ErrTXCountOverflow
		}
	}
	st.gas = 0
	if toaddr == nil { //YY
		log.Error("file state_transition", "func CallAuthTx()", "to is nil")
		return nil, 0, false, ErrTXToNil
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(tx.From(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			if toaddr == nil {
				log.Error("file state_transition", "func CallAuthTx()", "Extro to is nil")
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

	Authfrom := tx.From()
	delIndexList := make([]uint32, 0)
	err = json.Unmarshal(tx.Data(), &delIndexList) //EntrustList为被委托人的EntrustType切片
	if err != nil {
		log.Error("CallAuthTx Unmarshal err")
		return nil, 0, false, err
	}
	EntrustMarsha1Data := st.state.GetStateByteArray(Authfrom, common.BytesToHash(Authfrom[:]))
	if len(EntrustMarsha1Data) == 0 {
		log.Error("没有委托数据")
		return nil, 0, false, errors.New("without entrust data")
	}
	entrustDataList := make([]common.EntrustType, 0)
	err = json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	if err != nil {
		log.Error("CallAuthTx Unmarshal err")
	}
	newentrustDataList := make([]common.EntrustType, 0)
	for index, entrustFrom := range entrustDataList {
		if isContain(uint32(index), delIndexList) {
			//要删除的切片数据
			str_addres := entrustFrom.EntrustAddres //被委托人地址
			addres := base58.Base58DecodeToAddress(str_addres)
			marshaldata := st.state.GetStateByteArray(addres, common.BytesToHash(addres[:])) //获取之前的授权数据切片,marshal编码过的
			if len(marshaldata) > 0 {
				//oldAuthData := new(common.AuthType)   //oldAuthData的地址为0x地址
				oldAuthDataList := make([]common.AuthType, 0)
				err = json.Unmarshal(marshaldata, &oldAuthDataList) //oldAuthData的地址为0x地址
				if err != nil {
					return nil, 0, false, err
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
					return nil, 0, false, err
				}
				st.state.SetStateByteArray(addres, common.BytesToHash(addres[:]), newAuthDatalist)
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
	st.state.SetStateByteArray(Authfrom, common.BytesToHash(Authfrom[:]), newEntrustList)

	//YY
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, ErrTXCountOverflow
		}
	}
	st.gas = 0
	if toaddr == nil { //YY
		log.Error("file state_transition", "func CallAuthTx()", "to is nil")
		return nil, 0, false, ErrTXToNil
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(tx.From(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			if toaddr == nil {
				log.Error("file state_transition", "func CallAuthTx()", "Extro to is nil")
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
func (st *StateTransition) RefundGas() {
	// Apply refund counter, capped to half of the used gas.
	refund := st.GasUsed() / 2
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	// Return ETH for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.state.AddBalance(common.MainAccount, st.msg.From(), remaining)

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) GasUsed() uint64 {
	return st.initialGas - st.gas
}
