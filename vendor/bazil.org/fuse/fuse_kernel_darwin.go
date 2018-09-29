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
package fuse

import (
	"time"
)

type attr struct {
	Ino        uint64
	Size       uint64
	Blocks     uint64
	Atime      uint64
	Mtime      uint64
	Ctime      uint64
	Crtime_    uint64 // OS X only
	AtimeNsec  uint32
	MtimeNsec  uint32
	CtimeNsec  uint32
	CrtimeNsec uint32 // OS X only
	Mode       uint32
	Nlink      uint32
	Uid        uint32
	Gid        uint32
	Rdev       uint32
	Flags_     uint32 // OS X only; see chflags(2)
	Blksize    uint32
	padding    uint32
}

func (a *attr) SetCrtime(s uint64, ns uint32) {
	a.Crtime_, a.CrtimeNsec = s, ns
}

func (a *attr) SetFlags(f uint32) {
	a.Flags_ = f
}

type setattrIn struct {
	setattrInCommon

	// OS X only
	Bkuptime_    uint64
	Chgtime_     uint64
	Crtime       uint64
	BkuptimeNsec uint32
	ChgtimeNsec  uint32
	CrtimeNsec   uint32
	Flags_       uint32 // see chflags(2)
}

func (in *setattrIn) BkupTime() time.Time {
	return time.Unix(int64(in.Bkuptime_), int64(in.BkuptimeNsec))
}

func (in *setattrIn) Chgtime() time.Time {
	return time.Unix(int64(in.Chgtime_), int64(in.ChgtimeNsec))
}

func (in *setattrIn) Flags() uint32 {
	return in.Flags_
}

func openFlags(flags uint32) OpenFlags {
	return OpenFlags(flags)
}

type getxattrIn struct {
	getxattrInCommon

	// OS X only
	Position uint32
	Padding  uint32
}

func (g *getxattrIn) position() uint32 {
	return g.Position
}

type setxattrIn struct {
	setxattrInCommon

	// OS X only
	Position uint32
	Padding  uint32
}

func (s *setxattrIn) position() uint32 {
	return s.Position
}
