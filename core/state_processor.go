// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"math/big"
	"runtime"
	"sync"
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/consensus/misc"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/reward/blkreward"
	"github.com/matrix/go-matrix/reward/interest"
	"github.com/matrix/go-matrix/reward/lottery"
	"github.com/matrix/go-matrix/reward/slash"
	"github.com/matrix/go-matrix/reward/txsreward"
	"github.com/pkg/errors"
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
func (p *StateProcessor) getGas(state *state.StateDB, gas *big.Int) *big.Int {
	gasprice, err := matrixstate.GetTxpoolGasLimit(state)
	if err != nil {
		return big.NewInt(0)
	}
	allGas := new(big.Int).Mul(gas, new(big.Int).SetUint64(gasprice.Uint64()))
	log.INFO("奖励", "交易费奖励总额", allGas.String())
	balance := state.GetBalance(common.TxGasRewardAddress)

	if len(balance) == 0 {
		log.WARN("奖励", "交易费奖励账户余额不合法", "")
		return big.NewInt(0)
	}

	if balance[common.MainAccount].Balance.Cmp(big.NewInt(0)) <= 0 || balance[common.MainAccount].Balance.Cmp(allGas) <= 0 {
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

func (p *StateProcessor) ProcessReward(st *state.StateDB, header *types.Header, upTime map[common.Address]uint64, account []common.Address, usedGas uint64) []common.RewarTx {
	bcInterval, err := matrixstate.GetBroadcastInterval(st)
	if err != nil {
		log.Error("work", "获取广播周期失败", err)
		return nil
	}
	if bcInterval.IsBroadcastNumber(header.Number.Uint64()) {
		return nil
	}
	blkReward := blkreward.New(p.bc, st)
	rewardList := make([]common.RewarTx, 0)
	if nil != blkReward {
		//todo: read half number from state
		minersRewardMap := blkReward.CalcMinerRewards(header.Number.Uint64(), header.ParentHash)
		if 0 != len(minersRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: "MAN", Fromaddr: common.BlkMinerRewardAddress, To_Amont: minersRewardMap})
		}

		validatorsRewardMap := blkReward.CalcValidatorRewards(header.Leader, header.Number.Uint64())
		if 0 != len(validatorsRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: "MAN", Fromaddr: common.BlkValidatorRewardAddress, To_Amont: validatorsRewardMap})
		}
	}

	allGas := p.getGas(st, new(big.Int).SetUint64(usedGas))
	txsReward := txsreward.New(p.bc, st)
	if nil != txsReward {
		txsRewardMap := txsReward.CalcNodesRewards(allGas, header.Leader, header.Number.Uint64(), header.ParentHash)
		if 0 != len(txsRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: "MAN", Fromaddr: common.TxGasRewardAddress, To_Amont: txsRewardMap})
		}
	}
	lottery := lottery.New(p.bc, st, p.random)
	if nil != lottery {
		lotteryRewardMap := lottery.LotteryCalc(header.ParentHash, header.Number.Uint64())
		if 0 != len(lotteryRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: "MAN", Fromaddr: common.LotteryRewardAddress, To_Amont: lotteryRewardMap})
		}
		lottery.LotterySaveAccount(account, header.VrfValue)
	}

	////todo 利息
	interestReward := interest.New(st)
	if nil == interestReward {
		return p.reverse(rewardList)
	}
	interestCalcMap := interestReward.CalcInterest(st, header.Number.Uint64())

	slash := slash.New(p.bc, st)
	if nil != slash {
		slash.CalcSlash(st, header.Number.Uint64(), upTime, interestCalcMap)
	}
	interestPayMap := interestReward.PayInterest(st, header.Number.Uint64())
	if 0 != len(interestPayMap) {
		rewardList = append(rewardList, common.RewarTx{CoinType: "MAN", Fromaddr: common.InterestRewardAddress, To_Amont: interestPayMap, RewardTyp: common.RewardInerestType})
	}
	return p.reverse(rewardList)
}

func (p *StateProcessor) ProcessSuperBlk(block *types.Block, statedb *state.StateDB) error {
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

	root := statedb.IntermediateRoot(p.bc.chainConfig.IsEIP158(block.Number()))
	if root != block.Root() {
		return errors.Errorf("invalid super block root (remote: %x local: %x)", block.Root, root)
	}
	return nil
}

