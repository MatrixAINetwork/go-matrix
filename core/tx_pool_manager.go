// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package core

import (
	"encoding/json"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/p2p"
	"github.com/MatrixAINetwork/go-matrix/params"
)

var (
	ErrTxPoolAlreadyExist = errors.New("txpool already exist")
	ErrTxPoolIsNil        = errors.New("txpool is nil")
	ErrTxPoolNonexistent  = errors.New("txpool nonexistent")

	blockNumberByfilter = uint64(0)
)

type CoinPachFilter struct {
	mu      sync.Mutex
	coinNum map[string]uint64
}

var filtercoinnum = CoinPachFilter{coinNum: make(map[string]uint64)}

//
type RetChan struct {
	//Rxs   []types.SelfTransaction
	AllTxs []*RetCallTx
	Err    error
	Resqe  int
}
type RetChan_txpool struct {
	Rxs  []types.SelfTransaction
	Err  error
	Tx_t byte
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
func (b3 *byteNumber) catNumber() uint32 {
	b3.mu.Lock()
	defer b3.mu.Unlock()
	nodeNum, _ := ca.GetNodeNumber()
	num := b3.getNum()
	return (num << 7) + nodeNum
}

var byte3Number = &byteNumber{maxNum: 0x1ffff, num: 0}
var byte4Number = &byteNumber{maxNum: 0x1ffffff, num: 0}

type Blacklist struct {
	Bmap map[common.Address]bool
	mu   sync.RWMutex
}

func NewInitblacklist() *Blacklist {
	b := &Blacklist{}
	b.Bmap = make(map[common.Address]bool)
	b.Bmap[common.HexToAddress("0x7097f41F1C1847D52407C629d0E0ae0fDD24fd58")] = true
	return b
}

var SelfBlackList *Blacklist

func (b *Blacklist) FindBlackAddress(addr common.Address) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.Bmap[addr]
	return ok
}
func (b *Blacklist) AddBlackAddress(addr common.Address) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Bmap[addr] = true
}

// TxPoolManager
type TxPoolManager struct {
	txPoolsMutex sync.RWMutex
	once         sync.Once
	sub          event.Subscription
	txPools      map[byte]TxPool
	roleChan     chan common.RoleType
	quit         chan struct{}
	addPool      chan TxPool
	delPool      chan TxPool
	sendTxCh     chan NewTxsEvent
	txFeed       event.Feed
	scope        event.SubscriptionScope
	chain        blockChain
}

func NewTxPoolManager(config TxPoolConfig, chainconfig *params.ChainConfig, chain blockChain, path string) *TxPoolManager {
	txPoolManager := &TxPoolManager{
		txPoolsMutex: sync.RWMutex{},
		txPools:      make(map[byte]TxPool),
		quit:         make(chan struct{}),
		roleChan:     make(chan common.RoleType),
		addPool:      make(chan TxPool),
		delPool:      make(chan TxPool),
		sendTxCh:     make(chan NewTxsEvent),
		chain:        chain,
	}
	SelfBlackList = NewInitblacklist()
	go txPoolManager.loop(config, chainconfig, chain, path)
	return txPoolManager
}

// Subscribe a txpool into manager.
func (pm *TxPoolManager) Subscribe(pool TxPool) error {
	pm.txPoolsMutex.Lock()
	defer pm.txPoolsMutex.Unlock()

	if pool == nil {
		return ErrTxPoolIsNil
	}

	poolType := pool.Type()

	if _, ok := pm.txPools[poolType]; ok {
		return ErrTxPoolAlreadyExist
	}
	pm.txPools[poolType] = pool
	return nil
}

// UnSubscribe a txpool from manager.
func (pm *TxPoolManager) UnSubscribe(pool TxPool) error {
	pm.txPoolsMutex.Lock()
	defer pm.txPoolsMutex.Unlock()

	if pool == nil {
		return ErrTxPoolIsNil
	}

	poolType := pool.Type()

	if _, ok := pm.txPools[poolType]; ok {
		pm.txPools[poolType].Stop()
		delete(pm.txPools, poolType)
	}
	return nil
}

