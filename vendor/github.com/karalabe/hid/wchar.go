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
// This file is https://github.com/orofarne/gowchar/blob/master/gowchar.go
//
// It was vendored inline to work around CGO limitations that don't allow C types
// to directly cross package API boundaries.
//
// The vendored file is licensed under the 3-clause BSD license, according to:
// https://github.com/orofarne/gowchar/blob/master/LICENSE

// +build !ios
// +build linux darwin windows

package hid

/*
#include <wchar.h>

const size_t SIZEOF_WCHAR_T = sizeof(wchar_t);

void gowchar_set (wchar_t *arr, int pos, wchar_t val)
{
	arr[pos] = val;
}

wchar_t gowchar_get (wchar_t *arr, int pos)
{
	return arr[pos];
}
*/
import "C"

import (
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

var sizeofWcharT C.size_t = C.size_t(C.SIZEOF_WCHAR_T)

func stringToWcharT(s string) (*C.wchar_t, C.size_t) {
	switch sizeofWcharT {
	case 2:
		return stringToWchar2(s) // Windows
	case 4:
		return stringToWchar4(s) // Unix
	default:
		panic(fmt.Sprintf("Invalid sizeof(wchar_t) = %v", sizeofWcharT))
	}
}

func wcharTToString(s *C.wchar_t) (string, error) {
	switch sizeofWcharT {
	case 2:
		return wchar2ToString(s) // Windows
	case 4:
		return wchar4ToString(s) // Unix
	default:
		panic(fmt.Sprintf("Invalid sizeof(wchar_t) = %v", sizeofWcharT))
	}
}

func wcharTNToString(s *C.wchar_t, size C.size_t) (string, error) {
	switch sizeofWcharT {
	case 2:
		return wchar2NToString(s, size) // Windows
	case 4:
		return wchar4NToString(s, size) // Unix
	default:
		panic(fmt.Sprintf("Invalid sizeof(wchar_t) = %v", sizeofWcharT))
	}
}

// Windows
func stringToWchar2(s string) (*C.wchar_t, C.size_t) {
	var slen int
	s1 := s
	for len(s1) > 0 {
		r, size := utf8.DecodeRuneInString(s1)
		if er, _ := utf16.EncodeRune(r); er == '\uFFFD' {
			slen += 1
		} else {
			slen += 2
		}
		s1 = s1[size:]
	}
	slen++ // \0
	res := C.malloc(C.size_t(slen) * sizeofWcharT)
	var i int
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if r1, r2 := utf16.EncodeRune(r); r1 != '\uFFFD' {
			C.gowchar_set((*C.wchar_t)(res), C.int(i), C.wchar_t(r1))
			i++
			C.gowchar_set((*C.wchar_t)(res), C.int(i), C.wchar_t(r2))
			i++
		} else {
			C.gowchar_set((*C.wchar_t)(res), C.int(i), C.wchar_t(r))
			i++
		}
		s = s[size:]
	}
	C.gowchar_set((*C.wchar_t)(res), C.int(slen-1), C.wchar_t(0)) // \0
	return (*C.wchar_t)(res), C.size_t(slen)
}

// Unix
func stringToWchar4(s string) (*C.wchar_t, C.size_t) {
	slen := utf8.RuneCountInString(s)
	slen++ // \0
	res := C.malloc(C.size_t(slen) * sizeofWcharT)
	var i int
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		C.gowchar_set((*C.wchar_t)(res), C.int(i), C.wchar_t(r))
		s = s[size:]
		i++
	}
	C.gowchar_set((*C.wchar_t)(res), C.int(slen-1), C.wchar_t(0)) // \0
	return (*C.wchar_t)(res), C.size_t(slen)
}

// Windows
func wchar2ToString(s *C.wchar_t) (string, error) {
	var i int
	var res string
	for {
		ch := C.gowchar_get(s, C.int(i))
		if ch == 0 {
			break
		}
		r := rune(ch)
		i++
		if !utf16.IsSurrogate(r) {
			if !utf8.ValidRune(r) {
				err := fmt.Errorf("Invalid rune at position %v", i)
				return "", err
			}
			res += string(r)
		} else {
			ch2 := C.gowchar_get(s, C.int(i))
			r2 := rune(ch2)
			r12 := utf16.DecodeRune(r, r2)
			if r12 == '\uFFFD' {
				err := fmt.Errorf("Invalid surrogate pair at position %v", i-1)
				return "", err
			}
			res += string(r12)
			i++
		}
	}
	return res, nil
}

// Unix
func wchar4ToString(s *C.wchar_t) (string, error) {
	var i int
	var res string
	for {
		ch := C.gowchar_get(s, C.int(i))
		if ch == 0 {
			break
		}
		r := rune(ch)
		if !utf8.ValidRune(r) {
			err := fmt.Errorf("Invalid rune at position %v", i)
			return "", err
		}
		res += string(r)
		i++
	}
	return res, nil
}

// Windows
func wchar2NToString(s *C.wchar_t, size C.size_t) (string, error) {
	var i int
	var res string
	N := int(size)
	for i < N {
		ch := C.gowchar_get(s, C.int(i))
		if ch == 0 {
			break
		}
		r := rune(ch)
		i++
		if !utf16.IsSurrogate(r) {
			if !utf8.ValidRune(r) {
				err := fmt.Errorf("Invalid rune at position %v", i)
				return "", err
			}

			res += string(r)
		} else {
			if i >= N {
				err := fmt.Errorf("Invalid surrogate pair at position %v", i-1)
				return "", err
			}
			ch2 := C.gowchar_get(s, C.int(i))
			r2 := rune(ch2)
			r12 := utf16.DecodeRune(r, r2)
			if r12 == '\uFFFD' {
				err := fmt.Errorf("Invalid surrogate pair at position %v", i-1)
				return "", err
			}
			res += string(r12)
			i++
		}
	}
	return res, nil
}

// Unix
func wchar4NToString(s *C.wchar_t, size C.size_t) (string, error) {
	var i int
	var res string
	N := int(size)
	for i < N {
		ch := C.gowchar_get(s, C.int(i))
		r := rune(ch)
		if !utf8.ValidRune(r) {
			err := fmt.Errorf("Invalid rune at position %v", i)
			return "", err
		}
		res += string(r)
		i++
	}
	return res, nil
}
