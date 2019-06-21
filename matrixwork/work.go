// Copyright (c) 2018 The MATRIX Authors
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

	"github.com/MatrixAINetwork/go-matrix/accounts/abi"
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
)

type ChainReader interface {
	StateAt(root []common.CoinRoot) (*state.StateDBManage, error)
	GetBlockByHash(hash common.Hash) *types.Block
	Engine(version []byte) consensus.Engine
	GetHeader(common.Hash, uint64) *types.Header
	Processor(version []byte) core.Processor
}
type txPoolReader interface {
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[string]map[common.Address]types.SelfTransactions, error)
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

	State *state.StateDBManage // apply state changes here
	//ancestors *set.Set       // ancestor set (used for checking uncle parent validity)
	//family    *set.Set       // family set (used for checking uncle invalidity)
	//uncles    *set.Set       // uncle set
	tcount  int           // tx count in cycle
	gasPool *core.GasPool // available gas used to pack transactions

	Block *types.Block // the new block

	header *types.Header
	bc     ChainReader

	random   *baseinterface.Random
	txs      []types.CoinSelfTransaction
	Receipts []types.CoinReceipts

	transer []types.SelfTransaction
	recpts  []*types.Receipt

	createdAt     time.Time
	packNum       int //MAN以外的其他币种打包数量限制（不能超过MAN的1/100）
	coinType      string
	mapcoingasUse coingasUse
}
type coingasUse struct {
	mapcoin  map[string]*big.Int
	mapprice map[string]*big.Int
	mu       sync.RWMutex
}

