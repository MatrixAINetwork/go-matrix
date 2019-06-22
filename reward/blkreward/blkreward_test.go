// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkreward

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/mc"

	"github.com/MatrixAINetwork/go-matrix/core/state"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/reward"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

const (
	testAddress               = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
	MinersBlockRewardRate     = uint64(5000) //矿工网络奖励50%
	ValidatorsBlockRewardRate = uint64(5000) //验证者网络奖励50%

	MinerOutRewardRate        = uint64(4000) //出块矿工奖励40%
	ElectedMinerRewardRate    = uint64(5000) //当选矿工奖励50%
	FoundationMinerRewardRate = uint64(1000) //基金会网络奖励10%

	LeaderRewardRate               = uint64(4000) //出块验证者（leader）奖励40%
	ElectedValidatorsRewardRate    = uint64(5000) //当选验证者奖励60%
	FoundationValidatorsRewardRate = uint64(1000) //基金会网络奖励10%

	OriginElectOfflineRewardRate = uint64(5000) //初选下线验证者奖励50%
	BackupRate                   = uint64(5000) //当前替补验证者奖励50%
)

type Chain struct {
	blockCache map[uint64]*types.Block
	curtop     *mc.TopologyGraph
	elect      *mc.ElectGraph
}

func (chain *Chain) GetHeaderByHash(hash common.Hash) *types.Header {
	header := &types.Header{}
	return header
}
func (chain *Chain) StateAt(root []common.CoinRoot) (*state.StateDBManage, error) {
	return nil, nil
}

func (chain *Chain) State() (*state.StateDBManage, error) {
	return nil, nil
}

func (chain *Chain) GetGraphByState(state matrixstate.StateDB) (*mc.TopologyGraph, *mc.ElectGraph, error) {
	return nil, nil, nil
}

func (chain *Chain) StateAtBlockHash(hash common.Hash) (*state.StateDBManage, error) {
	return nil, nil
}

func (chain *Chain) GetAncestorHash(sonHash common.Hash, ancestorNumber uint64) (common.Hash, error) {
	return common.Hash{}, nil
}

func (chain *Chain) SetGraphByState(node *mc.TopologyGraph, elect *mc.ElectGraph) {
	chain.curtop = node
	chain.elect = elect
	return
}

type MyStateDB struct {
	balance map[string]common.BalanceType
}

func (s *MyStateDB) GetBalance(typ string, addr common.Address) common.BalanceType {
	return s.balance[typ]
}
func (s *MyStateDB) GetMatrixData(hash common.Hash) (val []byte) {
	return nil
}
func (s *MyStateDB) SetMatrixData(hash common.Hash, val []byte) {
	return
}

func (s *MyStateDB) SetBalance(typ string, AccountType uint32, addr common.Address, mount *big.Int) {
	BalanceType := make([]common.BalanceSlice, 0)
	BalanceType = append(BalanceType, common.BalanceSlice{AccountType: common.MainAccount, Balance: mount})
	s.balance[typ] = BalanceType
}

func TestNew(t *testing.T) {
	type args struct {
		chain  util.ChainReader
		st     util.StateDB
		preSt  util.StateDB
		ppreSt util.StateDB
	}
	tests := []struct {
		name string
		args args
		want reward.Reward
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.chain, tt.args.st, tt.args.preSt, tt.args.ppreSt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
	currentState := &MyStateDB{balance: make(map[string]common.BalanceType)}
	preState := &MyStateDB{balance: make(map[string]common.BalanceType)}
	ppreState := &MyStateDB{balance: make(map[string]common.BalanceType)}
	bc := &Chain{}
	matrixstate.SetBlkCalc(preState, "1")
	rrc := mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationMinerRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationValidatorsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}
	matrixstate.SetBlkRewardCfg(preState, &mc.BlkRewardCfg{MinerMount: 5, MinerAttenuationNum: 10000, ValidatorMount: 5, ValidatorAttenuationNum: 10000, RewardRate: rrc})
	matrixstate.SetBroadcastInterval(preState, &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100})
	matrixstate.SetFoundationAccount(preState, common.HexToAddress(testAddress))
	matrixstate.SetInnerMinerAccounts(preState, []common.Address{common.HexToAddress("0x1")})
	New(bc, currentState, preState, ppreState)
}
