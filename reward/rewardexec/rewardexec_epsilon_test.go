// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package rewardexec

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/reward/cfg"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

/*func TestBlockReward_getEpsilonMinerSelect(t *testing.T) {
	type fields struct {
		chain              util.ChainReader
		st                 util.StateDB
		rewardCfg          *cfg.RewardCfg
		foundationAccount  common.Address
		innerMinerAccounts []common.Address
		bcInterval         *mc.BCIntervalInfo
		topology           *mc.TopologyGraph
		elect              *mc.ElectGraph
	}
	type args struct {
		num     uint64
		rewards map[common.Address]*big.Int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := &BlockReward{
				chain:              tt.fields.chain,
				st:                 tt.fields.st,
				rewardCfg:          tt.fields.rewardCfg,
				foundationAccount:  tt.fields.foundationAccount,
				innerMinerAccounts: tt.fields.innerMinerAccounts,
				bcInterval:         tt.fields.bcInterval,
				topology:           tt.fields.topology,
				elect:              tt.fields.elect,
			}
			br.getEpsilonMinerSelect(tt.args.num, tt.args.rewards)
		})
	}
}
*/
func TestBlockReward_getEpsilonSelectAttenuationMount(t *testing.T) {
	type fields struct {
		chain              util.ChainReader
		st                 util.StateDB
		rewardCfg          *cfg.RewardCfg
		foundationAccount  common.Address
		innerMinerAccounts []common.Address
		bcInterval         *mc.BCIntervalInfo
		topology           *mc.TopologyGraph
		elect              *mc.ElectGraph
	}
	type args struct {
		num uint64
	}
	RewardMan := new(big.Int).Mul(new(big.Int).SetUint64(4800), util.GetPrice(util.CalcEpsilon))
	RewardMan1 := util.CalcRewardMount(RewardMan, 1, 8500)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *big.Int
		want1  *big.Int
	}{
		{name: "test0", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 3000000, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{num: 300}, want: RewardMan, want1: RewardMan},
		{name: "test1", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 3000000, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{num: 3000001}, want: RewardMan, want1: RewardMan},
		{name: "test2", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 3000000, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{num: 3000301}, want: RewardMan1, want1: RewardMan1},
		{name: "test3", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 360, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{num: 601}, want: RewardMan, want1: RewardMan1},
		{name: "test4", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 301, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{num: 601}, want: RewardMan1, want1: RewardMan1},
		{name: "test5", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 601, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 601, BCInterval: 100}}, args: args{num: 601}, want: RewardMan, want1: RewardMan},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := &BlockReward{
				chain:              tt.fields.chain,
				st:                 tt.fields.st,
				rewardCfg:          tt.fields.rewardCfg,
				foundationAccount:  tt.fields.foundationAccount,
				innerMinerAccounts: tt.fields.innerMinerAccounts,
				bcInterval:         tt.fields.bcInterval,
				topology:           tt.fields.topology,
				elect:              tt.fields.elect,
			}
			got, got1 := br.getEpsilonSelectAttenuationMount(tt.args.num)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockReward.getEpsilonSelectAttenuationMount() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BlockReward.getEpsilonSelectAttenuationMount() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestBlockReward_getEpsilonSelectAttenuationNum(t *testing.T) {
	type fields struct {
		chain              util.ChainReader
		st                 util.StateDB
		rewardCfg          *cfg.RewardCfg
		foundationAccount  common.Address
		innerMinerAccounts []common.Address
		bcInterval         *mc.BCIntervalInfo
		topology           *mc.TopologyGraph
		elect              *mc.ElectGraph
	}
	type args struct {
		originBlockRewardMount *big.Int
		finalBlockRewardMount  *big.Int
	}
	RewardMan := new(big.Int).Mul(new(big.Int).SetUint64(4800), util.GetPrice(util.CalcEpsilon))
	RewardMan1 := util.CalcRewardMount(RewardMan, 1, 8500)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
		want1  uint64
	}{
		{name: "test0", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 3000000, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{originBlockRewardMount: RewardMan, finalBlockRewardMount: RewardMan}, want: 0, want1: 297},
		{name: "test1", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 3000001, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{originBlockRewardMount: RewardMan, finalBlockRewardMount: RewardMan}, want: 0, want1: 297},
		{name: "test2", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 3000002, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{originBlockRewardMount: RewardMan, finalBlockRewardMount: RewardMan1}, want: 1, want1: 296},
		{name: "test3", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 3000060, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{originBlockRewardMount: RewardMan, finalBlockRewardMount: RewardMan1}, want: 59, want1: 238},
		{name: "test5", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 3000298, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{originBlockRewardMount: RewardMan, finalBlockRewardMount: RewardMan1}, want: 295, want1: 2},
		{name: "test6", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 3000299, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{originBlockRewardMount: RewardMan, finalBlockRewardMount: RewardMan1}, want: 296, want1: 1},
		{name: "test7", fields: fields{chain: nil, st: nil, rewardCfg: &cfg.RewardCfg{Calc: util.CalcEpsilon, MinersRate: 0, ValidatorsRate: 10000, RewardMount: &mc.BlkRewardCfg{MinerMount: 4800, MinerAttenuationRate: 8500, MinerAttenuationNum: 3000120, RewardRate: mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000}}}, bcInterval: &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}}, args: args{originBlockRewardMount: RewardMan, finalBlockRewardMount: RewardMan1}, want: 118, want1: 179},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := &BlockReward{
				chain:              tt.fields.chain,
				st:                 tt.fields.st,
				rewardCfg:          tt.fields.rewardCfg,
				foundationAccount:  tt.fields.foundationAccount,
				innerMinerAccounts: tt.fields.innerMinerAccounts,
				bcInterval:         tt.fields.bcInterval,
				topology:           tt.fields.topology,
				elect:              tt.fields.elect,
			}
			got, got1 := br.getEpsilonSelectAttenuationNum(tt.args.originBlockRewardMount, tt.args.finalBlockRewardMount)
			if got != tt.want {
				t.Errorf("BlockReward.getEpsilonSelectAttenuationNum() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("BlockReward.getEpsilonSelectAttenuationNum() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

//
//func TestBlockReward_getEpsilonMinerOutRewards(t *testing.T) {
//	type fields struct {
//		chain              util.ChainReader
//		st                 util.StateDB
//		rewardCfg          *cfg.RewardCfg
//		foundationAccount  common.Address
//		innerMinerAccounts []common.Address
//		bcInterval         *mc.BCIntervalInfo
//		topology           *mc.TopologyGraph
//		elect              *mc.ElectGraph
//	}
//	type args struct {
//		num        uint64
//		parentHash common.Hash
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   map[common.Address]*big.Int
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			br := &BlockReward{
//				chain:              tt.fields.chain,
//				st:                 tt.fields.st,
//				rewardCfg:          tt.fields.rewardCfg,
//				foundationAccount:  tt.fields.foundationAccount,
//				innerMinerAccounts: tt.fields.innerMinerAccounts,
//				bcInterval:         tt.fields.bcInterval,
//				topology:           tt.fields.topology,
//				elect:              tt.fields.elect,
//			}
//			if got := br.getEpsilonMinerOutRewards(tt.args.num, tt.args.parentHash); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("BlockReward.getEpsilonMinerOutRewards() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
