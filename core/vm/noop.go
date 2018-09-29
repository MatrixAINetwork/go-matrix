// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package vm

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/params"
)

func NoopCanTransfer(db StateDB, from common.Address, balance *big.Int) bool {
	return true
}
func NoopTransfer(db StateDB, from, to common.Address, amount *big.Int) {}

type NoopEVMCallContext struct{}

func (NoopEVMCallContext) Call(caller ContractRef, addr common.Address, data []byte, gas, value *big.Int) ([]byte, error) {
	return nil, nil
}
func (NoopEVMCallContext) CallCode(caller ContractRef, addr common.Address, data []byte, gas, value *big.Int) ([]byte, error) {
	return nil, nil
}
func (NoopEVMCallContext) Create(caller ContractRef, data []byte, gas, value *big.Int) ([]byte, common.Address, error) {
	return nil, common.Address{}, nil
}
func (NoopEVMCallContext) DelegateCall(me ContractRef, addr common.Address, data []byte, gas *big.Int) ([]byte, error) {
	return nil, nil
}

type NoopStateDB struct{}

func (NoopStateDB) CreateAccount(common.Address)                                       {}
func (NoopStateDB) SubBalance(common.Address, *big.Int)                                {}
func (NoopStateDB) AddBalance(common.Address, *big.Int)                                {}
func (NoopStateDB) GetBalance(common.Address) *big.Int                                 { return nil }
func (NoopStateDB) GetNonce(common.Address) uint64                                     { return 0 | params.NonceAddOne } //YY
func (NoopStateDB) SetNonce(common.Address, uint64)                                    {}
func (NoopStateDB) GetCodeHash(common.Address) common.Hash                             { return common.Hash{} }
func (NoopStateDB) GetCode(common.Address) []byte                                      { return nil }
func (NoopStateDB) SetCode(common.Address, []byte)                                     {}
func (NoopStateDB) GetCodeSize(common.Address) int                                     { return 0 }
func (NoopStateDB) AddRefund(uint64)                                                   {}
func (NoopStateDB) GetRefund() uint64                                                  { return 0 }
func (NoopStateDB) GetState(common.Address, common.Hash) common.Hash                   { return common.Hash{} }
func (NoopStateDB) SetState(common.Address, common.Hash, common.Hash)                  {}
func (NoopStateDB) Suicide(common.Address) bool                                        { return false }
func (NoopStateDB) HasSuicided(common.Address) bool                                    { return false }
func (NoopStateDB) Exist(common.Address) bool                                          { return false }
func (NoopStateDB) Empty(common.Address) bool                                          { return false }
func (NoopStateDB) RevertToSnapshot(int)                                               {}
func (NoopStateDB) Snapshot() int                                                      { return 0 }
func (NoopStateDB) AddLog(*types.Log)                                                  {}
func (NoopStateDB) AddPreimage(common.Hash, []byte)                                    {}
func (NoopStateDB) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) {}
