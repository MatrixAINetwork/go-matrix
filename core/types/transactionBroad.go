// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package types

import (
	"errors"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"io"
	"math/big"
	"sync/atomic"
	"time"
)

type TransactionBroad struct {
	data txdata
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

//
func NewBroadCastTransaction(txType byte, data []byte) *TransactionBroad {
	return newBroadCastTransaction(txType, data)
}

// EncodeRLP implements rlp.Encoder
func (tx *TransactionBroad) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &tx.data)
}

// DecodeRLP implements rlp.Decoder
func (tx *TransactionBroad) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}
	return err
}
func (tx *TransactionBroad) GetTxN(index int) uint32 {
	return 0
}

// 广播交易
func newBroadCastTransaction(txType byte, data []byte) *TransactionBroad {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	mx := Matrix_Extra{
		TxType: txType,
	}
	d := txdata{
		AccountNonce: 0,
		Recipient:    &common.Address{},
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     0,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
		TxEnterType:  BroadCastTxIndex,
		Extra:        make([]Matrix_Extra, 0),
	}

	d.Amount.Set(big.NewInt(0))
	d.Price.Set(big.NewInt(12))

	d.Extra = append(d.Extra, mx)
	tx := &TransactionBroad{data: d}
	return tx
}

func (tx *TransactionBroad) SetCoinType(typ string) {}
func (tx *TransactionBroad) TxType() byte           { return tx.data.TxEnterType }

func (tx *TransactionBroad) Data() []byte       { return common.CopyBytes(tx.data.Payload) }
func (tx *TransactionBroad) Gas() uint64        { return tx.data.GasLimit }
func (tx *TransactionBroad) GasPrice() *big.Int { return new(big.Int).Set(tx.data.Price) }
func (tx *TransactionBroad) Value() *big.Int    { return new(big.Int).Set(tx.data.Amount) }
func (tx *TransactionBroad) Nonce() uint64      { return tx.data.AccountNonce }
func (tx *TransactionBroad) CheckNonce() bool   { return true }
func (tx *TransactionBroad) ChainId() *big.Int {
	return deriveChainId(tx.data.V)
}
func (tx *TransactionBroad) GetMakeHashfield(chid *big.Int) []interface{} {
	return []interface{}{
		tx.data.AccountNonce,
		tx.data.Price,
		tx.data.GasLimit,
		tx.data.Recipient,
		tx.data.Amount,
		tx.data.Payload,
		tx.data.Extra,
		chid, uint(0), uint(0),
	}
}
func (tx *TransactionBroad) IsEntrustTx() bool            { return tx.data.IsEntrustTx == 1 }
func (tx *TransactionBroad) Setentrustfrom(x interface{}) {}
func (tx *TransactionBroad) GasFrom() common.Address {
	return common.Address{}
}
func (tx *TransactionBroad) AmontFrom() common.Address {
	return common.Address{}
}
func (tx *TransactionBroad) GetMatrixType() byte {
	return 1
}

//
func (tx *TransactionBroad) From() common.Address {
	return common.Address{}
}

func (tx *TransactionBroad) GetCreateTime() uint32 {
	return uint32(tx.data.CommitTime)
}

func (tx *TransactionBroad) GetLocalHeight() uint32 {
	if tx.data.Extra != nil && len(tx.data.Extra) > 0 {
		return uint32(tx.data.Extra[0].LockHeight)
	}
	return uint32(time.Now().Unix())
}
func (tx *TransactionBroad) SetTxV(v *big.Int) { tx.data.V = v }
func (tx *TransactionBroad) SetTxR(r *big.Int) { tx.data.R = r }

//
func (tx *TransactionBroad) GetTxFrom() (common.Address, error) {
	if tx.from.Load() == nil {
		//如果交易没有做过验签则err不为空。
		return common.Address{}, errors.New("Address is Nil")
	}
	//如果交易做过验签则err为空。
	return tx.from.Load().(sigCache).from, nil
}
func (tx *TransactionBroad) GetFromLoad() interface{} {
	return tx.from.Load()
}
func (tx *TransactionBroad) SetFromLoad(x interface{}) {
	tx.from.Store(x)
}

func (tx *TransactionBroad) SetTxCurrency(currency string) {

}
func (tx *TransactionBroad) GetTxCurrency() string {
	return params.MAN_COIN
}

//
func (tx *TransactionBroad) GetMatrix_EX() []Matrix_Extra { return tx.data.Extra }

//
func (tx *TransactionBroad) GetTxV() *big.Int { return tx.data.V }
func (tx *TransactionBroad) GetTxR() *big.Int { return tx.data.R }

//
func (tx *TransactionBroad) GetTxS() *big.Int { return tx.data.S }
func (tx *TransactionBroad) GetTxNLen() int {
	return 0
}

// 在传递交易时用来操作Nonce
func (tx *TransactionBroad) SetNonce(nc uint64) {
	tx.data.AccountNonce = nc
}

func (tx *TransactionBroad) GetIsEntrustGas() bool {
	return false
}

func (tx *TransactionBroad) GetIsEntrustByTime() bool {
	return false
}
func (tx *TransactionBroad) SetIsEntrustGas(b bool) {

}
func (tx *TransactionBroad) SetIsEntrustByTime(b bool) {
}
func (tx *TransactionBroad) SetIsEntrustByCount(b bool) {
}
func (tx *TransactionBroad) GetIsEntrustByCount() bool {
	return false
}

//
func (tx *TransactionBroad) SetTxS(S *big.Int) { tx.data.S = S }
func (tx *TransactionBroad) To() *common.Address {
	return tx.data.Recipient
	//if tx.data.Recipient == nil {
	//	return nil
	//}
	//to := *tx.data.Recipient
	//return &to
}

