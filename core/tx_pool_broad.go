// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	//"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/p2p"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	//"github.com/MatrixAINetwork/go-matrix/trie"
	//"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
)

type BroadCastTxPool struct {
	chain   blockChainBroadCast
	signer  types.Signer
	special map[common.Hash]types.SelfTransaction // All special transactions
	mu      sync.RWMutex
}

type blockChainBroadCast interface {
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
	SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription
	GetA0AccountFromAnyAccountAtSignHeight(account common.Address, blockHash common.Hash, signHeight uint64) (common.Address, common.Address, error)
}

func NewBroadTxPool(chainconfig *params.ChainConfig, chain blockChainBroadCast, path string) *BroadCastTxPool {
	bPool := &BroadCastTxPool{
		chain:   chain,
		signer:  types.NewEIP155Signer(chainconfig.ChainId),
		special: make(map[common.Hash]types.SelfTransaction, 0),
	}
	return bPool
}

// Type return txpool type.
func (bPool *BroadCastTxPool) Type() byte {
	return types.BroadCastTxIndex
}

// checkTxFrom check if tx has from.
func (bPool *BroadCastTxPool) checkTxFrom(tx types.SelfTransaction) (common.Address, error) {
	if from, err := tx.GetTxFrom(); err == nil {
		return from, nil
	}

	if from, err := types.Sender(bPool.signer, tx); err == nil {
		return from, nil
	}
	return common.Address{}, ErrInvalidSender
}

func ProduceMatrixStateData(block *types.Block, stateDb *state.StateDBManage, readFn PreStateReadFn) (interface{}, error) {
	if manparams.IsBroadcastNumberByHash(block.Number().Uint64(), block.ParentHash()) == false {
		return nil, nil
	}

	var (
		tempMap = make(map[string]map[common.Address][]byte)
	)
	log.Info("ProduceMatrixStateData message", "height", block.Number().Uint64(), "block.Hash=", block.Hash())

	tempMap[mc.Publickey] = make(map[common.Address][]byte)
	tempMap[mc.Heartbeat] = make(map[common.Address][]byte)
	tempMap[mc.Privatekey] = make(map[common.Address][]byte)
	tempMap[mc.CallTheRoll] = make(map[common.Address][]byte)
	txs := make(types.SelfTransactions, 0)
	for _, curr := range block.Currencies() {
		txs = append(txs, curr.Transactions.GetTransactions()...)
	}
	for _, tx := range txs {
		if len(tx.GetMatrix_EX()) > 0 && tx.GetMatrix_EX()[0].TxType == 1 {
			temp := make(map[string][]byte)
			if err := json.Unmarshal(tx.Data(), &temp); err != nil {
				log.Error("SetBroadcastTxs", "unmarshal error", err)
				continue
			}

			signer := types.NewEIP155Signer(tx.ChainId())
			from, err := types.Sender(signer, tx)
			if err != nil {
				log.Error("SetBroadcastTxs", "get from error", err)
				continue
			}
			for key, val := range temp {
				if strings.Contains(key, mc.Publickey) {
					tempMap[mc.Publickey][from] = val
				} else if strings.Contains(key, mc.Privatekey) {
					tempMap[mc.Privatekey][from] = val
				} else if strings.Contains(key, mc.Heartbeat) {
					tempMap[mc.Heartbeat][from] = val
				} else if strings.Contains(key, mc.CallTheRoll) {
					tempMap[mc.CallTheRoll][from] = val
				}
			}
		}
	}
	if len(tempMap) > 0 {
		log.INFO("ProduceMatrixStateData", "tempMap", tempMap)
		//这里需把map转成slice存储在状态树上
		var broadtxSlice common.BroadTxSlice
		for keystring, valmap := range tempMap {
			for keyaddr, valbyte := range valmap {
				broadtxSlice.Insert(keystring, keyaddr, valbyte)
			}
		}
		return broadtxSlice, nil
	}
	return nil, errors.New("without broadcatTxs")
}

type ChainReader interface {
	StateAt(root []common.CoinRoot) (*state.StateDBManage, error)
}

func GetBroadcastTxMap(bc ChainReader, root []common.CoinRoot, txtype string) (reqVal map[common.Address][]byte, err error) {
	state, err := bc.StateAt(root)
	if err != nil {
		log.Error("GetBroadcastTxMap StateAt err")
		return nil, err
	}

	mapdata, err := matrixstate.GetBroadcastTxs(state)
	if err != nil {
		log.Error("GetBroadcastTxMap GetDataByState err")
		return nil, err
	}
	//此处需将返回的common.BroadTxSlice转为map[]map[]
	reqVal = mapdata.FindKey(txtype)
	if reqVal != nil {
		return reqVal, nil
	}
	log.Error("GetBroadcastTxMap get broadcast map is nil")
	return nil, errors.New("GetBroadcastTxMap is nil")
}

