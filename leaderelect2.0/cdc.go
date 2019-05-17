// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect2

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/depoistInfo"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/pkg/errors"
)

type cdc struct {
	state            stateDef
	number           uint64
	selfAddr         common.Address // 自己的抵押账户A0
	selfNodeAddr     common.Address // 自己的实际node账户
	role             common.RoleType
	curConsensusTurn mc.ConsensusTurnInfo
	consensusLeader  common.Address
	curReelectTurn   uint32
	reelectMaster    common.Address
	isMaster         bool
	leaderCal        *leaderCalculator
	bcInterval       *mc.BCIntervalInfo
	parentState      StateReader
	turnTime         *turnTimes
	chain            *core.BlockChain
	logInfo          string
}

func newCDC(number uint64, chain *core.BlockChain, logInfo string) *cdc {
	dc := &cdc{
		state:            stIdle,
		number:           number,
		selfAddr:         common.Address{},
		selfNodeAddr:     common.Address{},
		role:             common.RoleNil,
		curConsensusTurn: mc.ConsensusTurnInfo{},
		consensusLeader:  common.Address{},
		curReelectTurn:   0,
		reelectMaster:    common.Address{},
		isMaster:         false,
		bcInterval:       nil,
		parentState:      nil,
		turnTime:         newTurnTimes(),
		chain:            chain,
		logInfo:          logInfo,
	}

	dc.leaderCal = newLeaderCalculator(chain, dc.number, dc.logInfo)
	return dc
}

func (dc *cdc) AnalysisState(parentHeader *types.Header, parentState StateReader) error {
	if parentState == nil || parentHeader == nil {
		return errors.New("parent state or parentHeader is nil")
	}

	validators, role, err := dc.readValidatorsAndRoleFromState(parentState)
	if err != nil {
		return err
	}
	specials, err := dc.readSpecialAccountsFromState(parentState)
	if err != nil {
		return err
	}
	config, err := dc.readLeaderConfigFromState(parentState)
	if err != nil {
		return err
	}
	bcInterval, err := dc.readBroadCastIntervalFromState(parentState)
	if err != nil {
		return err
	}

	if err := dc.leaderCal.SetValidatorsAndSpecials(parentHeader, validators, specials, bcInterval); err != nil {
		log.Warn(dc.logInfo, "SetValidatorsAndSpecials err", err)
		return err
	}

	consensusIndex := dc.curConsensusTurn.TotalTurns()
	consensusLeader, err := dc.GetLeader(consensusIndex, bcInterval)
	if err != nil {
		log.Warn(dc.logInfo, "dc.GetLeader() err", err)
		return err
	}
	if dc.curReelectTurn != 0 {
		reelectLeader, err := dc.GetLeader(consensusIndex+dc.curReelectTurn, bcInterval)
		if err != nil {
			return err
		}
		dc.reelectMaster.Set(reelectLeader)
	} else {
		dc.reelectMaster.Set(common.Address{})
	}
	if err := dc.turnTime.SetTimeConfig(config); err != nil {
		log.Error(dc.logInfo, "turnTime设置时间配置参数失败", err)
		return err
	}
	dc.bcInterval = bcInterval
	dc.consensusLeader.Set(consensusLeader)
	dc.parentState = parentState
	dc.role = role

	return nil
}

func (dc *cdc) SetConsensusTurn(consensusTurn mc.ConsensusTurnInfo) error {
	consensusLeader, err := dc.GetLeader(consensusTurn.TotalTurns(), dc.bcInterval)
	if err != nil {
		return errors.Errorf("获取共识leader错误(%v), 共识轮次: %s", err, consensusTurn.String())
	}

	dc.consensusLeader.Set(consensusLeader)
	dc.curConsensusTurn = consensusTurn
	dc.reelectMaster.Set(common.Address{})
	dc.curReelectTurn = 0
	return nil
}

