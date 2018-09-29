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
// Package escape contains utilities for escaping parts of InfluxQL
// and InfluxDB line protocol.
package escape // import "github.com/influxdata/influxdb/pkg/escape"

import (
	"bytes"
	"strings"
)

// Codes is a map of bytes to be escaped.
var Codes = map[byte][]byte{
	',': []byte(`\,`),
	'"': []byte(`\"`),
	' ': []byte(`\ `),
	'=': []byte(`\=`),
}

// Bytes escapes characters on the input slice, as defined by Codes.
func Bytes(in []byte) []byte {
	for b, esc := range Codes {
		in = bytes.Replace(in, []byte{b}, esc, -1)
	}
	return in
}

const escapeChars = `," =`

// IsEscaped returns whether b has any escaped characters,
// i.e. whether b seems to have been processed by Bytes.
func IsEscaped(b []byte) bool {
	for len(b) > 0 {
		i := bytes.IndexByte(b, '\\')
		if i < 0 {
			return false
		}

		if i+1 < len(b) && strings.IndexByte(escapeChars, b[i+1]) >= 0 {
			return true
		}
		b = b[i+1:]
	}
	return false
}

// AppendUnescaped appends the unescaped version of src to dst
// and returns the resulting slice.
func AppendUnescaped(dst, src []byte) []byte {
	var pos int
	for len(src) > 0 {
		next := bytes.IndexByte(src[pos:], '\\')
		if next < 0 || pos+next+1 >= len(src) {
			return append(dst, src...)
		}

		if pos+next+1 < len(src) && strings.IndexByte(escapeChars, src[pos+next+1]) >= 0 {
			if pos+next > 0 {
				dst = append(dst, src[:pos+next]...)
			}
			src = src[pos+next+1:]
			pos = 0
		} else {
			pos += next + 1
		}
	}

	return dst
}

// Unescape returns a new slice containing the unescaped version of in.
func Unescape(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}

	if bytes.IndexByte(in, '\\') == -1 {
		return in
	}

	i := 0
	inLen := len(in)

	// The output size will be no more than inLen. Preallocating the
	// capacity of the output is faster and uses less memory than
	// letting append() do its own (over)allocation.
	out := make([]byte, 0, inLen)

	for {
		if i >= inLen {
			break
		}
		if in[i] == '\\' && i+1 < inLen {
			switch in[i+1] {
			case ',':
				out = append(out, ',')
				i += 2
				continue
			case '"':
				out = append(out, '"')
				i += 2
				continue
			case ' ':
				out = append(out, ' ')
				i += 2
				continue
			case '=':
				out = append(out, '=')
				i += 2
				continue
			}
		}
		out = append(out, in[i])
		i += 1
	}
	return out
}
