// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package types

import (
	"container/heap"
	"errors"
	"github.com/MatrixAINetwork/go-matrix/base58"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"io"
	"math/big"
	"sync/atomic"
	"time"
)

//go:generate gencodec -type txdata -field-override txdataMarshaling -out gen_tx_json.go

var (
	ErrInvalidSig = errors.New("invalid transaction v, r, s values")
)

func init() {
	rlp.InterfaceConstructorMap[uint16(NormalTxIndex)] = func() interface{} {
		return &Transaction{}
	}
	rlp.InterfaceConstructorMap[uint16(BroadCastTxIndex)] = func() interface{} {
		return &TransactionBroad{}
	}
}

type Transaction struct {
	data txdata
	// caches
	hash        atomic.Value
	size        atomic.Value
	from        atomic.Value
	entrustfrom atomic.Value
	Currency    string //币种
	// by
	N                []uint32
	IsEntrustGas     bool
	IsEntrustByTime  bool //是否是按时间委托
	IsEntrustByCount bool //是否按次数委托
}
type TransactionCall struct {
	*Transaction
}

func (tc *TransactionCall) CheckNonce() bool { return false }

//
type Transaction_Mx struct {
	Data       txdata
	Currency   string
	TxType_Mx  byte
	LockHeight uint64  `json:"lockHeight" gencodec:"required"`
	ExtraTo    []Tx_to `json:"extra_to" gencodec:"required"`
}

//
type ExtraTo_tr struct {
	To_tr    *common.Address `json:"to"`
	Value_tr *hexutil.Big    `json:"value"`
	Input_tr *hexutil.Bytes  `json:"input"`
}

//
type Tx_to struct {
	Recipient *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount    *big.Int        `json:"value"    gencodec:"required"`
	Payload   []byte          `json:"input"    gencodec:"required"`
}

//
type Matrix_Extra struct {
	TxType     byte    `json:"txType" gencodec:"required"`
	LockHeight uint64  `json:"lockHeight" gencodec:"required"`
	ExtraTo    []Tx_to ` rlp:"tail"` //
}

