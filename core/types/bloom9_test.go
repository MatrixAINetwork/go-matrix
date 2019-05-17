// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package types

import (
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/common"
	"math/big"
	"testing"
	"time"
)

func TestBloom(t *testing.T) {
	positive := []string{
		"testtest",
		"test",
		"hallo",
		"other",
	}
	negative := []string{
		"tes",
		"lo",
	}
	time1 := time.Now().UnixNano()
	for i := 0; i < 10000; i++ {
		DeriveShaHash([]common.Hash{})
	}
	time2 := time.Now().UnixNano()
	fmt.Println(time2 - time1)
	var bloom Bloom
	for _, data := range positive {
		bloom.Add(new(big.Int).SetBytes([]byte(data)))
	}

	for _, data := range positive {
		if !bloom.TestBytes([]byte(data)) {
			t.Error("expected", data, "to test true")
		}
	}
	for _, data := range negative {
		if bloom.TestBytes([]byte(data)) {
			t.Error("did not expect", data, "to test true")
		}
	}
}

/*
import (
	"testing"

	"github.com/MatrixAINetwork/go-matrix/core/state"
)

func TestBloom9(t *testing.T) {
	testCase := []byte("testtest")
	bin := LogsBloom([]state.Log{
		{testCase, [][]byte{[]byte("hellohello")}, nil},
	}).Bytes()
	res := BloomLookup(bin, testCase)

	if !res {
		t.Errorf("Bloom lookup failed")
	}
}


func TestAddress(t *testing.T) {
	block := &Block{}
	block.Coinbase = common.Hex2Bytes("22341ae42d6dd7384bc8584e50419ea3ac75b83f")
	fmt.Printf("%x\n", crypto.Keccak256(block.Coinbase))

	bin := CreateBloom(block)
	fmt.Printf("bin = %x\n", common.LeftPadBytes(bin, 64))
}
*/
