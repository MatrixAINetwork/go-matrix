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
package duktape

const (
	CompileEval uint = 1 << iota
	CompileFunction
	CompileStrict
	CompileSafe
	CompileNoResult
	CompileNoSource
	CompileStrlen
)

const (
	TypeNone Type = iota
	TypeUndefined
	TypeNull
	TypeBoolean
	TypeNumber
	TypeString
	TypeObject
	TypeBuffer
	TypePointer
	TypeLightFunc
)

const (
	TypeMaskNone uint = 1 << iota
	TypeMaskUndefined
	TypeMaskNull
	TypeMaskBoolean
	TypeMaskNumber
	TypeMaskString
	TypeMaskObject
	TypeMaskBuffer
	TypeMaskPointer
	TypeMaskLightFunc
)

const (
	EnumIncludeNonenumerable uint = 1 << iota
	EnumIncludeInternal
	EnumOwnPropertiesOnly
	EnumArrayIndicesOnly
	EnumSortArrayIndices
	NoProxyBehavior
)

const (
	ErrNone int = 0

	// Internal to Duktape
	ErrUnimplemented int = 50 + iota
	ErrUnsupported
	ErrInternal
	ErrAlloc
	ErrAssertion
	ErrAPI
	ErrUncaughtError
)

const (
	// Common prototypes
	ErrError int = 1 + iota
	ErrEval
	ErrRange
	ErrReference
	ErrSyntax
	ErrType
	ErrURI
)

const (
	// Returned error values
	ErrRetUnimplemented int = -(ErrUnimplemented + iota)
	ErrRetUnsupported
	ErrRetInternal
	ErrRetAlloc
	ErrRetAssertion
	ErrRetAPI
	ErrRetUncaughtError
)

const (
	ErrRetError int = -(ErrError + iota)
	ErrRetEval
	ErrRetRange
	ErrRetReference
	ErrRetSyntax
	ErrRetType
	ErrRetURI
)

const (
	ExecSuccess = iota
	ExecError
)

const (
	LogTrace int = iota
	LogDebug
	LogInfo
	LogWarn
	LogError
	LogFatal
)

const (
	BufobjDuktapeAuffer     = 0
	BufobjNodejsAuffer      = 1
	BufobjArraybuffer       = 2
	BufobjDataview          = 3
	BufobjInt8array         = 4
	BufobjUint8array        = 5
	BufobjUint8clampedarray = 6
	BufobjInt16array        = 7
	BufobjUint16array       = 8
	BufobjInt32array        = 9
	BufobjUint32array       = 10
	BufobjFloat32array      = 11
	BufobjFloat64array      = 12
)
