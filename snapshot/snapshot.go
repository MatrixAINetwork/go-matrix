// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package snapshot

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
)

//need  to clear
//var SnapShotSync bool = false
//var SaveSnapShot bool = false
//var SAVESNAPSHOTPERIOD int
//var SyncSnapShootHight uint64
/*const (
SnapStartLimit=4
)*/

type SnapshotData struct {
	CoinTries []state.CoinTrie
	Td        *big.Int
	Block     types.Block
}

type SnapshotDatas struct {
	Datas      []SnapshotData
	OtherTries [][]state.CoinTrie
}