// 用于洪泛（传输）
type Floodtxdata struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`
	Currency     string
	// Signature values
	V           *big.Int       `json:"v" gencodec:"required"`
	R           *big.Int       `json:"r" gencodec:"required"`
	TxEnterType byte           `json:"TxEnterType" gencodec:"required"` //是否是委托
	IsEntrustTx byte           `json:"IsEntrustTx" gencodec:"required"` //是否是委托
	CommitTime  uint64         `json:"CommitTime" gencodec:"required"`  //创建交易时间
	Extra       []Matrix_Extra ` rlp:"tail"`
}

type txdata struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash        *common.Hash   `json:"hash" rlp:"-"`
	TxEnterType byte           `json:"TxEnterType" gencodec:"required"` //入池类型
	IsEntrustTx byte           `json:"IsEntrustTx" gencodec:"required"` //是否是委托
	CommitTime  uint64         `json:"CommitTime" gencodec:"required"`  //创建交易时间
	Extra       []Matrix_Extra ` rlp:"tail"`                            //
}

func TxdataAddresToString(currency string, data *txdata, data1 *txdata1) {
	data1.AccountNonce = data.AccountNonce
	data1.Price = data.Price
	data1.GasLimit = data.GasLimit
	data1.Amount = data.Amount
	data1.Payload = data.Payload
	data1.V = data.V
	data1.R = data.R
	data1.S = data.S
	data1.Hash = data.Hash
	data1.TxEnterType = data.TxEnterType
	data1.IsEntrustTx = data.IsEntrustTx
	data1.CommitTime = data.CommitTime
	if data.Recipient != nil {
		data1.Recipient = new(string)
		to := *data.Recipient
		*data1.Recipient = base58.Base58EncodeToString(currency, to)
	}
	if len(data.Extra) > 0 {
		tmpEx1 := make([]Matrix_Extra1, 0)
		for _, er := range data.Extra {
			tmpEr1 := new(Matrix_Extra1)
			tmpEr1.TxType = er.TxType
			tmpEr1.LockHeight = er.LockHeight
			exto := make([]Tx_to1, 0)
			if len(er.ExtraTo) > 0 {
				for _, tto := range er.ExtraTo {
					tmTo := new(Tx_to1)
					if tto.Recipient != nil {
						tmTo.Recipient = new(string)
						*tmTo.Recipient = base58.Base58EncodeToString(currency, *tto.Recipient)
					}
					tmTo.Payload = tto.Payload
					tmTo.Amount = tto.Amount
					exto = append(exto, *tmTo)
				}
			}
			tmpEr1.ExtraTo = exto
			tmpEx1 = append(tmpEx1, *tmpEr1)
		}
		data1.Extra = tmpEx1
	}
}
func TxdataStringToAddres(data1 *txdata1, data *txdata) {
	data.AccountNonce = data1.AccountNonce
	data.Price = data1.Price
	data.GasLimit = data1.GasLimit
	data.Amount = data1.Amount
	data.Payload = data1.Payload
	data.V = data1.V
	data.R = data1.R
	data.S = data1.S
	data.Hash = data1.Hash
	data.TxEnterType = data1.TxEnterType
	data.IsEntrustTx = data1.IsEntrustTx
	data.CommitTime = data1.CommitTime
	if data1.Recipient != nil {
		data.Recipient = new(common.Address)
		*data.Recipient, _ = base58.Base58DecodeToAddress(*data1.Recipient)
	}

	if len(data1.Extra) > 0 {
		tmpEx1 := make([]Matrix_Extra, 0)
		for _, er := range data1.Extra {
			tmpEr1 := new(Matrix_Extra)
			tmpEr1.TxType = er.TxType
			tmpEr1.LockHeight = er.LockHeight
			exto := make([]Tx_to, 0)
			if len(er.ExtraTo) > 0 {
				for _, tto := range er.ExtraTo {
					tmTo := new(Tx_to)
					if tto.Recipient != nil {
						tmTo.Recipient = new(common.Address)
						*tmTo.Recipient, _ = base58.Base58DecodeToAddress(*tto.Recipient)
					}
					tmTo.Payload = tto.Payload
					tmTo.Amount = tto.Amount
					exto = append(exto, *tmTo)
				}
			}
			tmpEr1.ExtraTo = exto
			tmpEx1 = append(tmpEx1, *tmpEr1)
		}
		data.Extra = tmpEx1
	}
}

type Tx_to1 struct {
	Recipient *string  `json:"to"       rlp:"nil"` // nil means contract creation
	Amount    *big.Int `json:"value"    gencodec:"required"`
	Payload   []byte   `json:"input"    gencodec:"required"`
}
type Matrix_Extra1 struct {
	TxType     byte     `json:"txType" gencodec:"required"`
	LockHeight uint64   `json:"lockHeight" gencodec:"required"`
	ExtraTo    []Tx_to1 `json:"extra_to" gencodec:"required"`
}

//to地址为string类型
type txdata1 struct {
	AccountNonce uint64   `json:"nonce"    gencodec:"required"`
	Price        *big.Int `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64   `json:"gas"      gencodec:"required"`
	Recipient    *string  `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int `json:"value"    gencodec:"required"`
	Payload      []byte   `json:"input"    gencodec:"required"`

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`
	// This is only used when marshaling to JSON.
	Hash        *common.Hash    `json:"hash" rlp:"-"`
	TxEnterType byte            `json:"TxEnterType" gencodec:"required"` //入池类型
	IsEntrustTx byte            `json:"IsEntrustTx" gencodec:"required"` //是否是委托
	CommitTime  uint64          `json:"CommitTime" gencodec:"required"`  //创建交易时间
	Extra       []Matrix_Extra1 ` rlp:"tail"`                            //
}

type txdataMarshaling struct {
	AccountNonce hexutil.Uint64
	Price        *hexutil.Big
	GasLimit     hexutil.Uint64
	Amount       *hexutil.Big
	Payload      hexutil.Bytes
	V            *hexutil.Big
	R            *hexutil.Big
	S            *hexutil.Big
}

func NewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, V *big.Int, R *big.Int, S *big.Int, typ byte, isEntrustTx byte, currency string, committime uint64) *Transaction {
	return newTransaction(nonce, &to, amount, gasLimit, gasPrice, data, V, R, S, typ, isEntrustTx, currency, committime)
}

func NewContractCreation(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, V *big.Int, R *big.Int, S *big.Int, typ byte, isEntrustTx byte, currency string, committime uint64) *Transaction {
	return newTransaction(nonce, nil, amount, gasLimit, gasPrice, data, V, R, S, typ, isEntrustTx, currency, committime)
}

