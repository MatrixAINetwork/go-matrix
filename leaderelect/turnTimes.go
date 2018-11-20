// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/matrix/go-matrix/params"
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

	posTime := params.LRSPOSOutTime
	if consensusTurn == 0 {
		posTime += params.LRSParentMiningTime
	}

	return tt.GetBeginTime(consensusTurn) + posTime
}

func (tt *turnTimes) CalState(consensusTurn uint32, time int64) (st state, remainTime int64, reelectTurn uint32) {
	posTime := params.LRSPOSOutTime
	if consensusTurn == 0 {
		posTime += params.LRSParentMiningTime
	}

	passTime := time - tt.GetBeginTime(consensusTurn)
	if passTime < posTime {
		return stPos, posTime - passTime, 0
	}

	st = stReelect
	reelectTurn = uint32((passTime-posTime)/params.LRSReelectOutTime) + 1
	_, endTime := tt.CalTurnTime(consensusTurn, reelectTurn)
	remainTime = endTime - time
	return
}

func (tt *turnTimes) CalRemainTime(consensusTurn uint32, reelectTurn uint32, time int64) int64 {
	_, endTime := tt.CalTurnTime(consensusTurn, reelectTurn)
	return endTime - time
}

func (tt *turnTimes) CalTurnTime(consensusTurn uint32, reelectTurn uint32) (beginTime int64, endTime int64) {
	posTime := params.LRSPOSOutTime
	if consensusTurn == 0 {
		posTime += params.LRSParentMiningTime
	}

	if reelectTurn == 0 {
		beginTime = tt.GetBeginTime(consensusTurn)
		endTime = beginTime + posTime
	} else {
		beginTime = tt.GetBeginTime(consensusTurn) + posTime + int64(reelectTurn-1)*params.LRSReelectOutTime
		endTime = beginTime + params.LRSReelectOutTime
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
