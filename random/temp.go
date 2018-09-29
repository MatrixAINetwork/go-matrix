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
package random

import (
	"fmt"
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
)

var Prv *big.Int
var Pub []byte
var Address *big.Int
var Address1 common.Address

func asd() {

	fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAA")
	if Prv == nil {
		Prv, Pub, _ = getkey()
		log.Info("asd", "a", Prv, "as", Pub, "asdsa", Address1)
	}

	Address = big.NewInt(100)
	Address1 = common.BigToAddress(Address)
	log.Info("asd", "a", Prv, "as", Pub, "asdsa", Address1)

}

func GetKeyTransInfo(Heigh uint64, types string) map[common.Address][]byte {
	asd()
	mmap_private := make(map[common.Address][]byte)
	mmap_public := make(map[common.Address][]byte)

	mmap_private[Address1] = Prv.Bytes()
	mmap_public[Address1] = Pub

	fmt.Println("GGGGGGGGGGGGG", Prv, Prv.Bytes(), Pub)

	/*
		fmt.Println(Address1, mmap_private, mmap_public)

		private1 := mmap_private[Address1]
		public1 := mmap_public[Address1]
		if compare(public1, private1) == true {
			fmt.Println("biggo")
		} else {
			fmt.Println("sb")
		}*/
	if types == "private" {
		return mmap_private
	}
	return mmap_public
}
