// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect2

import (
	"reflect"
	"testing"

	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
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

type testState struct {
	graphData        []byte
	accountsData     []byte
	leaderConfigData []byte
	bcIntervalData   []byte
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

func newTestState() *testState {
	ts := &testState{}
	ts.graphData, _ = json.Marshal(getTestValidatorGraph())

	return ts
}

func Test_cdc_AnalysisState(t *testing.T) {
	type fields struct {
		number  uint64
		logInfo string
	}
	type args struct {
		preHash    common.Hash
		preLeader  common.Address
		validators []mc.TopologyNodeInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "validators为nil",
			fields: fields{
				number:  88,
				logInfo: "test cdc",
			},
			args: args{
				preHash:    common.HexToHash("0x16663ee46380133bfb49410bac53a7d83b204d7c360ba23ad764fcaa36d58419"),
				preLeader:  common.HexToAddress("0x0EAd6cDB8D214389909a535d4Ccc21A393dDdBA9"),
				validators: nil,
			},
			wantErr: true,
		},
		{
			name: "leader不在validators中",
			fields: fields{
				number:  88,
				logInfo: "test cdc",
			},
			args: args{
				preHash:    common.HexToHash("0x16663ee46380133bfb49410bac53a7d83b204d7c360ba23ad764fcaa36d58419"),
				preLeader:  common.HexToAddress("0x011111DB8D214389909a535d4Ccc21A393dDdBA9"),
				validators: getTestValidatorList(),
			},
			wantErr: true,
		},
		{
			name: "正确入参",
			fields: fields{
				number:  88,
				logInfo: "test cdc",
			},
			args: args{
				preHash:   common.HexToHash("0x16663ee46380133bfb49410bac53a7d83b204d7c360ba23ad764fcaa36d58419"),
				preLeader: common.HexToAddress("0x0EAd6cDB8D214389909a535d4Ccc21A393dDdBA9"),
				validators: []mc.TopologyNodeInfo{
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
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := newCDC(tt.fields.number, nil, tt.fields.logInfo)
			if err := dc.SetValidators(tt.args.preHash, tt.args.preLeader, tt.args.validators); (err != nil) != tt.wantErr {
				t.Errorf("cdc.SetValidators() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_cdc_SetConsensusTurn(t *testing.T) {
	type fields struct {
		number     uint64
		logInfo    string
		preHash    common.Hash
		preLeader  common.Address
		validators []mc.TopologyNodeInfo
	}
	type args struct {
		consensusTurn uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "验证者列表未输入",
			fields: fields{
				number:     88,
				logInfo:    "test cdc",
				preHash:    common.Hash{},
				preLeader:  common.Address{},
				validators: nil,
			},
			args: args{
				consensusTurn: 1,
			},
			wantErr: true,
		},
		{
			name: "正确情况",
			fields: fields{
				number:     88,
				logInfo:    "test cdc",
				preHash:    common.HexToHash("0x16663ee46380133bfb49410bac53a7d83b204d7c360ba23ad764fcaa36d58419"),
				preLeader:  common.HexToAddress("0x0EAd6cDB8D214389909a535d4Ccc21A393dDdBA9"),
				validators: getTestValidatorList(),
			},
			args: args{
				consensusTurn: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := newCDC(tt.fields.number, nil, tt.fields.logInfo)
			dc.SetValidators(tt.fields.preHash, tt.fields.preLeader, tt.fields.validators)
			if err := dc.SetConsensusTurn(tt.args.consensusTurn); (err != nil) != tt.wantErr {
				t.Errorf("cdc.SetConsensusTurn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_cdc_GetLeader(t *testing.T) {
	manparams.BroadCastNodes = []manparams.NodeInfo{
		{
			NodeID:  discover.NodeID{},
			Address: common.HexToAddress("0xa9a2d445dd686d0bbfbffe3b5826a216e49d18a5"),
		},
	}
	type fields struct {
		number     uint64
		logInfo    string
		preHash    common.Hash
		preLeader  common.Address
		validators []mc.TopologyNodeInfo
	}
	type args struct {
		turn uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    common.Address
		wantErr bool
	}{
		{
			name: "验证者列表未输入",
			fields: fields{
				number:     88,
				logInfo:    "test cdc",
				preHash:    common.Hash{},
				preLeader:  common.Address{},
				validators: nil,
			},
			args: args{
				turn: 1,
			},
			want:    common.Address{},
			wantErr: true,
		},
		{
			name: "正确输入，第0轮次",
			fields: fields{
				number:     88,
				logInfo:    "test cdc",
				preHash:    common.HexToHash("0x16663ee46380133bfb49410bac53a7d83b204d7c360ba23ad764fcaa36d58419"),
				preLeader:  getTestValidatorList()[0].Account,
				validators: getTestValidatorList(),
			},
			args: args{
				turn: 0,
			},
			want:    getTestValidatorList()[1].Account,
			wantErr: false,
		},
		{
			name: "正确输入，第1轮次",
			fields: fields{
				number:     88,
				logInfo:    "test cdc",
				preHash:    common.HexToHash("0x16663ee46380133bfb49410bac53a7d83b204d7c360ba23ad764fcaa36d58419"),
				preLeader:  getTestValidatorList()[0].Account,
				validators: getTestValidatorList(),
			},
			args: args{
				turn: 1,
			},
			want:    getTestValidatorList()[2].Account,
			wantErr: false,
		},
		{
			name: "正确输入，第14轮次",
			fields: fields{
				number:     88,
				logInfo:    "test cdc",
				preHash:    common.HexToHash("0x16663ee46380133bfb49410bac53a7d83b204d7c360ba23ad764fcaa36d58419"),
				preLeader:  getTestValidatorList()[0].Account,
				validators: getTestValidatorList(),
			},
			args: args{
				turn: 14,
			},
			want:    getTestValidatorList()[15%4].Account,
			wantErr: false,
		},
		{
			name: "广播区块高度，第0轮次",
			fields: fields{
				number:     100,
				logInfo:    "test cdc",
				preHash:    common.HexToHash("0x16663ee46380133bfb49410bac53a7d83b204d7c360ba23ad764fcaa36d58419"),
				preLeader:  getTestValidatorList()[0].Account,
				validators: getTestValidatorList(),
			},
			args: args{
				turn: 0,
			},
			want:    common.HexToAddress("0xa9a2d445dd686d0bbfbffe3b5826a216e49d18a5"),
			wantErr: false,
		},
		{
			name: "广播区块高度，第3轮次",
			fields: fields{
				number:     100,
				logInfo:    "test cdc",
				preHash:    common.HexToHash("0x16663ee46380133bfb49410bac53a7d83b204d7c360ba23ad764fcaa36d58419"),
				preLeader:  getTestValidatorList()[0].Account,
				validators: getTestValidatorList(),
			},
			args: args{
				turn: 3,
			},
			want:    common.HexToAddress("0xa9a2d445dd686d0bbfbffe3b5826a216e49d18a5"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := newCDC(tt.fields.number, nil, tt.fields.logInfo)
			dc.SetValidators(tt.fields.preHash, tt.fields.preLeader, tt.fields.validators)
			got, err := dc.GetLeader(tt.args.turn)
			if (err != nil) != tt.wantErr {
				t.Errorf("cdc.GetLeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cdc.GetLeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cdc_GetConsensusLeader(t *testing.T) {
	type fields struct {
		state            state
		number           uint64
		curConsensusTurn uint32
		consensusLeader  common.Address
		curReelectTurn   uint32
		reelectMaster    common.Address
		isMaster         bool
		leaderCal        *leaderCalculator
		turnTime         *turnTimes
		chain            *core.BlockChain
		logInfo          string
	}
	tests := []struct {
		name   string
		fields fields
		want   common.Address
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &cdc{
				state:            tt.fields.state,
				number:           tt.fields.number,
				curConsensusTurn: tt.fields.curConsensusTurn,
				consensusLeader:  tt.fields.consensusLeader,
				curReelectTurn:   tt.fields.curReelectTurn,
				reelectMaster:    tt.fields.reelectMaster,
				isMaster:         tt.fields.isMaster,
				leaderCal:        tt.fields.leaderCal,
				turnTime:         tt.fields.turnTime,
				chain:            tt.fields.chain,
				logInfo:          tt.fields.logInfo,
			}
			if got := dc.GetConsensusLeader(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cdc.GetConsensusLeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cdc_GetReelectMaster(t *testing.T) {
	type fields struct {
		state            state
		number           uint64
		curConsensusTurn uint32
		consensusLeader  common.Address
		curReelectTurn   uint32
		reelectMaster    common.Address
		isMaster         bool
		leaderCal        *leaderCalculator
		turnTime         *turnTimes
		chain            *core.BlockChain
		logInfo          string
	}
	tests := []struct {
		name   string
		fields fields
		want   common.Address
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &cdc{
				state:            tt.fields.state,
				number:           tt.fields.number,
				curConsensusTurn: tt.fields.curConsensusTurn,
				consensusLeader:  tt.fields.consensusLeader,
				curReelectTurn:   tt.fields.curReelectTurn,
				reelectMaster:    tt.fields.reelectMaster,
				isMaster:         tt.fields.isMaster,
				leaderCal:        tt.fields.leaderCal,
				turnTime:         tt.fields.turnTime,
				chain:            tt.fields.chain,
				logInfo:          tt.fields.logInfo,
			}
			if got := dc.GetReelectMaster(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cdc.GetReelectMaster() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cdc_PrepareLeaderMsg(t *testing.T) {
	type fields struct {
		state            state
		number           uint64
		curConsensusTurn uint32
		consensusLeader  common.Address
		curReelectTurn   uint32
		reelectMaster    common.Address
		isMaster         bool
		leaderCal        *leaderCalculator
		turnTime         *turnTimes
		chain            *core.BlockChain
		logInfo          string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *mc.LeaderChangeNotify
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &cdc{
				state:            tt.fields.state,
				number:           tt.fields.number,
				curConsensusTurn: tt.fields.curConsensusTurn,
				consensusLeader:  tt.fields.consensusLeader,
				curReelectTurn:   tt.fields.curReelectTurn,
				reelectMaster:    tt.fields.reelectMaster,
				isMaster:         tt.fields.isMaster,
				leaderCal:        tt.fields.leaderCal,
				turnTime:         tt.fields.turnTime,
				chain:            tt.fields.chain,
				logInfo:          tt.fields.logInfo,
			}
			got, err := dc.PrepareLeaderMsg()
			if (err != nil) != tt.wantErr {
				t.Errorf("cdc.PrepareLeaderMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cdc.PrepareLeaderMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cdc_GetCurrentHash(t *testing.T) {
	type fields struct {
		state            state
		number           uint64
		curConsensusTurn uint32
		consensusLeader  common.Address
		curReelectTurn   uint32
		reelectMaster    common.Address
		isMaster         bool
		leaderCal        *leaderCalculator
		turnTime         *turnTimes
		chain            *core.BlockChain
		logInfo          string
	}
	tests := []struct {
		name   string
		fields fields
		want   common.Hash
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &cdc{
				state:            tt.fields.state,
				number:           tt.fields.number,
				curConsensusTurn: tt.fields.curConsensusTurn,
				consensusLeader:  tt.fields.consensusLeader,
				curReelectTurn:   tt.fields.curReelectTurn,
				reelectMaster:    tt.fields.reelectMaster,
				isMaster:         tt.fields.isMaster,
				leaderCal:        tt.fields.leaderCal,
				turnTime:         tt.fields.turnTime,
				chain:            tt.fields.chain,
				logInfo:          tt.fields.logInfo,
			}
			if got := dc.GetCurrentHash(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cdc.GetCurrentHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cdc_GetValidatorByHash(t *testing.T) {
	type fields struct {
		state            state
		number           uint64
		curConsensusTurn uint32
		consensusLeader  common.Address
		curReelectTurn   uint32
		reelectMaster    common.Address
		isMaster         bool
		leaderCal        *leaderCalculator
		turnTime         *turnTimes
		chain            *core.BlockChain
		logInfo          string
	}
	type args struct {
		hash common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *mc.TopologyGraph
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &cdc{
				state:            tt.fields.state,
				number:           tt.fields.number,
				curConsensusTurn: tt.fields.curConsensusTurn,
				consensusLeader:  tt.fields.consensusLeader,
				curReelectTurn:   tt.fields.curReelectTurn,
				reelectMaster:    tt.fields.reelectMaster,
				isMaster:         tt.fields.isMaster,
				leaderCal:        tt.fields.leaderCal,
				turnTime:         tt.fields.turnTime,
				chain:            tt.fields.chain,
				logInfo:          tt.fields.logInfo,
			}
			got, err := dc.GetValidatorByHash(tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("cdc.GetValidatorByHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cdc.GetValidatorByHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
