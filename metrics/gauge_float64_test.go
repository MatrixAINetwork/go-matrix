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

func BenchmarkGuageFloat64(b *testing.B) {
	g := NewGaugeFloat64()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(float64(i))
	}
}

func TestGaugeFloat64(t *testing.T) {
	g := NewGaugeFloat64()
	g.Update(float64(47.0))
	if v := g.Value(); float64(47.0) != v {
		t.Errorf("g.Value(): 47.0 != %v\n", v)
	}
}

func TestGaugeFloat64Snapshot(t *testing.T) {
	g := NewGaugeFloat64()
	g.Update(float64(47.0))
	snapshot := g.Snapshot()
	g.Update(float64(0))
	if v := snapshot.Value(); float64(47.0) != v {
		t.Errorf("g.Value(): 47.0 != %v\n", v)
	}
}

func TestGetOrRegisterGaugeFloat64(t *testing.T) {
	r := NewRegistry()
	NewRegisteredGaugeFloat64("foo", r).Update(float64(47.0))
	t.Logf("registry: %v", r)
	if g := GetOrRegisterGaugeFloat64("foo", r); float64(47.0) != g.Value() {
		t.Fatal(g)
	}
}

func TestFunctionalGaugeFloat64(t *testing.T) {
	var counter float64
	fg := NewFunctionalGaugeFloat64(func() float64 {
		counter++
		return counter
	})
	fg.Value()
	fg.Value()
	if counter != 2 {
		t.Error("counter != 2")
	}
}

func TestGetOrRegisterFunctionalGaugeFloat64(t *testing.T) {
	r := NewRegistry()
	NewRegisteredFunctionalGaugeFloat64("foo", r, func() float64 { return 47 })
	if g := GetOrRegisterGaugeFloat64("foo", r); 47 != g.Value() {
		t.Fatal(g)
	}
}
