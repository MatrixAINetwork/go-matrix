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
package testall

import (
	"fmt"
	"testing"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/p2p/discover"
)

func TestAsd(t *testing.T) {
	var v_s_account = []string{"0x1a8557a5830113ad675a9cb6f2d8a46d471edb8e", "0x20f4c8656cbac7de0e56e3e39d63e872393f089d"}
	var v_account = []common.Address{}
	var v_s_id = []string{"18b37bc680e739836fe0d8cca7c03a08a1162ff30fbb5049151d7bda951ec6e053916cb5df78f9b87e28a657230cc17d96cbde5fddc2f4571e5b606ec2a3a7a8", "4b2f638f46c7ae5b1564ca7015d716621848a0d9be66f1d1e91d566d2a70eedc2f11e92b743acb8d97dec3fb412c1b2f66afd7fbb9399d4fb2423619eaa514c7"}
	var v_id = []discover.NodeID{}
	/*
		var m = []string{}

		var account2node = map[string]string{
			"0x1a8557a5830113ad675a9cb6f2d8a46d471edb8e": "18b37bc680e739836fe0d8cca7c03a08a1162ff30fbb5049151d7bda951ec6e053916cb5df78f9b87e28a657230cc17d96cbde5fddc2f4571e5b606ec2a3a7a8",
			"0x20f4c8656cbac7de0e56e3e39d63e872393f089d": "4b2f638f46c7ae5b1564ca7015d716621848a0d9be66f1d1e91d566d2a70eedc2f11e92b743acb8d97dec3fb412c1b2f66afd7fbb9399d4fb2423619eaa514c7",
		}
	*/
	for _, v := range v_s_account {
		v_account = append(v_account, common.HexToAddress(v))
	}
	fmt.Println(v_account)

	for _, v := range v_s_id {
		ans, err := discover.HexID(v)
		fmt.Println(ans, err)
		v_id = append(v_id, ans)
	}
	fmt.Println(v_id)
}