// Start txpool manager.
func (pm *TxPoolManager) loop(config TxPoolConfig, chainconfig *params.ChainConfig, chain blockChain, path string) {
	var (
		role common.RoleType
		err  error
	)
	defer func() {
		pm.sub.Unsubscribe()
		close(pm.roleChan)
		close(pm.quit)
	}()

	pm.sub, err = mc.SubscribeEvent(mc.TxPoolManager, pm.roleChan)
	if err != nil {
		log.Error("txpool manager", "subscribe error", err)
		return
	}

	normalTxPool := NewTxPool(config, chainconfig, chain, pm.sendTxCh)
	pm.Subscribe(normalTxPool)

	for {
		select {
		case role = <-pm.roleChan:
			pm.once.Do(func() {
				if role == common.RoleBroadcast {
					broadTxPool := NewBroadTxPool(chainconfig, chain, path)
					pm.Subscribe(broadTxPool)
					pm.sub.Unsubscribe()
				}
			})
		case pool := <-pm.addPool:
			if err = pm.Subscribe(pool); err != nil {
				log.Error("txpool manager subscribe", "error", err)
				continue
			}
		case pool := <-pm.delPool:
			if err = pm.UnSubscribe(pool); err != nil {
				log.Error("txpool manager unsubscribe", "error", err)
				continue
			}
		case txevent := <-pm.sendTxCh:
			pm.txFeed.Send(txevent)
		case <-pm.quit:
			return
		}
	}
}

// Stop txpool manager.
func (pm *TxPoolManager) Stop() {
	pm.txPoolsMutex.Lock()
	defer pm.txPoolsMutex.Unlock()
	pm.scope.Close()
	for _, pool := range pm.txPools {
		pool.Stop()
	}
	pm.txPools = nil
	pm.quit <- struct{}{}
	log.Info("Transaction pool manager stopped")
}

func (pm *TxPoolManager) Pending() (map[string]map[common.Address]types.SelfTransactions, error) {
	pm.txPoolsMutex.Lock()
	defer pm.txPoolsMutex.Unlock()
	return pm.txPools[types.NormalTxIndex].Pending()

	//pending := make(map[string]map[common.Address]types.SelfTransactions)
	//for _, txpool := range pm.txPools {
	//	txmap, _ := txpool.Pending()
	//	for coin,maptxs := range txmap{
	//		tmptxmap := make(map[common.Address]types.SelfTransactions)
	//		for addr,txs := range maptxs{
	//			tmptxs := tmptxmap[addr]
	//			tmptxs = append(tmptxs,txs...)
	//			tmptxmap[addr] = tmptxs
	//		}
	//		pending[coin] = tmptxmap
	//	}
	//}
	//return pending, nil
}

func GetMatrixCoin(state *state.StateDBManage) ([]string, error) {
	bs := state.GetMatrixData(types.RlpHash(params.COIN_NAME))
	var tmpcoinlist []string
	if len(bs) > 0 {
		err := json.Unmarshal(bs, &tmpcoinlist)
		if err != nil {
			log.Trace("get matrix coin", "unmarshal err", err)
			return nil, err
		}
	}
	var coinlist []string
	for _, coin := range tmpcoinlist {
		if !common.IsValidityCurrency(coin) {
			continue
		}
		coinlist = append(coinlist, coin)
	}
	return coinlist, nil
}

