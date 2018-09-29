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
package base64vlq

import (
	"io"
)

const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

const (
	vlqBaseShift       = 5
	vlqBase            = 1 << vlqBaseShift
	vlqBaseMask        = vlqBase - 1
	vlqSignBit         = 1
	vlqContinuationBit = vlqBase
)

var decodeMap [256]byte

func init() {
	for i := 0; i < len(encodeStd); i++ {
		decodeMap[encodeStd[i]] = byte(i)
	}
}

func toVLQSigned(n int) int {
	if n < 0 {
		return -n<<1 + 1
	}
	return n << 1
}

func fromVLQSigned(n int) int {
	isNeg := n&vlqSignBit != 0
	n >>= 1
	if isNeg {
		return -n
	}
	return n
}

type Encoder struct {
	w io.ByteWriter
}

func NewEncoder(w io.ByteWriter) *Encoder {
	return &Encoder{
		w: w,
	}
}

func (enc Encoder) Encode(n int) error {
	n = toVLQSigned(n)
	for digit := vlqContinuationBit; digit&vlqContinuationBit != 0; {
		digit = n & vlqBaseMask
		n >>= vlqBaseShift
		if n > 0 {
			digit |= vlqContinuationBit
		}

		err := enc.w.WriteByte(encodeStd[digit])
		if err != nil {
			return err
		}
	}
	return nil
}

type Decoder struct {
	r io.ByteReader
}

func NewDecoder(r io.ByteReader) *Decoder {
	return &Decoder{
		r: r,
	}
}

func (dec Decoder) Decode() (n int, err error) {
	shift := uint(0)
	for continuation := true; continuation; {
		c, err := dec.r.ReadByte()
		if err != nil {
			return 0, err
		}

		c = decodeMap[c]
		continuation = c&vlqContinuationBit != 0
		n += int(c&vlqBaseMask) << shift
		shift += vlqBaseShift
	}
	return fromVLQSigned(n), nil
}
