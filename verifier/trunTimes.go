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
package verifier

import (
	"github.com/matrix/go-matrix/params/man"
)

type turnTimes struct {
	beginTimes map[uint32]int64
}

func newTurnTimes() *turnTimes {
	tt := &turnTimes{
		beginTimes: make(map[uint32]int64),
	}

	return tt
}

func (tt *turnTimes) SetBeginTime(consensusTurn uint32, time int64) bool {
	if oldTime, exist := tt.beginTimes[consensusTurn]; exist {
		if time <= oldTime {
			return false
		}
	}
	tt.beginTimes[consensusTurn] = time
	return true
}

func (tt *turnTimes) GetBeginTime(consensusTurn uint32) int64 {
	if beginTime, exist := tt.beginTimes[consensusTurn]; exist {
		return beginTime
	} else {
		return 0
	}
}

func (tt *turnTimes) CalState(consensusTurn uint32, time int64) (st state, remainTime int64, reelectTurn uint32) {
	posTime := man.LRSPOSOutTime
	if consensusTurn == 0 {
		posTime += man.LRSParentMiningTime
	}

	passTime := time - tt.GetBeginTime(consensusTurn)
	if passTime < posTime {
		return stPos, posTime - passTime, 0
	}

	st = stReelect
	reelectTurn = uint32((passTime-posTime)/man.LRSReelectOutTime) + 1
	remainTime = (passTime - posTime) % man.LRSReelectOutTime
	if remainTime == 0 {
		remainTime = man.LRSReelectOutTime
	}
	return
}

func (tt *turnTimes) CalRemainTime(consensusTurn uint32, reelectTurn uint32, time int64) int64 {
	posTime := man.LRSPOSOutTime
	if consensusTurn == 0 {
		posTime += man.LRSParentMiningTime
	}
	deadLine := tt.GetBeginTime(consensusTurn) + posTime + int64(reelectTurn)*man.LRSReelectOutTime
	return deadLine - time
}
