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
// +build !windows,!js

package runewidth

import (
	"os"
	"regexp"
	"strings"
)

var reLoc = regexp.MustCompile(`^[a-z][a-z][a-z]?(?:_[A-Z][A-Z])?\.(.+)`)

var mblenTable = map[string]int{
	"utf-8":   6,
	"utf8":    6,
	"jis":     8,
	"eucjp":   3,
	"euckr":   2,
	"euccn":   2,
	"sjis":    2,
	"cp932":   2,
	"cp51932": 2,
	"cp936":   2,
	"cp949":   2,
	"cp950":   2,
	"big5":    2,
	"gbk":     2,
	"gb2312":  2,
}

func isEastAsian(locale string) bool {
	charset := strings.ToLower(locale)
	r := reLoc.FindStringSubmatch(locale)
	if len(r) == 2 {
		charset = strings.ToLower(r[1])
	}

	if strings.HasSuffix(charset, "@cjk_narrow") {
		return false
	}

	for pos, b := range []byte(charset) {
		if b == '@' {
			charset = charset[:pos]
			break
		}
	}
	max := 1
	if m, ok := mblenTable[charset]; ok {
		max = m
	}
	if max > 1 && (charset[0] != 'u' ||
		strings.HasPrefix(locale, "ja") ||
		strings.HasPrefix(locale, "ko") ||
		strings.HasPrefix(locale, "zh")) {
		return true
	}
	return false
}

// IsEastAsian return true if the current locale is CJK
func IsEastAsian() bool {
	locale := os.Getenv("LC_CTYPE")
	if locale == "" {
		locale = os.Getenv("LANG")
	}

	// ignore C locale
	if locale == "POSIX" || locale == "C" {
		return false
	}
	if len(locale) > 1 && locale[0] == 'C' && (locale[1] == '.' || locale[1] == '-') {
		return false
	}

	return isEastAsian(locale)
}
