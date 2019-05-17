// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
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
	CA_ReqCurrentBlock

	//P2P
	P2P_BlkVerifyRequest // BlockVerifyReqMsg

	//Leader service
	Leader_LeaderChangeNotify // LeaderChangeNotify
	Leader_RecoveryState

	//BlockVerify service
	HD_BlkConsensusReq
	HD_BlkConsensusVote
	BlkVerify_VerifyConsensusOK //BlockVerifyConsensusOK
	BlkVerify_POSFinishedNotify //BlockPOSFinishedNotify

	//BlockGenor service
	BlockGenor_HeaderGenerateReq
	HD_NewBlockInsert
	BlockGenor_HeaderVerifyReq
	BlockGenor_NewBlockReady
	HD_FullBlockReq
	HD_FullBlockRsp

	//topnode online
	HD_TopNodeConsensusReq
	HD_TopNodeConsensusVote
	HD_TopNodeConsensusVoteResult

	//leader
	HD_LeaderReelectInquiryReq
	HD_LeaderReelectInquiryRsp
	HD_LeaderReelectReq
	HD_LeaderReelectVote
	HD_LeaderReelectBroadcast
	HD_LeaderReelectBroadcastRsp

	//Topology
	ReElec_MasterMinerReElectionReq
	ReElec_MasterValidatorElectionReq
	Topo_MasterMinerElectionRsp
	Topo_MasterValidatorElectionRsp

	//random
	ReElec_TopoSeedReq
	Random_TopoSeedRsp

	P2P_HDMSG

	BlockToBuckets
	BlockToLinkers
	SendUdpTx
	SendSyncRole //lb
	TxPoolManager

	//1030_scf
	EveryBlockSeedRsp

	BlockInserted

	//leaderV2
	HD_V2_LeaderReelectInquiryReq
	HD_V2_LeaderReelectInquiryRsp
	HD_V2_LeaderReelectReq
	HD_V2_LeaderReelectVote
	HD_V2_LeaderReelectBroadcast
	HD_V2_LeaderReelectBroadcastRsp

	LastEventCode
)
