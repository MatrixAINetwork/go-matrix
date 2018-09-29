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
// Copyright 2017 Zack Guo <zack.y.guo@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package termui

import (
	"fmt"
	"sync"
)

// event mixins
type WgtMgr map[string]WgtInfo

type WgtInfo struct {
	Handlers map[string]func(Event)
	WgtRef   Widget
	Id       string
}

type Widget interface {
	Id() string
}

func NewWgtInfo(wgt Widget) WgtInfo {
	return WgtInfo{
		Handlers: make(map[string]func(Event)),
		WgtRef:   wgt,
		Id:       wgt.Id(),
	}
}

func NewWgtMgr() WgtMgr {
	wm := WgtMgr(make(map[string]WgtInfo))
	return wm

}

func (wm WgtMgr) AddWgt(wgt Widget) {
	wm[wgt.Id()] = NewWgtInfo(wgt)
}

func (wm WgtMgr) RmWgt(wgt Widget) {
	wm.RmWgtById(wgt.Id())
}

func (wm WgtMgr) RmWgtById(id string) {
	delete(wm, id)
}

func (wm WgtMgr) AddWgtHandler(id, path string, h func(Event)) {
	if w, ok := wm[id]; ok {
		w.Handlers[path] = h
	}
}

func (wm WgtMgr) RmWgtHandler(id, path string) {
	if w, ok := wm[id]; ok {
		delete(w.Handlers, path)
	}
}

var counter struct {
	sync.RWMutex
	count int
}

func GenId() string {
	counter.Lock()
	defer counter.Unlock()

	counter.count += 1
	return fmt.Sprintf("%d", counter.count)
}

func (wm WgtMgr) WgtHandlersHook() func(Event) {
	return func(e Event) {
		for _, v := range wm {
			if k := findMatch(v.Handlers, e.Path); k != "" {
				v.Handlers[k](e)
			}
		}
	}
}

var DefaultWgtMgr WgtMgr

func (b *Block) Handle(path string, handler func(Event)) {
	if _, ok := DefaultWgtMgr[b.Id()]; !ok {
		DefaultWgtMgr.AddWgt(b)
	}

	DefaultWgtMgr.AddWgtHandler(b.Id(), path, handler)
}
