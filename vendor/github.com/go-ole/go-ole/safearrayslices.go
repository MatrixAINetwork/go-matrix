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
// +build windows

package ole

import (
	"unsafe"
)

func safeArrayFromByteSlice(slice []byte) *SafeArray {
	array, _ := safeArrayCreateVector(VT_UI1, 0, uint32(len(slice)))

	if array == nil {
		panic("Could not convert []byte to SAFEARRAY")
	}

	for i, v := range slice {
		safeArrayPutElement(array, int64(i), uintptr(unsafe.Pointer(&v)))
	}
	return array
}

func safeArrayFromStringSlice(slice []string) *SafeArray {
	array, _ := safeArrayCreateVector(VT_BSTR, 0, uint32(len(slice)))

	if array == nil {
		panic("Could not convert []string to SAFEARRAY")
	}
	// SysAllocStringLen(s)
	for i, v := range slice {
		safeArrayPutElement(array, int64(i), uintptr(unsafe.Pointer(SysAllocStringLen(v))))
	}
	return array
}
