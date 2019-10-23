// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenorV2

import (
	"strconv"
	"sync"

	"github.com/MatrixAINetwork/go-matrix/accounts/signhelper"
	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/olconsensus"
	"github.com/MatrixAINetwork/go-matrix/reelection"
)

type State uint16

const (
	StateIdle State = iota
	StateBlockBroadcast
	StateAIPick
	StateHeaderGen
	StatePOSWaiting
	StatePowCombine
	StateBlockInsert
	StateEnd
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "未运行状态"
	case StateBlockBroadcast:
		return "区块广播阶段"
	case StateAIPick:
		return "AI结果选取阶段"
	case StateHeaderGen:
		return "验证请求生成阶段"
	case StatePOSWaiting:
		return "POS等待阶段"
	case StatePowCombine:
		return "POW组合阶段"
	case StateBlockInsert:
		return "区块插入阶段"
	case StateEnd:
		return "完成状态"
	default:
		return "未知状态"
	}
}

type Process struct {
	mu                sync.Mutex
	curLeader         common.Address
	nextLeader        common.Address
	consensusTurn     mc.ConsensusTurnInfo
	parentHeader      *types.Header
	parentHash        common.Hash
	number            uint64
	role              common.RoleType
	state             State
	pm                *ProcessManage
	aiPool            *AIResultPool
	powPool           *PowPool
	basePowPool       *BasePowPool
	blockPool         *BlockPool
	broadcastRstCache map[common.Address]*bcBlockRspInfo
	insertBlockHash   []common.Hash
	FullBlockReqCache *common.ReuseMsgController
	msgSender         *common.ResendMsgCtrl
	bcInterval        *mc.BCIntervalInfo
}

func newProcess(number uint64, pm *ProcessManage) *Process {
	p := &Process{
		curLeader:         common.Address{},
		nextLeader:        common.Address{},
		consensusTurn:     mc.ConsensusTurnInfo{},
		parentHeader:      nil,
		parentHash:        common.Hash{},
		insertBlockHash:   make([]common.Hash, 0),
		number:            number,
		role:              common.RoleNil,
		state:             StateIdle,
		pm:                pm,
		aiPool:            NewAIResultPool(pm.logExtraInfo() + " AI结果池(" + strconv.Itoa(int(number)) + ")"),
		powPool:           NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(number))),
		basePowPool:       NewBasePowPool("算力检测池(高度)" + strconv.Itoa(int(number))),
		blockPool:         NewBlockPool(),
		broadcastRstCache: make(map[common.Address]*bcBlockRspInfo),
		FullBlockReqCache: common.NewReuseMsgController(3),
		msgSender:         nil,
	}

	return p
}

func (p *Process) StartRunning(role common.RoleType, bcInterval *mc.BCIntervalInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.role = role
	if p.bcInterval == nil {
		p.bcInterval = bcInterval
	}
	p.changeState(StateBlockBroadcast)
	p.startBcBlock()
}

func (p *Process) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = StateIdle
	p.curLeader = common.Address{}
	p.nextLeader = common.Address{}
	p.consensusTurn = mc.ConsensusTurnInfo{}
	p.parentHeader = nil
	p.parentHash = common.Hash{}
	p.bcInterval = nil
	p.closeMsgSender()
}

func (p *Process) ReInit() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.checkState(StateIdle) {
		return
	}
	p.state = StateBlockBroadcast
	p.curLeader = common.Address{}
	p.nextLeader = common.Address{}
	p.consensusTurn = mc.ConsensusTurnInfo{}
	p.parentHeader = nil
	p.parentHash = common.Hash{}
	p.closeMsgSender()
}

func (p *Process) ReInitNextLeader() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.nextLeader = common.Address{}
}

func (p *Process) SetCurLeader(leader common.Address, consensusTurn mc.ConsensusTurnInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.curLeader == leader && p.consensusTurn == consensusTurn {
		return
	}
	p.curLeader = leader
	p.consensusTurn = consensusTurn
	p.closeMsgSender()
	if p.checkState(StateIdle) {
		return
	}
	p.state = StateBlockBroadcast
	p.nextLeader = common.Address{}
	p.parentHeader = nil
	p.parentHash = common.Hash{}
	p.startBcBlock()
}

func (p *Process) SetNextLeader(preLeader common.Address, leader common.Address) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.nextLeader == leader {
		return
	}
	p.nextLeader = leader
	p.startBlockInsert(preLeader)
}

func (p *Process) AddInsertBlockMsg(blockInsert *mc.HD_BlockInsertNotify) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.processBlockInsertMsg(blockInsert)
}

