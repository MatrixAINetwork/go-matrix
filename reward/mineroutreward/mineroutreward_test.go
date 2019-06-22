// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mineroutreward

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/reward/util"

	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/mandb"
)

func TestSetPreMinerAlpha(t *testing.T) {
	chaindb := mandb.NewMemDatabase()
	state, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))

	matrixstate.SetVersionInfo(state, manparams.VersionAlpha)
	SetPreMinerReward(state, big.NewInt(int64(2e18)), util.TxsReward, params.MAN_COIN)
	out, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	if 1 != len(out) {
		t.Error("fail")
	}
	if out[0].CoinType != params.MAN_COIN {
		t.Error("fail")
	}
	if 0 != out[0].Reward.Cmp(big.NewInt(int64(2e18))) {
		t.Error("fail")
	}
}

func TestSetPreMinerReward(t *testing.T) {
	chaindb := mandb.NewMemDatabase()
	state, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))

	matrixstate.SetVersionInfo(state, manparams.VersionBeta)
	_, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	for i := 0; i < 100; i++ {
		SetPreMinerReward(state, big.NewInt(int64(i)), util.TxsReward, strconv.Itoa(i))
	}
	out2, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	if 100 != len(out2) {

	}
	for _, v := range out2 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
	for i := 100; i < 200; i++ {
		SetPreMinerReward(state, big.NewInt(int64(i)), util.TxsReward, strconv.Itoa(i))
	}
	out3, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	for _, v := range out3 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
}
func TestModifyPreMinerReward(t *testing.T) {
	chaindb := mandb.NewMemDatabase()
	state, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))

	matrixstate.SetVersionInfo(state, manparams.VersionBeta)
	_, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	for i := 0; i < 100; i++ {
		SetPreMinerReward(state, big.NewInt(int64(i)), util.TxsReward, strconv.Itoa(i))
	}
	out2, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	if 100 != len(out2) {

	}
	for _, v := range out2 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
	for i := 0; i < 200; i = i + 2 {
		SetPreMinerReward(state, big.NewInt(int64(i*1000)), util.TxsReward, strconv.Itoa(i))
	}
	out3, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	for _, v := range out3 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
}

func TestModifyVersionReward(t *testing.T) {
	chaindb := mandb.NewMemDatabase()
	state, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))

	matrixstate.SetVersionInfo(state, manparams.VersionAlpha)
	_, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}

	SetPreMinerReward(state, big.NewInt(int64(99)), util.TxsReward, params.MAN_COIN)

	out2, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	if 1 != len(out2) {

	}

	for _, v := range out2 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
	matrixstate.SetVersionInfo(state, manparams.VersionBeta)
	for i := 0; i < 200; i = i + 2 {
		SetPreMinerReward(state, big.NewInt(int64(i*1000)), util.TxsReward, strconv.Itoa(i))
	}
	out3, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	for _, v := range out3 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
}

func TestModifyVersion2Reward(t *testing.T) {
	chaindb := mandb.NewMemDatabase()
	state, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))

	matrixstate.SetVersionInfo(state, manparams.VersionAlpha)
	_, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}

	SetPreMinerReward(state, big.NewInt(int64(99)), util.TxsReward, params.MAN_COIN)

	out2, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	if 1 != len(out2) {

	}

	for _, v := range out2 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}

	matrixstate.SetVersionInfo(state, manparams.VersionBeta)
	_, err = util.GetPreMinerReward(state, util.TxsReward)
	if nil == err {
		t.Error("fail")
	}
}
