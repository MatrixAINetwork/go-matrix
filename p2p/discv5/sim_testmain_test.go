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

// +build go1.4,nacl,faketime_simulation

package discv5

import (
	"os"
	"runtime"
	"testing"
	"unsafe"
)

// Enable fake time mode in the runtime, like on the go playground.
// There is a slight chance that this won't work because some go code
// might have executed before the variable is set.

//go:linkname faketime runtime.faketime
var faketime = 1

func TestMain(m *testing.M) {
	// We need to use unsafe somehow in order to get access to go:linkname.
	_ = unsafe.Sizeof(0)

	// Run the actual test. runWithPlaygroundTime ensures that the only test
	// that runs is the one calling it.
	runtime.GOMAXPROCS(8)
	os.Exit(m.Run())
}
