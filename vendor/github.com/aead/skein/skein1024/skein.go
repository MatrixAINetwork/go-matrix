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
// Copyright (c) 2016 Andreas Auernhammer. All rights reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

// Package skein1024 implements the Skein1024 hash function
// based on the Threefish1024 tweakable block cipher.
package skein1024

import (
	"hash"

	"github.com/aead/skein"
)

// Sum512 computes the 512 bit Skein1024 checksum (or MAC if key is set) of msg
// and writes it to out. The key is optional and can be nil.
func Sum512(out *[64]byte, msg, key []byte) {
	var out1024 [128]byte

	s := new(hashFunc)
	s.initialize(64, &skein.Config{Key: key})

	s.Write(msg)

	s.finalizeHash()

	s.output(&out1024, 0)
	copy(out[:], out1024[:64])
}

// Sum384 computes the 384 bit Skein1024 checksum (or MAC if key is set) of msg
// and writes it to out. The key is optional and can be nil.
func Sum384(out *[48]byte, msg, key []byte) {
	var out1024 [128]byte

	s := new(hashFunc)
	s.initialize(48, &skein.Config{Key: key})

	s.Write(msg)

	s.finalizeHash()

	s.output(&out1024, 0)
	copy(out[:], out1024[:48])
}

// Sum256 computes the 256 bit Skein1024 checksum (or MAC if key is set) of msg
// and writes it to out. The key is optional and can be nil.
func Sum256(out *[32]byte, msg, key []byte) {
	var out1024 [128]byte

	s := new(hashFunc)
	s.initialize(32, &skein.Config{Key: key})

	s.Write(msg)

	s.finalizeHash()

	s.output(&out1024, 0)
	copy(out[:], out1024[:32])
}

// Sum160 computes the 160 bit Skein1024 checksum (or MAC if key is set) of msg
// and writes it to out. The key is optional and can be nil.
func Sum160(out *[20]byte, msg, key []byte) {
	var out1024 [128]byte

	s := new(hashFunc)
	s.initialize(20, &skein.Config{Key: key})

	s.Write(msg)

	s.finalizeHash()

	s.output(&out1024, 0)
	copy(out[:], out1024[:20])
}

// Sum returns the Skein1024 checksum with the given hash size of msg using the (optional)
// conf for configuration. The hashsize must be > 0.
func Sum(msg []byte, hashsize int, conf *skein.Config) []byte {
	s := New(hashsize, conf)
	s.Write(msg)
	return s.Sum(nil)
}

// New512 returns a hash.Hash computing the Skein1024 512 bit checksum.
// The key is optional and turns the hash into a MAC.
func New512(key []byte) hash.Hash {
	s := new(hashFunc)

	s.initialize(64, &skein.Config{Key: key})

	return s
}

// New256 returns a hash.Hash computing the Skein1024 256 bit checksum.
// The key is optional and turns the hash into a MAC.
func New256(key []byte) hash.Hash {
	s := new(hashFunc)

	s.initialize(32, &skein.Config{Key: key})

	return s
}

// New returns a hash.Hash computing the Skein1024 checksum with the given hash size.
// The conf is optional and configurates the hash.Hash
func New(hashsize int, conf *skein.Config) hash.Hash {
	s := new(hashFunc)
	s.initialize(hashsize, conf)
	return s
}
