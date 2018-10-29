// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package depoistInfo

import (
	"fmt"
	"github.com/matrix/go-matrix/common/hexutil"
	"github.com/matrix/go-matrix/rpc"
	"math/big"
	"testing"
)

func TestBigInt(t *testing.T) {
	var tm *big.Int
	var h rpc.BlockNumber
	tm = big.NewInt(100)
	encode := hexutil.EncodeBig(tm)
	err := h.UnmarshalJSON([]byte(encode))
	fmt.Println("err", err)
	fmt.Printf("encode:%T   %v\n", encode, encode)
}
