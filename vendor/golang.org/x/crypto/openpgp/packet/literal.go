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
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package packet

import (
	"encoding/binary"
	"io"
)

// LiteralData represents an encrypted file. See RFC 4880, section 5.9.
type LiteralData struct {
	IsBinary bool
	FileName string
	Time     uint32 // Unix epoch time. Either creation time or modification time. 0 means undefined.
	Body     io.Reader
}

// ForEyesOnly returns whether the contents of the LiteralData have been marked
// as especially sensitive.
func (l *LiteralData) ForEyesOnly() bool {
	return l.FileName == "_CONSOLE"
}

func (l *LiteralData) parse(r io.Reader) (err error) {
	var buf [256]byte

	_, err = readFull(r, buf[:2])
	if err != nil {
		return
	}

	l.IsBinary = buf[0] == 'b'
	fileNameLen := int(buf[1])

	_, err = readFull(r, buf[:fileNameLen])
	if err != nil {
		return
	}

	l.FileName = string(buf[:fileNameLen])

	_, err = readFull(r, buf[:4])
	if err != nil {
		return
	}

	l.Time = binary.BigEndian.Uint32(buf[:4])
	l.Body = r
	return
}

// SerializeLiteral serializes a literal data packet to w and returns a
// WriteCloser to which the data itself can be written and which MUST be closed
// on completion. The fileName is truncated to 255 bytes.
func SerializeLiteral(w io.WriteCloser, isBinary bool, fileName string, time uint32) (plaintext io.WriteCloser, err error) {
	var buf [4]byte
	buf[0] = 't'
	if isBinary {
		buf[0] = 'b'
	}
	if len(fileName) > 255 {
		fileName = fileName[:255]
	}
	buf[1] = byte(len(fileName))

	inner, err := serializeStreamHeader(w, packetTypeLiteralData)
	if err != nil {
		return
	}

	_, err = inner.Write(buf[:2])
	if err != nil {
		return
	}
	_, err = inner.Write([]byte(fileName))
	if err != nil {
		return
	}
	binary.BigEndian.PutUint32(buf[:], time)
	_, err = inner.Write(buf[:])
	if err != nil {
		return
	}

	plaintext = inner
	return
}
