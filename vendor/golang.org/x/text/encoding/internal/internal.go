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

// Package internal contains code that is shared among encoding implementations.
package internal

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/internal/identifier"
	"golang.org/x/text/transform"
)

// Encoding is an implementation of the Encoding interface that adds the String
// and ID methods to an existing encoding.
type Encoding struct {
	encoding.Encoding
	Name string
	MIB  identifier.MIB
}

// _ verifies that Encoding implements identifier.Interface.
var _ identifier.Interface = (*Encoding)(nil)

func (e *Encoding) String() string {
	return e.Name
}

func (e *Encoding) ID() (mib identifier.MIB, other string) {
	return e.MIB, ""
}

// SimpleEncoding is an Encoding that combines two Transformers.
type SimpleEncoding struct {
	Decoder transform.Transformer
	Encoder transform.Transformer
}

func (e *SimpleEncoding) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{Transformer: e.Decoder}
}

func (e *SimpleEncoding) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{Transformer: e.Encoder}
}

// FuncEncoding is an Encoding that combines two functions returning a new
// Transformer.
type FuncEncoding struct {
	Decoder func() transform.Transformer
	Encoder func() transform.Transformer
}

func (e FuncEncoding) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{Transformer: e.Decoder()}
}

func (e FuncEncoding) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{Transformer: e.Encoder()}
}

// A RepertoireError indicates a rune is not in the repertoire of a destination
// encoding. It is associated with an encoding-specific suggested replacement
// byte.
type RepertoireError byte

// Error implements the error interrface.
func (r RepertoireError) Error() string {
	return "encoding: rune not supported by encoding."
}

// Replacement returns the replacement string associated with this error.
func (r RepertoireError) Replacement() byte { return byte(r) }

var ErrASCIIReplacement = RepertoireError(encoding.ASCIISub)