// ProcessMsg
func (bPool *BroadCastTxPool) ProcessMsg(m NetworkMsgData) {
	if len(m.Data) <= 0 {
		log.Error("BroadCastTxPool", "ProcessMsg", "data is nil")
		return
	}
	if m.Data[0].Msgtype != BroadCast {
		return
	}

	txMx := &types.Transaction_Mx{}
	if err := json.Unmarshal(m.Data[0].MsgData, txMx); err != nil {
		log.Error("BroadCastTxPool", "ProcessMsg", err)
		return
	}

	tx := types.SetTransactionMx(txMx)
	bPool.AddTxPool(tx)
}

// SendMsg
func (bPool *BroadCastTxPool) SendMsg(data MsgStruct) {
	if data.Msgtype == BroadCast {
		data.TxpoolType = types.BroadCastTxIndex
		p2p.SendToSingle(data.SendAddr, common.NetworkMsg, []interface{}{data})
	}
}

// Stop terminates the transaction pool.
func (bPool *BroadCastTxPool) Stop() {
	log.Info("Broad Transaction pool stopped")
}

// AddTxPool
func (bPool *BroadCastTxPool) AddTxPool(tx types.SelfTransaction) (reerr error) {
	bPool.mu.Lock()
	defer bPool.mu.Unlock()
	if uint64(tx.Size()) > params.TxSize {
		log.Error("add broadcast tx pool", "tx size is too big", tx.Size())
		return reerr
	}
	if len(tx.GetMatrix_EX()) > 0 && tx.GetMatrix_EX()[0].TxType == 1 {
		from, addrerr := bPool.checkTxFrom(tx)
		if addrerr != nil {
			reerr = addrerr
			return reerr
		}
		tmpdt := make(map[string][]byte)
		err := json.Unmarshal(tx.Data(), &tmpdt)
		if err != nil {
			log.Error("add broadcast tx pool", "json.Unmarshal failed", err)
			reerr = err
			return reerr
		}
		for keydata, _ := range tmpdt {
			if !bPool.filter(from, keydata) {
				break
			}
			hash := types.RlpHash(keydata + from.String())
			if bPool.special[hash] != nil {
				log.Trace("Discarding already known broadcast transaction", "hash", hash)
				reerr = fmt.Errorf("known broadcast transaction: %x", hash)
				continue
			}
			bPool.special[hash] = tx
			log.Info("tx_pool_broad", "AddTxPool", "broadCast transaction add txpool success")
		}
	} else {
		reerr = errors.New("BroadCastTxPool:AddTxPool  Transaction type is error")
		if len(tx.GetMatrix_EX()) > 0 {
			log.Error("BroadCastTxPool:AddTxPool()", "transaction type error.Extra_tx type", tx.GetMatrix_EX()[0].TxType)
		} else {
			log.Error("BroadCastTxPool:AddTxPool()", "transaction type error.Extra_tx count", len(tx.GetMatrix_EX()))
		}
		return reerr
	}
	return reerr //bPool.addTxs(txs, false)
}
func (bPool *BroadCastTxPool) filter(from common.Address, keydata string) (isok bool) {
	/*    第三个问题不在这实现，上面已经做了判断了
			1、从ca模块中获取顶层节点的from 然后判断交易的具体类型（心跳、公钥、私钥）查找tx中的from是否存在。
	  		2、从ca模块中获取参选节点的from（不包括顶层节点） 然后判断交易的具体类型（心跳）查找tx中的from是否存在。
			3、判断同一个节点在此区间内是否发送过相同类型的交易（每个节点在一个区间内一种类型的交易只能发送一笔）。
			4、广播交易的类型必须是已知的如果是未知的则丢弃。（心跳、点名、公钥、私钥）
	*/

	bcInterval := manparams.GetBCIntervalInfo()

	height := bPool.chain.CurrentBlock().Number()
	blockHash := bPool.chain.CurrentBlock().Hash()
	curBlockNum := height.Uint64()
	tval := curBlockNum / bcInterval.GetBroadcastInterval()
	strVal := fmt.Sprintf("%v", tval+1)
	index := strings.Index(keydata, strVal)
	if index < 0 {
		return false
	}
	numStr := keydata[index:]
	if numStr != strVal {
		log.Error("Future broadCast block Height error.(func filter())")
		return false
	}
	str := keydata[0:index]
	bType := mc.ReturnBroadCastType()
	if _, ok := bType[str]; !ok {
		log.Error("BroadCast Transaction type unknown. (func filter())")
		return false
	}
	switch str {
	case mc.CallTheRoll:
		broadcastNum1 := curBlockNum + 1
		broadcastNum2 := curBlockNum + 2
		curBroadcastNum := bcInterval.GetNextBroadcastNumber(curBlockNum)
		if broadcastNum1 != curBroadcastNum && broadcastNum2 != curBroadcastNum {
			log.Error("The current block height is higher than the broadcast block height. (func filter())")
			return false
		}
		addrs := ca.GetRolesByGroup(common.RoleBroadcast)
		for _, addr := range addrs {
			if addr == from {
				return true
			}
		}
		log.Error("unknown broadcast Address. error (func filter()  BroadCastTxPool) ")
		return false
	case mc.Heartbeat:
		fromDepositAccount, _, err := bPool.chain.GetA0AccountFromAnyAccountAtSignHeight(from, blockHash, bcInterval.GetNextBroadcastNumber(height.Uint64()))
		if err != nil {
			log.Error("BroadCastTxPool", "convert from account to deposit account err", err, "from", from.Hex())
			return false
		}

		nodelist, err := ca.GetElectedByHeightByHash(blockHash)
		if err != nil {
			log.Error("getElected error (func filter()   BroadCastTxPool)", "error", err)
			return false
		}
		for _, node := range nodelist {
			if fromDepositAccount == node.Address {
				currentAcc := fromDepositAccount.Big()
				ret := new(big.Int).Rem(currentAcc, big.NewInt(int64(bcInterval.GetBroadcastInterval())-1))
				broadcastBlock := blockHash.Big()
				val := new(big.Int).Rem(broadcastBlock, big.NewInt(int64(bcInterval.GetBroadcastInterval())-1))
				if ret.Cmp(val) == 0 {
					return true
				}
			}
		}
		log.WARN("Unknown account information (func filter()   BroadCastTxPool),mc.Heartbeat")
		return false
	case mc.Privatekey, mc.Publickey:
		fromDepositAccount, _, err := bPool.chain.GetA0AccountFromAnyAccountAtSignHeight(from, blockHash, bcInterval.GetNextBroadcastNumber(height.Uint64()))
		if err != nil {
			log.Error("BroadCastTxPool", "convert from account to deposit account err", err, "from", from.Hex())
			return false
		}
		nodelist, err := ca.GetElectedByHeightAndRoleByHash(blockHash, common.RoleValidator)
		if err != nil {
			log.Error("broadCastTxPool filter getElected error", "error", err)
			return false
		}
		for _, node := range nodelist {
			if fromDepositAccount == node.Address {
				return true
			}
		}
		log.WARN("Unknown account information ,mc.Privatekey,mc.Publickey")
		return false
	default:
		log.WARN("Broadcast transaction type unknown")
		return false
	}
}

// Pending
func (bPool *BroadCastTxPool) Pending() (map[string]map[common.Address]types.SelfTransactions, error) {
	return nil, nil
}

// GetAllSpecialTxs get BroadCast transaction. (use apply SelfTransaction)
func (bPool *BroadCastTxPool) GetAllSpecialTxs() map[common.Address][]types.SelfTransaction {
	bPool.mu.Lock()
	defer bPool.mu.Unlock()
	reqVal := make(map[common.Address][]types.SelfTransaction, 0)
	log.Info("BroadCastTxPool getAllSpecialTxs", "len(bPool.special)", len(bPool.special))
	for _, tx := range bPool.special {
		from, err := bPool.checkTxFrom(tx)
		if err != nil {
			log.Error("BroadCastTxPool", "GetAllSpecialTxs err", err)
			continue
		}
		reqVal[from] = append(reqVal[from], tx)
	}
	bPool.special = make(map[common.Hash]types.SelfTransaction, 0)
	log.Info("BroadCastTxPool getAllSpecialTxs", "len(reqVal)", len(reqVal))
	return reqVal
}
func (bPool *BroadCastTxPool) ReturnAllTxsByN(listN []uint32, resqe byte, addr common.Address, retch chan *RetChan_txpool) {

}
