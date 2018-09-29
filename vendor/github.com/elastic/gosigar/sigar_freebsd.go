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
// Copied and modified from sigar_linux.go.

package gosigar

import (
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

/*
#include <sys/param.h>
#include <sys/mount.h>
#include <sys/ucred.h>
#include <sys/types.h>
#include <sys/sysctl.h>
#include <stdlib.h>
#include <stdint.h>
#include <unistd.h>
#include <time.h>
*/
import "C"

func init() {
	system.ticks = uint64(C.sysconf(C._SC_CLK_TCK))

	Procd = "/compat/linux/proc"

	getLinuxBootTime()
}

func getMountTableFileName() string {
	return Procd + "/mtab"
}

func (self *Uptime) Get() error {
	ts := C.struct_timespec{}

	if _, err := C.clock_gettime(C.CLOCK_UPTIME, &ts); err != nil {
		return err
	}

	self.Length = float64(ts.tv_sec) + 1e-9*float64(ts.tv_nsec)

	return nil
}

func (self *FDUsage) Get() error {
	val := C.uint32_t(0)
	sc := C.size_t(4)

	name := C.CString("kern.openfiles")
	_, err := C.sysctlbyname(name, unsafe.Pointer(&val), &sc, nil, 0)
	C.free(unsafe.Pointer(name))
	if err != nil {
		return err
	}
	self.Open = uint64(val)

	name = C.CString("kern.maxfiles")
	_, err = C.sysctlbyname(name, unsafe.Pointer(&val), &sc, nil, 0)
	C.free(unsafe.Pointer(name))
	if err != nil {
		return err
	}
	self.Max = uint64(val)

	self.Unused = self.Max - self.Open

	return nil
}

func (self *ProcFDUsage) Get(pid int) error {
	err := readFile("/proc/"+strconv.Itoa(pid)+"/rlimit", func(line string) bool {
		if strings.HasPrefix(line, "nofile") {
			fields := strings.Fields(line)
			if len(fields) == 3 {
				self.SoftLimit, _ = strconv.ParseUint(fields[1], 10, 64)
				self.HardLimit, _ = strconv.ParseUint(fields[2], 10, 64)
			}
			return false
		}
		return true
	})
	if err != nil {
		return err
	}

	// linprocfs only provides this information for this process (self).
	fds, err := ioutil.ReadDir(procFileName(pid, "fd"))
	if err != nil {
		return err
	}
	self.Open = uint64(len(fds))

	return nil
}

func (self *HugeTLBPages) Get() error {
	return ErrNotImplemented{runtime.GOOS}
}

func parseCpuStat(self *Cpu, line string) error {
	fields := strings.Fields(line)

	self.User, _ = strtoull(fields[1])
	self.Nice, _ = strtoull(fields[2])
	self.Sys, _ = strtoull(fields[3])
	self.Idle, _ = strtoull(fields[4])
	return nil
}
