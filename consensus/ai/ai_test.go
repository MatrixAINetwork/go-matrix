// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package ai

import (
	"github.com/MatrixAINetwork/go-matrix/log"
	"testing"
	"time"
)

func TestDigger(t *testing.T) {
	log.InitLog(5)

	Init("D:\\gopath\\bin\\picstore")

	resultCh := make(chan []byte, 0)
	stopCh := make(chan struct{})
	errCh := make(chan error)
	go Mining(1000, stopCh, resultCh, errCh)

	i := 0
loop:
	for {
		select {
		case result := <-resultCh:
			log.Info("挖矿结果获得", "result", result)
			return

		default:
			if i == 3 {
				log.Info("主动停止挖矿")
				close(stopCh)
				break loop
			}
			time.Sleep(time.Second)
			i++
		}
	}
	log.Info("结束")
}
