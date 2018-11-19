// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

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
