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
package otto

import (
	"reflect"
	"strconv"
)

func (runtime *_runtime) newGoArrayObject(value reflect.Value) *_object {
	self := runtime.newObject()
	self.class = "GoArray"
	self.objectClass = _classGoArray
	self.value = _newGoArrayObject(value)
	return self
}

type _goArrayObject struct {
	value        reflect.Value
	writable     bool
	propertyMode _propertyMode
}

func _newGoArrayObject(value reflect.Value) *_goArrayObject {
	writable := value.Kind() == reflect.Ptr // The Array is addressable (like a Slice)
	mode := _propertyMode(0010)
	if writable {
		mode = 0110
	}
	self := &_goArrayObject{
		value:        value,
		writable:     writable,
		propertyMode: mode,
	}
	return self
}

func (self _goArrayObject) getValue(index int64) (reflect.Value, bool) {
	value := reflect.Indirect(self.value)
	if index < int64(value.Len()) {
		return value.Index(int(index)), true
	}
	return reflect.Value{}, false
}

func (self _goArrayObject) setValue(index int64, value Value) bool {
	indexValue, exists := self.getValue(index)
	if !exists {
		return false
	}
	reflectValue, err := value.toReflectValue(reflect.Indirect(self.value).Type().Elem().Kind())
	if err != nil {
		panic(err)
	}
	indexValue.Set(reflectValue)
	return true
}

func goArrayGetOwnProperty(self *_object, name string) *_property {
	// length
	if name == "length" {
		return &_property{
			value: toValue(reflect.Indirect(self.value.(*_goArrayObject).value).Len()),
			mode:  0,
		}
	}

	// .0, .1, .2, ...
	index := stringToArrayIndex(name)
	if index >= 0 {
		object := self.value.(*_goArrayObject)
		value := Value{}
		reflectValue, exists := object.getValue(index)
		if exists {
			value = self.runtime.toValue(reflectValue.Interface())
		}
		return &_property{
			value: value,
			mode:  object.propertyMode,
		}
	}

	return objectGetOwnProperty(self, name)
}

func goArrayEnumerate(self *_object, all bool, each func(string) bool) {
	object := self.value.(*_goArrayObject)
	// .0, .1, .2, ...

	for index, length := 0, object.value.Len(); index < length; index++ {
		name := strconv.FormatInt(int64(index), 10)
		if !each(name) {
			return
		}
	}

	objectEnumerate(self, all, each)
}

func goArrayDefineOwnProperty(self *_object, name string, descriptor _property, throw bool) bool {
	if name == "length" {
		return self.runtime.typeErrorResult(throw)
	} else if index := stringToArrayIndex(name); index >= 0 {
		object := self.value.(*_goArrayObject)
		if object.writable {
			if self.value.(*_goArrayObject).setValue(index, descriptor.value.(Value)) {
				return true
			}
		}
		return self.runtime.typeErrorResult(throw)
	}
	return objectDefineOwnProperty(self, name, descriptor, throw)
}

func goArrayDelete(self *_object, name string, throw bool) bool {
	// length
	if name == "length" {
		return self.runtime.typeErrorResult(throw)
	}

	// .0, .1, .2, ...
	index := stringToArrayIndex(name)
	if index >= 0 {
		object := self.value.(*_goArrayObject)
		if object.writable {
			indexValue, exists := object.getValue(index)
			if exists {
				indexValue.Set(reflect.Zero(reflect.Indirect(object.value).Type().Elem()))
				return true
			}
		}
		return self.runtime.typeErrorResult(throw)
	}

	return self.delete(name, throw)
}
