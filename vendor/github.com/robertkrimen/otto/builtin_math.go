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
	"math/rand"
)

// Math

func builtinMath_abs(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Abs(number))
}

func builtinMath_acos(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Acos(number))
}

func builtinMath_asin(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Asin(number))
}

func builtinMath_atan(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Atan(number))
}

func builtinMath_atan2(call FunctionCall) Value {
	y := call.Argument(0).float64()
	if math.IsNaN(y) {
		return NaNValue()
	}
	x := call.Argument(1).float64()
	if math.IsNaN(x) {
		return NaNValue()
	}
	return toValue_float64(math.Atan2(y, x))
}

func builtinMath_cos(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Cos(number))
}

func builtinMath_ceil(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Ceil(number))
}

func builtinMath_exp(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Exp(number))
}

func builtinMath_floor(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Floor(number))
}

func builtinMath_log(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Log(number))
}

func builtinMath_max(call FunctionCall) Value {
	switch len(call.ArgumentList) {
	case 0:
		return negativeInfinityValue()
	case 1:
		return toValue_float64(call.ArgumentList[0].float64())
	}
	result := call.ArgumentList[0].float64()
	if math.IsNaN(result) {
		return NaNValue()
	}
	for _, value := range call.ArgumentList[1:] {
		value := value.float64()
		if math.IsNaN(value) {
			return NaNValue()
		}
		result = math.Max(result, value)
	}
	return toValue_float64(result)
}

func builtinMath_min(call FunctionCall) Value {
	switch len(call.ArgumentList) {
	case 0:
		return positiveInfinityValue()
	case 1:
		return toValue_float64(call.ArgumentList[0].float64())
	}
	result := call.ArgumentList[0].float64()
	if math.IsNaN(result) {
		return NaNValue()
	}
	for _, value := range call.ArgumentList[1:] {
		value := value.float64()
		if math.IsNaN(value) {
			return NaNValue()
		}
		result = math.Min(result, value)
	}
	return toValue_float64(result)
}

func builtinMath_pow(call FunctionCall) Value {
	// TODO Make sure this works according to the specification (15.8.2.13)
	x := call.Argument(0).float64()
	y := call.Argument(1).float64()
	if math.Abs(x) == 1 && math.IsInf(y, 0) {
		return NaNValue()
	}
	return toValue_float64(math.Pow(x, y))
}

func builtinMath_random(call FunctionCall) Value {
	var v float64
	if call.runtime.random != nil {
		v = call.runtime.random()
	} else {
		v = rand.Float64()
	}
	return toValue_float64(v)
}

func builtinMath_round(call FunctionCall) Value {
	number := call.Argument(0).float64()
	value := math.Floor(number + 0.5)
	if value == 0 {
		value = math.Copysign(0, number)
	}
	return toValue_float64(value)
}

func builtinMath_sin(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Sin(number))
}

func builtinMath_sqrt(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Sqrt(number))
}

func builtinMath_tan(call FunctionCall) Value {
	number := call.Argument(0).float64()
	return toValue_float64(math.Tan(number))
}
