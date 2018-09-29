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
// Copyright (c) 2014, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package errors provides common error types used throughout leveldb.
package errors

import (
	"errors"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Common errors.
var (
	ErrNotFound    = New("leveldb: not found")
	ErrReleased    = util.ErrReleased
	ErrHasReleaser = util.ErrHasReleaser
)

// New returns an error that formats as the given text.
func New(text string) error {
	return errors.New(text)
}

// ErrCorrupted is the type that wraps errors that indicate corruption in
// the database.
type ErrCorrupted struct {
	Fd  storage.FileDesc
	Err error
}

func (e *ErrCorrupted) Error() string {
	if !e.Fd.Zero() {
		return fmt.Sprintf("%v [file=%v]", e.Err, e.Fd)
	}
	return e.Err.Error()
}

// NewErrCorrupted creates new ErrCorrupted error.
func NewErrCorrupted(fd storage.FileDesc, err error) error {
	return &ErrCorrupted{fd, err}
}

// IsCorrupted returns a boolean indicating whether the error is indicating
// a corruption.
func IsCorrupted(err error) bool {
	switch err.(type) {
	case *ErrCorrupted:
		return true
	case *storage.ErrCorrupted:
		return true
	}
	return false
}

// ErrMissingFiles is the type that indicating a corruption due to missing
// files. ErrMissingFiles always wrapped with ErrCorrupted.
type ErrMissingFiles struct {
	Fds []storage.FileDesc
}

func (e *ErrMissingFiles) Error() string { return "file missing" }

// SetFd sets 'file info' of the given error with the given file.
// Currently only ErrCorrupted is supported, otherwise will do nothing.
func SetFd(err error, fd storage.FileDesc) error {
	switch x := err.(type) {
	case *ErrCorrupted:
		x.Fd = fd
		return x
	}
	return err
}
