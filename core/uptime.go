// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package core

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/readstatedb"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/depoistInfo"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/pkg/errors"
)

func (bc *BlockChain) getUpTimeAccounts(parentHash common.Hash, bcInterval *mc.BCIntervalInfo) ([]common.Address, error) {

	upTimeAccounts := make([]common.Address, 0)
	//todo:和老吕讨论Uptime使用当前抵押值
	//log.Debug(ModuleName, "参选矿工节点uptime高度", minerNum)
	ans, err := ca.GetElectedByHeightAndRoleByHash(parentHash, common.RoleMiner)
	if err != nil {
		return nil, err
	}

	for _, v := range ans {
		upTimeAccounts = append(upTimeAccounts, v.Address)
		//log.INFO("v.Address", "v.Address", v.Address)
	}
	//log.Debug(ModuleName, "参选验证节点uptime高度", validatorNum)
	ans1, err := ca.GetElectedByHeightAndRoleByHash(parentHash, common.RoleValidator)
	if err != nil {
		return upTimeAccounts, err
	}
	//log.Debug(ModuleName, "获取所有uptime账户为", "")
	for _, v := range ans1 {
		upTimeAccounts = append(upTimeAccounts, v.Address)
		//log.INFO("v.Address", "v.Address", v.Address)
	}

	return upTimeAccounts, nil
}
func (bc *BlockChain) getUpTimeData(root []common.CoinRoot, num uint64, parentHash common.Hash) (map[common.Address]uint32, map[common.Address][]byte, error) {
	heatBeatOriginMap, error := GetBroadcastTxMap(bc, root, mc.Heartbeat)
	if nil != error {
		log.WARN(ModuleName, "获取主动心跳交易错误", error)
	}
	headerBeatMap := make(map[common.Address][]byte, 0)
	for k, v := range heatBeatOriginMap {
		//log.INFO(ModuleName, "主动心跳交易A1/A2", k.Hex())
		account0, _, err := bc.GetA0AccountFromAnyAccount(k, parentHash)
		//log.INFO(ModuleName, "主动心跳交易A0", account0.Hex())
		if nil != err {
			continue
		}
		headerBeatMap[account0] = v

	}
	//每个广播周期发一次
	calltherollUnmarshall, error := GetBroadcastTxMap(bc, root, mc.CallTheRoll)
	if nil != error {
		log.ERROR(ModuleName, "获取点名心跳交易错误", error)
		return nil, nil, error
	}
	calltherollMap := make(map[common.Address]uint32, 0)
	for _, v := range calltherollUnmarshall {
		temp := make(map[string]uint32, 0)
		error := json.Unmarshal(v, &temp)
		if nil != error {
			log.ERROR(ModuleName, "序列化点名心跳交易错误", error)
			return nil, nil, error
		}
		for k, v := range temp {
			//log.INFO(ModuleName, "点名心跳交易A1/A2", k)
			account0, _, err := bc.GetA0AccountFromAnyAccount(common.HexToAddress(k), parentHash)
			//log.INFO(ModuleName, "点名心跳交易A0", account0.Hex())
			if nil != err {
				continue
			}
			calltherollMap[account0] = v
		}
	}
	return calltherollMap, headerBeatMap, nil
}
func (bc *BlockChain) handleUpTime(BeforeLastStateRoot []common.CoinRoot, state *state.StateDBManage, accounts []common.Address, calltherollRspAccounts map[common.Address]uint32, heatBeatAccounts map[common.Address][]byte, blockNum uint64, bcInterval *mc.BCIntervalInfo, parentHash common.Hash) (map[common.Address]uint64, error) {
	HeartBeatMap := bc.getHeatBeatAccount(BeforeLastStateRoot, bcInterval, blockNum, accounts, heatBeatAccounts)

	originValidatorMap, originMinerMap, err := bc.getElectMap(parentHash, bcInterval)
	if nil != err {
		return nil, err
	}

	return bc.calcUpTime(accounts, calltherollRspAccounts, HeartBeatMap, bcInterval, state, originValidatorMap, originMinerMap), nil
}

