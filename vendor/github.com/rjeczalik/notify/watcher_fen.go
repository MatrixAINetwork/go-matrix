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

import (
	"fmt"
	"os"
	"syscall"
)

// newTrigger returns implementation of trigger.
func newTrigger(pthLkp map[string]*watched) trigger {
	return &fen{
		pthLkp: pthLkp,
		cf:     newCfen(),
	}
}

// fen is a structure implementing trigger for FEN.
type fen struct {
	// p is a FEN port identifier
	p int
	// pthLkp is a structure mapping monitored files/dir with data about them,
	// shared with parent trg structure
	pthLkp map[string]*watched
	// cf wraps C operations for FEN
	cf cfen
}

// watched is a data structure representing watched file/directory.
type watched struct {
	trgWatched
}

// Stop implements trigger.
func (f *fen) Stop() error {
	return f.cf.portAlert(f.p)
}

// Close implements trigger.
func (f *fen) Close() (err error) {
	return syscall.Close(f.p)
}

// NewWatched implements trigger.
func (*fen) NewWatched(p string, fi os.FileInfo) (*watched, error) {
	return &watched{trgWatched{p: p, fi: fi}}, nil
}

// Record implements trigger.
func (f *fen) Record(w *watched) {
	f.pthLkp[w.p] = w
}

// Del implements trigger.
func (f *fen) Del(w *watched) {
	delete(f.pthLkp, w.p)
}

func inter2pe(n interface{}) PortEvent {
	pe, ok := n.(PortEvent)
	if !ok {
		panic(fmt.Sprintf("fen: type should be PortEvent, %T instead", n))
	}
	return pe
}

// Watched implements trigger.
func (f *fen) Watched(n interface{}) (*watched, int64, error) {
	pe := inter2pe(n)
	fo, ok := pe.PortevObject.(*FileObj)
	if !ok || fo == nil {
		panic(fmt.Sprintf("fen: type should be *FileObj, %T instead", fo))
	}
	w, ok := f.pthLkp[fo.Name]
	if !ok {
		return nil, 0, errNotWatched
	}
	return w, int64(pe.PortevEvents), nil
}

// init initializes FEN.
func (f *fen) Init() (err error) {
	f.p, err = f.cf.portCreate()
	return
}

func fi2fo(fi os.FileInfo, p string) FileObj {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		panic(fmt.Sprintf("fen: type should be *syscall.Stat_t, %T instead", st))
	}
	return FileObj{Name: p, Atim: st.Atim, Mtim: st.Mtim, Ctim: st.Ctim}
}

// Unwatch implements trigger.
func (f *fen) Unwatch(w *watched) error {
	return f.cf.portDissociate(f.p, FileObj{Name: w.p})
}

// Watch implements trigger.
func (f *fen) Watch(fi os.FileInfo, w *watched, e int64) error {
	return f.cf.portAssociate(f.p, fi2fo(fi, w.p), int(e))
}

// Wait implements trigger.
func (f *fen) Wait() (interface{}, error) {
	var (
		pe  PortEvent
		err error
	)
	err = f.cf.portGet(f.p, &pe)
	return pe, err
}

// IsStop implements trigger.
func (f *fen) IsStop(n interface{}, err error) bool {
	return err == syscall.EBADF || inter2pe(n).PortevSource == srcAlert
}

func init() {
	encode = func(e Event, dir bool) (o int64) {
		// Create event is not supported by FEN. Instead FileModified event will
		// be registered. If this event will be reported on dir which is to be
		// monitored for Create, dir will be rescanned and Create events will
		// be generated and returned for new files. In case of files,
		// if not requested FileModified event is reported, it will be ignored.
		o = int64(e &^ Create)
		if (e&Create != 0 && dir) || e&Write != 0 {
			o = (o &^ int64(Write)) | int64(FileModified)
		}
		// Following events are 'exception events' and as such cannot be requested
		// explicitly for monitoring or filtered out. If the will be reported
		// by FEN and not subscribed with by user, they will be filtered out by
		// watcher's logic.
		o &= int64(^Rename & ^Remove &^ FileDelete &^ FileRenameTo &^
			FileRenameFrom &^ Unmounted &^ MountedOver)
		return
	}
	nat2not = map[Event]Event{
		FileModified:   Write,
		FileRenameFrom: Rename,
		FileDelete:     Remove,
		FileAccess:     Event(0),
		FileAttrib:     Event(0),
		FileRenameTo:   Event(0),
		FileTrunc:      Event(0),
		FileNoFollow:   Event(0),
		Unmounted:      Event(0),
		MountedOver:    Event(0),
	}
	not2nat = map[Event]Event{
		Write:  FileModified,
		Rename: FileRenameFrom,
		Remove: FileDelete,
	}
}
