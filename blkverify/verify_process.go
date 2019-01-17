// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"sync"
	"time"

	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/matrixwork"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/reelection"
	"github.com/pkg/errors"
)

type State uint16

const (
	StateIdle State = iota
	StateStart
	StateReqVerify
	StateTxsVerify
	StateDPOSVerify
	StateEnd
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "未运行状态"
	case StateStart:
		return "开始状态"
	case StateReqVerify:
		return "请求验证阶段"
	case StateTxsVerify:
		return "交易验证阶段"
	case StateDPOSVerify:
		return "DPOS共识阶段"
	case StateEnd:
		return "完成状态"
	default:
		return "未知状态"
	}
}

type verifyResult uint8

const (
	localVerifyResultProcessing verifyResult = iota
	localVerifyResultSuccess
	localVerifyResultFailedButCanRecover
	localVerifyResultStateFailed
	localVerifyResultDBRecovery
)

func (vr verifyResult) String() string {
	switch vr {
	case localVerifyResultProcessing:
		return "未完成验证"
	case localVerifyResultSuccess:
		return "本地验证成功"
	case localVerifyResultFailedButCanRecover:
		return "验证失败(可恢复)"
	case localVerifyResultStateFailed:
		return "验证失败"
	case localVerifyResultDBRecovery:
		return "DB中的请求，之前验证过"
	default:
		return "未知状态"
	}
}

var (
	ErrParamIsNil = errors.New("param is nil")
	ErrExistVote  = errors.New("vote is existed")
)

type Process struct {
	mu               sync.Mutex
	leaderCache      mc.LeaderChangeNotify
	number           uint64
	role             common.RoleType
	state            State
	curProcessReq    *reqData
	reqCache         *reqCache
	unverifiedVotes  *unverifiedVotePool
	pm               *ProcessManage
	txsAcquireSeq    int
	voteMsgSender    *common.ResendMsgCtrl
	mineReqMsgSender *common.ResendMsgCtrl
	posedReqSender   *common.ResendMsgCtrl
}

func newProcess(number uint64, pm *ProcessManage) *Process {
	p := &Process{
		leaderCache: mc.LeaderChangeNotify{
			ConsensusState: false,
			Leader:         common.Address{},
			NextLeader:     common.Address{},
			Number:         number,
			ConsensusTurn:  mc.ConsensusTurnInfo{},
			ReelectTurn:    0,
			TurnBeginTime:  0,
			TurnEndTime:    0,
		},
		number:           number,
		role:             common.RoleNil,
		state:            StateIdle,
		curProcessReq:    nil,
		reqCache:         newReqCache(),
		unverifiedVotes:  newUnverifiedVotePool(pm.logExtraInfo()),
		pm:               pm,
		txsAcquireSeq:    0,
		voteMsgSender:    nil,
		mineReqMsgSender: nil,
		posedReqSender:   nil,
	}

	return p
}

func (p *Process) StartRunning(role common.RoleType) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.role = role
	p.changeState(StateStart)

	if p.role == common.RoleBroadcast {
		p.startReqVerifyBC()
	} else if p.role == common.RoleValidator {
		p.startReqVerifyCommon()
	}
}

func (p *Process) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = StateIdle
	p.curProcessReq = nil
	p.stopSender()
}

func (p *Process) SetLeaderInfo(info *mc.LeaderChangeNotify) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.leaderCache.ConsensusState = info.ConsensusState
	if p.leaderCache.ConsensusState == false {
		p.stopProcess()
		return
	}

	if p.leaderCache.Leader == info.Leader && p.leaderCache.ConsensusTurn == info.ConsensusTurn {
		//已处理过的leader消息，不处理
		return
	}

	//leader或轮次变化了，更新缓存
	p.leaderCache.Leader.Set(info.Leader)
	p.leaderCache.NextLeader.Set(info.NextLeader)
	p.leaderCache.ConsensusTurn = info.ConsensusTurn
	p.leaderCache.ReelectTurn = info.ReelectTurn
	p.leaderCache.TurnBeginTime = info.TurnBeginTime
	p.leaderCache.TurnEndTime = info.TurnEndTime
	p.curProcessReq = nil

	//维护req缓存
	p.reqCache.SetCurTurn(p.leaderCache.ConsensusTurn)

	//重启process
	p.stopSender()
	if p.state > StateIdle {
		p.state = StateStart
		if p.role == common.RoleValidator {
			p.startReqVerifyCommon()
		} else if p.role == common.RoleBroadcast {
			log.WARN(p.logExtraInfo(), "广播身份下收到leader变更消息", "不处理")
		}
	}
}

