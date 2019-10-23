// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package miner

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"reflect"
	"time"
)

type workerV2 struct {
	taskManger  *mineTaskManager
	curSuperSeq uint64
	curMineTask mineTask
	worker      *worker
	sendersMap  map[uint64][]*common.ResendMsgCtrl
	logInfo     string
}

func newWorkerV2(worker *worker) *workerV2 {
	return &workerV2{
		taskManger:  newMineTaskManager(worker.bc, "worker v2"),
		curSuperSeq: 0,
		curMineTask: nil,
		worker:      worker,
		sendersMap:  make(map[uint64][]*common.ResendMsgCtrl),
		logInfo:     "worker v2",
	}
}

func (self *workerV2) RoleUpdatedMsgHandler(data *mc.RoleUpdatedMsg) {
	if data.SuperSeq > self.curSuperSeq {
		self.worker.StopAgent()
		self.curMineTask = nil
		self.closeResultSenders(true)
		self.taskManger.Clear()
		self.curSuperSeq = data.SuperSeq
	} else if data.SuperSeq < self.curSuperSeq {
		return
	}

	// 更新高度和缓存
	self.taskManger.SetNewNumberAndRole(data.BlockNum+1, data.Role)

	// 关闭正在挖矿的过期任务
	self.closeExpiredCurTask()

	// 关闭过期的结果发送定时器
	self.closeResultSenders(false)

	canMining := self.taskManger.CanMining()
	log.Trace(self.logInfo, "高度", data.BlockNum+1, "角色", data.Role, "是否可以挖矿", canMining)
	if canMining {
		self.worker.StartAgent()
		self.processStartMining()
	} else {
		self.worker.StopAgent()
	}
}

func (self *workerV2) closeExpiredCurTask() {
	if isNilTask(self.curMineTask) {
		return
	}

	if self.curMineTask.MineNumber().Uint64() < self.taskManger.curNumber {
		log.Trace(self.logInfo, "关闭当前挖矿任务", self.curMineTask.TaskType().String(),
			"mine task number", self.curMineTask.MineNumber(), "cur number", self.taskManger.curNumber,
			"mine task hash", self.curMineTask.MineHash().TerminalString())
		self.worker.StopAgent()
		self.curMineTask = nil
	}
}

func (self *workerV2) MineReqMsgHandle(req *mc.HD_V2_MiningReqMsg) {
	if req == nil || req.Header == nil {
		return
	}

	if err := self.taskManger.AddMineHeader(req.Header); err != nil {
		log.Trace(self.logInfo, "挖矿请求添加缓存失败", err, "req number", req.Header.Number)
		return
	}
	log.Trace(self.logInfo, "挖矿请求添加缓存", "成功", "req number", req.Header.Number)

	self.processStartMining()
}

func (self *workerV2) processStartMining() {
	aiTask := self.taskManger.GetBestAITask()
	powTask := self.taskManger.GetBestPowTask()

	bestTask := self.packBestTask(aiTask, powTask)
	if isNilTask(bestTask) {
		log.Info(self.logInfo, "processStartMining", "当前无合适挖矿任务")
		return
	}

	// 当前有任务, 并且当前任务最优
	if isNilTask(self.curMineTask) == false && isSameTask(self.curMineTask, self.packBestTask(self.curMineTask, bestTask)) {
		log.Info(self.logInfo, "processStartMining", "当前任务为最优任务", "cur task timestamp", self.curMineTask.MineHeaderTime(), "cur task type", self.curMineTask.TaskType().String(),
			"cmp task timestamp", bestTask.MineHeaderTime(), "cmp task timestamp", bestTask.TaskType().String())
		return
	}

	// 使用最优任务替换当前任务
	self.curMineTask = bestTask
	self.worker.CommitNewWorkV2(self.curMineTask)
}

func isNilTask(task mineTask) bool {
	if task == nil {
		return true
	}
	value := reflect.ValueOf(task)
	if value.Kind() != reflect.Ptr {
		return false
	}
	return value.IsNil()
}

