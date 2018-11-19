// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package matrixwork

import (
	"encoding/json"
	"errors"
	"github.com/matrix/go-matrix/params/manparams"
	"math/big"
	"time"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/hexutil"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
	"sort"
)

var packagename string = "matrixwork"

// Work is the workers current environment and holds
// all of the current state information
type Work struct {
	config *params.ChainConfig
	signer types.Signer

	State *state.StateDB // apply state changes here
	//ancestors *set.Set       // ancestor set (used for checking uncle parent validity)
	//family    *set.Set       // family set (used for checking uncle invalidity)
	//uncles    *set.Set       // uncle set
	tcount  int           // tx count in cycle
	gasPool *core.GasPool // available gas used to pack transactions

	Block *types.Block // the new block

	header   *types.Header
	txs      []types.SelfTransaction
	Receipts []*types.Receipt

	createdAt time.Time
}

func NewWork(config *params.ChainConfig, bc *core.BlockChain, gasPool *core.GasPool, header *types.Header) (*Work, error) {

	Work := &Work{
		config:  config,
		signer:  types.NewEIP155Signer(config.ChainId),
		gasPool: gasPool,
		header:  header,
	}
	var err error

	Work.State, err = bc.StateAt(bc.GetBlockByHash(header.ParentHash).Root())

	if err != nil {
		return nil, err
	}
	return Work, nil
}

//func (env *Work) commitTransactions(mux *event.TypeMux, txs *types.TransactionsByPriceAndNonce, bc *core.BlockChain, coinbase common.Address) (listN []uint32, retTxs []types.SelfTransaction) {
func (env *Work) commitTransactions(mux *event.TypeMux, txser types.SelfTransactions, bc *core.BlockChain, coinbase common.Address) (listret []*common.RetCallTxN, retTxs []types.SelfTransaction) {
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(env.header.GasLimit)
	}

	var coalescedLogs []*types.Log
	tmpRetmap := make(map[common.TxTypeInt][]uint32)
	for _, txer := range txser {
		// If we don't have enough gas for any further transactions then we're done
		if env.gasPool.Gas() < params.TxGas {
			log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
			break
		}
		// Retrieve the next transaction and abort if all done
		//tx := txs.Peek()
		//if tx == nil {
		//	break
		//}

		if txer.GetTxNLen() == 0 {
			log.Info("===========tx.N is nil")
			//txs.Pop()
			continue
		}
		// Error may be ignored here. The error has already been checked
		// during transaction acceptance is the transaction pool.
		//
		// We use the eip155 signer regardless of the current hf.
		from, _ := txer.GetTxFrom()

		// Check whether the tx is replay protected. If we're not in the EIP155 hf
		// phase, start ignoring the sender until we do.
		//YYY TODO 是否需要当前这个if
		if txer.Protected() && !env.config.IsEIP155(env.header.Number) {
			log.Trace("Ignoring reply protected transaction", "hash", txer.Hash(), "eip155", env.config.EIP155Block)

			//txs.Pop()
			continue
		}
		// Start executing the transaction
		env.State.Prepare(txer.Hash(), common.Hash{}, env.tcount)
		err, logs := env.commitTransaction(txer, bc, coinbase, env.gasPool)
		switch err {
		case core.ErrGasLimitReached:
			// Pop the current out-of-gas transaction without shifting in the next from the account
			log.Trace("Gas limit exceeded for current block", "sender", from)
			//txs.Pop()

		case core.ErrNonceTooLow:
			// New head notification data race between the transaction pool and miner, shift
			log.Trace("Skipping transaction with low nonce", "sender", from, "nonce", txer.Nonce())
			//txs.Shift()

		case core.ErrNonceTooHigh:
			// Reorg notification data race between the transaction pool and miner, skip account =
			log.Trace("Skipping account with hight nonce", "sender", from, "nonce", txer.Nonce())
			//txs.Pop()

		case nil:
			// Everything ok, collect the logs and shift in the next transaction from the same account
			//==========hezi===================
			if txer.GetTxNLen() != 0 {
				n := txer.GetTxN(0)
				if listN, ok := tmpRetmap[txer.TxType()]; ok {
					listN = append(listN, n)
					tmpRetmap[txer.TxType()] = listN
				} else {
					listN := make([]uint32, 0)
					listN = append(listN, n)
					tmpRetmap[txer.TxType()] = listN
				}
				retTxs = append(retTxs, txer)
			}
			//==================================
			coalescedLogs = append(coalescedLogs, logs...)
			env.tcount++
			//txs.Shift()

		default:
			// Strange error, discard the transaction and get the next in line (note, the
			// nonce-too-high clause will prevent us from executing in vain).
			log.Debug("Transaction failed, account skipped", "hash", txer.Hash(), "err", err)
			//txs.Shift()
		}
	}
	for t, n := range tmpRetmap {
		ts := common.RetCallTxN{t, n}
		listret = append(listret, &ts)
	}
	if len(coalescedLogs) > 0 || env.tcount > 0 {
		// make a copy, the state caches the logs and these logs get "upgraded" from pending to mined
		// logs by filling in the block hash when the block was mined by the local miner. This can
		// cause a race condition if a log was "upgraded" before the PendingLogsEvent is processed.
		cpy := make([]*types.Log, len(coalescedLogs))
		for i, l := range coalescedLogs {
			cpy[i] = new(types.Log)
			*cpy[i] = *l
		}
		go func(logs []*types.Log, tcount int) {
			if len(logs) > 0 {
				mux.Post(core.PendingLogsEvent{Logs: logs})
			}
			if tcount > 0 {
				mux.Post(core.PendingStateEvent{})
			}
		}(cpy, env.tcount)
	}
	return listret, retTxs
}

