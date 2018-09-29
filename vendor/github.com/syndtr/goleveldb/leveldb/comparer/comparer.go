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

// Package comparer provides interface and implementation for ordering
// sets of data.
package comparer

// BasicComparer is the interface that wraps the basic Compare method.
type BasicComparer interface {
	// Compare returns -1, 0, or +1 depending on whether a is 'less than',
	// 'equal to' or 'greater than' b. The two arguments can only be 'equal'
	// if their contents are exactly equal. Furthermore, the empty slice
	// must be 'less than' any non-empty slice.
	Compare(a, b []byte) int
}

// Comparer defines a total ordering over the space of []byte keys: a 'less
// than' relationship.
type Comparer interface {
	BasicComparer

	// Name returns name of the comparer.
	//
	// The Level-DB on-disk format stores the comparer name, and opening a
	// database with a different comparer from the one it was created with
	// will result in an error.
	//
	// An implementation to a new name whenever the comparer implementation
	// changes in a way that will cause the relative ordering of any two keys
	// to change.
	//
	// Names starting with "leveldb." are reserved and should not be used
	// by any users of this package.
	Name() string

	// Bellow are advanced functions used used to reduce the space requirements
	// for internal data structures such as index blocks.

	// Separator appends a sequence of bytes x to dst such that a <= x && x < b,
	// where 'less than' is consistent with Compare. An implementation should
	// return nil if x equal to a.
	//
	// Either contents of a or b should not by any means modified. Doing so
	// may cause corruption on the internal state.
	Separator(dst, a, b []byte) []byte

	// Successor appends a sequence of bytes x to dst such that x >= b, where
	// 'less than' is consistent with Compare. An implementation should return
	// nil if x equal to b.
	//
	// Contents of b should not by any means modified. Doing so may cause
	// corruption on the internal state.
	Successor(dst, b []byte) []byte
}
