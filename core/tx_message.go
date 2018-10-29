// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package core

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/p2p"
	"encoding/json"
	"math/big"
	"github.com/matrix/go-matrix/core/types"
	"sync"
	"container/list"
	"time"
	"github.com/matrix/go-matrix/params"
)

//======struct// hezi
type mapst struct {
	//sendSNList  map[*big.Int]uint32
	slist []*big.Int
	mlock sync.RWMutex
}

// hezi
type listst struct {
	list *list.List
	lk   sync.RWMutex
}

// hezi
type sendst struct {
	snlist mapst
	lst    listst
	notice chan *big.Int
}

//global  // hezi
var gSendst sendst


//hezi
func (pool *TxPool) sendMsg(data MsgStruct) {
	selfRole := ca.GetRole()
	switch data.Msgtype {
	case SendFloodSN:
		if selfRole == common.RoleValidator || selfRole == common.RoleMiner {
			log.Info("===Transaction flood", "selfRole", selfRole)
			p2p.SendToGroupWithBackup(common.RoleValidator|common.RoleBackupValidator|common.RoleBroadcast, common.NetworkMsg, []interface{}{data})
		}
	case GetTxbyN, RecvTxbyN, BroadCast,GetConsensusTxbyN,RecvConsensusTxbyN: //YY
		//给固定的节点发送根据N获取Tx的请求
		log.Info("===sendMSG ======YY====", "Msgtype", data.Msgtype)
		p2p.SendToSingle(data.NodeId, common.NetworkMsg, []interface{}{data})
	case RecvErrTx: //YY 给全部验证者发送错误交易做共识
		if selfRole == common.RoleValidator {
			log.Info("===sendMsg ErrTx===YY===", "selfRole", selfRole)
			p2p.SendToGroup(common.RoleValidator, common.NetworkMsg, []interface{}{data})
		}
	}
}

//hezi
func (pool *TxPool) ProcessMsg(m NetworkMsgData) {
	log.Info("===========ProcessMsg", "aaaaa", 0)
	var msgdata *MsgStruct
	if len(m.Data) > 0 {
		msgdata = m.Data[0]
	} else {
		return
	}
	switch msgdata.Msgtype {
	case SendFloodSN: //YY
		snmap := make(map[uint32]*big.Int)
		btTmp := msgdata.MsgData
		err := json.Unmarshal(btTmp, &snmap)
		if err != nil {
			log.Info("====hezi====", "recv_snmap:err=", err)
		}
		nodeid := m.NodeId
		log.Info("====hezi====", "recv_snmap:nodeid=", nodeid)
		pool.msg_CheckTx(snmap, nodeid)
	case GetTxbyN: //YY
		listN := new([]uint32)
		json.Unmarshal(msgdata.MsgData, &listN)
		nodeid := m.NodeId
		pool.msg_GetTxByN(*listN, nodeid)
	case GetConsensusTxbyN: //add hezi
		listN := new([]uint32)
		json.Unmarshal(msgdata.MsgData, &listN)
		nodeid := m.NodeId
		pool.msg_GetConsensusTxByN(*listN, nodeid)
	case RecvTxbyN: //YY
		nodeid := m.NodeId
		ntx := make(map[uint32]*types.Floodtxdata, 0)
		json.Unmarshal(msgdata.MsgData, &ntx)
		pool.msg_RecvFloodTx(ntx, nodeid)
	case RecvConsensusTxbyN://add hezi
		nodeid := m.NodeId
		ntx := make(map[uint32]*types.Transaction, 0)
		json.Unmarshal(msgdata.MsgData, &ntx)
		pool.msg_RecvConsensusFloodTx(ntx, nodeid)
	case RecvErrTx: //YY
		nodeid := m.NodeId
		listS := new([]*big.Int)
		json.Unmarshal(msgdata.MsgData, &listS)
		pool.msg_RecvErrTx(common.HexToAddress(nodeid.String()), *listS)
	case BroadCast:
		var tx_mx *types.Transaction_Mx
		err := json.Unmarshal(msgdata.MsgData, &tx_mx)
		tx := types.SetTransactionMx(tx_mx)
		if err == nil {
			log.Info("========YY====1", "Unmarshal:OK=", tx)
			pool.addTx(tx, false)
		} else {
			log.Info("========YY====2", "Unmarshal-TX:err=", err)
		}
	}
}

type byteNumber struct {
	maxNum, num uint32
	mu          sync.Mutex
}

func (b3 *byteNumber) getNum() uint32 {
	if b3.num < b3.maxNum {
		b3.num++
	} else {
		b3.num = 0
	}
	return b3.num
}
func (b3 *byteNumber) catNumber(nodeNum uint32) uint32 {
	b3.mu.Lock()
	defer b3.mu.Unlock()
	num := b3.getNum()
	return (num << 7) + nodeNum
}

var byte3Number = &byteNumber{maxNum: 0x1ffff, num: 0}
var byte4Number = &byteNumber{maxNum: 0x1ffffff, num: 0}

//hezi
func (pool *TxPool) packageSNList() {
	if len(gSendst.snlist.slist) == 0 {
		return
	}

	gSendst.snlist.mlock.Lock()
	lst := gSendst.snlist.slist
	gSendst.snlist.slist = make([]*big.Int, 0)
	gSendst.snlist.mlock.Unlock()

	go func(lst []*big.Int) {
		tmpsnlst := make(map[uint32]*big.Int)
		nodeNum, _ := ca.GetNodeNumber()
		for _, s := range lst {
			if pool.sTxValIsNil(s) {
				tx := pool.getTxbyS(s)
				if tx == nil {
					log.Error("packageSNList", "tx is nil", 0)
					continue
				}
				tmpnum := byte4Number.catNumber(nodeNum)
				pool.setTxNum(tx, tmpnum)
				tmpsnlst[tmpnum] = s
				pool.setnTx(tmpnum, tx)
			}
		}
		log.Info("====hezi====", "send tmpsnlst", len(tmpsnlst))

		if len(tmpsnlst) > 0 {
			bt, _ := json.Marshal(tmpsnlst)
			pool.sendMsg(MsgStruct{Msgtype: SendFloodSN, MsgData: bt})
		}
	}(lst)
}

//hezi
func addSlist(s *big.Int) {
	gSendst.snlist.mlock.Lock()
	gSendst.snlist.slist = append(gSendst.snlist.slist, s)
	gSendst.snlist.mlock.Unlock()
}

//by hezi
func (pool *TxPool) checkList() {
	flood := time.NewTicker(params.FloodTime)
	defer flood.Stop()

	for {
		select {
		case <-flood.C:
			pool.packageSNList()

		case s := <-gSendst.notice:
			addSlist(s)
			if len(gSendst.snlist.slist) >= params.FloodMaxTransactions {
				pool.packageSNList()
			}
		}
	}
}

