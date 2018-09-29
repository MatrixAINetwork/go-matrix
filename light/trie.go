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

package light

import (
	"context"
	"errors"
	"fmt"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/trie"
)

func NewState(ctx context.Context, head *types.Header, odr OdrBackend) *state.StateDB {
	state, _ := state.New(head.Root, NewStateDatabase(ctx, head, odr))
	return state
}

func NewStateDatabase(ctx context.Context, head *types.Header, odr OdrBackend) state.Database {
	return &odrDatabase{ctx, StateTrieID(head), odr}
}

type odrDatabase struct {
	ctx     context.Context
	id      *TrieID
	backend OdrBackend
}

func (db *odrDatabase) OpenTrie(root common.Hash) (state.Trie, error) {
	return &odrTrie{db: db, id: db.id}, nil
}

func (db *odrDatabase) OpenStorageTrie(addrHash, root common.Hash) (state.Trie, error) {
	return &odrTrie{db: db, id: StorageTrieID(db.id, addrHash, root)}, nil
}

func (db *odrDatabase) CopyTrie(t state.Trie) state.Trie {
	switch t := t.(type) {
	case *odrTrie:
		cpy := &odrTrie{db: t.db, id: t.id}
		if t.trie != nil {
			cpytrie := *t.trie
			cpy.trie = &cpytrie
		}
		return cpy
	default:
		panic(fmt.Errorf("unknown trie type %T", t))
	}
}

func (db *odrDatabase) ContractCode(addrHash, codeHash common.Hash) ([]byte, error) {
	if codeHash == sha3_nil {
		return nil, nil
	}
	if code, err := db.backend.Database().Get(codeHash[:]); err == nil {
		return code, nil
	}
	id := *db.id
	id.AccKey = addrHash[:]
	req := &CodeRequest{Id: &id, Hash: codeHash}
	err := db.backend.Retrieve(db.ctx, req)
	return req.Data, err
}

func (db *odrDatabase) ContractCodeSize(addrHash, codeHash common.Hash) (int, error) {
	code, err := db.ContractCode(addrHash, codeHash)
	return len(code), err
}

func (db *odrDatabase) TrieDB() *trie.Database {
	return nil
}

type odrTrie struct {
	db   *odrDatabase
	id   *TrieID
	trie *trie.Trie
}

func (t *odrTrie) TryGet(key []byte) ([]byte, error) {
	key = crypto.Keccak256(key)
	var res []byte
	err := t.do(key, func() (err error) {
		res, err = t.trie.TryGet(key)
		return err
	})
	return res, err
}

func (t *odrTrie) TryUpdate(key, value []byte) error {
	key = crypto.Keccak256(key)
	return t.do(key, func() error {
		return t.trie.TryDelete(key)
	})
}

func (t *odrTrie) TryDelete(key []byte) error {
	key = crypto.Keccak256(key)
	return t.do(key, func() error {
		return t.trie.TryDelete(key)
	})
}

func (t *odrTrie) Commit(onleaf trie.LeafCallback) (common.Hash, error) {
	if t.trie == nil {
		return t.id.Root, nil
	}
	return t.trie.Commit(onleaf)
}

func (t *odrTrie) Hash() common.Hash {
	if t.trie == nil {
		return t.id.Root
	}
	return t.trie.Hash()
}

func (t *odrTrie) NodeIterator(startkey []byte) trie.NodeIterator {
	return newNodeIterator(t, startkey)
}

func (t *odrTrie) GetKey(sha []byte) []byte {
	return nil
}

func (t *odrTrie) Prove(key []byte, fromLevel uint, proofDb mandb.Putter) error {
	return errors.New("not implemented, needs client/server interface split")
}

// do tries and retries to execute a function until it returns with no error or
// an error type other than MissingNodeError
func (t *odrTrie) do(key []byte, fn func() error) error {
	for {
		var err error
		if t.trie == nil {
			t.trie, err = trie.New(t.id.Root, trie.NewDatabase(t.db.backend.Database()))
		}
		if err == nil {
			err = fn()
		}
		if _, ok := err.(*trie.MissingNodeError); !ok {
			return err
		}
		r := &TrieRequest{Id: t.id, Key: key}
		if err := t.db.backend.Retrieve(t.db.ctx, r); err != nil {
			return err
		}
	}
}

type nodeIterator struct {
	trie.NodeIterator
	t   *odrTrie
	err error
}

func newNodeIterator(t *odrTrie, startkey []byte) trie.NodeIterator {
	it := &nodeIterator{t: t}
	// Open the actual non-ODR trie if that hasn't happened yet.
	if t.trie == nil {
		it.do(func() error {
			t, err := trie.New(t.id.Root, trie.NewDatabase(t.db.backend.Database()))
			if err == nil {
				it.t.trie = t
			}
			return err
		})
	}
	it.do(func() error {
		it.NodeIterator = it.t.trie.NodeIterator(startkey)
		return it.NodeIterator.Error()
	})
	return it
}

func (it *nodeIterator) Next(descend bool) bool {
	var ok bool
	it.do(func() error {
		ok = it.NodeIterator.Next(descend)
		return it.NodeIterator.Error()
	})
	return ok
}

// do runs fn and attempts to fill in missing nodes by retrieving.
func (it *nodeIterator) do(fn func() error) {
	var lasthash common.Hash
	for {
		it.err = fn()
		missing, ok := it.err.(*trie.MissingNodeError)
		if !ok {
			return
		}
		if missing.NodeHash == lasthash {
			it.err = fmt.Errorf("retrieve loop for trie node %x", missing.NodeHash)
			return
		}
		lasthash = missing.NodeHash
		r := &TrieRequest{Id: it.t.id, Key: nibblesToKey(missing.Path)}
		if it.err = it.t.db.backend.Retrieve(it.t.db.ctx, r); it.err != nil {
			return
		}
	}
}

func (it *nodeIterator) Error() error {
	if it.err != nil {
		return it.err
	}
	return it.NodeIterator.Error()
}

func nibblesToKey(nib []byte) []byte {
	if len(nib) > 0 && nib[len(nib)-1] == 0x10 {
		nib = nib[:len(nib)-1] // drop terminator
	}
	if len(nib)&1 == 1 {
		nib = append(nib, 0) // make even
	}
	key := make([]byte, len(nib)/2)
	for bi, ni := 0, 0; ni < len(nib); bi, ni = bi+1, ni+2 {
		key[bi] = nib[ni]<<4 | nib[ni+1]
	}
	return key
}
