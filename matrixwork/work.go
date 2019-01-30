// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixwork

import (
	"errors"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/matrix/go-matrix/accounts/abi"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/hexutil"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
)

type ChainReader interface {
	StateAt(root common.Hash) (*state.StateDB, error)
	GetBlockByHash(hash common.Hash) *types.Block
	Engine(version []byte) consensus.Engine
	GetHeader(common.Hash, uint64) *types.Header
	Processor(version []byte) core.Processor
}
type txPoolReader interface {
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[common.Address]types.SelfTransactions, error)
}

var packagename string = "matrixwork"
var (
	depositDef = ` [{"constant": true,"inputs": [],"name": "getDepositList","outputs": [{"name": "","type": "address[]"}],"payable": false,"stateMutability": "view","type": "function"},
			{"constant": true,"inputs": [{"name": "addr","type": "address"}],"name": "getDepositInfo","outputs": [{"name": "","type": "uint256"},{"name": "","type": "bytes"},{"name": "","type": "uint256"}],"payable": false,"stateMutability": "view","type": "function"},
    		{"constant": false,"inputs": [{"name": "nodeID","type": "bytes"}],"name": "valiDeposit","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
    		{"constant": false,"inputs": [{"name": "nodeID","type": "bytes"}],"name": "minerDeposit","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
    		{"constant": false,"inputs": [],"name": "withdraw","outputs": [],"payable": false,"stateMutability": "nonpayable","type": "function"},
    		{"constant": false,"inputs": [],"name": "refund","outputs": [],"payable": false,"stateMutability": "nonpayable","type": "function"},
			{"constant": false,"inputs": [{"name": "addr","type": "address"}],"name": "interestAdd","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
			{"constant": false,"inputs": [{"name": "addr","type": "address"}],"name": "getinterest","outputs": [],"payable": false,"stateMutability": "payable","type": "function"}]`

	depositAbi, Abierr = abi.JSON(strings.NewReader(depositDef))
)

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

	header *types.Header
	bc     ChainReader

	txs      []types.SelfTransaction
	Receipts []*types.Receipt

	createdAt time.Time
}
type coingasUse struct {
	mapcoin  map[string]*big.Int
	mapprice map[string]*big.Int
	mu       sync.RWMutex
}

var mapcoingasUse coingasUse = coingasUse{mapcoin: make(map[string]*big.Int), mapprice: make(map[string]*big.Int)}

func (cu *coingasUse) setCoinGasUse(txer types.SelfTransaction, gasuse uint64) {
	cu.mu.Lock()
	defer cu.mu.Unlock()
	gasAll := new(big.Int).SetUint64(gasuse)
	priceAll := txer.GasPrice()
	if gas, ok := cu.mapcoin[txer.GetTxCurrency()]; ok {
		gasAll = new(big.Int).Add(gasAll, gas)
	}
	cu.mapcoin[txer.GetTxCurrency()] = gasAll

	if _, ok := cu.mapprice[txer.GetTxCurrency()]; !ok {
		cu.mapprice[txer.GetTxCurrency()] = priceAll
	}
}
func (cu *coingasUse) getCoinGasPrice(typ string) *big.Int {
	cu.mu.Lock()
	defer cu.mu.Unlock()
	price, ok := cu.mapprice[typ]
	if !ok {
		price = new(big.Int).SetUint64(0)
	}
	return price
}

func (cu *coingasUse) getCoinGasUse(typ string) *big.Int {
	cu.mu.Lock()
	defer cu.mu.Unlock()
	gas, ok := cu.mapcoin[typ]
	if !ok {
		gas = new(big.Int).SetUint64(0)
	}
	return gas
}
func (cu *coingasUse) clearmap() {
	cu.mu.Lock()
	defer cu.mu.Unlock()
	cu.mapcoin = make(map[string]*big.Int)
	cu.mapprice = make(map[string]*big.Int)
}
func NewWork(config *params.ChainConfig, bc ChainReader, gasPool *core.GasPool, header *types.Header) (*Work, error) {

	Work := &Work{
		config:  config,
		signer:  types.NewEIP155Signer(config.ChainId),
		gasPool: gasPool,
		header:  header,
		bc:      bc,
	}
	var err error

	Work.State, err = bc.StateAt(bc.GetBlockByHash(header.ParentHash).Root())

	if err != nil {
		return nil, err
	}
	return Work, nil
}

