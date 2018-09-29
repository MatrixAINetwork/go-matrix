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
// +build !windows

package ole

import (
	"unsafe"
)

// safeArrayAccessData returns raw array pointer.
//
// AKA: SafeArrayAccessData in Windows API.
func safeArrayAccessData(safearray *SafeArray) (uintptr, error) {
	return uintptr(0), NewError(E_NOTIMPL)
}

// safeArrayUnaccessData releases raw array.
//
// AKA: SafeArrayUnaccessData in Windows API.
func safeArrayUnaccessData(safearray *SafeArray) error {
	return NewError(E_NOTIMPL)
}

// safeArrayAllocData allocates SafeArray.
//
// AKA: SafeArrayAllocData in Windows API.
func safeArrayAllocData(safearray *SafeArray) error {
	return NewError(E_NOTIMPL)
}

// safeArrayAllocDescriptor allocates SafeArray.
//
// AKA: SafeArrayAllocDescriptor in Windows API.
func safeArrayAllocDescriptor(dimensions uint32) (*SafeArray, error) {
	return nil, NewError(E_NOTIMPL)
}

// safeArrayAllocDescriptorEx allocates SafeArray.
//
// AKA: SafeArrayAllocDescriptorEx in Windows API.
func safeArrayAllocDescriptorEx(variantType VT, dimensions uint32) (*SafeArray, error) {
	return nil, NewError(E_NOTIMPL)
}

// safeArrayCopy returns copy of SafeArray.
//
// AKA: SafeArrayCopy in Windows API.
func safeArrayCopy(original *SafeArray) (*SafeArray, error) {
	return nil, NewError(E_NOTIMPL)
}

// safeArrayCopyData duplicates SafeArray into another SafeArray object.
//
// AKA: SafeArrayCopyData in Windows API.
func safeArrayCopyData(original *SafeArray, duplicate *SafeArray) error {
	return NewError(E_NOTIMPL)
}

// safeArrayCreate creates SafeArray.
//
// AKA: SafeArrayCreate in Windows API.
func safeArrayCreate(variantType VT, dimensions uint32, bounds *SafeArrayBound) (*SafeArray, error) {
	return nil, NewError(E_NOTIMPL)
}

// safeArrayCreateEx creates SafeArray.
//
// AKA: SafeArrayCreateEx in Windows API.
func safeArrayCreateEx(variantType VT, dimensions uint32, bounds *SafeArrayBound, extra uintptr) (*SafeArray, error) {
	return nil, NewError(E_NOTIMPL)
}

// safeArrayCreateVector creates SafeArray.
//
// AKA: SafeArrayCreateVector in Windows API.
func safeArrayCreateVector(variantType VT, lowerBound int32, length uint32) (*SafeArray, error) {
	return nil, NewError(E_NOTIMPL)
}

// safeArrayCreateVectorEx creates SafeArray.
//
// AKA: SafeArrayCreateVectorEx in Windows API.
func safeArrayCreateVectorEx(variantType VT, lowerBound int32, length uint32, extra uintptr) (*SafeArray, error) {
	return nil, NewError(E_NOTIMPL)
}

// safeArrayDestroy destroys SafeArray object.
//
// AKA: SafeArrayDestroy in Windows API.
func safeArrayDestroy(safearray *SafeArray) error {
	return NewError(E_NOTIMPL)
}

// safeArrayDestroyData destroys SafeArray object.
//
// AKA: SafeArrayDestroyData in Windows API.
func safeArrayDestroyData(safearray *SafeArray) error {
	return NewError(E_NOTIMPL)
}

// safeArrayDestroyDescriptor destroys SafeArray object.
//
// AKA: SafeArrayDestroyDescriptor in Windows API.
func safeArrayDestroyDescriptor(safearray *SafeArray) error {
	return NewError(E_NOTIMPL)
}

// safeArrayGetDim is the amount of dimensions in the SafeArray.
//
// SafeArrays may have multiple dimensions. Meaning, it could be
// multidimensional array.
//
// AKA: SafeArrayGetDim in Windows API.
func safeArrayGetDim(safearray *SafeArray) (*uint32, error) {
	u := uint32(0)
	return &u, NewError(E_NOTIMPL)
}

// safeArrayGetElementSize is the element size in bytes.
//
// AKA: SafeArrayGetElemsize in Windows API.
func safeArrayGetElementSize(safearray *SafeArray) (*uint32, error) {
	u := uint32(0)
	return &u, NewError(E_NOTIMPL)
}

// safeArrayGetElement retrieves element at given index.
func safeArrayGetElement(safearray *SafeArray, index int64, pv unsafe.Pointer) error {
	return NewError(E_NOTIMPL)
}

// safeArrayGetElement retrieves element at given index and converts to string.
func safeArrayGetElementString(safearray *SafeArray, index int64) (string, error) {
	return "", NewError(E_NOTIMPL)
}

// safeArrayGetIID is the InterfaceID of the elements in the SafeArray.
//
// AKA: SafeArrayGetIID in Windows API.
func safeArrayGetIID(safearray *SafeArray) (*GUID, error) {
	return nil, NewError(E_NOTIMPL)
}

// safeArrayGetLBound returns lower bounds of SafeArray.
//
// SafeArrays may have multiple dimensions. Meaning, it could be
// multidimensional array.
//
// AKA: SafeArrayGetLBound in Windows API.
func safeArrayGetLBound(safearray *SafeArray, dimension uint32) (int64, error) {
	return int64(0), NewError(E_NOTIMPL)
}

// safeArrayGetUBound returns upper bounds of SafeArray.
//
// SafeArrays may have multiple dimensions. Meaning, it could be
// multidimensional array.
//
// AKA: SafeArrayGetUBound in Windows API.
func safeArrayGetUBound(safearray *SafeArray, dimension uint32) (int64, error) {
	return int64(0), NewError(E_NOTIMPL)
}

// safeArrayGetVartype returns data type of SafeArray.
//
// AKA: SafeArrayGetVartype in Windows API.
func safeArrayGetVartype(safearray *SafeArray) (uint16, error) {
	return uint16(0), NewError(E_NOTIMPL)
}

// safeArrayLock locks SafeArray for reading to modify SafeArray.
//
// This must be called during some calls to ensure that another process does not
// read or write to the SafeArray during editing.
//
// AKA: SafeArrayLock in Windows API.
func safeArrayLock(safearray *SafeArray) error {
	return NewError(E_NOTIMPL)
}

// safeArrayUnlock unlocks SafeArray for reading.
//
// AKA: SafeArrayUnlock in Windows API.
func safeArrayUnlock(safearray *SafeArray) error {
	return NewError(E_NOTIMPL)
}

// safeArrayPutElement stores the data element at the specified location in the
// array.
//
// AKA: SafeArrayPutElement in Windows API.
func safeArrayPutElement(safearray *SafeArray, index int64, element uintptr) error {
	return NewError(E_NOTIMPL)
}

// safeArrayGetRecordInfo accesses IRecordInfo info for custom types.
//
// AKA: SafeArrayGetRecordInfo in Windows API.
//
// XXX: Must implement IRecordInfo interface for this to return.
func safeArrayGetRecordInfo(safearray *SafeArray) (interface{}, error) {
	return nil, NewError(E_NOTIMPL)
}

// safeArraySetRecordInfo mutates IRecordInfo info for custom types.
//
// AKA: SafeArraySetRecordInfo in Windows API.
//
// XXX: Must implement IRecordInfo interface for this to return.
func safeArraySetRecordInfo(safearray *SafeArray, recordInfo interface{}) error {
	return NewError(E_NOTIMPL)
}
