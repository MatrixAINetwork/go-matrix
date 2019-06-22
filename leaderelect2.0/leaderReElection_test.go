// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect2

import (
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

func getTestValidatorGraph() *mc.TopologyGraph {
	return &mc.TopologyGraph{
		CurNodeNumber: 0,
		NodeList: []mc.TopologyNodeInfo{
			{
				Account:    common.HexToAddress("0x0EAd6cDB8D214389909a535d4Ccc21A393dDdBA9"),
				Position:   0,
				Type:       common.RoleValidator,
				NodeNumber: 0,
			},
			{
				Account:    common.HexToAddress("0x6a3217d128A76e4777403E092bde8362d4117773"),
				Position:   1,
				Type:       common.RoleValidator,
				NodeNumber: 1,
			},
			{
				Account:    common.HexToAddress("0xf9E18AcC86179925353713a4A5D0E9BF381fBc17"),
				Position:   2,
				Type:       common.RoleValidator,
				NodeNumber: 2,
			},
			{
				Account:    common.HexToAddress("0xa121E6670439ba37E7244d4EB18E42bd6724Ef0F"),
				Position:   3,
				Type:       common.RoleValidator,
				NodeNumber: 3,
			},
		},
	}
}

func getSpecialAccounts() *mc.MatrixSpecialAccounts {
	return &mc.MatrixSpecialAccounts{
		BroadcastAccount:     mc.NodeInfo{Address: common.HexToAddress("0x4444444444444444444444444444444444444444")},
		InnerMinerAccounts:   nil,
		FoundationAccount:    common.HexToAddress("0x4444444444444444444444444444444444444444"),
		VersionSuperAccounts: nil,
		BlockSuperAccounts:   nil,
	}
}

func getLeaderConfig() *mc.LeaderConfig {
	return &mc.LeaderConfig{
		ParentMiningTime:      20,
		PosOutTime:            20,
		ReelectOutTime:        40,
		ReelectHandleInterval: 3,
	}
}

func getbcInterval() *mc.BCIntervalInfo {
	return &mc.BCIntervalInfo{
		LastBCNumber:       0,
		LastReelectNumber:  0,
		BCInterval:         100,
		BackupEnableNumber: 0,
		BackupBCInterval:   0,
	}
}

type testState struct {
	graphData        []byte
	accountsData     []byte
	leaderConfigData []byte
	bcIntervalData   []byte
}

func newTestState() *testState {
	ts := &testState{}
	ts.graphData, _ = json.Marshal(getTestValidatorGraph())
	ts.accountsData, _ = json.Marshal(getSpecialAccounts())
	ts.leaderConfigData, _ = json.Marshal(getLeaderConfig())
	ts.bcIntervalData, _ = json.Marshal(getbcInterval())
	return ts
}

func (self *testState) GetMatrixData(hash common.Hash) (val []byte) {
	if hash == matrixstate.GetKeyHash(mc.MSKeyTopologyGraph) {
		return self.graphData
	}
	if hash == matrixstate.GetKeyHash(mc.MSKeyMatrixAccount) {
		return self.accountsData
	}
	if hash == matrixstate.GetKeyHash(mc.MSKeyLeaderConfig) {
		return self.leaderConfigData
	}
	if hash == matrixstate.GetKeyHash(mc.MSKeyBroadcastInterval) {
		return self.bcIntervalData
	}
	return nil
}

func (self *testState) SetMatrixData(hash common.Hash, val []byte) {
	return
}

func (self *testState) GetAuthFrom(entrustFrom common.Address, height uint64) common.Address {
	return common.Address{}
}

func (self *testState) GetEntrustFrom(authFrom common.Address, height uint64) []common.Address {
	return nil
}
