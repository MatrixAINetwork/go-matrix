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

// Package leveldb provides implementation of LevelDB key/value database.
//
// Create or open a database:
//
//	// The returned DB instance is safe for concurrent use. Which mean that all
//	// DB's methods may be called concurrently from multiple goroutine.
//	db, err := leveldb.OpenFile("path/to/db", nil)
//	...
//	defer db.Close()
//	...
//
// Read or modify the database content:
//
//	// Remember that the contents of the returned slice should not be modified.
//	data, err := db.Get([]byte("key"), nil)
//	...
//	err = db.Put([]byte("key"), []byte("value"), nil)
//	...
//	err = db.Delete([]byte("key"), nil)
//	...
//
// Iterate over database content:
//
//	iter := db.NewIterator(nil, nil)
//	for iter.Next() {
//		// Remember that the contents of the returned slice should not be modified, and
//		// only valid until the next call to Next.
//		key := iter.Key()
//		value := iter.Value()
//		...
//	}
//	iter.Release()
//	err = iter.Error()
//	...
//
// Iterate over subset of database content with a particular prefix:
//	iter := db.NewIterator(util.BytesPrefix([]byte("foo-")), nil)
//	for iter.Next() {
//		// Use key/value.
//		...
//	}
//	iter.Release()
//	err = iter.Error()
//	...
//
// Seek-then-Iterate:
//
// 	iter := db.NewIterator(nil, nil)
// 	for ok := iter.Seek(key); ok; ok = iter.Next() {
// 		// Use key/value.
// 		...
// 	}
// 	iter.Release()
// 	err = iter.Error()
// 	...
//
// Iterate over subset of database content:
//
// 	iter := db.NewIterator(&util.Range{Start: []byte("foo"), Limit: []byte("xoo")}, nil)
// 	for iter.Next() {
// 		// Use key/value.
// 		...
// 	}
// 	iter.Release()
// 	err = iter.Error()
// 	...
//
// Batch writes:
//
//	batch := new(leveldb.Batch)
//	batch.Put([]byte("foo"), []byte("value"))
//	batch.Put([]byte("bar"), []byte("another value"))
//	batch.Delete([]byte("baz"))
//	err = db.Write(batch, nil)
//	...
//
// Use bloom filter:
//
//	o := &opt.Options{
//		Filter: filter.NewBloomFilter(10),
//	}
//	db, err := leveldb.OpenFile("path/to/db", o)
//	...
//	defer db.Close()
//	...
package leveldb
