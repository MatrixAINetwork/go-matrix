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
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/matrix/go-matrix/log"
)

/*
DPA provides the client API entrypoints Store and Retrieve to store and retrieve
It can store anything that has a byte slice representation, so files or serialised objects etc.

Storage: DPA calls the Chunker to segment the input datastream of any size to a merkle hashed tree of chunks. The key of the root block is returned to the client.

Retrieval: given the key of the root block, the DPA retrieves the block chunks and reconstructs the original data and passes it back as a lazy reader. A lazy reader is a reader with on-demand delayed processing, i.e. the chunks needed to reconstruct a large file are only fetched and processed if that particular part of the document is actually read.

As the chunker produces chunks, DPA dispatches them to its own chunk store
implementation for storage or retrieval.
*/

const (
	storeChanCapacity           = 100
	retrieveChanCapacity        = 100
	singletonSwarmDbCapacity    = 50000
	singletonSwarmCacheCapacity = 500
	maxStoreProcesses           = 8
	maxRetrieveProcesses        = 8
)

var (
	notFound = errors.New("not found")
)

type DPA struct {
	ChunkStore
	storeC    chan *Chunk
	retrieveC chan *Chunk
	Chunker   Chunker

	lock    sync.Mutex
	running bool
	quitC   chan bool
}

// for testing locally
func NewLocalDPA(datadir string) (*DPA, error) {

	hash := MakeHashFunc("SHA256")

	dbStore, err := NewDbStore(datadir, hash, singletonSwarmDbCapacity, 0)
	if err != nil {
		return nil, err
	}

	return NewDPA(&LocalStore{
		NewMemStore(dbStore, singletonSwarmCacheCapacity),
		dbStore,
	}, NewChunkerParams()), nil
}

func NewDPA(store ChunkStore, params *ChunkerParams) *DPA {
	chunker := NewTreeChunker(params)
	return &DPA{
		Chunker:    chunker,
		ChunkStore: store,
	}
}

// Public API. Main entry point for document retrieval directly. Used by the
// FS-aware API and httpaccess
// Chunk retrieval blocks on netStore requests with a timeout so reader will
// report error if retrieval of chunks within requested range time out.
func (self *DPA) Retrieve(key Key) LazySectionReader {
	return self.Chunker.Join(key, self.retrieveC)
}

// Public API. Main entry point for document storage directly. Used by the
// FS-aware API and httpaccess
func (self *DPA) Store(data io.Reader, size int64, swg *sync.WaitGroup, wwg *sync.WaitGroup) (key Key, err error) {
	return self.Chunker.Split(data, size, self.storeC, swg, wwg)
}

func (self *DPA) Start() {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.running {
		return
	}
	self.running = true
	self.retrieveC = make(chan *Chunk, retrieveChanCapacity)
	self.storeC = make(chan *Chunk, storeChanCapacity)
	self.quitC = make(chan bool)
	self.storeLoop()
	self.retrieveLoop()
}

func (self *DPA) Stop() {
	self.lock.Lock()
	defer self.lock.Unlock()
	if !self.running {
		return
	}
	self.running = false
	close(self.quitC)
}

// retrieveLoop dispatches the parallel chunk retrieval requests received on the
// retrieve channel to its ChunkStore  (NetStore or LocalStore)
func (self *DPA) retrieveLoop() {
	for i := 0; i < maxRetrieveProcesses; i++ {
		go self.retrieveWorker()
	}
	log.Trace(fmt.Sprintf("dpa: retrieve loop spawning %v workers", maxRetrieveProcesses))
}

func (self *DPA) retrieveWorker() {
	for chunk := range self.retrieveC {
		log.Trace(fmt.Sprintf("dpa: retrieve loop : chunk %v", chunk.Key.Log()))
		storedChunk, err := self.Get(chunk.Key)
		if err == notFound {
			log.Trace(fmt.Sprintf("chunk %v not found", chunk.Key.Log()))
		} else if err != nil {
			log.Trace(fmt.Sprintf("error retrieving chunk %v: %v", chunk.Key.Log(), err))
		} else {
			chunk.SData = storedChunk.SData
			chunk.Size = storedChunk.Size
		}
		close(chunk.C)

		select {
		case <-self.quitC:
			return
		default:
		}
	}
}

// storeLoop dispatches the parallel chunk store request processors
// received on the store channel to its ChunkStore (NetStore or LocalStore)
func (self *DPA) storeLoop() {
	for i := 0; i < maxStoreProcesses; i++ {
		go self.storeWorker()
	}
	log.Trace(fmt.Sprintf("dpa: store spawning %v workers", maxStoreProcesses))
}

func (self *DPA) storeWorker() {

	for chunk := range self.storeC {
		self.Put(chunk)
		if chunk.wg != nil {
			log.Trace(fmt.Sprintf("dpa: store processor %v", chunk.Key.Log()))
			chunk.wg.Done()

		}
		select {
		case <-self.quitC:
			return
		default:
		}
	}
}

// DpaChunkStore implements the ChunkStore interface,
// this chunk access layer assumed 2 chunk stores
// local storage eg. LocalStore and network storage eg., NetStore
// access by calling network is blocking with a timeout

type dpaChunkStore struct {
	n          int
	localStore ChunkStore
	netStore   ChunkStore
}

func NewDpaChunkStore(localStore, netStore ChunkStore) *dpaChunkStore {
	return &dpaChunkStore{0, localStore, netStore}
}

// Get is the entrypoint for local retrieve requests
// waits for response or times out
func (self *dpaChunkStore) Get(key Key) (chunk *Chunk, err error) {
	chunk, err = self.netStore.Get(key)
	// timeout := time.Now().Add(searchTimeout)
	if chunk.SData != nil {
		log.Trace(fmt.Sprintf("DPA.Get: %v found locally, %d bytes", key.Log(), len(chunk.SData)))
		return
	}
	// TODO: use self.timer time.Timer and reset with defer disableTimer
	timer := time.After(searchTimeout)
	select {
	case <-timer:
		log.Trace(fmt.Sprintf("DPA.Get: %v request time out ", key.Log()))
		err = notFound
	case <-chunk.Req.C:
		log.Trace(fmt.Sprintf("DPA.Get: %v retrieved, %d bytes (%p)", key.Log(), len(chunk.SData), chunk))
	}
	return
}

// Put is the entrypoint for local store requests coming from storeLoop
func (self *dpaChunkStore) Put(entry *Chunk) {
	chunk, err := self.localStore.Get(entry.Key)
	if err != nil {
		log.Trace(fmt.Sprintf("DPA.Put: %v new chunk. call netStore.Put", entry.Key.Log()))
		chunk = entry
	} else if chunk.SData == nil {
		log.Trace(fmt.Sprintf("DPA.Put: %v request entry found", entry.Key.Log()))
		chunk.SData = entry.SData
		chunk.Size = entry.Size
	} else {
		log.Trace(fmt.Sprintf("DPA.Put: %v chunk already known", entry.Key.Log()))
		return
	}
	// from this point on the storage logic is the same with network storage requests
	log.Trace(fmt.Sprintf("DPA.Put %v: %v", self.n, chunk.Key.Log()))
	self.n++
	self.netStore.Put(chunk)
}

// Close chunk store
func (self *dpaChunkStore) Close() {}
