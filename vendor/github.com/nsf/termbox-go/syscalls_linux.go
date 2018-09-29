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
// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs syscalls.go

package termbox

import "syscall"

type syscall_Termios syscall.Termios

const (
	syscall_IGNBRK = syscall.IGNBRK
	syscall_BRKINT = syscall.BRKINT
	syscall_PARMRK = syscall.PARMRK
	syscall_ISTRIP = syscall.ISTRIP
	syscall_INLCR  = syscall.INLCR
	syscall_IGNCR  = syscall.IGNCR
	syscall_ICRNL  = syscall.ICRNL
	syscall_IXON   = syscall.IXON
	syscall_OPOST  = syscall.OPOST
	syscall_ECHO   = syscall.ECHO
	syscall_ECHONL = syscall.ECHONL
	syscall_ICANON = syscall.ICANON
	syscall_ISIG   = syscall.ISIG
	syscall_IEXTEN = syscall.IEXTEN
	syscall_CSIZE  = syscall.CSIZE
	syscall_PARENB = syscall.PARENB
	syscall_CS8    = syscall.CS8
	syscall_VMIN   = syscall.VMIN
	syscall_VTIME  = syscall.VTIME

	syscall_TCGETS = syscall.TCGETS
	syscall_TCSETS = syscall.TCSETS
)
