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
// Copyright (c) 2014-2015 The Notify Authors. All rights reserved.
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

// +build darwin,kqueue dragonfly freebsd netbsd openbsd

package notify

import "syscall"

// TODO(pblaszczyk): ensure in runtime notify built-in event values do not
// overlap with platform-defined ones.

// Platform independent event values.
const (
	osSpecificCreate Event = 0x0100 << iota
	osSpecificRemove
	osSpecificWrite
	osSpecificRename
	// internal
	// recursive is used to distinguish recursive eventsets from non-recursive ones
	recursive
	// omit is used for dispatching internal events; only those events are sent
	// for which both the event and the watchpoint has omit in theirs event sets.
	omit
)

const (
	// NoteDelete is an event reported when the unlink() system call was called
	// on the file referenced by the descriptor.
	NoteDelete = Event(syscall.NOTE_DELETE)
	// NoteWrite is an event reported when a write occurred on the file
	// referenced by the descriptor.
	NoteWrite = Event(syscall.NOTE_WRITE)
	// NoteExtend is an event reported when the file referenced by the
	// descriptor was extended.
	NoteExtend = Event(syscall.NOTE_EXTEND)
	// NoteAttrib is an event reported when the file referenced
	// by the descriptor had its attributes changed.
	NoteAttrib = Event(syscall.NOTE_ATTRIB)
	// NoteLink is an event reported when the link count on the file changed.
	NoteLink = Event(syscall.NOTE_LINK)
	// NoteRename is an event reported when the file referenced
	// by the descriptor was renamed.
	NoteRename = Event(syscall.NOTE_RENAME)
	// NoteRevoke is an event reported when access to the file was revoked via
	// revoke(2) or	the underlying file system was unmounted.
	NoteRevoke = Event(syscall.NOTE_REVOKE)
)

var osestr = map[Event]string{
	NoteDelete: "notify.NoteDelete",
	NoteWrite:  "notify.NoteWrite",
	NoteExtend: "notify.NoteExtend",
	NoteAttrib: "notify.NoteAttrib",
	NoteLink:   "notify.NoteLink",
	NoteRename: "notify.NoteRename",
	NoteRevoke: "notify.NoteRevoke",
}
