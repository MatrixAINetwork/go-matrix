// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package vm

import (
	"math/big"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/params"
)

type dummyContractRef struct {
	calledForEach bool
}

func (dummyContractRef) ReturnGas(*big.Int)          {}
func (dummyContractRef) Address() common.Address     { return common.Address{} }
func (dummyContractRef) Value() *big.Int             { return new(big.Int) }
func (dummyContractRef) SetCode(common.Hash, []byte) {}
func (d *dummyContractRef) ForEachStorage(callback func(key, value common.Hash) bool) {
	d.calledForEach = true
}
func (d *dummyContractRef) SubBalance(amount *big.Int) {}
func (d *dummyContractRef) AddBalance(amount *big.Int) {}
func (d *dummyContractRef) SetBalance(*big.Int)        {}
func (d *dummyContractRef) SetNonce(uint64)            {}
func (d *dummyContractRef) Balance() *big.Int          { return new(big.Int) }

type dummyStateDB struct {
	NoopStateDB
	ref *dummyContractRef
}

func TestStoreCapture(t *testing.T) {
	var (
		env      = NewEVM(Context{}, nil, params.TestChainConfig, Config{})
		logger   = NewStructLogger(nil)
		mem      = NewMemory()
		stack    = newstack()
		contract = NewContract(&dummyContractRef{}, &dummyContractRef{}, new(big.Int), 0)
	)
	stack.push(big.NewInt(1))
	stack.push(big.NewInt(0))

	var index common.Hash

	logger.CaptureState(env, 0, SSTORE, 0, 0, mem, stack, contract, 0, nil)
	if len(logger.changedValues[contract.Address()]) == 0 {
		t.Fatalf("expected exactly 1 changed value on address %x, got %d", contract.Address(), len(logger.changedValues[contract.Address()]))
	}
	exp := common.BigToHash(big.NewInt(1))
	if logger.changedValues[contract.Address()][index] != exp {
		t.Errorf("expected %x, got %x", exp, logger.changedValues[contract.Address()][index])
	}
}