func (env *Work) commitTransaction(tx types.SelfTransaction, bc *core.BlockChain, coinbase common.Address, gp *core.GasPool) (error, []*types.Log) {
	snap := env.State.Snapshot()

	receipt, _, err := core.ApplyTransaction(env.config, bc, &coinbase, gp, env.State, env.header, tx, &env.header.GasUsed, vm.Config{})
	if err != nil {
		log.Info("*************", "ApplyTransaction:err", err)
		env.State.RevertToSnapshot(snap)
		return err, nil
	}
	env.txs = append(env.txs, tx)
	env.Receipts = append(env.Receipts, receipt)

	return nil, receipt.Logs
}
func (env *Work) s_commitTransaction(tx types.SelfTransaction, bc *core.BlockChain, coinbase common.Address, gp *core.GasPool) (error, []*types.Log) {
	env.State.Prepare(tx.Hash(), common.Hash{}, env.tcount)
	snap := env.State.Snapshot()
	receipt, _, err := core.ApplyTransaction(env.config, bc, &coinbase, gp, env.State, env.header, tx, &env.header.GasUsed, vm.Config{})
	if err != nil {
		log.Info("*************","ApplyTransaction:err",err)
		env.State.RevertToSnapshot(snap)
		return err, nil
	}
	tmps :=make([]types.SelfTransaction,0)
	tmps = append(tmps, tx)
	tmps = append(tmps, env.txs...)
	env.txs = tmps

	tmpr := make([]*types.Receipt,0)
	tmpr = append(tmpr, receipt)
	tmpr = append(tmpr, env.Receipts...)
	env.Receipts = tmpr
	env.tcount++
	return nil, receipt.Logs
}
//Leader
var lostCnt int = 0

type retStruct struct {
	no  []uint32
	txs []*types.Transaction
}

