// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"strconv"
	"sync"

	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/reelection"
)

type State uint16

const (
	StateIdle State = iota
	StateBlockBroadcast
	StateHeaderGen
	StateMinerResultVerify
	StateBlockInsert
	StateEnd
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "未运行状态"
	case StateBlockBroadcast:
		return "区块广播阶段"
	case StateHeaderGen:
		return "验证请求生成阶段"
	case StateMinerResultVerify:
		return "矿工结果验证阶段"
	case StateBlockInsert:
		return "区块插入阶段"
	case StateEnd:
		return "完成状态"
	default:
		return "未知状态"
	}
}

type Process struct {
	mu                 sync.Mutex
	curLeader          common.Address
	nextLeader         common.Address
	consensusTurn      uint32
	preBlockHash       common.Hash
	number             uint64
	role               common.RoleType
	state              State
	pm                 *ProcessManage
	powPool            *PowPool
	broadcastRstCache  []*mc.BlockData
	blockCache         *blockCache
	insertBlockHash    []common.Hash
	FullBlockReqCache  *common.ReuseMsgController
	consensusReqSender *common.ResendMsgCtrl
}

func newProcess(number uint64, pm *ProcessManage) *Process {
	p := &Process{
		curLeader:          common.Address{},
		nextLeader:         common.Address{},
		consensusTurn:      0,
		preBlockHash:       common.Hash{},
		insertBlockHash:    make([]common.Hash, 0),
		number:             number,
		role:               common.RoleNil,
		state:              StateIdle,
		pm:                 pm,
		powPool:            NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(number))),
		broadcastRstCache:  make([]*mc.BlockData, 0),
		blockCache:         newBlockCache(),
		FullBlockReqCache:  common.NewReuseMsgController(3),
		consensusReqSender: nil,
	}

	return p
}

func (p *Process) StartRunning(role common.RoleType) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.role = role
	p.changeState(StateBlockBroadcast)
	p.startBcBlock()
}

func (p *Process) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = StateIdle
	p.curLeader = common.Address{}
	p.nextLeader = common.Address{}
	p.consensusTurn = 0
	p.preBlockHash = common.Hash{}
	p.closeConsensusReqSender()
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
	p.consensusTurn = 0
	p.preBlockHash = common.Hash{}
	p.closeConsensusReqSender()
}

func (p *Process) ReInitNextLeader() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.nextLeader = common.Address{}
}

func (p *Process) SetCurLeader(leader common.Address, consensusTurn uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.curLeader == leader && p.consensusTurn == consensusTurn {
		return
	}
	p.curLeader = leader
	p.consensusTurn = consensusTurn
	p.closeConsensusReqSender()
	log.INFO(p.logExtraInfo(), "process设置当前leader成功", p.curLeader.Hex(), "高度", p.number)
	if p.checkState(StateIdle) {
		return
	}
	p.state = StateBlockBroadcast
	p.nextLeader = common.Address{}
	p.preBlockHash = common.Hash{}
	p.startBcBlock()
}

func (p *Process) SetNextLeader(leader common.Address) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.nextLeader == leader {
		return
	}
	p.nextLeader = leader
	log.INFO(p.logExtraInfo(), "process设置next leader成功", p.nextLeader.Hex(), "高度", p.number)
	if p.state < StateBlockInsert {
		return
	}
	p.processBlockInsert()
}

func (p *Process) AddInsertBlockInfo(blockInsert *mc.HD_BlockInsertNotify) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.startBlockInsert(blockInsert)
}
func (p *Process) startBlockInsert(blkInsertMsg *mc.HD_BlockInsertNotify) {
	blockHash := blkInsertMsg.Header.Hash()
	log.INFO(p.logExtraInfo(), "区块插入", "启动", "block hash", blockHash.TerminalString())

	if p.checkRepeatInsert(blockHash) {
		log.WARN(p.logExtraInfo(), "插入区块已处理", p.number, "block hash", blockHash.TerminalString())
		return
	}

	header := blkInsertMsg.Header
	if common.IsBroadcastNumber(p.number) {
		signAccount, _, err := crypto.VerifySignWithValidate(header.HashNoSignsAndNonce().Bytes(), header.Signatures[0].Bytes())
		if err != nil {
			log.ERROR(p.logExtraInfo(), "广播区块插入消息非法, 签名解析错误", err)
			return
		}

		if signAccount != header.Leader {
			log.ERROR(p.logExtraInfo(), "广播区块插入消息非法, 签名不匹配，签名人", signAccount.Hex(), "Leader", header.Leader.Hex())
			return
		}

		if role, _ := ca.GetAccountOriginalRole(signAccount, p.number-1); common.RoleBroadcast != role {
			log.ERROR(p.logExtraInfo(), "广播区块插入消息非法，签名人别是广播身份, role", role.String())
			return
		}
		log.Info(p.logExtraInfo(), "开始插入", "广播区块")
	} else {
		if err := p.dposEngine().VerifyBlock(p.blockChain(), header); err != nil {
			log.ERROR(p.logExtraInfo(), "区块插入消息DPOS共识失败", err)
			return
		}

		//todo 不是原始难度的结果，需要修改POW seal验证过程
		if err := p.engine().VerifySeal(p.blockChain(), header); err != nil {
			log.ERROR(p.logExtraInfo(), "区块插入消息POW验证失败", err)
			return
		}
		log.Info(p.logExtraInfo(), "开始插入", "普通区块")
	}

	if _, err := p.insertAndBcBlock(false, header); err != nil {
		log.INFO(p.logExtraInfo(), "区块插入失败, err", err, "fetch 高度", p.number, "fetch hash", blockHash.TerminalString())
		p.backend().FetcherNotify(blockHash, p.number)
	}

	p.saveInsertedBlockHash(blockHash)
}