func (dc *cdc) SetReelectTurn(reelectTurn uint32) error {
	if dc.curReelectTurn == reelectTurn {
		return nil
	}
	if reelectTurn == 0 {
		dc.reelectMaster.Set(common.Address{})
		dc.curReelectTurn = 0
		return nil
	}
	master, err := dc.GetLeader(dc.curConsensusTurn.TotalTurns()+reelectTurn, dc.bcInterval)
	if err != nil {
		return errors.Errorf("获取master错误(%v), 重选轮次(%d), 共识轮次(%d)", err, reelectTurn, dc.curConsensusTurn.String())
	}
	dc.reelectMaster.Set(master)
	dc.curReelectTurn = reelectTurn
	return nil
}

func (dc *cdc) GetLeader(turn uint32, bcInterval *mc.BCIntervalInfo) (common.Address, error) {
	leaders, err := dc.leaderCal.GetLeader(turn, bcInterval)
	if err != nil {
		return common.Address{}, err
	}
	return leaders.leader, nil
}

func (dc *cdc) GetConsensusLeader() common.Address {
	return dc.consensusLeader
}

func (dc *cdc) GetReelectMaster() common.Address {
	return dc.reelectMaster
}

func (dc *cdc) PrepareLeaderMsg() (*mc.LeaderChangeNotify, error) {
	leaders, err := dc.leaderCal.GetLeader(dc.curConsensusTurn.TotalTurns()+dc.curReelectTurn, dc.bcInterval)
	if err != nil {
		return nil, err
	}

	beginTime, endTime := dc.turnTime.CalTurnTime(dc.curConsensusTurn.TotalTurns(), dc.curReelectTurn)
	return &mc.LeaderChangeNotify{
		PreLeader:      dc.leaderCal.preLeader,
		Leader:         leaders.leader,
		NextLeader:     leaders.nextLeader,
		ConsensusTurn:  dc.curConsensusTurn,
		ReelectTurn:    dc.curReelectTurn,
		Number:         dc.number,
		ConsensusState: dc.state != stReelect,
		TurnBeginTime:  beginTime,
		TurnEndTime:    endTime,
	}, nil
}

func (dc *cdc) readValidatorsAndRoleFromState(state StateReader) ([]mc.TopologyNodeInfo, common.RoleType, error) {
	topology, err := matrixstate.GetTopologyGraph(state)
	if err != nil {
		return nil, common.RoleNil, err
	}
	if topology == nil {
		return nil, common.RoleNil, errors.New("topology data is nil")
	}

	role := dc.getRoleFromTopology(topology)

	validators := make([]mc.TopologyNodeInfo, 0)
	for _, node := range topology.NodeList {
		if node.Type == common.RoleValidator {
			validators = append(validators, node)
		}
	}
	return validators, role, nil
}

func (dc *cdc) getRoleFromTopology(TopologyGraph *mc.TopologyGraph) common.RoleType {
	for _, v := range TopologyGraph.NodeList {
		if v.Account == dc.selfAddr {
			return v.Type
		}
	}
	return common.RoleNil
}

func (dc *cdc) readSpecialAccountsFromState(state StateReader) (*specialAccounts, error) {
	broadcasts, err := matrixstate.GetBroadcastAccounts(state)
	if err != nil {
		return nil, err
	}

	versionSupers, err := matrixstate.GetVersionSuperAccounts(state)
	if err != nil {
		return nil, err
	}

	blockSupers, err := matrixstate.GetBlockSuperAccounts(state)
	if err != nil {
		return nil, err
	}

	return &specialAccounts{
		broadcasts:    broadcasts,
		versionSupers: versionSupers,
		blockSupers:   blockSupers,
	}, nil
}

func (dc *cdc) readLeaderConfigFromState(state StateReader) (*mc.LeaderConfig, error) {
	config, err := matrixstate.GetLeaderConfig(state)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("LeaderConfig == nil")
	}
	return config, nil
}

func (dc *cdc) readBroadCastIntervalFromState(state StateReader) (*mc.BCIntervalInfo, error) {
	interval, err := matrixstate.GetBroadcastInterval(state)
	if err != nil {
		return nil, err
	}
	if interval == nil {
		return nil, errors.New("broadcast interval is nil")
	}
	return interval, nil
}

