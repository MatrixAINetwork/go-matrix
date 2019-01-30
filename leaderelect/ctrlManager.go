// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
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

func (cm *ControllerManager) ClearController() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.curNumber = 0
	for _, ctrl := range cm.ctrlMap {
		ctrl.Close()
	}
	cm.ctrlMap = make(map[uint64]*controller)
}

func (cm *ControllerManager) StartController(number uint64, msg *startControllerMsg) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.curNumber > number {
		log.Debug(cm.logInfo, "处理start controller消息", "高度低于当前高度,不处理", "curNumber", cm.curNumber, "number", number)
		return
	} else if cm.curNumber < number {
		cm.curNumber = number
		cm.fixCtrlMap()
	}
	cm.getController(number).ReceiveMsg(msg)
}

func (cm *ControllerManager) ReceiveMsgByCur(msg interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.curNumber <= 0 {
		return
	}
	ctrl := cm.getController(cm.curNumber)
	ctrl.ReceiveMsg(msg)
}

func (cm *ControllerManager) ReceiveMsg(number uint64, msg interface{}) error {
	if number <= 0 {
		return errors.New("number(0) is illegal")
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if err := cm.isLegalNumber(number); err != nil {
		return err
	}
	ctrl := cm.getController(number)
	ctrl.ReceiveMsg(msg)
	return nil
}

func (cm *ControllerManager) fixCtrlMap() {
	if len(cm.ctrlMap) == 0 {
		return
	}
	delKeys := make([]uint64, 0)
	for key, ctrl := range cm.ctrlMap {
		if err := cm.isLegalNumber(key); err != nil {
			ctrl.Close()
			delKeys = append(delKeys, key)
		}
	}
	for _, delKey := range delKeys {
		delete(cm.ctrlMap, delKey)
	}
}

func (cm *ControllerManager) isLegalNumber(number uint64) error {
	if number < cm.curNumber {
		return errors.Errorf("number(%d) is less than current number(%d)", number, cm.curNumber)
	}
	if number > cm.curNumber+mangerCacheMax {
		return errors.Errorf("number(%d) is too big than current number(%d)", number, cm.curNumber)
	}
	return nil
}

func (cm *ControllerManager) getController(number uint64) *controller {
	ctrl, OK := cm.ctrlMap[number]
	if OK == false {
		ctrl = newController(cm.matrix, cm.logInfo, number)
		cm.ctrlMap[number] = ctrl
	}
	return ctrl
}
