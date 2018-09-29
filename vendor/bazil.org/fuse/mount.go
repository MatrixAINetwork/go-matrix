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
package fuse

import (
	"bufio"
	"errors"
	"io"
	"log"
	"sync"
)

var (
	// ErrOSXFUSENotFound is returned from Mount when the OSXFUSE
	// installation is not detected.
	//
	// Only happens on OS X. Make sure OSXFUSE is installed, or see
	// OSXFUSELocations for customization.
	ErrOSXFUSENotFound = errors.New("cannot locate OSXFUSE")
)

func neverIgnoreLine(line string) bool {
	return false
}

func lineLogger(wg *sync.WaitGroup, prefix string, ignore func(line string) bool, r io.ReadCloser) {
	defer wg.Done()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if ignore(line) {
			continue
		}
		log.Printf("%s: %s", prefix, line)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("%s, error reading: %v", prefix, err)
	}
}
