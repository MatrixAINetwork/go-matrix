// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"sync"

	"github.com/MatrixAINetwork/go-matrix/consensus/blkmanage"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"

	"github.com/MatrixAINetwork/go-matrix/accounts/signhelper"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/msgsend"
	"github.com/MatrixAINetwork/go-matrix/reelection"
	"github.com/pkg/errors"
)

type ProcessManage struct {
	mu             sync.Mutex
	curChainState  mc.ChainState
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
		curChainState:  mc.ChainState{},
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
	pm.checkVerifiedBlocksCache()
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

	delKeys := make([]uint64, 0)
	for key, process := range pm.processMap {
		if key < pm.curChainState.CurNumber() {
			process.Close()
			delKeys = append(delKeys, key)
		}
	}

	for _, delKey := range delKeys {
		delete(pm.processMap, delKey)
	}
}

func (pm *ProcessManage) clearProcessMap() {
	if len(pm.processMap) == 0 {
		return
	}
	for _, process := range pm.processMap {
		process.Close()
	}
	pm.processMap = make(map[uint64]*Process)
}

func (pm *ProcessManage) isLegalNumber(number uint64) error {
	if number < pm.curChainState.CurNumber() {
		return errors.Errorf("number(%d) is less than current number(%d)", number, pm.curChainState.CurNumber())
	}

	if number > pm.curChainState.CurNumber()+2 {
		return errors.Errorf("number(%d) is too big than current number(%d)", number, pm.curChainState.CurNumber())
	}

	return nil
}

func (pm *ProcessManage) getProcess(number uint64) *Process {
	process, OK := pm.processMap[number]
	if OK == false {
		process = newProcess(number, pm)
		pm.processMap[number] = process
	}
	return process
}

func (pm *ProcessManage) checkVerifiedBlocksCache() {
	if len(pm.verifiedBlocks) <= 0 {
		return
	}
	curProcess := pm.getProcess(pm.curChainState.CurNumber())
	del := make([]common.Hash, 0)
	for key, block := range pm.verifiedBlocks {
		number := block.req.Header.Number.Uint64()
		if number > pm.curChainState.CurNumber() {
			continue
		}

		if number == pm.curChainState.CurNumber() {
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
