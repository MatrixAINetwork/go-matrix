// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package ca

import "github.com/matrix/go-matrix/params"

const (
	VerifyNetChangeUpTime = params.VerifyNetChangeUpTime //验证者网络切换时间点(提前量)
	MinerNetChangeUpTime  = params.MinerNetChangeUpTime  //矿工网络切换时间点(提前量)
)

const (
	LevelDBTopologyGraph = "TopologyGraph"
	LevelDBOriginalRole  = "OriginalRole"
)

const (
	TopNode     = 5
	DefaultNode = 6
	ErrNode     = 0
)
