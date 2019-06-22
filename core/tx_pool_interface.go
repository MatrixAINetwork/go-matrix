// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package core

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
)

//消息类型
const (
	tmpEmpty = iota //
	SendFloodSN
	GetTxbyN
	RecvTxbyN //
	RecvErrTx //
	BroadCast //
	GetConsensusTxbyN
	RecvConsensusTxbyN
)

// TxPool interface
type TxPool interface {
	Type() byte
	Stop()
	AddTxPool(tx types.SelfTransaction) error
	Pending() (map[string]map[common.Address]types.SelfTransactions, error)
	ReturnAllTxsByN(listN []uint32, resqe byte, addr common.Address, retch chan *RetChan_txpool)
}

type TxpoolEx interface {
	DemoteUnexecutables()
	ListenUdp()
}

//Expansion interface

type RetCallTx struct {
	TXt byte
	//ListN []uint32
	Txser []types.SelfTransaction
}

//
type NetworkMsgData struct {
	SendAddress common.Address
	Data        []*MsgStruct
}

//
type MsgStruct struct {
	Msgtype    uint32
	SendAddr   common.Address
	MsgData    []byte
	TxpoolType byte
}

//消息中心的接口（如果需要消息中心就要实现这两个方法）
type MessageProcess interface {
	ProcessMsg(m NetworkMsgData)
	SendMsg(data MsgStruct)
}

//洪泛交易的接口（如果需要洪泛交易就要实现以下方法，同时还包括链表、交易流水线等）
type TxFlood interface {
	CheckTx(mapSN map[uint32]*big.Int, nid discover.NodeID)
	GetTxByN(listN []uint32, nid discover.NodeID)
	GetConsensusTxByN(listN []uint32, nid discover.NodeID)
	RecvConsensusFloodTx(mapNtx map[uint32]types.SelfTransaction, nid discover.NodeID)
	RecvFloodTx(mapNtx map[uint32]*types.Floodtxdata, nid discover.NodeID)
	RecvErrTx(addr common.Address, listS []*big.Int)
}