//func (env *Work) commitTransactions(mux *event.TypeMux, txs *types.TransactionsByPriceAndNonce, bc *core.BlockChain, coinbase common.Address) (listN []uint32, retTxs []types.SelfTransaction) {
func (env *Work) commitTransactions(mux *event.TypeMux, txser types.SelfTransactions, coinbase common.Address) (listret []*common.RetCallTxN, retTxs []types.SelfTransaction) {
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(env.header.GasLimit)
	}

	var coalescedLogs []*types.Log
	tmpRetmap := make(map[byte][]uint32)
	for _, txer := range txser {
		// If we don't have enough gas for any further transactions then we're done
		if env.gasPool.Gas() < params.TxGas {
			log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
			break
		}
		if txer.GetTxNLen() == 0 {
			log.Info("file work func commitTransactions err: tx.N is nil")
			continue
		}
		// We use the eip155 signer regardless of the current hf.
		from, _ := txer.GetTxFrom()

		// Start executing the transaction
		env.State.Prepare(txer.Hash(), common.Hash{}, env.tcount)
		err, logs := env.commitTransaction(txer, env.bc, coinbase, env.gasPool)
		switch err {
		case core.ErrGasLimitReached:
			// Pop the current out-of-gas transaction without shifting in the next from the account
			log.Trace("Gas limit exceeded for current block", "sender", from)
		case core.ErrNonceTooLow:
			// New head notification data race between the transaction pool and miner, shift
			log.Trace("Skipping transaction with low nonce", "sender", from, "nonce", txer.Nonce())
		case core.ErrNonceTooHigh:
			// Reorg notification data race between the transaction pool and miner, skip account =
			log.Trace("Skipping account with hight nonce", "sender", from, "nonce", txer.Nonce())
		case nil:
			// Everything ok, collect the logs and shift in the next transaction from the same account
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
			coalescedLogs = append(coalescedLogs, logs...)
			env.tcount++
		default:
			// Strange error, discard the transaction and get the next in line (note, the
			// nonce-too-high clause will prevent us from executing in vain).
			log.Debug("Transaction failed, account skipped", "hash", txer.Hash(), "err", err)
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

func (env *Work) commitTransaction(tx types.SelfTransaction, bc ChainReader, coinbase common.Address, gp *core.GasPool) (error, []*types.Log) {
	snap := env.State.Snapshot()
	receipt, _, err := core.ApplyTransaction(env.config, bc, &coinbase, gp, env.State, env.header, tx, &env.header.GasUsed, vm.Config{})
	if err != nil && err != core.ErrSpecialTxFailed {
		log.Info("file work", "func commitTransaction", err)
		env.State.RevertToSnapshot(snap)
		return err, nil
	}
	env.txs = append(env.txs, tx)
	env.Receipts = append(env.Receipts, receipt)
	mapcoingasUse.setCoinGasUse(tx, receipt.GasUsed)
	return nil, receipt.Logs
}
func (env *Work) s_commitTransaction(tx types.SelfTransaction, coinbase common.Address, gp *core.GasPool) (error, []*types.Log) {
	env.State.Prepare(tx.Hash(), common.Hash{}, env.tcount)
	snap := env.State.Snapshot()
	receipt, _, err := core.ApplyTransaction(env.config, env.bc, &coinbase, gp, env.State, env.header, tx, &env.header.GasUsed, vm.Config{})
	if err != nil && err != core.ErrSpecialTxFailed {
		log.Info("file work", "func s_commitTransaction", err)
		env.State.RevertToSnapshot(snap)
		return err, nil
	}
	tmps := make([]types.SelfTransaction, 0)
	tmps = append(tmps, tx)
	tmps = append(tmps, env.txs...)
	env.txs = tmps

	tmpr := make([]*types.Receipt, 0)
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

func (env *Work) ProcessTransactions(mux *event.TypeMux, tp txPoolReader, upTime map[common.Address]uint64) (listret []*common.RetCallTxN, originalTxs []types.SelfTransaction, finalTxs []types.SelfTransaction) {
	pending, err := tp.Pending()
	if err != nil {
		log.Error("Failed to fetch pending transactions", "err", err)
		return nil, nil, nil
	}
	mapcoingasUse.clearmap()
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	listTx := make(types.SelfTransactions, 0)
	for _, txser := range pending {
		listTx = append(listTx, txser...)
	}
	listret, originalTxs = env.commitTransactions(mux, listTx, common.Address{})
	finalTxs = append(finalTxs, originalTxs...)
	tmps := make([]types.SelfTransaction, 0)
	from := make([]common.Address, 0)
	for _, tx := range originalTxs {
		from = append(from, tx.From())
	}
	rewart := env.bc.Processor(env.header.Version).ProcessReward(env.State, env.header, upTime, from, mapcoingasUse.getCoinGasUse("MAN").Uint64())
	txers := env.makeTransaction(rewart)
	for _, tx := range txers {
		err, _ := env.s_commitTransaction(tx, common.Address{}, new(core.GasPool).AddGas(0))
		if err != nil {
			log.Error("file work", "func ProcessTransactions:::reward Tx call Error", err)
			continue
		}
		tmptxs := make([]types.SelfTransaction, 0)
		tmptxs = append(tmptxs, tx)
		tmptxs = append(tmptxs, tmps...)
		tmps = tmptxs
	}
	tmps = append(tmps, finalTxs...)
	finalTxs = tmps
	return
}

func (env *Work) makeTransaction(rewarts []common.RewarTx) (txers []types.SelfTransaction) {
	for _, rewart := range rewarts {
		sorted_keys := make([]string, 0)
		for k, _ := range rewart.To_Amont {
			sorted_keys = append(sorted_keys, k.String())
		}
		sort.Strings(sorted_keys)
		extra := make([]*types.ExtraTo_tr, 0)
		var to common.Address
		var value *big.Int
		databytes := make([]byte, 0)
		isfirst := true
		for _, addr := range sorted_keys {
			k := common.HexToAddress(addr)
			v := rewart.To_Amont[k]
			if isfirst {
				if rewart.RewardTyp == common.RewardInerestType {
					if k != common.ContractAddress {
						databytes = append(databytes, depositAbi.Methods["interestAdd"].Id()...)
						tmpbytes, _ := depositAbi.Methods["interestAdd"].Inputs.Pack(k)
						databytes = append(databytes, tmpbytes...)
						to = common.ContractAddress
						value = v
					} else {
						continue
					}
				} else {
					to = k
					value = v
				}
				isfirst = false
				continue
			}
			tmp := new(types.ExtraTo_tr)
			vv := new(big.Int).Set(v)
			var kk common.Address = k
			tmp.To_tr = &kk
			tmp.Value_tr = (*hexutil.Big)(vv)
			if rewart.RewardTyp == common.RewardInerestType {
				if kk != common.ContractAddress {
					bytes := make([]byte, 0)
					bytes = append(bytes, depositAbi.Methods["interestAdd"].Id()...)
					tmpbytes, _ := depositAbi.Methods["interestAdd"].Inputs.Pack(k)
					bytes = append(bytes, tmpbytes...)
					b := hexutil.Bytes(bytes)
					tmp.Input_tr = &b
					tmp.To_tr = &common.ContractAddress
				} else {
					continue
				}
			}
			extra = append(extra, tmp)
		}

		tx := types.NewTransactions(env.State.GetNonce(rewart.Fromaddr), to, value, 0, new(big.Int), databytes,nil,nil,nil, extra, 0, common.ExtraUnGasTxType, 0,rewart.CoinType,0)
		tx.SetFromLoad(rewart.Fromaddr)
		tx.SetTxS(big.NewInt(1))
		tx.SetTxV(big.NewInt(1))
		tx.SetTxR(big.NewInt(1))
		tx.SetTxCurrency(rewart.CoinType)
		txers = append(txers, tx)
	}

	return
}

//Broadcast
func (env *Work) ProcessBroadcastTransactions(mux *event.TypeMux, txs []types.SelfTransaction) {
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	mapcoingasUse.clearmap()
	for _, tx := range txs {
		env.commitTransaction(tx, env.bc, common.Address{}, nil)
	}

	rewart := env.bc.Processor(env.header.Version).ProcessReward(env.State, env.header, nil, nil, mapcoingasUse.getCoinGasUse("MAN").Uint64())
	txers := env.makeTransaction(rewart)
	for _, tx := range txers {
		err, _ := env.s_commitTransaction(tx, common.Address{}, new(core.GasPool).AddGas(0))
		if err != nil {
			log.Error("file work", "func ProcessTransactions:::reward Tx call Error", err)
		}
	}
	return
}

func (env *Work) ConsensusTransactions(mux *event.TypeMux, txs []types.SelfTransaction, upTime map[common.Address]uint64) error {
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(env.header.GasLimit)
	}
	mapcoingasUse.clearmap()
	var coalescedLogs []*types.Log
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	from := make([]common.Address, 0)
	for _, tx := range txs {
		// If we don't have enough gas for any further transactions then we're done
		if env.gasPool.Gas() < params.TxGas {
			log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
			return errors.New("Not enough gas for further transactions")
		}

		// Start executing the transaction
		env.State.Prepare(tx.Hash(), common.Hash{}, env.tcount)
		err, logs := env.commitTransaction(tx, env.bc, common.Address{}, env.gasPool)
		if err == nil {
			env.tcount++
			coalescedLogs = append(coalescedLogs, logs...)
		} else {
			return err
		}
		from = append(from, tx.From())
	}

	rewart := env.bc.Processor(env.header.Version).ProcessReward(env.State, env.header, upTime, from, mapcoingasUse.getCoinGasUse("MAN").Uint64())
	txers := env.makeTransaction(rewart)
	for _, tx := range txers {
		err, _ := env.s_commitTransaction(tx, common.Address{}, new(core.GasPool).AddGas(0))
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
func (env *Work) GetTxs() []types.SelfTransaction {
	return env.txs
}
