// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

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
