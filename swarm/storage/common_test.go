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
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/matrix/go-matrix/log"
)

type brokenLimitedReader struct {
	lr    io.Reader
	errAt int
	off   int
	size  int
}

func brokenLimitReader(data io.Reader, size int, errAt int) *brokenLimitedReader {
	return &brokenLimitedReader{
		lr:    data,
		errAt: errAt,
		size:  size,
	}
}

func testDataReader(l int) (r io.Reader) {
	return io.LimitReader(rand.Reader, int64(l))
}

func (self *brokenLimitedReader) Read(buf []byte) (int, error) {
	if self.off+len(buf) > self.errAt {
		return 0, fmt.Errorf("Broken reader")
	}
	self.off += len(buf)
	return self.lr.Read(buf)
}

func testDataReaderAndSlice(l int) (r io.Reader, slice []byte) {
	slice = make([]byte, l)
	if _, err := rand.Read(slice); err != nil {
		panic("rand error")
	}
	r = io.LimitReader(bytes.NewReader(slice), int64(l))
	return
}

func testStore(m ChunkStore, l int64, branches int64, t *testing.T) {

	chunkC := make(chan *Chunk)
	go func() {
		for chunk := range chunkC {
			m.Put(chunk)
			if chunk.wg != nil {
				chunk.wg.Done()
			}
		}
	}()
	chunker := NewTreeChunker(&ChunkerParams{
		Branches: branches,
		Hash:     SHA3Hash,
	})
	swg := &sync.WaitGroup{}
	key, _ := chunker.Split(rand.Reader, l, chunkC, swg, nil)
	swg.Wait()
	close(chunkC)
	chunkC = make(chan *Chunk)

	quit := make(chan bool)

	go func() {
		for ch := range chunkC {
			go func(chunk *Chunk) {
				storedChunk, err := m.Get(chunk.Key)
				if err == notFound {
					log.Trace(fmt.Sprintf("chunk '%v' not found", chunk.Key.Log()))
				} else if err != nil {
					log.Trace(fmt.Sprintf("error retrieving chunk %v: %v", chunk.Key.Log(), err))
				} else {
					chunk.SData = storedChunk.SData
					chunk.Size = storedChunk.Size
				}
				log.Trace(fmt.Sprintf("chunk '%v' not found", chunk.Key.Log()))
				close(chunk.C)
			}(ch)
		}
		close(quit)
	}()
	r := chunker.Join(key, chunkC)

	b := make([]byte, l)
	n, err := r.ReadAt(b, 0)
	if err != io.EOF {
		t.Fatalf("read error (%v/%v) %v", n, l, err)
	}
	close(chunkC)
	<-quit
}
