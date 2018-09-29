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
package colorable

import (
	"bytes"
	"io"
)

// NonColorable hold writer but remove escape sequence.
type NonColorable struct {
	out     io.Writer
	lastbuf bytes.Buffer
}

// NewNonColorable return new instance of Writer which remove escape sequence from Writer.
func NewNonColorable(w io.Writer) io.Writer {
	return &NonColorable{out: w}
}

// Write write data on console
func (w *NonColorable) Write(data []byte) (n int, err error) {
	er := bytes.NewReader(data)
	var bw [1]byte
loop:
	for {
		c1, err := er.ReadByte()
		if err != nil {
			break loop
		}
		if c1 != 0x1b {
			bw[0] = c1
			w.out.Write(bw[:])
			continue
		}
		c2, err := er.ReadByte()
		if err != nil {
			w.lastbuf.WriteByte(c1)
			break loop
		}
		if c2 != 0x5b {
			w.lastbuf.WriteByte(c1)
			w.lastbuf.WriteByte(c2)
			continue
		}

		var buf bytes.Buffer
		for {
			c, err := er.ReadByte()
			if err != nil {
				w.lastbuf.WriteByte(c1)
				w.lastbuf.WriteByte(c2)
				w.lastbuf.Write(buf.Bytes())
				break loop
			}
			if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '@' {
				break
			}
			buf.Write([]byte(string(c)))
		}
	}
	return len(data) - w.lastbuf.Len(), nil
}
