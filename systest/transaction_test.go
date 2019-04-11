// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package tests

import (
	"math/big"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/params"
)

func TestTransaction(t *testing.T) {
	t.Parallel()

	txt := new(testMatcher)
	txt.config(`^Homestead/`, params.ChainConfig{
		HomesteadBlock: big.NewInt(0),
	})
	txt.config(`^EIP155/`, params.ChainConfig{
		HomesteadBlock: big.NewInt(0),
		EIP150Block:    big.NewInt(0),
		EIP155Block:    big.NewInt(0),
		EIP158Block:    big.NewInt(0),
		ChainId:        big.NewInt(1),
	})
	txt.config(`^Byzantium/`, params.ChainConfig{
		HomesteadBlock: big.NewInt(0),
		EIP150Block:    big.NewInt(0),
		EIP155Block:    big.NewInt(0),
		EIP158Block:    big.NewInt(0),
		ByzantiumBlock: big.NewInt(0),
	})

	txt.walk(t, transactionTestDir, func(t *testing.T, name string, test *TransactionTest) {
		cfg := txt.findConfig(name)
		if err := txt.checkFailure(t, name, test.Run(cfg)); err != nil {
			t.Error(err)
		}
	})
}
