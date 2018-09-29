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

// +build windows

package notify

import (
	"os"
	"path/filepath"
	"syscall"
)

// Platform independent event values.
const (
	osSpecificCreate Event = 1 << (20 + iota)
	osSpecificRemove
	osSpecificWrite
	osSpecificRename
	// recursive is used to distinguish recursive eventsets from non-recursive ones
	recursive
	// omit is used for dispatching internal events; only those events are sent
	// for which both the event and the watchpoint has omit in theirs event sets.
	omit
	// dirmarker TODO(pknap)
	dirmarker
)

// ReadDirectoryChangesW filters
// On Windows the following events can be passed to Watch. A different set of
// events (see actions below) are received on the channel passed to Watch.
// For more information refer to
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa365465(v=vs.85).aspx
const (
	FileNotifyChangeFileName   = Event(syscall.FILE_NOTIFY_CHANGE_FILE_NAME)
	FileNotifyChangeDirName    = Event(syscall.FILE_NOTIFY_CHANGE_DIR_NAME)
	FileNotifyChangeAttributes = Event(syscall.FILE_NOTIFY_CHANGE_ATTRIBUTES)
	FileNotifyChangeSize       = Event(syscall.FILE_NOTIFY_CHANGE_SIZE)
	FileNotifyChangeLastWrite  = Event(syscall.FILE_NOTIFY_CHANGE_LAST_WRITE)
	FileNotifyChangeLastAccess = Event(syscall.FILE_NOTIFY_CHANGE_LAST_ACCESS)
	FileNotifyChangeCreation   = Event(syscall.FILE_NOTIFY_CHANGE_CREATION)
	FileNotifyChangeSecurity   = Event(syscallFileNotifyChangeSecurity)
)

const (
	fileNotifyChangeAll      = 0x17f // logical sum of all FileNotifyChange* events.
	fileNotifyChangeModified = fileNotifyChangeAll &^ (FileNotifyChangeFileName | FileNotifyChangeDirName)
)

// according to: http://msdn.microsoft.com/en-us/library/windows/desktop/aa365465(v=vs.85).aspx
// this flag should be declared in: http://golang.org/src/pkg/syscall/ztypes_windows.go
const syscallFileNotifyChangeSecurity = 0x00000100

// ReadDirectoryChangesW actions
// The following events are returned on the channel passed to Watch, but cannot
// be passed to Watch itself (see filters above). You can find a table showing
// the relation between actions and filteres at
// https://github.com/rjeczalik/notify/issues/10#issuecomment-66179535
// The msdn documentation on actions is part of
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa364391(v=vs.85).aspx
const (
	FileActionAdded          = Event(syscall.FILE_ACTION_ADDED) << 12
	FileActionRemoved        = Event(syscall.FILE_ACTION_REMOVED) << 12
	FileActionModified       = Event(syscall.FILE_ACTION_MODIFIED) << 14
	FileActionRenamedOldName = Event(syscall.FILE_ACTION_RENAMED_OLD_NAME) << 15
	FileActionRenamedNewName = Event(syscall.FILE_ACTION_RENAMED_NEW_NAME) << 16
)

const fileActionAll = 0x7f000 // logical sum of all FileAction* events.

var osestr = map[Event]string{
	FileNotifyChangeFileName:   "notify.FileNotifyChangeFileName",
	FileNotifyChangeDirName:    "notify.FileNotifyChangeDirName",
	FileNotifyChangeAttributes: "notify.FileNotifyChangeAttributes",
	FileNotifyChangeSize:       "notify.FileNotifyChangeSize",
	FileNotifyChangeLastWrite:  "notify.FileNotifyChangeLastWrite",
	FileNotifyChangeLastAccess: "notify.FileNotifyChangeLastAccess",
	FileNotifyChangeCreation:   "notify.FileNotifyChangeCreation",
	FileNotifyChangeSecurity:   "notify.FileNotifyChangeSecurity",

	FileActionAdded:          "notify.FileActionAdded",
	FileActionRemoved:        "notify.FileActionRemoved",
	FileActionModified:       "notify.FileActionModified",
	FileActionRenamedOldName: "notify.FileActionRenamedOldName",
	FileActionRenamedNewName: "notify.FileActionRenamedNewName",
}

const (
	fTypeUnknown uint8 = iota
	fTypeFile
	fTypeDirectory
)

// TODO(ppknap) : doc.
type event struct {
	pathw  []uint16
	name   string
	ftype  uint8
	action uint32
	filter uint32
	e      Event
}

func (e *event) Event() Event     { return e.e }
func (e *event) Path() string     { return filepath.Join(syscall.UTF16ToString(e.pathw), e.name) }
func (e *event) Sys() interface{} { return e.ftype }

func (e *event) isDir() (bool, error) {
	if e.ftype != fTypeUnknown {
		return e.ftype == fTypeDirectory, nil
	}
	fi, err := os.Stat(e.Path())
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}
