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

package asm

import (
	"reflect"
	"testing"
)

func lexAll(src string) []token {
	ch := Lex("test.asm", []byte(src), false)

	var tokens []token
	for i := range ch {
		tokens = append(tokens, i)
	}
	return tokens
}

func TestLexer(t *testing.T) {
	tests := []struct {
		input  string
		tokens []token
	}{
		{
			input:  ";; this is a comment",
			tokens: []token{{typ: lineStart}, {typ: eof}},
		},
		{
			input:  "0x12345678",
			tokens: []token{{typ: lineStart}, {typ: number, text: "0x12345678"}, {typ: eof}},
		},
		{
			input:  "0x123ggg",
			tokens: []token{{typ: lineStart}, {typ: number, text: "0x123"}, {typ: element, text: "ggg"}, {typ: eof}},
		},
		{
			input:  "12345678",
			tokens: []token{{typ: lineStart}, {typ: number, text: "12345678"}, {typ: eof}},
		},
		{
			input:  "123abc",
			tokens: []token{{typ: lineStart}, {typ: number, text: "123"}, {typ: element, text: "abc"}, {typ: eof}},
		},
		{
			input:  "0123abc",
			tokens: []token{{typ: lineStart}, {typ: number, text: "0123"}, {typ: element, text: "abc"}, {typ: eof}},
		},
	}

	for _, test := range tests {
		tokens := lexAll(test.input)
		if !reflect.DeepEqual(tokens, test.tokens) {
			t.Errorf("input %q\ngot:  %+v\nwant: %+v", test.input, tokens, test.tokens)
		}
	}
}
