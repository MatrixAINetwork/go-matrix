// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"encoding/json"
	"errors"
	"math"
	"math/big"

	"strings"

	"bufio"
	"github.com/MatrixAINetwork/go-matrix/base58"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/txinterface"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"os"
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
	state      vm.StateDBManager
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
	//gasprice, err := matrixstate.GetTxpoolGasLimit(evm.StateDB)
	//if err != nil {
	//	log.Error("NewStateTransition err")
	//	return nil
	//}
	return &StateTransition{
		gp:       gp,
		evm:      evm,
		msg:      msg,
		gasPrice: new(big.Int).SetUint64(params.TxGasPrice),
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
func (st *StateTransition) getCoinAddress(cointyp string) (rewardaddr common.Address, coinrange string) {
	if cointyp == params.MAN_COIN || cointyp == "" {
		return common.TxGasRewardAddress, cointyp
	}
	coinconfig := st.state.GetMatrixData(types.RlpHash(common.COINPREFIX + mc.MSCurrencyConfig))
	var coincfglist []common.CoinConfig
	if len(coinconfig) > 0 {
		err := json.Unmarshal(coinconfig, &coincfglist)
		if err != nil {
			log.Trace("get coin config list", "unmarshal err", err)
			return common.TxGasRewardAddress, params.MAN_COIN
		}
	}
	for _, cc := range coincfglist {
		if cc.PackNum <= 0 {
			continue
		}
		if cc.CoinType == cointyp {
			rewardaddr = cc.CoinAddress
			coinrange = cc.CoinRange
			break
		}
	}
	return rewardaddr, coinrange
}

//按币种分区扣gas（该接口废弃）
func (st *StateTransition) BuyGas_coin() error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Gas()), st.gasPrice)
	for _, tAccount := range st.state.GetBalance(st.msg.GetTxCurrency(), st.msg.From()) {
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
	coinCfglist, err := matrixstate.GetCoinConfig(st.state)
	if err != nil {
		return errors.New("get GetCoinConfig err")
	}
	var payGasType string = params.MAN_COIN
	for _, coinCfg := range coinCfglist {
		if coinCfg.CoinType == st.msg.GetTxCurrency() {
			payGasType = coinCfg.CoinRange
			break
		}
	}
	balance := st.state.GetBalanceByType(payGasType, st.msg.AmontFrom(), common.MainAccount)
	if balance.Cmp(mgval) < 0 {
		log.Error("MAN", "BuyGas err", "MAN Coin : Insufficient account balance.")
		return errors.New("MAN Coin : Insufficient account balance.")
	}
	st.state.SubBalance(payGasType, common.MainAccount, st.msg.AmontFrom(), mgval)
	return nil
}

//扣各自币种的gas
func (st *StateTransition) BuyGas() error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Gas()), st.gasPrice)
	for _, tAccount := range st.state.GetBalance(st.msg.GetTxCurrency(), st.msg.From()) {
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
	balance := st.state.GetBalanceByType(st.msg.GetTxCurrency(), st.msg.AmontFrom(), common.MainAccount)
	if balance.Cmp(mgval) < 0 {
		log.Error("MAN", "BuyGas err", "MAN Coin : Insufficient account balance.")
		return errors.New("MAN Coin : Insufficient account balance.")
	}
	st.state.SubBalance(st.msg.GetTxCurrency(), common.MainAccount, st.msg.AmontFrom(), mgval)
	return nil
}

