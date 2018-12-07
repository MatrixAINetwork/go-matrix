package types

import (
	"math/big"
	"github.com/matrix/go-matrix/common"
)


const (
	NormalTxIndex    byte = iota // NormalPool save normal transaction
	BroadCastTxIndex                   // BroadcastPool save broadcast transaction

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
	GetTxFrom() (common.Address,error)
	SetNonce(nc uint64)
	GetTxS() *big.Int
	GetTxR() *big.Int
	GetTxV() *big.Int
	SetTxS(S *big.Int)
	SetTxV(v *big.Int)
	SetTxR(r *big.Int)
	To() *common.Address
	Hash() common.Hash
	GetTxHashStruct()   //获取交易结构中需要哈希的成员  返回值应该是什么？？？？？
	Call() error     //执行交易
	Size() common.StorageSize
	GetFromLoad() interface{}
	SetFromLoad(x interface{})
	ChainId() *big.Int
	WithSignature(signer Signer, sig []byte) (SelfTransaction, error)
	GetTxNLen()int
	GetTxN(index int) uint32
	RawSignatureValues() (*big.Int, *big.Int, *big.Int)
	//Protected() bool
	GetConstructorType()uint16
	GasFrom() common.Address
	AmontFrom() common.Address
	GetMatrixType() byte
	Setentrustfrom(x interface{})
	CoinType()string
	SetCoinType(typ string)
}

func SetTransactionToMx(txer SelfTransaction)(txm *Transaction_Mx){
	if txer.TxType() == BroadCastTxIndex{
		txm = GetTransactionMx(txer)
	}else if txer.TxType() == NormalTxIndex{
		txm = ConvTxtoMxtx(txer)
	}
	return
}

func SetMxToTransaction(txm *Transaction_Mx)(txer SelfTransaction){
	if txm.TxType_Mx == common.ExtraNormalTxType{
		txer = ConvMxtotx(txm)
	} else if txm.TxType_Mx == common.ExtraBroadTxType {
		txer = SetTransactionMx(txm)
	}
	return
}