func (env *Work) ProcessTransactions(mux *event.TypeMux, tp *core.TxPoolManager, bc *core.BlockChain,blkRewar map[common.Address]*big.Int,gasRewar map[common.Address]*big.Int) (listret []*common.RetCallTxN, retTxs []types.SelfTransaction) {
	pending, err := tp.Pending()
	if err != nil {
		log.Error("Failed to fetch pending transactions", "err", err)
		return nil, nil
	}
	//log.INFO("===========", "ProcessTransactions:pending:", len(pending))
	listTx := make(types.SelfTransactions, 0)
	for _, txser := range pending {
		listTx = append(listTx, txser...)
	}
	listret,retTxs = env.commitTransactions(mux, listTx, bc, common.Address{})
	tmps :=make([]types.SelfTransaction,0)
	var (
		tx1 types.SelfTransaction
		tx2 types.SelfTransaction
	)
	if len(gasRewar) > 0{
		tx2 = env.makeTransaction(common.TxGasRewardAddress,gasRewar)//交易费奖励
		env.s_commitTransaction(tx2,bc,common.Address{},new(core.GasPool).AddGas(0))
	}

	if len(blkRewar) > 0{
		tx1 = env.makeTransaction(common.BlkRewardAddress,blkRewar) //区块奖励
		env.s_commitTransaction(tx1,bc,common.Address{},new(core.GasPool).AddGas(0))
	}


	if tx1 != nil{
		tmps = append(tmps, tx1)
	}
	if tx2 != nil {
		tmps = append(tmps, tx2)
	}
	tmps = append(tmps, retTxs...)
	retTxs = tmps
	return
}
func (env *Work)makeTransaction(from common.Address,val map[common.Address]*big.Int) types.SelfTransaction{
	sorted_keys := make([]string, 0)

	for k, _ := range val {
		sorted_keys = append(sorted_keys, k.String())
	}
	sort.Strings(sorted_keys)

	extra := make([]*types.ExtraTo_tr,0)
	var to common.Address
	var value *big.Int
	tmpv := new(big.Int).SetUint64(10000)
	tmpgas := new(big.Int).Mul(new(big.Int).SetUint64(env.header.GasUsed),new(big.Int).SetUint64(params.TxGasPrice))
	isfirst := true
	for _,addr := range sorted_keys{
		k :=common.HexToAddress(addr)
		v := val[k]
		if from == common.TxGasRewardAddress{
			v = new(big.Int).Mul(v, tmpgas)
			v = new(big.Int).Quo(v,tmpv)
		}
		if isfirst{
			to = k
			value = v
			isfirst = false
			continue
		}
		tmp := new(types.ExtraTo_tr)
		vv := new(big.Int).Set(v)
		var kk common.Address = k
		tmp.To_tr = &kk
		tmp.Value_tr = (*hexutil.Big)(vv)
		extra = append(extra, tmp)
	}
	tx := types.NewTransactions(env.State.GetNonce(from),to,value,0,new(big.Int),nil,extra,0,common.ExtraUnGasTxType)
	tx.SetFromLoad(from)
	tx.SetTxS(big.NewInt(1))
	tx.SetTxV(big.NewInt(1))
	tx.SetTxR(big.NewInt(1))
	return tx
}
//Broadcast
func (env *Work) ProcessBroadcastTransactions(mux *event.TypeMux, txs []types.SelfTransaction, bc *core.BlockChain,blkRewar map[common.Address]*big.Int,gasRewar map[common.Address]*big.Int) {
	for _, tx := range txs {
		env.commitTransaction(tx, bc, common.Address{}, nil)
	}
	if len(gasRewar) > 0{
		tx2 := env.makeTransaction(common.TxGasRewardAddress,gasRewar)//交易费奖励
		env.s_commitTransaction(tx2,bc,common.Address{},new(core.GasPool).AddGas(0))
	}
	if len(blkRewar) > 0{
		tx1 := env.makeTransaction(common.BlkRewardAddress,blkRewar) //区块奖励
		env.s_commitTransaction(tx1,bc,common.Address{},new(core.GasPool).AddGas(0))
	}
	return
}

