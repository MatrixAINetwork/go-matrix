// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect2

import (
	"time"

	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
)

type turnTimes struct {
	beginTime             int64 // 轮次开始时间, 即区块头时间
	parentMiningTime      int64 // 预留父区块挖矿时间
	turnOutTime           int64 // 轮次超时时间
	reelectHandleInterval int64 // 重选处理间隔时间
}

func newTurnTimes() *turnTimes {
	tt := &turnTimes{
		beginTime:             0,
		parentMiningTime:      0,
		turnOutTime:           0,
		reelectHandleInterval: 0,
	}

	return tt
}

func (tt *turnTimes) SetTimeConfig(config *mc.LeaderConfig) error {
	if config == nil {
		return ErrParamsIsNil
	}

	tt.parentMiningTime = int64(config.ParentMiningTime)
	tt.turnOutTime = int64(config.PosOutTime)
	tt.reelectHandleInterval = int64(config.ReelectHandleInterval)
	return nil
}

func (tt *turnTimes) SetBeginTime(time int64) bool {
	if time <= tt.beginTime {
		return false
	}
	tt.beginTime = time
	return true
}

func (tt *turnTimes) GetBeginTime(consensusTurn uint32) int64 {
	beginTime, _ := tt.CalTurnTime(consensusTurn, 0)
	return beginTime
}

func (tt *turnTimes) GetPosEndTime(consensusTurn uint32) int64 {
	_, endTime := tt.CalTurnTime(consensusTurn, 0)
	return endTime
}

func (tt *turnTimes) CalState(consensusTurn uint32, time int64) (st stateDef, remainTime int64, reelectTurn uint32) {
	if tt.turnOutTime == 0 {
		log.Error("critical", "turnTimes", "turnOutTime == 0")
		return stReelect, 0, 0
	}
	posTime := tt.turnOutTime
	if consensusTurn == 0 {
		posTime += tt.parentMiningTime
	}

	passTime := time - tt.GetBeginTime(consensusTurn)
	if passTime < posTime {
		return stPos, posTime - passTime, 0
	}

	st = stReelect
	reelectTurn = uint32((passTime-posTime)/tt.turnOutTime) + 1
	_, endTime := tt.CalTurnTime(consensusTurn, reelectTurn)
	remainTime = endTime - time
	return
}

func (tt *turnTimes) CalTurnTime(consensusTurn uint32, reelectTurn uint32) (beginTime int64, endTime int64) {
	totalTurn := consensusTurn + reelectTurn
	if totalTurn == 0 {
		beginTime = tt.beginTime
		endTime = beginTime + tt.parentMiningTime + tt.turnOutTime
	} else {
		beginTime = tt.beginTime + tt.parentMiningTime + tt.turnOutTime*int64(totalTurn)
		endTime = beginTime + tt.turnOutTime
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
