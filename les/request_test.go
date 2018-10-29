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

package les

import (
	"context"
	"testing"
	"time"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/rawdb"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/man"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/light"
)

var testBankSecureTrieKey = secAddr(testBankAddress)

func secAddr(addr common.Address) []byte {
	return crypto.Keccak256(addr[:])
}

type accessTestFn func(db mandb.Database, bhash common.Hash, number uint64) light.OdrRequest

func TestBlockAccessLes1(t *testing.T) { testAccess(t, 1, tfBlockAccess) }

func TestBlockAccessLes2(t *testing.T) { testAccess(t, 2, tfBlockAccess) }

func tfBlockAccess(db mandb.Database, bhash common.Hash, number uint64) light.OdrRequest {
	return &light.BlockRequest{Hash: bhash, Number: number}
}

func TestReceiptsAccessLes1(t *testing.T) { testAccess(t, 1, tfReceiptsAccess) }

func TestReceiptsAccessLes2(t *testing.T) { testAccess(t, 2, tfReceiptsAccess) }

func tfReceiptsAccess(db mandb.Database, bhash common.Hash, number uint64) light.OdrRequest {
	return &light.ReceiptsRequest{Hash: bhash, Number: number}
}

func TestTrieEntryAccessLes1(t *testing.T) { testAccess(t, 1, tfTrieEntryAccess) }

func TestTrieEntryAccessLes2(t *testing.T) { testAccess(t, 2, tfTrieEntryAccess) }

func tfTrieEntryAccess(db mandb.Database, bhash common.Hash, number uint64) light.OdrRequest {
	if number := rawdb.ReadHeaderNumber(db, bhash); number != nil {
		return &light.TrieRequest{Id: light.StateTrieID(rawdb.ReadHeader(db, bhash, *number)), Key: testBankSecureTrieKey}
	}
	return nil
}

func TestCodeAccessLes1(t *testing.T) { testAccess(t, 1, tfCodeAccess) }

func TestCodeAccessLes2(t *testing.T) { testAccess(t, 2, tfCodeAccess) }

func tfCodeAccess(db mandb.Database, bhash common.Hash, num uint64) light.OdrRequest {
	number := rawdb.ReadHeaderNumber(db, bhash)
	if number != nil {
		return nil
	}
	header := rawdb.ReadHeader(db, bhash, *number)
	if header.Number.Uint64() < testContractDeployed {
		return nil
	}
	sti := light.StateTrieID(header)
	ci := light.StorageTrieID(sti, crypto.Keccak256Hash(testContractAddr[:]), common.Hash{})
	return &light.CodeRequest{Id: ci, Hash: crypto.Keccak256Hash(testContractCodeDeployed)}
}

func testAccess(t *testing.T, protocol int, fn accessTestFn) {
	// Assemble the test environment
	peers := newPeerSet()
	dist := newRequestDistributor(peers, make(chan struct{}))
	rm := newRetrieveManager(peers, dist, nil)
	db := mandb.NewMemDatabase()
	ldb := mandb.NewMemDatabase()
	odr := NewLesOdr(ldb, light.NewChtIndexer(db, true), light.NewBloomTrieIndexer(db, true), man.NewBloomIndexer(db, light.BloomTrieFrequency), rm)

	pm := newTestProtocolManagerMust(t, false, 4, testChainGen, nil, nil, db)
	lpm := newTestProtocolManagerMust(t, true, 0, nil, peers, odr, ldb)
	_, err1, lpeer, err2 := newTestPeerPair("peer", protocol, pm, lpm)
	select {
	case <-time.After(time.Millisecond * 100):
	case err := <-err1:
		t.Fatalf("peer 1 handshake error: %v", err)
	case err := <-err2:
		t.Fatalf("peer 1 handshake error: %v", err)
	}

	lpm.synchronise(lpeer)

	test := func(expFail uint64) {
		for i := uint64(0); i <= pm.blockchain.CurrentHeader().Number.Uint64(); i++ {
			bhash := rawdb.ReadCanonicalHash(db, i)
			if req := fn(ldb, bhash, i); req != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancel()

				err := odr.Retrieve(ctx, req)
				got := err == nil
				exp := i < expFail
				if exp && !got {
					t.Errorf("object retrieval failed")
				}
				if !exp && got {
					t.Errorf("unexpected object retrieval success")
				}
			}
		}
	}

	// temporarily remove peer to test odr fails
	peers.Unregister(lpeer.id)
	time.Sleep(time.Millisecond * 10) // ensure that all peerSetNotify callbacks are executed
	// expect retrievals to fail (except genesis block) without a les peer
	test(0)

	peers.Register(lpeer)
	time.Sleep(time.Millisecond * 10) // ensure that all peerSetNotify callbacks are executed
	lpeer.lock.Lock()
	lpeer.hasBlock = func(common.Hash, uint64) bool { return true }
	lpeer.lock.Unlock()
	// expect all retrievals to pass
	test(5)
}
