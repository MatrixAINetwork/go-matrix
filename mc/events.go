// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package mc

type EventCode int

const (
	NewBlockMessage EventCode = iota
	SendBroadCastTx
	HD_MiningReq
	HD_MiningRsp
	HD_BroadcastMiningReq
	HD_BroadcastMiningRsp

	//CA
	CA_RoleUpdated // RoleUpdatedMsg

	//P2P
	P2P_BlkVerifyRequest // BlockVerifyReqMsg

	//Leader service
	Leader_LeaderChangeNotify // LeaderChangeNotify

	//BlockVerify service
	HD_BlkConsensusReq
	HD_BlkConsensusVote
	BlkVerify_VerifyConsensusOK //BlockVerifyConsensusOK
	BlkVerify_VerifyStateNotify //BlockVerifyStateNotify
	BlockGenor_NewBlockReady

	ReElect_MasterMinerReElectionReqMsg

	//BlockGenor service
	BlockGenor_HeaderGenerateReq
	BlockGenor_HeaderBlockReq
	HD_NewBlockInsert
	BlockGenor_HeaderVerifyReq
	HDBlockGenor_BlockVerifyReqMsg
	BCBlockGenor_HeaderBlockReq
	GBlock_HeaderGenNotify
	BlockGenor_PreBlockBroadcastFinished

	//ReElection service
	ReElec_RandomSeedReq
	Random_ElectionSeedRsp

	//topnode online
	HD_TopNodeConsensusReq
	HD_TopNodeConsensusVote
	HD_TopNodeConsensusVoteResult

	//leader
	HD_LeaderReelectVoteReq
	HD_LeaderReelectVoteRsp
	HD_LeaderReelectConsensusBroadcast

	//Topology
	ReElec_MasterMinerReElectionReq
	ReElec_MasterValidatorElectionReq
	Topo_MasterMinerElectionRsp
	Topo_MasterValidatorElectionRsp

	//random
	ReElec_TopoSeedReq
	Random_TopoSeedRsp

	P2P_HDMSG
	P2PSENDDISPATCHERMSG

	BlockToBuckets
	BlockToElected
	BlockToLinkers
	SendUdpTx
	LastEventCode
)
