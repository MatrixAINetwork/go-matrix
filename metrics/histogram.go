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

// Histograms calculate distribution statistics from a series of int64 values.
type Histogram interface {
	Clear()
	Count() int64
	Max() int64
	Mean() float64
	Min() int64
	Percentile(float64) float64
	Percentiles([]float64) []float64
	Sample() Sample
	Snapshot() Histogram
	StdDev() float64
	Sum() int64
	Update(int64)
	Variance() float64
}

// GetOrRegisterHistogram returns an existing Histogram or constructs and
// registers a new StandardHistogram.
func GetOrRegisterHistogram(name string, r Registry, s Sample) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() Histogram { return NewHistogram(s) }).(Histogram)
}

// NewHistogram constructs a new StandardHistogram from a Sample.
func NewHistogram(s Sample) Histogram {
	if !Enabled {
		return NilHistogram{}
	}
	return &StandardHistogram{sample: s}
}

// NewRegisteredHistogram constructs and registers a new StandardHistogram from
// a Sample.
func NewRegisteredHistogram(name string, r Registry, s Sample) Histogram {
	c := NewHistogram(s)
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// HistogramSnapshot is a read-only copy of another Histogram.
type HistogramSnapshot struct {
	sample *SampleSnapshot
}

// Clear panics.
func (*HistogramSnapshot) Clear() {
	panic("Clear called on a HistogramSnapshot")
}

// Count returns the number of samples recorded at the time the snapshot was
// taken.
func (h *HistogramSnapshot) Count() int64 { return h.sample.Count() }

// Max returns the maximum value in the sample at the time the snapshot was
// taken.
func (h *HistogramSnapshot) Max() int64 { return h.sample.Max() }

// Mean returns the mean of the values in the sample at the time the snapshot
// was taken.
func (h *HistogramSnapshot) Mean() float64 { return h.sample.Mean() }

// Min returns the minimum value in the sample at the time the snapshot was
// taken.
func (h *HistogramSnapshot) Min() int64 { return h.sample.Min() }

// Percentile returns an arbitrary percentile of values in the sample at the
// time the snapshot was taken.
func (h *HistogramSnapshot) Percentile(p float64) float64 {
	return h.sample.Percentile(p)
}

// Percentiles returns a slice of arbitrary percentiles of values in the sample
// at the time the snapshot was taken.
func (h *HistogramSnapshot) Percentiles(ps []float64) []float64 {
	return h.sample.Percentiles(ps)
}

// Sample returns the Sample underlying the histogram.
func (h *HistogramSnapshot) Sample() Sample { return h.sample }

// Snapshot returns the snapshot.
func (h *HistogramSnapshot) Snapshot() Histogram { return h }

// StdDev returns the standard deviation of the values in the sample at the
// time the snapshot was taken.
func (h *HistogramSnapshot) StdDev() float64 { return h.sample.StdDev() }

// Sum returns the sum in the sample at the time the snapshot was taken.
func (h *HistogramSnapshot) Sum() int64 { return h.sample.Sum() }

// Update panics.
func (*HistogramSnapshot) Update(int64) {
	panic("Update called on a HistogramSnapshot")
}

// Variance returns the variance of inputs at the time the snapshot was taken.
func (h *HistogramSnapshot) Variance() float64 { return h.sample.Variance() }

// NilHistogram is a no-op Histogram.
type NilHistogram struct{}

// Clear is a no-op.
func (NilHistogram) Clear() {}

// Count is a no-op.
func (NilHistogram) Count() int64 { return 0 }

// Max is a no-op.
func (NilHistogram) Max() int64 { return 0 }

// Mean is a no-op.
func (NilHistogram) Mean() float64 { return 0.0 }

// Min is a no-op.
func (NilHistogram) Min() int64 { return 0 }

// Percentile is a no-op.
func (NilHistogram) Percentile(p float64) float64 { return 0.0 }

// Percentiles is a no-op.
func (NilHistogram) Percentiles(ps []float64) []float64 {
	return make([]float64, len(ps))
}

// Sample is a no-op.
func (NilHistogram) Sample() Sample { return NilSample{} }

// Snapshot is a no-op.
func (NilHistogram) Snapshot() Histogram { return NilHistogram{} }

// StdDev is a no-op.
func (NilHistogram) StdDev() float64 { return 0.0 }

// Sum is a no-op.
func (NilHistogram) Sum() int64 { return 0 }

// Update is a no-op.
func (NilHistogram) Update(v int64) {}

// Variance is a no-op.
func (NilHistogram) Variance() float64 { return 0.0 }

// StandardHistogram is the standard implementation of a Histogram and uses a
// Sample to bound its memory use.
type StandardHistogram struct {
	sample Sample
}

// Clear clears the histogram and its sample.
func (h *StandardHistogram) Clear() { h.sample.Clear() }

// Count returns the number of samples recorded since the histogram was last
// cleared.
func (h *StandardHistogram) Count() int64 { return h.sample.Count() }

// Max returns the maximum value in the sample.
func (h *StandardHistogram) Max() int64 { return h.sample.Max() }

// Mean returns the mean of the values in the sample.
func (h *StandardHistogram) Mean() float64 { return h.sample.Mean() }

// Min returns the minimum value in the sample.
func (h *StandardHistogram) Min() int64 { return h.sample.Min() }

// Percentile returns an arbitrary percentile of the values in the sample.
func (h *StandardHistogram) Percentile(p float64) float64 {
	return h.sample.Percentile(p)
}

// Percentiles returns a slice of arbitrary percentiles of the values in the
// sample.
func (h *StandardHistogram) Percentiles(ps []float64) []float64 {
	return h.sample.Percentiles(ps)
}

// Sample returns the Sample underlying the histogram.
func (h *StandardHistogram) Sample() Sample { return h.sample }

// Snapshot returns a read-only copy of the histogram.
func (h *StandardHistogram) Snapshot() Histogram {
	return &HistogramSnapshot{sample: h.sample.Snapshot().(*SampleSnapshot)}
}

// StdDev returns the standard deviation of the values in the sample.
func (h *StandardHistogram) StdDev() float64 { return h.sample.StdDev() }

// Sum returns the sum in the sample.
func (h *StandardHistogram) Sum() int64 { return h.sample.Sum() }

// Update samples a new value.
func (h *StandardHistogram) Update(v int64) { h.sample.Update(v) }

// Variance returns the variance of the values in the sample.
func (h *StandardHistogram) Variance() float64 { return h.sample.Variance() }
