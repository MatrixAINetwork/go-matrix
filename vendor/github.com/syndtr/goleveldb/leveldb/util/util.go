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
// Copyright (c) 2013, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package util provides utilities used throughout leveldb.
package util

import (
	"errors"
)

var (
	ErrReleased    = errors.New("leveldb: resource already relesed")
	ErrHasReleaser = errors.New("leveldb: releaser already defined")
)

// Releaser is the interface that wraps the basic Release method.
type Releaser interface {
	// Release releases associated resources. Release should always success
	// and can be called multiple times without causing error.
	Release()
}

// ReleaseSetter is the interface that wraps the basic SetReleaser method.
type ReleaseSetter interface {
	// SetReleaser associates the given releaser to the resources. The
	// releaser will be called once coresponding resources released.
	// Calling SetReleaser with nil will clear the releaser.
	//
	// This will panic if a releaser already present or coresponding
	// resource is already released. Releaser should be cleared first
	// before assigned a new one.
	SetReleaser(releaser Releaser)
}

// BasicReleaser provides basic implementation of Releaser and ReleaseSetter.
type BasicReleaser struct {
	releaser Releaser
	released bool
}

// Released returns whether Release method already called.
func (r *BasicReleaser) Released() bool {
	return r.released
}

// Release implements Releaser.Release.
func (r *BasicReleaser) Release() {
	if !r.released {
		if r.releaser != nil {
			r.releaser.Release()
			r.releaser = nil
		}
		r.released = true
	}
}

// SetReleaser implements ReleaseSetter.SetReleaser.
func (r *BasicReleaser) SetReleaser(releaser Releaser) {
	if r.released {
		panic(ErrReleased)
	}
	if r.releaser != nil && releaser != nil {
		panic(ErrHasReleaser)
	}
	r.releaser = releaser
}

type NoopReleaser struct{}

func (NoopReleaser) Release() {}