func isSameTask(taskA mineTask, taskB mineTask) bool {
	return taskA.MineHash() == taskB.MineHash() && taskA.TaskType() == taskB.TaskType()
}

func (self *workerV2) packBestTask(taskA mineTask, taskB mineTask) mineTask {
	if isNilTask(taskA) {
		return taskB
	}
	if isNilTask(taskB) {
		return taskA
	}

	switch taskA.MineHeaderTime().Cmp(taskB.MineHeaderTime()) {
	case 0: // taskA time == taskB time
		// 时间搓相同，优先进行POW挖矿任务
		if taskA.TaskType() == mineTaskTypePow {
			return taskA
		}
		if taskB.TaskType() == mineTaskTypePow {
			return taskB
		}
		return taskA

	case 1: // taskA time > taskB time
		return taskA
	case -1: // taskA time < taskB time
		return taskB
	default:
		return nil
	}
}

func (self *workerV2) foundHandle(header *types.Header) {
	log.Trace(self.logInfo, "foundHandle", "begin", "key hash", header.ParentHash.TerminalString(), "ai hash", header.AIHash.TerminalString(), "nonce", header.Nonce, "sm3 nonce", header.Sm3Nonce)
	defer log.Trace(self.logInfo, "foundHandle", "end", "key hash", header.ParentHash.TerminalString(), "ai hash", header.AIHash.TerminalString(), "nonce", header.Nonce, "sm3 nonce", header.Sm3Nonce)
	if (header.AICoinbase != common.Address{}) && (header.AIHash != common.Hash{}) {
		log.Trace(self.logInfo, "得到AI挖矿结果", header.AIHash.Hex(), "number", header.Number, "parent hash", header.ParentHash.TerminalString())
		self.handleAIResult(header)
	}

	if (header.Coinbase != common.Address{}) && (header.Nonce != types.BlockNonce{}) {
		if (header.Sm3Nonce != types.BlockNonce{}) {
			log.Trace(self.logInfo, "得到POW挖矿结果", header.Nonce.Uint64(), "number", header.Number, "parent hash", header.ParentHash.TerminalString())
			self.handlePowResult(header)
		} else {
			log.Trace(self.logInfo, "得到 Base POW 挖矿结果", header.Nonce.Uint64(), "number", header.Number, "parent hash", header.ParentHash.TerminalString())
			self.handleBasePowResult(header)
		}
	}

	if isNilTask(self.curMineTask) && self.taskManger.CanMining() {
		// 挖矿结束，且当前可以挖矿，继续开始下个挖矿任务
		self.processStartMining()
	}

}

func (self *workerV2) handleBasePowResult(header *types.Header) {
	if isNilTask(self.curMineTask) {
		log.Info(self.logInfo, "Base POW挖矿结果处理", "当前无挖矿任务")
		return
	}
	if self.curMineTask.MineHash() != header.ParentHash {
		// todo 其他挖矿结果延迟达到，考虑更新缓存中的信息
		log.Info(self.logInfo, "Base POW挖矿结果处理", "挖矿结果不匹配当前挖矿任务", "task mine hash", self.curMineTask.MineHash().TerminalString(), "result key hash", header.ParentHash.TerminalString())
		return
	}
	if self.curMineTask.TaskType() != mineTaskTypePow {
		log.Info(self.logInfo, "Base POW挖矿结果处理", "当前任务非POW任务")
		return
	}
	task, ok := self.curMineTask.(*powMineTask)
	if ok == false {
		log.Info(self.logInfo, "Base POW挖矿结果处理", "当前任务反射失败")
		return
	}

	if task.minedPow {
		log.Info(self.logInfo, "Base POW挖矿结果处理", "已得到过挖矿结果", "parent hash", header.ParentHash.TerminalString())
		return
	}

	if task.minedBasePow == false {
		// 未发送过基础算力结果
		if header.Difficulty.Cmp(params.BasePowerDifficulty) < 0 {
			log.Info(self.logInfo, "Base POW挖矿结果处理", "挖矿难度小于基础算力难度", "mined difficulty", header.Difficulty, "base difficulty", params.BasePowerDifficulty)
			return
		}

		task.minedBasePow = true
		sendData := &powMineTask{
			mineHash:        task.mineHash,
			mineHeader:      task.mineHeader,
			minedBasePow:    true,
			powMiningNumber: task.powMiningNumber,
			powMiner:        header.Coinbase,
			mixDigest:       header.MixDigest,
			nonce:           header.Nonce,
		}
		self.startBasePowResultSender(sendData)
	}
}

