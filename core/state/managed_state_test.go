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

package state

import (
	"testing"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mandb"
)

var addr = common.BytesToAddress([]byte("test"))

func create() (*ManagedState, *account) {
	statedb, _ := New(common.Hash{}, NewDatabase(mandb.NewMemDatabase()))
	ms := ManageState(statedb)
	ms.StateDB.SetNonce(addr, 100)
	ms.accounts[addr] = newAccount(ms.StateDB.getStateObject(addr))
	return ms, ms.accounts[addr]
}

func TestNewNonce(t *testing.T) {
	ms, _ := create()

	nonce := ms.NewNonce(addr)
	if nonce != 100 {
		t.Error("expected nonce 100. got", nonce)
	}

	nonce = ms.NewNonce(addr)
	if nonce != 101 {
		t.Error("expected nonce 101. got", nonce)
	}
}

func TestRemove(t *testing.T) {
	ms, account := create()

	nn := make([]bool, 10)
	for i := range nn {
		nn[i] = true
	}
	account.nonces = append(account.nonces, nn...)

	i := uint64(5)
	ms.RemoveNonce(addr, account.nstart+i)
	if len(account.nonces) != 5 {
		t.Error("expected", i, "'th index to be false")
	}
}

func TestReuse(t *testing.T) {
	ms, account := create()

	nn := make([]bool, 10)
	for i := range nn {
		nn[i] = true
	}
	account.nonces = append(account.nonces, nn...)

	i := uint64(5)
	ms.RemoveNonce(addr, account.nstart+i)
	nonce := ms.NewNonce(addr)
	if nonce != 105 {
		t.Error("expected nonce to be 105. got", nonce)
	}
}

func TestRemoteNonceChange(t *testing.T) {
	ms, account := create()
	nn := make([]bool, 10)
	for i := range nn {
		nn[i] = true
	}
	account.nonces = append(account.nonces, nn...)
	ms.NewNonce(addr)

	ms.StateDB.stateObjects[addr].data.Nonce = 200
	nonce := ms.NewNonce(addr)
	if nonce != 200 {
		t.Error("expected nonce after remote update to be", 200, "got", nonce)
	}
	ms.NewNonce(addr)
	ms.NewNonce(addr)
	ms.NewNonce(addr)
	ms.StateDB.stateObjects[addr].data.Nonce = 200
	nonce = ms.NewNonce(addr)
	if nonce != 204 {
		t.Error("expected nonce after remote update to be", 204, "got", nonce)
	}
}

func TestSetNonce(t *testing.T) {
	ms, _ := create()

	var addr common.Address
	ms.SetNonce(addr, 10)

	if ms.GetNonce(addr) != 10 {
		t.Error("Expected nonce of 10, got", ms.GetNonce(addr))
	}

	addr[0] = 1
	ms.StateDB.SetNonce(addr, 1)

	if ms.GetNonce(addr) != 1 {
		t.Error("Expected nonce of 1, got", ms.GetNonce(addr))
	}
}
