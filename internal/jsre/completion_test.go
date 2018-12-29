// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

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
