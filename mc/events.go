// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
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
	SendSyncRole      //lb
	TxPoolManager
	LastEventCode
)
