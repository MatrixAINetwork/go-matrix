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
/*
Package registry is an expirmental package to facillitate altering the otto runtime via import.

This interface can change at any time.
*/
package registry

var registry []*Entry = make([]*Entry, 0)

type Entry struct {
	active bool
	source func() string
}

func newEntry(source func() string) *Entry {
	return &Entry{
		active: true,
		source: source,
	}
}

func (self *Entry) Enable() {
	self.active = true
}

func (self *Entry) Disable() {
	self.active = false
}

func (self Entry) Source() string {
	return self.source()
}

func Apply(callback func(Entry)) {
	for _, entry := range registry {
		if !entry.active {
			continue
		}
		callback(*entry)
	}
}

func Register(source func() string) *Entry {
	entry := newEntry(source)
	registry = append(registry, entry)
	return entry
}