func (p *Process) startBcBlock() {
	if p.checkState(StateBlockBroadcast) == false {
		log.Warn(p.logExtraInfo(), "准备向验证者和广播节点广播区块，状态错误", p.state.String(), "区块高度", p.number-1)
		return
	}

	switch p.role {
	case common.RoleBroadcast:
		p.processBcBlock()
		if p.bcInterval.IsBroadcastNumber(p.number) {
			p.state = StateAIPick
			p.startHeaderGen(nil)
		} else {
			log.Debug(p.logExtraInfo(), "区块广播阶段", "广播节点，当前不是广播区块", "直接进入POS结果等待阶段", p.number)
			p.state = StatePOSWaiting
			p.processPosWaiting()
		}

	case common.RoleValidator:
		if p.bcInterval.IsBroadcastNumber(p.number) {
			log.Debug(p.logExtraInfo(), "区块广播阶段", "广播区块，直接进入广播区块结果验证阶段")
			p.state = StatePOSWaiting
			p.processBCBlockVerify()
			return
		}

		if (p.curLeader == common.Address{}) {
			log.Warn(p.logExtraInfo(), "区块广播阶段", "当前leader为空，等待leader消息", "高度", p.number)
			return
		}

		if p.curLeader == ca.GetDepositAddress() && p.bcInterval.IsBroadcastNumber(p.number) == false {
			p.processBcBlock()
			p.startAIPick()
		} else {
			log.Debug(p.logExtraInfo(), "区块广播阶段", "不是当前leader或当前是广播区块", "直接进入POS结果等待阶段", p.number)
			p.state = StatePOSWaiting
			p.processPosWaiting()
		}

	default:
		log.Warn(p.logExtraInfo(), "区块广播阶段, 错误的身份", p.role.String(), "高度", p.number)
		return
	}
}

func (p *Process) processBcBlock() {
	parentBlock := p.blockChain().GetBlockByNumber(p.number - 1)
	if parentBlock == nil {
		log.Error(p.logExtraInfo(), "无法获取父区块", "高度", p.number)
		return
	}
	parentHash := parentBlock.Hash()
	if p.number != 1 {
		log.Debug(p.logExtraInfo(), "开始广播区块, 高度", p.number-1, "区块 hash", parentHash)
		// peer层广播
		p.eventMux().Post(core.NewMinedBlockEvent{Block: parentBlock})
		// 高层广播
		msg := &mc.HD_BlockInsertNotify{Header: parentBlock.Header()}
		p.pm.hd.SendNodeMsg(mc.HD_NewBlockInsert, msg, common.RoleValidator|common.RoleBroadcast, nil)
	}
	p.parentHeader = parentBlock.Header()
	p.parentHash = parentHash
}

func (p *Process) startHeaderGen(aiResult *mc.HD_V2_AIMiningRspMsg) {
	if p.checkState(StateAIPick) == false {
		log.Warn(p.logExtraInfo(), "准备开始生成验证请求，状态错误", p.state.String(), "高度", p.number)
		return
	}
	p.state = StateHeaderGen

	if p.bcInterval.IsBroadcastNumber(p.number) {
		log.Info(p.logExtraInfo(), "开始生成广播区块, 高度", p.number)
		err := p.processBroadcastBlockGen()
		if err != nil {
			log.Error(p.logExtraInfo(), "生成广播区块验证请求错误", err, "高度", p.number)
			return
		}
		// 广播区块生成完毕后结束
		p.state = StateEnd
	} else {
		log.Info(p.logExtraInfo(), "开始生成普通区块验证请求, 高度", p.number)
		err := p.processHeaderGen(aiResult)
		if err != nil {
			log.Error(p.logExtraInfo(), "生成普通区块验证请求错误", err, "高度", p.number)
			return
		}
		p.state = StatePOSWaiting
		p.processPosWaiting()
	}
}

func (p *Process) checkState(state State) bool {
	return p.state == state
}

func (p *Process) changeState(targetState State) {
	if p.state == targetState-1 {
		p.state = targetState
	}
}

func (p *Process) checkRepeatInsert(blockHash common.Hash) bool {
	for _, insertedHash := range p.insertBlockHash {
		if blockHash == insertedHash {
			return true
		}
	}

	return false
}

func (p *Process) saveInsertedBlockHash(blockHash common.Hash) {
	p.insertBlockHash = append(p.insertBlockHash, blockHash)
}

func (p *Process) closeMsgSender() {
	if p.msgSender == nil {
		return
	}
	p.msgSender.Close()
	p.msgSender = nil
}

func (p *Process) logExtraInfo() string {
	return p.pm.logExtraInfo()
}

func (p *Process) blockChain() *core.BlockChain { return p.pm.bc }

func (p *Process) txPool() *core.TxPoolManager { return p.pm.txPool } //Y

func (p *Process) signHelper() *signhelper.SignHelper { return p.pm.signHelper }

func (p *Process) eventMux() *event.TypeMux { return p.pm.matrix.EventMux() }

func (p *Process) reElection() *reelection.ReElection { return p.pm.reElection }

func (p *Process) topNode() *olconsensus.TopNodeService { return p.pm.olConsensus }

func (p *Process) backend() Backend { return p.pm.matrix }
