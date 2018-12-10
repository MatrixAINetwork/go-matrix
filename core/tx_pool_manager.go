package core

import (
	"errors"
	"sync"

	"fmt"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/ca"
)

var (
	ErrTxPoolAlreadyExist = errors.New("txpool already exist")
	ErrTxPoolIsNil        = errors.New("txpool is nil")
	ErrTxPoolNonexistent  = errors.New("txpool nonexistent")
)
//YY
type RetChan struct {
	Rxs   []types.SelfTransaction
	Err   error
	Resqe int
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
	txPools      map[types.TxTypeInt]TxPool
	roleChan     chan common.RoleType
	quit         chan struct{}
	addPool      chan TxPool
	delPool      chan TxPool
}

func NewTxPoolManager(config TxPoolConfig, chainconfig *params.ChainConfig, chain blockChain, path string) *TxPoolManager {
	txPoolManager := &TxPoolManager{
		txPoolsMutex: sync.RWMutex{},
		txPools:      make(map[types.TxTypeInt]TxPool),
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

//TODO 返回的交易可能需要用map存储，因为会有多个pending，每个pending中存放的都是不同类型交易
func (pm *TxPoolManager) Pending() (map[common.Address]types.SelfTransactions, error) {
	//TODO 循环遍历所有的交易池
	for _,txpool := range pm.txPools{
		txpool.Pending()
	}
	return nil, nil
}

func (pm *TxPoolManager) AddRemotes(txs []types.SelfTransaction) []error {
	for _, tx := range txs {
		//TODO 根据交易类型判断调用哪个池的实例
		fmt.Println(tx)
	}
	return nil
}

func (pm *TxPoolManager) SubscribeNewTxsEvent(ch chan NewTxsEvent) (ev event.Subscription) {
	t := <-ch
	pm.txPoolsMutex.RLock()
	bpool, ok := pm.txPools[t.poolType]
	pm.txPoolsMutex.RUnlock()
	if !ok {
		log.Error("TxPoolManager", "unknown type txpool (SubscribeNewTxsEvent)", t.poolType)
		return nil
	}
	ev = bpool.SubscribeNewTxsEvent(ch)
	return ev
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
func (pm *TxPoolManager) GetTxPoolByType(tp types.TxTypeInt) (txPool TxPool, err error) {
	pm.txPoolsMutex.RLock()
	defer pm.txPoolsMutex.RUnlock()

	if val, ok := pm.txPools[tp]; ok {
		return val, nil
	}
	return nil, ErrTxPoolNonexistent
}

//TODO 根据交易类型判断调用哪个池的实例(如何判断N对应的是哪个交易池，N的编号是否应该再加上交易类型)
func (pm *TxPoolManager) ReturnAllTxsByN(listN []uint32, resqe int, addr common.Address, retch chan *RetChan) {
	pm.txPoolsMutex.RLock()
	npool, ok := pm.txPools[types.NormalTxIndex].(*NormalTxPool)
	pm.txPoolsMutex.RUnlock()
	if ok {
		npool.ReturnAllTxsByN(listN, resqe, addr, retch)
	} else {
		log.Error("TxPoolManager:ReturnAllTxsByN unknown txpool")
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

//TODO 根据交易类型判断调用哪个池的实例
func (pm *TxPoolManager) Stats() (int, int) {
	return 0, 0
}
