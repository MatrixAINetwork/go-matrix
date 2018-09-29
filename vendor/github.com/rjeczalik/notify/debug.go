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
// Copyright (c) 2014-2015 The Notify Authors. All rights reserved.
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package notify

import (
	"log"
	"os"
	"runtime"
	"strings"
)

var dbgprint func(...interface{})

var dbgprintf func(string, ...interface{})

var dbgcallstack func(max int) []string

func init() {
	if _, ok := os.LookupEnv("NOTIFY_DEBUG"); ok || debugTag {
		log.SetOutput(os.Stdout)
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
		dbgprint = func(v ...interface{}) {
			v = append([]interface{}{"[D] "}, v...)
			log.Println(v...)
		}
		dbgprintf = func(format string, v ...interface{}) {
			format = "[D] " + format
			log.Printf(format, v...)
		}
		dbgcallstack = func(max int) []string {
			pc, stack := make([]uintptr, max), make([]string, 0, max)
			runtime.Callers(2, pc)
			for _, pc := range pc {
				if f := runtime.FuncForPC(pc); f != nil {
					fname := f.Name()
					idx := strings.LastIndex(fname, string(os.PathSeparator))
					if idx != -1 {
						stack = append(stack, fname[idx+1:])
					} else {
						stack = append(stack, fname)
					}
				}
			}
			return stack
		}
		return
	}
	dbgprint = func(v ...interface{}) {}
	dbgprintf = func(format string, v ...interface{}) {}
	dbgcallstack = func(max int) []string { return nil }
}