// Process processes the state changes according to the Matrix rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) ProcessTxs(block *types.Block, statedb *state.StateDB, cfg vm.Config, upTime map[common.Address]uint64) (types.Receipts, []*types.Log, uint64, error) {
	var (
		receipts types.Receipts
		usedGas  = new(uint64)
		header   = block.Header()
		allLogs  []*types.Log
		gp       = new(GasPool).AddGas(block.GasLimit())
	)
	// Mutate the the block and state according to any hard-fork specs
	if p.config.DAOForkSupport && p.config.DAOForkBlock != nil && p.config.DAOForkBlock.Cmp(block.Number()) == 0 {
		misc.ApplyDAOHardFork(statedb)
	}
	// Iterate over and process the individual transactions
	statedb.UpdateTxForBtree(uint32(block.Time().Uint64()))
	statedb.UpdateTxForBtreeBytime(uint32(block.Time().Uint64()))
	stxs := make([]types.SelfTransaction, 0)
	var txcount int
	txs := block.Transactions()
	var waitG = &sync.WaitGroup{}
	maxProcs := runtime.NumCPU() //获取cpu个数
	if maxProcs >= 2 {
		runtime.GOMAXPROCS(maxProcs - 1) //限制同时运行的goroutines数量
	}
	normalTxindex := 0
	for _, tx := range txs {
		if tx.GetMatrixType() == common.ExtraUnGasTxType {
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
	for i, tx := range txs[normalTxindex:] {
		if tx.GetMatrixType() == common.ExtraUnGasTxType {
			tmpstxs := make([]types.SelfTransaction, 0)
			tmpstxs = append(tmpstxs, tx)
			tmpstxs = append(tmpstxs, stxs...)
			stxs = tmpstxs
			continue
		}
		if tx.IsEntrustTx() {
			from := tx.From()
			entrustFrom := statedb.GetGasAuthFrom(from, p.bc.CurrentBlock().NumberU64()) //
			if !entrustFrom.Equal(common.Address{}) {
				tx.Setentrustfrom(entrustFrom)
				//tx.IsEntrustGas = true
				tx.SetIsEntrustGas(true)
			} else {
				entrustFrom := statedb.GetGasAuthFromByTime(from, uint64(block.Time().Uint64()))
				if !entrustFrom.Equal(common.Address{}) {
					tx.Setentrustfrom(entrustFrom)
					//tx.IsEntrustGas = true
					//tx.IsEntrustByTime = true
					tx.SetIsEntrustGas(true)
					tx.SetIsEntrustByTime(true)
				} else {
					log.Error("下载过程:该用户没有被授权过委托Gas")
					return nil, nil, 0, ErrWithoutAuth
				}
			}
		}
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		receipt, _, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil && err != ErrSpecialTxFailed {
			return nil, nil, 0, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
		txcount = i
		from = append(from, tx.From())
	}
	p.ProcessReward(statedb, block.Header(), upTime, from, *usedGas)

	for _, tx := range stxs {
		statedb.Prepare(tx.Hash(), block.Hash(), txcount+1)
		receipt, _, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil && err != ErrSpecialTxFailed {
			return nil, nil, 0, err
		}
		tmpr := make(types.Receipts, 0)
		tmpr = append(tmpr, receipt)
		tmpr = append(tmpr, receipts...)
		receipts = tmpr
		tmpl := make([]*types.Log, 0)
		tmpl = append(tmpl, receipt.Logs...)
		tmpl = append(tmpl, allLogs...)
		allLogs = tmpl
	}

	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Transactions(), block.Uncles(), receipts)

	return receipts, allLogs, *usedGas, nil
}

func (p *StateProcessor) Process(block *types.Block, parent *types.Block, statedb *state.StateDB, cfg vm.Config) error {

	err := p.bc.ProcessStateVersion(block.Header().Version, statedb)
	if err != nil {
		log.Trace("BlockChain insertChain in3 Process Block err0")
		return err
	}

	uptimeMap, err := p.bc.ProcessUpTime(statedb, block.Header())
	if err != nil {
		log.Trace("BlockChain insertChain in3 Process Block err1")
		p.bc.reportBlock(block, nil, err)
		return err
	}
	err = p.bc.ProcessBlockGProduceSlash(statedb, block.Header())
	if err != nil {
		log.Trace("BlockChain insertChain in3 Process Block err6")
		p.bc.reportBlock(block, nil, err)
		return err
	}
	// Process block using the parent state as reference point.
	receipts, _, usedGas, err := p.ProcessTxs(block, statedb, cfg, uptimeMap)
	if err != nil {
		log.Trace("BlockChain insertChain in3 Process Block err2")
		p.bc.reportBlock(block, receipts, err)
		return err
	}

	// Process matrix state
	err = p.bc.matrixProcessor.ProcessMatrixState(block, statedb)
	if err != nil {
		log.Trace("BlockChain insertChain in3 Process Block err3")
		return err
	}

	// Validate the state using the default validator
	err = p.bc.Validator(block.Header().Version).ValidateState(block, parent, statedb, receipts, usedGas)
	if err != nil {
		log.Trace("BlockChain insertChain in3 Process Block err4")
		p.bc.reportBlock(block, receipts, err)
		return err
	}

	return nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx types.SelfTransaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	// Create a new context to be used in the EVM environment
	from, err := tx.GetTxFrom()
	if err != nil {
		from, err = types.Sender(types.NewEIP155Signer(config.ChainId), tx)
	}
	context := NewEVMContext(from, tx.GasPrice(), header, bc, author)

	vmenv := vm.NewEVM(context, statedb, config, cfg)

	//如果是委托gas并且是按时间委托
	if tx.GetIsEntrustGas() && tx.GetIsEntrustByTime() {
		if !statedb.GetIsEntrustByTime(from, header.Time.Uint64()) {
			log.Error("按时间委托gas的交易失效")
			return nil, 0, errors.New("entrustTx is invalid")
		}
	}

	// Apply the transaction to the current state (included in the env)
	var gas uint64
	var failed bool
	if tx.TxType() == types.BroadCastTxIndex {
		if extx := tx.GetMatrix_EX(); (extx != nil) && len(extx) > 0 && extx[0].TxType == 1 {
			gas = uint64(0)
			failed = true
		}
	} else {
		_, gas, failed, err = ApplyMessage(vmenv, tx, gp)
		if err != nil && err != ErrSpecialTxFailed {
			return nil, 0, err
		}
	}

	// Update the state with pending changes
	var root []byte
	if config.IsByzantium(header.Number) {
		statedb.Finalise(true)
	} else {
		root = statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes()
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
	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, gas, err
}
