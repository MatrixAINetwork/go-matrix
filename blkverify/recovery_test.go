// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"testing"
)

func TestDataCode(t *testing.T) {
	data, err := encodeVerifiedBlock(&mc.HD_BlkConsensusReqMsg{
		From:   common.Address{},
		Header: nil,
		ConsensusTurn: mc.ConsensusTurnInfo{
			PreConsensusTurn: 1,
			UsedReelectTurn:  3,
		},
		TxsCode:                nil,
		OnlineConsensusResults: nil,
	}, nil)

	if err != nil {
		t.Fatalf("encode err %v", err)
	}

	req, txs, err := decodeVerifiedBlock(data)
	t.Log(req, txs, err)

	indexData, err := encodeVerifiedBlockIndex(newVerifiedBlockIndex(4))
	if err != nil {
		t.Fatalf("encode err %v", err)
	}

	index, err := decodeVerifiedBlockIndex(indexData)
	t.Log(index, err)
}