func (p *Process) stopProcess() {
	p.closeMineReqMsgSender()
	p.leaderCache.Leader.Set(common.Address{})
	p.leaderCache.NextLeader.Set(common.Address{})
	p.leaderCache.ConsensusTurn = mc.ConsensusTurnInfo{}
	p.leaderCache.ReelectTurn = 0
	p.leaderCache.TurnBeginTime = 0
	p.leaderCache.TurnEndTime = 0
	p.curProcessReq = nil

	if p.state > StateIdle {
		p.state = StateStart
	}
}

func (p *Process) ProcessRecoveryMsg(msg *mc.RecoveryStateMsg) {
	p.mu.Lock()
	defer p.mu.Unlock()
	msgHeaderHash := msg.Header.HashNoSignsAndNonce()
	reqData, err := p.reqCache.GetLeaderReqByHash(msgHeaderHash)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "处理状态恢复消息", "本地请求获取失败", "err", err)
		return
	}
	if reqData.hash != msgHeaderHash {
		log.ERROR(p.logExtraInfo(), "处理状态恢复消息", "本地请求hash不匹配，忽略消息",
			"本地hash", reqData.hash.TerminalString(), "消息hash", msgHeaderHash.TerminalString())
		return
	}

	log.INFO(p.logExtraInfo(), "处理状态恢复消息", "开始重置POS投票")
	reqData.clearVotes()
	parentHash := reqData.req.Header.ParentHash
	//添加投票
	for _, sign := range msg.Header.Signatures {
		verifiedVote, err := p.verifyVote(reqData.hash, sign, common.Address{}, parentHash, false)
		if err != nil {
			log.Info(p.logExtraInfo(), "处理状态恢复消息", "签名验证失败", "err", err)
			continue
		}
		reqData.addVote(verifiedVote)
	}

	if p.role == common.RoleBroadcast {
		p.startReqVerifyBC()
	} else if p.role == common.RoleValidator {
		p.startReqVerifyCommon()
	}

	log.INFO(p.logExtraInfo(), "处理状态恢复消息", "完成")
}

func (p *Process) AddVerifiedBlock(block *verifiedBlock) {
	if block.req == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	log.Info(p.logExtraInfo(), "AddVerifiedBlock", "开始处理", "block hash", block.hash.TerminalString(), "leader", block.req.Header.Leader.Hex(), "高度", p.number)
	reqData, err := p.reqCache.AddReq(block.req, true)
	if err != nil {
		log.Info(p.logExtraInfo(), "AddVerifiedBlock", "添加缓存失败", "err", err, "高度", p.number)
		return
	}
	reqData.originalTxs = block.txs
	return
}

func (p *Process) AddReq(reqMsg *mc.HD_BlkConsensusReqMsg) {
	p.mu.Lock()
	defer p.mu.Unlock()

	reqData, err := p.reqCache.AddReq(reqMsg, false)
	if err != nil {
		//log.Trace(p.logExtraInfo(), "请求添加缓存失败", err, "from", reqMsg.From, "高度", p.number)
		return
	}
	log.INFO(p.logExtraInfo(), "区块共识请求处理", "请求添加缓存成功", "from", reqMsg.From.Hex(), "高度", p.number, "reqHash", reqData.hash.TerminalString())

	if p.role == common.RoleBroadcast {
		p.startReqVerifyBC()
	} else if p.role == common.RoleValidator {
		p.startReqVerifyCommon()
	}
}

func (p *Process) AddLocalReq(localReq *mc.LocalBlockVerifyConsensusReq) {
	p.mu.Lock()
	defer p.mu.Unlock()

	leader := localReq.BlkVerifyConsensusReq.Header.Leader
	reqData, err := p.reqCache.AddLocalReq(localReq)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "本地请求添加缓存失败", err, "高度", p.number, "leader", leader.Hex())
		return
	}
	log.INFO(p.logExtraInfo(), "本地请求添加成功, 高度", p.number, "leader", leader.Hex())
	// 添加早于请求达到的投票
	parentHash := reqData.req.Header.ParentHash
	votes := p.unverifiedVotes.GetVotes(reqData.hash)
	for _, vote := range votes {
		if vote.signHash != reqData.hash {
			log.Info(p.logExtraInfo(), "区块共识请求(本地)处理", "添加早的投票, signHash不匹配")
			continue
		}
		verifiedVote, err := p.verifyVote(reqData.hash, vote.sign, vote.from, parentHash, true)
		if err != nil {
			log.Info(p.logExtraInfo(), "区块共识请求(本地)处理", "添加早的投票, 签名验证失败", "err", err, "from", vote.from.Hex(), "reqHash", reqData.hash.TerminalString())
			continue
		}
		reqData.addVote(verifiedVote)
	}

	if p.role == common.RoleBroadcast {
		p.startReqVerifyBC()
	} else if p.role == common.RoleValidator {
		p.startReqVerifyCommon()
	}
}