func (cu *coingasUse) setCoinGasUse(txer types.SelfTransaction, gasuse uint64) {
	cu.mu.Lock()
	defer cu.mu.Unlock()
	coin := txer.GetTxCurrency()
	//coin = params.MAN_COIN
	gasAll := new(big.Int).SetUint64(gasuse)
	priceAll := txer.GasPrice()
	if gas, ok := cu.mapcoin[coin]; ok {
		gasAll = new(big.Int).Add(gasAll, gas)
	}
	cu.mapcoin[coin] = gasAll
	if _, ok := cu.mapprice[coin]; !ok {
		cu.mapprice[coin] = priceAll
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
	Work.mapcoingasUse = coingasUse{mapcoin: make(map[string]*big.Int), mapprice: make(map[string]*big.Int)}
	var err error

	Work.State, err = bc.StateAt(bc.GetBlockByHash(header.ParentHash).Root())

	if err != nil {
		return nil, err
	}
	return Work, nil
}

func (env *Work) commitTransactions(mux *event.TypeMux, txser map[common.Address]types.SelfTransactions, coinbase common.Address) (listret []*common.RetCallTxN, retTxs []types.SelfTransaction) {
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(env.header.GasLimit)
	}

	coalescedLogs := make([]types.CoinLogs, 0, 1024)
	retTxs = make([]types.SelfTransaction, 0, 1024)
	tmpRetmap := make(map[byte][]uint32)
	isExceed := false
	for _, txers := range txser {
		if isExceed {
			break
		}
		//txs := types.GetCoinTX(txers)
		for _, txer := range txers {
			if uint64(env.packNum) >= params.OtherCoinPackNum && env.coinType != params.MAN_COIN {
				isExceed = true
				break
			}
			// If we don't have enough gas for any further transactions then we're done
			if env.gasPool.Gas() < params.TxGas {
				log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
				break
			}
			if txer.GetTxNLen() == 0 {
				log.Info("work.go commitTransactions err: tx.N is nil")
				continue
			}
			// We use the eip155 signer regardless of the current hf.
			from, _ := txer.GetTxFrom()

			// Start executing the transaction
			env.State.Prepare(txer.Hash(), common.Hash{}, env.tcount)
			err, logs := env.commitTransaction(txer, env.bc, coinbase, env.gasPool)
			isSkipFrom := false
			switch err {
			case core.ErrGasLimitReached:
				// Pop the current out-of-gas transaction without shifting in the next from the account
				log.Trace("Gas limit exceeded for current block", "sender", from)
				isSkipFrom = true
			case core.ErrNonceTooLow:
				// New head notification data race between the transaction pool and miner, shift
				log.Trace("Skipping transaction with low nonce", "sender", from, "nonce", txer.Nonce())
			case core.ErrNonceTooHigh:
				// Reorg notification data race between the transaction pool and miner, skip account =
				log.Trace("Skipping account with hight nonce", "sender", from, "nonce", txer.Nonce())
				isSkipFrom = true
			case core.ErrBlackListTx: //黑名单交易
				log.Trace("Skipping account in blacklist", "sender", from, "nonce", txer.Nonce())
				isSkipFrom = true
			case nil:
				// Everything ok, collect the logs and shift in the next transaction from the same account
				if txer.GetTxNLen() != 0 {
					n := txer.GetTxN(0)
					tmpRetmap[txer.TxType()] = append(tmpRetmap[txer.TxType()], n)
					/*
						if listN, ok := tmpRetmap[txer.TxType()]; ok {
							listN = append(listN, n)
							tmpRetmap[txer.TxType()] = listN
						} else {
							listN := make([]uint32, 1)
							listN[0] = n
							tmpRetmap[txer.TxType()] = listN
						}
					*/
					retTxs = append(retTxs, txer)

				}
				coalescedLogs = append(coalescedLogs, types.CoinLogs{txer.GetTxCurrency(), logs})
				env.tcount++
				env.packNum++
			default:
				// Strange error, discard the transaction and get the next in line (note, the
				// nonce-too-high clause will prevent us from executing in vain).
				log.Debug("Transaction failed, account skipped", "hash", txer.Hash(), "err", err)
			}
			if isSkipFrom {
				break
			}
		}
	}
	//env.State.Finalise("MAN",true)
	listret = make([]*common.RetCallTxN, 0, len(tmpRetmap))
	for t, n := range tmpRetmap {
		ts := common.RetCallTxN{t, n}
		listret = append(listret, &ts)
	}
	if len(coalescedLogs) > 0 || env.tcount > 0 {
		// make a copy, the state caches the logs and these logs get "upgraded" from pending to mined
		// logs by filling in the block hash when the block was mined by the local miner. This can
		// cause a race condition if a log was "upgraded" before the PendingLogsEvent is processed.
		cpy := make([]types.CoinLogs, len(coalescedLogs))
		for i, l := range coalescedLogs {
			cpy[i] = *new(types.CoinLogs)
			cpy[i] = l
		}
		go func(logs []types.CoinLogs, tcount int) {
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
func isInBlackList(from common.Address) bool {
	isOk := false
	for _, blackaccount := range common.BlackList {
		if from.Equal(blackaccount) {
			isOk = true
			break
		}
	}
	return isOk
}

func (env *Work) commitTransaction(tx types.SelfTransaction, bc ChainReader, coinbase common.Address, gp *core.GasPool) (error, []*types.Log) {
	//leader和follower过滤黑名单交易
	if isInBlackList(tx.From()) {
		log.Error("commitTransaction", "tx.from is in blacklist", tx.From().String())
		return core.ErrBlackListTx, nil
	}

	snap := env.State.Snapshot(tx.GetTxCurrency())
	var snap1 []int
	if tx.GetTxCurrency() != params.MAN_COIN {
		snap1 = env.State.Snapshot(params.MAN_COIN)
	}
	receipt, _, _, err := core.ApplyTransaction(env.config, bc, &coinbase, gp, env.State, env.header, tx, &env.header.GasUsed, vm.Config{})
	if err != nil {
		log.Error(packagename, "ApplyTransaction,err", err)
		env.State.RevertToSnapshot(tx.GetTxCurrency(), snap)
		if tx.GetTxCurrency() != params.MAN_COIN {
			env.State.RevertToSnapshot(params.MAN_COIN, snap1)
		}
		return err, nil
	}
	env.transer = append(env.transer, tx)
	env.recpts = append(env.recpts, receipt)
	env.mapcoingasUse.setCoinGasUse(tx, receipt.GasUsed)
	return nil, receipt.Logs
}
func (env *Work) s_commitTransaction(tx types.SelfTransaction, coinbase common.Address, gp *core.GasPool) (error, []*types.Log, *types.Receipt) {
	env.State.Prepare(tx.Hash(), common.Hash{}, env.tcount)
	snap := env.State.Snapshot(tx.GetTxCurrency())
	receipt, _, _, err := core.ApplyTransaction(env.config, env.bc, &coinbase, gp, env.State, env.header, tx, &env.header.GasUsed, vm.Config{})
	if err != nil {
		log.Error("s_commitTransaction commit err. ", "err", err)
		env.State.RevertToSnapshot(tx.GetTxCurrency(), snap)
		return err, nil, nil
	}
	env.tcount++
	return nil, receipt.Logs, receipt
}

//Leader
var lostCnt int = 0

type retStruct struct {
	no  []uint32
	txs []*types.Transaction
}

func myCoinsort(coins []string) []string {
	coinsnoman := make([]string, 0, len(coins))
	retCoins := make([]string, 0, len(coins))
	for _, coinname := range coins {
		if coinname == params.MAN_COIN {
			continue
		}
		coinsnoman = append(coinsnoman, coinname)
	}
	retCoins = append(retCoins, params.MAN_COIN)
	retCoins = append(retCoins, coinsnoman...)
	return retCoins
}

func (env *Work) ProcessTransactions(mux *event.TypeMux, tp txPoolReader, upTime map[common.Address]uint64) (listret []*common.RetCallTxN, originalTxs []types.SelfTransaction) {
	pending, err := tp.Pending()
	if err != nil {
		log.Error("Failed to fetch pending transactions", "err", err)
		return nil, nil
	}
	env.mapcoingasUse.clearmap()
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	log.Info("work", "关键时间点", "开始执行交易", "time", time.Now(), "块高", env.header.Number, "MAN交易数量", len(pending[params.MAN_COIN]))

	coins := make([]string, 0)
	coinsnoman := make([]string, 0)
	for coinname, _ := range pending {
		if coinname == params.MAN_COIN {
			continue
		}
		coinsnoman = append(coinsnoman, coinname) //leader
	}
	sort.Strings(coinsnoman)
	coins = append(coins, params.MAN_COIN)
	coins = append(coins, coinsnoman...)
	//先跑MAN交易，后跑其他币种交易
	for _, coinname := range coins {
		env.packNum = 0
		env.coinType = coinname
		tmplistret, tmporiginalTxs := env.commitTransactions(mux, pending[coinname], common.Address{})
		originalTxs = append(originalTxs, tmporiginalTxs...)
		listret = append(listret, tmplistret...)
	}

	//finalCoinTxs -按币种存放的所有交易;finalCoinRecpets -按币种存放的所有收据
	tCoinTxs, tCoinRecpets := types.GetCoinTXRS(env.transer, env.recpts) // env.transer就是originalTxs
	finalCoinTxs := make([]types.CoinSelfTransaction, 0, len(coins))
	finalCoinRecpets := make([]types.CoinReceipts, 0, len(coins))
	//查看是否有MAN分区（MAN币）,如果有直接append到finalCoinTxs，没有就创建MAN分区用于后面存奖励交易
	isHaveManCoin := false
	for _, tcoin := range tCoinTxs {
		if tcoin.CoinType == params.MAN_COIN {
			isHaveManCoin = true
			break
		}
	}
	if !isHaveManCoin {
		var MANCoinReceipt types.CoinReceipts
		var MANtmCointxs types.CoinSelfTransaction
		MANtmCointxs.CoinType = params.MAN_COIN //MAN分区必须有，用于存放矿工和验证者奖励费
		MANCoinReceipt.CoinType = params.MAN_COIN
		finalCoinTxs = append(finalCoinTxs, MANtmCointxs)
		finalCoinRecpets = append(finalCoinRecpets, MANCoinReceipt)
	}
	finalCoinTxs = append(finalCoinTxs, tCoinTxs...)
	finalCoinRecpets = append(finalCoinRecpets, tCoinRecpets...)

	CoinsMap := make(map[string]bool) //存放所有币种
	for _, cointxs := range finalCoinTxs {
		CoinsMap[cointxs.CoinType] = true
	}

	from := make(map[string][]common.Address)
	for _, tx := range originalTxs {
		from[tx.GetTxCurrency()] = append(from[tx.GetTxCurrency()], tx.From())
		log.Trace("关键时间点", "txNonce", tx.Nonce(), "txfrom", tx.From(), "txhash", tx.Hash().Hex(), "N", tx.GetTxN(0))
	}

	log.Info("work", "关键时间点", "执行交易完成，开始执行奖励", "time", time.Now(), "块高", env.header.Number, "tx num ", len(originalTxs))
	rewart := env.bc.Processor(env.header.Version).ProcessReward(env.State, env.header, upTime, from, env.mapcoingasUse.mapcoin)
	rewardTxmap := env.makeTransaction(rewart)
	allfinalTxs := make([]types.CoinSelfTransaction, 0, len(coins)) //按币种存放的所有交易切片(先放分区币种的奖励交易，然后存该币种的普通交易)
	allfinalRecpets := make([]types.CoinReceipts, 0, len(coins))    //按币种存放的所有收据切片(先放分区币种的奖励收据，然后存该币种的普通收据)

	//防止多币种交易下一个区块的没有该币种的交易，但有该币种奖励
	tmpcoins := make([]string, 0)
	for rewardCoinname, _ := range rewardTxmap {
		//tmpcoins = append(tmpcoins,rewardCoinname)
		CoinsMap[rewardCoinname] = true
	}
	for coinname, _ := range CoinsMap {
		tmpcoins = append(tmpcoins, coinname)
	}
	coins = myCoinsort(tmpcoins)

	//先跑MAN奖励交易，后跑其他币种奖励交易
	for _, coinname := range coins {
		var tCoinReceipt types.CoinReceipts
		var tCointxs types.CoinSelfTransaction
		tmpTxs := make([]types.SelfTransaction, 0)
		tmpRecepts := make(types.Receipts, 0)
		//tmpcoinRecpets := make([]types.CoinReceipts,0,len(coins))
		for _, tx := range rewardTxmap[coinname] {
			err, _, recpts := env.s_commitTransaction(tx, common.Address{}, new(core.GasPool).AddGas(0))
			if err != nil {
				log.Error("work.go", "ProcessTransactions:::reward Tx call Error", err)
				continue
			}
			tmpTxs = append(tmpTxs, tx)             //奖励放对应分区币种的前面
			tmpRecepts = append(tmpRecepts, recpts) //奖励放对应分区币种的前面
		}

		//tmpTxs = append(tmpTxs,tRewartTx...) //奖励交易放对应分区币种的前面
		for _, cointxs := range finalCoinTxs {
			if coinname == cointxs.CoinType {
				tmpTxs = append(tmpTxs, cointxs.Txser...) //普通交易放后面
			}
		}
		tCointxs.CoinType = coinname
		tCointxs.Txser = tmpTxs
		allfinalTxs = append(allfinalTxs, tCointxs)

		for _, coinrecpts := range finalCoinRecpets {
			if coinname == coinrecpts.CoinType {
				tmpRecepts = append(tmpRecepts, coinrecpts.Receiptlist...)
			}
		}
		tCoinReceipt.CoinType = coinname
		tCoinReceipt.Receiptlist = tmpRecepts
		allfinalRecpets = append(allfinalRecpets, tCoinReceipt)
	}
	env.State.Finalise("", true)
	env.txs = allfinalTxs
	env.Receipts = allfinalRecpets
	log.Info("work", "关键时间点", "奖励执行完成", "time", time.Now(), "块高", env.header.Number)
	return
}

func (env *Work) makeTransaction(rewarts []common.RewarTx) (coinTxs map[string]types.SelfTransactions) {
	nonceMap := make(map[common.Address]uint64)
	coinTxs = make(map[string]types.SelfTransactions)
	for _, rewart := range rewarts {
		sorted_keys := make([]string, 0)
		for k := range rewart.To_Amont {
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
				if rewart.RewardTyp == common.RewardInterestType {
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
			if rewart.RewardTyp == common.RewardInterestType {
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
		if _, ok := nonceMap[rewart.Fromaddr]; ok {
			nonceMap[rewart.Fromaddr] = nonceMap[rewart.Fromaddr] + 1
		} else {
			nonceMap[rewart.Fromaddr] = env.State.GetNonce(rewart.CoinType, rewart.Fromaddr)
		}
		tx := types.NewTransactions(nonceMap[rewart.Fromaddr], to, value, 0, new(big.Int), databytes, nil, nil, nil, extra, 0, env.rewardTypetransformation(rewart.RewardTyp), 0, rewart.CoinType, 0)
		tx.SetFromLoad(rewart.Fromaddr)
		//txers = append(txers, tx)
		rewartTxs := coinTxs[rewart.CoinRange]
		rewartTxs = append(rewartTxs, tx)
		coinTxs[rewart.CoinRange] = rewartTxs
	}
	return
}
func (env *Work) rewardTypetransformation(inputType byte) byte {
	switch inputType {
	case common.RewardMinerType:
		return common.ExtraUnGasMinerTxType
	case common.RewardValidatorType:
		return common.ExtraUnGasValidatorTxType
	case common.RewardInterestType:
		return common.ExtraUnGasInterestTxType
	case common.RewardTxsType:
		return common.ExtraUnGasTxsType
	case common.RewardLotteryType:
		return common.ExtraUnGasLotteryTxType
	default:
		log.Error("work.go", "rewardTypetransformation:Unknown reward type.", inputType)
		panic("rewardTypetransformation:Unknown reward type.")
		return common.ExtraUnGasMinerTxType
	}
}

//Broadcast
func (env *Work) ProcessBroadcastTransactions(mux *event.TypeMux, txs []types.CoinSelfTransaction) {
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	coins := make([]string, 0, len(txs)+1)
	if len(txs) > 1 {
		txs = mysort(txs)
	}
	env.mapcoingasUse.clearmap()
	for _, tx := range txs {
		env.packNum = 0
		env.coinType = tx.CoinType
		coins = append(coins, tx.CoinType)
		for _, t := range tx.Txser {
			if uint64(env.packNum) >= params.OtherCoinPackNum && env.coinType != params.MAN_COIN {
				break
			}
			env.commitTransaction(t, env.bc, common.Address{}, nil)
			env.packNum++
		}
	}
	coinsnoman := make([]string, 0, len(txs)+1)
	for _, coinname := range coins {
		if coinname == params.MAN_COIN {
			continue
		}
		coinsnoman = append(coinsnoman, coinname)
	}
	sort.Strings(coinsnoman)
	tcoins := make([]string, 0, len(txs)+1)
	tcoins = append(tcoins, params.MAN_COIN)
	tcoins = append(tcoins, coinsnoman...)
	coins = tcoins //前面是man币，后面是排序过的币种

	//finalCoinTxs -按币种存放的所有交易;finalCoinRecpets -按币种存放的所有收据
	tCoinTxs, tCoinRecpets := types.GetCoinTXRS(env.transer, env.recpts) // env.transer就是originalTxs
	finalCoinTxs := make([]types.CoinSelfTransaction, 0)
	finalCoinRecpets := make([]types.CoinReceipts, 0)
	//env.State.Finalise("MAN",true)
	//查看是否有MAN分区（MAN币）,如果有直接append到finalCoinTxs，没有就创建MAN分区用于后面存奖励交易
	isHaveManCoin := false
	for _, tcoin := range tCoinTxs {
		if tcoin.CoinType == params.MAN_COIN {
			isHaveManCoin = true
			break
		}
	}
	if !isHaveManCoin {
		var MANCoinReceipt types.CoinReceipts
		var MANtmCointxs types.CoinSelfTransaction
		MANtmCointxs.CoinType = params.MAN_COIN //MAN分区必须有，用于存放矿工和验证者奖励费
		MANCoinReceipt.CoinType = params.MAN_COIN
		finalCoinTxs = append(finalCoinTxs, MANtmCointxs)
		finalCoinRecpets = append(finalCoinRecpets, MANCoinReceipt)
	}
	finalCoinTxs = append(finalCoinTxs, tCoinTxs...)
	finalCoinRecpets = append(finalCoinRecpets, tCoinRecpets...)

	CoinsMap := make(map[string]bool) //存放所有币种
	for _, cointxs := range finalCoinTxs {
		CoinsMap[cointxs.CoinType] = true
	}

	rewart := env.bc.Processor(env.header.Version).ProcessReward(env.State, env.header, nil, nil, nil)
	rewardTxmap := env.makeTransaction(rewart)

	allfinalTxs := make([]types.CoinSelfTransaction, 0, len(coins)) //按币种存放的所有交易切片(先放分区币种的奖励交易，然后存该币种的普通交易)
	allfinalRecpets := make([]types.CoinReceipts, 0, len(coins))    //按币种存放的所有收据切片(先放分区币种的奖励收据，然后存该币种的普通收据)
	//防止多币种交易下一个区块的没有该币种的交易，但有该币种奖励
	tmpcoins := make([]string, 0)
	for rewardCoinname, _ := range rewardTxmap {
		CoinsMap[rewardCoinname] = true
	}
	for coinname, _ := range CoinsMap {
		tmpcoins = append(tmpcoins, coinname)
	}
	coins = myCoinsort(tmpcoins)

	//先跑MAN奖励交易，后跑其他币种奖励交易
	for _, coinname := range coins {
		var tCoinReceipt types.CoinReceipts
		var tCointxs types.CoinSelfTransaction
		tmpTxs := make([]types.SelfTransaction, 0)
		tmpRecepts := make(types.Receipts, 0)
		//tmpcoinRecpets := make([]types.CoinReceipts,0,len(coins))
		for _, tx := range rewardTxmap[coinname] {
			err, _, recpts := env.s_commitTransaction(tx, common.Address{}, new(core.GasPool).AddGas(0))
			if err != nil {
				log.Error("work.go", "ProcessTransactions:::reward Tx call Error", err)
				continue
			}
			tmpTxs = append(tmpTxs, tx)             //奖励放对应分区币种的前面
			tmpRecepts = append(tmpRecepts, recpts) //奖励放对应分区币种的前面
		}

		//tmpTxs = append(tmpTxs,tRewartTx...) //奖励交易放对应分区币种的前面
		for _, cointxs := range finalCoinTxs {
			if coinname == cointxs.CoinType {
				tmpTxs = append(tmpTxs, cointxs.Txser...)
			}
		}
		tCointxs.CoinType = coinname
		tCointxs.Txser = tmpTxs
		allfinalTxs = append(allfinalTxs, tCointxs)

		for _, coinrecpts := range finalCoinRecpets {
			if coinname == coinrecpts.CoinType {
				tmpRecepts = append(tmpRecepts, coinrecpts.Receiptlist...)
			}
		}
		tCoinReceipt.CoinType = coinname
		tCoinReceipt.Receiptlist = tmpRecepts
		allfinalRecpets = append(allfinalRecpets, tCoinReceipt)
	}
	env.State.Finalise("", true)
	env.txs = allfinalTxs
	env.Receipts = allfinalRecpets
	return
}

func mysort(s []types.CoinSelfTransaction) (sortCointxs []types.CoinSelfTransaction) {
	tmap := make(map[string][]types.SelfTransaction)
	keys := make([]string, 0)
	coinsnoman := make([]string, 0)
	coins := make([]string, 0)
	for _, k := range s {
		txs := tmap[k.CoinType]
		txs = append(txs, k.Txser...)
		tmap[k.CoinType] = txs
		keys = append(keys, k.CoinType)
	}
	sort.Strings(keys)
	isHaveMan := false
	for _, key := range keys {
		if key == params.MAN_COIN {
			isHaveMan = true
			continue
		}
		coinsnoman = append(coinsnoman, key)
	}
	if isHaveMan {
		coins = append(coins, params.MAN_COIN)
	}
	coins = append(coins, coinsnoman...)
	for _, coinname := range coins {
		var cointx types.CoinSelfTransaction
		cointx.CoinType = coinname
		cointx.Txser = append(cointx.Txser, tmap[coinname]...)
		sortCointxs = append(sortCointxs, cointx)
	}
	return sortCointxs
}

func (env *Work) ConsensusTransactions(mux *event.TypeMux, txs []types.CoinSelfTransaction, upTime map[common.Address]uint64) error {
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(env.header.GasLimit)
	}
	env.mapcoingasUse.clearmap()
	var coalescedLogs []types.CoinLogs
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	from := make(map[string][]common.Address)
	coins := make([]string, 0, len(txs)+1)
	log.Info("work", "关键时间点", "开始执行交易", "time", time.Now(), "块高", env.header.Number)
	if len(txs) > 1 {
		txs = mysort(txs)
	}

	for _, tx := range txs {
		env.packNum = 0
		env.coinType = tx.CoinType
		coins = append(coins, tx.CoinType)
		// If we don't have enough gas for any further transactions then we're done
		if env.gasPool.Gas() < params.TxGas {
			log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
			return errors.New("Not enough gas for further transactions")
		}
		// Start executing the transaction
		for _, t := range tx.Txser {
			if uint64(env.packNum) >= params.OtherCoinPackNum && env.coinType != params.MAN_COIN {
				break
			}
			env.State.Prepare(t.Hash(), common.Hash{}, env.tcount)
			err, logs := env.commitTransaction(t, env.bc, common.Address{}, env.gasPool)
			if err == nil {
				env.tcount++
				coalescedLogs = append(coalescedLogs, types.CoinLogs{t.GetTxCurrency(), logs})
				env.packNum++
			} else {
				return err
			}
			from[t.GetTxCurrency()] = append(from[t.GetTxCurrency()], t.From())
		}
	}

	coinsnoman := make([]string, 0, len(txs)+1)
	for _, coinname := range coins {
		if coinname == params.MAN_COIN {
			continue
		}
		coinsnoman = append(coinsnoman, coinname) //fllow
	}
	sort.Strings(coinsnoman)
	tcoins := make([]string, 0, len(txs)+1)
	tcoins = append(tcoins, params.MAN_COIN)
	tcoins = append(tcoins, coinsnoman...)
	coins = tcoins //前面是man币，后面是排序过的币种

	//finalCoinTxs -按币种存放的所有交易;finalCoinRecpets -按币种存放的所有收据
	tCoinTxs, tCoinRecpets := types.GetCoinTXRS(env.transer, env.recpts) // env.transer就是originalTxs
	finalCoinTxs := make([]types.CoinSelfTransaction, 0, len(coins))
	finalCoinRecpets := make([]types.CoinReceipts, 0, len(coins))
	//查看是否有MAN分区（MAN币）,如果有直接append到finalCoinTxs，没有就创建MAN分区用于后面存奖励交易
	isHaveManCoin := false
	for _, tcoin := range tCoinTxs {
		if tcoin.CoinType == params.MAN_COIN {
			isHaveManCoin = true
			break
		}
	}
	if !isHaveManCoin {
		var MANCoinReceipt types.CoinReceipts
		var MANtmCointxs types.CoinSelfTransaction
		MANtmCointxs.CoinType = params.MAN_COIN //MAN分区必须有，用于存放矿工和验证者奖励费
		MANCoinReceipt.CoinType = params.MAN_COIN
		finalCoinTxs = append(finalCoinTxs, MANtmCointxs)
		finalCoinRecpets = append(finalCoinRecpets, MANCoinReceipt)
	}
	finalCoinTxs = append(finalCoinTxs, tCoinTxs...)
	finalCoinRecpets = append(finalCoinRecpets, tCoinRecpets...)
	//env.State.Finalise("MAN",true)
	CoinsMap := make(map[string]bool) //存放所有币种
	for _, cointxs := range finalCoinTxs {
		CoinsMap[cointxs.CoinType] = true
	}

	log.Info("work", "关键时间点", "执行交易完成，开始执行奖励", "time", time.Now(), "块高", env.header.Number)
	rewart := env.bc.Processor(env.header.Version).ProcessReward(env.State, env.header, upTime, from, env.mapcoingasUse.mapcoin)
	rewardTxmap := env.makeTransaction(rewart)

	allfinalTxs := make([]types.CoinSelfTransaction, 0, len(coins)) //按币种存放的所有交易切片(先放分区币种的奖励交易，然后存该币种的普通交易)
	allfinalRecpets := make([]types.CoinReceipts, 0, len(coins))    //按币种存放的所有收据切片(先放分区币种的奖励收据，然后存该币种的普通收据)
	//防止多币种交易下一个区块的没有该币种的交易，但有该币种奖励
	tmpcoins := make([]string, 0)
	for rewardCoinname, _ := range rewardTxmap {
		CoinsMap[rewardCoinname] = true
	}
	for coinname, _ := range CoinsMap {
		tmpcoins = append(tmpcoins, coinname)
	}
	coins = myCoinsort(tmpcoins)

	//先跑MAN奖励交易，后跑其他币种奖励交易
	for _, coinname := range coins {
		var tCoinReceipt types.CoinReceipts
		var tCointxs types.CoinSelfTransaction
		tmpTxs := make([]types.SelfTransaction, 0)
		tmpRecepts := make(types.Receipts, 0)
		//tmpcoinRecpets := make([]types.CoinReceipts,0,len(coins))
		for _, tx := range rewardTxmap[coinname] {
			err, _, recpts := env.s_commitTransaction(tx, common.Address{}, new(core.GasPool).AddGas(0))
			if err != nil {
				log.Error("work.go", "ProcessTransactions:::reward Tx call Error", err)
				continue
			}
			tmpTxs = append(tmpTxs, tx)             //奖励放对应分区币种的前面
			tmpRecepts = append(tmpRecepts, recpts) //奖励放对应分区币种的前面
		}

		//tmpTxs = append(tmpTxs,tRewartTx...) //奖励交易放对应分区币种的前面
		for _, cointxs := range finalCoinTxs {
			if coinname == cointxs.CoinType {
				tmpTxs = append(tmpTxs, cointxs.Txser...)
			}
		}
		tCointxs.CoinType = coinname
		tCointxs.Txser = tmpTxs
		allfinalTxs = append(allfinalTxs, tCointxs)
		for _, coinrecpts := range finalCoinRecpets {
			if coinname == coinrecpts.CoinType {
				tmpRecepts = append(tmpRecepts, coinrecpts.Receiptlist...)
			}
		}
		tCoinReceipt.CoinType = coinname
		tCoinReceipt.Receiptlist = tmpRecepts
		allfinalRecpets = append(allfinalRecpets, tCoinReceipt)
	}
	env.txs = allfinalTxs
	env.Receipts = allfinalRecpets
	env.State.Finalise("", true)

	if len(coalescedLogs) > 0 || env.tcount > 0 {
		go func(logs []types.CoinLogs, tcount int) {
			if len(logs) > 0 {
				mux.Post(core.PendingLogsEvent{Logs: logs})
			}
			if tcount > 0 {
				mux.Post(core.PendingStateEvent{})
			}
		}(coalescedLogs, env.tcount)
	}
	log.Info("work", "关键时间点", "奖励执行完成", "time", time.Now(), "块高", env.header.Number)
	return nil
}
func (env *Work) GetTxs() []types.CoinSelfTransaction {
	return env.txs
}
