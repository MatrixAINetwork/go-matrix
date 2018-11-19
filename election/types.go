// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package election

import "github.com/matrix/go-matrix/common"

type CandidateInfo struct {
	Address common.Address
	TPS     uint64
	UpTime  uint64
	Deposit uint64
}

type ElectionResultInfo struct {
	Address common.Address
	Stake   uint64
}

type topoGen interface {
	MinerTopoGen()
	// Hashrate returns the current mining hashrate of a PoW consensus engine.
	ValidatorTopoGen()
}
