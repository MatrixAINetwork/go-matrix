// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package amhash

import (
	"encoding/binary"
	"github.com/MatrixAINetwork/go-matrix/consensus/sm3"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11"
	"github.com/MatrixAINetwork/go-matrix/log"
	"math/rand"
)

func x11PowHash(src []byte, nonce uint64) []byte {
	return x11.Hash(append(src, uint32ToBytes(uint32(nonce))...))
}

func sm3PowHash(src []byte, nonce uint64) []byte {
	return sm3.Sm3Sum(append(src, uint32ToBytes(uint32(nonce))...))
}

func uint32ToBytes(num uint32) []byte {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, num)
	return data
}

func getRandNums(seed int64, max int, count int) []int {
	if count > max {
		log.Error("get rand numbers", "params err", "count > max", "count", count, "max", max)
		return []int{}
	}

	myRand := rand.New(rand.NewSource(seed))
	result := make([]int, 0)
	for len(result) < count {
		num := myRand.Intn(max)
		if isExistNum(num, result) {
			continue
		}
		result = append(result, num)
	}
	return result
}

func isExistNum(num int, list []int) bool {
	for _, item := range list {
		if num == item {
			return true
		}
	}
	return false
}

func Reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