//
func NewTransactions(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, V *big.Int, R *big.Int, S *big.Int, ex []*ExtraTo_tr, localtime uint64, txType byte, isEntrustTx byte, currency string, committime uint64) *Transaction {
	return newTransactions(nonce, &to, amount, gasLimit, gasPrice, data, V, R, S, ex, localtime, txType, isEntrustTx, currency, committime)
}

//
func newTransactions(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, V *big.Int, R *big.Int, S *big.Int, ex []*ExtraTo_tr, localtime uint64, txType byte, isEntrustTx byte, currency string, committime uint64) *Transaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
		TxEnterType:  NormalTxIndex,
		IsEntrustTx:  isEntrustTx,
		CommitTime:   committime,
		Extra:        make([]Matrix_Extra, 0),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}
	if V != nil {
		d.V.Set(V)
	}
	if R != nil {
		d.R.Set(R)
	}
	if S != nil {
		d.S.Set(S)
	}
	//
	matrixEx := new(Matrix_Extra)
	arrayTx := make([]Tx_to, 0)
	if len(ex) > 0 {
		for _, extro := range ex {
			var input []byte
			if extro.Input_tr == nil {
				input = make([]byte, 0)
			} else {
				input = *extro.Input_tr
			}
			txto := new(Tx_to)
			txto.Amount = (*big.Int)(extro.Value_tr)
			txto.Recipient = extro.To_tr
			txto.Payload = input
			arrayTx = append(arrayTx, *txto)
		}
	}
	d.CommitTime = committime
	matrixEx.TxType = txType
	matrixEx.LockHeight = localtime
	matrixEx.ExtraTo = arrayTx
	d.Extra = append(d.Extra, *matrixEx)
	tx := &Transaction{Currency: currency, data: d}
	return tx
}

func newTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, V *big.Int, R *big.Int, S *big.Int, typ byte, isEntrustTx byte, currency string, committime uint64) *Transaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
		TxEnterType:  NormalTxIndex,
		IsEntrustTx:  isEntrustTx,
		CommitTime:   committime,
	}
	mx := new(Matrix_Extra)
	mx.TxType = typ
	d.Extra = append(d.Extra, *mx)
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}
	if V != nil {
		d.V.Set(V)
	}
	if R != nil {
		d.R.Set(R)
	}
	if S != nil {
		d.S.Set(S)
	}
	tx := &Transaction{Currency: currency, data: d}
	return tx
}

// ChainId returns which chain id this transaction was signed for (if at all)
func (tx *Transaction) ChainId() *big.Int {
	return deriveChainId(tx.data.V)
}

func isProtectedV(V *big.Int) bool {
	if V.BitLen() <= 8 {
		v := V.Uint64()
		return v != 27 && v != 28
	}
	// anything not 27 or 28 are considered unprotected
	return true
}

type extTransaction struct {
	Data     txdata
	Currency string
	From     common.Address
}

// EncodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	etx := &extTransaction{Data: tx.data, Currency: tx.Currency}
	if tx.GetMatrixType() == common.ExtraUnGasMinerTxType || tx.GetMatrixType() == common.ExtraUnGasValidatorTxType ||
		tx.GetMatrixType() == common.ExtraUnGasInterestTxType || tx.GetMatrixType() == common.ExtraUnGasTxsType || tx.GetMatrixType() == common.ExtraUnGasLotteryTxType {
		etx.From = tx.From()
	}
	return rlp.Encode(w, etx)
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	var err error
	_, size, _ := s.Kind()
	var extData extTransaction
	err = s.Decode(&extData)
	tx.data = extData.Data
	tx.Currency = extData.Currency
	if tx.GetMatrixType() == common.ExtraUnGasMinerTxType || tx.GetMatrixType() == common.ExtraUnGasValidatorTxType ||
		tx.GetMatrixType() == common.ExtraUnGasInterestTxType || tx.GetMatrixType() == common.ExtraUnGasTxsType || tx.GetMatrixType() == common.ExtraUnGasLotteryTxType {
		tx.SetFromLoad(extData.From)
	}
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}
	return err
}

// MarshalJSON encodes the web3 RPC transaction format.
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	hash := tx.Hash()
	data := tx.data
	data.Hash = &hash
	return data.MarshalJSON()
}

