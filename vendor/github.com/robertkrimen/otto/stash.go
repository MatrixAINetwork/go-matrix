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
	"fmt"
)

// ======
// _stash
// ======

type _stash interface {
	hasBinding(string) bool            //
	createBinding(string, bool, Value) // CreateMutableBinding
	setBinding(string, Value, bool)    // SetMutableBinding
	getBinding(string, bool) Value     // GetBindingValue
	deleteBinding(string) bool         //
	setValue(string, Value, bool)      // createBinding + setBinding

	outer() _stash
	runtime() *_runtime

	newReference(string, bool, _at) _reference

	clone(clone *_clone) _stash
}

// ==========
// _objectStash
// ==========

type _objectStash struct {
	_runtime *_runtime
	_outer   _stash
	object   *_object
}

func (self *_objectStash) runtime() *_runtime {
	return self._runtime
}

func (runtime *_runtime) newObjectStash(object *_object, outer _stash) *_objectStash {
	if object == nil {
		object = runtime.newBaseObject()
		object.class = "environment"
	}
	return &_objectStash{
		_runtime: runtime,
		_outer:   outer,
		object:   object,
	}
}

func (in *_objectStash) clone(clone *_clone) _stash {
	out, exists := clone.objectStash(in)
	if exists {
		return out
	}
	*out = _objectStash{
		clone.runtime,
		clone.stash(in._outer),
		clone.object(in.object),
	}
	return out
}

func (self *_objectStash) hasBinding(name string) bool {
	return self.object.hasProperty(name)
}

func (self *_objectStash) createBinding(name string, deletable bool, value Value) {
	if self.object.hasProperty(name) {
		panic(hereBeDragons())
	}
	mode := _propertyMode(0111)
	if !deletable {
		mode = _propertyMode(0110)
	}
	// TODO False?
	self.object.defineProperty(name, value, mode, false)
}

func (self *_objectStash) setBinding(name string, value Value, strict bool) {
	self.object.put(name, value, strict)
}

func (self *_objectStash) setValue(name string, value Value, throw bool) {
	if !self.hasBinding(name) {
		self.createBinding(name, true, value) // Configurable by default
	} else {
		self.setBinding(name, value, throw)
	}
}

func (self *_objectStash) getBinding(name string, throw bool) Value {
	if self.object.hasProperty(name) {
		return self.object.get(name)
	}
	if throw { // strict?
		panic(self._runtime.panicReferenceError("Not Defined", name))
	}
	return Value{}
}

func (self *_objectStash) deleteBinding(name string) bool {
	return self.object.delete(name, false)
}

func (self *_objectStash) outer() _stash {
	return self._outer
}

func (self *_objectStash) newReference(name string, strict bool, at _at) _reference {
	return newPropertyReference(self._runtime, self.object, name, strict, at)
}

// =========
// _dclStash
// =========

type _dclStash struct {
	_runtime *_runtime
	_outer   _stash
	property map[string]_dclProperty
}

type _dclProperty struct {
	value     Value
	mutable   bool
	deletable bool
	readable  bool
}

func (runtime *_runtime) newDeclarationStash(outer _stash) *_dclStash {
	return &_dclStash{
		_runtime: runtime,
		_outer:   outer,
		property: map[string]_dclProperty{},
	}
}

func (in *_dclStash) clone(clone *_clone) _stash {
	out, exists := clone.dclStash(in)
	if exists {
		return out
	}
	property := make(map[string]_dclProperty, len(in.property))
	for index, value := range in.property {
		property[index] = clone.dclProperty(value)
	}
	*out = _dclStash{
		clone.runtime,
		clone.stash(in._outer),
		property,
	}
	return out
}

func (self *_dclStash) hasBinding(name string) bool {
	_, exists := self.property[name]
	return exists
}

func (self *_dclStash) runtime() *_runtime {
	return self._runtime
}

func (self *_dclStash) createBinding(name string, deletable bool, value Value) {
	_, exists := self.property[name]
	if exists {
		panic(fmt.Errorf("createBinding: %s: already exists", name))
	}
	self.property[name] = _dclProperty{
		value:     value,
		mutable:   true,
		deletable: deletable,
		readable:  false,
	}
}

func (self *_dclStash) setBinding(name string, value Value, strict bool) {
	property, exists := self.property[name]
	if !exists {
		panic(fmt.Errorf("setBinding: %s: missing", name))
	}
	if property.mutable {
		property.value = value
		self.property[name] = property
	} else {
		self._runtime.typeErrorResult(strict)
	}
}

func (self *_dclStash) setValue(name string, value Value, throw bool) {
	if !self.hasBinding(name) {
		self.createBinding(name, false, value) // NOT deletable by default
	} else {
		self.setBinding(name, value, throw)
	}
}

// FIXME This is called a __lot__
func (self *_dclStash) getBinding(name string, throw bool) Value {
	property, exists := self.property[name]
	if !exists {
		panic(fmt.Errorf("getBinding: %s: missing", name))
	}
	if !property.mutable && !property.readable {
		if throw { // strict?
			panic(self._runtime.panicTypeError())
		}
		return Value{}
	}
	return property.value
}

func (self *_dclStash) deleteBinding(name string) bool {
	property, exists := self.property[name]
	if !exists {
		return true
	}
	if !property.deletable {
		return false
	}
	delete(self.property, name)
	return true
}

func (self *_dclStash) outer() _stash {
	return self._outer
}

func (self *_dclStash) newReference(name string, strict bool, _ _at) _reference {
	return &_stashReference{
		name: name,
		base: self,
	}
}

// ========
// _fnStash
// ========

type _fnStash struct {
	_dclStash
	arguments           *_object
	indexOfArgumentName map[string]string
}

func (runtime *_runtime) newFunctionStash(outer _stash) *_fnStash {
	return &_fnStash{
		_dclStash: _dclStash{
			_runtime: runtime,
			_outer:   outer,
			property: map[string]_dclProperty{},
		},
	}
}

func (in *_fnStash) clone(clone *_clone) _stash {
	out, exists := clone.fnStash(in)
	if exists {
		return out
	}
	dclStash := in._dclStash.clone(clone).(*_dclStash)
	index := make(map[string]string, len(in.indexOfArgumentName))
	for name, value := range in.indexOfArgumentName {
		index[name] = value
	}
	*out = _fnStash{
		_dclStash:           *dclStash,
		arguments:           clone.object(in.arguments),
		indexOfArgumentName: index,
	}
	return out
}

func getStashProperties(stash _stash) (keys []string) {
	switch vars := stash.(type) {
	case *_dclStash:
		for k := range vars.property {
			keys = append(keys, k)
		}
	case *_fnStash:
		for k := range vars.property {
			keys = append(keys, k)
		}
	case *_objectStash:
		for k := range vars.object.property {
			keys = append(keys, k)
		}
	default:
		panic("unknown stash type")
	}

	return
}
