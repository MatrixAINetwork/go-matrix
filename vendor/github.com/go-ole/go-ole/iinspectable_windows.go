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
	"bytes"
	"encoding/binary"
	"reflect"
	"syscall"
	"unsafe"
)

func (v *IInspectable) GetIids() (iids []*GUID, err error) {
	var count uint32
	var array uintptr
	hr, _, _ := syscall.Syscall(
		v.VTable().GetIIds,
		3,
		uintptr(unsafe.Pointer(v)),
		uintptr(unsafe.Pointer(&count)),
		uintptr(unsafe.Pointer(&array)))
	if hr != 0 {
		err = NewError(hr)
		return
	}
	defer CoTaskMemFree(array)

	iids = make([]*GUID, count)
	byteCount := count * uint32(unsafe.Sizeof(GUID{}))
	slicehdr := reflect.SliceHeader{Data: array, Len: int(byteCount), Cap: int(byteCount)}
	byteSlice := *(*[]byte)(unsafe.Pointer(&slicehdr))
	reader := bytes.NewReader(byteSlice)
	for i := range iids {
		guid := GUID{}
		err = binary.Read(reader, binary.LittleEndian, &guid)
		if err != nil {
			return
		}
		iids[i] = &guid
	}
	return
}

func (v *IInspectable) GetRuntimeClassName() (s string, err error) {
	var hstring HString
	hr, _, _ := syscall.Syscall(
		v.VTable().GetRuntimeClassName,
		2,
		uintptr(unsafe.Pointer(v)),
		uintptr(unsafe.Pointer(&hstring)),
		0)
	if hr != 0 {
		err = NewError(hr)
		return
	}
	s = hstring.String()
	DeleteHString(hstring)
	return
}

func (v *IInspectable) GetTrustLevel() (level uint32, err error) {
	hr, _, _ := syscall.Syscall(
		v.VTable().GetTrustLevel,
		2,
		uintptr(unsafe.Pointer(v)),
		uintptr(unsafe.Pointer(&level)),
		0)
	if hr != 0 {
		err = NewError(hr)
	}
	return
}
