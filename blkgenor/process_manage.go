// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"sync"

	"github.com/MatrixAINetwork/go-matrix/accounts/signhelper"
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/consensus/blkmanage"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/msgsend"
	"github.com/MatrixAINetwork/go-matrix/olconsensus"
	"github.com/MatrixAINetwork/go-matrix/reelection"
	"github.com/pkg/errors"
)

type ProcessManage struct {
	mu            sync.Mutex
	curChainState mc.ChainState
	processMap    map[uint64]*Process
	matrix        Backend
	hd            *msgsend.HD
	signHelper    *signhelper.SignHelper
	bc            *core.BlockChain
	txPool        *core.TxPoolManager //Y
	reElection    *reelection.ReElection
	olConsensus   *olconsensus.TopNodeService
	random        *baseinterface.Random
	manblk        *blkmanage.ManBlkManage
}

func NewProcessManage(matrix Backend) *ProcessManage {
	return &ProcessManage{
		curChainState: mc.ChainState{},
		processMap:    make(map[uint64]*Process),
		matrix:        matrix,
		hd:            matrix.HD(),
		signHelper:    matrix.SignHelper(),
		bc:            matrix.BlockChain(),
		txPool:        matrix.TxPool(),
		reElection:    matrix.ReElection(),
		olConsensus:   matrix.OLConsensus(),
		random:        matrix.Random(),
		manblk:        matrix.ManBlkDeal(),
	}
}

func (pm *ProcessManage) SetCurNumber(number uint64, superSeq uint64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	switch pm.curChainState.Cmp(superSeq, number) {
	case 1: //cur > input
		log.Debug(pm.logExtraInfo(), "SetCurNumber", "超级序号或高度低于当前值,不处理",
			"curNumber", pm.curChainState.CurNumber(), "number", number,
			"curSuperSeq", pm.curChainState.SuperSeq(), "superSeq", superSeq)
		return
	case -1: // cur < input
		if pm.curChainState.SuperSeq() < superSeq {
			log.Debug(pm.logExtraInfo(), "SetCurNumber", "超级区块序号变更,清空之前状态",
				"curSuperSeq", pm.curChainState.SuperSeq(), "superSeq", superSeq)
			pm.clearProcessMap()
		}
		pm.curChainState.Reset(superSeq, number)
		pm.fixProcessMap()
	}
}

func (pm *ProcessManage) GetCurNumber() uint64 {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	return pm.curChainState.CurNumber()
}

func (pm *ProcessManage) GetCurrentProcess() *Process {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	return pm.getProcess(pm.curChainState.CurNumber())
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
	if pm.curChainState.CurNumber() == 0 {
		return
	}

	if len(pm.processMap) == 0 {
		return
	}

	delKeys := make([]uint64, 0)
	for key, process := range pm.processMap {
		if key < pm.curChainState.CurNumber()-1 {
			process.Close()
			delKeys = append(delKeys, key)
		}
	}

	for _, delKey := range delKeys {
		delete(pm.processMap, delKey)
	}

	//log.INFO(pm.logExtraInfo(), "PM 结束修正map, process数量", len(pm.processMap))
}

func (pm *ProcessManage) clearProcessMap() {
	for _, process := range pm.processMap {
		process.Close()
	}
	pm.processMap = make(map[uint64]*Process)
}

func (pm *ProcessManage) isLegalNumber(number uint64) error {
	var minNumber uint64
	if pm.curChainState.CurNumber() < 1 {
		minNumber = 0
	} else {
		minNumber = pm.curChainState.CurNumber() - 1
	}

	if number < minNumber {
		return errors.Errorf("高度(%d) 过于小于当前高度 范围(%d)", number, pm.curChainState.CurNumber())
	}

	if number > pm.curChainState.CurNumber()+2 {
		return errors.Errorf("高度(%d) 过于大于当前高度 范围(%d)", number, pm.curChainState.CurNumber())
	}

	return nil
}

func (pm *ProcessManage) getProcess(number uint64) *Process {
	process, OK := pm.processMap[number]
	if OK == false {
		//log.INFO(pm.logExtraInfo(), "PM 创建process，高度", number)
		process = newProcess(number, pm)
		pm.processMap[number] = process
	}

	return process
}

func (pm *ProcessManage) logExtraInfo() string {
	return "区块生成"
}
