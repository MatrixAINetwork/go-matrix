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
	"fmt"
	"path/filepath"
	"time"

	"github.com/matrix/go-matrix/log"
)

/*
NetStore is a cloud storage access abstaction layer for swarm
it contains the shared logic of network served chunk store/retrieval requests
both local (coming from DPA api) and remote (coming from peers via bzz protocol)
it implements the ChunkStore interface and embeds LocalStore

It is called by the bzz protocol instances via Depo (the store/retrieve request handler)
a protocol instance is running on each peer, so this is heavily parallelised.
NetStore falls back to a backend (CloudStorage interface)
implemented by bzz/network/forwarder. forwarder or IPFS or IPΞS
*/
type NetStore struct {
	hashfunc   SwarmHasher
	localStore *LocalStore
	cloud      CloudStore
}

// backend engine for cloud store
// It can be aggregate dispatching to several parallel implementations:
// bzz/network/forwarder. forwarder or IPFS or IPΞS
type CloudStore interface {
	Store(*Chunk)
	Deliver(*Chunk)
	Retrieve(*Chunk)
}

type StoreParams struct {
	ChunkDbPath   string
	DbCapacity    uint64
	CacheCapacity uint
	Radius        int
}

//create params with default values
func NewDefaultStoreParams() (self *StoreParams) {
	return &StoreParams{
		DbCapacity:    defaultDbCapacity,
		CacheCapacity: defaultCacheCapacity,
		Radius:        defaultRadius,
	}
}

//this can only finally be set after all config options (file, cmd line, env vars)
//have been evaluated
func (self *StoreParams) Init(path string) {
	self.ChunkDbPath = filepath.Join(path, "chunks")
}

// netstore contructor, takes path argument that is used to initialise dbStore,
// the persistent (disk) storage component of LocalStore
// the second argument is the hive, the connection/logistics manager for the node
func NewNetStore(hash SwarmHasher, lstore *LocalStore, cloud CloudStore, params *StoreParams) *NetStore {
	return &NetStore{
		hashfunc:   hash,
		localStore: lstore,
		cloud:      cloud,
	}
}

var (
	// timeout interval before retrieval is timed out
	searchTimeout = 3 * time.Second
)

// store logic common to local and network chunk store requests
// ~ unsafe put in localdb no check if exists no extra copy no hash validation
// the chunk is forced to propagate (Cloud.Store) even if locally found!
// caller needs to make sure if that is wanted
func (self *NetStore) Put(entry *Chunk) {
	self.localStore.Put(entry)

	// handle deliveries
	if entry.Req != nil {
		log.Trace(fmt.Sprintf("NetStore.Put: localStore.Put %v hit existing request...delivering", entry.Key.Log()))
		// closing C signals to other routines (local requests)
		// that the chunk is has been retrieved
		close(entry.Req.C)
		// deliver the chunk to requesters upstream
		go self.cloud.Deliver(entry)
	} else {
		log.Trace(fmt.Sprintf("NetStore.Put: localStore.Put %v stored locally", entry.Key.Log()))
		// handle propagating store requests
		// go self.cloud.Store(entry)
		go self.cloud.Store(entry)
	}
}

// retrieve logic common for local and network chunk retrieval requests
func (self *NetStore) Get(key Key) (*Chunk, error) {
	var err error
	chunk, err := self.localStore.Get(key)
	if err == nil {
		if chunk.Req == nil {
			log.Trace(fmt.Sprintf("NetStore.Get: %v found locally", key))
		} else {
			log.Trace(fmt.Sprintf("NetStore.Get: %v hit on an existing request", key))
			// no need to launch again
		}
		return chunk, err
	}
	// no data and no request status
	log.Trace(fmt.Sprintf("NetStore.Get: %v not found locally. open new request", key))
	chunk = NewChunk(key, newRequestStatus(key))
	self.localStore.memStore.Put(chunk)
	go self.cloud.Retrieve(chunk)
	return chunk, nil
}

// Close netstore
func (self *NetStore) Close() {}