func (self *workerV2) handlePowResult(header *types.Header) {
	if isNilTask(self.curMineTask) {
		log.Info(self.logInfo, "POW挖矿结果处理", "当前无挖矿任务")
		return
	}
	if self.curMineTask.MineHash() != header.ParentHash {
		// todo 其他挖矿结果延迟达到，考虑更新缓存中的信息
		log.Info(self.logInfo, "POW挖矿结果处理", "挖矿结果不匹配当前挖矿任务", "task mine hash", self.curMineTask.MineHash().TerminalString(), "result key hash", header.ParentHash.TerminalString())
		return
	}
	if self.curMineTask.TaskType() != mineTaskTypePow {
		log.Info(self.logInfo, "POW挖矿结果处理", "当前任务非POW任务")
		return
	}
	task, ok := self.curMineTask.(*powMineTask)
	if ok == false {
		log.Info(self.logInfo, "POW挖矿结果处理", "当前任务反射失败")
		return
	}

	if task.minedPow {
		log.Info(self.logInfo, "POW挖矿结果处理", "已得到过挖矿结果", "parent hash", header.ParentHash.TerminalString())
		self.curMineTask = nil
		return
	}

	if header.Difficulty.Cmp(task.powMiningDifficulty) != 0 {
		log.Info(self.logInfo, "POW挖矿结果处理", "挖矿难度不匹配", "mined difficulty", task.powMiningDifficulty, "result difficulty", header.Difficulty)
		return
	}

	self.curMineTask = nil
	task.minedPow = true
	task.powMiner = header.Coinbase
	task.mixDigest = header.MixDigest
	task.nonce = header.Nonce
	task.sm3Nonce = header.Sm3Nonce

	sendData := &powMineTask{
		mineHash:            task.mineHash,
		mineHeader:          task.mineHeader,
		minedPow:            true,
		powMiningNumber:     task.powMiningNumber,
		powMiningDifficulty: task.powMiningDifficulty,
		powMiner:            task.powMiner,
		mixDigest:           task.mixDigest,
		nonce:               task.nonce,
		sm3Nonce:            task.sm3Nonce,
	}
	self.startPOWMineResultSender(sendData)

	if task.minedBasePow == false {
		// 未发送过基础算力结果
		if header.Difficulty.Cmp(params.BasePowerDifficulty) < 0 {
			log.Info(self.logInfo, "POW挖矿结果处理", "挖矿难度小于基础算力难度", "mined difficulty", header.Difficulty, "base difficulty", params.BasePowerDifficulty)
			return
		}

		task.minedBasePow = true
		basePowSendData := &powMineTask{
			mineHash:        task.mineHash,
			mineHeader:      task.mineHeader,
			minedBasePow:    true,
			powMiningNumber: task.powMiningNumber,
			powMiner:        header.Coinbase,
			mixDigest:       header.MixDigest,
			nonce:           header.Nonce,
		}
		self.startBasePowResultSender(basePowSendData)
	}
}

