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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/MatrixAINetwork/go-matrix/mandb"

	"github.com/MatrixAINetwork/go-matrix/core/types"

	"github.com/MatrixAINetwork/go-matrix/common"

	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/log"
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
func (w *wizard) MakeSuperGenesis(bc *core.BlockChain, db mandb.Database, num uint64, electFlag bool) {
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
		Leader:            common.HexToAddress("8111111111111111111111111111111111111111"),
		Mixhash:           parentHeader.MixDigest,
		Coinbase:          common.HexToAddress("8111111111111111111111111111111111111111"),
		Signatures:        make([]common.Signature, 0),
		Timestamp:         uint64(time.Now().Unix()),
		GasLimit:          parentHeader.GasLimit,
		Difficulty:        parentHeader.Difficulty,
		Alloc:             make(core.GenesisAlloc),
		ExtraData:         make([]byte, 8),
		Version:           string(parentHeader.Version),
		VersionSignatures: parentHeader.VersionSignatures,
		Nonce:             parentHeader.Nonce.Uint64(),
		Number:            num,
		GasUsed:           parentHeader.GasUsed,
		VrfValue:          make([]byte, 0),
	}

	sbs, err := bc.GetSuperBlockSeq()
	if nil != err {
		log.Error("Failed get SuperBlockSeq", "err", err)
		return
	}
	sbs = sbs + 1
	binary.BigEndian.PutUint64(genesis.ExtraData, sbs)
	fmt.Println("超级区块序号", sbs)
	if curHeader != nil {
		if electFlag {
			genesis.MState = new(core.GenesisMState)
			stateDB, err := bc.StateAtBlockHash(curHeader.Hash())
			if nil != err {
				log.Error("Failed get stateDB", "err", err)
				return
			}
			topology, elect, err := bc.GetGraphByState(stateDB)
			if nil != err {
				log.Error("Failed get graph", "err", err)
				return
			}
			curElect := make([]core.GenesisElect, 0)
			for _, v := range elect.ElectList {
				curElect = append(curElect, core.GenesisElect{Account: core.GenesisAddress(v.Account), Stock: v.Stock, Type: v.Type.Transfer2ElectRole(), VIP: v.VIPLevel})
			}
			genesis.MState.CurElect = &curElect
			genesis.NextElect = elect.TransferNextElect2CommonElect()

			topologyData := make([]common.NetTopologyData, 0)
			for _, v := range topology.NodeList {
				topologyData = append(topologyData, common.NetTopologyData{Account: v.Account, Position: v.Position})
			}
			genesis.NetTopology.NetTopologyData = topologyData
			genesis.NetTopology.Type = common.NetTopoTypeAll
			//stateDB.RawAccount(common.ContractAddress)
		} else {
			genesis.NetTopology.NetTopologyData = curHeader.NetTopology.NetTopologyData
			genesis.NetTopology.Type = curHeader.NetTopology.Type
			genesis.NextElect = curHeader.Elect
		}

	} else {
		if electFlag {
			genesis.MState = new(core.GenesisMState)
			stateDB, err := bc.StateAtBlockHash(parentHeader.Hash())
			if nil != err {
				log.Error("Failed get stateDB", "err", err)
				return
			}
			topology, elect, err := bc.GetGraphByState(stateDB)
			if nil != err {
				log.Error("Failed get graph", "err", err)
				return
			}
			curElect := make([]core.GenesisElect, 0)
			for _, v := range elect.ElectList {
				curElect = append(curElect, core.GenesisElect{Account: core.GenesisAddress(v.Account), Stock: v.Stock, Type: v.Type.Transfer2ElectRole(), VIP: v.VIPLevel})
			}
			genesis.MState.CurElect = &curElect
			genesis.NextElect = elect.TransferNextElect2CommonElect()

			topologyData := make([]common.NetTopologyData, 0)
			for _, v := range topology.NodeList {
				topologyData = append(topologyData, common.NetTopologyData{Account: v.Account, Position: v.Position})
			}
			genesis.NetTopology.NetTopologyData = topologyData
			genesis.NetTopology.Type = common.NetTopoTypeAll

			//genesis.Alloc = make(map[common.Address]core.GenesisAccount)
			//depositBalance := stateDB.GetBalance(common.ContractAddress)
			//depositBalance := stateDB.sets(common.ContractAddress)

			//stateDB.RawAccount(common.ContractAddress)

		} else {
			genesis.NextElect = make([]common.Elect, 0)

			genesis.NetTopology = common.NetTopology{Type: common.NetTopoTypeChange, NetTopologyData: make([]common.NetTopologyData, 0)}
		}

	}

	// Figure out which consensus engine to choose
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
