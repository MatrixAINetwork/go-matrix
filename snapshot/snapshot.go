// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package snapshot

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
)

type SnapshotData struct {
	CoinTries []state.CoinTrie
	Td        *big.Int
	Block     types.Block
	Seq       uint64 //lb
}

type SnapshotDatas struct {
	Datas      []SnapshotData
	OtherTries [][]state.CoinTrie
}

type SnapshotDataV1 struct {
	CoinTries []state.CoinTrie
	Td        *big.Int
	Block     types.BlockV1
	Seq       uint64 //lb
}

type SnapshotDatasV1 struct {
	Datas      []SnapshotDataV1
	OtherTries [][]state.CoinTrie
}
