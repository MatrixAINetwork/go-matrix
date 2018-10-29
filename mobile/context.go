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

// Contains all the wrappers from the golang.org/x/net/context package to support
// client side context management on mobile platforms.

package gman

import (
	"context"
	"time"
)

// Context carries a deadline, a cancelation signal, and other values across API
// boundaries.
type Context struct {
	context context.Context
	cancel  context.CancelFunc
}

// NewContext returns a non-nil, empty Context. It is never canceled, has no
// values, and has no deadline. It is typically used by the main function,
// initialization, and tests, and as the top-level Context for incoming requests.
func NewContext() *Context {
	return &Context{
		context: context.Background(),
	}
}

// WithCancel returns a copy of the original context with cancellation mechanism
// included.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
func (c *Context) WithCancel() *Context {
	child, cancel := context.WithCancel(c.context)
	return &Context{
		context: child,
		cancel:  cancel,
	}
}

// WithDeadline returns a copy of the original context with the deadline adjusted
// to be no later than the specified time.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
func (c *Context) WithDeadline(sec int64, nsec int64) *Context {
	child, cancel := context.WithDeadline(c.context, time.Unix(sec, nsec))
	return &Context{
		context: child,
		cancel:  cancel,
	}
}

// WithTimeout returns a copy of the original context with the deadline adjusted
// to be no later than now + the duration specified.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
func (c *Context) WithTimeout(nsec int64) *Context {
	child, cancel := context.WithTimeout(c.context, time.Duration(nsec))
	return &Context{
		context: child,
		cancel:  cancel,
	}
}