func (self *workerV2) handleAIResult(header *types.Header) {
	if isNilTask(self.curMineTask) {
		log.Info(self.logInfo, "AI挖矿结果处理", "当前无挖矿任务")
		return
	}
	if self.curMineTask.MineHash() != header.ParentHash {
		// todo 其他挖矿结果延迟达到，考虑更新缓存中的信息
		log.Info(self.logInfo, "AI挖矿结果处理", "挖矿结果不匹配当前挖矿任务", "task mine hash", self.curMineTask.MineHash().TerminalString(), "result key hash", header.ParentHash.TerminalString())
		return
	}
	if self.curMineTask.TaskType() != mineTaskTypeAI {
		log.Info(self.logInfo, "AI挖矿结果处理", "当前任务非POW任务")
		return
	}
	task, ok := self.curMineTask.(*aiMineTask)
	if ok == false {
		log.Info(self.logInfo, "AI挖矿结果处理", "当前任务反射失败")
		return
	}

	if task.minedAI {
		log.Info(self.logInfo, "AI挖矿结果处理", "已得到过挖矿结果", "parent hash", header.ParentHash.TerminalString())
		self.curMineTask = nil
		return
	}

	self.curMineTask = nil
	task.minedAI = true
	task.aiMiner = header.AICoinbase
	task.aiHash = header.AIHash

	sendData := &aiMineTask{
		mineHash:       task.mineHash,
		mineHeader:     task.mineHeader,
		minedAI:        true,
		aiMiningNumber: task.aiMiningNumber,
		aiMiner:        task.aiMiner,
		aiHash:         task.aiHash,
	}
	self.startAIMineResultSender(sendData)
}

func (self *workerV2) addResultSenders(number uint64, sender *common.ResendMsgCtrl) {
	lists, exist := self.sendersMap[number]
	if exist {
		self.sendersMap[number] = append(lists, sender)
	} else {
		self.sendersMap[number] = []*common.ResendMsgCtrl{sender}
	}
}

func (self *workerV2) closeResultSenders(force bool) {
	for number, senders := range self.sendersMap {
		if force || number < self.taskManger.curNumber {
			for _, sender := range senders {
				sender.Close()
			}

			log.Trace(self.logInfo, "结果发送器", "停止", "number", number, "cur number", self.taskManger.curNumber, "count", len(senders))
			delete(self.sendersMap, number)
		}
	}
}

func (self *workerV2) startPOWMineResultSender(task *powMineTask) {
	var sender *common.ResendMsgCtrl = nil
	var err error
	if self.taskManger.curRole == common.RoleInnerMiner {
		sender, err = common.NewResendMsgCtrl(task, self.innerMinerSendPOWMineResultFunc, 1, 0)
		if err != nil {
			log.Error(self.logInfo, "创建POW挖矿结果发送器", "失败", "err", err)
			return
		}
	} else {
		sender, err = common.NewResendMsgCtrl(task, self.sendPOWMineResultFunc, manparams.MinerResultSendInterval, 0)
		if err != nil {
			log.Error(self.logInfo, "创建POW挖矿结果发送器", "失败", "err", err)
			return
		}
	}

	log.Trace(self.logInfo, "创建POW挖矿结果发送器", "成功", "高度", task.powMiningNumber, "mine hash", task.mineHash.TerminalString(), "role", self.taskManger.curRole.String())
	self.addResultSenders(task.powMiningNumber, sender)
}

func (self *workerV2) sendPOWMineResultFunc(data interface{}, times uint32) {
	resultData, OK := data.(*powMineTask)
	if !OK {
		log.Error(self.logInfo, "发出POW挖矿结果", "反射消息失败", "次数", times)
		return
	}

	if nil == resultData {
		log.Error(self.logInfo, "发出POW挖矿结果", "入参错误", "次数", times)
		return
	}

	rsp := &mc.HD_V2_PowMiningRspMsg{
		Number:     resultData.powMiningNumber,
		BlockHash:  resultData.mineHash,
		Difficulty: resultData.powMiningDifficulty,
		Nonce:      resultData.nonce,
		Coinbase:   resultData.powMiner,
		MixDigest:  resultData.mixDigest,
		Sm3Nonce:   resultData.sm3Nonce,
	}

	self.worker.hd.SendNodeMsg(mc.HD_V2_PowMiningRsp, rsp, common.RoleValidator|common.RoleBroadcast, nil)
	log.Trace(self.logInfo, "POW挖矿结果", "发送", "parent hash", rsp.BlockHash.TerminalString(), "次数", times, "高度", rsp.Number, "x11 Nonce", rsp.Nonce, "sm3 Nonce", rsp.Sm3Nonce, "difficulty", rsp.Difficulty)
}