// UnmarshalJSON decodes the web3 RPC transaction format.
func (tx *Transaction) UnmarshalJSON(input []byte) error {
	var dec txdata
	if err := dec.UnmarshalJSON(input); err != nil {
		return err
	}
	var V byte
	if isProtectedV(dec.V) {
		chainID := deriveChainId(dec.V).Uint64()
		V = byte(dec.V.Uint64() - 35 - 2*chainID)
	} else {
		V = byte(dec.V.Uint64() - 27)
	}
	if !crypto.ValidateSignatureValues(V, dec.R, dec.S, false) {
		return ErrInvalidSig
	}
	*tx = Transaction{data: dec}
	return nil
}

type man_txdata struct {
	currency string
	data     txdata1
}

func (tx *Transaction) ManTx_UnmarshalJSON(input []byte) error {
	var dec man_txdata
	if err := dec.UnmarshalJSON(input); err != nil {
		return err
	}
	var V byte
	if isProtectedV(dec.data.V) {
		chainID := deriveChainId(dec.data.V).Uint64()
		V = byte(dec.data.V.Uint64() - 35 - 2*chainID)
	} else {
		V = byte(dec.data.V.Uint64() - 27)
	}
	if !crypto.ValidateSignatureValues(V, dec.data.R, dec.data.S, false) {
		return ErrInvalidSig
	}
	TxdataStringToAddres(&dec.data, &tx.data)
	*tx = Transaction{Currency: dec.currency, data: tx.data}
	return nil
}
func (tx *Transaction) GetConstructorType() uint16 {
	return uint16(NormalTxIndex)
}
func (tx *Transaction) Data() []byte       { return common.CopyBytes(tx.data.Payload) }
func (tx *Transaction) Gas() uint64        { return tx.data.GasLimit }
func (tx *Transaction) GasPrice() *big.Int { return new(big.Int).Set(tx.data.Price) }
func (tx *Transaction) Value() *big.Int    { return new(big.Int).Set(tx.data.Amount) }
func (tx *Transaction) Nonce() uint64      { return tx.data.AccountNonce }
func (tx *Transaction) CheckNonce() bool   { return true }
func (tx *Transaction) GetMakeHashfield(chid *big.Int) []interface{} {
	var data1 txdata1
	TxdataAddresToString(tx.Currency, &tx.data, &data1)
	return []interface{}{
		data1.AccountNonce,
		data1.Price,
		data1.GasLimit,
		data1.Recipient,
		data1.Amount,
		data1.Payload,
		chid, uint(0), uint(0),
		data1.TxEnterType,
		data1.IsEntrustTx,
		data1.CommitTime,
		data1.Extra,
	}
}
func (tx *Transaction) GetTxHashStruct() {

}

func (tx *Transaction) GetCreateTime() uint32 {
	return uint32(tx.data.CommitTime)
}

func (tx *Transaction) Call() error {
	return nil
}

func (tx *Transaction) GetLocalHeight() uint32 {
	if tx.data.Extra != nil && len(tx.data.Extra) > 0 {
		return uint32(tx.data.Extra[0].LockHeight)
	}
	return uint32(time.Now().Unix())
}
func (tx *Transaction) TxType() byte { return tx.data.TxEnterType }

func (tx *Transaction) IsEntrustTx() bool { return tx.data.IsEntrustTx == 1 }

//
func (tx *Transaction) GetMatrix_EX() []Matrix_Extra { return tx.data.Extra }

func (tx *Transaction) GetMatrixType() byte {
	if tx.data.Extra != nil && len(tx.data.Extra) > 0 {
		return tx.data.Extra[0].TxType
	}
	return common.ExtraNormalTxType
}

//Y 为了兼容去掉的Message结构体
func (tx *Transaction) From() common.Address {
	addr, err := tx.GetTxFrom()
	if err != nil {
		return common.Address{}
	}
	return addr
}
func (tx *Transaction) GetFromLoad() interface{} {
	return tx.from.Load()
}

func (tx *Transaction) SetFromLoad(x interface{}) {
	from, ok := x.(common.Address)
	if ok {
		tx.from.Store(sigCache{signer: NewEIP155Signer(tx.ChainId()), from: from})
	} else {
		tx.from.Store(x)
	}
}

