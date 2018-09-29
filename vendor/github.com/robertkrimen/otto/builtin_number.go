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
	"math"
	"strconv"
)

// Number

func numberValueFromNumberArgumentList(argumentList []Value) Value {
	if len(argumentList) > 0 {
		return argumentList[0].numberValue()
	}
	return toValue_int(0)
}

func builtinNumber(call FunctionCall) Value {
	return numberValueFromNumberArgumentList(call.ArgumentList)
}

func builtinNewNumber(self *_object, argumentList []Value) Value {
	return toValue_object(self.runtime.newNumber(numberValueFromNumberArgumentList(argumentList)))
}

func builtinNumber_toString(call FunctionCall) Value {
	// Will throw a TypeError if ThisObject is not a Number
	value := call.thisClassObject("Number").primitiveValue()
	radix := 10
	radixArgument := call.Argument(0)
	if radixArgument.IsDefined() {
		integer := toIntegerFloat(radixArgument)
		if integer < 2 || integer > 36 {
			panic(call.runtime.panicRangeError("toString() radix must be between 2 and 36"))
		}
		radix = int(integer)
	}
	if radix == 10 {
		return toValue_string(value.string())
	}
	return toValue_string(numberToStringRadix(value, radix))
}

func builtinNumber_valueOf(call FunctionCall) Value {
	return call.thisClassObject("Number").primitiveValue()
}

func builtinNumber_toFixed(call FunctionCall) Value {
	precision := toIntegerFloat(call.Argument(0))
	if 20 < precision || 0 > precision {
		panic(call.runtime.panicRangeError("toFixed() precision must be between 0 and 20"))
	}
	if call.This.IsNaN() {
		return toValue_string("NaN")
	}
	value := call.This.float64()
	if math.Abs(value) >= 1e21 {
		return toValue_string(floatToString(value, 64))
	}
	return toValue_string(strconv.FormatFloat(call.This.float64(), 'f', int(precision), 64))
}

func builtinNumber_toExponential(call FunctionCall) Value {
	if call.This.IsNaN() {
		return toValue_string("NaN")
	}
	precision := float64(-1)
	if value := call.Argument(0); value.IsDefined() {
		precision = toIntegerFloat(value)
		if 0 > precision {
			panic(call.runtime.panicRangeError("toString() radix must be between 2 and 36"))
		}
	}
	return toValue_string(strconv.FormatFloat(call.This.float64(), 'e', int(precision), 64))
}

func builtinNumber_toPrecision(call FunctionCall) Value {
	if call.This.IsNaN() {
		return toValue_string("NaN")
	}
	value := call.Argument(0)
	if value.IsUndefined() {
		return toValue_string(call.This.string())
	}
	precision := toIntegerFloat(value)
	if 1 > precision {
		panic(call.runtime.panicRangeError("toPrecision() precision must be greater than 1"))
	}
	return toValue_string(strconv.FormatFloat(call.This.float64(), 'g', int(precision), 64))
}

func builtinNumber_toLocaleString(call FunctionCall) Value {
	return builtinNumber_toString(call)
}
