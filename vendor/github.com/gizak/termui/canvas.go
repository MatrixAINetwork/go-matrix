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

/*
dots:
   ,___,
   |1 4|
   |2 5|
   |3 6|
   |7 8|
   `````
*/

var brailleBase = '\u2800'

var brailleOftMap = [4][2]rune{
	{'\u0001', '\u0008'},
	{'\u0002', '\u0010'},
	{'\u0004', '\u0020'},
	{'\u0040', '\u0080'}}

// Canvas contains drawing map: i,j -> rune
type Canvas map[[2]int]rune

// NewCanvas returns an empty Canvas
func NewCanvas() Canvas {
	return make(map[[2]int]rune)
}

func chOft(x, y int) rune {
	return brailleOftMap[y%4][x%2]
}

func (c Canvas) rawCh(x, y int) rune {
	if ch, ok := c[[2]int{x, y}]; ok {
		return ch
	}
	return '\u0000' //brailleOffset
}

// return coordinate in terminal
func chPos(x, y int) (int, int) {
	return y / 4, x / 2
}

// Set sets a point (x,y) in the virtual coordinate
func (c Canvas) Set(x, y int) {
	i, j := chPos(x, y)
	ch := c.rawCh(i, j)
	ch |= chOft(x, y)
	c[[2]int{i, j}] = ch
}

// Unset removes point (x,y)
func (c Canvas) Unset(x, y int) {
	i, j := chPos(x, y)
	ch := c.rawCh(i, j)
	ch &= ^chOft(x, y)
	c[[2]int{i, j}] = ch
}

// Buffer returns un-styled points
func (c Canvas) Buffer() Buffer {
	buf := NewBuffer()
	for k, v := range c {
		buf.Set(k[0], k[1], Cell{Ch: v + brailleBase})
	}
	return buf
}
