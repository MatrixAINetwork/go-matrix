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

package rpc

import (
	"encoding/json"
	"testing"

	"github.com/matrix/go-matrix/common/math"
)

func TestBlockNumberJSONUnmarshal(t *testing.T) {
	tests := []struct {
		input    string
		mustFail bool
		expected BlockNumber
	}{
		0:  {`"0x"`, true, BlockNumber(0)},
		1:  {`"0x0"`, false, BlockNumber(0)},
		2:  {`"0X1"`, false, BlockNumber(1)},
		3:  {`"0x00"`, true, BlockNumber(0)},
		4:  {`"0x01"`, true, BlockNumber(0)},
		5:  {`"0x1"`, false, BlockNumber(1)},
		6:  {`"0x12"`, false, BlockNumber(18)},
		7:  {`"0x7fffffffffffffff"`, false, BlockNumber(math.MaxInt64)},
		8:  {`"0x8000000000000000"`, true, BlockNumber(0)},
		9:  {"0", true, BlockNumber(0)},
		10: {`"ff"`, true, BlockNumber(0)},
		11: {`"pending"`, false, PendingBlockNumber},
		12: {`"latest"`, false, LatestBlockNumber},
		13: {`"earliest"`, false, EarliestBlockNumber},
		14: {`someString`, true, BlockNumber(0)},
		15: {`""`, true, BlockNumber(0)},
		16: {``, true, BlockNumber(0)},
	}

	for i, test := range tests {
		var num BlockNumber
		err := json.Unmarshal([]byte(test.input), &num)
		if test.mustFail && err == nil {
			t.Errorf("Test %d should fail", i)
			continue
		}
		if !test.mustFail && err != nil {
			t.Errorf("Test %d should pass but got err: %v", i, err)
			continue
		}
		if num != test.expected {
			t.Errorf("Test %d got unexpected value, want %d, got %d", i, test.expected, num)
		}
	}
}
