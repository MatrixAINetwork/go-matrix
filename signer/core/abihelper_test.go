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

package core

import (
	"fmt"
	"strings"
	"testing"

	"io/ioutil"
	"math/big"
	"reflect"

	"github.com/matrix/go-matrix/accounts/abi"
	"github.com/matrix/go-matrix/common"
)

func verify(t *testing.T, jsondata, calldata string, exp []interface{}) {

	abispec, err := abi.JSON(strings.NewReader(jsondata))
	if err != nil {
		t.Fatal(err)
	}
	cd := common.Hex2Bytes(calldata)
	sigdata, argdata := cd[:4], cd[4:]
	method, err := abispec.MethodById(sigdata)

	if err != nil {
		t.Fatal(err)
	}

	data, err := method.Inputs.UnpackValues(argdata)

	if len(data) != len(exp) {
		t.Fatalf("Mismatched length, expected %d, got %d", len(exp), len(data))
	}
	for i, elem := range data {
		if !reflect.DeepEqual(elem, exp[i]) {
			t.Fatalf("Unpack error, arg %d, got %v, want %v", i, elem, exp[i])
		}
	}
}
func TestNewUnpacker(t *testing.T) {
	type unpackTest struct {
		jsondata string
		calldata string
		exp      []interface{}
	}
	testcases := []unpackTest{
		{ // https://solidity.readthedocs.io/en/develop/abi-spec.html#use-of-dynamic-types
			`[{"type":"function","name":"f", "inputs":[{"type":"uint256"},{"type":"uint32[]"},{"type":"bytes10"},{"type":"bytes"}]}]`,
			// 0x123, [0x456, 0x789], "1234567890", "Hello, world!"
			"8be65246" + "00000000000000000000000000000000000000000000000000000000000001230000000000000000000000000000000000000000000000000000000000000080313233343536373839300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000004560000000000000000000000000000000000000000000000000000000000000789000000000000000000000000000000000000000000000000000000000000000d48656c6c6f2c20776f726c642100000000000000000000000000000000000000",
			[]interface{}{
				big.NewInt(0x123),
				[]uint32{0x456, 0x789},
				[10]byte{49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
				common.Hex2Bytes("48656c6c6f2c20776f726c6421"),
			},
		}, { // https://github.com/matrix/wiki/wiki/Matrix-Contract-ABI#examples
			`[{"type":"function","name":"sam","inputs":[{"type":"bytes"},{"type":"bool"},{"type":"uint256[]"}]}]`,
			//  "dave", true and [1,2,3]
			"a5643bf20000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000464617665000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003",
			[]interface{}{
				[]byte{0x64, 0x61, 0x76, 0x65},
				true,
				[]*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
			},
		}, {
			`[{"type":"function","name":"send","inputs":[{"type":"uint256"}]}]`,
			"a52c101e0000000000000000000000000000000000000000000000000000000000000012",
			[]interface{}{big.NewInt(0x12)},
		}, {
			`[{"type":"function","name":"compareAndApprove","inputs":[{"type":"address"},{"type":"uint256"},{"type":"uint256"}]}]`,
			"751e107900000000000000000000000000000133700000deadbeef00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001",
			[]interface{}{
				common.HexToAddress("0x00000133700000deadbeef000000000000000000"),
				new(big.Int).SetBytes([]byte{0x00}),
				big.NewInt(0x1),
			},
		},
	}
	for _, c := range testcases {
		verify(t, c.jsondata, c.calldata, c.exp)
	}

}

/*
func TestReflect(t *testing.T) {
	a := big.NewInt(0)
	b := new(big.Int).SetBytes([]byte{0x00})
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Nope, %v != %v", a, b)
	}
}
*/

func TestCalldataDecoding(t *testing.T) {

	// send(uint256)                              : a52c101e
	// compareAndApprove(address,uint256,uint256) : 751e1079
	// issue(address[],uint256)                   : 42958b54
	jsondata := `
[
	{"type":"function","name":"send","inputs":[{"name":"a","type":"uint256"}]},
	{"type":"function","name":"compareAndApprove","inputs":[{"name":"a","type":"address"},{"name":"a","type":"uint256"},{"name":"a","type":"uint256"}]},
	{"type":"function","name":"issue","inputs":[{"name":"a","type":"address[]"},{"name":"a","type":"uint256"}]},
	{"type":"function","name":"sam","inputs":[{"name":"a","type":"bytes"},{"name":"a","type":"bool"},{"name":"a","type":"uint256[]"}]}
]`
	//Expected failures
	for _, hexdata := range []string{
		"a52c101e00000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000042",
		"a52c101e000000000000000000000000000000000000000000000000000000000000001200",
		"a52c101e00000000000000000000000000000000000000000000000000000000000000",
		"a52c101e",
		"a52c10",
		"",
		// Too short
		"751e10790000000000000000000000000000000000000000000000000000000000000012",
		"751e1079FFffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		//Not valid multiple of 32
		"deadbeef00000000000000000000000000000000000000000000000000000000000000",
		//Too short 'issue'
		"42958b5400000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000042",
		// Too short compareAndApprove
		"a52c101e00ff0000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000042",
		// From https://github.com/matrix/wiki/wiki/Matrix-Contract-ABI
		// contains a bool with illegal values
		"a5643bf20000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000001100000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000464617665000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003",
	} {
		_, err := parseCallData(common.Hex2Bytes(hexdata), jsondata)
		if err == nil {
			t.Errorf("Expected decoding to fail: %s", hexdata)
		}
	}

	//Expected success
	for _, hexdata := range []string{
		// From https://github.com/matrix/wiki/wiki/Matrix-Contract-ABI
		"a5643bf20000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000464617665000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003",
		"a52c101e0000000000000000000000000000000000000000000000000000000000000012",
		"a52c101eFFffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"751e1079000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"42958b54" +
			// start of dynamic type
			"0000000000000000000000000000000000000000000000000000000000000040" +
			//uint256
			"0000000000000000000000000000000000000000000000000000000000000001" +
			// length of  array
			"0000000000000000000000000000000000000000000000000000000000000002" +
			// array values
			"000000000000000000000000000000000000000000000000000000000000dead" +
			"000000000000000000000000000000000000000000000000000000000000beef",
	} {
		_, err := parseCallData(common.Hex2Bytes(hexdata), jsondata)
		if err != nil {
			t.Errorf("Unexpected failure on input %s:\n %v (%d bytes) ", hexdata, err, len(common.Hex2Bytes(hexdata)))
		}
	}
}

func TestSelectorUnmarshalling(t *testing.T) {
	var (
		db        *AbiDb
		err       error
		abistring []byte
		abistruct abi.ABI
	)

	db, err = NewAbiDBFromFile("../../cmd/clef/4byte.json")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("DB size %v\n", db.Size())
	for id, selector := range db.db {

		abistring, err = MethodSelectorToAbi(selector)
		if err != nil {
			t.Error(err)
			return
		}
		abistruct, err = abi.JSON(strings.NewReader(string(abistring)))
		if err != nil {
			t.Error(err)
			return
		}
		m, err := abistruct.MethodById(common.Hex2Bytes(id[2:]))
		if err != nil {
			t.Error(err)
			return
		}
		if m.Sig() != selector {
			t.Errorf("Expected equality: %v != %v", m.Sig(), selector)
		}
	}

}

func TestCustomABI(t *testing.T) {
	d, err := ioutil.TempDir("", "signer-4byte-test")
	if err != nil {
		t.Fatal(err)
	}
	filename := fmt.Sprintf("%s/4byte_custom.json", d)
	abidb, err := NewAbiDBFromFiles("../../cmd/clef/4byte.json", filename)
	if err != nil {
		t.Fatal(err)
	}
	// Now we'll remove all existing signatures
	abidb.db = make(map[string]string)
	calldata := common.Hex2Bytes("a52c101edeadbeef")
	_, err = abidb.LookupMethodSelector(calldata)
	if err == nil {
		t.Fatalf("Should not find a match on empty db")
	}
	if err = abidb.AddSignature("send(uint256)", calldata); err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}
	_, err = abidb.LookupMethodSelector(calldata)
	if err != nil {
		t.Fatalf("Should find a match for abi signature, got: %v", err)
	}
	//Check that it wrote to file
	abidb2, err := NewAbiDBFromFile(filename)
	if err != nil {
		t.Fatalf("Failed to create new abidb: %v", err)
	}
	_, err = abidb2.LookupMethodSelector(calldata)
	if err != nil {
		t.Fatalf("Save failed: should find a match for abi signature after loading from disk")
	}
}
