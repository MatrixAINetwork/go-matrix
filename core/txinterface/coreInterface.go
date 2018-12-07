package txinterface

import (
	"github.com/matrix/go-matrix/common"
	"math/big"
	"github.com/matrix/go-matrix/core/types"
)
// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	//FromFrontier() (common.Address, error)
	To() *common.Address
	GasFrom() common.Address
	AmontFrom() common.Address
	GasPrice() *big.Int
	Gas() uint64
	Value() *big.Int
	Hash() common.Hash
	Nonce() uint64
	CheckNonce() bool
	Data() []byte
	GetMatrixType() byte
	GetMatrix_EX() []types.Matrix_Extra //YYY  注释 Extra() 方法 改用此方法
	TxType() byte
}

type StateTransitioner interface {
	//InitStateTransition(evm *vm.EVM, msg Message, gp uint64)
	TransitionDb() (ret []byte, usedGas uint64, failed bool, err error)
	To() common.Address
	UseGas(amount uint64) error
	BuyGas() error
	PreCheck() error
	RefundGas()
	GasUsed() uint64
	//CreateTransition(evm *vm.EVM, msg Message, gp uint64)StateTransitioner
}