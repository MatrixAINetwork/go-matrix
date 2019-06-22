// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package vm

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/params"
)

type MatrixDepositVersion struct {
	Deposit001 MatrixDeposit001
	Deposit002 MatrixDeposit002
}

func (mdv *MatrixDepositVersion) RequiredGas(input []byte) uint64 {
	if len(input) < 4 {
		return 0
	}
	var methodIdArr [4]byte
	copy(methodIdArr[:], input[:4])
	if methodIdArr == interestAddArr {
		return 0
	}
	return params.SstoreSetGas * 2
}

func (mdv *MatrixDepositVersion) Run(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	if in == nil || len(in) == 0 {
		return nil, nil
	}
	if len(in) < 4 {
		return nil, errParameters
	}
	var methodIdArr [4]byte
	copy(methodIdArr[:], in[:4])
	ret := evm.StateDB.GetState(params.MAN_COIN, common.Address{}, common.BytesToHash([]byte(params.DepositVersionKey_1)))
	if ret.Equal(common.BytesToHash([]byte(params.DepositVersion_1))) {
		return mdv.Deposit002.Run(in, contract, evm)
	} else {
		return mdv.Deposit001.Run(in, contract, evm)
	}
}