func (st *StateTransition) PreCheck() error {
	// Make sure this transaction's nonce is correct.
	if st.msg.CheckNonce() {
		nonce := st.state.GetNonce(st.msg.GetTxCurrency(), st.msg.From())
		if nonce < st.msg.Nonce() {
			log.Error("ErrNonceTooHigh", "txNonce", st.msg.Nonce(), "stateNonce", nonce, "txfrom", st.msg.From(), "txhash", st.msg.Hash().Hex())
			return ErrNonceTooHigh
		} else if nonce > st.msg.Nonce() {
			log.Error("ErrNonceTooLow", "txNonce", st.msg.Nonce(), "stateNonce", nonce, "txfrom", st.msg.From(), "txhash", st.msg.Hash().Hex())
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
func ApplyMessage(evm *vm.EVM, tx txinterface.Message, gp *GasPool) ([]byte, uint64, bool, []uint, error) {
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
func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, failed bool, shardings []uint, err error) {
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	txtype := tx.GetMatrixType()
	if txtype != common.ExtraNormalTxType && txtype != common.ExtraAItxType {
		switch txtype {
		case common.ExtraRevocable:
			return st.CallRevocableNormalTx()
		case common.ExtraRevertTxType:
			return st.CallRevertNormalTx()
		case common.ExtraUnGasMinerTxType, common.ExtraUnGasValidatorTxType, common.ExtraUnGasInterestTxType, common.ExtraUnGasTxsType, common.ExtraUnGasLotteryTxType:
			return st.CallUnGasNormalTx()
		case common.ExtraTimeTxType:
			return st.CallTimeNormalTx()
		case common.ExtraAuthTx:
			log.INFO("授权交易", "交易类型", txtype)
			return st.CallAuthTx()
		case common.ExtraCancelEntrust:
			log.INFO("取消委托", "交易类型", txtype)
			return st.CallCancelAuthTx()
		case common.ExtraMakeCoinType:
			return st.CallMakeCoinTx()
		case common.ExtraSetBlackListTxType:
			return st.CallSetBlackListTx()
		default:
			log.Info("state transition unknown extra txtype")
			return nil, 0, false, nil, ErrTXUnknownType
		}

	} else {
		return st.CallNormalTx()
	}
}
func (st *StateTransition) CallTimeNormalTx() (ret []byte, usedGas uint64, failed bool, shardings []uint, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, shardings, errors.New("CallTimeNormalTx from is nil")
	}
	usefrom := tx.From()
	if usefrom == addr {
		return nil, 0, false, shardings, errors.New("CallTimeNormalTx usefrom is nil")
	}
	var (
		vmerr error
	)
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, shardings, err
	}
	mapTOAmonts := make([]common.AddrAmont, 0)
	//
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, shardings, ErrTXCountOverflow
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return nil, 0, false, shardings, err
			}
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, shardings, err
	}
	st.state.SetNonce(st.msg.GetTxCurrency(), from, st.state.GetNonce(st.msg.GetTxCurrency(), from)+1)
	gasaddr, coinrange := st.getCoinAddress(tx.GetTxCurrency())
	st.RefundGas(coinrange)
	st.state.AddBalance(coinrange, common.MainAccount, gasaddr, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice))
	st.state.AddBalance(st.msg.GetTxCurrency(), common.WithdrawAccount, usefrom, st.value)
	st.state.SubBalance(st.msg.GetTxCurrency(), common.MainAccount, usefrom, st.value)
	shardings = append(shardings, uint(from[0]))
	shardings = append(shardings, uint(st.To()[0]))
	mapTOAmont := common.AddrAmont{Addr: st.To(), Amont: st.value}
	mapTOAmonts = append(mapTOAmonts, mapTOAmont)
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			st.state.AddBalance(st.msg.GetTxCurrency(), common.WithdrawAccount, usefrom, ex.Amount)
			st.state.SubBalance(st.msg.GetTxCurrency(), common.MainAccount, usefrom, ex.Amount)
			mapTOAmont = common.AddrAmont{Addr: *ex.Recipient, Amont: ex.Amount}
			shardings = append(shardings, uint(ex.Recipient[0]))
			mapTOAmonts = append(mapTOAmonts, mapTOAmont)
			if vmerr != nil {
				break
			}
		}
	}
	//costGas := new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, nil, vmerr
		}
	}
	rt := new(common.RecorbleTx)
	rt.From = tx.From()
	rt.Tim = tx.GetCreateTime()
	rt.Typ = tx.GetMatrixType()
	rt.Cointyp = tx.GetTxCurrency()
	rt.Adam = append(rt.Adam, mapTOAmonts...)
	b, marshalerr := json.Marshal(rt)
	if marshalerr != nil {
		return nil, 0, false, nil, marshalerr
	}
	txHash := tx.Hash()
	mapHashamont := make(map[common.Hash][]byte)
	mapHashamont[txHash] = b
	st.state.SaveTx(st.msg.GetTxCurrency(), st.msg.From(), tx.GetMatrixType(), rt.Tim, mapHashamont)
	return ret, st.GasUsed(), vmerr != nil, shardings, err
}
func (st *StateTransition) CallRevertNormalTx() (ret []byte, usedGas uint64, failed bool, shardings []uint, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	hashlist := make([]common.Hash, 0)
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, shardings, errors.New("CallRevertNormalTx from is nil")
	}
	var (
		vmerr error
	)
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, shardings, err
	}

	//
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, shardings, ErrTXCountOverflow
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return nil, 0, false, shardings, err
			}
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, shardings, err
	}
	st.state.SetNonce(st.msg.GetTxCurrency(), from, st.state.GetNonce(st.msg.GetTxCurrency(), from)+1)
	var hash common.Hash = common.BytesToHash(tx.Data())
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
	//costGas := new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)
	gasaddr, coinrange := st.getCoinAddress(tx.GetTxCurrency())
	st.RefundGas(coinrange)
	st.state.AddBalance(coinrange, common.MainAccount, gasaddr, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)) //给对应币种奖励账户加钱

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
		if rt.From != from {
			log.Error("state_transition", "CallRevertNormalTx, err", "Revert tx from different Revocable tx from")
			continue
		}
		if rt.Typ != common.ExtraRevocable {
			log.Info("state_transition", "CallRevertNormalTx:err:type is ", rt.Typ, "Revert tx type should ", common.ExtraRevocable)
			continue
		}
		if rt.Cointyp != st.msg.GetTxCurrency() {
			log.Info("state_transition", "CallRevertNormalTx:err:tx coin type", st.msg.GetTxCurrency(), "statedb val coin type", rt.Cointyp)
			continue
		}
		for _, vv := range rt.Adam { //一对多交易
			log.Info("state_transition", "CallRevertNormalTx:vv.Addr", vv.Addr, "vv.Amont", vv.Amont)
			log.Info("state_transition", "CallRevertNormalTx:from", rt.From, "vv.Amont", vv.Amont)
			st.state.AddBalance(rt.Cointyp, common.MainAccount, rt.From, vv.Amont)
			st.state.SubBalance(rt.Cointyp, common.WithdrawAccount, rt.From, vv.Amont)
			shardings = append(shardings, uint(vv.Addr[0]))
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
		shardings = append(shardings, uint(rt.From[0]))
	}
	for k, v := range delval {
		st.state.GetSaveTx(st.msg.GetTxCurrency(), st.msg.From(), common.ExtraRevocable, k, v, true)
	}
	return ret, st.GasUsed(), vmerr != nil, shardings, err
}
func isExistCoin(newCoin string, coinlist []string) bool {
	for _, coin := range coinlist {
		if coin == newCoin {
			return true
		}
	}
	return false
}
func (st *StateTransition) CallMakeCoinTx() (ret []byte, usedGas uint64, failed bool, shardings []uint, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	toaddr := tx.To()
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, shardings, errors.New("state_transition,make coin ,from is nil")
	}
	sender := vm.AccountRef(from)

	var (
		evm   = st.evm
		vmerr error
	)
	tmpshard := make([]uint, 0)
	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, shardings, err
	}
	//
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, shardings, ErrTXCountOverflow
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return nil, 0, false, shardings, err
			}
			//0.7+0.3*pow(0.9,(num-1))
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, shardings, err
	}
	if toaddr == nil {
		var caddr common.Address
		ret, caddr, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
		tmpshard = append(tmpshard, uint(caddr[0]))
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(tx.GetTxCurrency(), from, st.msg.Nonce()+1) //st.state.GetNonce(tx.GetTxCurrency(), from)
		ret, st.gas, tmpshard, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}

	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, nil, vmerr
		}
	}
	shardings = append(shardings, tmpshard...)

	by := tx.Data()
	var makecoin common.SMakeCoin
	makecerr := json.Unmarshal(by, &makecoin)
	if makecerr != nil {
		log.Trace("Make Coin", "Unmarshal err", makecerr)
		return nil, 0, false, shardings, makecerr
	}

	if !common.IsValidityCurrency(makecoin.CoinName) {
		return nil, 0, false, shardings, errors.New("state_transition,make coin err, coin name Wrongful")
	}
	if len(makecoin.AddrAmount) <= 0 {
		return nil, 0, false, shardings, errors.New("state_transition,make coin err, address and amount is nil")
	}
	addrVal := make(map[common.Address]*big.Int)
	for str, amount := range makecoin.AddrAmount {
		if str == "" {
			return nil, 0, false, shardings, errors.New("state_transition,make coin err, coin addr is nil")
		}
		addr, err := base58.Base58DecodeToAddress(str)
		if err != nil {
			log.Trace("Make Coin", "invalid send address", "Base58toAddr err", "base58 addr", str)
			return nil, 0, false, shardings, errors.New("Base58toAddr err")
		}
		strcoin := strings.Split(str, ".")
		if makecoin.CoinName != strcoin[0] {
			log.Error("Make Coin", "invalid send address", "Currency mismatch with account")
			return nil, 0, false, shardings, errors.New("Currency mismatch with account")
		}
		addrVal[addr] = (*big.Int)(amount)
	}
	key := types.RlpHash(params.COIN_NAME)
	coinlistbyte := st.state.GetMatrixData(key)
	var coinlist []string
	if len(coinlistbyte) > 0 {
		err := json.Unmarshal(coinlistbyte, &coinlist)
		if err != nil {
			log.Trace("get coin list", "unmarshal err", err)
			return nil, 0, false, nil, err
		}
	}
	if isExistCoin(makecoin.CoinName, coinlist) {
		return nil, 0, false, shardings, errors.New("Coin exist")
	}
	clmap := make(map[string]bool)
	coinlist = append(coinlist, makecoin.CoinName)
	var clslice []string
	for _, cl := range coinlist {
		if _, ok := clmap[cl]; !ok {
			clmap[cl] = true
			clslice = append(clslice, cl)
		}
	}
	coinby, _ := json.Marshal(clslice)
	st.state.SetMatrixData(key, coinby)

	coinconfig := st.state.GetMatrixData(types.RlpHash(common.COINPREFIX + mc.MSCurrencyConfig))
	var coincfglist []common.CoinConfig
	json.Unmarshal(coinconfig, &coincfglist)
	if makecoin.PackNum == 0 {
		makecoin.PackNum = params.CallTxPachNum
	}
	if makecoin.CoinUnit == nil {
		makecoin.CoinUnit = (*hexutil.Big)(new(big.Int).SetUint64(params.CoinTypeUnit))
	}
	if makecoin.CoinAddress == (common.Address{}) {
		makecoin.CoinAddress = common.TxGasRewardAddress
	}
	st.state.MakeStatedb(makecoin.CoinName, false)
	totalAmont := big.NewInt(0)
	for address, val := range addrVal {
		st.state.SetBalance(makecoin.CoinName, common.MainAccount, address, val)
		totalAmont = new(big.Int).Add(totalAmont, val)
	}
	isCoin := true
	for i, cc := range coincfglist {
		if cc.CoinType == makecoin.CoinName {
			coincfglist[i].CoinType = makecoin.CoinName
			coincfglist[i].PackNum = makecoin.PackNum
			coincfglist[i].CoinTotal = (*hexutil.Big)(totalAmont)
			coincfglist[i].CoinUnit = makecoin.CoinUnit
			coincfglist[i].CoinAddress = makecoin.CoinAddress
			coincfglist[i].CoinRange = makecoin.CoinName //coinrange和cointype是一个类型，为了扩展方便保留该字段
			isCoin = false
		}
	}
	if isCoin {
		tmpcc := common.CoinConfig{
			CoinType:    makecoin.CoinName,
			PackNum:     makecoin.PackNum,
			CoinTotal:   new(hexutil.Big),
			CoinUnit:    new(hexutil.Big),
			CoinAddress: makecoin.CoinAddress,
			CoinRange:   makecoin.CoinName,
		}
		tmpcc.CoinTotal = (*hexutil.Big)(totalAmont)
		tmpcc.CoinUnit = makecoin.CoinUnit
		coincfglist = append(coincfglist, tmpcc)
	}
	//coinCfgbs, _ := rlp.EncodeToBytes(coincfglist)
	coinCfgbs, _ := json.Marshal(coincfglist)
	st.state.SetMatrixData(types.RlpHash(common.COINPREFIX+mc.MSCurrencyConfig), coinCfgbs)
	gasaddr, coinrange := st.getCoinAddress(tx.GetTxCurrency())
	st.RefundGas(coinrange)
	st.state.AddBalance(coinrange, common.MainAccount, gasaddr, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)) //给对应币种奖励账户加钱
	return ret, 0, false, shardings, err
}
func (st *StateTransition) CallRevocableNormalTx() (ret []byte, usedGas uint64, failed bool, shardings []uint, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, shardings, errors.New("CallRevocableNormalTx from is nil")
	}
	usefrom := tx.From()
	if usefrom == addr {
		return nil, 0, false, shardings, errors.New("CallRevocableNormalTx usefrom is nil")
	}
	var (
		vmerr error
	)
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, shardings, err
	}
	mapTOAmonts := make([]common.AddrAmont, 0)
	//
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, shardings, ErrTXCountOverflow
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return nil, 0, false, shardings, err
			}
			//0.7+0.3*pow(0.9,(num-1))
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, shardings, err
	}
	st.state.SetNonce(st.msg.GetTxCurrency(), from, st.state.GetNonce(st.msg.GetTxCurrency(), from)+1)
	st.state.AddBalance(st.msg.GetTxCurrency(), common.WithdrawAccount, usefrom, st.value)
	st.state.SubBalance(st.msg.GetTxCurrency(), common.MainAccount, usefrom, st.value)
	shardings = append(shardings, uint(from[0]))
	shardings = append(shardings, uint(st.To()[0]))
	mapTOAmont := common.AddrAmont{Addr: st.To(), Amont: st.value}
	mapTOAmonts = append(mapTOAmonts, mapTOAmont)
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			st.state.AddBalance(st.msg.GetTxCurrency(), common.WithdrawAccount, usefrom, ex.Amount)
			st.state.SubBalance(st.msg.GetTxCurrency(), common.MainAccount, usefrom, ex.Amount)
			mapTOAmont = common.AddrAmont{Addr: *ex.Recipient, Amont: ex.Amount}
			shardings = append(shardings, uint(ex.Recipient[0]))
			mapTOAmonts = append(mapTOAmonts, mapTOAmont)
			if vmerr != nil {
				break
			}
		}
	}
	//costGas := new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, nil, vmerr
		}
	}
	var rt common.RecorbleTx
	rt.From = tx.From()
	rt.Tim = tx.GetCreateTime()
	rt.Typ = tx.GetMatrixType()
	rt.Cointyp = tx.GetTxCurrency()
	rt.Adam = append(rt.Adam, mapTOAmonts...)
	b, marshalerr := json.Marshal(&rt)
	if marshalerr != nil {
		return nil, 0, false, nil, marshalerr
	}
	txHash := tx.Hash()
	//log.Info("file state_transition","func CallRevocableNormalTx:txHash",txHash)
	mapHashamont := make(map[common.Hash][]byte)
	mapHashamont[txHash] = b
	st.state.SaveTx(st.msg.GetTxCurrency(), st.msg.From(), tx.GetMatrixType(), rt.Tim, mapHashamont)
	st.state.SetMatrixData(txHash, b)
	gasaddr, coinrange := st.getCoinAddress(tx.GetTxCurrency())
	st.RefundGas(coinrange)
	st.state.AddBalance(coinrange, common.MainAccount, gasaddr, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)) //给对应币种奖励账户加钱
	return ret, st.GasUsed(), vmerr != nil, shardings, err
}
func (st *StateTransition) CallUnGasNormalTx() (ret []byte, usedGas uint64, failed bool, shardings []uint, err error) {
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	toaddr := tx.To()
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, shardings, errors.New("CallUnGasNormalTx from is nil")
	}
	sender := vm.AccountRef(from)
	var (
		evm   = st.evm
		vmerr error
	)
	tmpshard := make([]uint, 0)
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, shardings, ErrTXCountOverflow
		}
	}
	st.gas = uint64(10000000)
	issendFromContract := false
	beforAmont := st.state.GetBalanceByType(st.msg.GetTxCurrency(), common.ContractAddress, common.MainAccount)
	interestbefor := st.state.GetBalanceByType(st.msg.GetTxCurrency(), common.InterestRewardAddress, common.MainAccount) // Test
	interset := big.NewInt(0)
	if toaddr == nil { //
		log.Error("state_transition callUnGasNormalTx to is nil")
		return nil, 0, false, shardings, ErrTXToNil
	} else {
		// Increment the nonce for the next transaction
		if st.To() == common.ContractAddress {
			interset = new(big.Int).Add(interset, st.value)
			issendFromContract = true
		}
		st.state.SetNonce(st.msg.GetTxCurrency(), tx.From(), st.state.GetNonce(st.msg.GetTxCurrency(), sender.Address())+1)
		ret, st.gas, tmpshard, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			if toaddr == nil {
				log.Error("state_transition callUnGasNormalTx Extro to is nil")
				return nil, 0, false, shardings, ErrTXToNil
			} else {
				if *ex.Recipient == common.ContractAddress {
					interset = new(big.Int).Add(interset, ex.Amount)
					issendFromContract = true
				}
				// Increment the nonce for the next transaction
				ret, st.gas, tmpshard, vmerr = evm.Call(sender, *ex.Recipient, ex.Payload, st.gas, ex.Amount)
			}
			if vmerr != nil {
				break
			}
		}
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, shardings, vmerr
		}
	}
	if issendFromContract {

		afterAmont := st.state.GetBalanceByType(st.msg.GetTxCurrency(), common.ContractAddress, common.MainAccount)
		difAmont := new(big.Int).Sub(afterAmont, beforAmont)
		if difAmont.Cmp(interset) != 0 {
			log.Info("state_transition", "rewardTx", "ContractAddress 余额与增加的钱不一致")
			return nil, 0, false, shardings, ErrinterestAmont
		}
		interestafter := st.state.GetBalanceByType(st.msg.GetTxCurrency(), common.InterestRewardAddress, common.MainAccount)
		dif := new(big.Int).Sub(interestbefor, interestafter)
		if difAmont.Cmp(dif) != 0 {
			log.Info("state_transition", "rewardTx", "InterestRewardAddress 余额与扣除的钱不一致")
			log.Error("state_transition", "difAmont", difAmont, "dif", dif, "afterAmont", afterAmont, "beforAmont", beforAmont, "interestafter", interestafter, "interestbefor", interestbefor)
			return nil, 0, false, shardings, ErrinterestAmont
		}
	}
	shardings = append(shardings, tmpshard...)
	return ret, 0, vmerr != nil, shardings, err
}

