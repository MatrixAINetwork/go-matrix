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

// Package notify implements access to filesystem events.
//
// Notify is a high-level abstraction over filesystem watchers like inotify,
// kqueue, FSEvents, FEN or ReadDirectoryChangesW. Watcher implementations are
// split into two groups: ones that natively support recursive notifications
// (FSEvents and ReadDirectoryChangesW) and ones that do not (inotify, kqueue, FEN).
// For more details see watcher and recursiveWatcher interfaces in watcher.go
// source file.
//
// On top of filesystem watchers notify maintains a watchpoint tree, which provides
// a strategy for creating and closing filesystem watches and dispatching filesystem
// events to user channels.
//
// An event set is just an event list joint using bitwise OR operator
// into a single event value.
// Both the platform-independent (see Constants) and specific events can be used.
// Refer to the event_*.go source files for information about the available
// events.
//
// A filesystem watch or just a watch is platform-specific entity which represents
// a single path registered for notifications for specific event set. Setting a watch
// means using platform-specific API calls for creating / initializing said watch.
// For each watcher the API call is:
//
//   - FSEvents: FSEventStreamCreate
//   - inotify:  notify_add_watch
//   - kqueue:   kevent
//   - ReadDirectoryChangesW: CreateFile+ReadDirectoryChangesW
//   - FEN:      port_get
//
// To rewatch means to either shrink or expand an event set that was previously
// registered during watch operation for particular filesystem watch.
//
// A watchpoint is a list of user channel and event set pairs for particular
// path (watchpoint tree's node). A single watchpoint can contain multiple
// different user channels registered to listen for one or more events. A single
// user channel can be registered in one or more watchpoints, recursive and
// non-recursive ones as well.
package notify