func (self *workerV2) innerMinerSendPOWMineResultFunc(data interface{}, times uint32) {
	resultData, OK := data.(*powMineTask)
	if !OK {
		log.Error(self.logInfo, "基金会矿工发出POW挖矿结果", "反射消息失败", "次数", times)
		return
	}

	if nil == resultData {
		log.Error(self.logInfo, "基金会矿工发出POW挖矿结果", "入参错误", "次数", times)
		return
	}

	curTime := time.Now().Unix()
	deadline := resultData.mineHeader.Time.Int64() + params.InnerMinerPowSendDelay
	if curTime < deadline {
		log.Trace(self.logInfo, "基金会矿工发出POW挖矿结果", "未到延迟时间不发送", "cur time", curTime, "deadline", deadline, "header time", resultData.mineHeader.Time)
		return
	}

	rsp := &mc.HD_V2_PowMiningRspMsg{
		Number:     resultData.powMiningNumber,
		BlockHash:  resultData.mineHash,
		Difficulty: resultData.powMiningDifficulty,
		Nonce:      resultData.nonce,
		Coinbase:   resultData.powMiner,
		MixDigest:  resultData.mixDigest,
		Sm3Nonce:   resultData.sm3Nonce,
	}

	self.worker.hd.SendNodeMsg(mc.HD_V2_PowMiningRsp, rsp, common.RoleValidator|common.RoleBroadcast, nil)
	log.Trace(self.logInfo, "基金会矿工POW挖矿结果", "发送", "parent hash", rsp.BlockHash.TerminalString(), "次数", times, "高度", rsp.Number, "x11 Nonce", rsp.Nonce, "sm3 Nonce", rsp.Sm3Nonce, "difficulty", rsp.Difficulty)
}

func (self *workerV2) startBasePowResultSender(task *powMineTask) {
	if self.taskManger.curRole == common.RoleInnerMiner {
		log.Trace(self.logInfo, "创建算力检测结果发送器", "基金会矿工不发送算力检测结果")
		return
	}
	sender, err := common.NewResendMsgCtrl(task, self.sendBasePowResultFunc, manparams.MinerResultSendInterval, 0)
	if err != nil {
		log.Error(self.logInfo, "创建算力检测结果发送器", "失败", "err", err)
		return
	}
	log.Trace(self.logInfo, "创建算力检测结果发送器", "成功", "高度", task.powMiningNumber, "mine hash", task.mineHash.TerminalString())
	self.addResultSenders(task.powMiningNumber, sender)
}

func (self *workerV2) sendBasePowResultFunc(data interface{}, times uint32) {
	resultData, OK := data.(*powMineTask)
	if !OK {
		log.Error(self.logInfo, "发出算力检测结果", "反射消息失败", "次数", times)
		return
	}

	if nil == resultData {
		log.Error(self.logInfo, "发出算力检测结果", "入参错误", "次数", times)
		return
	}

	rsp := &mc.HD_BasePowerDifficulty{
		Number:     resultData.powMiningNumber,
		BlockHash:  resultData.mineHash,
		Difficulty: params.BasePowerDifficulty,
		Nonce:      resultData.nonce,
		Coinbase:   resultData.powMiner,
		MixDigest:  resultData.mixDigest,
	}

	self.worker.hd.SendNodeMsg(mc.HD_BasePowerResult, rsp, common.RoleValidator|common.RoleBroadcast, nil)
	log.Trace(self.logInfo, "算力检测结果", "发送", "parent hash", rsp.BlockHash.TerminalString(), "次数", times, "高度", rsp.Number, "Nonce", rsp.Nonce)
}

