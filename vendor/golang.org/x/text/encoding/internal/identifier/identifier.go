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

//go:generate go run gen.go

// Package identifier defines the contract between implementations of Encoding
// and Index by defining identifiers that uniquely identify standardized coded
// character sets (CCS) and character encoding schemes (CES), which we will
// together refer to as encodings, for which Encoding implementations provide
// converters to and from UTF-8. This package is typically only of concern to
// implementers of Indexes and Encodings.
//
// One part of the identifier is the MIB code, which is defined by IANA and
// uniquely identifies a CCS or CES. Each code is associated with data that
// references authorities, official documentation as well as aliases and MIME
// names.
//
// Not all CESs are covered by the IANA registry. The "other" string that is
// returned by ID can be used to identify other character sets or versions of
// existing ones.
//
// It is recommended that each package that provides a set of Encodings provide
// the All and Common variables to reference all supported encodings and
// commonly used subset. This allows Index implementations to include all
// available encodings without explicitly referencing or knowing about them.
package identifier

// Note: this package is internal, but could be made public if there is a need
// for writing third-party Indexes and Encodings.

// References:
// - http://source.icu-project.org/repos/icu/icu/trunk/source/data/mappings/convrtrs.txt
// - http://www.iana.org/assignments/character-sets/character-sets.xhtml
// - http://www.iana.org/assignments/ianacharset-mib/ianacharset-mib
// - http://www.ietf.org/rfc/rfc2978.txt
// - http://www.unicode.org/reports/tr22/
// - http://www.w3.org/TR/encoding/
// - https://encoding.spec.whatwg.org/
// - https://encoding.spec.whatwg.org/encodings.json
// - https://tools.ietf.org/html/rfc6657#section-5

// Interface can be implemented by Encodings to define the CCS or CES for which
// it implements conversions.
type Interface interface {
	// ID returns an encoding identifier. Exactly one of the mib and other
	// values should be non-zero.
	//
	// In the usual case it is only necessary to indicate the MIB code. The
	// other string can be used to specify encodings for which there is no MIB,
	// such as "x-mac-dingbat".
	//
	// The other string may only contain the characters a-z, A-Z, 0-9, - and _.
	ID() (mib MIB, other string)

	// NOTE: the restrictions on the encoding are to allow extending the syntax
	// with additional information such as versions, vendors and other variants.
}

// A MIB identifies an encoding. It is derived from the IANA MIB codes and adds
// some identifiers for some encodings that are not covered by the IANA
// standard.
//
// See http://www.iana.org/assignments/ianacharset-mib.
type MIB uint16

// These additional MIB types are not defined in IANA. They are added because
// they are common and defined within the text repo.
const (
	// Unofficial marks the start of encodings not registered by IANA.
	Unofficial MIB = 10000 + iota

	// Replacement is the WhatWG replacement encoding.
	Replacement

	// XUserDefined is the code for x-user-defined.
	XUserDefined

	// MacintoshCyrillic is the code for x-mac-cyrillic.
	MacintoshCyrillic
)