func (p *Process) HandleVote(signHash common.Hash, vote common.Signature, from common.Address) {
	if (signHash == common.Hash{}) || (vote == common.Signature{}) || (from == common.Address{}) {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 签名不是当前处理的请求
	if p.curProcessReq == nil || p.curProcessReq.hash != signHash {
		// 将投票存入未验证票池中
		p.unverifiedVotes.AddVote(signHash, vote, from)
		return
	}

	if p.curProcessReq.isAccountExistVote(from) {
		log.Trace(p.logExtraInfo(), "处理投票消息", "已存在的投票", "from", from.Hex())
		return
	}

	verifiedVote, err := p.verifyVote(signHash, vote, from, p.curProcessReq.req.Header.ParentHash, true)
	if err != nil {
		log.Info(p.logExtraInfo(), "处理投票消息", "签名验证失败", "err", err)
		return
	}

	p.curProcessReq.addVote(verifiedVote)
	p.processDPOSOnce()
}

func (p *Process) ProcessDPOSOnce() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.processDPOSOnce()
}

func (p *Process) startReqVerifyCommon() {
	if p.checkState(StateStart) == false {
		log.WARN(p.logExtraInfo(), "准备开始请求验证阶段，状态错误", p.state.String(), "高度", p.number)
		return
	}

	if p.leaderCache.ConsensusState == false {
		log.WARN(p.logExtraInfo(), "请求验证阶段", "当前leader未共识完成，等待leader消息", "高度", p.number)
		return
	}

	req, err := p.reqCache.GetLeaderReq(p.leaderCache.Leader, p.leaderCache.ConsensusTurn)
	if err != nil {
		log.WARN(p.logExtraInfo(), "请求验证阶段,寻找leader的请求错误,继续等待请求", err,
			"Leader", p.leaderCache.Leader.Hex(), "轮次", p.leaderCache.ConsensusTurn, "高度", p.number)
		return
	}

	p.curProcessReq = req
	log.INFO(p.logExtraInfo(), "请求验证阶段", "开始", "高度", p.number, "HeaderHash", p.curProcessReq.hash.TerminalString(), "parent hash", p.curProcessReq.req.Header.ParentHash.TerminalString(), "之前状态", p.state.String())
	p.state = StateReqVerify
	p.processReqOnce()
}

