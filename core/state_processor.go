// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"math/big"
	"runtime"
	"sync"

	"github.com/MatrixAINetwork/go-matrix/reward/util"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"

	"bufio"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/reward/blkreward"
	"github.com/MatrixAINetwork/go-matrix/reward/interest"
	"github.com/MatrixAINetwork/go-matrix/reward/lottery"
	"github.com/MatrixAINetwork/go-matrix/reward/slash"
	"github.com/MatrixAINetwork/go-matrix/reward/txsreward"
	"github.com/pkg/errors"
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
	random *baseinterface.Random
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}
func (p *StateProcessor) SetRandom(random *baseinterface.Random) {
	p.random = random
}
func (p *StateProcessor) getCoinAddress(cointypelist map[string]types.SelfTransactions) map[string]common.Address{
	statedbM,_ := p.bc.State()
	coinconfig := statedbM.GetMatrixData(types.RlpHash(common.COINPREFIX+mc.MSCurrencyConfig))
	ret:=make(map[string]common.Address)
	var coincfglist []common.CoinConfig
	if len(coinconfig) > 0{
		err := json.Unmarshal(coinconfig,&coincfglist)
		if err != nil{
			log.Trace("get coin config list","unmarshal err",err)
			return nil
		}
	}
	for _,cc:=range coincfglist{
		if cc.PackNum <=0{
			continue
		}
		if _,ok:=cointypelist[cc.CoinType];ok{
			ret[cc.CoinType] = cc.CoinAddress
		}
	}
	if _,ok:=cointypelist[params.MAN_COIN];ok{
		ret[params.MAN_COIN] = common.TxGasRewardAddress
	}
	return ret
}
func (p *StateProcessor) getGas(state *state.StateDBManage, gas *big.Int) *big.Int {
	gasprice, err := matrixstate.GetTxpoolGasLimit(state)
	if err != nil {
		return big.NewInt(0)
	}
	allGas := new(big.Int).Mul(gas, new(big.Int).SetUint64(gasprice.Uint64()))
	log.INFO("奖励", "交易费奖励总额", allGas.String())
	balance := state.GetBalance(params.MAN_COIN, common.TxGasRewardAddress)

	if len(balance) == 0 {
		log.WARN("奖励", "交易费奖励账户余额不合法", "")
		return big.NewInt(0)
	}

	if balance[common.MainAccount].Balance.Cmp(big.NewInt(0)) <= 0 || balance[common.MainAccount].Balance.Cmp(allGas) < 0 {
		log.WARN("奖励", "交易费奖励账户余额不合法，余额", balance)
		return big.NewInt(0)
	}
	return allGas
}
func (env *StateProcessor) reverse(s []common.RewarTx) []common.RewarTx {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func (p *StateProcessor) ProcessReward(st *state.StateDBManage, header *types.Header, upTime map[common.Address]uint64, account []common.Address, usedGas uint64) []common.RewarTx {
	bcInterval, err := matrixstate.GetBroadcastInterval(st)
	if err != nil {
		log.Error("奖励", "获取广播周期失败", err)
		return nil
	}
	if bcInterval.IsBroadcastNumber(header.Number.Uint64()) {
		return nil
	}
	preState, err := p.bc.StateAtBlockHash(header.ParentHash)
	if err != nil {
		log.Error("奖励", "获取前一个状态错误", err)
		return nil
	}

	var ppreState *state.StateDBManage
	if header.Number.Uint64() == 1 {
		ppreState = preState
	} else {
		block := p.bc.GetBlockByHash(header.ParentHash)
		ppreState, err = p.bc.StateAtBlockHash(block.ParentHash())
		if err != nil {
			log.Error("奖励", "获取前一个状态错误", err)
			return nil
		}
	}

	blkReward := blkreward.New(p.bc, st, preState, ppreState)
	rewardList := make([]common.RewarTx, 0)
	if nil != blkReward {
		//todo: read half number from state
		minersRewardMap := blkReward.CalcMinerRewards(header.Number.Uint64(), header.ParentHash)
		if 0 != len(minersRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: params.MAN_COIN, Fromaddr: common.BlkMinerRewardAddress, To_Amont: minersRewardMap, RewardTyp: common.RewardMinerType})
		}

		validatorsRewardMap := blkReward.CalcValidatorRewards(header.Leader, header.Number.Uint64())
		if 0 != len(validatorsRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: params.MAN_COIN, Fromaddr: common.BlkValidatorRewardAddress, To_Amont: validatorsRewardMap, RewardTyp: common.RewardValidatorType})
		}
	}

	allGas := p.getGas(st, new(big.Int).SetUint64(usedGas))
	txsReward := txsreward.New(p.bc, st, preState)
	if nil != txsReward {
		txsRewardMap := txsReward.CalcNodesRewards(allGas, header.Leader, header.Number.Uint64(), header.ParentHash)
		if 0 != len(txsRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: params.MAN_COIN, Fromaddr: common.TxGasRewardAddress, To_Amont: txsRewardMap, RewardTyp: common.RewardTxsType})
		}
	}

	lottery := lottery.New(p.bc, st, p.random, preState)
	if nil != lottery {
		lotteryRewardMap := lottery.LotteryCalc(header.ParentHash, header.Number.Uint64())
		if 0 != len(lotteryRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: params.MAN_COIN, Fromaddr: common.LotteryRewardAddress, To_Amont: lotteryRewardMap, RewardTyp: common.RewardLotteryType})
		}
		lottery.LotterySaveAccount(account, header.VrfValue)
	}

	////todo 利息
	interestReward := interest.New(st, preState)

	if nil == interestReward {
		return p.reverse(util.Accumulator(st, rewardList))
	}
	interestReward.CalcReward(st, header.Number.Uint64(), header.ParentHash)

	slash := slash.New(p.bc, st, preState)
	if nil != slash {
		slash.CalcSlash(st, header.Number.Uint64(), upTime, header.ParentHash)
	}
	interestPayMap := interestReward.PayInterest(st, header.Number.Uint64())
	if 0 != len(interestPayMap) {
		rewardList = append(rewardList, common.RewarTx{CoinType: params.MAN_COIN, Fromaddr: common.InterestRewardAddress, To_Amont: interestPayMap, RewardTyp: common.RewardInterestType})
	}
	return p.reverse(util.Accumulator(st, rewardList))
}

