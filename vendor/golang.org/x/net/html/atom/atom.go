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
// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package atom provides integer codes (also known as atoms) for a fixed set of
// frequently occurring HTML strings: tag names and attribute keys such as "p"
// and "id".
//
// Sharing an atom's name between all elements with the same tag can result in
// fewer string allocations when tokenizing and parsing HTML. Integer
// comparisons are also generally faster than string comparisons.
//
// The value of an atom's particular code is not guaranteed to stay the same
// between versions of this package. Neither is any ordering guaranteed:
// whether atom.H1 < atom.H2 may also change. The codes are not guaranteed to
// be dense. The only guarantees are that e.g. looking up "div" will yield
// atom.Div, calling atom.Div.String will return "div", and atom.Div != 0.
package atom // import "golang.org/x/net/html/atom"

// Atom is an integer code for a string. The zero value maps to "".
type Atom uint32

// String returns the atom's name.
func (a Atom) String() string {
	start := uint32(a >> 8)
	n := uint32(a & 0xff)
	if start+n > uint32(len(atomText)) {
		return ""
	}
	return atomText[start : start+n]
}

func (a Atom) string() string {
	return atomText[a>>8 : a>>8+a&0xff]
}

// fnv computes the FNV hash with an arbitrary starting value h.
func fnv(h uint32, s []byte) uint32 {
	for i := range s {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func match(s string, t []byte) bool {
	for i, c := range t {
		if s[i] != c {
			return false
		}
	}
	return true
}

// Lookup returns the atom whose name is s. It returns zero if there is no
// such atom. The lookup is case sensitive.
func Lookup(s []byte) Atom {
	if len(s) == 0 || len(s) > maxAtomLen {
		return 0
	}
	h := fnv(hash0, s)
	if a := table[h&uint32(len(table)-1)]; int(a&0xff) == len(s) && match(a.string(), s) {
		return a
	}
	if a := table[(h>>16)&uint32(len(table)-1)]; int(a&0xff) == len(s) && match(a.string(), s) {
		return a
	}
	return 0
}

// String returns a string whose contents are equal to s. In that sense, it is
// equivalent to string(s) but may be more efficient.
func String(s []byte) string {
	if a := Lookup(s); a != 0 {
		return a.String()
	}
	return string(s)
}
