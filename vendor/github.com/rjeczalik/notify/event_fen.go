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

// +build solaris

package notify

const (
	osSpecificCreate Event = 0x00000100 << iota
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
	// FileAccess is an event reported when monitored file/directory was accessed.
	FileAccess = fileAccess
	// FileModified is an event reported when monitored file/directory was modified.
	FileModified = fileModified
	// FileAttrib is an event reported when monitored file/directory's ATTRIB
	// was changed.
	FileAttrib = fileAttrib
	// FileDelete is an event reported when monitored file/directory was deleted.
	FileDelete = fileDelete
	// FileRenameTo to is an event reported when monitored file/directory was renamed.
	FileRenameTo = fileRenameTo
	// FileRenameFrom is an event reported when monitored file/directory was renamed.
	FileRenameFrom = fileRenameFrom
	// FileTrunc is an event reported when monitored file/directory was truncated.
	FileTrunc = fileTrunc
	// FileNoFollow is an flag to indicate not to follow symbolic links.
	FileNoFollow = fileNoFollow
	// Unmounted is an event reported when monitored filesystem was unmounted.
	Unmounted = unmounted
	// MountedOver is an event reported when monitored file/directory was mounted on.
	MountedOver = mountedOver
)

var osestr = map[Event]string{
	FileAccess:     "notify.FileAccess",
	FileModified:   "notify.FileModified",
	FileAttrib:     "notify.FileAttrib",
	FileDelete:     "notify.FileDelete",
	FileRenameTo:   "notify.FileRenameTo",
	FileRenameFrom: "notify.FileRenameFrom",
	FileTrunc:      "notify.FileTrunc",
	FileNoFollow:   "notify.FileNoFollow",
	Unmounted:      "notify.Unmounted",
	MountedOver:    "notify.MountedOver",
}
