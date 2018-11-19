package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/p2p"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/trie"
	"github.com/matrix/go-matrix/core/rawdb"
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
func (bPool *BroadCastTxPool) Type() common.TxTypeInt {
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

// SetBroadcastTxs
func SetBroadcastTxs(head *types.Block, chainId *big.Int) {
	if head.Number().Uint64()%common.GetBroadcastInterval() != 0 {
		return
	}

	var (
		signer  = types.NewEIP155Signer(chainId)
		tempMap = make(map[string]map[common.Address][]byte)
	)
	log.Info("Block insert message", "height", head.Number().Uint64(), "head.Hash=", head.Hash())

	tempMap[mc.Publickey] = make(map[common.Address][]byte)
	tempMap[mc.Heartbeat] = make(map[common.Address][]byte)
	tempMap[mc.Privatekey] = make(map[common.Address][]byte)
	tempMap[mc.CallTheRoll] = make(map[common.Address][]byte)
	txs := head.Transactions()
	for _, tx := range txs {
		if len(tx.GetMatrix_EX()) > 0 && tx.GetMatrix_EX()[0].TxType == 1 {
			temp := make(map[string][]byte)
			if err := json.Unmarshal(tx.Data(), &temp); err != nil {
				log.Error("SetBroadcastTxs", "unmarshal error", err)
				continue
			}

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

	hash := head.Hash()
	for typeStr, content := range tempMap {
		if err := insertManTrie(typeStr,hash,content); err != nil {
			log.Error("SetBroadcastTxs insertDB", "height", head.Number().Uint64(),"hash",hash)
		}else{
			log.Info("SetBroadcastTxs success","content",content)
		}
	}
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
	//txs := make([]types.SelfTransaction, 0)
	//txs = append(txs, tx)
	bPool.AddTxPool(tx)
}

// SendMsg
func (bPool *BroadCastTxPool) SendMsg(data MsgStruct) {
	if data.Msgtype == BroadCast {
		data.TxpoolType = types.BroadCastTxIndex
		p2p.SendToSingle(data.NodeId, common.NetworkMsg, []interface{}{data})
	}
}

// Stop terminates the transaction pool.
func (bPool *BroadCastTxPool) Stop() {
	// Unsubscribe subscriptions registered from blockchain
	//bPool.chainHeadSub.Unsubscribe()
	//bPool.wg.Wait()
	//if ldb != nil {
	//	ldb.Close()
	//}
	log.Info("Broad Transaction pool stopped")
}



// AddTxPool
func (bPool *BroadCastTxPool) AddTxPool(tx types.SelfTransaction) (reerr error) {
	bPool.mu.Lock()
	defer bPool.mu.Unlock()
	//TODO 1、将交易dncode,2、过滤交易（白名单）
	//for _, tx := range txs {
	if uint64(tx.Size()) > params.TxSize {
		log.Error("add broadcast tx pool", "tx`s size is too big", tx.Size())
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
			log.Info("file tx_pool_broad", "func AddTxPool", "broadCast transaction add txpool success")
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
	//}
	//if len(txs) <= 0 {
	//	log.Trace("transfer txs is nil")
	//}
	return reerr //bPool.addTxs(txs, false)
}
func (bPool *BroadCastTxPool) filter(from common.Address, keydata string) (isok bool) {
	/*   TODO 第三个问题不在这实现，上面已经做了判断了
			1、从ca模块中获取顶层节点的from 然后判断交易的具体类型（心跳、公钥、私钥）查找tx中的from是否存在。
	  		2、从ca模块中获取参选节点的from（不包括顶层节点） 然后判断交易的具体类型（心跳）查找tx中的from是否存在。
			3、判断同一个节点在此区间内是否发送过相同类型的交易（每个节点在一个区间内一种类型的交易只能发送一笔）。
			4、广播交易的类型必须是已知的如果是未知的则丢弃。（心跳、点名、公钥、私钥）
	*/
	height := bPool.chain.CurrentBlock().Number()
	blockHash := bPool.chain.CurrentBlock().Hash()
	curBlockNum := height.Uint64()
	tval := curBlockNum / common.GetBroadcastInterval()
	strVal := fmt.Sprintf("%v", tval+1)
	index := strings.Index(keydata, strVal)
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
		curBroadcastNum := common.GetNextBroadcastNumber(curBlockNum)
		if broadcastNum1 != curBroadcastNum && broadcastNum2 != curBroadcastNum {
			log.Error("The current block height is higher than the broadcast block height. (func filter())")
			return false
		}
		bids := ca.GetRolesByGroup(common.RoleBroadcast)
		for _, bid := range bids {
			addr, err := ca.ConvertNodeIdToAddress(bid)
			if err != nil {
				log.Error("ConvertNodeIdToAddress error (func filter()  BroadCastTxPool)", "error", err)
				return false
			}
			if addr == from {
				return true
			}
		}
		log.Error("unknown broadcast Address. error (func filter()  BroadCastTxPool) ")
		return false
	case mc.Heartbeat:
		nodelist, err := ca.GetElectedByHeight(height)
		if err != nil {
			log.Error("getElected error (func filter()   BroadCastTxPool)", "error", err)
			return false
		}
		for _, node := range nodelist {
			if from == node.Address {
				currentAcc := from.Big()
				ret := new(big.Int).Rem(currentAcc, big.NewInt(int64(common.GetBroadcastInterval())-1))
				broadcastBlock := blockHash.Big()
				val := new(big.Int).Rem(broadcastBlock, big.NewInt(int64(common.GetBroadcastInterval())-1))
				if ret.Cmp(val) == 0 {
					return true
				}
			}
		}
		log.WARN("Unknown account information (func filter()   BroadCastTxPool),mc.Heartbeat")
		return false
	case mc.Privatekey, mc.Publickey:
		nodelist, err := ca.GetElectedByHeightAndRole(height, common.RoleValidator)
		if err != nil {
			log.Error("getElected error (func filter()   BroadCastTxPool)", "error", err)
			return false
		}
		for _, node := range nodelist {
			if from == node.Address {
				return true
			}
		}
		log.WARN("Unknown account information (func filter()   BroadCastTxPool),mc.Privatekey,mc.Publickey")
		return false
	default:
		log.WARN("Broadcast transaction type unknown (func filter()  BroadCastTxPool),default")
		return false
	}
}

// Pending
func (bPool *BroadCastTxPool) Pending() (map[common.Address][]types.SelfTransaction, error) {
	return nil, nil
}

// insertDB
//func insertManTrie(keyData []byte, val map[common.Address][]byte,bc *BlockChain) error {
func insertManTrie(txtype string,hash common.Hash, val map[common.Address][]byte) error {
	keyData := types.RlpHash(txtype + hash.String())
	dataVal, err := json.Marshal(val)
	if err != nil {
		log.Error("insertDB", "json.Marshal(val) err", err)
		return err
	}
	key := append(rawdb.BroadcastPrefix,keyData.Bytes()...)
	return rawdb.SetManTrie(key,dataVal)
}

// GetBroadcastTxs get broadcast transactions' data from stateDB.
func GetBroadcastTxs(hash common.Hash, txtype string) (reqVal map[common.Address][]byte, err error) {
	keyData := types.RlpHash(txtype + hash.String())
	key := append(rawdb.BroadcastPrefix,keyData.Bytes()...)
	dataVal,err := trie.ManTrie.TryGet(key)
	//dataVal, err := ldb.Get(hv.Bytes(), nil)
	if err != nil {
		log.Error("GetBroadcastTxs from trie failed","keydata",key)
		return nil, err
	}

	err = json.Unmarshal(dataVal, &reqVal)
	if err != nil {
		log.Error("GetBroadcastTxs", "Unmarshal failed", err)
	}
	log.Info("GetBroadcastTxs","type",txtype,"reqval",reqVal,"keydata",key)
	return reqVal, err
}

// GetAllSpecialTxs get BroadCast transaction. (use apply SelfTransaction)
func (bPool *BroadCastTxPool) GetAllSpecialTxs() map[common.Address][]types.SelfTransaction {
	bPool.mu.Lock()
	defer bPool.mu.Unlock()
	reqVal := make(map[common.Address][]types.SelfTransaction, 0)
	log.Info("File tx_pool_broad", "func GetAllSpecialTxs:len(bPool.special)", len(bPool.special))
	for _, tx := range bPool.special {
		from, err := bPool.checkTxFrom(tx)
		if err != nil {
			log.Error("BroadCastTxPool", "GetAllSpecialTxs", err)
			continue
		}
		reqVal[from] = append(reqVal[from], tx)
	}
	bPool.special = make(map[common.Hash]types.SelfTransaction, 0)
	log.Info("File tx_pool_broad", "func GetAllSpecialTxs::len(reqVal)", len(reqVal))
	return reqVal
}

func (bPool *BroadCastTxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription {
	return nil
}
func (bPool *BroadCastTxPool) ReturnAllTxsByN(listN []uint32, resqe common.TxTypeInt, addr common.Address, retch chan *RetChan_txpool) {

}