func (p *StateProcessor) ProcessSuperBlk(block *types.Block, statedb *state.StateDBManage) error {
	log.Trace("BlockChain insertChain in3 IsSuperBlock")
	sbs, err := p.bc.GetSuperBlockSeq()
	if nil != err {
		return errors.New("get super seq error")
	}
	if block.Header().SuperBlockSeq() <= sbs {
		return errors.Errorf("invalid super block seq (remote: %x local: %x)", block.Header().SuperBlockSeq(), sbs)
	}
	log.Trace("BlockChain insertChain in3 IsSuperBlock processSuperBlockState")
	err = p.bc.processSuperBlockState(block, statedb)
	if err != nil {
		p.bc.reportBlock(block, nil, err)
		return err
	}
	var root []common.CoinRoot
	root, _ = statedb.IntermediateRoot(p.bc.chainConfig.IsEIP158(block.Number()))
	intermediateroothash := types.RlpHash(root)
	blockroothash := types.RlpHash(block.Root())
	isok := false
	for _, cr := range root {
		for _, br := range block.Root() {
			if cr.Cointyp == br.Cointyp {
				if cr.Root != br.Root {
					isok = true
				}
			}
		}
	}
	if isok {
		return errors.Errorf("invalid super block root (remote: %x local: %x)", blockroothash, intermediateroothash)
	}
	return nil
}

func (p *StateProcessor) readShardConfig(filePth string) ([]common.CoinSharding, error) {
	f, err := os.Open(filePth)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bfRd := bufio.NewReader(f)
	coinshard := make([]common.CoinSharding, 0)
	for {
		ushards := make([]uint, 0)
		buf, _, err := bfRd.ReadLine()
		if len(buf) <= 0 {
			if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
				if err == io.EOF {
					break
				}
				return nil, err
			}
		}
		str := string(buf)
		idx := strings.Index(str, "=")
		coin := str[0:idx]
		shards := str[idx+1:]
		strs := strings.Split(shards, ",")
		for _, str := range strs {
			i, _ := strconv.Atoi(str)
			ushards = append(ushards, uint(i))
		}
		coinshard = append(coinshard, common.CoinSharding{CoinType: coin, Shardings: ushards})
	}
	return coinshard, nil
}
func (p *StateProcessor) checkCoinShard(coinShard []common.CoinSharding) []common.CoinSharding {
	isexistCoin := false
	isexistShard := false
	for _, cs := range coinShard {
		if cs.CoinType == params.MAN_COIN {
			isexistCoin = true
			for _, s := range cs.Shardings {
				if s == 0 {
					isexistShard = true
				}
			}
			break
		}
	}
	if !isexistCoin {
		coinShard = append(coinShard, common.CoinSharding{CoinType: params.MAN_COIN, Shardings: []uint{0}})
	} else if isexistShard {
		for i, cs := range coinShard {
			ui := make([]uint, 0)
			if cs.CoinType == params.MAN_COIN {
				ui = append(ui, uint(0))
				ui = append(ui, coinShard[i].Shardings...)
				coinShard[i].Shardings = ui
			}
		}
	}
	return coinShard
}