func (self *workerV2) startAIMineResultSender(task *aiMineTask) {
	var sender *common.ResendMsgCtrl = nil
	var err error
	if self.taskManger.curRole == common.RoleInnerMiner {
		sender, err = common.NewResendMsgCtrl(task, self.innerMinerSendAIMineResultFunc, 1, 0)
		if err != nil {
			log.Error(self.logInfo, "创建AI结果发送器", "失败", "err", err)
			return
		}
	} else {
		sender, err = common.NewResendMsgCtrl(task, self.sendAIMineResultFunc, manparams.MinerResultSendInterval, 0)
		if err != nil {
			log.Error(self.logInfo, "创建AI结果发送器", "失败", "err", err)
			return
		}
	}

	log.Trace(self.logInfo, "创建AI结果发送器", "成功", "高度", task.aiMiningNumber, "mine hash", task.mineHash.TerminalString(), "role", self.taskManger.curRole.String())
	self.addResultSenders(task.aiMiningNumber, sender)
}

func (self *workerV2) sendAIMineResultFunc(data interface{}, times uint32) {
	resultData, OK := data.(*aiMineTask)
	if !OK {
		log.Error(self.logInfo, "发出AI挖矿结果", "反射消息失败", "次数", times)
		return
	}

	if nil == resultData {
		log.Error(self.logInfo, "发出AI挖矿结果", "入参错误", "次数", times)
		return
	}

	rsp := &mc.HD_V2_AIMiningRspMsg{
		Number:     resultData.aiMiningNumber,
		BlockHash:  resultData.mineHash,
		AIHash:     resultData.aiHash,
		AICoinbase: resultData.aiMiner,
	}

	self.worker.hd.SendNodeMsg(mc.HD_V2_AIMiningRsp, rsp, common.RoleValidator|common.RoleBroadcast, nil)
	log.Trace(self.logInfo, "AI挖矿结果", "发送", "parent hash", rsp.BlockHash.TerminalString(), "次数", times, "高度", rsp.Number, "AIHash", rsp.AIHash.TerminalString(), "AIMiner", rsp.AICoinbase.Hex())
}

func (self *workerV2) innerMinerSendAIMineResultFunc(data interface{}, times uint32) {
	resultData, OK := data.(*aiMineTask)
	if !OK {
		log.Error(self.logInfo, "基金会矿工发出AI挖矿结果", "反射消息失败", "次数", times)
		return
	}

	if nil == resultData {
		log.Error(self.logInfo, "基金会矿工发出AI挖矿结果", "入参错误", "次数", times)
		return
	}

	curTime := time.Now().Unix()
	deadline := resultData.mineHeader.Time.Int64() + params.InnerMinerAISendDelay
	if curTime < deadline {
		log.Trace(self.logInfo, "基金会矿工发出AI挖矿结果", "未到延迟时间不发送", "cur time", curTime, "deadline", deadline, "header time", resultData.mineHeader.Time)
		return
	}

	rsp := &mc.HD_V2_AIMiningRspMsg{
		Number:     resultData.aiMiningNumber,
		BlockHash:  resultData.mineHash,
		AIHash:     resultData.aiHash,
		AICoinbase: resultData.aiMiner,
	}

	self.worker.hd.SendNodeMsg(mc.HD_V2_AIMiningRsp, rsp, common.RoleValidator|common.RoleBroadcast, nil)
	log.Trace(self.logInfo, "基金会矿工AI挖矿结果", "发送", "parent hash", rsp.BlockHash.TerminalString(), "次数", times, "高度", rsp.Number, "AIHash", rsp.AIHash.TerminalString(), "AIMiner", rsp.AICoinbase.Hex())
}
