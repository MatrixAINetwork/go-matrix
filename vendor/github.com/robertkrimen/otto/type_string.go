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
	"unicode/utf8"
)

type _stringObject interface {
	Length() int
	At(int) rune
	String() string
}

type _stringASCII string

func (str _stringASCII) Length() int {
	return len(str)
}

func (str _stringASCII) At(at int) rune {
	return rune(str[at])
}

func (str _stringASCII) String() string {
	return string(str)
}

type _stringWide struct {
	string string
	length int
	runes  []rune
}

func (str _stringWide) Length() int {
	return str.length
}

func (str _stringWide) At(at int) rune {
	if str.runes == nil {
		str.runes = []rune(str.string)
	}
	return str.runes[at]
}

func (str _stringWide) String() string {
	return str.string
}

func _newStringObject(str string) _stringObject {
	for i := 0; i < len(str); i++ {
		if str[i] >= utf8.RuneSelf {
			goto wide
		}
	}

	return _stringASCII(str)

wide:
	return &_stringWide{
		string: str,
		length: utf8.RuneCountInString(str),
	}
}

func stringAt(str _stringObject, index int) rune {
	if 0 <= index && index < str.Length() {
		return str.At(index)
	}
	return utf8.RuneError
}

func (runtime *_runtime) newStringObject(value Value) *_object {
	str := _newStringObject(value.string())

	self := runtime.newClassObject("String")
	self.defineProperty("length", toValue_int(str.Length()), 0, false)
	self.objectClass = _classString
	self.value = str
	return self
}

func (self *_object) stringValue() _stringObject {
	if str, ok := self.value.(_stringObject); ok {
		return str
	}
	return nil
}

func stringEnumerate(self *_object, all bool, each func(string) bool) {
	if str := self.stringValue(); str != nil {
		length := str.Length()
		for index := 0; index < length; index++ {
			if !each(strconv.FormatInt(int64(index), 10)) {
				return
			}
		}
	}
	objectEnumerate(self, all, each)
}

func stringGetOwnProperty(self *_object, name string) *_property {
	if property := objectGetOwnProperty(self, name); property != nil {
		return property
	}
	// TODO Test a string of length >= +int32 + 1?
	if index := stringToArrayIndex(name); index >= 0 {
		if chr := stringAt(self.stringValue(), int(index)); chr != utf8.RuneError {
			return &_property{toValue_string(string(chr)), 0}
		}
	}
	return nil
}
