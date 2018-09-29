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
	"strconv"
)

func (runtime *_runtime) newArgumentsObject(indexOfParameterName []string, stash _stash, length int) *_object {
	self := runtime.newClassObject("Arguments")

	for index, _ := range indexOfParameterName {
		name := strconv.FormatInt(int64(index), 10)
		objectDefineOwnProperty(self, name, _property{Value{}, 0111}, false)
	}

	self.objectClass = _classArguments
	self.value = _argumentsObject{
		indexOfParameterName: indexOfParameterName,
		stash:                stash,
	}

	self.prototype = runtime.global.ObjectPrototype

	self.defineProperty("length", toValue_int(length), 0101, false)

	return self
}

type _argumentsObject struct {
	indexOfParameterName []string
	// function(abc, def, ghi)
	// indexOfParameterName[0] = "abc"
	// indexOfParameterName[1] = "def"
	// indexOfParameterName[2] = "ghi"
	// ...
	stash _stash
}

func (in _argumentsObject) clone(clone *_clone) _argumentsObject {
	indexOfParameterName := make([]string, len(in.indexOfParameterName))
	copy(indexOfParameterName, in.indexOfParameterName)
	return _argumentsObject{
		indexOfParameterName,
		clone.stash(in.stash),
	}
}

func (self _argumentsObject) get(name string) (Value, bool) {
	index := stringToArrayIndex(name)
	if index >= 0 && index < int64(len(self.indexOfParameterName)) {
		name := self.indexOfParameterName[index]
		if name == "" {
			return Value{}, false
		}
		return self.stash.getBinding(name, false), true
	}
	return Value{}, false
}

func (self _argumentsObject) put(name string, value Value) {
	index := stringToArrayIndex(name)
	name = self.indexOfParameterName[index]
	self.stash.setBinding(name, value, false)
}

func (self _argumentsObject) delete(name string) {
	index := stringToArrayIndex(name)
	self.indexOfParameterName[index] = ""
}

func argumentsGet(self *_object, name string) Value {
	if value, exists := self.value.(_argumentsObject).get(name); exists {
		return value
	}
	return objectGet(self, name)
}

func argumentsGetOwnProperty(self *_object, name string) *_property {
	property := objectGetOwnProperty(self, name)
	if value, exists := self.value.(_argumentsObject).get(name); exists {
		property.value = value
	}
	return property
}

func argumentsDefineOwnProperty(self *_object, name string, descriptor _property, throw bool) bool {
	if _, exists := self.value.(_argumentsObject).get(name); exists {
		if !objectDefineOwnProperty(self, name, descriptor, false) {
			return self.runtime.typeErrorResult(throw)
		}
		if value, valid := descriptor.value.(Value); valid {
			self.value.(_argumentsObject).put(name, value)
		}
		return true
	}
	return objectDefineOwnProperty(self, name, descriptor, throw)
}

func argumentsDelete(self *_object, name string, throw bool) bool {
	if !objectDelete(self, name, throw) {
		return false
	}
	if _, exists := self.value.(_argumentsObject).get(name); exists {
		self.value.(_argumentsObject).delete(name)
	}
	return true
}
