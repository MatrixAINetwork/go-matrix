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

// RoInitialize
func RoInitialize(thread_type uint32) (err error) {
	return NewError(E_NOTIMPL)
}

// RoActivateInstance
func RoActivateInstance(clsid string) (ins *IInspectable, err error) {
	return nil, NewError(E_NOTIMPL)
}

// RoGetActivationFactory
func RoGetActivationFactory(clsid string, iid *GUID) (ins *IInspectable, err error) {
	return nil, NewError(E_NOTIMPL)
}

// HString is handle string for pointers.
type HString uintptr

// NewHString returns a new HString for Go string.
func NewHString(s string) (hstring HString, err error) {
	return HString(uintptr(0)), NewError(E_NOTIMPL)
}

// DeleteHString deletes HString.
func DeleteHString(hstring HString) (err error) {
	return NewError(E_NOTIMPL)
}

// String returns Go string value of HString.
func (h HString) String() string {
	return ""
}
