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
// Helper for converting SafeArray to array of objects.

package ole

import (
	"unsafe"
)

type SafeArrayConversion struct {
	Array *SafeArray
}

func (sac *SafeArrayConversion) ToStringArray() (strings []string) {
	totalElements, _ := sac.TotalElements(0)
	strings = make([]string, totalElements)

	for i := int64(0); i < totalElements; i++ {
		strings[int32(i)], _ = safeArrayGetElementString(sac.Array, i)
	}

	return
}

func (sac *SafeArrayConversion) ToByteArray() (bytes []byte) {
	totalElements, _ := sac.TotalElements(0)
	bytes = make([]byte, totalElements)

	for i := int64(0); i < totalElements; i++ {
		safeArrayGetElement(sac.Array, i, unsafe.Pointer(&bytes[int32(i)]))
	}

	return
}

func (sac *SafeArrayConversion) ToValueArray() (values []interface{}) {
	totalElements, _ := sac.TotalElements(0)
	values = make([]interface{}, totalElements)
	vt, _ := safeArrayGetVartype(sac.Array)

	for i := 0; i < int(totalElements); i++ {
		switch VT(vt) {
		case VT_BOOL:
			var v bool
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_I1:
			var v int8
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_I2:
			var v int16
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_I4:
			var v int32
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_I8:
			var v int64
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_UI1:
			var v uint8
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_UI2:
			var v uint16
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_UI4:
			var v uint32
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_UI8:
			var v uint64
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_R4:
			var v float32
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_R8:
			var v float64
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_BSTR:
			var v string
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v
		case VT_VARIANT:
			var v VARIANT
			safeArrayGetElement(sac.Array, int64(i), unsafe.Pointer(&v))
			values[i] = v.Value()
		default:
			// TODO
		}
	}

	return
}

func (sac *SafeArrayConversion) GetType() (varType uint16, err error) {
	return safeArrayGetVartype(sac.Array)
}

func (sac *SafeArrayConversion) GetDimensions() (dimensions *uint32, err error) {
	return safeArrayGetDim(sac.Array)
}

func (sac *SafeArrayConversion) GetSize() (length *uint32, err error) {
	return safeArrayGetElementSize(sac.Array)
}

func (sac *SafeArrayConversion) TotalElements(index uint32) (totalElements int64, err error) {
	if index < 1 {
		index = 1
	}

	// Get array bounds
	var LowerBounds int64
	var UpperBounds int64

	LowerBounds, err = safeArrayGetLBound(sac.Array, index)
	if err != nil {
		return
	}

	UpperBounds, err = safeArrayGetUBound(sac.Array, index)
	if err != nil {
		return
	}

	totalElements = UpperBounds - LowerBounds + 1
	return
}

// Release Safe Array memory
func (sac *SafeArrayConversion) Release() {
	safeArrayDestroy(sac.Array)
}