func (bc *BlockChain) getElectMap(parentHash common.Hash, bcInterval *mc.BCIntervalInfo) (map[common.Address]uint32, map[common.Address]uint32, error) {
	eleNum := bcInterval.GetLastBroadcastNumber() - 2
	stHash, err := bc.GetAncestorHash(parentHash, eleNum)
	if err != nil {
		log.Error(ModuleName, "获取选举高度的hash败", err, "eleNum", eleNum)
		return nil, nil, err
	}
	st, err := bc.StateAtBlockHash(stHash)
	if err != nil {
		log.Error(ModuleName, "获取选举高度的状态树失败", err, "eleNum", eleNum)
		return nil, nil, err
	}
	electGraph, err := matrixstate.GetElectGraph(st)
	if err != nil {
		log.Error(ModuleName, "获取拓扑图错误", err)
		return nil, nil, errors.New("获取拓扑图错误")
	}
	if electGraph == nil {
		log.Error(ModuleName, "获取拓扑图反射错误")
		return nil, nil, errors.New("获取拓扑图反射错误")
	}
	if 0 == len(electGraph.ElectList) {
		log.Error(ModuleName, "get获取初选列表为空", "")
		return nil, nil, errors.New("get获取初选列表为空")
	}
	//log.Debug(ModuleName, "获取原始拓扑图所有的验证者和矿工，高度为", eleNum)
	originValidatorMap := make(map[common.Address]uint32, 0)
	originMinerMap := make(map[common.Address]uint32, 0)
	for _, v := range electGraph.ElectList {
		if v.Type == common.RoleValidator || v.Type == common.RoleBackupValidator {
			originValidatorMap[v.Account] = 0
		} else if v.Type == common.RoleMiner || v.Type == common.RoleBackupMiner {
			originMinerMap[v.Account] = 0
		}
	}
	return originValidatorMap, originMinerMap, nil
}

func (bc *BlockChain) getHeatBeatAccount(beforeLastStateRoot []common.CoinRoot, bcInterval *mc.BCIntervalInfo, blockNum uint64, accounts []common.Address, heatBeatAccounts map[common.Address][]byte) map[common.Address]bool {
	HeatBeatReqAccounts := make([]common.Address, 0)
	HeartBeatMap := make(map[common.Address]bool, 0)
	//subVal就是最新的广播区块，例如当前区块高度是198或者是101，那么subVal就是100

	broadcastBlock := types.RlpHash(beforeLastStateRoot).Big()
	val := new(big.Int).Rem(broadcastBlock, big.NewInt(int64(bcInterval.GetBroadcastInterval())-1))
	for _, v := range accounts {
		currentAcc := v.Big()
		ret := new(big.Int).Rem(currentAcc, big.NewInt(int64(bcInterval.GetBroadcastInterval())-1))
		if ret.Cmp(val) == 0 {
			HeatBeatReqAccounts = append(HeatBeatReqAccounts, v)
			if _, ok := heatBeatAccounts[v]; ok {
				HeartBeatMap[v] = true
			} else {
				HeartBeatMap[v] = false

			}
			log.Debug(ModuleName, "计算主动心跳的账户", v, "心跳状态", HeartBeatMap[v])
		}
	}
	return HeartBeatMap
}

func (bc *BlockChain) calcUpTime(accounts []common.Address, calltherollRspAccounts map[common.Address]uint32, HeartBeatMap map[common.Address]bool, bcInterval *mc.BCIntervalInfo, state *state.StateDBManage, originValidatorMap map[common.Address]uint32, originMinerMap map[common.Address]uint32) map[common.Address]uint64 {
	var upTime uint64
	maxUptime := bcInterval.GetBroadcastInterval() - 3
	upTimeMap := make(map[common.Address]uint64, 0)
	for _, account := range accounts {
		onlineBlockNum, ok := calltherollRspAccounts[account]
		if ok { //被点名,使用点名的uptime
			upTime = uint64(onlineBlockNum)
			if upTime < maxUptime {
				log.Debug(ModuleName, "点名账号", account, "uptime异常", upTime)
			}

		} else { //没被点名，没有主动上报，则为最大值，

			if v, ok := HeartBeatMap[account]; ok { //有主动上报
				if v {
					upTime = maxUptime
					//log.Debug(ModuleName, "没被点名，有主动上报有响应", account, "uptime", upTime)
				} else {
					upTime = 0
					log.Debug(ModuleName, "没被点名，有主动上报无响应", account, "uptime", upTime)
				}
			} else { //没被点名和主动上报
				upTime = maxUptime
				//log.Debug(ModuleName, "没被点名，没要求主动上报", account, "uptime", upTime)

			}
		}
		upTimeMap[account] = upTime
		// todo: add
		bc.saveUptime(account, upTime, state, originValidatorMap, originMinerMap)
	}
	return upTimeMap
}

