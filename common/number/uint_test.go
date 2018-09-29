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

package number

import (
	"math/big"
	"testing"

	"github.com/matrix/go-matrix/common"
)

func TestSet(t *testing.T) {
	a := Uint(0)
	b := Uint(10)
	a.Set(b)
	if a.num.Cmp(b.num) != 0 {
		t.Error("didn't compare", a, b)
	}

	c := Uint(0).SetBytes(common.Hex2Bytes("0a"))
	if c.num.Cmp(big.NewInt(10)) != 0 {
		t.Error("c set bytes failed.")
	}
}

func TestInitialiser(t *testing.T) {
	check := false
	init := NewInitialiser(func(x *Number) *Number {
		check = true
		return x
	})
	a := init(0).Add(init(1), init(2))
	if a.Cmp(init(3)) != 0 {
		t.Error("expected 3. got", a)
	}
	if !check {
		t.Error("expected limiter to be called")
	}
}

func TestGet(t *testing.T) {
	a := Uint(10)
	if a.Uint64() != 10 {
		t.Error("expected to get 10. got", a.Uint64())
	}

	a = Uint(10)
	if a.Int64() != 10 {
		t.Error("expected to get 10. got", a.Int64())
	}
}

func TestCmp(t *testing.T) {
	a := Uint(10)
	b := Uint(10)
	c := Uint(11)

	if a.Cmp(b) != 0 {
		t.Error("a b == 0 failed", a, b)
	}

	if a.Cmp(c) >= 0 {
		t.Error("a c < 0 failed", a, c)
	}

	if c.Cmp(b) <= 0 {
		t.Error("c b > 0 failed", c, b)
	}
}

func TestMaxArith(t *testing.T) {
	a := Uint(0).Add(MaxUint256, One)
	if a.Cmp(Zero) != 0 {
		t.Error("expected max256 + 1 = 0 got", a)
	}

	a = Uint(0).Sub(Uint(0), One)
	if a.Cmp(MaxUint256) != 0 {
		t.Error("expected 0 - 1 = max256 got", a)
	}

	a = Int(0).Sub(Int(0), One)
	if a.Cmp(MinOne) != 0 {
		t.Error("expected 0 - 1 = -1 got", a)
	}
}

func TestConversion(t *testing.T) {
	a := Int(-1)
	b := a.Uint256()
	if b.Cmp(MaxUint256) != 0 {
		t.Error("expected -1 => unsigned to return max. got", b)
	}
}
