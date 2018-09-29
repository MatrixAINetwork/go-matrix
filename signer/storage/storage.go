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
//

package storage

import (
	"fmt"
)

type Storage interface {
	// Put stores a value by key. 0-length keys results in no-op
	Put(key, value string)
	// Get returns the previously stored value, or the empty string if it does not exist or key is of 0-length
	Get(key string) string
}

// EphemeralStorage is an in-memory storage that does
// not persist values to disk. Mainly used for testing
type EphemeralStorage struct {
	data      map[string]string
	namespace string
}

func (s *EphemeralStorage) Put(key, value string) {
	if len(key) == 0 {
		return
	}
	fmt.Printf("storage: put %v -> %v\n", key, value)
	s.data[key] = value
}

func (s *EphemeralStorage) Get(key string) string {
	if len(key) == 0 {
		return ""
	}
	fmt.Printf("storage: get %v\n", key)
	if v, exist := s.data[key]; exist {
		return v
	}
	return ""
}

func NewEphemeralStorage() Storage {
	s := &EphemeralStorage{
		data: make(map[string]string),
	}
	return s
}
