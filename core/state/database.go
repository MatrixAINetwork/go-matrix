// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package state

import (
	"fmt"
	"sync"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/trie"
	lru "github.com/hashicorp/golang-lru"
)

// Trie cache generation limit after which to evict trie nodes from memory.
var MaxTrieCacheGen = uint16(120)

const (
	// Number of past tries to keep. This value is chosen such that
	// reasonable chain reorg depths will hit an existing trie.
	maxPastTries  = 12
	PastTriesSize = 1024

	// Number of codehash->size associations to keep.
	codeSizeCacheSize = 100000
)

// Database wraps access to tries and contract code.
type Database interface {
	// OpenTrie opens the main account trie.
	OpenTrie(root common.Hash) (Trie, error)

	// OpenStorageTrie opens the storage trie of an account.
	OpenStorageTrie(addrHash, root common.Hash) (Trie, error)

	// CopyTrie returns an independent copy of the given trie.
	CopyTrie(Trie) Trie

	// ContractCode retrieves a particular contract's code.
	ContractCode(addrHash, codeHash common.Hash) ([]byte, error)

	// ContractCodeSize retrieves a particular contracts code's size.
	ContractCodeSize(addrHash, codeHash common.Hash) (int, error)

	// TrieDB retrieves the low level trie database used for data storage.
	TrieDB() *trie.Database
}

// Trie is a Matrix Merkle Trie.
type Trie interface {
	TryGet(key []byte) ([]byte, error)
	TryUpdate(key, value []byte) error
	TryDelete(key []byte) error
	Commit(onleaf trie.LeafCallback) (common.Hash, error)
	Hash() common.Hash
	NodeIterator(startKey []byte) trie.NodeIterator
	GetKey([]byte) []byte // TODO(fjl): remove this when SecureTrie is removed
	Prove(key []byte, fromLevel uint, proofDb mandb.Putter) error
}

// NewDatabase creates a backing store for state. The returned database is safe for
// concurrent use and retains cached trie nodes in memory. The pool is an optional
// intermediate trie-node memory pool between the low level storage layer and the
// high level trie abstraction.
func NewDatabase(db mandb.Database) Database {
	csc, _ := lru.New(codeSizeCacheSize)
	past, _ := lru.New(PastTriesSize)
	return &cachingDB{
		db:            trie.NewDatabase(db),
		pastTries:     past,
		codeSizeCache: csc,
	}
}

type cachingDB struct {
	db        *trie.Database
	mu        sync.Mutex
	pastTries *lru.Cache
	//	pastTries     []*trie.SecureTrie
	codeSizeCache *lru.Cache
}

// OpenTrie opens the main account trie.
func (db *cachingDB) OpenTrie(root common.Hash) (Trie, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if pastTree, exist := db.pastTries.Get(root); exist {
		return cachedTrie{pastTree.(*trie.SecureTrie).Copy(), db}, nil
	} else {
		tr, err := trie.NewSecure(root, db.db, MaxTrieCacheGen)
		if err != nil {
			return nil, err
		}
		db.pastTries.Add(root, tr)
		return cachedTrie{tr.Copy(), db}, nil
	}
	/*
		for i := len(db.pastTries) - 1; i >= 0; i-- {
			if db.pastTries[i].Hash() == root {
				return cachedTrie{db.pastTries[i].Copy(), db}, nil
			}
		}
		tr, err := trie.NewSecure(root, db.db, MaxTrieCacheGen)
		if err != nil {
			return nil, err
		}
		return cachedTrie{tr, db}, nil
	*/
}

func (db *cachingDB) pushTrie(t *trie.SecureTrie) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.pastTries.Add(t.Hash(), t)
	/*
		if len(db.pastTries) >= maxPastTries {
			copy(db.pastTries, db.pastTries[1:])
			db.pastTries[len(db.pastTries)-1] = t
		} else {
			db.pastTries = append(db.pastTries, t)
		}
	*/
}

// OpenStorageTrie opens the storage trie of an account.
func (db *cachingDB) OpenStorageTrie(addrHash, root common.Hash) (Trie, error) {
	return trie.NewSecure(root, db.db, 0)
}

// CopyTrie returns an independent copy of the given trie.
func (db *cachingDB) CopyTrie(t Trie) Trie {
	switch t := t.(type) {
	case cachedTrie:
		return cachedTrie{t.SecureTrie.Copy(), db}
	case *trie.SecureTrie:
		return t.Copy()
	default:
		panic(fmt.Errorf("unknown trie type %T", t))
	}
}

// ContractCode retrieves a particular contract's code.
func (db *cachingDB) ContractCode(addrHash, codeHash common.Hash) ([]byte, error) {
	code, err := db.db.Node(codeHash)
	if err == nil {
		db.codeSizeCache.Add(codeHash, len(code))
	}
	return code, err
}

// ContractCodeSize retrieves a particular contracts code's size.
func (db *cachingDB) ContractCodeSize(addrHash, codeHash common.Hash) (int, error) {
	if cached, ok := db.codeSizeCache.Get(codeHash); ok {
		return cached.(int), nil
	}
	code, err := db.ContractCode(addrHash, codeHash)
	return len(code), err
}

// TrieDB retrieves any intermediate trie-node caching layer.
func (db *cachingDB) TrieDB() *trie.Database {
	return db.db
}

// cachedTrie inserts its trie into a cachingDB on commit.
type cachedTrie struct {
	*trie.SecureTrie
	db *cachingDB
}

func (m cachedTrie) Commit(onleaf trie.LeafCallback) (common.Hash, error) {
	root, err := m.SecureTrie.Commit(onleaf)
	if err == nil {
		m.db.pushTrie(m.SecureTrie)
	}
	return root, err
}

func (m cachedTrie) Prove(key []byte, fromLevel uint, proofDb mandb.Putter) error {
	return m.SecureTrie.Prove(key, fromLevel, proofDb)
}
