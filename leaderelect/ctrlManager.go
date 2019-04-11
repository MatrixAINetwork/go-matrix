// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
	"sync"
)

type ControllerManager struct {
	mu            sync.Mutex
	curChainState mc.ChainState
	ctrlMap       map[uint64]*controller
	matrix        Matrix
	logInfo       string
}

func NewControllerManager(matrix Matrix, logInfo string) *ControllerManager {
	return &ControllerManager{
		curChainState: mc.ChainState{},
		ctrlMap:       make(map[uint64]*controller),
		matrix:        matrix,
		logInfo:       logInfo,
	}
}

func (cm *ControllerManager) StartController(number uint64, superBlkSeq uint64, msg *startControllerMsg) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	switch cm.curChainState.Cmp(superBlkSeq, number) {
	case 1: //cur > input
		log.Debug(cm.logInfo, "处理start controller消息", "超级序号或高度低于当前值,不处理",
			"curNumber", cm.curChainState.CurNumber(), "number", number,
			"curSuperSeq", cm.curChainState.SuperSeq(), "superSeq", superBlkSeq)
		return
	case -1: // cur < input
		if cm.curChainState.SuperSeq() < superBlkSeq {
			log.Debug(cm.logInfo, "处理start controller消息", "超级区块序号变更,清空之前状态",
				"curSuperSeq", cm.curChainState.SuperSeq(), "superSeq", superBlkSeq)
			cm.clearCtrlMap()
		}
		cm.curChainState.Reset(superBlkSeq, number)
		cm.fixCtrlMap()
	}

	cm.getController(number).ReceiveMsg(msg)
}

func (cm *ControllerManager) ReceiveMsgByCur(msg interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.curChainState.CurNumber() <= 0 {
		return
	}
	ctrl := cm.getController(cm.curChainState.CurNumber())
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

func (cm *ControllerManager) clearCtrlMap() {
	if len(cm.ctrlMap) == 0 {
		return
	}
	for _, ctrl := range cm.ctrlMap {
		ctrl.Close()
	}
	cm.ctrlMap = make(map[uint64]*controller)
}

func (cm *ControllerManager) isLegalNumber(number uint64) error {
	if number < cm.curChainState.CurNumber() {
		return errors.Errorf("number(%d) is less than current number(%d)", number, cm.curChainState.CurNumber())
	}
	if number > cm.curChainState.CurNumber()+mangerCacheMax {
		return errors.Errorf("number(%d) is too big than current number(%d)", number, cm.curChainState.CurNumber())
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
