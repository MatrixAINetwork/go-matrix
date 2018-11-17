// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// simple nonconcurrent reference implementation for hashsize segment based
// Binary Merkle tree hash on arbitrary but fixed maximum chunksize
//
// This implementation does not take advantage of any paralellisms and uses
// far more memory than necessary, but it is easy to see that it is correct.
// It can be used for generating test cases for optimized implementations.
// see testBMTHasherCorrectness function in bmt_test.go
package bmt

import (
	"hash"
)

// RefHasher is the non-optimized easy to read reference implementation of BMT
type RefHasher struct {
	span    int
	section int
	cap     int
	h       hash.Hash
}

// NewRefHasher returns a new RefHasher
func NewRefHasher(hasher BaseHasher, count int) *RefHasher {
	h := hasher()
	hashsize := h.Size()
	maxsize := hashsize * count
	c := 2
	for ; c < count; c *= 2 {
	}
	if c > 2 {
		c /= 2
	}
	return &RefHasher{
		section: 2 * hashsize,
		span:    c * hashsize,
		cap:     maxsize,
		h:       h,
	}
}

// Hash returns the BMT hash of the byte slice
// implements the SwarmHash interface
func (rh *RefHasher) Hash(d []byte) []byte {
	if len(d) > rh.cap {
		d = d[:rh.cap]
	}

	return rh.hash(d, rh.span)
}

func (rh *RefHasher) hash(d []byte, s int) []byte {
	l := len(d)
	left := d
	var right []byte
	if l > rh.section {
		for ; s >= l; s /= 2 {
		}
		left = rh.hash(d[:s], s)
		right = d[s:]
		if l-s > rh.section/2 {
			right = rh.hash(right, s)
		}
	}
	defer rh.h.Reset()
	rh.h.Write(left)
	rh.h.Write(right)
	h := rh.h.Sum(nil)
	return h
}
