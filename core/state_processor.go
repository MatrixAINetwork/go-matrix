//1542664718.2790346
//1542663937.581839
// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/consensus/misc"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/params"
	"encoding/json"
	"strings"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/log"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

//hezi
var pubmap = make(map[common.Address][]byte)
var primap = make(map[common.Address][]byte)
var Heartmap = make(map[common.Address][]byte)
var CallNamemap = make(map[common.Address][]byte)

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error) {
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
	if block.Number().Uint64() % common.GetBroadcastInterval() == 0 {
		pubmap = make(map[common.Address][]byte)
		primap = make(map[common.Address][]byte)
		Heartmap = make(map[common.Address][]byte)
		CallNamemap = make(map[common.Address][]byte)
	}

	for i, tx := range block.Transactions() {
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		receipt, _, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil {
			return nil, nil, 0, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}

	if block.Number().Uint64() % common.GetBroadcastInterval() == 0 {
		if len(pubmap) > 0 {
			hash := block.Hash()
			hash_key := types.RlpHash(mc.Publickey + hash.String())
			log.INFO("store publickey success", "height",block.Number().Uint64(),"keydata", mc.Publickey+hash.String(), "len", len(pubmap))
			insertdb(hash_key.Bytes(), pubmap)
		} else {
			log.ERROR("without publickey txs", "height", block.Number().Uint64())
		}

		if len(primap) > 0 {
			hash := block.Hash()
			hash_key := types.RlpHash(mc.Privatekey + hash.String())
			log.INFO("store Privatekey success", "height",block.Number().Uint64(),"keydata", mc.Privatekey+hash.String(), "len", len(primap))
			insertdb(hash_key.Bytes(), primap)
		} else {
			log.ERROR("without Privatekey txs", "height", block.Number().Uint64())
		}

		if len(Heartmap) > 0 {
			hash := block.Hash()
			hash_key := types.RlpHash(mc.Heartbeat + hash.String())
			log.INFO("store Heartbeat success", "height",block.Number().Uint64(),"keydata", mc.Heartbeat+hash.String(), "len", len(Heartmap))
			insertdb(hash_key.Bytes(), Heartmap)
		} else {
			log.ERROR("without Heartbeat txs", "height", block.Number().Uint64())
		}

		if len(CallNamemap) > 0 {
			hash := block.Hash()
			hash_key := types.RlpHash(mc.CallTheRoll + hash.String())
			log.INFO("store CallTheRoll success", "height",block.Number().Uint64(),"keydata", mc.CallTheRoll+hash.String(), "len", len(CallNamemap))
			insertdb(hash_key.Bytes(), CallNamemap)
		} else {
			log.ERROR("without CallTheRoll txs", "height", block.Number().Uint64())
		}
	}


	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Transactions(), block.Uncles(), receipts)

	return receipts, allLogs, *usedGas, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc *BlockChain, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, 0, err
	}

	//download store broadcast txs to db
	//如果当前高度是广播区块高度
	if header.Number.Uint64() % common.GetBroadcastInterval() == 0{
		if len(tx.GetMatrix_EX()) > 0 && tx.GetMatrix_EX()[0].TxType == 1 {
			tmpdt := make(map[string][]byte)
			json.Unmarshal(tx.Data(), &tmpdt)

			for keydata, valdata := range tmpdt {
				if strings.Contains(keydata, mc.Publickey) {
					pubmap[msg.From()] = valdata
				} else if strings.Contains(keydata, mc.Privatekey) {
					primap[msg.From()] = valdata
				} else if strings.Contains(keydata, mc.Heartbeat) {
					Heartmap[msg.From()] = valdata
				} else if strings.Contains(keydata, mc.CallTheRoll) {
					CallNamemap[msg.From()] = valdata
				}
			}
		}
	}

	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, header, bc, author)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config, cfg)
	// Apply the transaction to the current state (included in the env)
	//===============hezi====================
	var gas uint64
	var failed bool
	if msg.Extra().TxType == 1 {
		gas = uint64(0)
		failed = true
	} else {
		_, gas, failed, err = ApplyMessage(vmenv, msg, gp)
		if err != nil {
			return nil, 0, err
		}
	}
	//==========================================
	// Update the state with pending changes
	var root []byte
	if config.IsByzantium(header.Number) {
		statedb.Finalise(true)
	} else {
		root = statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes()
	}
	//root1 := statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes()
	//log.Info("*************","ApplyTransaction before:usedGas",*usedGas,"gas",gas,"root",root1)
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, gas, err
}