//f(x)=ax+b
func (bc *BlockChain) upTimesReset(oldUpTime *big.Int, a float64, b int64) *big.Int {

	return big.NewInt(int64(a*float64(oldUpTime.Int64())) + b)

}
func (bc *BlockChain) saveUptime(account common.Address, upTime uint64, state *state.StateDBManage, originValidatorMap map[common.Address]uint32, originMinerMap map[common.Address]uint32) {
	old, err := depoistInfo.GetOnlineTime(state, account)
	if nil != err {
		return
	}
	//log.Debug(ModuleName, "读取状态树", account, "upTime处理前", old)
	var newTime *big.Int
	if _, ok := originValidatorMap[account]; ok {

		newTime = bc.upTimesReset(old, 1, int64(upTime))
		//log.Debug(ModuleName, "是原始验证节点，upTime累加", account, "upTime", newTime.Uint64())

	} else if _, ok := originMinerMap[account]; ok {
		newTime = bc.upTimesReset(old, 1, int64(upTime))
		//log.Debug(ModuleName, "是原始矿工节点，upTime累加", account, "upTime", newTime.Uint64())

	} else {
		newTime = bc.upTimesReset(old, 1, int64(upTime))
		//log.Debug(ModuleName, "其它节点，upTime累加", account, "upTime", newTime.Uint64())
	}

	depoistInfo.SetOnlineTime(state, account, newTime)

	depoistInfo.GetOnlineTime(state, account)
	//log.Debug(ModuleName, "读取存入upTime账户", account, "upTime处理后", newTime.Uint64())
}

func (bc *BlockChain) HandleUpTimeWithSuperBlock(state *state.StateDBManage, accounts []common.Address, blockNum uint64, bcInterval *mc.BCIntervalInfo) (map[common.Address]uint64, error) {
	broadcastInterval := bcInterval.GetBroadcastInterval()
	originTopologyNum := blockNum - blockNum%broadcastInterval - 1
	originTopology, err := ca.GetTopologyByNumber(common.RoleValidator|common.RoleBackupValidator|common.RoleMiner|common.RoleBackupMiner, originTopologyNum)
	if err != nil {
		return nil, err
	}
	originTopologyMap := make(map[common.Address]uint32, 0)
	for _, v := range originTopology.NodeList {
		originTopologyMap[v.Account] = 0
	}
	upTimeMap := make(map[common.Address]uint64, 0)
	for _, account := range accounts {

		upTime := broadcastInterval - 3
		//log.Debug(ModuleName, "没被点名，没要求主动上报", account, "uptime", upTime)

		// todo: add
		depoistInfo.AddOnlineTime(state, account, new(big.Int).SetUint64(upTime))
		//read, err := depoistInfo.GetOnlineTime(state, account)
		upTimeMap[account] = upTime
		if nil == err {
			//log.Debug(ModuleName, "读取状态树", account, "upTime累加", read)
			if _, ok := originTopologyMap[account]; ok {
				//updateData := new(big.Int).SetUint64(read.Uint64() / 2)
				//log.INFO(ModuleName, "是原始拓扑图节点，upTime减半", account, "upTime", read.Uint64())
				//depoistInfo.AddOnlineTime(state, account, updateData)
			}
		}

	}
	return upTimeMap, nil

}