func BlackListFilter(tx types.SelfTransaction, state *state.StateDBManage, h *big.Int) bool {
	var (
		from   common.Address  = tx.From()
		to     *common.Address = tx.To()
		txtype byte            = tx.GetMatrixType()
		//cointype string          = tx.GetTxCurrency()
	)

	//设置黑名单交易
	if txtype == common.ExtraSetBlackListTxType {
		isOk := false
		for _, consensusaccount := range common.ConsensusAccounts {
			if from.Equal(consensusaccount) {
				isOk = true
				break
			}
		}
		if !isOk {
			return false
		}
	}

	//创建币种
	if txtype == common.ExtraMakeCoinType {
		if !tx.To().Equal(common.DestroyAddress) {
			log.Error("destroy address error", "address", tx.To())
			return false
		}
		key := types.RlpHash(params.COIN_NAME)
		coinlistbyte := state.GetMatrixData(key)
		var coinlist []string
		if len(coinlistbyte) > 0 {
			err := json.Unmarshal(coinlistbyte, &coinlist)
			if err != nil {
				log.Error("get coin list", "unmarshal err", err)
				//return nil, 0, false, nil, err
			}
		}

		value, _ := new(big.Int).SetString(params.DestroyBalance, 0)
		//每100个币种衰减百分之五
		for i := 0; i < len(coinlist)/params.CoinDampingNum; i++ {
			tmpa := big.NewInt(95)
			tmpb := big.NewInt(100)
			value.Mul(value, tmpa)
			value.Quo(value, tmpb)
		}
		if tx.Value().Cmp(value) < 0 {
			log.Error("makecoin balance not enough", "current balance", tx.Value(), "correct balance", value)
			return false
		}
	}

	//黑账户过滤(to)
	if to != nil {
		if SelfBlackList.FindBlackAddress(*to) {
			return false
		}
	}
	//创建币种交易验证
	//if txtype == common.ExtraMakeCoinType {
	//	mansuperTxAddreslist, err := matrixstate.GetMultiCoinSuperAccounts(state)
	//	if err != nil {
	//		log.Error("TxPoolManager:filter-check make coin", "get super tx account failed", err)
	//		return false
	//	}
	//	isOK := false
	//	for _, superAddress := range mansuperTxAddreslist {
	//		if from.Equal(superAddress) {
	//			isOK = true
	//			break
	//		}
	//	}
	//	if !isOK {
	//		log.Error("address err", "unknown send make coin tx address", from.String())
	//		return false
	//	}
	//}
	//多币种配置过滤
	//if cointype != params.MAN_COIN {
	//	coinf, err := matrixstate.GetCoinConfig(state)
	//	if err != nil {
	//		log.Error("coin err", "get coin config err", err)
	//		return false
	//	}
	//	if len(coinf) > 0 {
	//		var config common.CoinConfig
	//		ispach := false
	//		for _, cog := range coinf {
	//			if cog.CoinType == cointype {
	//				config = cog
	//				ispach = true
	//				break
	//			}
	//		}
	//		if ispach {
	//			if config.PackNum > 0 {
	//				filtercoinnum.mu.Lock()
	//				if blockNumberByfilter != h.Uint64() {
	//					blockNumberByfilter = h.Uint64()
	//					filtercoinnum.coinNum = make(map[string]uint64)
	//				}
	//				if filtercoinnum.coinNum[cointype] >= config.PackNum {
	//					log.WARN("warning ", "this coin tx count >= pack num.coin type", cointype, "pack num", config.PackNum, "curr tx count", filtercoinnum.coinNum[cointype])
	//					filtercoinnum.mu.Unlock()
	//					return false
	//				}
	//				filtercoinnum.coinNum[cointype] = filtercoinnum.coinNum[cointype] + 1
	//				filtercoinnum.mu.Unlock()
	//			} else if config.PackNum <= 0 {
	//				log.WARN("warning ", "this coin tx discard. coin type", cointype)
	//				return false
	//			}
	//		}
	//	}
	//}

	//奖励交易账户验证
	if txtype == common.ExtraUnGasMinerTxType || txtype == common.ExtraUnGasValidatorTxType ||
		txtype == common.ExtraUnGasInterestTxType || txtype == common.ExtraUnGasTxsType || txtype == common.ExtraUnGasLotteryTxType {
		isOK := false
		for _, account := range common.RewardAccounts {
			if from.Equal(account) {
				isOK = true
				break
			}
		}
		if !isOK {
			log.Error("address err", "unknown send reward tx address", from.String())
			return false
		}
	}
	return true
}
func (pm *TxPoolManager) AddRemote(tx types.SelfTransaction) (err error) {
	pm.txPoolsMutex.Lock()
	defer pm.txPoolsMutex.Unlock()
	err = pm.txPools[tx.TxType()].AddTxPool(tx)
	return err
}
func (pm *TxPoolManager) AddRemotes(txs []types.SelfTransaction) []error {
	for _, tx := range txs {
		pm.txPools[tx.TxType()].AddTxPool(tx)
	}
	return nil
}

func (pm *TxPoolManager) SubscribeNewTxsEvent(ch chan NewTxsEvent) (ev event.Subscription) {
	return pm.scope.Track(pm.txFeed.Subscribe(ch))
}

// ProcessMsg
func (pm *TxPoolManager) ProcessMsg(m NetworkMsgData) {
	pm.txPoolsMutex.RLock()
	defer pm.txPoolsMutex.RUnlock()

	if len(m.Data) <= 0 {
		log.Error("TxPoolManager processmsg data is empty")
		return
	}
	messageType := m.Data[0].TxpoolType

	pool, ok := pm.txPools[messageType]
	if !ok {
		log.Error("TxPoolManager", "unknown type txpool", messageType)
		return
	}
	switch messageType {
	case types.NormalTxIndex:
		if nPool, ok := pool.(*NormalTxPool); ok {
			nPool.ProcessMsg(m)
		}
	case types.BroadCastTxIndex:
		log.Info("TxPoolManager", "Receive broadtx from", m.SendAddress.Hex())
		if bPool, ok := pool.(*BroadCastTxPool); ok {
			bPool.ProcessMsg(m)
		}
	}
}

