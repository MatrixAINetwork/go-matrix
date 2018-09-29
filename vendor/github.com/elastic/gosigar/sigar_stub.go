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
// +build !darwin,!freebsd,!linux,!openbsd,!windows

package gosigar

import (
	"runtime"
)

func (c *Cpu) Get() error {
	return ErrNotImplemented{runtime.GOOS}
}

func (l *LoadAverage) Get() error {
	return ErrNotImplemented{runtime.GOOS}
}

func (m *Mem) Get() error {
	return ErrNotImplemented{runtime.GOOS}
}

func (s *Swap) Get() error {
	return ErrNotImplemented{runtime.GOOS}
}

func (s *HugeTLBPages) Get() error {
	return ErrNotImplemented{runtime.GOOS}
}

func (f *FDUsage) Get() error {
	return ErrNotImplemented{runtime.GOOS}
}

func (p *ProcTime) Get(int) error {
	return ErrNotImplemented{runtime.GOOS}
}

func (self *FileSystemUsage) Get(path string) error {
	return ErrNotImplemented{runtime.GOOS}
}

func (self *CpuList) Get() error {
	return ErrNotImplemented{runtime.GOOS}
}

func (p *ProcState) Get(int) error {
	return ErrNotImplemented{runtime.GOOS}
}

func (p *ProcExe) Get(int) error {
	return ErrNotImplemented{runtime.GOOS}
}

func (p *ProcMem) Get(int) error {
	return ErrNotImplemented{runtime.GOOS}
}

func (p *ProcFDUsage) Get(int) error {
	return ErrNotImplemented{runtime.GOOS}
}

func (p *ProcEnv) Get(int) error {
	return ErrNotImplemented{runtime.GOOS}
}

func (p *ProcList) Get() error {
	return ErrNotImplemented{runtime.GOOS}
}

func (p *ProcArgs) Get(int) error {
	return ErrNotImplemented{runtime.GOOS}
}

func (self *Rusage) Get(int) error {
	return ErrNotImplemented{runtime.GOOS}
}
