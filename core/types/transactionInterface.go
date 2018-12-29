package types

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"math/big"
)

const (
	NormalTxIndex    byte = iota // NormalPool save normal transaction
	BroadCastTxIndex             // BroadcastPool save broadcast transaction

)

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
			log.Info("file transactionInterface", "func SetMxToTransaction1", "tx is nil", "Transaction_Mx", txm)
		}
	} else if txm.TxType_Mx == common.ExtraBroadTxType {
		tx := SetTransactionMx(txm)
		if tx != nil {
			txer = tx
		} else {
			log.Info("file transactionInterface", "func SetMxToTransaction2", "tx is nil", "Transaction_Mx", txm)
		}
	} else {
		log.Info("file transactionInterface", "func SetMxToTransaction", "Transaction_Mx is nil", txm)
	}
	return
}
