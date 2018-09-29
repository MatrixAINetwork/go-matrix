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
// Copyright (c) 2012 VMware, Inc.

// +build darwin freebsd linux

package gosigar

import (
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func (self *FileSystemUsage) Get(path string) error {
	stat := syscall.Statfs_t{}
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return err
	}

	self.Total = uint64(stat.Blocks) * uint64(stat.Bsize)
	self.Free = uint64(stat.Bfree) * uint64(stat.Bsize)
	self.Avail = uint64(stat.Bavail) * uint64(stat.Bsize)
	self.Used = self.Total - self.Free
	self.Files = stat.Files
	self.FreeFiles = uint64(stat.Ffree)

	return nil
}

func (r *Rusage) Get(who int) error {
	ru, err := getResourceUsage(who)
	if err != nil {
		return err
	}

	uTime := convertRtimeToDur(ru.Utime)
	sTime := convertRtimeToDur(ru.Stime)

	r.Utime = uTime
	r.Stime = sTime
	r.Maxrss = int64(ru.Maxrss)
	r.Ixrss = int64(ru.Ixrss)
	r.Idrss = int64(ru.Idrss)
	r.Isrss = int64(ru.Isrss)
	r.Minflt = int64(ru.Minflt)
	r.Majflt = int64(ru.Majflt)
	r.Nswap = int64(ru.Nswap)
	r.Inblock = int64(ru.Inblock)
	r.Oublock = int64(ru.Oublock)
	r.Msgsnd = int64(ru.Msgsnd)
	r.Msgrcv = int64(ru.Msgrcv)
	r.Nsignals = int64(ru.Nsignals)
	r.Nvcsw = int64(ru.Nvcsw)
	r.Nivcsw = int64(ru.Nivcsw)

	return nil
}

func getResourceUsage(who int) (unix.Rusage, error) {
	r := unix.Rusage{}
	err := unix.Getrusage(who, &r)

	return r, err
}

func convertRtimeToDur(t unix.Timeval) time.Duration {
	return time.Duration(t.Nano())
}