func (st *StateTransition) CallSetBlackListTx() (ret []byte, usedGas uint64, failed bool, shardings []uint, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, shardings, errors.New("state_transition,make coin ,from is nil")
	}
	sender := vm.AccountRef(from)
	data := tx.Data()
	var blacklist []string
	err = json.Unmarshal(data, &blacklist)
	if err != nil {
		log.Error("CallSetBlackListTx", "Unmarshal err", err)
		return nil, 0, false, shardings, err
	}
	st.gas = 0
	st.state.SetNonce(st.msg.GetTxCurrency(), tx.From(), st.state.GetNonce(st.msg.GetTxCurrency(), sender.Address())+1)

	file, err := os.OpenFile(common.WorkPath+"/blacklist.txt", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		file.Close()
		log.Error("CallSetBlackListTx", "OpenFile err", err)
		return nil, 0, false, shardings, err
	}
	var tmpBlackListString []string
	var tmpBlackList []common.Address
	writer := bufio.NewWriter(file)
	for _, black := range blacklist {
		addr, err := base58.Base58DecodeToAddress(black)
		if err != nil {
			log.Error("invalidate black", "black", black)
			continue
		}
		writer.WriteString(black)
		writer.WriteString("\x0D\x0A")
		writer.Flush()
		tmpBlackListString = append(tmpBlackListString, black)
		tmpBlackList = append(tmpBlackList, addr)
	}
	file.Close()
	if len(tmpBlackListString) > 0 || len(blacklist) == 0 {
		common.BlackListString = make([]string, 0, len(tmpBlackListString))
		common.BlackList = make([]common.Address, 0, len(tmpBlackList))
		common.BlackListString = append(common.BlackListString, tmpBlackListString...)
		common.BlackList = append(common.BlackList, tmpBlackList...)
	}
	gasaddr, coinrange := st.getCoinAddress(tx.GetTxCurrency())
	st.RefundGas(coinrange)
	st.state.AddBalance(coinrange, common.MainAccount, gasaddr, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)) //给对应币种奖励账户加钱
	return ret, st.GasUsed(), true, shardings, err
}

