// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package olconsensus

import (
	"testing"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

func TestNewDPosVoteRing(t *testing.T) {
	ring := NewDPosVoteRing(64)
	test := mc.OnlineConsensusReq{}
	for i := 0; i < 100; i++ {
		test.OnlineState = 1
		test.Leader = common.Address{}
		test.Seq = uint64(i + 1)
		test.Node = common.Address{}
		hash := types.RlpHash(&test)
		t.Log(hash)
		ring.addProposal(hash, &test)
		t.Log(ring.last)
		if ring.last != i%64 {
			t.Fatalf("Ring add Error,(%v),(%v)", ring.last, i)
		}
	}
}
