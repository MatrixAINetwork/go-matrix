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
	"github.com/pkg/errors"
	"time"
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

func (tt *turnTimes) GetPosEndTime(consensusTurn uint32) int64 {
	_, endTime := tt.CalTurnTime(consensusTurn, 0)
	return endTime

	posTime := man.LRSPOSOutTime
	if consensusTurn == 0 {
		posTime += man.LRSParentMiningTime
	}

	return tt.GetBeginTime(consensusTurn) + posTime
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
	_, endTime := tt.CalTurnTime(consensusTurn, reelectTurn)
	remainTime = endTime - time
	return
}

func (tt *turnTimes) CalRemainTime(consensusTurn uint32, reelectTurn uint32, time int64) int64 {
	_, endTime := tt.CalTurnTime(consensusTurn, reelectTurn)
	return endTime - time
}

func (tt *turnTimes) CalTurnTime(consensusTurn uint32, reelectTurn uint32) (beginTime int64, endTime int64) {
	posTime := man.LRSPOSOutTime
	if consensusTurn == 0 {
		posTime += man.LRSParentMiningTime
	}

	if reelectTurn == 0 {
		beginTime = tt.GetBeginTime(consensusTurn)
		endTime = beginTime + posTime
	} else {
		beginTime = tt.GetBeginTime(consensusTurn) + posTime + int64(reelectTurn-1)*man.LRSReelectOutTime
		endTime = beginTime + man.LRSReelectOutTime
	}
	return
}

func (tt *turnTimes) CheckTimeLegal(consensusTurn uint32, reelectTurn uint32, checkTime int64) error {
	beginTime, endTime := tt.CalTurnTime(consensusTurn, reelectTurn)
	if checkTime <= beginTime || checkTime >= endTime {
		return errors.Errorf("时间(%s)非法,轮次起止时间(%s - %s)",
			time.Unix(checkTime, 0).String(), time.Unix(beginTime, 0).String(), time.Unix(endTime, 0).String())
	}
	return nil
}