func (p *Process) processReqOnce() {
	if p.checkState(StateReqVerify) == false {
		return
	}

	// if is local req, skip local verify step
	if p.curProcessReq.localReq {
		log.INFO(p.logExtraInfo(), "共识请求验证", "", "请求为本地请求", "跳过验证阶段", "高度", p.number)
		p.startDPOSVerify(localVerifyResultSuccess)
		return
	}

	if p.curProcessReq.localVerifyResult != localVerifyResultProcessing && p.curProcessReq.localVerifyResult != localVerifyResultDBRecovery {
		log.Info(p.logExtraInfo(), "共识请求验证", "", "请求已验证过", "投票并进入POS共识阶段", "验证结果", p.curProcessReq.localVerifyResult, "高度", p.number, "请求leader", p.curProcessReq.req.Header.Leader.Hex(), "请求hash", p.curProcessReq.hash.TerminalString())
		p.startDPOSVerify(p.curProcessReq.localVerifyResult)
		return
	}

	if p.curProcessReq.localVerifyResult != localVerifyResultDBRecovery {
		// verify timestamp (对从DB中恢复的请求，不验证时间戳)
		headerTime := p.curProcessReq.req.Header.Time.Int64()
		if headerTime < p.leaderCache.TurnBeginTime || headerTime > p.leaderCache.TurnEndTime {
			log.ERROR(p.logExtraInfo(), "验证请求头时间戳", "时间戳不合法", "头时间", headerTime,
				"轮次开始时间", p.leaderCache.TurnBeginTime, "轮次结束时间", p.leaderCache.TurnEndTime,
				"轮次", p.leaderCache.ConsensusTurn, "高度", p.number)
			p.startDPOSVerify(localVerifyResultStateFailed)
			return
		}
	}

	// verify header
	if err := p.blockChain().VerifyHeader(p.curProcessReq.req.Header); err != nil {
		log.ERROR(p.logExtraInfo(), "预验证头信息失败", err, "高度", p.number)
		p.startDPOSVerify(localVerifyResultStateFailed)
		return
	}

	// verify net topology info
	if err := p.verifyNetTopology(p.curProcessReq.req.Header, p.curProcessReq.req.OnlineConsensusResults); err != nil {
		log.ERROR(p.logExtraInfo(), "验证拓扑信息失败", err, "高度", p.number)
		p.startDPOSVerify(localVerifyResultFailedButCanRecover)
		return
	}

	//verify Version
	if err := p.blockChain().DPOSEngine().VerifyVersion(p.blockChain(), p.curProcessReq.req.Header); err != nil {
		log.ERROR(p.logExtraInfo(), "验证版本号失败", err, "高度", p.number)
		p.startDPOSVerify(localVerifyResultFailedButCanRecover)
		return
	}

	//verify vrf
	if err := p.verifyVrf(p.curProcessReq.req.Header); err != nil {
		log.Error(p.logExtraInfo(), "验证vrf失败", err, "高度", p.number)
		p.startDPOSVerify(localVerifyResultFailedButCanRecover)
		return
	}
	log.INFO(p.logExtraInfo(), "验证vrf成功 高度", p.number)

	p.startTxsVerify()
}

func (p *Process) startTxsVerify() {
	if p.checkState(StateReqVerify) == false {
		return
	}
	log.INFO(p.logExtraInfo(), "交易获取", "开始", "当前身份", p.role.String(), "高度", p.number)

	p.changeState(StateTxsVerify)

	txsCodeSize := p.curProcessReq.req.TxsCodeCount()
	if txsCodeSize == 0 || txsCodeSize == len(p.curProcessReq.originalTxs) {
		// 交易为空，或者已经得到全交易，跳过交易获取
		log.Info(p.logExtraInfo(), "无需获取交易", "直接进入交易及状态验证", "txsCode size", txsCodeSize, "txs size", len(p.curProcessReq.originalTxs))
		p.verifyTxsAndState()
	} else {
		// 开启交易获
		p.txsAcquireSeq++
		leader := p.curProcessReq.req.Header.Leader
		log.INFO(p.logExtraInfo(), "开始交易获取,seq", p.txsAcquireSeq, "数量", p.curProcessReq.req.TxsCodeCount(), "leader", leader.Hex(), "高度", p.number)
		txAcquireCh := make(chan *core.RetChan, 1)
		go p.txPool().ReturnAllTxsByN(p.curProcessReq.req.TxsCode, p.txsAcquireSeq, leader, txAcquireCh)
		go p.processTxsAcquire(txAcquireCh, p.txsAcquireSeq)
	}
}

func (p *Process) processTxsAcquire(txsAcquireCh <-chan *core.RetChan, seq int) {
	log.INFO(p.logExtraInfo(), "交易获取协程", "启动", "当前身份", p.role.String(), "高度", p.number)
	defer log.INFO(p.logExtraInfo(), "交易获取协程", "退出", "当前身份", p.role.String(), "高度", p.number)

	outTime := time.NewTimer(time.Second * 5)
	select {
	case txsResult := <-txsAcquireCh:
		go p.StartVerifyTxsAndState(txsResult)

	case <-outTime.C:
		log.INFO(p.logExtraInfo(), "交易获取协程", "获取交易超时", "高度", p.number, "seq", seq)
		go p.ProcessTxsAcquireTimeOut(seq)
		return
	}
}

