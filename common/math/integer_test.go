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

package math

import (
	"testing"
)

type operation byte

const (
	sub operation = iota
	add
	mul
)

func TestOverflow(t *testing.T) {
	for i, test := range []struct {
		x        uint64
		y        uint64
		overflow bool
		op       operation
	}{
		// add operations
		{MaxUint64, 1, true, add},
		{MaxUint64 - 1, 1, false, add},

		// sub operations
		{0, 1, true, sub},
		{0, 0, false, sub},

		// mul operations
		{0, 0, false, mul},
		{10, 10, false, mul},
		{MaxUint64, 2, true, mul},
		{MaxUint64, 1, false, mul},
	} {
		var overflows bool
		switch test.op {
		case sub:
			_, overflows = SafeSub(test.x, test.y)
		case add:
			_, overflows = SafeAdd(test.x, test.y)
		case mul:
			_, overflows = SafeMul(test.x, test.y)
		}

		if test.overflow != overflows {
			t.Errorf("%d failed. Expected test to be %v, got %v", i, test.overflow, overflows)
		}
	}
}

func TestHexOrDecimal64(t *testing.T) {
	tests := []struct {
		input string
		num   uint64
		ok    bool
	}{
		{"", 0, true},
		{"0", 0, true},
		{"0x0", 0, true},
		{"12345678", 12345678, true},
		{"0x12345678", 0x12345678, true},
		{"0X12345678", 0x12345678, true},
		// Tests for leading zero behaviour:
		{"0123456789", 123456789, true}, // note: not octal
		{"0x00", 0, true},
		{"0x012345678abc", 0x12345678abc, true},
		// Invalid syntax:
		{"abcdef", 0, false},
		{"0xgg", 0, false},
		// Doesn't fit into 64 bits:
		{"18446744073709551617", 0, false},
	}
	for _, test := range tests {
		var num HexOrDecimal64
		err := num.UnmarshalText([]byte(test.input))
		if (err == nil) != test.ok {
			t.Errorf("ParseUint64(%q) -> (err == nil) = %t, want %t", test.input, err == nil, test.ok)
			continue
		}
		if err == nil && uint64(num) != test.num {
			t.Errorf("ParseUint64(%q) -> %d, want %d", test.input, num, test.num)
		}
	}
}

func TestMustParseUint64(t *testing.T) {
	if v := MustParseUint64("12345"); v != 12345 {
		t.Errorf(`MustParseUint64("12345") = %d, want 12345`, v)
	}
}

func TestMustParseUint64Panic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("MustParseBig should've panicked")
		}
	}()
	MustParseUint64("ggg")
}