// WithSignature returns a new transaction with the given signature.
// This signature needs to be formatted as described in the yellow paper (v+27).
func (tx *TransactionBroad) WithSignature(signer Signer, sig []byte) (SelfTransaction, error) {
	r, s, v, err := signer.SignatureValues(tx, sig)
	if err != nil {
		return nil, err
	}
	cpy := &TransactionBroad{data: tx.data}
	cpy.data.R, cpy.data.S, cpy.data.V = r, s, v
	////
	//if len(cpy.data.Extra) > 0 {
	//	cpy.data.V.Add(cpy.data.V, big.NewInt(128))
	//}
	return cpy, nil
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *TransactionBroad) Hash() common.Hash {
	v := rlpHash(tx)
	return v
}

func (tx *TransactionBroad) GetTxHashStruct() {

}
func (tx *TransactionBroad) Call() error {
	return nil
}
func (tx *TransactionBroad) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &tx.data)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

////
//func (tx *TransactionBroad) SetTransactionMx(tx_Mx *Transaction_Mx)(txer SelfTransaction ){
//	if tx_Mx == nil{
//		return nil
//	}
//
//	tx.data.AccountNonce=tx_Mx.Data.AccountNonce
//	tx.data.Price=tx_Mx.Data.Price
//	tx.data.GasLimit=tx_Mx.Data.GasLimit
//	tx.data.Recipient=tx_Mx.Data.Recipient
//	tx.data.Amount=tx_Mx.Data.Amount
//	tx.data.Payload=tx_Mx.Data.Payload
//	// Signature values
//	tx.data.V=tx_Mx.Data.V
//	tx.data.R=tx_Mx.Data.R
//	tx.data.S=tx_Mx.Data.S
//	tx.data.TxEnterType=BroadCastTxIndex
//	tx.data.Extra=tx_Mx.Data.Extra
//
//	mx := Matrix_Extra{
//		TxType: tx_Mx.TxType_Mx,
//	}
//	tx.data.Extra = append(tx.data.Extra, mx)
//	txa := &TransactionBroad{data: tx.data}
//	txer = txa
//	return
//}
//
////
//func (tx *TransactionBroad)GetTransactionMx(stx SelfTransaction) *Transaction_Mx {
//	btx,ok:=stx.(*TransactionBroad)
//	if !ok {
//		return nil
//	}
//	tx_Mx := &Transaction_Mx{}
//	tx_Mx.Data.AccountNonce = btx.data.AccountNonce
//	tx_Mx.Data.Price = btx.data.Price
//	tx_Mx.Data.GasLimit = btx.data.GasLimit
//	tx_Mx.Data.Recipient = btx.data.Recipient
//	tx_Mx.Data.Amount = btx.data.Amount
//	tx_Mx.Data.Payload = btx.data.Payload
//	// Signature values
//	tx_Mx.Data.V = btx.data.V
//	tx_Mx.Data.R = btx.data.R
//	tx_Mx.Data.S = btx.data.S
//	tx_Mx.Data.Extra = btx.data.Extra
//	tx_Mx.Data.TxEnterType = btx.data.TxEnterType
//	if len(btx.data.Extra) > 0 {
//		tx_Mx.TxType_Mx = btx.data.Extra[0].TxType
//	}
//	return tx_Mx
//}
//
func SetTransactionMx(tx_Mx *Transaction_Mx) *TransactionBroad {
	if tx_Mx == nil {
		return nil
	}
	tx := txdata{
		AccountNonce: tx_Mx.Data.AccountNonce,
		Price:        tx_Mx.Data.Price,
		GasLimit:     tx_Mx.Data.GasLimit,
		Recipient:    tx_Mx.Data.Recipient,
		Amount:       tx_Mx.Data.Amount,
		Payload:      tx_Mx.Data.Payload,
		// Signature values
		V:           tx_Mx.Data.V,
		R:           tx_Mx.Data.R,
		S:           tx_Mx.Data.S,
		TxEnterType: BroadCastTxIndex,
		Extra:       tx_Mx.Data.Extra,
	}
	if len(tx.Extra) == 0 {
		mx := Matrix_Extra{
			TxType: tx_Mx.TxType_Mx,
		}
		tx.Extra = append(tx.Extra, mx)
	}

	return &TransactionBroad{data: tx}
}

//
func GetTransactionMx(stx SelfTransaction) *Transaction_Mx {
	tx, ok := stx.(*TransactionBroad)
	if !ok {
		return nil
	}
	tx_Mx := &Transaction_Mx{}
	tx_Mx.Data.AccountNonce = tx.data.AccountNonce
	tx_Mx.Data.Price = tx.data.Price
	tx_Mx.Data.GasLimit = tx.data.GasLimit
	tx_Mx.Data.Recipient = tx.data.Recipient
	tx_Mx.Data.Amount = tx.data.Amount
	tx_Mx.Data.Payload = tx.data.Payload
	// Signature values
	tx_Mx.Data.V = tx.data.V
	tx_Mx.Data.R = tx.data.R
	tx_Mx.Data.S = tx.data.S
	tx_Mx.Data.Extra = tx.data.Extra
	tx_Mx.Data.TxEnterType = tx.data.TxEnterType
	if len(tx.data.Extra) > 0 {
		tx_Mx.TxType_Mx = tx.data.Extra[0].TxType
	}
	return tx_Mx
}

func (tx *TransactionBroad) RawSignatureValues() (*big.Int, *big.Int, *big.Int) {
	return tx.data.V, tx.data.R, tx.data.S
}
func (tx *TransactionBroad) Protected() bool {
	return isProtectedV(tx.data.V)
}
func (tx *TransactionBroad) GetConstructorType() uint16 {
	return uint16(BroadCastTxIndex)
}