func (p *Process) ProcessTxsAcquireTimeOut(seq int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.INFO(p.logExtraInfo(), "交易获取超时处理", "开始", "高度", p.number, "seq", seq, "cur seq", p.txsAcquireSeq)
	defer log.INFO(p.logExtraInfo(), "交易获取超时处理", "结束", "高度", p.number, "seq", seq)

	if seq != p.txsAcquireSeq {
		log.WARN(p.logExtraInfo(), "交易获取超时处理", "Seq不匹配，忽略", "高度", p.number, "seq", seq, "cur seq", p.txsAcquireSeq)
		return
	}

	if p.checkState(StateTxsVerify) == false {
		log.INFO(p.logExtraInfo(), "交易获取超时处理", "状态不正确，不处理", "高度", p.number, "seq", seq)
		return
	}

	p.startDPOSVerify(localVerifyResultFailedButCanRecover)
}

func (p *Process) StartVerifyTxsAndState(result *core.RetChan) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.checkState(StateTxsVerify) == false {
		return
	}

	log.INFO(p.logExtraInfo(), "交易验证，交易数据 result.seq", result.Resqe, "当前 reqSeq", p.txsAcquireSeq, "高度", p.number)
	if result.Resqe != p.txsAcquireSeq {
		log.WARN(p.logExtraInfo(), "交易验证", "seq不匹配，跳过", "高度", p.number)
		return
	}

	if result.Err != nil {
		log.ERROR(p.logExtraInfo(), "交易验证，交易数据错误", result.Err, "高度", p.number)
		p.startDPOSVerify(localVerifyResultFailedButCanRecover)
		return
	}
	for _, listN := range result.AllTxs {
		p.curProcessReq.originalTxs = append(p.curProcessReq.originalTxs, listN.Txser...)
	}

	p.verifyTxsAndState()
}

func (p *Process) verifyTxsAndState() {
	log.INFO(p.logExtraInfo(), "开始交易验证, 数量", len(p.curProcessReq.originalTxs), "高度", p.number)

	//跑交易交易验证， Root TxHash ReceiptHash Bloom GasLimit GasUsed
	remoteHeader := p.curProcessReq.req.Header
	localHeader := types.CopyHeader(remoteHeader)
	localHeader.GasUsed = 0

	work, err := matrixwork.NewWork(p.blockChain().Config(), p.blockChain(), nil, localHeader, p.pm.random)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "交易验证，创建work失败!", err, "高度", p.number)
		p.startDPOSVerify(localVerifyResultFailedButCanRecover)
		return
	}

	// process state version
	err = p.blockChain().ProcessStateVersion(p.curProcessReq.req.Header.Version, work.State)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "状态树验证,错误", "运行状态树版本更新失败", "err", err)
		p.startDPOSVerify(localVerifyResultStateFailed)
		return
	}

	uptimeMap, err := p.blockChain().ProcessUpTime(work.State, localHeader)
	if err != nil {
		log.Error(p.logExtraInfo(), "uptime处理错误", err)
		return
	}
	err = work.ConsensusTransactions(p.pm.event, p.curProcessReq.originalTxs, uptimeMap)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "交易验证，共识执行交易出错!", err, "高度", p.number)
		p.startDPOSVerify(localVerifyResultStateFailed)
		return
	}
	finalTxs := work.GetTxs()
	localBlock := types.NewBlock(localHeader, finalTxs, nil, work.Receipts)
	// process matrix state
	err = p.blockChain().ProcessMatrixState(localBlock, work.State)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "matrix状态验证,错误", "运行matrix状态出错", "err", err)
		p.startDPOSVerify(localVerifyResultStateFailed)
		return
	}

	// 运行完matrix state后，生成root
	localBlock, err = p.blockChain().Engine().Finalize(p.blockChain(), localHeader, work.State, finalTxs, nil, work.Receipts)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "matrix状态验证,错误", "Failed to finalize block for sealing", "err", err)
		p.startDPOSVerify(localVerifyResultStateFailed)
		return
	}

	log.Info(p.logExtraInfo(), "共识后的交易本地hash", localBlock.TxHash(), "共识后的交易远程hash", remoteHeader.TxHash)
	log.Info("miss tree node debug", "finalize root", localBlock.Root().Hex(), "remote root", remoteHeader.Root.Hex())

	// verify election info
	if err := p.verifyElection(p.curProcessReq.req.Header, work.State); err != nil {
		log.ERROR(p.logExtraInfo(), "验证选举信息失败", err, "高度", p.number)
		p.startDPOSVerify(localVerifyResultStateFailed)
		return
	}

	//localBlock check
	localHeader = localBlock.Header()
	localHash := localHeader.HashNoSignsAndNonce()

	if localHash != p.curProcessReq.hash {
		log.ERROR(p.logExtraInfo(), "交易验证及状态，错误", "block hash不匹配",
			"local hash", localHash.TerminalString(), "remote hash", p.curProcessReq.hash.TerminalString(),
			"local root", localHeader.Root.TerminalString(), "remote root", remoteHeader.Root.TerminalString(),
			"local txHash", localHeader.TxHash.TerminalString(), "remote txHash", remoteHeader.TxHash.TerminalString(),
			"local ReceiptHash", localHeader.ReceiptHash.TerminalString(), "remote ReceiptHash", remoteHeader.ReceiptHash.TerminalString(),
			"local Bloom", localHeader.Bloom.Big(), "remote Bloom", remoteHeader.Bloom.Big(),
			"local GasLimit", localHeader.GasLimit, "remote GasLimit", remoteHeader.GasLimit,
			"local GasUsed", localHeader.GasUsed, "remote GasUsed", remoteHeader.GasUsed)
		p.startDPOSVerify(localVerifyResultStateFailed)
		return
	}

	p.curProcessReq.receipts = work.Receipts
	p.curProcessReq.stateDB = work.State
	p.curProcessReq.finalTxs = finalTxs
	// 开始DPOS共识验证
	p.startDPOSVerify(localVerifyResultSuccess)
}

