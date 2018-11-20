// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"sync"

	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/msgsend"
	"github.com/matrix/go-matrix/reelection"
	"github.com/pkg/errors"
)

type ProcessManage struct {
	mu         sync.Mutex
	curNumber  uint64
	processMap map[uint64]*Process
	matrix     Backend
	hd         *msgsend.HD
	signHelper *signhelper.SignHelper
	bc         *core.BlockChain
	txPool     *core.TxPoolManager //YYY
	reElection *reelection.ReElection
	engine     consensus.Engine
	dposEngine consensus.DPOSEngine
}

func NewProcessManage(matrix Backend) *ProcessManage {
	return &ProcessManage{
		curNumber:  0,
		processMap: make(map[uint64]*Process),
		matrix:     matrix,
		hd:         matrix.HD(),
		signHelper: matrix.SignHelper(),
		bc:         matrix.BlockChain(),
		txPool:     matrix.TxPool(),
		reElection: matrix.ReElection(),
		engine:     matrix.BlockChain().Engine(),
		dposEngine: matrix.BlockChain().DPOSEngine(),
	}
}

func (pm *ProcessManage) SetCurNumber(number uint64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.curNumber = number
	pm.fixProcessMap()
}

func (pm *ProcessManage) GetCurNumber() uint64 {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	return pm.curNumber
}

func (pm *ProcessManage) GetCurrentProcess() *Process {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	return pm.getProcess(pm.curNumber)
}

func (pm *ProcessManage) GetProcess(number uint64) (*Process, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if err := pm.isLegalNumber(number); err != nil {
		return nil, err
	}
	return pm.getProcess(number), nil
}

func (pm *ProcessManage) GetProcessAndPreProcess(number uint64) (*Process, *Process, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if err := pm.isLegalNumber(number); err != nil {
		return nil, nil, err
	}

	if number == 0 {
		return pm.getProcess(number), nil, nil
	} else {
		return pm.getProcess(number), pm.getProcess(number - 1), nil
	}
}

func (pm *ProcessManage) fixProcessMap() {
	if pm.curNumber == 0 {
		return
	}

	if len(pm.processMap) == 0 {
		return
	}

	log.INFO(pm.logExtraInfo(), "PM 开始修正map, process数量", len(pm.processMap), "修复高度", pm.curNumber)

	delKeys := make([]uint64, 0)
	for key, process := range pm.processMap {
		if key < pm.curNumber-1 {
			process.Close()
			delKeys = append(delKeys, key)
		}
	}

	for _, delKey := range delKeys {
		delete(pm.processMap, delKey)
	}

	log.INFO(pm.logExtraInfo(), "PM 结束修正map, process数量", len(pm.processMap))
}

func (pm *ProcessManage) isLegalNumber(number uint64) error {
	var minNumber uint64
	if pm.curNumber < 1 {
		minNumber = 0
	} else {
		minNumber = pm.curNumber - 1
	}

	if number < minNumber {
		return errors.Errorf("number(%d) is less than current number(%d)", number, pm.curNumber)
	}

	if number > pm.curNumber+2 {
		return errors.Errorf("number(%d) is too big than current number(%d)", number, pm.curNumber)
	}

	return nil
}

func (pm *ProcessManage) getProcess(number uint64) *Process {
	process, OK := pm.processMap[number]
	if OK == false {
		log.INFO(pm.logExtraInfo(), "PM 创建process，高度", number)
		process = newProcess(number, pm)
		pm.processMap[number] = process
	}

	return process
}

func (pm *ProcessManage) logExtraInfo() string {
	return "区块生成"
}
