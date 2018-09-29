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
package oleutil

import ole "github.com/go-ole/go-ole"

// ClassIDFrom retrieves class ID whether given is program ID or application string.
func ClassIDFrom(programID string) (classID *ole.GUID, err error) {
	return ole.ClassIDFrom(programID)
}

// CreateObject creates object from programID based on interface type.
//
// Only supports IUnknown.
//
// Program ID can be either program ID or application string.
func CreateObject(programID string) (unknown *ole.IUnknown, err error) {
	classID, err := ole.ClassIDFrom(programID)
	if err != nil {
		return
	}

	unknown, err = ole.CreateInstance(classID, ole.IID_IUnknown)
	if err != nil {
		return
	}

	return
}

// GetActiveObject retrieves active object for program ID and interface ID based
// on interface type.
//
// Only supports IUnknown.
//
// Program ID can be either program ID or application string.
func GetActiveObject(programID string) (unknown *ole.IUnknown, err error) {
	classID, err := ole.ClassIDFrom(programID)
	if err != nil {
		return
	}

	unknown, err = ole.GetActiveObject(classID, ole.IID_IUnknown)
	if err != nil {
		return
	}

	return
}

// CallMethod calls method on IDispatch with parameters.
func CallMethod(disp *ole.IDispatch, name string, params ...interface{}) (result *ole.VARIANT, err error) {
	return disp.InvokeWithOptionalArgs(name, ole.DISPATCH_METHOD, params)
}

// MustCallMethod calls method on IDispatch with parameters or panics.
func MustCallMethod(disp *ole.IDispatch, name string, params ...interface{}) (result *ole.VARIANT) {
	r, err := CallMethod(disp, name, params...)
	if err != nil {
		panic(err.Error())
	}
	return r
}

// GetProperty retrieves property from IDispatch.
func GetProperty(disp *ole.IDispatch, name string, params ...interface{}) (result *ole.VARIANT, err error) {
	return disp.InvokeWithOptionalArgs(name, ole.DISPATCH_PROPERTYGET, params)
}

// MustGetProperty retrieves property from IDispatch or panics.
func MustGetProperty(disp *ole.IDispatch, name string, params ...interface{}) (result *ole.VARIANT) {
	r, err := GetProperty(disp, name, params...)
	if err != nil {
		panic(err.Error())
	}
	return r
}

// PutProperty mutates property.
func PutProperty(disp *ole.IDispatch, name string, params ...interface{}) (result *ole.VARIANT, err error) {
	return disp.InvokeWithOptionalArgs(name, ole.DISPATCH_PROPERTYPUT, params)
}

// MustPutProperty mutates property or panics.
func MustPutProperty(disp *ole.IDispatch, name string, params ...interface{}) (result *ole.VARIANT) {
	r, err := PutProperty(disp, name, params...)
	if err != nil {
		panic(err.Error())
	}
	return r
}

// PutPropertyRef mutates property reference.
func PutPropertyRef(disp *ole.IDispatch, name string, params ...interface{}) (result *ole.VARIANT, err error) {
	return disp.InvokeWithOptionalArgs(name, ole.DISPATCH_PROPERTYPUTREF, params)
}

// MustPutPropertyRef mutates property reference or panics.
func MustPutPropertyRef(disp *ole.IDispatch, name string, params ...interface{}) (result *ole.VARIANT) {
	r, err := PutPropertyRef(disp, name, params...)
	if err != nil {
		panic(err.Error())
	}
	return r
}

func ForEach(disp *ole.IDispatch, f func(v *ole.VARIANT) error) error {
	newEnum, err := disp.GetProperty("_NewEnum")
	if err != nil {
		return err
	}
	defer newEnum.Clear()

	enum, err := newEnum.ToIUnknown().IEnumVARIANT(ole.IID_IEnumVariant)
	if err != nil {
		return err
	}
	defer enum.Release()

	for item, length, err := enum.Next(1); length > 0; item, length, err = enum.Next(1) {
		if err != nil {
			return err
		}
		if ferr := f(&item); ferr != nil {
			return ferr
		}
	}
	return nil
}
