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
package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb/storage"
	"sync/atomic"
)

type iStorage struct {
	storage.Storage
	read  uint64
	write uint64
}

func (c *iStorage) Open(fd storage.FileDesc) (storage.Reader, error) {
	r, err := c.Storage.Open(fd)
	return &iStorageReader{r, c}, err
}

func (c *iStorage) Create(fd storage.FileDesc) (storage.Writer, error) {
	w, err := c.Storage.Create(fd)
	return &iStorageWriter{w, c}, err
}

func (c *iStorage) reads() uint64 {
	return atomic.LoadUint64(&c.read)
}

func (c *iStorage) writes() uint64 {
	return atomic.LoadUint64(&c.write)
}

// newIStorage returns the given storage wrapped by iStorage.
func newIStorage(s storage.Storage) *iStorage {
	return &iStorage{s, 0, 0}
}

type iStorageReader struct {
	storage.Reader
	c *iStorage
}

func (r *iStorageReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	atomic.AddUint64(&r.c.read, uint64(n))
	return n, err
}

func (r *iStorageReader) ReadAt(p []byte, off int64) (n int, err error) {
	n, err = r.Reader.ReadAt(p, off)
	atomic.AddUint64(&r.c.read, uint64(n))
	return n, err
}

type iStorageWriter struct {
	storage.Writer
	c *iStorage
}

func (w *iStorageWriter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	atomic.AddUint64(&w.c.write, uint64(n))
	return n, err
}