func (env *Work) ConsensusTransactions(mux *event.TypeMux, txs []types.SelfTransaction, bc *core.BlockChain,blkRewar map[common.Address]*big.Int,gasRewar map[common.Address]*big.Int) error {
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(env.header.GasLimit)
	}
	var coalescedLogs []*types.Log

	for _, tx := range txs {
		// If we don't have enough gas for any further transactions then we're done
		if env.gasPool.Gas() < params.TxGas {
			log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
			return errors.New("Not enough gas for further transactions")
		}

		// Start executing the transaction
		env.State.Prepare(tx.Hash(), common.Hash{}, env.tcount)
		err, logs := env.commitTransaction(tx, bc, common.Address{}, env.gasPool)
		if err == nil {
			//log.Info("=========","ConsensusTransactions:tx.N",tx.N)
			env.tcount++
			coalescedLogs = append(coalescedLogs, logs...)
		} else {
			return err
		}
	}

	if len(gasRewar) > 0{
		tx2 := env.makeTransaction(common.TxGasRewardAddress,gasRewar)//交易费奖励
		err, _ :=env.s_commitTransaction(tx2,bc,common.Address{},new(core.GasPool).AddGas(0))
		if err != nil {
			return err
		}
	}

	if len(blkRewar) > 0{
		tx1 := env.makeTransaction(common.BlkRewardAddress,blkRewar) //区块奖励
		err,_:=env.s_commitTransaction(tx1,bc,common.Address{},new(core.GasPool).AddGas(0))
		if err != nil {
			return err
		}
	}
	if len(coalescedLogs) > 0 || env.tcount > 0 {
		// make a copy, the state caches the logs and these logs get "upgraded" from pending to mined
		// logs by filling in the block hash when the block was mined by the local miner. This can
		// cause a race condition if a log was "upgraded" before the PendingLogsEvent is processed.
		cpy := make([]*types.Log, len(coalescedLogs))
		for i, l := range coalescedLogs {
			cpy[i] = new(types.Log)
			*cpy[i] = *l
		}
		go func(logs []*types.Log, tcount int) {
			if len(logs) > 0 {
				mux.Post(core.PendingLogsEvent{Logs: logs})
			}
			if tcount > 0 {
				mux.Post(core.PendingStateEvent{})
			}
		}(cpy, env.tcount)
	}

	return nil
}
func (env *Work) GetTxs()[]types.SelfTransaction{
	return env.txs
}
func (env *Work) GetUpTimeAccounts(num uint64) ([]common.Address, error) {

	log.INFO(packagename, "获取所有参与uptime点名高度", num)

	upTimeAccounts := make([]common.Address, 0)

	minerNum := num - (num % common.GetBroadcastInterval()) - manparams.MinerTopologyGenerateUpTime
	log.INFO(packagename, "参选矿工节点uptime高度", minerNum)
	ans, err := ca.GetElectedByHeightAndRole(big.NewInt(int64(minerNum)), common.RoleMiner)
	if err != nil {
		return nil, err
	}

	for _, v := range ans {
		upTimeAccounts = append(upTimeAccounts, v.Address)
		log.INFO("packagename", "矿工节点账户", v.Address.Hex())
	}
	validatorNum := num - (num % common.GetBroadcastInterval()) - manparams.VerifyTopologyGenerateUpTime
	log.INFO(packagename, "参选验证节点uptime高度", validatorNum)
	ans1, err := ca.GetElectedByHeightAndRole(big.NewInt(int64(validatorNum)), common.RoleValidator)
	if err != nil {
		return upTimeAccounts, err
	}
	for _, v := range ans1 {
		upTimeAccounts = append(upTimeAccounts, v.Address)
		log.INFO("packagename", "验证者节点账户", v.Address.Hex())
	}
	return upTimeAccounts, nil
}

func (env *Work) GetUpTimeData(hash common.Hash) (map[common.Address]uint32, map[common.Address][]byte, error) {

	log.INFO(packagename, "获取所有心跳交易", "")
	//%99
	heatBeatUnmarshallMMap, error := core.GetBroadcastTxs(hash, mc.Heartbeat)
	if nil != error {
		log.WARN(packagename, "获取主动心跳交易错误", error)
	}
	//每个广播周期发一次
	calltherollUnmarshall, error := core.GetBroadcastTxs(hash, mc.CallTheRoll)
	if nil != error {
		log.ERROR(packagename, "获取点名心跳交易错误", error)
		return nil, nil, error
	}
	calltherollMap := make(map[common.Address]uint32, 0)
	for _, v := range calltherollUnmarshall {
		temp := make(map[string]uint32, 0)
		error := json.Unmarshal(v, &temp)
		if nil != error {
			log.ERROR(packagename, "序列化点名心跳交易错误", error)
			return nil, nil, error
		}
		log.INFO(packagename, "++++++++点名心跳交易++++++++", temp)
		for k, v := range temp {
			calltherollMap[common.HexToAddress(k)] = v
		}
	}
	return calltherollMap, heatBeatUnmarshallMMap, nil
}

