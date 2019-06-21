// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package manapi

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"math/big"
	"testing"
)

type JsonTest1 struct {
	Address1 common.Address
	Address2 common.Address `json:"address_2"`
}
type JsonTest2 struct {
	Num1 *big.Int
	Num2 *big.Int `json:"num_2"`
}
type JsonTest3 struct {
	Test1 JsonTest1
	Test11 []JsonTest1
	Test2 JsonTest2
	Test22 []JsonTest2
}

func TestJsonMarshal(t* testing.T)  {
	var testa *big.Int
	MarshalInterface(&testa)
	test3 := JsonTest3{
		Test11:[]JsonTest1{JsonTest1{},JsonTest1{}},
		Test2:JsonTest2{big.NewInt(10),big.NewInt(2)},
		Test22:[]JsonTest2{JsonTest2{},JsonTest2{}},
	}
	data,err := MarshalInterface(&test3)
	t.Log(string(data),err)
}