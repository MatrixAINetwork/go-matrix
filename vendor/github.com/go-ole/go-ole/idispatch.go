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

type IDispatch struct {
	IUnknown
}

type IDispatchVtbl struct {
	IUnknownVtbl
	GetTypeInfoCount uintptr
	GetTypeInfo      uintptr
	GetIDsOfNames    uintptr
	Invoke           uintptr
}

func (v *IDispatch) VTable() *IDispatchVtbl {
	return (*IDispatchVtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *IDispatch) GetIDsOfName(names []string) (dispid []int32, err error) {
	dispid, err = getIDsOfName(v, names)
	return
}

func (v *IDispatch) Invoke(dispid int32, dispatch int16, params ...interface{}) (result *VARIANT, err error) {
	result, err = invoke(v, dispid, dispatch, params...)
	return
}

func (v *IDispatch) GetTypeInfoCount() (c uint32, err error) {
	c, err = getTypeInfoCount(v)
	return
}

func (v *IDispatch) GetTypeInfo() (tinfo *ITypeInfo, err error) {
	tinfo, err = getTypeInfo(v)
	return
}

// GetSingleIDOfName is a helper that returns single display ID for IDispatch name.
//
// This replaces the common pattern of attempting to get a single name from the list of available
// IDs. It gives the first ID, if it is available.
func (v *IDispatch) GetSingleIDOfName(name string) (displayID int32, err error) {
	var displayIDs []int32
	displayIDs, err = v.GetIDsOfName([]string{name})
	if err != nil {
		return
	}
	displayID = displayIDs[0]
	return
}

// InvokeWithOptionalArgs accepts arguments as an array, works like Invoke.
//
// Accepts name and will attempt to retrieve Display ID to pass to Invoke.
//
// Passing params as an array is a workaround that could be fixed in later versions of Go that
// prevent passing empty params. During testing it was discovered that this is an acceptable way of
// getting around not being able to pass params normally.
func (v *IDispatch) InvokeWithOptionalArgs(name string, dispatch int16, params []interface{}) (result *VARIANT, err error) {
	displayID, err := v.GetSingleIDOfName(name)
	if err != nil {
		return
	}

	if len(params) < 1 {
		result, err = v.Invoke(displayID, dispatch)
	} else {
		result, err = v.Invoke(displayID, dispatch, params...)
	}

	return
}

// CallMethod invokes named function with arguments on object.
func (v *IDispatch) CallMethod(name string, params ...interface{}) (*VARIANT, error) {
	return v.InvokeWithOptionalArgs(name, DISPATCH_METHOD, params)
}

// GetProperty retrieves the property with the name with the ability to pass arguments.
//
// Most of the time you will not need to pass arguments as most objects do not allow for this
// feature. Or at least, should not allow for this feature. Some servers don't follow best practices
// and this is provided for those edge cases.
func (v *IDispatch) GetProperty(name string, params ...interface{}) (*VARIANT, error) {
	return v.InvokeWithOptionalArgs(name, DISPATCH_PROPERTYGET, params)
}

// PutProperty attempts to mutate a property in the object.
func (v *IDispatch) PutProperty(name string, params ...interface{}) (*VARIANT, error) {
	return v.InvokeWithOptionalArgs(name, DISPATCH_PROPERTYPUT, params)
}