func (p *Process) sendVote(validate bool) {
	signHash := p.curProcessReq.hash
	sign, err := p.signHelper().SignHashWithValidate(signHash.Bytes(), validate, p.curProcessReq.req.Header.ParentHash)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "投票签名失败", err, "高度", p.number)
		return
	}

	p.startVoteMsgSender(&mc.HD_ConsensusVote{SignHash: signHash, Sign: sign, Number: p.number})

	//将自己的投票加入票池 todo 股权
	p.curProcessReq.addVote(&common.VerifiedSign{
		Sign:     sign,
		Account:  ca.GetAddress(),
		Validate: true,
		Stock:    0,
	})

	// notify block genor server the result
	result := mc.BlockLocalVerifyOK{
		Header:      p.curProcessReq.req.Header,
		BlockHash:   p.curProcessReq.hash,
		OriginalTxs: p.curProcessReq.originalTxs,
		FinalTxs:    p.curProcessReq.finalTxs,
		Receipts:    p.curProcessReq.receipts,
		State:       p.curProcessReq.stateDB,
	}
	//log.INFO(p.logExtraInfo(), "发出区块共识结果消息", result, "高度", p.number)
	mc.PublishEvent(mc.BlkVerify_VerifyConsensusOK, &result)
}

func (p *Process) startDPOSVerify(lvResult verifyResult) {
	if p.state >= StateDPOSVerify {
		return
	}

	if p.role == common.RoleBroadcast {
		//广播节点，跳过DPOS投票验证阶段
		p.bcFinishedProcess(lvResult)
		return
	}

	log.INFO(p.logExtraInfo(), "开始POS阶段,验证结果", lvResult.String(), "高度", p.number)

	if lvResult == localVerifyResultSuccess {
		p.sendVote(true)
		// 验证成功的请求，做持久化缓存
		log.Info(p.logExtraInfo(), "区块持久化", "开始缓存")
		if err := saveVerifiedBlockToDB(p.ChainDb(), p.curProcessReq.hash, p.curProcessReq.req, p.curProcessReq.originalTxs); err != nil {
			log.Error(p.logExtraInfo(), "验证成功的区块持久化缓存失败", err)
		}
	}
	p.curProcessReq.localVerifyResult = lvResult

	// 添加早于请求达到的投票
	parentHash := p.curProcessReq.req.Header.ParentHash
	votes := p.unverifiedVotes.GetVotes(p.curProcessReq.hash)
	for _, vote := range votes {
		if vote.signHash != p.curProcessReq.hash {
			log.Info(p.logExtraInfo(), "开始POS阶段", "添加早的投票, signHash不匹配")
			continue
		}
		verifiedVote, err := p.verifyVote(p.curProcessReq.hash, vote.sign, vote.from, parentHash, true)
		if err != nil {
			log.Info(p.logExtraInfo(), "开始POS阶段", "添加早的投票, 签名验证失败", "err", err, "from", vote.from.Hex(), "reqHash", p.curProcessReq.hash.TerminalString())
			continue
		}
		p.curProcessReq.addVote(verifiedVote)
	}

	p.state = StateDPOSVerify
	p.processDPOSOnce()
}