//////////////////////////////////////////////////////////////////////////////////////////
//提供共识引擎调用，获取数据的接口
func (dc *cdc) GetCurrentHash() common.Hash {
	return dc.leaderCal.preHash
}

func (dc *cdc) GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error) {
	if (hash == common.Hash{}) {
		return nil, nil, errors.New("输入hash为空")
	}
	if hash == dc.leaderCal.preHash {
		return dc.chain.GetGraphByState(dc.parentState)
	}
	return dc.chain.GetGraphByHash(hash)
}

func (dc *cdc) GetBroadcastAccounts(blockHash common.Hash) ([]common.Address, error) {
	if (blockHash == common.Hash{}) {
		return nil, errors.New("输入hash为空")
	}
	if blockHash == dc.leaderCal.preHash {
		return dc.leaderCal.specialAccounts.broadcasts, nil
	}
	return dc.chain.GetBroadcastAccounts(blockHash)
}

func (dc *cdc) GetVersionSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	if (blockHash == common.Hash{}) {
		return nil, errors.New("输入hash为空")
	}
	if blockHash == dc.leaderCal.preHash {
		return dc.leaderCal.specialAccounts.versionSupers, nil
	}
	return dc.chain.GetVersionSuperAccounts(blockHash)
}

func (dc *cdc) GetBlockSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	if (blockHash == common.Hash{}) {
		return nil, errors.New("输入hash为空")
	}
	if blockHash == dc.leaderCal.preHash {
		return dc.leaderCal.specialAccounts.blockSupers, nil
	}
	return dc.chain.GetBlockSuperAccounts(blockHash)
}

func (dc *cdc) GetBroadcastIntervalByHash(blockHash common.Hash) (*mc.BCIntervalInfo, error) {
	if (blockHash == common.Hash{}) {
		return nil, errors.New("输入hash为空")
	}
	if blockHash == dc.leaderCal.preHash {
		if dc.bcInterval == nil {
			return nil, errors.New("缓存中不存在广播周期信息")
		}
		return dc.bcInterval, nil
	}
	return dc.chain.GetBroadcastIntervalByHash(blockHash)
}

func (dc *cdc) GetSignAccountPassword(signAccounts []common.Address) (common.Address, string, error) {
	return dc.chain.GetSignAccountPassword(signAccounts)
}

func (dc *cdc) GetA2AccountsFromA0Account(a0Account common.Address, blockHash common.Hash) ([]common.Address, error) {
	if blockHash.Equal(common.Hash{}) {
		log.Error(common.SignLog, "cdc获取A2账户", "输入数据区块hash为空")
		return nil, errors.New("cdc:输入hash为空")
	}

	if blockHash != dc.leaderCal.preHash {
		log.Info(common.SignLog, "cdc获取A2账户", "调blockchain接口")
		return dc.chain.GetA2AccountsFromA0Account(a0Account, blockHash)
	}

	return dc.getA2Accounts(a0Account, blockHash, dc.number-1)
}

func (dc *cdc) GetA0AccountFromAnyAccount(account common.Address, blockHash common.Hash) (common.Address, common.Address, error) {
	if blockHash == (common.Hash{}) {
		log.ERROR(common.SignLog, "CDC获取A0账户", "输入的hash为空")
		return common.Address{}, common.Address{}, errors.New("cdc: 输入hash为空")
	}
	if blockHash != dc.leaderCal.preHash {
		log.Warn(common.SignLog, "CDC获取A0账户", "采用blockchain的接口")
		return dc.chain.GetA0AccountFromAnyAccount(account, blockHash)
	}

	return dc.getA0Account(account, blockHash, dc.number-1)
}

func (dc *cdc) GetA2AccountsFromA0AccountAtSignHeight(a0Account common.Address, blockHash common.Hash, signHeight uint64) ([]common.Address, error) {
	if blockHash.Equal(common.Hash{}) {
		log.Error(common.SignLog, "cdc获取A2账户", "输入数据区块hash为空")
		return nil, errors.New("cdc:输入hash为空")
	}

	if blockHash != dc.leaderCal.preHash {
		log.Info(common.SignLog, "cdc获取A2账户", "调blockchain接口")
		return dc.chain.GetA2AccountsFromA0AccountAtSignHeight(a0Account, blockHash, signHeight)
	}

	return dc.getA2Accounts(a0Account, blockHash, signHeight)
}