func (tx *Transaction) Setentrustfrom(x interface{}) {
	tx.entrustfrom.Store(x)
}
func (tx *Transaction) GasFrom() (from common.Address) {
	tmp, ok := tx.from.Load().(sigCache)
	if !ok {
		tmpfrom, isok := tx.from.Load().(common.Address)
		if !isok {
			return common.Address{}
		}
		from = tmpfrom
	} else {
		from = tmp.from
	}
	return
}
func (tx *Transaction) AmontFrom() (from common.Address) {
	//TODO from 要改为entrustfrom
	tmp, ok := tx.entrustfrom.Load().(sigCache)
	if !ok {
		tmpfrom, isok := tx.entrustfrom.Load().(common.Address)
		if !isok {
			return tx.From()
		}
		from = tmpfrom
	} else {
		from = tmp.from
	}
	return
}

//
func (tx *Transaction) GetTxFrom() (from common.Address, err error) {
	if tx.from.Load() == nil {
		//如果交易没有做过验签则err不为空。
		return common.Address{}, errors.New("Address is Nil")
	}
	var tf common.Address
	//如果交易做过验签则err为空。
	tmp, ok := tx.from.Load().(sigCache)
	if !ok {
		tmpfrom, isok := tx.from.Load().(common.Address)
		if !isok {
			return common.Address{}, errors.New("load Address is Nil")
		}
		if tmpfrom != tf {
			from = tmpfrom
		} else {
			return common.Address{}, errors.New("load Address is Nil")
		}
	} else {
		from = tmp.from
	}
	return
}

func (tx *Transaction) TotalAmount() *big.Int {
	amount := tx.data.Amount
	txEx := tx.GetMatrix_EX()
	total := new(big.Int)
	if len(txEx) > 0 {
		for _, extra := range tx.data.Extra[0].ExtraTo {
			total.Add(total, extra.Amount)
		}
	}
	return total.Add(total, amount)
}

//// Cost returns amount + gasprice * gaslimit.
func (tx *Transaction) CostALL() *big.Int {
	total := new(big.Int).Mul(tx.data.Price, new(big.Int).SetUint64(tx.data.GasLimit))
	total.Add(total, tx.data.Amount)
	txEx := tx.GetMatrix_EX()
	if len(txEx) > 0 {
		for _, extra := range tx.data.Extra[0].ExtraTo {
			total.Add(total, extra.Amount)
		}
	}
	return total
}

func (tx *Transaction) GetTxNLen() int {
	return len(tx.N)
}
func (tx *Transaction) GetIsEntrustGas() bool {
	return tx.IsEntrustGas
}

func (tx *Transaction) GetIsEntrustByTime() bool {
	return tx.IsEntrustByTime
}
func (tx *Transaction) GetIsEntrustByCount() bool {
	return tx.IsEntrustByCount
}
func (tx *Transaction) SetIsEntrustGas(b bool) {
	tx.IsEntrustGas = b
}
func (tx *Transaction) SetIsEntrustByTime(b bool) {
	tx.IsEntrustByTime = b
}
func (tx *Transaction) SetIsEntrustByCount(b bool) {
	tx.IsEntrustByCount = b
}

//
func (tx *Transaction) GetTxV() *big.Int  { return tx.data.V }
func (tx *Transaction) SetTxV(v *big.Int) { tx.data.V = v }
func (tx *Transaction) SetTxR(r *big.Int) { tx.data.R = r }

//
func (tx *Transaction) GetTxS() *big.Int { return tx.data.S }

//
func (tx *Transaction) GetTxR() *big.Int { return tx.data.R }
func (tx *Transaction) GetTxN(index int) uint32 {
	return tx.N[index]
}

// 在传递交易时用来操作Nonce
func (tx *Transaction) SetNonce(nc uint64) {
	tx.data.AccountNonce = nc
}

//
func GetFloodData(tx *Transaction) *Floodtxdata {

	floodtx := &Floodtxdata{
		AccountNonce: tx.data.AccountNonce & params.NonceSubOne,
		Price:        tx.data.Price,
		GasLimit:     tx.data.GasLimit,
		Recipient:    tx.data.Recipient,
		Amount:       tx.data.Amount,
		Payload:      tx.data.Payload,
		Currency:     tx.Currency,
		// Signature values
		V:           tx.data.V,
		R:           tx.data.R,
		TxEnterType: tx.data.TxEnterType,
		IsEntrustTx: tx.data.IsEntrustTx,
		CommitTime:  tx.data.CommitTime,
		Extra:       tx.data.Extra,
	}
	return floodtx
}

