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
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

const testDataSize = 0x1000000

func TestDPArandom(t *testing.T) {
	dbStore := initDbStore(t)
	dbStore.setCapacity(50000)
	memStore := NewMemStore(dbStore, defaultCacheCapacity)
	localStore := &LocalStore{
		memStore,
		dbStore,
	}
	chunker := NewTreeChunker(NewChunkerParams())
	dpa := &DPA{
		Chunker:    chunker,
		ChunkStore: localStore,
	}
	dpa.Start()
	defer dpa.Stop()
	defer os.RemoveAll("/tmp/bzz")

	reader, slice := testDataReaderAndSlice(testDataSize)
	wg := &sync.WaitGroup{}
	key, err := dpa.Store(reader, testDataSize, wg, nil)
	if err != nil {
		t.Errorf("Store error: %v", err)
	}
	wg.Wait()
	resultReader := dpa.Retrieve(key)
	resultSlice := make([]byte, len(slice))
	n, err := resultReader.ReadAt(resultSlice, 0)
	if err != io.EOF {
		t.Errorf("Retrieve error: %v", err)
	}
	if n != len(slice) {
		t.Errorf("Slice size error got %d, expected %d.", n, len(slice))
	}
	if !bytes.Equal(slice, resultSlice) {
		t.Errorf("Comparison error.")
	}
	ioutil.WriteFile("/tmp/slice.bzz.16M", slice, 0666)
	ioutil.WriteFile("/tmp/result.bzz.16M", resultSlice, 0666)
	localStore.memStore = NewMemStore(dbStore, defaultCacheCapacity)
	resultReader = dpa.Retrieve(key)
	for i := range resultSlice {
		resultSlice[i] = 0
	}
	n, err = resultReader.ReadAt(resultSlice, 0)
	if err != io.EOF {
		t.Errorf("Retrieve error after removing memStore: %v", err)
	}
	if n != len(slice) {
		t.Errorf("Slice size error after removing memStore got %d, expected %d.", n, len(slice))
	}
	if !bytes.Equal(slice, resultSlice) {
		t.Errorf("Comparison error after removing memStore.")
	}
}

func TestDPA_capacity(t *testing.T) {
	dbStore := initDbStore(t)
	memStore := NewMemStore(dbStore, defaultCacheCapacity)
	localStore := &LocalStore{
		memStore,
		dbStore,
	}
	memStore.setCapacity(0)
	chunker := NewTreeChunker(NewChunkerParams())
	dpa := &DPA{
		Chunker:    chunker,
		ChunkStore: localStore,
	}
	dpa.Start()
	reader, slice := testDataReaderAndSlice(testDataSize)
	wg := &sync.WaitGroup{}
	key, err := dpa.Store(reader, testDataSize, wg, nil)
	if err != nil {
		t.Errorf("Store error: %v", err)
	}
	wg.Wait()
	resultReader := dpa.Retrieve(key)
	resultSlice := make([]byte, len(slice))
	n, err := resultReader.ReadAt(resultSlice, 0)
	if err != io.EOF {
		t.Errorf("Retrieve error: %v", err)
	}
	if n != len(slice) {
		t.Errorf("Slice size error got %d, expected %d.", n, len(slice))
	}
	if !bytes.Equal(slice, resultSlice) {
		t.Errorf("Comparison error.")
	}
	// Clear memStore
	memStore.setCapacity(0)
	// check whether it is, indeed, empty
	dpa.ChunkStore = memStore
	resultReader = dpa.Retrieve(key)
	if _, err = resultReader.ReadAt(resultSlice, 0); err == nil {
		t.Errorf("Was able to read %d bytes from an empty memStore.", len(slice))
	}
	// check how it works with localStore
	dpa.ChunkStore = localStore
	//	localStore.dbStore.setCapacity(0)
	resultReader = dpa.Retrieve(key)
	for i := range resultSlice {
		resultSlice[i] = 0
	}
	n, err = resultReader.ReadAt(resultSlice, 0)
	if err != io.EOF {
		t.Errorf("Retrieve error after clearing memStore: %v", err)
	}
	if n != len(slice) {
		t.Errorf("Slice size error after clearing memStore got %d, expected %d.", n, len(slice))
	}
	if !bytes.Equal(slice, resultSlice) {
		t.Errorf("Comparison error after clearing memStore.")
	}
}