// Process processes the state changes according to the Matrix rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) ProcessTxs(block *types.Block, statedb *state.StateDBManage, cfg vm.Config, upTime map[common.Address]uint64) ([]types.CoinLogs, uint64, error) {
	var (
		receipts    types.Receipts
		allreceipts = make(map[string]types.Receipts)
		usedGas     = new(uint64)
		header      = block.Header()
		allLogs     []types.CoinLogs
		gp                 = new(GasPool).AddGas(block.GasLimit())
		retAllGas   uint64 = 0
	)
	cs, cserr := p.readShardConfig("")
	var coinShard []common.CoinSharding
	if cserr == nil {
		coinShard = make([]common.CoinSharding, len(cs))
		copy(coinShard, cs)
		coinShard = p.checkCoinShard(coinShard)
	}
	// Iterate over and process the individual transactions
	statedb.UpdateTxForBtree(uint32(block.Time().Uint64()))
	statedb.UpdateTxForBtreeBytime(uint32(block.Time().Uint64()))
	stxs := make([]types.SelfTransaction, 0)
	ftxs := make([]types.SelfTransaction, 0)
	txs := make([]types.SelfTransaction, 0)
	var txcount int
	tmpMaptx := make(map[string]types.SelfTransactions)
	tmpMapre := make(map[string]types.Receipts)
	for _, cb := range block.Currencies() {
		txs = append(txs, cb.Transactions.GetTransactions()...)
	}
	var waitG = &sync.WaitGroup{}
	maxProcs := runtime.NumCPU() //获取cpu个数
	if maxProcs >= 2 {
		runtime.GOMAXPROCS(maxProcs - 1) //限制同时运行的goroutines数量
	}
	normalTxindex := 0
	for _, tx := range txs {
		if tx.GetMatrixType() == common.ExtraUnGasMinerTxType || tx.GetMatrixType() == common.ExtraUnGasValidatorTxType ||
			tx.GetMatrixType() == common.ExtraUnGasInterestTxType || tx.GetMatrixType() == common.ExtraUnGasTxsType || tx.GetMatrixType() == common.ExtraUnGasLotteryTxType {
			tmpstxs := make([]types.SelfTransaction, 0)
			tmpstxs = append(tmpstxs, tx)
			tmpstxs = append(tmpstxs, stxs...)
			stxs = tmpstxs
			normalTxindex++
			continue
		}
		sig := types.NewEIP155Signer(tx.ChainId())
		waitG.Add(1)
		ttx := tx
		go types.Sender_self(sig, ttx, waitG)
	}
	waitG.Wait()
	from := make([]common.Address, 0)
	isvadter := p.isValidater(header.ParentHash)
	for i, tx := range txs[normalTxindex:] {
		if tx.GetMatrixType() == common.ExtraUnGasMinerTxType || tx.GetMatrixType() == common.ExtraUnGasValidatorTxType ||
			tx.GetMatrixType() == common.ExtraUnGasInterestTxType || tx.GetMatrixType() == common.ExtraUnGasTxsType || tx.GetMatrixType() == common.ExtraUnGasLotteryTxType {
			tmpstxs := make([]types.SelfTransaction, 0)
			tmpstxs = append(tmpstxs, tx)
			tmpstxs = append(tmpstxs, stxs...)
			stxs = tmpstxs
			continue
		}
		if tx.IsEntrustTx() {
			from := tx.From()
			entrustFrom := statedb.GetGasAuthFrom(tx.GetTxCurrency(), from, p.bc.CurrentBlock().NumberU64()) //
			if !entrustFrom.Equal(common.Address{}) {
				tx.Setentrustfrom(entrustFrom)
				tx.SetIsEntrustGas(true)
			} else {
				entrustFrom := statedb.GetGasAuthFromByTime(tx.GetTxCurrency(), from, uint64(block.Time().Uint64()))
				if !entrustFrom.Equal(common.Address{}) {
					tx.Setentrustfrom(entrustFrom)
					tx.SetIsEntrustGas(true)
					tx.SetIsEntrustByTime(true)
				} else {
					log.Error("下载过程:该用户没有被授权过委托Gas")
					return nil, 0, ErrWithoutAuth
				}
			}
		}
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		receipt, gas, shard, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil {
			return nil, 0, err
		}
		allreceipts[tx.GetTxCurrency()] = append(allreceipts[tx.GetTxCurrency()], receipt)
		retAllGas += gas
		if isvadter {
			//receipts = append(receipts, receipt)
			allLogs = append(allLogs, types.CoinLogs{CoinType: tx.GetTxCurrency(), Logs: receipt.Logs})
			tmpMaptx[tx.GetTxCurrency()] = append(tmpMaptx[tx.GetTxCurrency()], tx)
			tmpMapre[tx.GetTxCurrency()] = append(tmpMapre[tx.GetTxCurrency()], receipt)
		} else {
			if p.isaddSharding(shard, coinShard, tx.GetTxCurrency()) {
				//receipts = append(receipts, receipt)
				allLogs = append(allLogs, types.CoinLogs{CoinType: tx.GetTxCurrency(), Logs: receipt.Logs})
				tmpMaptx[tx.GetTxCurrency()] = append(tmpMaptx[tx.GetTxCurrency()], tx)
				tmpMapre[tx.GetTxCurrency()] = append(tmpMapre[tx.GetTxCurrency()], receipt)
			} else {
				if _, ok := tmpMaptx[tx.GetTxCurrency()]; ok {
					//receipts = append(receipts, nil)
					tmpMaptx[tx.GetTxCurrency()] = append(tmpMaptx[tx.GetTxCurrency()], nil)
					tmpMapre[tx.GetTxCurrency()] = append(tmpMapre[tx.GetTxCurrency()], nil)
				}
			}
		}
		txcount = i
		from = append(from, tx.From())
	}
	p.ProcessReward(statedb, block.Header(), upTime, from, retAllGas)
	for _, tx := range stxs {
		statedb.Prepare(tx.Hash(), block.Hash(), txcount+1)
		receipt, _, shard, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil {
			return nil, 0, err
		}
		tmpr2 := make(types.Receipts, 1+len(allreceipts[tx.GetTxCurrency()]))
		tmpr2[0] = receipt
		copy(tmpr2[1:], allreceipts[tx.GetTxCurrency()])
		allreceipts[tx.GetTxCurrency()] = tmpr2

		var tmptx types.SelfTransaction
		if isvadter {
			tmptx = tx
		} else {
			if p.isaddSharding(shard, coinShard, tx.GetTxCurrency()) {
				tmptx = tx
			} else {
				tmptx = nil
				receipt = nil
			}
		}
		tmpr := make(types.Receipts, 1+len(receipts))
		tmpr[0] = receipt
		copy(tmpr[1:], receipts)
		receipts = tmpr
		if receipt != nil {
			tmpl := make([]types.CoinLogs, 0)
			tmpl = append(tmpl, types.CoinLogs{CoinType: params.MAN_COIN, Logs: receipt.Logs})
			tmpl = append(tmpl, allLogs...)
			allLogs = tmpl
		}
		ftxs = append(ftxs, tmptx)
	}
	receipts = append(receipts, tmpMapre[params.MAN_COIN]...)
	tmpMapre[params.MAN_COIN] = receipts
	ftxs = append(ftxs, tmpMaptx[params.MAN_COIN]...)
	tmpMaptx[params.MAN_COIN] = ftxs

	currblock := make([]types.CurrencyBlock, 0)
	for i, bc := range block.Currencies() {
		if !isvadter {
			if len(coinShard) > 0 {
				for _, cs := range coinShard {
					if bc.CurrencyName == cs.CoinType {
						block.Currencies()[i].Receipts = types.SetReceipts(tmpMapre[bc.CurrencyName], allreceipts[bc.CurrencyName].HashList(), cs.Shardings)
						block.Currencies()[i].Transactions = types.SetTransactions(tmpMaptx[bc.CurrencyName], types.TxHashList(txs), cs.Shardings)
						currblock = append(currblock, block.Currencies()[i])
						break
					}
				}
			} else {
				block.Currencies()[i].Receipts = types.SetReceipts(tmpMapre[bc.CurrencyName], allreceipts[bc.CurrencyName].HashList(), nil)
				currblock = append(currblock, block.Currencies()[i])
			}

		} else {
			block.Currencies()[i].Receipts = types.SetReceipts(tmpMapre[bc.CurrencyName], allreceipts[bc.CurrencyName].HashList(), nil)
			currblock = append(currblock, block.Currencies()[i])
		}
	}
	block.SetCurrencies(currblock)
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Uncles(), block.Currencies())

	return allLogs, *usedGas, nil
}
func (p *StateProcessor) isValidater(hash common.Hash) bool {
	roles, _ := ca.GetElectedByHeightAndRoleByHash(hash, common.RoleValidator)
	for _, role := range roles {
		if role.SignAddress == ca.GetSignAddress() {
			return true
		}
	}
	if ca.GetRole() == common.RoleBroadcast {
		return true
	}
	return false
}
func (p *StateProcessor) isaddSharding(shard []uint, shardings []common.CoinSharding, cointyp string) bool {
	if len(shardings) == 0 {
		return true
	}
	for _, s := range shard {
		if s == 0 && cointyp == params.MAN_COIN {
			return true
		}
		for _, ss := range shardings {
			if cointyp == ss.CoinType {
				for _, sd := range ss.Shardings {
					if sd == s {
						return true
					}
				}
				break
			}
		}
	}
	return false
}

