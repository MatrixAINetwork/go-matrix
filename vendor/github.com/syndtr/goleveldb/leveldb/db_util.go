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
// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Reader is the interface that wraps basic Get and NewIterator methods.
// This interface implemented by both DB and Snapshot.
type Reader interface {
	Get(key []byte, ro *opt.ReadOptions) (value []byte, err error)
	NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator
}

// Sizes is list of size.
type Sizes []int64

// Sum returns sum of the sizes.
func (sizes Sizes) Sum() int64 {
	var sum int64
	for _, size := range sizes {
		sum += size
	}
	return sum
}

// Logging.
func (db *DB) log(v ...interface{})                 { db.s.log(v...) }
func (db *DB) logf(format string, v ...interface{}) { db.s.logf(format, v...) }

// Check and clean files.
func (db *DB) checkAndCleanFiles() error {
	v := db.s.version()
	defer v.release()

	tmap := make(map[int64]bool)
	for _, tables := range v.levels {
		for _, t := range tables {
			tmap[t.fd.Num] = false
		}
	}

	fds, err := db.s.stor.List(storage.TypeAll)
	if err != nil {
		return err
	}

	var nt int
	var rem []storage.FileDesc
	for _, fd := range fds {
		keep := true
		switch fd.Type {
		case storage.TypeManifest:
			keep = fd.Num >= db.s.manifestFd.Num
		case storage.TypeJournal:
			if !db.frozenJournalFd.Zero() {
				keep = fd.Num >= db.frozenJournalFd.Num
			} else {
				keep = fd.Num >= db.journalFd.Num
			}
		case storage.TypeTable:
			_, keep = tmap[fd.Num]
			if keep {
				tmap[fd.Num] = true
				nt++
			}
		}

		if !keep {
			rem = append(rem, fd)
		}
	}

	if nt != len(tmap) {
		var mfds []storage.FileDesc
		for num, present := range tmap {
			if !present {
				mfds = append(mfds, storage.FileDesc{storage.TypeTable, num})
				db.logf("db@janitor table missing @%d", num)
			}
		}
		return errors.NewErrCorrupted(storage.FileDesc{}, &errors.ErrMissingFiles{Fds: mfds})
	}

	db.logf("db@janitor F·%d G·%d", len(fds), len(rem))
	for _, fd := range rem {
		db.logf("db@janitor removing %s-%d", fd.Type, fd.Num)
		if err := db.s.stor.Remove(fd); err != nil {
			return err
		}
	}
	return nil
}
