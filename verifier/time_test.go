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
	"github.com/matrix/go-matrix/log"
	"testing"
	"time"
)

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
		case <-timer.C:
			log.Info("超时了!!!")
		}
	}
}