func (p *Process) processDPOSOnce() {
	if p.checkState(StateDPOSVerify) == false {
		return
	}

	if p.curProcessReq.req == nil {
		return
	}

	if p.curProcessReq.posFinished {
		log.Trace(p.logExtraInfo(), "POS验证处理", "已完成，不重复处理")
		return
	}

	signs := p.curProcessReq.getVotes()
	log.INFO(p.logExtraInfo(), "POS验证处理", "执行POS", "投票数量", len(signs), "hash", p.curProcessReq.hash.TerminalString(), "高度", p.number)
	rightSigns, err := p.blockChain().DPOSEngine().VerifyHashWithVerifiedSignsAndBlock(p.blockChain(), signs, p.curProcessReq.req.Header.ParentHash)
	if err != nil {
		log.Trace(p.logExtraInfo(), "POS验证处理", "POS未通过", "err", err, "高度", p.number)
		return
	}
	log.INFO(p.logExtraInfo(), "POS验证处理", "POS通过", "正确签名数量", len(rightSigns), "高度", p.number)
	p.curProcessReq.posFinished = true
	p.curProcessReq.req.Header.Signatures = rightSigns

	p.finishedProcess()
}

func (p *Process) finishedProcess() {
	result := p.curProcessReq.localVerifyResult
	if result == localVerifyResultProcessing {
		log.ERROR(p.logExtraInfo(), "req is processing now, can't finish!", "validator", "高度", p.number)
		return
	}
	if result == localVerifyResultStateFailed {
		log.Error(p.logExtraInfo(), "local verify header err, but dpos pass! please check your state!", "validator", "高度", p.number)
		//todo 硬分叉了，以后加需要处理
		return
	}

	if result == localVerifyResultSuccess {
		// notify leader server the verify state
		notify := mc.BlockPOSFinishedNotify{
			Number:        p.number,
			Header:        p.curProcessReq.req.Header,
			ConsensusTurn: p.curProcessReq.req.ConsensusTurn,
			TxsCode:       p.curProcessReq.req.TxsCode,
		}
		mc.PublishEvent(mc.BlkVerify_POSFinishedNotify, &notify)
	}

	log.Info(p.logExtraInfo(), "关键时间点", "共识投票完毕，发送挖矿请求", "time", time.Now(), "块高", p.number)
	//给矿工发送区块验证结果
	p.startSendMineReq(&mc.HD_MiningReqMsg{Header: p.curProcessReq.req.Header})
	//给广播节点发送区块验证请求(带签名列表)
	p.startPosedReqSender(p.curProcessReq.req)
	p.state = StateEnd
}

func (p *Process) checkState(state State) bool {
	return p.state == state
}

func (p *Process) changeState(targetState State) {
	if p.state == targetState-1 {
		log.WARN(p.logExtraInfo(), "切换状态成功, 原状态", p.state.String(), "新状态", targetState.String(), "高度", p.number)
		p.state = targetState
	} else {
		log.WARN(p.logExtraInfo(), "切换状态失败, 原状态", p.state.String(), "目标状态", targetState.String(), "高度", p.number)
	}
}

func (p *Process) verifyVote(signHash common.Hash, vote common.Signature, from common.Address, blkHash common.Hash, verifyFrom bool) (*common.VerifiedSign, error) {
	signAccount, validate, err := p.signHelper().VerifySignWithValidateDependHash(signHash.Bytes(), vote.Bytes(), blkHash)
	if err != nil {
		return nil, err
	}

	if verifyFrom && signAccount != from {
		return nil, errors.Errorf("vote sign account[%s] != from account[%s]", signAccount.Hex(), from.Hex())
	}

	//todo 股权消息
	return &common.VerifiedSign{
		Sign:     vote,
		Account:  signAccount,
		Validate: validate,
		Stock:    0,
	}, nil
}

func (p *Process) signHelper() *signhelper.SignHelper { return p.pm.signHelper }

func (p *Process) blockChain() *core.BlockChain { return p.pm.bc }

func (p *Process) txPool() *core.TxPoolManager { return p.pm.txPool } //Y

func (p *Process) reElection() *reelection.ReElection { return p.pm.reElection }

func (p *Process) logExtraInfo() string { return p.pm.logExtraInfo() }

func (p *Process) eventMux() *event.TypeMux { return p.pm.event }

func (p *Process) ChainDb() mandb.Database { return p.pm.chainDB }
