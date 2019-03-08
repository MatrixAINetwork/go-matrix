package types

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"math/big"
	"sort"
	"github.com/MatrixAINetwork/go-matrix/params"
)

const (
	NormalTxIndex    byte = iota // NormalPool save normal transaction
	BroadCastTxIndex             // BroadcastPool save broadcast transaction

)

type CoinSelfTransaction struct {
	CoinType string
	Txser    SelfTransactions
}

type SelfTransaction interface {
	TxType() byte
	Data() []byte
	Gas() uint64
	GasPrice() *big.Int
	Value() *big.Int
	Nonce() uint64
	CheckNonce() bool
	GetMatrix_EX() []Matrix_Extra
	From() common.Address
	GetTxFrom() (common.Address, error)
	SetNonce(nc uint64)
	GetTxS() *big.Int
	GetTxR() *big.Int
	GetTxV() *big.Int
	SetTxS(S *big.Int)
	SetTxV(v *big.Int)
	SetTxR(r *big.Int)
	To() *common.Address
	Hash() common.Hash
	Size() common.StorageSize
	GetFromLoad() interface{}
	SetFromLoad(x interface{})
	ChainId() *big.Int
	WithSignature(signer Signer, sig []byte) (SelfTransaction, error)
	GetTxNLen() int
	GetTxN(index int) uint32
	RawSignatureValues() (*big.Int, *big.Int, *big.Int)
	//Protected() bool
	GetConstructorType() uint16
	GasFrom() common.Address
	AmontFrom() common.Address
	GetMatrixType() byte
	Setentrustfrom(x interface{})
	IsEntrustTx() bool
	SetTxCurrency(currency string)
	GetTxCurrency() string
	GetCreateTime() uint32
	GetLocalHeight() uint32
	GetIsEntrustGas() bool
	GetIsEntrustByTime() bool
	GetMakeHashfield(chid *big.Int) []interface{}
	SetIsEntrustGas(b bool)
	SetIsEntrustByTime(b bool)
}

func SetTransactionToMx(txer SelfTransaction) (txm *Transaction_Mx) {
	if txer.TxType() == BroadCastTxIndex {
		txm = GetTransactionMx(txer)
	} else if txer.TxType() == NormalTxIndex {
		txm = ConvTxtoMxtx(txer)
	}
	return
}

func SetMxToTransaction(txm *Transaction_Mx) (txer SelfTransaction) {
	txer = nil
	if txm.TxType_Mx == common.ExtraNormalTxType {
		tx := ConvMxtotx(txm)
		if tx != nil {
			txer = tx
		} else {
			log.Info("transactionInterface", "SetMxToTransaction1", "tx is nil", "Transaction_Mx", txm)
		}
	} else if txm.TxType_Mx == common.ExtraBroadTxType {
		tx := SetTransactionMx(txm)
		if tx != nil {
			txer = tx
		} else {
			log.Info("transactionInterface", "SetMxToTransaction2", "tx is nil", "Transaction_Mx", txm)
		}
	} else {
		log.Info("transactionInterface", "SetMxToTransaction", "Transaction_Mx is nil", txm)
	}
	return
}

func GetTX(ctx []CoinSelfTransaction)[] SelfTransaction  {
	var txs []SelfTransaction
	for _,tx:= range ctx{
		for _,t:= range tx.Txser  {
			txs=append(txs,t)
		}
	}
	return txs
}

func GetCoinTX(txs []SelfTransaction)[]CoinSelfTransaction  {
	mm := make(map[string][]SelfTransaction) //BB
	for _, tx := range txs {
		cointype := tx.GetTxCurrency()
		mm[cointype] = append(mm[cointype], tx)
	}
	cs := []CoinSelfTransaction{}
	sorted_keys := make([]string, 0)
	for k, _ := range mm {
		sorted_keys = append(sorted_keys, k)
	}
	sort.Strings(sorted_keys)
	cs = append(cs, CoinSelfTransaction{params.MAN_COIN, mm[params.MAN_COIN]})
	for _, k := range sorted_keys {
		if k == params.MAN_COIN{
			continue
		}
		cs = append(cs, CoinSelfTransaction{k, mm[k]})
	}
	return cs
}

func GetCoinTXRS(txs []SelfTransaction,rxs []*Receipt) ([]CoinSelfTransaction,[]CoinReceipts) {
	var tx []CoinSelfTransaction	//BB
	var rx []CoinReceipts
	tm := make(map[string][]SelfTransaction)
	rm := make(map[string][]*Receipt)
	for i,t := range txs  {
		tm[t.GetTxCurrency()]=append(tm[t.GetTxCurrency()],t)
		rm[t.GetTxCurrency()]=append(rm[t.GetTxCurrency()],rxs[i])
	}
	sorted_keys := make([]string, 0)
	for k, _ := range tm {
		sorted_keys = append(sorted_keys, k)
	}
	sort.Strings(sorted_keys)
	tx=append(tx,CoinSelfTransaction{params.MAN_COIN,tm[params.MAN_COIN]})
	rx=append(rx,CoinReceipts{params.MAN_COIN,rm[params.MAN_COIN]})
	for _,k:=range sorted_keys  {
		if k == params.MAN_COIN{
			continue
		}
		tx=append(tx,CoinSelfTransaction{k,tm[k]})
		rx=append(rx,CoinReceipts{k,rm[k]})
	}
	return tx,rx
}

func TxHashList(txs SelfTransactions)(list []common.Hash){
	for _,tx := range txs{
		list = append(list,tx.Hash())
	}
	return
}