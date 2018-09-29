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
	"math"
	"reflect"
	"unicode/utf16"
)

func (value Value) bool() bool {
	if value.kind == valueBoolean {
		return value.value.(bool)
	}
	if value.IsUndefined() {
		return false
	}
	if value.IsNull() {
		return false
	}
	switch value := value.value.(type) {
	case bool:
		return value
	case int, int8, int16, int32, int64:
		return 0 != reflect.ValueOf(value).Int()
	case uint, uint8, uint16, uint32, uint64:
		return 0 != reflect.ValueOf(value).Uint()
	case float32:
		return 0 != value
	case float64:
		if math.IsNaN(value) || value == 0 {
			return false
		}
		return true
	case string:
		return 0 != len(value)
	case []uint16:
		return 0 != len(utf16.Decode(value))
	}
	if value.IsObject() {
		return true
	}
	panic(fmt.Errorf("toBoolean(%T)", value.value))
}