func (p *StateProcessor) Process(block *types.Block, parent *types.Block, statedb *state.StateDBManage, cfg vm.Config) ([]types.CoinReceipts, []types.CoinLogs, uint64, error) {

	err := p.bc.ProcessStateVersion(block.Version(), statedb)
	if err != nil {
		log.Trace("BlockChain insertChain in3 Process Block err0")
		return nil, nil, 0, err
	}

	uptimeMap, err := p.bc.ProcessUpTime(statedb, block.Header())
	if err != nil {
		log.Trace("BlockChain insertChain in3 Process Block err1")
		p.bc.reportBlock(block, nil, err)
		return nil, nil, 0, err
	}

	err = p.bc.ProcessBlockGProduceSlash(statedb, block.Header())
	if err != nil {
		log.Trace("BlockChain insertChain in3 Process Block err2")
		p.bc.reportBlock(block, nil, err)
		return nil, nil, 0, err
	}

	// Process block using the parent state as reference point.
	logs, usedGas, err := p.ProcessTxs(block, statedb, cfg, uptimeMap)
	if err != nil {
		log.Trace("BlockChain insertChain in3 Process Block err3")
		p.bc.reportBlock(block, nil, err)
		return nil, logs, usedGas, err
	}

	// Process matrix state
	err = p.bc.matrixProcessor.ProcessMatrixState(block, string(parent.Version()), statedb)
	if err != nil {
		log.Trace("BlockChain insertChain in3 Process Block err4")
		return nil, logs, usedGas, err
	}

	return nil, logs, usedGas, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDBManage, header *types.Header, tx types.SelfTransaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, []uint, error) {
	if !BlackListFilter(tx, statedb, header.Number) {
		return nil, 0, nil, errors.New("blacklist account")
	}
	// Create a new context to be used in the EVM environment
	from, err := tx.GetTxFrom()
	if err != nil {
		from, err = types.Sender(types.NewEIP155Signer(config.ChainId), tx)
	}
	statedb.MakeStatedb(tx.GetTxCurrency(), true)
	context := NewEVMContext(from, tx.GasPrice(), header, bc, author)

	vmenv := vm.NewEVM(context, statedb, config, cfg, tx.GetTxCurrency())
	//如果是委托gas并且是按时间委托
	if tx.GetIsEntrustGas() && tx.GetIsEntrustByTime() {
		if !statedb.GetIsEntrustByTime(tx.GetTxCurrency(), from, header.Time.Uint64()) {
			log.Error("按时间委托gas的交易失效")
			return nil, 0, nil, errors.New("entrustTx is invalid")
		}
	}

	// Apply the transaction to the current state (included in the env)
	var gas uint64
	var failed bool
	shardings := make([]uint, 0)
	if tx.TxType() == types.BroadCastTxIndex {
		if extx := tx.GetMatrix_EX(); (extx != nil) && len(extx) > 0 && extx[0].TxType == 1 {
			gas = uint64(0)
			failed = true
		}
	} else {
		_, gas, failed, shardings, err = ApplyMessage(vmenv, tx, gp)
		if err != nil {
			return nil, 0, nil, err
		}
	}

	// Update the state with pending changes
	var root []byte
	if config.IsByzantium(header.Number) {
		statedb.Finalise(tx.GetTxCurrency(), true)
	} else {
		root = statedb.IntermediateRootByCointype(tx.GetTxCurrency(), config.IsEIP158(header.Number)).Bytes()
	}
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if tx.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.GetTxCurrency(), tx.From(), tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, gas, shardings, err
}
