package core

import (
	"errors"
	"sync"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/ca"
	"time"
)

var (
	ErrTxPoolAlreadyExist = errors.New("txpool already exist")
	ErrTxPoolIsNil        = errors.New("txpool is nil")
	ErrTxPoolNonexistent  = errors.New("txpool nonexistent")
)
//YY
type RetChan struct {
	//Rxs   []types.SelfTransaction
	AllTxs []*RetCallTx
	Err   error
	Resqe int
}
type RetChan_txpool struct {
	Rxs   []types.SelfTransaction
	Err   error
	Tx_t common.TxTypeInt
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


// TxPoolManager
type TxPoolManager struct {
	txPoolsMutex sync.RWMutex
	once         sync.Once
	sub          event.Subscription
	txPools      map[common.TxTypeInt]TxPool
	roleChan     chan common.RoleType
	quit         chan struct{}
	addPool      chan TxPool
	delPool      chan TxPool
}

func NewTxPoolManager(config TxPoolConfig, chainconfig *params.ChainConfig, chain blockChain, path string) *TxPoolManager {
	txPoolManager := &TxPoolManager{
		txPoolsMutex: sync.RWMutex{},
		txPools:      make(map[common.TxTypeInt]TxPool),
		quit:         make(chan struct{}),
		roleChan:     make(chan common.RoleType),
		addPool:      make(chan TxPool),
		delPool:      make(chan TxPool),
	}
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
		log.Error("txpool manage", "subscribe error", err)
		return
	}

	normalTxPool := NewTxPool(config, chainconfig, chain)
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
		case <-pm.quit:
			return
		}
	}
}

// Stop txpool manager.
func (pm *TxPoolManager) Stop() {
	pm.txPoolsMutex.Lock()
	defer pm.txPoolsMutex.Unlock()

	for _, pool := range pm.txPools {
		pool.Stop()
	}
	pm.txPools = nil
	pm.quit <- struct{}{}
	log.Info("Transaction pool manager stopped")
}

func (pm *TxPoolManager) Pending() (map[common.Address]types.SelfTransactions, error) {
	pm.txPoolsMutex.Lock()
	defer pm.txPoolsMutex.Unlock()
	txser := make(map[common.Address]types.SelfTransactions)
	for _,txpool := range pm.txPools{
		txmap,_ := txpool.Pending()
		for addr,txs:=range txmap{
			if txlist,ok:=txser[addr];ok{
				txlist = append(txlist, txs...)
				txser[addr] = txlist
			}else {
				txser[addr] = txs
			}
		}
	}
	return txser, nil
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
	//if len(ch) <= 0{
	//	return nil
	//}
	//t := <-ch
	//TODO
	pm.txPoolsMutex.RLock()
	defer pm.txPoolsMutex.RUnlock()
	return pm.txPools[types.NormalTxIndex].SubscribeNewTxsEvent(ch)
}

// ProcessMsg
func (pm *TxPoolManager) ProcessMsg(m NetworkMsgData) {
	pm.txPoolsMutex.RLock()
	defer pm.txPoolsMutex.RUnlock()

	if len(m.Data) <= 0 {
		log.Error("TxPoolManager", "ProcessMsg", "data is empty")
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
		if bPool, ok := pool.(*BroadCastTxPool); ok {
			bPool.ProcessMsg(m)
		}
	}
}

// GetTxPoolByType get txpool by given type from manager.
func (pm *TxPoolManager) GetTxPoolByType(tp common.TxTypeInt) (txPool TxPool, err error) {
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
	if len(listretctx) <= 0{
		retch <- &RetChan{nil, nil, resqe}
		return
	}
	txAcquireCh := make(chan *RetChan_txpool, len(listretctx))
	for _,retctx := range listretctx{
		go pm.txPools[retctx.TXt].ReturnAllTxsByN(retctx.ListN, retctx.TXt, addr, txAcquireCh)
	}
	timeOut := time.NewTimer(5*time.Second)
	allTxs := make([]*RetCallTx,0)
	for {
		select {
		case txch := <- txAcquireCh:
			if txch.Err != nil{
				log.Info("File txpoolManager", "ReturnAllTxsByN:loss tx=", 0)
				txerr := errors.New("File txpoolManager loss tx")
				retch <- &RetChan{nil, txerr, resqe}
				return
			}
			allTxs = append(allTxs,&RetCallTx{txch.Tx_t,txch.Rxs})
			if len(allTxs) == len(listretctx){
				retch <- &RetChan{allTxs, nil, resqe}
				return
			}
		case <-timeOut.C:
			log.Info("File txpoolManager", "ReturnAllTxsByN:time out =", 0)
			txerr := errors.New("File txpoolManager time out")
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

func (pm *TxPoolManager)Stats()(int,int)  {
	return 0, 0
}