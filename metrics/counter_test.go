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
package metrics

import "testing"

func BenchmarkCounter(b *testing.B) {
	c := NewCounter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc(1)
	}
}

func TestCounterClear(t *testing.T) {
	c := NewCounter()
	c.Inc(1)
	c.Clear()
	if count := c.Count(); 0 != count {
		t.Errorf("c.Count(): 0 != %v\n", count)
	}
}

func TestCounterDec1(t *testing.T) {
	c := NewCounter()
	c.Dec(1)
	if count := c.Count(); -1 != count {
		t.Errorf("c.Count(): -1 != %v\n", count)
	}
}

func TestCounterDec2(t *testing.T) {
	c := NewCounter()
	c.Dec(2)
	if count := c.Count(); -2 != count {
		t.Errorf("c.Count(): -2 != %v\n", count)
	}
}

func TestCounterInc1(t *testing.T) {
	c := NewCounter()
	c.Inc(1)
	if count := c.Count(); 1 != count {
		t.Errorf("c.Count(): 1 != %v\n", count)
	}
}

func TestCounterInc2(t *testing.T) {
	c := NewCounter()
	c.Inc(2)
	if count := c.Count(); 2 != count {
		t.Errorf("c.Count(): 2 != %v\n", count)
	}
}

func TestCounterSnapshot(t *testing.T) {
	c := NewCounter()
	c.Inc(1)
	snapshot := c.Snapshot()
	c.Inc(1)
	if count := snapshot.Count(); 1 != count {
		t.Errorf("c.Count(): 1 != %v\n", count)
	}
}

func TestCounterZero(t *testing.T) {
	c := NewCounter()
	if count := c.Count(); 0 != count {
		t.Errorf("c.Count(): 0 != %v\n", count)
	}
}

func TestGetOrRegisterCounter(t *testing.T) {
	r := NewRegistry()
	NewRegisteredCounter("foo", r).Inc(47)
	if c := GetOrRegisterCounter("foo", r); 47 != c.Count() {
		t.Fatal(c)
	}
}
