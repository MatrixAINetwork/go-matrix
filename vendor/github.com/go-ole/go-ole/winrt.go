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
	"reflect"
	"syscall"
	"unicode/utf8"
	"unsafe"
)

var (
	procRoInitialize              = modcombase.NewProc("RoInitialize")
	procRoActivateInstance        = modcombase.NewProc("RoActivateInstance")
	procRoGetActivationFactory    = modcombase.NewProc("RoGetActivationFactory")
	procWindowsCreateString       = modcombase.NewProc("WindowsCreateString")
	procWindowsDeleteString       = modcombase.NewProc("WindowsDeleteString")
	procWindowsGetStringRawBuffer = modcombase.NewProc("WindowsGetStringRawBuffer")
)

func RoInitialize(thread_type uint32) (err error) {
	hr, _, _ := procRoInitialize.Call(uintptr(thread_type))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}

func RoActivateInstance(clsid string) (ins *IInspectable, err error) {
	hClsid, err := NewHString(clsid)
	if err != nil {
		return nil, err
	}
	defer DeleteHString(hClsid)

	hr, _, _ := procRoActivateInstance.Call(
		uintptr(unsafe.Pointer(hClsid)),
		uintptr(unsafe.Pointer(&ins)))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}

func RoGetActivationFactory(clsid string, iid *GUID) (ins *IInspectable, err error) {
	hClsid, err := NewHString(clsid)
	if err != nil {
		return nil, err
	}
	defer DeleteHString(hClsid)

	hr, _, _ := procRoGetActivationFactory.Call(
		uintptr(unsafe.Pointer(hClsid)),
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&ins)))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}

// HString is handle string for pointers.
type HString uintptr

// NewHString returns a new HString for Go string.
func NewHString(s string) (hstring HString, err error) {
	u16 := syscall.StringToUTF16Ptr(s)
	len := uint32(utf8.RuneCountInString(s))
	hr, _, _ := procWindowsCreateString.Call(
		uintptr(unsafe.Pointer(u16)),
		uintptr(len),
		uintptr(unsafe.Pointer(&hstring)))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}

// DeleteHString deletes HString.
func DeleteHString(hstring HString) (err error) {
	hr, _, _ := procWindowsDeleteString.Call(uintptr(hstring))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}

// String returns Go string value of HString.
func (h HString) String() string {
	var u16buf uintptr
	var u16len uint32
	u16buf, _, _ = procWindowsGetStringRawBuffer.Call(
		uintptr(h),
		uintptr(unsafe.Pointer(&u16len)))

	u16hdr := reflect.SliceHeader{Data: u16buf, Len: int(u16len), Cap: int(u16len)}
	u16 := *(*[]uint16)(unsafe.Pointer(&u16hdr))
	return syscall.UTF16ToString(u16)
}
