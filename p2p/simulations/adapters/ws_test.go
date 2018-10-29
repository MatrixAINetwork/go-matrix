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
package adapters

import (
	"bytes"
	"testing"
	"time"
)

func TestFindWSAddr(t *testing.T) {
	line := `t=2018-05-02T19:00:45+0200 lvl=info msg="WebSocket endpoint opened"  node.id=26c65a606d1125a44695bc08573190d047152b6b9a776ccbbe593e90f91444d9c1ebdadac6a775ad9fdd0923468a1d698ed3a842c1fb89c1bc0f9d4801f8c39c url=ws://127.0.0.1:59975`
	buf := bytes.NewBufferString(line)
	got, err := findWSAddr(buf, 10*time.Second)
	if err != nil {
		t.Fatalf("Failed to find addr: %v", err)
	}
	expected := `ws://127.0.0.1:59975`

	if got != expected {
		t.Fatalf("Expected to get '%s', but got '%s'", expected, got)
	}
}