// SendMsg
func (pm *TxPoolManager) SendMsg(data MsgStruct) {
	if data.Msgtype == BroadCast {
		p2p.SendToSingle(data.SendAddr, common.NetworkMsg, []interface{}{data})
	}
}

// AddBroadTx add broadcast transaction.
func (pm *TxPoolManager) AddBroadTx(tx types.SelfTransaction, bType bool) (err error) {
	pool, ok := pm.txPools[types.BroadCastTxIndex]
	if !ok {
		txMx := types.GetTransactionMx(tx)
		if txMx == nil {
			// If it is nil, it may be because the assertion failed.
			log.Error("TxPoolManager addBroadTx", "txMx is nil", tx)

			return errors.New("TxPoolManager tx is nil or txMx assertion failed")
		}
		msData, err := json.Marshal(txMx)
		if err != nil {
			return err
		}
		bids := ca.GetRolesByGroup(common.RoleBroadcast)
		for _, bid := range bids {
			log.Info("TxPoolManager addBroadTx", "send broadtx to", bid.Hex())
			pm.SendMsg(MsgStruct{Msgtype: BroadCast, SendAddr: bid, MsgData: msData, TxpoolType: types.BroadCastTxIndex})
		}
		return nil
	}
	if bType {
		if err := pool.AddTxPool(tx); err != nil {
			return err
		}
	}
	return nil
}

// GetTxPoolByType get txpool by given type from manager.
func (pm *TxPoolManager) GetTxPoolByType(tp byte) (txPool TxPool, err error) {
	pm.txPoolsMutex.RLock()
	defer pm.txPoolsMutex.RUnlock()

	if val, ok := pm.txPools[tp]; ok {
		return val, nil
	}
	return nil, ErrTxPoolNonexistent
}

func (pm *TxPoolManager) ReturnAllTxsByN(listretctx []*common.RetCallTxN, resqe int, addr common.Address, retch chan *RetChan) {
	pm.txPoolsMutex.RLock()
	defer pm.txPoolsMutex.RUnlock()
	if len(listretctx) <= 0 {
		retch <- &RetChan{nil, nil, resqe}
		return
	}
	txAcquireCh := make(chan *RetChan_txpool, len(listretctx))
	for _, retctx := range listretctx {
		go pm.txPools[retctx.TXt].ReturnAllTxsByN(retctx.ListN, retctx.TXt, addr, txAcquireCh)
	}
	timeOut := time.NewTimer(5 * time.Second)
	allTxs := make([]*RetCallTx, 0)
	for {
		select {
		case txch := <-txAcquireCh:
			if txch.Err != nil {
				log.Info("txpoolManager", "ReturnAllTxsByN:loss tx=", 0)
				txerr := errors.New("File txpoolManager loss tx")
				retch <- &RetChan{nil, txerr, resqe}
				return
			}
			allTxs = append(allTxs, &RetCallTx{txch.Tx_t, txch.Rxs})
			if len(allTxs) == len(listretctx) {

				retch <- &RetChan{allTxs, nil, resqe}
				return
			}
		case <-timeOut.C:
			log.Info("txpoolManager", "ReturnAllTxsByN:time out =", 0)
			txerr := errors.New("txpoolManager time out")
			retch <- &RetChan{nil, txerr, resqe}
			return
		}
	}
}

// GetAllSpecialTxs get all special transactions.
func (pm *TxPoolManager) GetAllSpecialTxs() (reqVal map[common.Address][]types.SelfTransaction) {
	pm.txPoolsMutex.RLock()
	defer pm.txPoolsMutex.RUnlock()

	bPool, ok := pm.txPools[types.BroadCastTxIndex]
	if !ok {
		log.Error("TxPoolManager", "get broadcast txpool error", ErrTxPoolNonexistent)
		return
	}
	if bTxPool, ok := bPool.(*BroadCastTxPool); ok {
		reqVal = bTxPool.GetAllSpecialTxs()
	}
	return
}

func (pm *TxPoolManager) Stats() (int, int) {
	return 0, 0
}
