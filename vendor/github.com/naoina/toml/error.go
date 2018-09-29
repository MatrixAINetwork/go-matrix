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
package toml

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var (
	errArrayMultiType = errors.New("array can't contain multiple types")
)

// LineError is returned by Unmarshal, UnmarshalTable and Parse
// if the error is local to a line.
type LineError struct {
	Line        int
	StructField string
	Err         error
}

func (err *LineError) Error() string {
	field := ""
	if err.StructField != "" {
		field = "(" + err.StructField + ") "
	}
	return fmt.Sprintf("line %d: %s%v", err.Line, field, err.Err)
}

func lineError(line int, err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*LineError); ok {
		return err
	}
	return &LineError{Line: line, Err: err}
}

func lineErrorField(line int, field string, err error) error {
	if lerr, ok := err.(*LineError); ok {
		return lerr
	} else if err != nil {
		err = &LineError{Line: line, StructField: field, Err: err}
	}
	return err
}

type overflowError struct {
	kind reflect.Kind
	v    string
}

func (err *overflowError) Error() string {
	return fmt.Sprintf("value %s is out of range for %v", err.v, err.kind)
}

func convertNumError(kind reflect.Kind, err error) error {
	if numerr, ok := err.(*strconv.NumError); ok && numerr.Err == strconv.ErrRange {
		return &overflowError{kind, numerr.Num}
	}
	return err
}

type invalidUnmarshalError struct {
	typ reflect.Type
}

func (err *invalidUnmarshalError) Error() string {
	if err.typ == nil {
		return "toml: Unmarshal(nil)"
	}
	if err.typ.Kind() != reflect.Ptr {
		return "toml: Unmarshal(non-pointer " + err.typ.String() + ")"
	}
	return "toml: Unmarshal(nil " + err.typ.String() + ")"
}

type unmarshalTypeError struct {
	what string
	want string
	typ  reflect.Type
}

func (err *unmarshalTypeError) Error() string {
	msg := fmt.Sprintf("cannot unmarshal TOML %s into %s", err.what, err.typ)
	if err.want != "" {
		msg += " (need " + err.want + ")"
	}
	return msg
}

type marshalNilError struct {
	typ reflect.Type
}

func (err *marshalNilError) Error() string {
	return fmt.Sprintf("toml: cannot marshal nil %s", err.typ)
}

type marshalTableError struct {
	typ reflect.Type
}

func (err *marshalTableError) Error() string {
	return fmt.Sprintf("toml: cannot marshal %s as table, want struct or map type", err.typ)
}
