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

// +build !amd64

package skein

func bytesToBlock(block *[8]uint64, src []byte) {
	for i := range block {
		j := i * 8
		block[i] = uint64(src[j]) | uint64(src[j+1])<<8 | uint64(src[j+2])<<16 |
			uint64(src[j+3])<<24 | uint64(src[j+4])<<32 | uint64(src[j+5])<<40 |
			uint64(src[j+6])<<48 | uint64(src[j+7])<<56
	}
}

func blockToBytes(dst []byte, block *[8]uint64) {
	i := 0
	for _, v := range block {
		dst[i] = byte(v)
		dst[i+1] = byte(v >> 8)
		dst[i+2] = byte(v >> 16)
		dst[i+3] = byte(v >> 24)
		dst[i+4] = byte(v >> 32)
		dst[i+5] = byte(v >> 40)
		dst[i+6] = byte(v >> 48)
		dst[i+7] = byte(v >> 56)
		i += 8
	}
}
