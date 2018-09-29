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

// RegExp

func builtinRegExp(call FunctionCall) Value {
	pattern := call.Argument(0)
	flags := call.Argument(1)
	if object := pattern._object(); object != nil {
		if object.class == "RegExp" && flags.IsUndefined() {
			return pattern
		}
	}
	return toValue_object(call.runtime.newRegExp(pattern, flags))
}

func builtinNewRegExp(self *_object, argumentList []Value) Value {
	return toValue_object(self.runtime.newRegExp(
		valueOfArrayIndex(argumentList, 0),
		valueOfArrayIndex(argumentList, 1),
	))
}

func builtinRegExp_toString(call FunctionCall) Value {
	thisObject := call.thisObject()
	source := thisObject.get("source").string()
	flags := []byte{}
	if thisObject.get("global").bool() {
		flags = append(flags, 'g')
	}
	if thisObject.get("ignoreCase").bool() {
		flags = append(flags, 'i')
	}
	if thisObject.get("multiline").bool() {
		flags = append(flags, 'm')
	}
	return toValue_string(fmt.Sprintf("/%s/%s", source, flags))
}

func builtinRegExp_exec(call FunctionCall) Value {
	thisObject := call.thisObject()
	target := call.Argument(0).string()
	match, result := execRegExp(thisObject, target)
	if !match {
		return nullValue
	}
	return toValue_object(execResultToArray(call.runtime, target, result))
}

func builtinRegExp_test(call FunctionCall) Value {
	thisObject := call.thisObject()
	target := call.Argument(0).string()
	match, _ := execRegExp(thisObject, target)
	return toValue_bool(match)
}

func builtinRegExp_compile(call FunctionCall) Value {
	// This (useless) function is deprecated, but is here to provide some
	// semblance of compatibility.
	// Caveat emptor: it may not be around for long.
	return Value{}
}