func (bc *BlockChain) ProcessUpTime(state *state.StateDBManage, header *types.Header) (map[common.Address]uint64, error) {
	latestNum, err := matrixstate.GetUpTimeNum(state)
	if nil != err {
		return nil, err
	}

	bcInterval, err := manparams.GetBCIntervalInfoByHash(header.ParentHash)
	if err != nil {
		log.Error(ModuleName, "获取广播周期失败", err)
		return nil, err
	}

	if header.Number.Uint64() < bcInterval.GetBroadcastInterval() {
		return nil, err
	}
	superBlkCfg, err := matrixstate.GetSuperBlockCfg(state)
	if err != nil {
		return nil, errors.Errorf("get super seq error")
	}
	sbh := superBlkCfg.Num
	if latestNum < bcInterval.GetLastBroadcastNumber()+1 {
		//log.Debug(ModuleName, "区块插入验证", "完成创建work, 开始执行uptime", "高度", header.Number.Uint64())
		matrixstate.SetUpTimeNum(state, header.Number.Uint64())
		upTimeAccounts, err := bc.getUpTimeAccounts(header.ParentHash, bcInterval)
		if err != nil {
			log.ERROR("core", "获取所有抵押账户错误!", err, "高度", header.Number.Uint64())
			return nil, err
		}
		if sbh < bcInterval.GetLastBroadcastNumber() &&
			sbh >= bcInterval.GetLastBroadcastNumber()-bcInterval.GetBroadcastInterval() {
			upTimeMap, err := bc.HandleUpTimeWithSuperBlock(state, upTimeAccounts, header.Number.Uint64(), bcInterval)
			if nil != err {
				log.ERROR("core", "处理uptime错误", err)
				return nil, err
			}
			return upTimeMap, nil
		} else {
			//log.Debug(ModuleName, "获取所有心跳交易", "")
			LastStateRoot, BeforeLastStateRoot, err := bc.getPreRoot(header, bcInterval)
			if nil != err {
				return nil, err
			}

			calltherollMap, heatBeatUnmarshallMMap, err := bc.getUpTimeData(LastStateRoot, header.Number.Uint64(), header.ParentHash)
			if err != nil {
				log.WARN("core", "获取心跳交易错误!", err, "高度", header.Number.Uint64())
			}
			upTimeMap, err := bc.handleUpTime(BeforeLastStateRoot, state, upTimeAccounts, calltherollMap, heatBeatUnmarshallMMap, header.Number.Uint64(), bcInterval, header.ParentHash)
			if nil != err {
				log.ERROR("core", "处理uptime错误", err)
				return nil, err
			}
			return upTimeMap, nil
		}

	}

	return nil, nil
}
func (bc *BlockChain) getPreRoot(header *types.Header, bcInterval *mc.BCIntervalInfo) ([]common.CoinRoot, []common.CoinRoot, error) {
	var LastStateRoot []common.CoinRoot
	var BeforeLastStateRoot []common.CoinRoot
	if header.Number.Uint64() == bcInterval.GetLastBroadcastNumber()+1 {
		preBroadcastRoot, err := readstatedb.GetPreBroadcastRoot(bc, header.ParentHash)
		if err != nil {
			log.Error(ModuleName, "获取之前广播区块的root值失败 err", err)
			return nil, nil, fmt.Errorf("从状态树获取前2个广播区块root失败")
		}
		BeforeLastStateRoot = preBroadcastRoot.LastStateRoot
		LastStateRoot = bc.GetBlockByHash(header.ParentHash).Root()

	} else {
		preBroadcastRoot, err := readstatedb.GetPreBroadcastRoot(bc, header.ParentHash)
		if err != nil {
			log.Error(ModuleName, "获取之前广播区块的root值失败 err", err)
			return nil, nil, fmt.Errorf("从状态树获取前2个广播区块root失败")
		}
		LastStateRoot = preBroadcastRoot.LastStateRoot
		BeforeLastStateRoot = preBroadcastRoot.BeforeLastStateRoot

	}
	log.Debug(ModuleName, "获取最新的root", LastStateRoot, "上一个root", BeforeLastStateRoot)
	return LastStateRoot, BeforeLastStateRoot, nil
}
