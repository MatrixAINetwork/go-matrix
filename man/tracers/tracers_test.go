// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package tracers

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"github.com/MatrixAINetwork/go-matrix/common/math"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"github.com/MatrixAINetwork/go-matrix/tests"
)

// To generate a new callTracer test, copy paste the makeTest method below into
// a Gman console and call it with a transaction hash you which to export.

/*
// makeTest generates a callTracer test by running a prestate reassembled and a
// call trace run, assembling all the gathered information into a test case.
var makeTest = function(tx, rewind) {
  // Generate the genesis block from the block, transaction and prestate data
  var block   = man.getBlock(man.getTransaction(tx).blockHash);
  var genesis = man.getBlock(block.parentHash);

  delete genesis.gasUsed;
  delete genesis.logsBloom;
  delete genesis.parentHash;
  delete genesis.receiptsRoot;
  delete genesis.sha3Uncles;
  delete genesis.size;
  delete genesis.transactions;
  delete genesis.transactionsRoot;
  delete genesis.uncles;

  genesis.gasLimit  = genesis.gasLimit.toString();
  genesis.number    = genesis.number.toString();
  genesis.timestamp = genesis.timestamp.toString();

  genesis.alloc = debug.traceTransaction(tx, {tracer: "prestateTracer", rewind: rewind});
  for (var key in genesis.alloc) {
    genesis.alloc[key].nonce = genesis.alloc[key].nonce.toString();
  }
  genesis.config = admin.nodeInfo.protocols.man.config;

  // Generate the call trace and produce the test input
  var result = debug.traceTransaction(tx, {tracer: "callTracer", rewind: rewind});
  delete result.time;

  console.log(JSON.stringify({
    genesis: genesis,
    context: {
      number:     block.number.toString(),
      difficulty: block.difficulty,
      timestamp:  block.timestamp.toString(),
      gasLimit:   block.gasLimit.toString(),
      miner:      block.miner,
    },
    input:  man.getRawTransaction(tx),
    result: result,
  }, null, 2));
}
*/

// callTrace is the result of a callTracer run.
type callTrace struct {
	Type    string          `json:"type"`
	From    common.Address  `json:"from"`
	To      common.Address  `json:"to"`
	Input   hexutil.Bytes   `json:"input"`
	Output  hexutil.Bytes   `json:"output"`
	Gas     *hexutil.Uint64 `json:"gas,omitempty"`
	GasUsed *hexutil.Uint64 `json:"gasUsed,omitempty"`
	Value   *hexutil.Big    `json:"value,omitempty"`
	Error   string          `json:"error,omitempty"`
	Calls   []callTrace     `json:"calls,omitempty"`
}

type callContext struct {
	Number     math.HexOrDecimal64   `json:"number"`
	Difficulty *math.HexOrDecimal256 `json:"difficulty"`
	Time       math.HexOrDecimal64   `json:"timestamp"`
	GasLimit   math.HexOrDecimal64   `json:"gasLimit"`
	Miner      common.Address        `json:"miner"`
}

// callTracerTest defines a single test to check the call tracer against.
type callTracerTest struct {
	Genesis *core.Genesis `json:"genesis"`
	Context *callContext  `json:"context"`
	Input   string        `json:"input"`
	Result  *callTrace    `json:"result"`
}

// Iterates over all the input-output datasets in the tracer test harness and
// runs the JavaScript tracers against them.
func TestCallTracer(t *testing.T) {
	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		t.Fatalf("failed to retrieve tracer test suite: %v", err)
	}
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), "call_tracer_") {
			continue
		}
		file := file // capture range variable
		t.Run(camel(strings.TrimSuffix(strings.TrimPrefix(file.Name(), "call_tracer_"), ".json")), func(t *testing.T) {
			t.Parallel()

			// Call tracer test found, read if from disk
			blob, err := ioutil.ReadFile(filepath.Join("testdata", file.Name()))
			if err != nil {
				t.Fatalf("failed to read testcase: %v", err)
			}
			test := new(callTracerTest)
			if err := json.Unmarshal(blob, test); err != nil {
				t.Fatalf("failed to parse testcase: %v", err)
			}
			// Configure a blockchain with the given prestate
			tx := new(types.Transaction)
			if err := rlp.DecodeBytes(common.FromHex(test.Input), tx); err != nil {
				t.Fatalf("failed to parse testcase input: %v", err)
			}
			signer := types.MakeSigner(test.Genesis.Config, new(big.Int).SetUint64(uint64(test.Context.Number)))
			origin, _ := signer.Sender(tx)

			context := vm.Context{
				CanTransfer: core.CanTransfer,
				Transfer:    core.Transfer,
				Origin:      origin,
				Coinbase:    test.Context.Miner,
				BlockNumber: new(big.Int).SetUint64(uint64(test.Context.Number)),
				Time:        new(big.Int).SetUint64(uint64(test.Context.Time)),
				Difficulty:  (*big.Int)(test.Context.Difficulty),
				GasLimit:    uint64(test.Context.GasLimit),
				GasPrice:    tx.GasPrice(),
			}
			statedb := tests.MakePreState(mandb.NewMemDatabase(), test.Genesis.Alloc)

			// Create the tracer, the EVM environment and run it
			tracer, err := New("callTracer")
			if err != nil {
				t.Fatalf("failed to create call tracer: %v", err)
			}
			evm := vm.NewEVM(context, statedb, test.Genesis.Config, vm.Config{Debug: true, Tracer: tracer})

			msg, err := tx.AsMessage(signer)
			if err != nil {
				t.Fatalf("failed to prepare transaction for tracing: %v", err)
			}
			st := core.NewStateTransition(evm, msg, new(core.GasPool).AddGas(tx.Gas()))
			if _, _, _, err = st.TransitionDb(); err != nil {
				t.Fatalf("failed to execute transaction: %v", err)
			}
			// Retrieve the trace result and compare against the etalon
			res, err := tracer.GetResult()
			if err != nil {
				t.Fatalf("failed to retrieve trace result: %v", err)
			}
			ret := new(callTrace)
			if err := json.Unmarshal(res, ret); err != nil {
				t.Fatalf("failed to unmarshal trace result: %v", err)
			}
			if !reflect.DeepEqual(ret, test.Result) {
				t.Fatalf("trace mismatch: have %+v, want %+v", ret, test.Result)
			}
		})
	}
}
