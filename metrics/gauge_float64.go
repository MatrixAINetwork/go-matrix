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

import "sync"

// GaugeFloat64s hold a float64 value that can be set arbitrarily.
type GaugeFloat64 interface {
	Snapshot() GaugeFloat64
	Update(float64)
	Value() float64
}

// GetOrRegisterGaugeFloat64 returns an existing GaugeFloat64 or constructs and registers a
// new StandardGaugeFloat64.
func GetOrRegisterGaugeFloat64(name string, r Registry) GaugeFloat64 {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, NewGaugeFloat64()).(GaugeFloat64)
}

// NewGaugeFloat64 constructs a new StandardGaugeFloat64.
func NewGaugeFloat64() GaugeFloat64 {
	if !Enabled {
		return NilGaugeFloat64{}
	}
	return &StandardGaugeFloat64{
		value: 0.0,
	}
}

// NewRegisteredGaugeFloat64 constructs and registers a new StandardGaugeFloat64.
func NewRegisteredGaugeFloat64(name string, r Registry) GaugeFloat64 {
	c := NewGaugeFloat64()
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewFunctionalGauge constructs a new FunctionalGauge.
func NewFunctionalGaugeFloat64(f func() float64) GaugeFloat64 {
	if !Enabled {
		return NilGaugeFloat64{}
	}
	return &FunctionalGaugeFloat64{value: f}
}

// NewRegisteredFunctionalGauge constructs and registers a new StandardGauge.
func NewRegisteredFunctionalGaugeFloat64(name string, r Registry, f func() float64) GaugeFloat64 {
	c := NewFunctionalGaugeFloat64(f)
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// GaugeFloat64Snapshot is a read-only copy of another GaugeFloat64.
type GaugeFloat64Snapshot float64

// Snapshot returns the snapshot.
func (g GaugeFloat64Snapshot) Snapshot() GaugeFloat64 { return g }

// Update panics.
func (GaugeFloat64Snapshot) Update(float64) {
	panic("Update called on a GaugeFloat64Snapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g GaugeFloat64Snapshot) Value() float64 { return float64(g) }

// NilGauge is a no-op Gauge.
type NilGaugeFloat64 struct{}

// Snapshot is a no-op.
func (NilGaugeFloat64) Snapshot() GaugeFloat64 { return NilGaugeFloat64{} }

// Update is a no-op.
func (NilGaugeFloat64) Update(v float64) {}

// Value is a no-op.
func (NilGaugeFloat64) Value() float64 { return 0.0 }

// StandardGaugeFloat64 is the standard implementation of a GaugeFloat64 and uses
// sync.Mutex to manage a single float64 value.
type StandardGaugeFloat64 struct {
	mutex sync.Mutex
	value float64
}

// Snapshot returns a read-only copy of the gauge.
func (g *StandardGaugeFloat64) Snapshot() GaugeFloat64 {
	return GaugeFloat64Snapshot(g.Value())
}

// Update updates the gauge's value.
func (g *StandardGaugeFloat64) Update(v float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.value = v
}

// Value returns the gauge's current value.
func (g *StandardGaugeFloat64) Value() float64 {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.value
}

// FunctionalGaugeFloat64 returns value from given function
type FunctionalGaugeFloat64 struct {
	value func() float64
}

// Value returns the gauge's current value.
func (g FunctionalGaugeFloat64) Value() float64 {
	return g.value()
}

// Snapshot returns the snapshot.
func (g FunctionalGaugeFloat64) Snapshot() GaugeFloat64 { return GaugeFloat64Snapshot(g.Value()) }

// Update panics.
func (FunctionalGaugeFloat64) Update(float64) {
	panic("Update called on a FunctionalGaugeFloat64")
}
