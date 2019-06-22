// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package txinterface

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"math/big"
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
	GetMatrix_EX() []types.Matrix_Extra //Y  注释 Extra() 方法 改用此方法
	TxType() byte
	IsEntrustTx() bool
	GetCreateTime() uint32
	GetTxCurrency() string
}

type StateTransitioner interface {
	//InitStateTransition(evm *vm.EVM, msg Message, gp uint64)
	TransitionDb() (ret []byte, usedGas uint64, failed bool, shardings []uint, err error)
	To() common.Address
	UseGas(amount uint64) error
	BuyGas() error
	PreCheck() error
	RefundGas(coinrange string)
	GasUsed() uint64
	//CreateTransition(evm *vm.EVM, msg Message, gp uint64)StateTransitioner
}
