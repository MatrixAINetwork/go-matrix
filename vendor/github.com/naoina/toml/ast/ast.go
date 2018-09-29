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
package ast

import (
	"strconv"
	"strings"
	"time"
)

type Position struct {
	Begin int
	End   int
}

type Value interface {
	Pos() int
	End() int
	Source() string
}

type String struct {
	Position Position
	Value    string
	Data     []rune
}

func (s *String) Pos() int {
	return s.Position.Begin
}

func (s *String) End() int {
	return s.Position.End
}

func (s *String) Source() string {
	return string(s.Data)
}

type Integer struct {
	Position Position
	Value    string
	Data     []rune
}

func (i *Integer) Pos() int {
	return i.Position.Begin
}

func (i *Integer) End() int {
	return i.Position.End
}

func (i *Integer) Source() string {
	return string(i.Data)
}

func (i *Integer) Int() (int64, error) {
	return strconv.ParseInt(i.Value, 10, 64)
}

type Float struct {
	Position Position
	Value    string
	Data     []rune
}

func (f *Float) Pos() int {
	return f.Position.Begin
}

func (f *Float) End() int {
	return f.Position.End
}

func (f *Float) Source() string {
	return string(f.Data)
}

func (f *Float) Float() (float64, error) {
	return strconv.ParseFloat(f.Value, 64)
}

type Boolean struct {
	Position Position
	Value    string
	Data     []rune
}

func (b *Boolean) Pos() int {
	return b.Position.Begin
}

func (b *Boolean) End() int {
	return b.Position.End
}

func (b *Boolean) Source() string {
	return string(b.Data)
}

func (b *Boolean) Boolean() (bool, error) {
	return strconv.ParseBool(b.Value)
}

type Datetime struct {
	Position Position
	Value    string
	Data     []rune
}

func (d *Datetime) Pos() int {
	return d.Position.Begin
}

func (d *Datetime) End() int {
	return d.Position.End
}

func (d *Datetime) Source() string {
	return string(d.Data)
}

func (d *Datetime) Time() (time.Time, error) {
	switch {
	case !strings.Contains(d.Value, ":"):
		return time.Parse("2006-01-02", d.Value)
	case !strings.Contains(d.Value, "-"):
		return time.Parse("15:04:05.999999999", d.Value)
	default:
		return time.Parse(time.RFC3339Nano, d.Value)
	}
}

type Array struct {
	Position Position
	Value    []Value
	Data     []rune
}

func (a *Array) Pos() int {
	return a.Position.Begin
}

func (a *Array) End() int {
	return a.Position.End
}

func (a *Array) Source() string {
	return string(a.Data)
}

type TableType uint8

const (
	TableTypeNormal TableType = iota
	TableTypeArray
)

var tableTypes = [...]string{
	"normal",
	"array",
}

func (t TableType) String() string {
	return tableTypes[t]
}

type Table struct {
	Position Position
	Line     int
	Name     string
	Fields   map[string]interface{}
	Type     TableType
	Data     []rune
}

func (t *Table) Pos() int {
	return t.Position.Begin
}

func (t *Table) End() int {
	return t.Position.End
}

func (t *Table) Source() string {
	return string(t.Data)
}

type KeyValue struct {
	Key   string
	Value Value
	Line  int
}
