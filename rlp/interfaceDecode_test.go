// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package rlp

import (
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/common"
	"math/big"
	"testing"
)

type testInterface interface {
	test1()
	test2()
	test3()
	GetConstructorType() uint16
}
type testStruct1 struct {
	A uint64
	B uint64
	C uint64
}

func (t *testStruct1) test1() {

}
func (t *testStruct1) test2() {

}
func (t *testStruct1) test3() {

}
func (t *testStruct1) GetConstructorType() uint16 {
	return 10
}

type testStruct2 struct {
	A uint64
	B uint64
	C uint64
	D uint64
}

func (t *testStruct2) test1() {

}
func (t *testStruct2) test2() {

}
func (t *testStruct2) test3() {

}
func (t *testStruct2) GetConstructorType() uint16 {
	return 20
}

type testStruct struct {
	Test1 testInterface //`rlp:"interface"`
	Test2 testInterface //`rlp:"interface"`
	//	Test3
}

func TestDecodeInterface1(t *testing.T) {

	testRlp := testStruct{&testStruct1{100, 100, 100}, &testStruct2{100, 100, 100, 100}}
	b, _ := EncodeToBytes(testRlp)
	t.Log(b)
	testRlp1 := testStruct{}
	//	testSlice1 := []testInterface{}
	InterfaceConstructorMap[testRlp.Test1.GetConstructorType()] = func() interface{} {
		return &testStruct1{}
	}
	InterfaceConstructorMap[testRlp.Test2.GetConstructorType()] = func() interface{} {
		return &testStruct2{}
	}
	DecodeBytes(b, &testRlp1)
	t.Log(testRlp1.Test1, testRlp1.Test2)
}
func TestDecodeInterface(t *testing.T) {
	testSlice := []testInterface{}
	test1 := testStruct1{100, 100, 100}
	testSlice = append(testSlice, &testStruct1{100, 100, 100}, &testStruct1{200, 200, 200})
	testSlice = append(testSlice, &testStruct2{100, 100, 100, 100}, &testStruct2{100, 100, 100, 100})
	InterfaceConstructorMap[test1.GetConstructorType()] = func() interface{} {
		return &testStruct1{}
	}
	InterfaceConstructorMap[testSlice[2].GetConstructorType()] = func() interface{} {
		return &testStruct2{}
	}
	b1, _ := EncodeToBytes(test1)
	b, _ := EncodeToBytes(testSlice)
	testSlice1 := []testInterface{}
	DecodeBytes(b, &testSlice1)
	DecodeBytes(b1, test1)
	t.Log(testSlice1[0], testSlice1[1], testSlice1[2], testSlice1[3])
	t.Log(test1)
}

type ByteStruct struct {
	A, B, C, D, E uint8
}

func TestByteDecode(t *testing.T) {
	test1 := ByteStruct{30, 1, 2, 3, 4}
	b1, _ := EncodeToBytes(test1)
	t.Log(b1)
	DecodeBytes(b1, &test1)
	t.Log(test1)
}

type Tx_to1 struct {
	Recipient *string `json:"to"       rlp:"nil"` // nil means contract creation
}
type Matrix_Extra1 struct {
	TxType     byte   `json:"txType" gencodec:"required"`
	LockHeight uint64 `json:"lockHeight" gencodec:"required"`
	//ExtraTo    []Tx_to1 `json:"extra_to" gencodec:"required"`
	ExtraTo []Tx_to1 ` rlp:"tail"` //
}
type txdata1 struct {
	AccountNonce uint64   `json:"nonce"    gencodec:"required"`
	Price        *big.Int `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64   `json:"gas"      gencodec:"required"`
	Recipient    *string  `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int `json:"value"    gencodec:"required"`
	Payload      []byte   `json:"input"    gencodec:"required"`

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`
	// This is only used when marshaling to JSON.
	Hash        *common.Hash `json:"hash" rlp:"-"`
	TxEnterType byte         `json:"TxEnterType" gencodec:"required"` //入池类型
	IsEntrustTx bool         `json:"TxEnterType" gencodec:"required"` //是否是委托
	//Extra []Matrix_Extra1 ` rlp:"tail"` //
	Extra []Matrix_Extra1 `json:"Extra" gencodec:"required"` //
}

//{"nonce":"0x10000000000000","gasPrice":"0x098bca5a00","gasLimit":"0x033450","to":"MAN.3eKSYmc89mnRMeekbjL2WtPAKaL4zhAw656wmpUBsQVjGaU668XTAPNj","value":"0x0de0b6b3a7640000","data":"0x","TxEnterType":"0x01","IsEntrustTx":1,"extra_to":[[1,0]],"chainId":20}
func TestByteDecode1(t *testing.T) {
	var test1 txdata1
	test1.AccountNonce = 0x10000000000000
	test1.Price = big.NewInt(0x098bca5a00)
	test1.GasLimit = 0x033450
	test1.Recipient = new(string)
	*test1.Recipient = "MAN.3eKSYmc89mnRMeekbjL2WtPAKaL4zhAw656wmpUBsQVjGaU668XTAPNj"
	test1.Amount = big.NewInt(0x0de0b6b3a7640000)
	test1.TxEnterType = 0
	test1.IsEntrustTx = false
	ext := new(Matrix_Extra1)
	ext.LockHeight = 0
	ext.TxType = 1

	//exto := make([]Tx_to1,1)
	//to1 := "MAN.3Yd45AeqcojaRmdkCDRJzPZFWqJZkysE76CyUNiqfsnqhjZ6u5BGjoN7"
	//exto[0].Recipient = &to1
	//ext.ExtraTo = exto
	test1.Extra = make([]Matrix_Extra1, 1)
	test1.Extra[0] = *ext

	b1, err := EncodeToBytes(test1)
	if err != nil {
		fmt.Println(1111)
	}
	fmt.Println(common.Bytes2Hex(b1))
	var test2 txdata1
	b1 = common.Hex2Bytes("f862871000000000000085098bca5a0083033450b83c4d414e2e33654b53596d6338396d6e524d65656b626a4c32577450414b614c347a684177363536776d7055427351566a476155363638585441504e6a880de0b6b3a7640000808080808080c20180")
	err = DecodeBytes(b1, &test2)
	if err != nil {
		//fmt.Println(*test2.Extra[0].ExtraTo[0].Recipient)
		fmt.Println(err)
	}
	t.Log(test2)
}
