package core

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/p2p/discover"
)

//消息类型
const (
	tmpEmpty = iota //YY
	SendFloodSN
	GetTxbyN
	RecvTxbyN //YY
	RecvErrTx //YY
	BroadCast //YY
	GetConsensusTxbyN
	RecvConsensusTxbyN
)


// TxPool interface
type TxPool interface {
	Type() types.TxTypeInt
	Stop()
	AddTxPool(txs []types.SelfTransaction) []error
	Pending() (map[common.Address][]*types.Transaction, error)
	SubscribeNewTxsEvent(chan<- NewTxsEvent) event.Subscription
}

//Expansion interface
type TxpoolEx interface {
	DemoteUnexecutables()
	ListenUdp()
	ReturnAllTxsByN(listN []uint32, resqe int, addr common.Address, retch chan *RetChan)
}

// hezi
type NetworkMsgData struct {
	NodeId discover.NodeID
	Data   []*MsgStruct
}

// hezi
type MsgStruct struct {
	Msgtype    uint32
	NodeId     discover.NodeID
	MsgData    []byte
	TxpoolType types.TxTypeInt
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