//
func SetFloodData(floodtx *Floodtxdata) *Transaction {
	tx := &Transaction{}
	tx.data.AccountNonce = floodtx.AccountNonce | params.NonceAddOne
	tx.data.Price = floodtx.Price
	tx.data.GasLimit = floodtx.GasLimit
	tx.data.Recipient = floodtx.Recipient
	tx.data.Amount = floodtx.Amount
	tx.data.Payload = floodtx.Payload
	// Signature values
	tx.data.V = floodtx.V
	tx.data.R = floodtx.R
	tx.data.TxEnterType = floodtx.TxEnterType
	tx.data.IsEntrustTx = floodtx.IsEntrustTx
	tx.data.CommitTime = floodtx.CommitTime
	tx.data.Extra = floodtx.Extra
	tx.Currency = floodtx.Currency
	return tx
}

func ConvTxtoMxtx(txer SelfTransaction) *Transaction_Mx {
	tx, ok := txer.(*Transaction)
	if !ok {
		return nil
	}
	tx_Mx := &Transaction_Mx{}
	tx_Mx.Data.AccountNonce = tx.data.AccountNonce & params.NonceSubOne
	tx_Mx.Data.Price = tx.data.Price
	tx_Mx.Data.GasLimit = tx.data.GasLimit
	tx_Mx.Data.Recipient = tx.data.Recipient
	tx_Mx.Data.Amount = tx.data.Amount
	tx_Mx.Data.Payload = tx.data.Payload
	// Signature values
	tx_Mx.Data.V = tx.data.V
	tx_Mx.Data.R = tx.data.R
	tx_Mx.Data.S = tx.data.S
	tx_Mx.Data.TxEnterType = tx.data.TxEnterType
	tx_Mx.Data.IsEntrustTx = tx.data.IsEntrustTx
	tx_Mx.Data.CommitTime = tx.data.CommitTime
	tx_Mx.Data.Extra = tx.data.Extra
	tx_Mx.Currency = tx.Currency
	if len(tx.data.Extra) > 0 {
		tx_Mx.TxType_Mx = tx.data.Extra[0].TxType
		tx_Mx.LockHeight = tx.data.Extra[0].LockHeight
		tx_Mx.ExtraTo = tx.data.Extra[0].ExtraTo
	} else {
		log.Error("tx.data.Extra is nil")
		return nil
	}
	return tx_Mx
}

func ConvMxtotx(tx_Mx *Transaction_Mx) *Transaction {
	txd := txdata{
		AccountNonce: tx_Mx.Data.AccountNonce | params.NonceAddOne,
		Price:        tx_Mx.Data.Price,
		GasLimit:     tx_Mx.Data.GasLimit,
		Recipient:    tx_Mx.Data.Recipient,
		Amount:       tx_Mx.Data.Amount,
		Payload:      tx_Mx.Data.Payload,
		// Signature values
		V:           tx_Mx.Data.V,
		R:           tx_Mx.Data.R,
		S:           tx_Mx.Data.S,
		TxEnterType: tx_Mx.Data.TxEnterType,
		IsEntrustTx: tx_Mx.Data.IsEntrustTx,
		CommitTime:  tx_Mx.Data.CommitTime,
		Extra:       tx_Mx.Data.Extra,
	}
	tx := &Transaction{Currency: tx_Mx.Currency, data: txd}
	return tx
}

//
func (tx *Transaction) SetTxS(S *big.Int) { tx.data.S = S }
func (tx *Transaction) SetTxCurrency(currency string) {
	tx.Currency = currency
}
func (tx *Transaction) GetTxCurrency() string {
	str := tx.Currency
	if str == "" {
		str = params.MAN_COIN
	}
	return str
}

// To returns the recipient address of the transaction.
// It returns nil if the transaction is a contract creation.
func (tx *Transaction) To() *common.Address {
	//return tx.data.Recipient
	if tx.data.Recipient == nil {
		return nil
	}
	to := *tx.data.Recipient
	return &to
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	if len(tx.data.Extra) != 1 {
		panic("Transaction data Extra length must be 1")
	}
	v := rlpHash(tx)
	tx.hash.Store(v)
	return v
}

// Size returns the true RLP encoded storage size of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &tx.data)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

