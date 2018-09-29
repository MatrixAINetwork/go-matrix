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

package jsre

import (
	"os"
	"reflect"
	"testing"
)

func TestCompleteKeywords(t *testing.T) {
	re := New("", os.Stdout)
	re.Run(`
		function theClass() {
			this.foo = 3;
			this.gazonk = {xyz: 4};
		}
		theClass.prototype.someMethod = function () {};
  		var x = new theClass();
  		var y = new theClass();
		y.someMethod = function override() {};
	`)

	var tests = []struct {
		input string
		want  []string
	}{
		{
			input: "x",
			want:  []string{"x."},
		},
		{
			input: "x.someMethod",
			want:  []string{"x.someMethod("},
		},
		{
			input: "x.",
			want: []string{
				"x.constructor",
				"x.foo",
				"x.gazonk",
				"x.someMethod",
			},
		},
		{
			input: "y.",
			want: []string{
				"y.constructor",
				"y.foo",
				"y.gazonk",
				"y.someMethod",
			},
		},
		{
			input: "x.gazonk.",
			want: []string{
				"x.gazonk.constructor",
				"x.gazonk.hasOwnProperty",
				"x.gazonk.isPrototypeOf",
				"x.gazonk.propertyIsEnumerable",
				"x.gazonk.toLocaleString",
				"x.gazonk.toString",
				"x.gazonk.valueOf",
				"x.gazonk.xyz",
			},
		},
	}
	for _, test := range tests {
		cs := re.CompleteKeywords(test.input)
		if !reflect.DeepEqual(cs, test.want) {
			t.Errorf("wrong completions for %q\ngot  %v\nwant %v", test.input, cs, test.want)
		}
	}
}
