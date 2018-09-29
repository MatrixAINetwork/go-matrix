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
package fuse

import "unsafe"

// buffer provides a mechanism for constructing a message from
// multiple segments.
type buffer []byte

// alloc allocates size bytes and returns a pointer to the new
// segment.
func (w *buffer) alloc(size uintptr) unsafe.Pointer {
	s := int(size)
	if len(*w)+s > cap(*w) {
		old := *w
		*w = make([]byte, len(*w), 2*cap(*w)+s)
		copy(*w, old)
	}
	l := len(*w)
	*w = (*w)[:l+s]
	return unsafe.Pointer(&(*w)[l])
}

// reset clears out the contents of the buffer.
func (w *buffer) reset() {
	for i := range (*w)[:cap(*w)] {
		(*w)[i] = 0
	}
	*w = (*w)[:0]
}

func newBuffer(extra uintptr) buffer {
	const hdrSize = unsafe.Sizeof(outHeader{})
	buf := make(buffer, hdrSize, hdrSize+extra)
	return buf
}
