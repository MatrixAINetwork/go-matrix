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
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"testing"
	"time"
)

const FANOUT = 128

// Stop the compiler from complaining during debugging.
var (
	_ = ioutil.Discard
	_ = log.LstdFlags
)

func BenchmarkMetrics(b *testing.B) {
	r := NewRegistry()
	c := NewRegisteredCounter("counter", r)
	g := NewRegisteredGauge("gauge", r)
	gf := NewRegisteredGaugeFloat64("gaugefloat64", r)
	h := NewRegisteredHistogram("histogram", r, NewUniformSample(100))
	m := NewRegisteredMeter("meter", r)
	t := NewRegisteredTimer("timer", r)
	RegisterDebugGCStats(r)
	RegisterRuntimeMemStats(r)
	b.ResetTimer()
	ch := make(chan bool)

	wgD := &sync.WaitGroup{}
	/*
		wgD.Add(1)
		go func() {
			defer wgD.Done()
			//log.Println("go CaptureDebugGCStats")
			for {
				select {
				case <-ch:
					//log.Println("done CaptureDebugGCStats")
					return
				default:
					CaptureDebugGCStatsOnce(r)
				}
			}
		}()
	//*/

	wgR := &sync.WaitGroup{}
	//*
	wgR.Add(1)
	go func() {
		defer wgR.Done()
		//log.Println("go CaptureRuntimeMemStats")
		for {
			select {
			case <-ch:
				//log.Println("done CaptureRuntimeMemStats")
				return
			default:
				CaptureRuntimeMemStatsOnce(r)
			}
		}
	}()
	//*/

	wgW := &sync.WaitGroup{}
	/*
		wgW.Add(1)
		go func() {
			defer wgW.Done()
			//log.Println("go Write")
			for {
				select {
				case <-ch:
					//log.Println("done Write")
					return
				default:
					WriteOnce(r, ioutil.Discard)
				}
			}
		}()
	//*/

	wg := &sync.WaitGroup{}
	wg.Add(FANOUT)
	for i := 0; i < FANOUT; i++ {
		go func(i int) {
			defer wg.Done()
			//log.Println("go", i)
			for i := 0; i < b.N; i++ {
				c.Inc(1)
				g.Update(int64(i))
				gf.Update(float64(i))
				h.Update(int64(i))
				m.Mark(1)
				t.Update(1)
			}
			//log.Println("done", i)
		}(i)
	}
	wg.Wait()
	close(ch)
	wgD.Wait()
	wgR.Wait()
	wgW.Wait()
}

func Example() {
	c := NewCounter()
	Register("money", c)
	c.Inc(17)

	// Threadsafe registration
	t := GetOrRegisterTimer("db.get.latency", nil)
	t.Time(func() { time.Sleep(10 * time.Millisecond) })
	t.Update(1)

	fmt.Println(c.Count())
	fmt.Println(t.Min())
	// Output: 17
	// 1
}
