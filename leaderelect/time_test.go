// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/MatrixAINetwork/go-matrix/log"
	"math/big"
	"sort"
	"testing"
	"time"
)

func TestSliceSort(t *testing.T) {
	log.InitLog(3)
	testSlice := make([]*big.Int, 0)
	testSlice = append(testSlice, big.NewInt(10), big.NewInt(-1), big.NewInt(22), big.NewInt(6))
	log.INFO("before sort", "slice", testSlice)
	sort.Slice(testSlice, func(i, j int) bool {
		return testSlice[i].Cmp(testSlice[j]) > 0
	})
	log.INFO("after sort", "slice", testSlice)
}

func TestTimer(t *testing.T) {
	log.InitLog(3)
	recvCh := make(chan struct{})
	go TimerRunning(t, recvCh)

	//time.Sleep(7 * time.Second)
	recvCh <- struct{}{}
	time.Sleep(11111 * time.Second)
}

func TimerRunning(t *testing.T, recv chan struct{}) {
	timer := time.NewTimer(10 * time.Second)
	log.Info("开始定时器")
	for {
		select {
		case <-recv:
			log.Info("收到停止消息")
			time.Sleep(12 * time.Second)
			log.Info("停止定时器")
			result := timer.Reset(10 * time.Second)
			log.Info("重置定时器", "结果", result)
			if result == false {
				<-timer.C
			}

		case <-timer.C:
			log.Info("超时了!!!")
		}
	}
}
