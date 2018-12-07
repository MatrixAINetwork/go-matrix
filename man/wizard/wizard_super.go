// Copyright 2017 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package wizard

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/params/manparams"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	"github.com/matrix/go-matrix/core/types"

	"github.com/matrix/go-matrix/common"

	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/log"
)

func MakeWizard(network string) *wizard {
	return &wizard{
		network: network,
		conf: config{
			Servers: make(map[string][]byte),
		},
		services: make(map[string][]string),
		in:       bufio.NewReader(os.Stdin),
	}
}

// makeGenesis creates a new genesis struct based on some user input.
func (w *wizard) MakeSuperGenesis(bc *core.BlockChain, db mandb.Database, num uint64) {
	// Construct a default genesis block
	var parentHeader, curHeader *types.Header
	if num > 1 {
		parentHeader = bc.GetHeaderByNumber(num - 1)
		curHeader = bc.GetHeaderByNumber(num)
	} else if num == 0 {
		parentHeader = bc.Genesis().Header()
		curHeader = parentHeader
	} else if num == 1 {
		parentHeader = bc.Genesis().Header()
		curHeader = bc.GetHeaderByNumber(num)
	}
	if parentHeader == nil {
		log.Error("get parent header err!")
		return
	}

	genesis := &core.Genesis{
		ParentHash:        parentHeader.Hash(),
		Leader:            common.HexToAddress("0x8111111111111111111111111111111111111111"),
		Mixhash:           parentHeader.MixDigest,
		Coinbase:          manparams.InnerMinerNodes[0].Address,
		Signatures:        make([]common.Signature, 0),
		Timestamp:         uint64(time.Now().Unix()),
		GasLimit:          parentHeader.GasLimit,
		Difficulty:        parentHeader.Difficulty,
		Alloc:             make(core.GenesisAlloc),
		ExtraData:         make([]byte, 0),
		Version:           string(parentHeader.Version),
		VersionSignatures: parentHeader.VersionSignatures,
		Nonce:             parentHeader.Nonce.Uint64(),
		Number:            num,
		GasUsed:           parentHeader.GasUsed,
	}

	if curHeader != nil {
		genesis.Elect = curHeader.Elect
		genesis.NetTopology = curHeader.NetTopology
	} else {
		genesis.Elect = make([]common.Elect, 0)
		genesis.NetTopology = common.NetTopology{Type: common.NetTopoTypeChange, NetTopologyData: make([]common.NetTopologyData, 0)}
	}

	// Figure out which consensus engine to choose
	genesis.Alloc[common.BlkMinerRewardAddress] = core.GenesisAccount{Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(200), big.NewInt(0))}
	genesis.Alloc[common.BlkValidatorRewardAddress] = core.GenesisAccount{Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(200), big.NewInt(0))}
	genesis.Alloc[common.TxGasRewardAddress] = core.GenesisAccount{Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(200), big.NewInt(0))}
	genesis.Alloc[common.LotteryRewardAddress] = core.GenesisAccount{Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(200), big.NewInt(0))}
	// All done, store the genesis and flush to disk
	log.Info("Configured new genesis block")

	w.conf.Genesis = genesis
	//w.conf.flush()
	fmt.Printf("Which file to save the genesis into %s", w.network)
	out, _ := json.MarshalIndent(w.conf.Genesis, "", "  ")
	if err := ioutil.WriteFile(w.network, out, 0644); err != nil {
		log.Error("Failed to save genesis file", "err", err)
		return
	}
	if err := json.Unmarshal(out, w.conf.Genesis); err != nil {
		log.Error("Failed to save genesis file", "err", err)
		return
	}
	if err := ioutil.WriteFile(w.network, out, 0644); err != nil {
		log.Error("Failed to save genesis file", "err", err)
		return
	}
	log.Info("Exported existing genesis block")
}
