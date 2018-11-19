package types

import (
	"math/big"
	"github.com/matrix/go-matrix/common"
)


const (
	NormalTxIndex    common.TxTypeInt = iota // NormalPool save normal transaction
	BroadCastTxIndex                   // BroadcastPool save broadcast transaction
)

type SelfTransaction interface {
	TxType() common.TxTypeInt
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
	Protected() bool
	GetConstructorType()uint16
}
