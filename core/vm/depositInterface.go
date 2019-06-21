// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package vm

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"math/big"
)

type DepositInterface interface {
	GetDepositBase(contract *Contract, stateDB StateDBManager, addr common.Address) *common.DepositBase
}
type TransInterestsInterface interface {
	TransferInterests(amount *big.Int,position uint64,time uint64,address common.Address,state StateDBManager)error
}