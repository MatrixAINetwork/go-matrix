// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package common

import (
	"github.com/pkg/errors"
	"time"
)

var (
	ErrParam = errors.New("param err")
)

type ResendMsgCtrl struct {
	curTimes uint32
	maxTimes uint32
	interval time.Duration
	msg      interface{}
	sendFunc func(interface{}, uint32)
	closeCh  chan struct{}
}

func NewResendMsgCtrl(msg interface{}, sendFunc func(interface{}, uint32), interval int64, times uint32) (*ResendMsgCtrl, error) {
	if sendFunc == nil || interval <= 0 {
		return nil, ErrParam
	}
	ctrl := &ResendMsgCtrl{
		curTimes: 0,
		maxTimes: times,
		interval: time.Duration(interval) * time.Second,
		msg:      msg,
		sendFunc: sendFunc,
		closeCh:  make(chan struct{}),
	}
	go ctrl.running()
	return ctrl, nil
}

func (self *ResendMsgCtrl) Close() {
	close(self.closeCh)
}

func (self *ResendMsgCtrl) running() {
	self.curTimes = 1
	self.sendFunc(self.msg, self.curTimes)

	timer := time.NewTimer(self.interval)
	for {
		select {
		case <-timer.C:
			if self.maxTimes != 0 && self.curTimes == self.maxTimes {
				return
			}
			self.curTimes++
			self.sendFunc(self.msg, self.curTimes)
			timer.Reset(self.interval)

		case <-self.closeCh:
			return
		}
	}
}
