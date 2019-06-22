// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixstate

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"testing"
)

type TestState struct {
	cache map[common.Hash][]byte
}

func newTestState() *TestState {
	return &TestState{
		cache: make(map[common.Hash][]byte),
	}
}

func (st *TestState) GetMatrixData(hash common.Hash) (val []byte) {
	data, exist := st.cache[hash]
	if exist {
		return data
	}
	return nil
}

func (st *TestState) SetMatrixData(hash common.Hash, val []byte) {
	st.cache[hash] = val
}

func Test_Manager(t *testing.T) {
	log.InitLog(3)

	st := newTestState()

	account1 := common.HexToAddress("0x12345")
	account2 := common.HexToAddress("0x543210")
	optV2, _ := mangerBeta.FindOperator(mc.MSKeyAccountBroadcasts)
	optV2.SetValue(st, []common.Address{account2, account1})

	use_st(st)
}

func use_st(state *TestState) {
	optV2, _ := mangerBeta.FindOperator(mc.MSKeyAccountBroadcasts)
	accounts, err := optV2.GetValue(state)
	log.Info("new get", "accounts", accounts.([]common.Address), "err", err)
}

func Test_GetUpTime(t *testing.T) {
	log.InitLog(3)
	st := newTestState()
	if err := SetUpTimeNum(st, uint64(333)); err != nil {
		t.Fatal(err)
	}
	num, err := GetUpTimeNum(st)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(num)
}
