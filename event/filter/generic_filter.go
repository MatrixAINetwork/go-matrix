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

package filter

type Generic struct {
	Str1, Str2, Str3 string
	Data             map[string]struct{}

	Fn func(data interface{})
}

// self = registered, f = incoming
func (self Generic) Compare(f Filter) bool {
	var strMatch, dataMatch = true, true

	filter := f.(Generic)
	if (len(self.Str1) > 0 && filter.Str1 != self.Str1) ||
		(len(self.Str2) > 0 && filter.Str2 != self.Str2) ||
		(len(self.Str3) > 0 && filter.Str3 != self.Str3) {
		strMatch = false
	}

	for k := range self.Data {
		if _, ok := filter.Data[k]; !ok {
			return false
		}
	}

	return strMatch && dataMatch
}

func (self Generic) Trigger(data interface{}) {
	self.Fn(data)
}
