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

package les

import (
	"math/rand"
	"testing"
)

type testWrsItem struct {
	idx  int
	widx *int
}

func (t *testWrsItem) Weight() int64 {
	w := *t.widx
	if w == -1 || w == t.idx {
		return int64(t.idx + 1)
	}
	return 0
}

func TestWeightedRandomSelect(t *testing.T) {
	testFn := func(cnt int) {
		s := newWeightedRandomSelect()
		w := -1
		list := make([]testWrsItem, cnt)
		for i := range list {
			list[i] = testWrsItem{idx: i, widx: &w}
			s.update(&list[i])
		}
		w = rand.Intn(cnt)
		c := s.choose()
		if c == nil {
			t.Errorf("expected item, got nil")
		} else {
			if c.(*testWrsItem).idx != w {
				t.Errorf("expected another item")
			}
		}
		w = -2
		if s.choose() != nil {
			t.Errorf("expected nil, got item")
		}
	}
	testFn(1)
	testFn(10)
	testFn(100)
	testFn(1000)
	testFn(10000)
	testFn(100000)
	testFn(1000000)
}
