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

package filter

import (
	"testing"
	"time"
)

// Simple test to check if baseline matching/mismatching filtering works.
func TestFilters(t *testing.T) {
	fm := New()
	fm.Start()

	// Register two filters to catch posted data
	first := make(chan struct{})
	fm.Install(Generic{
		Str1: "hello",
		Fn: func(data interface{}) {
			first <- struct{}{}
		},
	})
	second := make(chan struct{})
	fm.Install(Generic{
		Str1: "hello1",
		Str2: "hello",
		Fn: func(data interface{}) {
			second <- struct{}{}
		},
	})
	// Post an event that should only match the first filter
	fm.Notify(Generic{Str1: "hello"}, true)
	fm.Stop()

	// Ensure only the mathcing filters fire
	select {
	case <-first:
	case <-time.After(100 * time.Millisecond):
		t.Error("matching filter timed out")
	}
	select {
	case <-second:
		t.Error("mismatching filter fired")
	case <-time.After(100 * time.Millisecond):
	}
}
