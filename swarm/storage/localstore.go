// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package storage

import (
	"encoding/binary"

	"github.com/matrix/go-matrix/metrics"
)

//metrics variables
var (
	dbStorePutCounter = metrics.NewRegisteredCounter("storage.db.dbstore.put.count", nil)
)

// LocalStore is a combination of inmemory db over a disk persisted db
// implements a Get/Put with fallback (caching) logic using any 2 ChunkStores
type LocalStore struct {
	memStore ChunkStore
	DbStore  ChunkStore
}

// This constructor uses MemStore and DbStore as components
func NewLocalStore(hash SwarmHasher, params *StoreParams) (*LocalStore, error) {
	dbStore, err := NewDbStore(params.ChunkDbPath, hash, params.DbCapacity, params.Radius)
	if err != nil {
		return nil, err
	}
	return &LocalStore{
		memStore: NewMemStore(dbStore, params.CacheCapacity),
		DbStore:  dbStore,
	}, nil
}

func (self *LocalStore) CacheCounter() uint64 {
	return uint64(self.memStore.(*MemStore).Counter())
}

func (self *LocalStore) DbCounter() uint64 {
	return self.DbStore.(*DbStore).Counter()
}

// LocalStore is itself a chunk store
// unsafe, in that the data is not integrity checked
func (self *LocalStore) Put(chunk *Chunk) {
	chunk.dbStored = make(chan bool)
	self.memStore.Put(chunk)
	if chunk.wg != nil {
		chunk.wg.Add(1)
	}
	go func() {
		dbStorePutCounter.Inc(1)
		self.DbStore.Put(chunk)
		if chunk.wg != nil {
			chunk.wg.Done()
		}
	}()
}

// Get(chunk *Chunk) looks up a chunk in the local stores
// This method is blocking until the chunk is retrieved
// so additional timeout may be needed to wrap this call if
// ChunkStores are remote and can have long latency
func (self *LocalStore) Get(key Key) (chunk *Chunk, err error) {
	chunk, err = self.memStore.Get(key)
	if err == nil {
		return
	}
	chunk, err = self.DbStore.Get(key)
	if err != nil {
		return
	}
	chunk.Size = int64(binary.LittleEndian.Uint64(chunk.SData[0:8]))
	self.memStore.Put(chunk)
	return
}

// Close local store
func (self *LocalStore) Close() {}
