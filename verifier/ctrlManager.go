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
	"github.com/pkg/errors"
	"sync"
)

type ControllerManager struct {
	mu        sync.Mutex
	curNumber uint64
	ctrlMap   map[uint64]*controller
	matrix    Matrix
	logInfo   string
}

func NewControllerManager(matrix Matrix, logInfo string) *ControllerManager {
	return &ControllerManager{
		curNumber: 0,
		ctrlMap:   make(map[uint64]*controller),
		matrix:    matrix,
		logInfo:   logInfo,
	}
}

func (cm *ControllerManager) StartController(number uint64, msg *startControllerMsg) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.curNumber > number {
		log.INFO(cm.logInfo, "处理start controller消息", "高度低于当前高度,不处理", "curNumber", cm.curNumber, "number", number)
		return
	} else if cm.curNumber < number {
		cm.curNumber = number
		cm.fixCtrlMap()
	}
	cm.getController(number).ReceiveMsg(msg)
}

func (cm *ControllerManager) GetController(number uint64) (*controller, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if err := cm.isLegalNumber(number); err != nil {
		return nil, err
	}
	return cm.getController(number), nil
}

func (cm *ControllerManager) GetCurController() *controller {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.getController(cm.curNumber)
}

func (cm *ControllerManager) fixCtrlMap() {
	if len(cm.ctrlMap) == 0 {
		return
	}

	log.INFO(cm.logInfo, "ctrlManager 开始修正map, controller数量", len(cm.ctrlMap), "修复高度", cm.curNumber)

	delKeys := make([]uint64, 0)
	for key, ctrl := range cm.ctrlMap {
		if key < cm.curNumber {
			ctrl.Close()
			delKeys = append(delKeys, key)
		}
	}

	for _, delKey := range delKeys {
		delete(cm.ctrlMap, delKey)
	}

	log.INFO(cm.logInfo, "PM 结束修正map, controller数量", len(cm.ctrlMap))
}

func (cm *ControllerManager) isLegalNumber(number uint64) error {
	if number < cm.curNumber {
		return errors.Errorf("number(%d) is less than current number(%d)", number, cm.curNumber)
	}
	if number > cm.curNumber+2 {
		return errors.Errorf("number(%d) is too big than current number(%d)", number, cm.curNumber)
	}
	return nil
}

func (cm *ControllerManager) getController(number uint64) *controller {
	ctrl, OK := cm.ctrlMap[number]
	if OK == false {
		log.INFO(cm.logInfo, "ctrlManager 创建controller，高度", number)
		ctrl = newController(cm.matrix, cm.logInfo, number)
		cm.ctrlMap[number] = ctrl
	}
	return ctrl
}
