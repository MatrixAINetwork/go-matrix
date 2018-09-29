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

import "image"

// Cell is a rune with assigned Fg and Bg
type Cell struct {
	Ch rune
	Fg Attribute
	Bg Attribute
}

// Buffer is a renderable rectangle cell data container.
type Buffer struct {
	Area    image.Rectangle // selected drawing area
	CellMap map[image.Point]Cell
}

// At returns the cell at (x,y).
func (b Buffer) At(x, y int) Cell {
	return b.CellMap[image.Pt(x, y)]
}

// Set assigns a char to (x,y)
func (b Buffer) Set(x, y int, c Cell) {
	b.CellMap[image.Pt(x, y)] = c
}

// Bounds returns the domain for which At can return non-zero color.
func (b Buffer) Bounds() image.Rectangle {
	x0, y0, x1, y1 := 0, 0, 0, 0
	for p := range b.CellMap {
		if p.X > x1 {
			x1 = p.X
		}
		if p.X < x0 {
			x0 = p.X
		}
		if p.Y > y1 {
			y1 = p.Y
		}
		if p.Y < y0 {
			y0 = p.Y
		}
	}
	return image.Rect(x0, y0, x1+1, y1+1)
}

// SetArea assigns a new rect area to Buffer b.
func (b *Buffer) SetArea(r image.Rectangle) {
	b.Area.Max = r.Max
	b.Area.Min = r.Min
}

// Sync sets drawing area to the buffer's bound
func (b *Buffer) Sync() {
	b.SetArea(b.Bounds())
}

// NewCell returns a new cell
func NewCell(ch rune, fg, bg Attribute) Cell {
	return Cell{ch, fg, bg}
}

// Merge merges bs Buffers onto b
func (b *Buffer) Merge(bs ...Buffer) {
	for _, buf := range bs {
		for p, v := range buf.CellMap {
			b.Set(p.X, p.Y, v)
		}
		b.SetArea(b.Area.Union(buf.Area))
	}
}

// NewBuffer returns a new Buffer
func NewBuffer() Buffer {
	return Buffer{
		CellMap: make(map[image.Point]Cell),
		Area:    image.Rectangle{}}
}

// Fill fills the Buffer b with ch,fg and bg.
func (b Buffer) Fill(ch rune, fg, bg Attribute) {
	for x := b.Area.Min.X; x < b.Area.Max.X; x++ {
		for y := b.Area.Min.Y; y < b.Area.Max.Y; y++ {
			b.Set(x, y, Cell{ch, fg, bg})
		}
	}
}

// NewFilledBuffer returns a new Buffer filled with ch, fb and bg.
func NewFilledBuffer(x0, y0, x1, y1 int, ch rune, fg, bg Attribute) Buffer {
	buf := NewBuffer()
	buf.Area.Min = image.Pt(x0, y0)
	buf.Area.Max = image.Pt(x1, y1)

	for x := buf.Area.Min.X; x < buf.Area.Max.X; x++ {
		for y := buf.Area.Min.Y; y < buf.Area.Max.Y; y++ {
			buf.Set(x, y, Cell{ch, fg, bg})
		}
	}
	return buf
}
