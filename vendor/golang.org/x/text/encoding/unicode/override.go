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
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unicode

import (
	"golang.org/x/text/transform"
)

// BOMOverride returns a new decoder transformer that is identical to fallback,
// except that the presence of a Byte Order Mark at the start of the input
// causes it to switch to the corresponding Unicode decoding. It will only
// consider BOMs for UTF-8, UTF-16BE, and UTF-16LE.
//
// This differs from using ExpectBOM by allowing a BOM to switch to UTF-8, not
// just UTF-16 variants, and allowing falling back to any encoding scheme.
//
// This technique is recommended by the W3C for use in HTML 5: "For
// compatibility with deployed content, the byte order mark (also known as BOM)
// is considered more authoritative than anything else."
// http://www.w3.org/TR/encoding/#specification-hooks
//
// Using BOMOverride is mostly intended for use cases where the first characters
// of a fallback encoding are known to not be a BOM, for example, for valid HTML
// and most encodings.
func BOMOverride(fallback transform.Transformer) transform.Transformer {
	// TODO: possibly allow a variadic argument of unicode encodings to allow
	// specifying details of which fallbacks are supported as well as
	// specifying the details of the implementations. This would also allow for
	// support for UTF-32, which should not be supported by default.
	return &bomOverride{fallback: fallback}
}

type bomOverride struct {
	fallback transform.Transformer
	current  transform.Transformer
}

func (d *bomOverride) Reset() {
	d.current = nil
	d.fallback.Reset()
}

var (
	// TODO: we could use decode functions here, instead of allocating a new
	// decoder on every NewDecoder as IgnoreBOM decoders can be stateless.
	utf16le = UTF16(LittleEndian, IgnoreBOM)
	utf16be = UTF16(BigEndian, IgnoreBOM)
)

const utf8BOM = "\ufeff"

func (d *bomOverride) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	if d.current != nil {
		return d.current.Transform(dst, src, atEOF)
	}
	if len(src) < 3 && !atEOF {
		return 0, 0, transform.ErrShortSrc
	}
	d.current = d.fallback
	bomSize := 0
	if len(src) >= 2 {
		if src[0] == 0xFF && src[1] == 0xFE {
			d.current = utf16le.NewDecoder()
			bomSize = 2
		} else if src[0] == 0xFE && src[1] == 0xFF {
			d.current = utf16be.NewDecoder()
			bomSize = 2
		} else if len(src) >= 3 &&
			src[0] == utf8BOM[0] &&
			src[1] == utf8BOM[1] &&
			src[2] == utf8BOM[2] {
			d.current = transform.Nop
			bomSize = 3
		}
	}
	if bomSize < len(src) {
		nDst, nSrc, err = d.current.Transform(dst, src[bomSize:], atEOF)
	}
	return nDst, nSrc + bomSize, err
}