func (env *Work) HandleUpTime(state *state.StateDB, accounts []common.Address, calltherollRspAccounts map[common.Address]uint32, heatBeatAccounts map[common.Address][]byte, blockNum uint64, bc *core.BlockChain) error {
	var blockHash common.Hash
	HeatBeatReqAccounts := make([]common.Address, 0)
	HeartBeatMap := make(map[common.Address]bool, 0)
	blockNumRem := blockNum % common.GetBroadcastInterval()

	//subVal就是最新的广播区块，例如当前区块高度是198或者是101，那么subVal就是100
	subVal := blockNum - blockNumRem
	//subVal就是最新的广播区块，例如当前区块高度是198或者是101，那么subVal就是100
	subVal = subVal
	if blockNum < common.GetBroadcastInterval() { //当前区块小于100说明是100区块内 (下面的if else是为了应对中途加入的参选节点)
		blockHash = bc.GetBlockByNumber(0).Hash() //创世区块的hash
	} else {
		blockHash = bc.GetBlockByNumber(subVal).Hash() //获取最近的广播区块的hash
	}
	// todo: remove
	//blockHash = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3e4")
	broadcastBlock := blockHash.Big()
	val := broadcastBlock.Uint64() % ((common.GetBroadcastInterval()) - 1)

	for _, v := range accounts {
		currentAcc := v.Big() //YY TODO 这里应该是广播账户。后期需要修改
		ret := currentAcc.Uint64() % (common.GetBroadcastInterval() - 1)
		if ret == val {
			HeatBeatReqAccounts = append(HeatBeatReqAccounts, v)
			if _, ok := heatBeatAccounts[v]; ok {
				HeartBeatMap[v] = true
			} else {
				HeartBeatMap[v] = false

			}
			log.Info(packagename, "计算主动心跳的账户", v, "心跳状态", HeartBeatMap[v])
		}
	}

	var upTime uint64
	originTopologyNum := blockNum - blockNum%common.GetBroadcastInterval() - 1
	log.Info(packagename, "获取原始拓扑图所有的验证者和矿工，高度为", originTopologyNum)
	originTopology, err := ca.GetTopologyByNumber(common.RoleValidator|common.RoleBackupValidator|common.RoleMiner|common.RoleBackupMiner, originTopologyNum)
	if err != nil {
		return err
	}
	originTopologyMap := make(map[common.Address]uint32, 0)
	for _, v := range originTopology.NodeList {
		originTopologyMap[v.Account] = 0
	}
	for _, account := range accounts {
		onlineBlockNum, ok := calltherollRspAccounts[account]
		if ok { //被点名,使用点名的uptime
			upTime = uint64(onlineBlockNum)
			log.INFO(packagename, "点名账号", account, "uptime", upTime)

		} else { //没被点名，没有主动上报，则为最大值，
			if v, ok := HeartBeatMap[account]; ok { //有主动上报
				if v {
					upTime = common.GetBroadcastInterval() - 3
					log.INFO(packagename, "没被点名，有主动上报有响应", account, "uptime", upTime)
				} else {
					upTime = 0
					log.INFO(packagename, "没被点名，有主动上报无响应", account, "uptime", upTime)
				}
			} else { //没被点名和主动上报
				upTime = common.GetBroadcastInterval() - 3
				log.INFO(packagename, "没被点名，没要求主动上报", account, "uptime", upTime)

			}
		}
		// todo: add
		depoistInfo.AddOnlineTime(state, account, new(big.Int).SetUint64(upTime))
		read, err := depoistInfo.GetOnlineTime(state, account)
		if nil == err {
			log.INFO(packagename, "读取状态树", account, "upTime减半", read)
			if _, ok := originTopologyMap[account]; ok {
				updateData := new(big.Int).SetUint64(read.Uint64() / 2)
				log.INFO(packagename, "是原始拓扑图节点，upTime减半", account, "upTime", updateData.Uint64())
				depoistInfo.AddOnlineTime(state, account, updateData)
			}
		}

	}

	return nil
}
