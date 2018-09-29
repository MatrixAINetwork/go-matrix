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
package reelection

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/random"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/man"
)

//func Post() {
//	blockNum := 20
//	for {
//
//		err := mc.PostEvent("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(blockNum)})
//		blockNum++
//		//fmt.Println("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(blockNum)})
//		log.Info("err", err)
//		time.Sleep(5 * time.Second)
//
//	}
//}
//
//func TestReElect(t *testing.T) {
//
//	electseed, err := random.NewElectionSeed()
//
//	log.Info("electseed", electseed)
//	log.Info("seed err", err)
//
//	var man *man.Matrix
//	reElect, err := New(man)
//	log.Info("err", err)
//
//	go Post()
//
//	time.Sleep(10000 * time.Second)
//	time.Sleep(3 * time.Second)
//	ans1, ans2, ans3 := reElect.readElectData(common.RoleMiner, 240)
//	fmt.Println("READ ELECT", ans1, ans2, ans3)
//	fmt.Println("READ ELECT", 240)
//
//	fmt.Println(reElect)
//}

func TestT(t *testing.T) {
	ans := big.NewInt(100)
	ans1 := common.BigToHash(ans)
	fmt.Println(ans1)

}
func TestCase(t *testing.T) {
	ans1, ans2 := GetAllElectedByHeight(big.NewInt(100), common.RoleMiner)
	fmt.Println(ans1, ans2)
}
