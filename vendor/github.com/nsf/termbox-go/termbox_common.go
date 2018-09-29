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
package termbox

// private API, common OS agnostic part

type cellbuf struct {
	width  int
	height int
	cells  []Cell
}

func (this *cellbuf) init(width, height int) {
	this.width = width
	this.height = height
	this.cells = make([]Cell, width*height)
}

func (this *cellbuf) resize(width, height int) {
	if this.width == width && this.height == height {
		return
	}

	oldw := this.width
	oldh := this.height
	oldcells := this.cells

	this.init(width, height)
	this.clear()

	minw, minh := oldw, oldh

	if width < minw {
		minw = width
	}
	if height < minh {
		minh = height
	}

	for i := 0; i < minh; i++ {
		srco, dsto := i*oldw, i*width
		src := oldcells[srco : srco+minw]
		dst := this.cells[dsto : dsto+minw]
		copy(dst, src)
	}
}

func (this *cellbuf) clear() {
	for i := range this.cells {
		c := &this.cells[i]
		c.Ch = ' '
		c.Fg = foreground
		c.Bg = background
	}
}

const cursor_hidden = -1

func is_cursor_hidden(x, y int) bool {
	return x == cursor_hidden || y == cursor_hidden
}