func (st *StateTransition) CallNormalTx() (ret []byte, usedGas uint64, failed bool, shardings []uint, err error) {
	if err = st.PreCheck(); err != nil {
		return
	}
	tx := st.msg //因为st.msg的接口全部在transaction中实现,所以此处的局部变量msg实际是transaction类型
	toaddr := tx.To()
	var addr common.Address
	from := tx.From()
	if from == addr {
		return nil, 0, false, shardings, errors.New("CallNormalTx from is nil")
	}
	sender := vm.AccountRef(from)
	var (
		evm   = st.evm
		vmerr error
	)
	tmpshard := make([]uint, 0)
	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data)
	if err != nil {
		return nil, 0, false, shardings, err
	}
	//
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, shardings, ErrTXCountOverflow
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return nil, 0, false, shardings, err
			}
			//0.7+0.3*pow(0.9,(num-1))
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, shardings, err
	}
	if toaddr == nil {
		var caddr common.Address
		ret, caddr, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
		tmpshard = append(tmpshard, uint(caddr[0]))
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(tx.GetTxCurrency(), from, st.msg.Nonce()+1) //st.state.GetNonce(tx.GetTxCurrency(), from)
		ret, st.gas, tmpshard, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			var caddr common.Address
			if toaddr == nil {
				ret, caddr, st.gas, vmerr = evm.Create(sender, ex.Payload, st.gas, ex.Amount)
				tmpshard = append(tmpshard, uint(caddr[0]))
			} else {
				// Increment the nonce for the next transaction
				ret, st.gas, tmpshard, vmerr = evm.Call(sender, *ex.Recipient, ex.Payload, st.gas, ex.Amount)
			}
			if vmerr != nil {
				break
			}

		}
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, nil, vmerr
		}
	}
	shardings = append(shardings, tmpshard...)
	gasaddr, coinrange := st.getCoinAddress(tx.GetTxCurrency())
	st.RefundGas(coinrange)
	st.state.AddBalance(coinrange, common.MainAccount, gasaddr, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)) //给对应币种奖励账户加钱
	return ret, st.GasUsed(), vmerr != nil, shardings, err
}

