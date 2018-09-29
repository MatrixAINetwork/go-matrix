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
// Copyright (c) 2013, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package storage

import (
	"bytes"
	"os"
	"sync"
)

const typeShift = 4

// Verify at compile-time that typeShift is large enough to cover all FileType
// values by confirming that 0 == 0.
var _ [0]struct{} = [TypeAll >> typeShift]struct{}{}

type memStorageLock struct {
	ms *memStorage
}

func (lock *memStorageLock) Unlock() {
	ms := lock.ms
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.slock == lock {
		ms.slock = nil
	}
	return
}

// memStorage is a memory-backed storage.
type memStorage struct {
	mu    sync.Mutex
	slock *memStorageLock
	files map[uint64]*memFile
	meta  FileDesc
}

// NewMemStorage returns a new memory-backed storage implementation.
func NewMemStorage() Storage {
	return &memStorage{
		files: make(map[uint64]*memFile),
	}
}

func (ms *memStorage) Lock() (Locker, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.slock != nil {
		return nil, ErrLocked
	}
	ms.slock = &memStorageLock{ms: ms}
	return ms.slock, nil
}

func (*memStorage) Log(str string) {}

func (ms *memStorage) SetMeta(fd FileDesc) error {
	if !FileDescOk(fd) {
		return ErrInvalidFile
	}

	ms.mu.Lock()
	ms.meta = fd
	ms.mu.Unlock()
	return nil
}

func (ms *memStorage) GetMeta() (FileDesc, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.meta.Zero() {
		return FileDesc{}, os.ErrNotExist
	}
	return ms.meta, nil
}

func (ms *memStorage) List(ft FileType) ([]FileDesc, error) {
	ms.mu.Lock()
	var fds []FileDesc
	for x := range ms.files {
		fd := unpackFile(x)
		if fd.Type&ft != 0 {
			fds = append(fds, fd)
		}
	}
	ms.mu.Unlock()
	return fds, nil
}

func (ms *memStorage) Open(fd FileDesc) (Reader, error) {
	if !FileDescOk(fd) {
		return nil, ErrInvalidFile
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	if m, exist := ms.files[packFile(fd)]; exist {
		if m.open {
			return nil, errFileOpen
		}
		m.open = true
		return &memReader{Reader: bytes.NewReader(m.Bytes()), ms: ms, m: m}, nil
	}
	return nil, os.ErrNotExist
}

func (ms *memStorage) Create(fd FileDesc) (Writer, error) {
	if !FileDescOk(fd) {
		return nil, ErrInvalidFile
	}

	x := packFile(fd)
	ms.mu.Lock()
	defer ms.mu.Unlock()
	m, exist := ms.files[x]
	if exist {
		if m.open {
			return nil, errFileOpen
		}
		m.Reset()
	} else {
		m = &memFile{}
		ms.files[x] = m
	}
	m.open = true
	return &memWriter{memFile: m, ms: ms}, nil
}

func (ms *memStorage) Remove(fd FileDesc) error {
	if !FileDescOk(fd) {
		return ErrInvalidFile
	}

	x := packFile(fd)
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if _, exist := ms.files[x]; exist {
		delete(ms.files, x)
		return nil
	}
	return os.ErrNotExist
}

func (ms *memStorage) Rename(oldfd, newfd FileDesc) error {
	if !FileDescOk(oldfd) || !FileDescOk(newfd) {
		return ErrInvalidFile
	}
	if oldfd == newfd {
		return nil
	}

	oldx := packFile(oldfd)
	newx := packFile(newfd)
	ms.mu.Lock()
	defer ms.mu.Unlock()
	oldm, exist := ms.files[oldx]
	if !exist {
		return os.ErrNotExist
	}
	newm, exist := ms.files[newx]
	if (exist && newm.open) || oldm.open {
		return errFileOpen
	}
	delete(ms.files, oldx)
	ms.files[newx] = oldm
	return nil
}

func (*memStorage) Close() error { return nil }

type memFile struct {
	bytes.Buffer
	open bool
}

type memReader struct {
	*bytes.Reader
	ms     *memStorage
	m      *memFile
	closed bool
}

func (mr *memReader) Close() error {
	mr.ms.mu.Lock()
	defer mr.ms.mu.Unlock()
	if mr.closed {
		return ErrClosed
	}
	mr.m.open = false
	return nil
}

type memWriter struct {
	*memFile
	ms     *memStorage
	closed bool
}

func (*memWriter) Sync() error { return nil }

func (mw *memWriter) Close() error {
	mw.ms.mu.Lock()
	defer mw.ms.mu.Unlock()
	if mw.closed {
		return ErrClosed
	}
	mw.memFile.open = false
	return nil
}

func packFile(fd FileDesc) uint64 {
	return uint64(fd.Num)<<typeShift | uint64(fd.Type)
}

func unpackFile(x uint64) FileDesc {
	return FileDesc{FileType(x) & TypeAll, int64(x >> typeShift)}
}
