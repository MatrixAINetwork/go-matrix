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
// Copyright 2016 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"fmt"
	"strings"
)

// Palette defines the color used.
//
// An empty object Palette{} can be used to disable coloring.
type Palette struct {
	EOLReset string

	// Routine header.
	RoutineFirst string // The first routine printed.
	Routine      string // Following routines.
	CreatedBy    string

	// Call line.
	Package                string
	SourceFile             string
	FunctionStdLib         string
	FunctionStdLibExported string
	FunctionMain           string
	FunctionOther          string
	FunctionOtherExported  string
	Arguments              string
}

// CalcLengths returns the maximum length of the source lines and package names.
func CalcLengths(buckets Buckets, fullPath bool) (int, int) {
	srcLen := 0
	pkgLen := 0
	for _, bucket := range buckets {
		for _, line := range bucket.Signature.Stack.Calls {
			l := 0
			if fullPath {
				l = len(line.FullSourceLine())
			} else {
				l = len(line.SourceLine())
			}
			if l > srcLen {
				srcLen = l
			}
			l = len(line.Func.PkgName())
			if l > pkgLen {
				pkgLen = l
			}
		}
	}
	return srcLen, pkgLen
}

// functionColor returns the color to be used for the function name based on
// the type of package the function is in.
func (p *Palette) functionColor(line *Call) string {
	if line.IsStdlib() {
		if line.Func.IsExported() {
			return p.FunctionStdLibExported
		}
		return p.FunctionStdLib
	} else if line.IsPkgMain() {
		return p.FunctionMain
	} else if line.Func.IsExported() {
		return p.FunctionOtherExported
	}
	return p.FunctionOther
}

// routineColor returns the color for the header of the goroutines bucket.
func (p *Palette) routineColor(bucket *Bucket, multipleBuckets bool) string {
	if bucket.First() && multipleBuckets {
		return p.RoutineFirst
	}
	return p.Routine
}

// BucketHeader prints the header of a goroutine signature.
func (p *Palette) BucketHeader(bucket *Bucket, fullPath, multipleBuckets bool) string {
	extra := ""
	if bucket.SleepMax != 0 {
		if bucket.SleepMin != bucket.SleepMax {
			extra += fmt.Sprintf(" [%d~%d minutes]", bucket.SleepMin, bucket.SleepMax)
		} else {
			extra += fmt.Sprintf(" [%d minutes]", bucket.SleepMax)
		}
	}
	if bucket.Locked {
		extra += " [locked]"
	}
	created := bucket.CreatedBy.Func.PkgDotName()
	if created != "" {
		created += " @ "
		if fullPath {
			created += bucket.CreatedBy.FullSourceLine()
		} else {
			created += bucket.CreatedBy.SourceLine()
		}
		extra += p.CreatedBy + " [Created by " + created + "]"
	}
	return fmt.Sprintf(
		"%s%d: %s%s%s\n",
		p.routineColor(bucket, multipleBuckets), len(bucket.Routines),
		bucket.State, extra,
		p.EOLReset)
}

// callLine prints one stack line.
func (p *Palette) callLine(line *Call, srcLen, pkgLen int, fullPath bool) string {
	src := ""
	if fullPath {
		src = line.FullSourceLine()
	} else {
		src = line.SourceLine()
	}
	return fmt.Sprintf(
		"    %s%-*s %s%-*s %s%s%s(%s)%s",
		p.Package, pkgLen, line.Func.PkgName(),
		p.SourceFile, srcLen, src,
		p.functionColor(line), line.Func.Name(),
		p.Arguments, line.Args,
		p.EOLReset)
}

// StackLines prints one complete stack trace, without the header.
func (p *Palette) StackLines(signature *Signature, srcLen, pkgLen int, fullPath bool) string {
	out := make([]string, len(signature.Stack.Calls))
	for i := range signature.Stack.Calls {
		out[i] = p.callLine(&signature.Stack.Calls[i], srcLen, pkgLen, fullPath)
	}
	if signature.Stack.Elided {
		out = append(out, "    (...)")
	}
	return strings.Join(out, "\n") + "\n"
}