func (dc *cdc) GetA0AccountFromAnyAccountAtSignHeight(account common.Address, blockHash common.Hash, signHeight uint64) (common.Address, common.Address, error) {
	if blockHash == (common.Hash{}) {
		log.ERROR(common.SignLog, "CDC获取A0账户", "输入的hash为空")
		return common.Address{}, common.Address{}, errors.New("cdc: 输入hash为空")
	}
	if blockHash != dc.leaderCal.preHash {
		log.Warn(common.SignLog, "CDC获取A0账户", "采用blockchain的接口")
		return dc.chain.GetA0AccountFromAnyAccountAtSignHeight(account, blockHash, signHeight)
	}
	return dc.getA0Account(account, blockHash, signHeight)
}

func (dc *cdc) getA2Accounts(a0Account common.Address, blockHash common.Hash, signHeight uint64) ([]common.Address, error) {
	if nil == dc.parentState {
		log.Info(common.SignLog, "cdc获取A2账户", "dc.parentState是空")
		return nil, errors.New("cdc: parent stateDB is nil, can't reader data")
	}

	a1Account := depoistInfo.GetAuthAccount(dc.parentState, a0Account)
	if a1Account == (common.Address{}) {
		log.Error(common.SignLog, "cdc获取A2账户", " 不存在A1账户", " a0Account", a0Account.Hex())
		return nil, errors.New("不存在A1账户")
	}

	a2Accounts := dc.parentState.GetEntrustFrom(params.MAN_COIN, a1Account, signHeight)
	if len(a2Accounts) == 0 {
		log.INFO(common.SignLog, "cdc获得A2账户", "失败", "无委托交易,使用A1账户", a1Account.String(), "签名高度", signHeight)
	} else {
		log.Info(common.SignLog, "cdc获得A2账户", "成功", "账户数量", len(a2Accounts), "签名高度", signHeight)
		for i, account := range a2Accounts {
			log.Info(common.SignLog, "A2账户", i, "account", account.Hex(), "签名高度", signHeight)
		}
	}
	a2Accounts = append(a2Accounts, a1Account)
	return a2Accounts, nil
}

func (dc *cdc) getA0Account(account common.Address, blockHash common.Hash, signHeight uint64) (common.Address, common.Address, error) {
	if nil == dc.parentState {
		log.ERROR(common.SignLog, "CDC获取A0账户", "dc.parentState is nil")
		return common.Address{}, common.Address{}, errors.New("cdc: parent stateDB is nil, can't reader data")
	}

	//假设传入的account为A1账户, 获取A1账户
	a0Account := depoistInfo.GetDepositAccount(dc.parentState, account)
	if a0Account != (common.Address{}) {
		log.Debug(common.SignLog, "CDC获取A0账户", "成功", "输入A1", account.Hex(), "输出A0", a0Account.Hex())
		return a0Account, account, nil
	}

	//账户为A2账户，获取A1
	a1Account := dc.parentState.GetAuthFrom(params.MAN_COIN, account, signHeight)
	if a1Account == (common.Address{}) {
		log.Error(common.SignLog, "CDC获取A0账户", "账户不是A1也不是A2账户", "Account", account.Hex())
		return common.Address{}, common.Address{}, errors.New("账户不是A1也不是A2账户")
	}

	// 根据A1获取A0
	a0Account = depoistInfo.GetDepositAccount(dc.parentState, a1Account)
	if a0Account != (common.Address{}) {
		log.Debug(common.SignLog, "CDC获取A0账户", "成功", "输入A1", a1Account.Hex(), "输出A0", a0Account.Hex())
		return a0Account, a1Account, nil
	} else {
		log.Error(common.SignLog, "CDC获取A0账户", "A1账户获取A0账户失败", "A1Account", a1Account.Hex())
		return common.Address{}, common.Address{}, errors.New("获取A0账户失败")
	}
}
