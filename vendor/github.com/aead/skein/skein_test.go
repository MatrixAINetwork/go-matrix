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

package skein

import (
	"bytes"
	"encoding/hex"
	"hash"
	"testing"
)

func testWrite(msg string, t *testing.T, h hash.Hash, c *Config) {
	var msg1 []byte
	msg0 := make([]byte, 64)
	for i := range msg0 {
		h.Write(msg0[:i])
		msg1 = append(msg1, msg0[:i]...)
	}
	tag0 := h.Sum(nil)
	tag1 := Sum(msg1, h.Size(), c)

	if !bytes.Equal(tag0, tag1) {
		t.Fatalf("%s\nSum differ from Sum\n Sum: %s \n skein.Sum: %s", msg, hex.EncodeToString(tag0), hex.EncodeToString(tag1))
	}
}

func TestWrite(t *testing.T) {
	testWrite("testWrite(t, New256(nil), nil)", t, New256(nil), nil)
	testWrite("testWrite(t, New256(make([]byte, 16)), &Config{Key: make([]byte, 16)})", t, New256(make([]byte, 16)), &Config{Key: make([]byte, 16)})

	testWrite("testWrite(t, New512(nil), nil)", t, New512(nil), nil)
	testWrite("testWrite(t, New512(make([]byte, 16)), &Config{Key: make([]byte, 16)})", t, New512(make([]byte, 16)), &Config{Key: make([]byte, 16)})

	testWrite("testWrite(t, New(128, nil), nil)", t, New(128, nil), nil)
	testWrite("testWrite(t, New(128, &Config{Key: make([]byte, 16)}), &Config{Key: make([]byte, 16)})", t, New(128, &Config{Key: make([]byte, 16)}), &Config{Key: make([]byte, 16)})
}

func TestBlockSize(t *testing.T) {
	h := New(64, nil)
	if bs := h.BlockSize(); bs != BlockSize {
		t.Fatalf("BlockSize() returned: %d - but expected: %d", bs, BlockSize)
	}
}

func TestSum(t *testing.T) {
	sizes := []int{20, 32, 48, 64}
	key := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	msg := make([]byte, 512)
	for i := range msg {
		msg[i] = byte(i) + key[i%len(key)]
	}
	c := &Config{Key: key}

	for _, hashsize := range sizes {
		switch hashsize {
		case 20:
			{
				var sum0 [20]byte
				Sum160(&sum0, msg, key)
				sum1 := Sum(msg, hashsize, c)
				if !bytes.Equal(sum0[:], sum1) {
					t.Fatalf("Sum160 differ from Sum: Sum160: %s\n Sum: %s", hex.EncodeToString(sum0[:]), hex.EncodeToString(sum1))
				}
			}
		case 32:
			{
				var sum0 [32]byte
				Sum256(&sum0, msg, key)
				sum1 := Sum(msg, hashsize, c)
				if !bytes.Equal(sum0[:], sum1) {
					t.Fatalf("Sum256 differ from Sum: Sum256: %s\n Sum: %s", hex.EncodeToString(sum0[:]), hex.EncodeToString(sum1))
				}
			}
		case 48:
			{
				var sum0 [48]byte
				Sum384(&sum0, msg, key)
				sum1 := Sum(msg, hashsize, c)
				if !bytes.Equal(sum0[:], sum1) {
					t.Fatalf("Sum384 differ from Sum: Sum384: %s\n Sum: %s", hex.EncodeToString(sum0[:]), hex.EncodeToString(sum1))
				}
			}
		case 64:
			{
				var sum0 [64]byte
				Sum512(&sum0, msg, key)
				sum1 := Sum(msg, hashsize, c)
				if !bytes.Equal(sum0[:], sum1) {
					t.Fatalf("Sum512 differ from Sum: Sum512: %s\n Sum: %s", hex.EncodeToString(sum0[:]), hex.EncodeToString(sum1))
				}
			}
		}
	}
}

func TestInitialize(t *testing.T) {
	rec := func() {
		if err := recover(); err == nil {
			t.Fatal("Recover expected error, but no one occured")
		}
	}
	mustFail := func() {
		defer rec()
		s := new(hashFunc)
		s.initialize(0, nil)
	}
	mustFail()

	c := &Config{
		Key:       make([]byte, 16),
		KeyID:     make([]byte, 16),
		Personal:  make([]byte, 8),
		Nonce:     make([]byte, 12),
		PublicKey: make([]byte, 128),
	}
	testWrite("testWrite(t, New(64, c), c)", t, New(64, c), c)
}

// Benchmarks

func benchmarkSum(b *testing.B, size int) {
	msg := make([]byte, size)
	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sum(msg, 64, nil)
	}
}

func benchmarkSum512(b *testing.B, size int) {
	var sum [64]byte
	msg := make([]byte, size)
	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sum512(&sum, msg, nil)
	}
}

func BenchmarkSum_64(b *testing.B)    { benchmarkSum(b, 64) }
func BenchmarkSum_1K(b *testing.B)    { benchmarkSum(b, 1024) }
func BenchmarkSum512_64(b *testing.B) { benchmarkSum512(b, 64) }
func BenchmarkSum512_1K(b *testing.B) { benchmarkSum512(b, 1024) }

func benchmarkWrite(b *testing.B, size int) {
	h := New512(nil)
	msg := make([]byte, size)
	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Write(msg)
	}
}

func BenchmarkWrite_64(b *testing.B) { benchmarkWrite(b, 64) }
func BenchmarkWrite_1K(b *testing.B) { benchmarkWrite(b, 1024) }
