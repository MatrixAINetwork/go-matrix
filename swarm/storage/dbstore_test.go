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

package storage

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/matrix/go-matrix/common"
)

func initDbStore(t *testing.T) *DbStore {
	dir, err := ioutil.TempDir("", "bzz-storage-test")
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewDbStore(dir, MakeHashFunc(SHA3Hash), defaultDbCapacity, defaultRadius)
	if err != nil {
		t.Fatal("can't create store:", err)
	}
	return m
}

func testDbStore(l int64, branches int64, t *testing.T) {
	m := initDbStore(t)
	defer m.Close()
	testStore(m, l, branches, t)
}

func TestDbStore128_0x1000000(t *testing.T) {
	testDbStore(0x1000000, 128, t)
}

func TestDbStore128_10000_(t *testing.T) {
	testDbStore(10000, 128, t)
}

func TestDbStore128_1000_(t *testing.T) {
	testDbStore(1000, 128, t)
}

func TestDbStore128_100_(t *testing.T) {
	testDbStore(100, 128, t)
}

func TestDbStore2_100_(t *testing.T) {
	testDbStore(100, 2, t)
}

func TestDbStoreNotFound(t *testing.T) {
	m := initDbStore(t)
	defer m.Close()
	_, err := m.Get(ZeroKey)
	if err != notFound {
		t.Errorf("Expected notFound, got %v", err)
	}
}

func TestDbStoreSyncIterator(t *testing.T) {
	m := initDbStore(t)
	defer m.Close()
	keys := []Key{
		Key(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000")),
		Key(common.Hex2Bytes("4000000000000000000000000000000000000000000000000000000000000000")),
		Key(common.Hex2Bytes("5000000000000000000000000000000000000000000000000000000000000000")),
		Key(common.Hex2Bytes("3000000000000000000000000000000000000000000000000000000000000000")),
		Key(common.Hex2Bytes("2000000000000000000000000000000000000000000000000000000000000000")),
		Key(common.Hex2Bytes("1000000000000000000000000000000000000000000000000000000000000000")),
	}
	for _, key := range keys {
		m.Put(NewChunk(key, nil))
	}
	it, err := m.NewSyncIterator(DbSyncState{
		Start: Key(common.Hex2Bytes("1000000000000000000000000000000000000000000000000000000000000000")),
		Stop:  Key(common.Hex2Bytes("4000000000000000000000000000000000000000000000000000000000000000")),
		First: 2,
		Last:  4,
	})
	if err != nil {
		t.Fatalf("unexpected error creating NewSyncIterator")
	}

	var chunk Key
	var res []Key
	for {
		chunk = it.Next()
		if chunk == nil {
			break
		}
		res = append(res, chunk)
	}
	if len(res) != 1 {
		t.Fatalf("Expected 1 chunk, got %v: %v", len(res), res)
	}
	if !bytes.Equal(res[0][:], keys[3]) {
		t.Fatalf("Expected %v chunk, got %v", keys[3], res[0])
	}

	if err != nil {
		t.Fatalf("unexpected error creating NewSyncIterator")
	}

	it, err = m.NewSyncIterator(DbSyncState{
		Start: Key(common.Hex2Bytes("1000000000000000000000000000000000000000000000000000000000000000")),
		Stop:  Key(common.Hex2Bytes("5000000000000000000000000000000000000000000000000000000000000000")),
		First: 2,
		Last:  4,
	})

	res = nil
	for {
		chunk = it.Next()
		if chunk == nil {
			break
		}
		res = append(res, chunk)
	}
	if len(res) != 2 {
		t.Fatalf("Expected 2 chunk, got %v: %v", len(res), res)
	}
	if !bytes.Equal(res[0][:], keys[3]) {
		t.Fatalf("Expected %v chunk, got %v", keys[3], res[0])
	}
	if !bytes.Equal(res[1][:], keys[2]) {
		t.Fatalf("Expected %v chunk, got %v", keys[2], res[1])
	}

	if err != nil {
		t.Fatalf("unexpected error creating NewSyncIterator")
	}

	it, _ = m.NewSyncIterator(DbSyncState{
		Start: Key(common.Hex2Bytes("1000000000000000000000000000000000000000000000000000000000000000")),
		Stop:  Key(common.Hex2Bytes("4000000000000000000000000000000000000000000000000000000000000000")),
		First: 2,
		Last:  5,
	})
	res = nil
	for {
		chunk = it.Next()
		if chunk == nil {
			break
		}
		res = append(res, chunk)
	}
	if len(res) != 2 {
		t.Fatalf("Expected 2 chunk, got %v", len(res))
	}
	if !bytes.Equal(res[0][:], keys[4]) {
		t.Fatalf("Expected %v chunk, got %v", keys[4], res[0])
	}
	if !bytes.Equal(res[1][:], keys[3]) {
		t.Fatalf("Expected %v chunk, got %v", keys[3], res[1])
	}

	it, _ = m.NewSyncIterator(DbSyncState{
		Start: Key(common.Hex2Bytes("2000000000000000000000000000000000000000000000000000000000000000")),
		Stop:  Key(common.Hex2Bytes("4000000000000000000000000000000000000000000000000000000000000000")),
		First: 2,
		Last:  5,
	})
	res = nil
	for {
		chunk = it.Next()
		if chunk == nil {
			break
		}
		res = append(res, chunk)
	}
	if len(res) != 1 {
		t.Fatalf("Expected 1 chunk, got %v", len(res))
	}
	if !bytes.Equal(res[0][:], keys[3]) {
		t.Fatalf("Expected %v chunk, got %v", keys[3], res[0])
	}
}
