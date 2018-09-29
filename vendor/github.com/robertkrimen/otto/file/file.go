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
// Package file encapsulates the file abstractions used by the ast & parser.
//
package file

import (
	"fmt"
	"strings"

	"gopkg.in/sourcemap.v1"
)

// Idx is a compact encoding of a source position within a file set.
// It can be converted into a Position for a more convenient, but much
// larger, representation.
type Idx int

// Position describes an arbitrary source position
// including the filename, line, and column location.
type Position struct {
	Filename string // The filename where the error occurred, if any
	Offset   int    // The src offset
	Line     int    // The line number, starting at 1
	Column   int    // The column number, starting at 1 (The character count)

}

// A Position is valid if the line number is > 0.

func (self *Position) isValid() bool {
	return self.Line > 0
}

// String returns a string in one of several forms:
//
//	file:line:column    A valid position with filename
//	line:column         A valid position without filename
//	file                An invalid position with filename
//	-                   An invalid position without filename
//
func (self *Position) String() string {
	str := self.Filename
	if self.isValid() {
		if str != "" {
			str += ":"
		}
		str += fmt.Sprintf("%d:%d", self.Line, self.Column)
	}
	if str == "" {
		str = "-"
	}
	return str
}

// FileSet

// A FileSet represents a set of source files.
type FileSet struct {
	files []*File
	last  *File
}

// AddFile adds a new file with the given filename and src.
//
// This an internal method, but exported for cross-package use.
func (self *FileSet) AddFile(filename, src string) int {
	base := self.nextBase()
	file := &File{
		name: filename,
		src:  src,
		base: base,
	}
	self.files = append(self.files, file)
	self.last = file
	return base
}

func (self *FileSet) nextBase() int {
	if self.last == nil {
		return 1
	}
	return self.last.base + len(self.last.src) + 1
}

func (self *FileSet) File(idx Idx) *File {
	for _, file := range self.files {
		if idx <= Idx(file.base+len(file.src)) {
			return file
		}
	}
	return nil
}

// Position converts an Idx in the FileSet into a Position.
func (self *FileSet) Position(idx Idx) *Position {
	for _, file := range self.files {
		if idx <= Idx(file.base+len(file.src)) {
			return file.Position(idx - Idx(file.base))
		}
	}

	return nil
}

type File struct {
	name string
	src  string
	base int // This will always be 1 or greater
	sm   *sourcemap.Consumer
}

func NewFile(filename, src string, base int) *File {
	return &File{
		name: filename,
		src:  src,
		base: base,
	}
}

func (fl *File) WithSourceMap(sm *sourcemap.Consumer) *File {
	fl.sm = sm
	return fl
}

func (fl *File) Name() string {
	return fl.name
}

func (fl *File) Source() string {
	return fl.src
}

func (fl *File) Base() int {
	return fl.base
}

func (fl *File) Position(idx Idx) *Position {
	position := &Position{}

	offset := int(idx) - fl.base

	if offset >= len(fl.src) || offset < 0 {
		return nil
	}

	src := fl.src[:offset]

	position.Filename = fl.name
	position.Offset = offset
	position.Line = strings.Count(src, "\n") + 1

	if index := strings.LastIndex(src, "\n"); index >= 0 {
		position.Column = offset - index
	} else {
		position.Column = len(src) + 1
	}

	if fl.sm != nil {
		if f, _, l, c, ok := fl.sm.Source(position.Line, position.Column); ok {
			position.Filename, position.Line, position.Column = f, l, c
		}
	}

	return position
}
