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
package gosigar

import (
	"time"
)

type ConcreteSigar struct{}

func (c *ConcreteSigar) CollectCpuStats(collectionInterval time.Duration) (<-chan Cpu, chan<- struct{}) {
	// samplesCh is buffered to 1 value to immediately return first CPU sample
	samplesCh := make(chan Cpu, 1)

	stopCh := make(chan struct{})

	go func() {
		var cpuUsage Cpu

		// Immediately provide non-delta value.
		// samplesCh is buffered to 1 value, so it will not block.
		cpuUsage.Get()
		samplesCh <- cpuUsage

		ticker := time.NewTicker(collectionInterval)

		for {
			select {
			case <-ticker.C:
				previousCpuUsage := cpuUsage

				cpuUsage.Get()

				select {
				case samplesCh <- cpuUsage.Delta(previousCpuUsage):
				default:
					// Include default to avoid channel blocking
				}

			case <-stopCh:
				return
			}
		}
	}()

	return samplesCh, stopCh
}

func (c *ConcreteSigar) GetLoadAverage() (LoadAverage, error) {
	l := LoadAverage{}
	err := l.Get()
	return l, err
}

func (c *ConcreteSigar) GetMem() (Mem, error) {
	m := Mem{}
	err := m.Get()
	return m, err
}

func (c *ConcreteSigar) GetSwap() (Swap, error) {
	s := Swap{}
	err := s.Get()
	return s, err
}

func (c *ConcreteSigar) GetHugeTLBPages() (HugeTLBPages, error) {
	p := HugeTLBPages{}
	err := p.Get()
	return p, err
}

func (c *ConcreteSigar) GetFileSystemUsage(path string) (FileSystemUsage, error) {
	f := FileSystemUsage{}
	err := f.Get(path)
	return f, err
}

func (c *ConcreteSigar) GetFDUsage() (FDUsage, error) {
	fd := FDUsage{}
	err := fd.Get()
	return fd, err
}

// GetRusage return the resource usage of the process
// Possible params: 0 = RUSAGE_SELF, 1 = RUSAGE_CHILDREN, 2 = RUSAGE_THREAD
func (c *ConcreteSigar) GetRusage(who int) (Rusage, error) {
	r := Rusage{}
	err := r.Get(who)
	return r, err
}
