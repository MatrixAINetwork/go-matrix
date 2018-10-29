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

// +build none
//sed -e 's/_N_/Hash/g' -e 's/_S_/32/g' -e '1d' types_template.go | gofmt -w hash.go

package common

import "math/big"

type _N_ [_S_]byte

func BytesTo_N_(b []byte) _N_ {
	var h _N_
	h.SetBytes(b)
	return h
}
func StringTo_N_(s string) _N_ { return BytesTo_N_([]byte(s)) }
func BigTo_N_(b *big.Int) _N_  { return BytesTo_N_(b.Bytes()) }
func HexTo_N_(s string) _N_    { return BytesTo_N_(FromHex(s)) }

// Don't use the default 'String' method in case we want to overwrite

// Get the string representation of the underlying hash
func (h _N_) Str() string   { return string(h[:]) }
func (h _N_) Bytes() []byte { return h[:] }
func (h _N_) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }
func (h _N_) Hex() string   { return "0x" + Bytes2Hex(h[:]) }

// Sets the hash to the value of b. If b is larger than len(h) it will panic
func (h *_N_) SetBytes(b []byte) {
	// Use the right most bytes
	if len(b) > len(h) {
		b = b[len(b)-_S_:]
	}

	// Reverse the loop
	for i := len(b) - 1; i >= 0; i-- {
		h[_S_-len(b)+i] = b[i]
	}
}

// Set string `s` to h. If s is larger than len(h) it will panic
func (h *_N_) SetString(s string) { h.SetBytes([]byte(s)) }

// Sets h to other
func (h *_N_) Set(other _N_) {
	for i, v := range other {
		h[i] = v
	}
}
