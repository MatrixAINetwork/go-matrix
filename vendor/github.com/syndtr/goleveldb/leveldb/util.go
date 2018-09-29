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
// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package leveldb

import (
	"fmt"
	"sort"

	"github.com/syndtr/goleveldb/leveldb/storage"
)

func shorten(str string) string {
	if len(str) <= 8 {
		return str
	}
	return str[:3] + ".." + str[len(str)-3:]
}

var bunits = [...]string{"", "Ki", "Mi", "Gi", "Ti"}

func shortenb(bytes int) string {
	i := 0
	for ; bytes > 1024 && i < 4; i++ {
		bytes /= 1024
	}
	return fmt.Sprintf("%d%sB", bytes, bunits[i])
}

func sshortenb(bytes int) string {
	if bytes == 0 {
		return "~"
	}
	sign := "+"
	if bytes < 0 {
		sign = "-"
		bytes *= -1
	}
	i := 0
	for ; bytes > 1024 && i < 4; i++ {
		bytes /= 1024
	}
	return fmt.Sprintf("%s%d%sB", sign, bytes, bunits[i])
}

func sint(x int) string {
	if x == 0 {
		return "~"
	}
	sign := "+"
	if x < 0 {
		sign = "-"
		x *= -1
	}
	return fmt.Sprintf("%s%d", sign, x)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type fdSorter []storage.FileDesc

func (p fdSorter) Len() int {
	return len(p)
}

func (p fdSorter) Less(i, j int) bool {
	return p[i].Num < p[j].Num
}

func (p fdSorter) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func sortFds(fds []storage.FileDesc) {
	sort.Sort(fdSorter(fds))
}

func ensureBuffer(b []byte, n int) []byte {
	if cap(b) < n {
		return make([]byte, n)
	}
	return b[:n]
}
