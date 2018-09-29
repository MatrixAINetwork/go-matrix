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
package ole

import "unsafe"

type IUnknown struct {
	RawVTable *interface{}
}

type IUnknownVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

type UnknownLike interface {
	QueryInterface(iid *GUID) (disp *IDispatch, err error)
	AddRef() int32
	Release() int32
}

func (v *IUnknown) VTable() *IUnknownVtbl {
	return (*IUnknownVtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *IUnknown) PutQueryInterface(interfaceID *GUID, obj interface{}) error {
	return reflectQueryInterface(v, v.VTable().QueryInterface, interfaceID, obj)
}

func (v *IUnknown) IDispatch(interfaceID *GUID) (dispatch *IDispatch, err error) {
	err = v.PutQueryInterface(interfaceID, &dispatch)
	return
}

func (v *IUnknown) IEnumVARIANT(interfaceID *GUID) (enum *IEnumVARIANT, err error) {
	err = v.PutQueryInterface(interfaceID, &enum)
	return
}

func (v *IUnknown) QueryInterface(iid *GUID) (*IDispatch, error) {
	return queryInterface(v, iid)
}

func (v *IUnknown) MustQueryInterface(iid *GUID) (disp *IDispatch) {
	unk, err := queryInterface(v, iid)
	if err != nil {
		panic(err)
	}
	return unk
}

func (v *IUnknown) AddRef() int32 {
	return addRef(v)
}

func (v *IUnknown) Release() int32 {
	return release(v)
}