// WithSignature returns a new transaction with the given signature.
// This signature needs to be formatted as described in the yellow paper (v+27).
func (tx *Transaction) WithSignature(signer Signer, sig []byte) (SelfTransaction, error) {
	r, s, v, err := signer.SignatureValues(tx, sig)
	if err != nil {
		return nil, err
	}
	cpy := &Transaction{Currency: tx.Currency, data: tx.data}
	cpy.data.R, cpy.data.S, cpy.data.V = r, s, v
	return cpy, nil
}

// Cost returns amount + gasprice * gaslimit.
func (tx *Transaction) Cost() *big.Int {
	total := new(big.Int).Mul(tx.data.Price, new(big.Int).SetUint64(tx.data.GasLimit))
	total.Add(total, tx.data.Amount)
	return total
}

func (tx *Transaction) RawSignatureValues() (*big.Int, *big.Int, *big.Int) {
	return tx.data.V, tx.data.R, tx.data.S
}

// Transactions is a Transaction slice type for basic sorting.
type SelfTransactions []SelfTransaction

// Len returns the length of s.
func (s SelfTransactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s SelfTransactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s SelfTransactions) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}

// TxDifference returns a new set t which is the difference between a to b.
func TxDifference(a, b SelfTransactions) (keep SelfTransactions) {
	keep = make(SelfTransactions, 0, len(a))

	remove := make(map[common.Hash]struct{})
	for _, tx := range b {
		remove[tx.Hash()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[tx.Hash()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}

// TxByNonce implements the sort interface to allow sorting a list of transactions
// by their nonces. This is usually only useful for sorting transactions from a
// single account, otherwise a nonce comparison doesn't make much sense.
type TxByNonce []SelfTransaction

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].Nonce() < s[j].Nonce() }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s *TxByNonce) Push(x interface{}) {
}

func (s *TxByNonce) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

// TxByPrice implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPrice SelfTransactions

func (s TxByPrice) Len() int           { return len(s) }
func (s TxByPrice) Less(i, j int) bool { return s[i].GasPrice().Cmp(s[j].GasPrice()) > 0 }
func (s TxByPrice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *TxByPrice) Push(x interface{}) {
}

func (s *TxByPrice) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

// TransactionsByPriceAndNonce represents a set of transactions that can return
// transactions in a profit-maximizing sorted order, while supporting removing
// entire batches of transactions for non-executable accounts.
type TransactionsByPriceAndNonce struct {
	txs    map[common.Address]SelfTransactions // Per account nonce-sorted list of transactions
	heads  TxByNonce
	signer Signer // Signer for the set of transactions
}

// NewTransactionsByPriceAndNonce creates a transaction set that can retrieve
// price sorted transactions in a nonce-honouring way.
//
// Note, the input map is reowned so the caller should not interact any more with
// if after providing it to the constructor.
func NewTransactionsByPriceAndNonce(signer Signer, txs map[common.Address]SelfTransactions) *TransactionsByPriceAndNonce {
	// Initialize a price based heap with the head transactions
	heads := make(TxByNonce, 0, len(txs))
	for from, accTxs := range txs {
		heads = append(heads, accTxs[0])
		// Ensure the sender address is from the signer
		acc, _ := Sender(signer, accTxs[0])
		txs[acc] = accTxs[1:]
		if from != acc {
			delete(txs, from)
		}
	}
	// Assemble and return the transaction set
	return &TransactionsByPriceAndNonce{
		txs:    txs,
		heads:  heads,
		signer: signer,
	}
}

// Peek returns the next transaction by price.
func (t *TransactionsByPriceAndNonce) Peek() SelfTransaction {
	if len(t.heads) == 0 {
		return nil
	}
	return t.heads[0]
}

// Shift replaces the current best head with the next one from the same account.
func (t *TransactionsByPriceAndNonce) Shift() {
	acc, _ := Sender(t.signer, t.heads[0])
	if txs, ok := t.txs[acc]; ok && len(txs) > 0 {
		t.heads[0], t.txs[acc] = txs[0], txs[1:]
		heap.Fix(&t.heads, 0)
	} else {
		heap.Pop(&t.heads)
	}
}

// Pop removes the best transaction, *not* replacing it with the next one from
// the same account. This should be used when a transaction cannot be executed
// and hence all subsequent ones should be discarded from the same account.
func (t *TransactionsByPriceAndNonce) Pop() {
	heap.Pop(&t.heads)
}
