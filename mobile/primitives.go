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

// Contains various wrappers for primitive types.

package gman

import (
	"errors"
	"fmt"
)

// Strings represents s slice of strs.
type Strings struct{ strs []string }

// Size returns the number of strs in the slice.
func (s *Strings) Size() int {
	return len(s.strs)
}

// Get returns the string at the given index from the slice.
func (s *Strings) Get(index int) (str string, _ error) {
	if index < 0 || index >= len(s.strs) {
		return "", errors.New("index out of bounds")
	}
	return s.strs[index], nil
}

// Set sets the string at the given index in the slice.
func (s *Strings) Set(index int, str string) error {
	if index < 0 || index >= len(s.strs) {
		return errors.New("index out of bounds")
	}
	s.strs[index] = str
	return nil
}

// String implements the Stringer interface.
func (s *Strings) String() string {
	return fmt.Sprintf("%v", s.strs)
}
