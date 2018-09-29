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

type syscall_Termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]uint8
	Ispeed uint32
	Ospeed uint32
}

const (
	syscall_IGNBRK = 0x1
	syscall_BRKINT = 0x2
	syscall_PARMRK = 0x8
	syscall_ISTRIP = 0x20
	syscall_INLCR  = 0x40
	syscall_IGNCR  = 0x80
	syscall_ICRNL  = 0x100
	syscall_IXON   = 0x200
	syscall_OPOST  = 0x1
	syscall_ECHO   = 0x8
	syscall_ECHONL = 0x10
	syscall_ICANON = 0x100
	syscall_ISIG   = 0x80
	syscall_IEXTEN = 0x400
	syscall_CSIZE  = 0x300
	syscall_PARENB = 0x1000
	syscall_CS8    = 0x300
	syscall_VMIN   = 0x10
	syscall_VTIME  = 0x11

	syscall_TCGETS = 0x402c7413
	syscall_TCSETS = 0x802c7414
)