func (p *Process) startBcBlock() {
	if p.checkState(StateBlockBroadcast) == false {
		log.WARN(p.logExtraInfo(), "准备向验证者和广播节点广播区块，状态错误", p.state.String(), "区块高度", p.number-1)
		return
	}

	if p.canBcBlock() == false {
		return
	}

	header := p.blockChain().GetHeaderByNumber(p.number - 1)
	if p.number != 1 { //todo 不好理解
		log.INFO(p.logExtraInfo(), "开始广播区块, 高度", p.number-1, "block hash", header.Hash())
		p.pm.hd.SendNodeMsg(mc.HD_NewBlockInsert, &mc.HD_BlockInsertNotify{Header: header}, common.RoleValidator|common.RoleBroadcast, nil)
	}
	p.preBlockHash = header.Hash()
	p.state = StateHeaderGen
	p.startHeaderGen()
}

func (p *Process) canBcBlock() bool {
	switch p.role {
	case common.RoleBroadcast:
		return true
	case common.RoleValidator:
		if (p.curLeader == common.Address{}) {
			log.WARN(p.logExtraInfo(), "广播区块阶段", "当前leader为空，等待leader消息", "高度", p.number)
			return false
		}
	default:
		log.ERROR(p.logExtraInfo(), "广播区块阶段, 错误的身份", p.role.String(), "高度", p.number)
		return false
	}
	return true
}

func (p *Process) startHeaderGen() {
	if p.checkState(StateHeaderGen) == false {
		log.WARN(p.logExtraInfo(), "准备开始生成验证请求，状态错误", p.state.String(), "高度", p.number)
		return
	}

	if p.canGenHeader() == false {
		return
	}

	log.INFO(p.logExtraInfo(), "开始生成验证请求, 高度", p.number)
	err := p.processHeaderGen()
	if err != nil {
		log.ERROR(p.logExtraInfo(), "生成验证请求错误", err, "高度", p.number)
		return
	}

	p.state = StateMinerResultVerify
	p.processMinerResultVerify(p.curLeader, true)
}

func (p *Process) canGenHeader() bool {
	switch p.role {
	case common.RoleBroadcast:
		if false == common.IsBroadcastNumber(p.number) {
			log.INFO(p.logExtraInfo(), "广播身份，当前不是广播区块，不生成区块", "直接进入挖矿结果验证阶段", "高度", p.number)
			p.state = StateMinerResultVerify
			p.processMinerResultVerify(p.curLeader, true)
			return false
		}
	case common.RoleValidator:
		if common.IsBroadcastNumber(p.number) {
			log.INFO(p.logExtraInfo(), "验证者身份，当前是广播区块，不生成区块", "直接进入挖矿结果验证阶段", "高度", p.number)
			p.state = StateMinerResultVerify
			p.processMinerResultVerify(p.curLeader, true)
			return false
		}

		if (p.curLeader == common.Address{}) {
			log.WARN(p.logExtraInfo(), "准备开始生成验证请求", "当前leader为空，等待leader消息", "高度", p.number)
			return false
		}

		if p.curLeader != ca.GetAddress() {
			log.INFO(p.logExtraInfo(), "自己不是当前leader，进入挖矿结果验证阶段, 高度", p.number, "self", ca.GetAddress().Hex(), "leader", p.curLeader.Hex())
			p.state = StateMinerResultVerify
			p.processMinerResultVerify(p.curLeader, true)
			return false
		}

	default:
		log.ERROR(p.logExtraInfo(), "错误的身份", p.role.String(), "高度", p.number)
		return false
	}

	return true
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

func (p *Process) logExtraInfo() string {
	return p.pm.logExtraInfo()
}

func (p *Process) blockChain() *core.BlockChain { return p.pm.bc }

func (p *Process) engine() consensus.Engine { return p.pm.engine }

func (p *Process) dposEngine() consensus.DPOSEngine { return p.pm.dposEngine }

func (p *Process) txPool() *core.TxPoolManager { return p.pm.txPool } //YYY

func (p *Process) signHelper() *signhelper.SignHelper { return p.pm.signHelper }

func (p *Process) eventMux() *event.TypeMux { return p.pm.matrix.EventMux() }

func (p *Process) reElection() *reelection.ReElection { return p.pm.reElection }

func (p *Process) backend() Backend { return p.pm.matrix }
