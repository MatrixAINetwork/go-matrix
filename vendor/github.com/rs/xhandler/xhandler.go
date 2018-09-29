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
// Package xhandler provides a bridge between http.Handler and net/context.
//
// xhandler enforces net/context in your handlers without sacrificing
// compatibility with existing http.Handlers nor imposing a specific router.
//
// Thanks to net/context deadline management, xhandler is able to enforce
// a per request deadline and will cancel the context in when the client close
// the connection unexpectedly.
//
// You may create net/context aware middlewares pretty much the same way as
// you would with http.Handler.
package xhandler // import "github.com/rs/xhandler"

import (
	"net/http"

	"golang.org/x/net/context"
)

// HandlerC is a net/context aware http.Handler
type HandlerC interface {
	ServeHTTPC(context.Context, http.ResponseWriter, *http.Request)
}

// HandlerFuncC type is an adapter to allow the use of ordinary functions
// as an xhandler.Handler. If f is a function with the appropriate signature,
// xhandler.HandlerFuncC(f) is a xhandler.Handler object that calls f.
type HandlerFuncC func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTPC calls f(ctx, w, r).
func (f HandlerFuncC) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

// New creates a conventional http.Handler injecting the provided root
// context to sub handlers. This handler is used as a bridge between conventional
// http.Handler and context aware handlers.
func New(ctx context.Context, h HandlerC) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTPC(ctx, w, r)
	})
}
