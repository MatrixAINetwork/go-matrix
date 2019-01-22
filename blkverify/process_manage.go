// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"sync"

	"github.com/matrix/go-matrix/consensus/blkmanage"

	"github.com/matrix/go-matrix/baseinterface"

	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/msgsend"
	"github.com/matrix/go-matrix/reelection"
	"github.com/pkg/errors"
)

type ProcessManage struct {
	mu             sync.Mutex
	curNumber      uint64
	processMap     map[uint64]*Process
	hd             *msgsend.HD
	signHelper     *signhelper.SignHelper
	bc             *core.BlockChain
	txPool         *core.TxPoolManager //Y
	reElection     *reelection.ReElection
	event          *event.TypeMux
	random         *baseinterface.Random
	chainDB        mandb.Database
	verifiedBlocks map[common.Hash]*verifiedBlock
	manblk         *blkmanage.ManBlkManage
}

func NewProcessManage(matrix Matrix) *ProcessManage {
	return &ProcessManage{
		curNumber:      0,
		processMap:     make(map[uint64]*Process),
		hd:             matrix.HD(),
		signHelper:     matrix.SignHelper(),
		bc:             matrix.BlockChain(),
		txPool:         matrix.TxPool(),
		reElection:     matrix.ReElection(),
		event:          matrix.EventMux(),
		random:         matrix.Random(),
		chainDB:        matrix.ChainDb(),
		verifiedBlocks: make(map[common.Hash]*verifiedBlock),
		manblk:         matrix.ManBlkDeal(),
	}
}

func (pm *ProcessManage) AddVerifiedBlock(block *verifiedBlock) {
	if block == nil || block.req == nil {
		return
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.verifiedBlocks[block.hash] = block
}

func (pm *ProcessManage) SetCurNumber(number uint64, preSuperBlock bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.curNumber = number
	if preSuperBlock {
		pm.clearProcessMap()
	} else {
		pm.fixProcessMap()
	}
	pm.checkVerifiedBlocksCache()
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

/*func (pm *ProcessManage) GetProcessByRole(number uint64, role common.RoleType) (*Process, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if err := pm.isLegalNumber(number); err != nil {
		return nil, err
	}
	process := pm.getProcess(number)
	pRole := process.GetRole()
	if pRole != role {
		return nil, errors.Errorf("process.role(%s) != params.role(%s)", pRole.String(), role.String())
	}
	return process, nil
}*/

func (pm *ProcessManage) fixProcessMap() {
	if len(pm.processMap) == 0 {
		return
	}

	log.INFO(pm.logExtraInfo(), "PM 开始修正map, process数量", len(pm.processMap), "修复高度", pm.curNumber)

	delKeys := make([]uint64, 0)
	for key, process := range pm.processMap {
		if key < pm.curNumber {
			process.Close()
			delKeys = append(delKeys, key)
		}
	}

	for _, delKey := range delKeys {
		delete(pm.processMap, delKey)
	}

	log.INFO(pm.logExtraInfo(), "PM 结束修正map, process数量", len(pm.processMap))
}

func (pm *ProcessManage) clearProcessMap() {
	if pm.curNumber == 0 {
		return
	}

	if len(pm.processMap) == 0 {
		return
	}

	log.INFO(pm.logExtraInfo(), "超级区块：PM 开始删除map, process数量", len(pm.processMap), "修复高度", pm.curNumber)

	delKeys := make([]uint64, 0)
	for key, process := range pm.processMap {
		process.Close()
		delKeys = append(delKeys, key)
	}

	for _, delKey := range delKeys {
		delete(pm.processMap, delKey)
	}

	log.INFO(pm.logExtraInfo(), "超级区块：PM 结束删除map, process数量", len(pm.processMap))
}

func (pm *ProcessManage) isLegalNumber(number uint64) error {
	if number < pm.curNumber {
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

func (pm *ProcessManage) checkVerifiedBlocksCache() {
	if len(pm.verifiedBlocks) <= 0 {
		return
	}
	curProcess := pm.getProcess(pm.curNumber)
	del := make([]common.Hash, 0)
	for key, block := range pm.verifiedBlocks {
		number := block.req.Header.Number.Uint64()
		if number > pm.curNumber {
			continue
		}

		if number == pm.curNumber {
			curProcess.AddVerifiedBlock(block)
		}
		del = append(del, key)
	}

	for _, delKey := range del {
		delete(pm.verifiedBlocks, delKey)
	}
}

func (pm *ProcessManage) logExtraInfo() string {
	return "区块验证服务"
}