//授权交易的from和to是同一个地址
func (st *StateTransition) CallAuthTx() (ret []byte, usedGas uint64, failed bool, shardings []uint, err error) {
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
		return nil, 0, false, shardings, err
	}
	//
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, shardings, ErrTXCountOverflow
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return nil, 0, false, shardings, err
			}
			//0.7+0.3*pow(0.9,(num-1))
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, shardings, err
	}
	tmpshard := make([]uint, 0)
	if toaddr == nil { //
		log.Error("state_transition callAuthTx to is nil")
		return nil, 0, false, shardings, ErrTXToNil
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(tx.GetTxCurrency(), tx.From(), st.state.GetNonce(tx.GetTxCurrency(), sender.Address())+1)
		ret, st.gas, tmpshard, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			if toaddr == nil {
				log.Error("state_transition callAuthTx extro to is nil")
				return nil, 0, false, shardings, ErrTXToNil
			} else {
				// Increment the nonce for the next transaction
				ret, st.gas, tmpshard, vmerr = evm.Call(sender, *ex.Recipient, ex.Payload, st.gas, ex.Amount)
			}
			if vmerr != nil {
				break
			}
		}
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, shardings, vmerr
		}
	}
	shardings = append(shardings, tmpshard...)
	gasaddr, coinrange := st.getCoinAddress(tx.GetTxCurrency())
	st.RefundGas(coinrange)
	st.state.AddBalance(coinrange, common.MainAccount, gasaddr, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)) //给对应币种奖励账户加钱

	isModiEntrustCount := false
	entrustOK := false
	Authfrom := tx.From()
	EntrustList := make([]common.EntrustType, 0)
	err = json.Unmarshal(tx.Data(), &EntrustList) //EntrustList为被委托人的EntrustType切片
	if err != nil {
		log.Error("CallAuthTx Unmarshal err")
		return nil, st.GasUsed(), true, shardings, nil
	}

	HeightAuthDataList := make([]common.AuthType, 0) //按高度存储授权数据列表
	TimeAuthDataList := make([]common.AuthType, 0)   //按时间存储授权数据列表
	CountAuthDataList := make([]common.AuthType, 0)  //按次数存储授权数据列表
	for _, EntrustData := range EntrustList {
		str_addres := EntrustData.EntrustAddres //被委托人地址
		addres, err := base58.Base58DecodeToAddress(str_addres)
		if err != nil {
			return nil, st.GasUsed(), true, shardings, nil
		}
		tCoin := strings.Split(str_addres, ".")[0]
		if tx.GetTxCurrency() != tCoin {
			log.Error("不能跨币种委托", "当前币种", tx.GetTxCurrency(), "委托币种", tCoin)
			return nil, st.GasUsed(), true, shardings, nil
		}
		tmpAuthMarsha1Data := st.state.GetAuthStateByteArray(tx.GetTxCurrency(), addres) //获取授权数据
		if len(tmpAuthMarsha1Data) != 0 {
			//AuthData := new(common.AuthType)
			AuthDataList := make([]common.AuthType, 0)
			err = json.Unmarshal(tmpAuthMarsha1Data, &AuthDataList)
			if err != nil {
				log.Error("CallAuthTx AuthDataList Unmarshal err")
				return nil, st.GasUsed(), true, shardings, nil
			}
			for _, AuthData := range AuthDataList {
				if AuthData.IsEntrustGas == false && AuthData.IsEntrustSign == false {
					continue
				}
				if AuthData.AuthAddres != (common.Address{}) && !(AuthData.AuthAddres.Equal(Authfrom)) {
					//如果不是同一个人授权，先判断之前的授权人权限是否失效，如果之前的授权权限没失效则不能被重复委托
					if AuthData.EnstrustSetType == params.EntrustByHeight {
						if st.evm.BlockNumber.Uint64() <= AuthData.EndHeight {
							//按高度委托未失效
							log.Error("该委托人已经被委托过了，不能重复委托", "from", tx.From(), "Nonce", tx.Nonce())
							return nil, st.GasUsed(), true, shardings, nil //如果一个不满足就返回，不continue
						}
					} else if AuthData.EnstrustSetType == params.EntrustByTime {
						if st.evm.Time.Uint64() <= AuthData.EndTime {
							//按时间委托未失效
							log.Error("该委托人已经被委托过了，不能重复委托", "from", tx.From(), "Nonce", tx.Nonce())
							return nil, st.GasUsed(), true, shardings, nil //如果一个不满足就返回，不continue
						}
					} else if AuthData.EnstrustSetType == params.EntrustByCount {
						if AuthData.EntrustCount > 0 {
							log.Error("该委托人已经被委托过了，不能重复委托", "from", tx.From(), "Nonce", tx.Nonce())
							return nil, st.GasUsed(), true, shardings, nil //如果一个不满足就返回，不continue
						}
					} else {
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
							return nil, st.GasUsed(), true, shardings, nil
						}
						HeightAuthDataList = append(HeightAuthDataList, AuthData)
					} else if EntrustData.EnstrustSetType == params.EntrustByTime {
						//按时间委托
						if EntrustData.StartTime <= AuthData.EndTime {
							log.Error("同一个授权人的委托时间不能重合", "from", tx.From(), "Nonce", tx.Nonce())
							return nil, st.GasUsed(), true, shardings, nil
						}
						TimeAuthDataList = append(TimeAuthDataList, AuthData)
					} else if EntrustData.EnstrustSetType == params.EntrustByCount {
						//读取以前的按次数授权是同一个人，则修改以前的授权次数
						for index, oldAuthData := range CountAuthDataList {
							if oldAuthData.AuthAddres.Equal(addres) {
								AuthData.EntrustCount = EntrustData.EntrustCount
								CountAuthDataList[index] = AuthData
								isModiEntrustCount = true
							}
						}
					} else {
						log.Error("未设置委托类型", "from", tx.From(), "Nonce", tx.Nonce())
						return nil, st.GasUsed(), true, shardings, nil
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
			t_authData.EntrustCount = EntrustData.EntrustCount
			HeightAuthDataList = append(HeightAuthDataList, *t_authData)
			marshalAuthData, err := json.Marshal(HeightAuthDataList)
			if err != nil {
				log.Error("Marshal err")
				return nil, st.GasUsed(), true, shardings, nil
			}
			//marsha1AuthData是authData的Marsha1编码
			st.state.SetAuthStateByteArray(tx.GetTxCurrency(), addres, marshalAuthData) //设置授权数据
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
			t_authData.EntrustCount = EntrustData.EntrustCount
			TimeAuthDataList = append(TimeAuthDataList, *t_authData)
			marshalAuthData, err := json.Marshal(TimeAuthDataList)
			if err != nil {
				log.Error("Marshal err")
				return nil, st.GasUsed(), true, shardings, nil
			}
			//marsha1AuthData是authData的Marsha1编码
			st.state.SetAuthStateByteArray(tx.GetTxCurrency(), addres, marshalAuthData) //设置授权数据
		}

		if EntrustData.EnstrustSetType == params.EntrustByCount {
			//按次数委托
			if !isModiEntrustCount {
				t_authData := new(common.AuthType)
				t_authData.EnstrustSetType = EntrustData.EnstrustSetType
				t_authData.StartTime = EntrustData.StartTime
				t_authData.EndTime = EntrustData.EndTime
				t_authData.IsEntrustSign = EntrustData.IsEntrustSign
				t_authData.IsEntrustGas = EntrustData.IsEntrustGas
				t_authData.AuthAddres = Authfrom
				t_authData.EntrustCount = EntrustData.EntrustCount
				CountAuthDataList = append(CountAuthDataList, *t_authData)
				isModiEntrustCount = false
			}
			marshalAuthData, err := json.Marshal(CountAuthDataList)
			if err != nil {
				log.Error("Marshal err")
				return nil, st.GasUsed(), true, shardings, nil
			}
			//marsha1AuthData是authData的Marsha1编码
			st.state.SetAuthStateByteArray(tx.GetTxCurrency(), addres, marshalAuthData) //设置授权数据
		}
	}
	if entrustOK {
		//获取之前的委托数据(结构体切片经过marshal编码)
		AllEntrustList := make([]common.EntrustType, 0)
		oldEntrustList := st.state.GetEntrustStateByteArray(tx.GetTxCurrency(), Authfrom) //获取委托数据
		if len(oldEntrustList) != 0 {
			err = json.Unmarshal(oldEntrustList, &AllEntrustList)
			if err != nil {
				log.Error("CallAuthTx Unmarshal err")
				return nil, st.GasUsed(), true, shardings, nil
			}
		}

		//遍历之前的委托数据，看是否有同一个人按次数授权的，如果有，直接修改之前的次数
		isHave := false
		tmpEntrustList := make([]common.EntrustType, 0)
		for _, newEntrustData := range EntrustList {
			if newEntrustData.EnstrustSetType == params.EntrustByCount {
				for i, oldEntrustData := range AllEntrustList {
					if oldEntrustData.EntrustAddres == newEntrustData.EntrustAddres && oldEntrustData.EnstrustSetType == params.EntrustByCount {
						AllEntrustList[i].EntrustCount = newEntrustData.EntrustCount
						isHave = true
						break
					}
				}
			}
			if !isHave {
				tmpEntrustList = append(tmpEntrustList, newEntrustData)
			}
		}

		AllEntrustList = append(AllEntrustList, tmpEntrustList...)
		allDataList, err := json.Marshal(AllEntrustList)
		if err != nil {
			log.Error("Marshal error")
		}
		st.state.SetEntrustStateByteArray(tx.GetTxCurrency(), Authfrom, allDataList) //设置委托数据
		entrustOK = false
	} else {
		log.Error("委托条件不满足")
		return nil, st.GasUsed(), true, shardings, nil
	}

	return ret, st.GasUsed(), vmerr != nil, shardings, nil
}

func isContain(a uint32, list []uint32) bool {
	for _, data := range list {
		if data == a {
			return true
		}
	}
	return false
}

func (st *StateTransition) CallCancelAuthTx() (ret []byte, usedGas uint64, failed bool, shardings []uint, err error) {
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
		return nil, 0, false, shardings, err
	}
	//
	tmpExtra := tx.GetMatrix_EX() //Extra()
	if (&tmpExtra) != nil && len(tmpExtra) > 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return nil, 0, false, shardings, ErrTXCountOverflow
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return nil, 0, false, shardings, err
			}
			//0.7+0.3*pow(0.9,(num-1))
			gas += tmpgas
		}
	}
	if err = st.UseGas(gas); err != nil {
		return nil, 0, false, shardings, err
	}
	tmpshard := make([]uint, 0)
	if toaddr == nil { //
		log.Error("state transition callAuthTx to is nil")
		return nil, 0, false, shardings, ErrTXToNil
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(tx.GetTxCurrency(), tx.From(), st.state.GetNonce(tx.GetTxCurrency(), sender.Address())+1)
		ret, st.gas, tmpshard, vmerr = evm.Call(sender, st.To(), st.data, st.gas, st.value)
	}
	if vmerr == nil && (&tmpExtra) != nil && len(tmpExtra) > 0 {
		for _, ex := range tmpExtra[0].ExtraTo {
			if toaddr == nil {
				log.Error("state transition callAuthTx Extro to is nil")
				return nil, 0, false, shardings, ErrTXToNil
			} else {
				// Increment the nonce for the next transaction
				ret, st.gas, tmpshard, vmerr = evm.Call(sender, *ex.Recipient, ex.Payload, st.gas, ex.Amount)
			}
			if vmerr != nil {
				break
			}
		}
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, shardings, vmerr
		}
	}
	shardings = append(shardings, tmpshard...)
	gasaddr, coinrange := st.getCoinAddress(tx.GetTxCurrency())
	st.RefundGas(coinrange)
	st.state.AddBalance(coinrange, common.MainAccount, gasaddr, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.gasPrice)) //给对应币种奖励账户加钱

	Authfrom := tx.From()
	delIndexList := make([]uint32, 0)
	err = json.Unmarshal(tx.Data(), &delIndexList) //EntrustList为被委托人的EntrustType切片
	if err != nil {
		log.Error("CallAuthTx Unmarshal err")
		return nil, st.GasUsed(), true, shardings, nil
	}
	EntrustMarsha1Data := st.state.GetEntrustStateByteArray(tx.GetTxCurrency(), Authfrom) //获取委托数据
	if len(EntrustMarsha1Data) == 0 {
		log.Error("没有委托数据")
		return nil, st.GasUsed(), true, shardings, nil
	}
	allentrustDataList := make([]common.EntrustType, 0)
	err = json.Unmarshal(EntrustMarsha1Data, &allentrustDataList)
	if err != nil {
		log.Error("CallAuthTx Unmarshal err")
		return nil, st.GasUsed(), true, shardings, nil
	}

	entrustDataList := make([]common.EntrustType, 0, len(allentrustDataList))
	for _, entrustdata := range allentrustDataList {
		if entrustdata.EnstrustSetType == params.EntrustByTime && st.evm.Time.Uint64() > entrustdata.EndTime {
			continue
		}
		if entrustdata.EnstrustSetType == params.EntrustByHeight && st.evm.BlockNumber.Uint64() > entrustdata.EndHeight {
			continue
		}
		if entrustdata.EnstrustSetType == params.EntrustByCount && entrustdata.EntrustCount <= 0 {
			continue
		}
		entrustDataList = append(entrustDataList, entrustdata)
	}

	newentrustDataList := make([]common.EntrustType, 0)
	for index, entrustFrom := range entrustDataList {
		if isContain(uint32(index), delIndexList) {
			//要删除的切片数据
			str_addres := entrustFrom.EntrustAddres //被委托人地址
			addres, err := base58.Base58DecodeToAddress(str_addres)
			if err != nil {
				return nil, st.GasUsed(), true, shardings, nil
			}
			marshaldata := st.state.GetAuthStateByteArray(tx.GetTxCurrency(), addres) //获取之前的授权数据切片,marshal编码过的  //获取授权数据
			if len(marshaldata) > 0 {
				//oldAuthData := new(common.AuthType)   //oldAuthData的地址为0x地址
				oldAuthDataList := make([]common.AuthType, 0)
				err = json.Unmarshal(marshaldata, &oldAuthDataList) //oldAuthData的地址为0x地址
				if err != nil {
					return nil, st.GasUsed(), true, shardings, nil
				}
				newAuthDataList := make([]common.AuthType, 0)
				for _, oldAuthData := range oldAuthDataList {
					//只要起始高度或时间能对应上，就是要删除的切片
					if entrustFrom.EnstrustSetType == oldAuthData.EnstrustSetType {
						if entrustFrom.StartHeight == oldAuthData.StartHeight || entrustFrom.StartTime == oldAuthData.StartTime || entrustFrom.EntrustCount == oldAuthData.EntrustCount {
							oldAuthData.IsEntrustGas = false
							oldAuthData.IsEntrustSign = false
							continue
						}
					}
					newAuthDataList = append(newAuthDataList, oldAuthData)
				}
				newAuthDatalist, err := json.Marshal(newAuthDataList)
				if err != nil {
					return nil, st.GasUsed(), true, shardings, nil
				}
				st.state.SetAuthStateByteArray(tx.GetTxCurrency(), addres, newAuthDatalist) //设置授权数据
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
	st.state.SetEntrustStateByteArray(tx.GetTxCurrency(), Authfrom, newEntrustList) //设置委托数据

	return ret, st.GasUsed(), vmerr != nil, shardings, nil
}

func (st *StateTransition) RefundGas(coinrange string) {
	// Apply refund counter, capped to half of the used gas.
	refund := st.GasUsed() / 2
	if refund > st.state.GetRefund(coinrange, st.msg.From()) {
		refund = st.state.GetRefund(coinrange, st.msg.From())
	}
	st.gas += refund
	if st.gas == 0 {
		return
	}
	// Return ETH for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.state.AddBalance(coinrange, common.MainAccount, st.msg.AmontFrom(), remaining)
	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) GasUsed() uint64 {
	return st.initialGas - st.gas
}
