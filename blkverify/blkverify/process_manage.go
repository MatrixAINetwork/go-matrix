// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"sync"

	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/blkverify/votepool"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/msgsend"
	"github.com/matrix/go-matrix/olconsensus"
	"github.com/matrix/go-matrix/reelection"
	"github.com/pkg/errors"
)

type ProcessManage struct {
	mu         sync.Mutex
	curNumber  uint64
	processMap map[uint64]*Process
	votePool   *votepool.VotePool
	hd         *msgsend.HD
	signHelper *signhelper.SignHelper
	bc         *core.BlockChain
	txPool     *core.TxPoolManager //YYY
	reElection *reelection.ReElection
	topNode    *olconsensus.TopNodeService
	event      *event.TypeMux
}

func NewProcessManage(matrix Matrix) *ProcessManage {
	return &ProcessManage{
		curNumber:  0,
		processMap: make(map[uint64]*Process),
		votePool:   votepool.NewVotePool(common.RoleValidator, "区块验证服务票池"),
		hd:         matrix.HD(),
		signHelper: matrix.SignHelper(),
		bc:         matrix.BlockChain(),
		txPool:     matrix.TxPool(),
		reElection: matrix.ReElection(),
		topNode:    matrix.TopNode(),
		event:      matrix.EventMux(),
	}
}

func (pm *ProcessManage) SetCurNumber(number uint64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.curNumber = number
	pm.fixProcessMap()
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

func (pm *ProcessManage) AddVoteToPool(signHash common.Hash, sign common.Signature, fromAccount common.Address, height uint64) error {
	return pm.votePool.AddVote(signHash, sign, fromAccount, height, true)
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

func (pm *ProcessManage) logExtraInfo() string {
	return "区块验证服务"
}
