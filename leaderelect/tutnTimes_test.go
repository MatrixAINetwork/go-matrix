// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/matrix/go-matrix/log"
	"testing"
	"time"
)

func TestTurnTime(t *testing.T) {
	log.InitLog(3)
	tt := newTurnTimes()

	tt.SetBeginTime(0, 121)

	curTime := time.Now().Unix()
	st, remainTime, reelectTurn := tt.CalState(0, curTime)
	log.INFO("TestTurnTime", "状态", st, "剩余时间", remainTime, "重选轮次", reelectTurn, "endTime", time.Unix(curTime+remainTime, 0).String())

	beginTime, endTime := tt.CalTurnTime(0, reelectTurn)
	log.INFO("TestTurnTime N轮次", "开始时间", time.Unix(beginTime, 0).String(), "结束", time.Unix(endTime, 0).String())
	beginTime, endTime = tt.CalTurnTime(0, reelectTurn+1)
	log.INFO("TestTurnTime N+1轮次", "开始时间", time.Unix(beginTime, 0).String(), "结束", time.Unix(endTime, 0).String())
	beginTime, endTime = tt.CalTurnTime(0, reelectTurn+2)
	log.INFO("TestTurnTime N+2轮次", "开始时间", time.Unix(beginTime, 0).String(), "结束", time.Unix(endTime, 0).String())

	err := tt.CheckTimeLegal(0, reelectTurn, time.Now().Unix())
	if err != nil {
		log.INFO("TestTurnTime", "检查消息", err)
	}

}
