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

import (
	"runtime"
	"testing"
	"time"
)

func BenchmarkRuntimeMemStats(b *testing.B) {
	r := NewRegistry()
	RegisterRuntimeMemStats(r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CaptureRuntimeMemStatsOnce(r)
	}
}

func TestRuntimeMemStats(t *testing.T) {
	r := NewRegistry()
	RegisterRuntimeMemStats(r)
	CaptureRuntimeMemStatsOnce(r)
	zero := runtimeMetrics.MemStats.PauseNs.Count() // Get a "zero" since GC may have run before these tests.
	runtime.GC()
	CaptureRuntimeMemStatsOnce(r)
	if count := runtimeMetrics.MemStats.PauseNs.Count(); 1 != count-zero {
		t.Fatal(count - zero)
	}
	runtime.GC()
	runtime.GC()
	CaptureRuntimeMemStatsOnce(r)
	if count := runtimeMetrics.MemStats.PauseNs.Count(); 3 != count-zero {
		t.Fatal(count - zero)
	}
	for i := 0; i < 256; i++ {
		runtime.GC()
	}
	CaptureRuntimeMemStatsOnce(r)
	if count := runtimeMetrics.MemStats.PauseNs.Count(); 259 != count-zero {
		t.Fatal(count - zero)
	}
	for i := 0; i < 257; i++ {
		runtime.GC()
	}
	CaptureRuntimeMemStatsOnce(r)
	if count := runtimeMetrics.MemStats.PauseNs.Count(); 515 != count-zero { // We lost one because there were too many GCs between captures.
		t.Fatal(count - zero)
	}
}

func TestRuntimeMemStatsNumThread(t *testing.T) {
	r := NewRegistry()
	RegisterRuntimeMemStats(r)
	CaptureRuntimeMemStatsOnce(r)

	if value := runtimeMetrics.NumThread.Value(); value < 1 {
		t.Fatalf("got NumThread: %d, wanted at least 1", value)
	}
}

func TestRuntimeMemStatsBlocking(t *testing.T) {
	if g := runtime.GOMAXPROCS(0); g < 2 {
		t.Skipf("skipping TestRuntimeMemStatsBlocking with GOMAXPROCS=%d\n", g)
	}
	ch := make(chan int)
	go testRuntimeMemStatsBlocking(ch)
	var memStats runtime.MemStats
	t0 := time.Now()
	runtime.ReadMemStats(&memStats)
	t1 := time.Now()
	t.Log("i++ during runtime.ReadMemStats:", <-ch)
	go testRuntimeMemStatsBlocking(ch)
	d := t1.Sub(t0)
	t.Log(d)
	time.Sleep(d)
	t.Log("i++ during time.Sleep:", <-ch)
}

func testRuntimeMemStatsBlocking(ch chan int) {
	i := 0
	for {
		select {
		case ch <- i:
			return
		default:
			i++
		}
	}
}
