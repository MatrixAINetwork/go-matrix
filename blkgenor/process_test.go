// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/accounts"
	"github.com/MatrixAINetwork/go-matrix/accounts/keystore"
	"github.com/MatrixAINetwork/go-matrix/accounts/signhelper"
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/consensus/manash"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/rawdb"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	_ "github.com/MatrixAINetwork/go-matrix/crypto/vrf"
	"github.com/MatrixAINetwork/go-matrix/depoistInfo"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/msgsend"
	"github.com/MatrixAINetwork/go-matrix/olconsensus"
	"github.com/MatrixAINetwork/go-matrix/p2p"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/pod"
	_ "github.com/MatrixAINetwork/go-matrix/random/electionseed"
	_ "github.com/MatrixAINetwork/go-matrix/random/ereryblockseed"
	_ "github.com/MatrixAINetwork/go-matrix/random/everybroadcastseed"
	"github.com/MatrixAINetwork/go-matrix/reelection"
	"io/ioutil"
	"math/big"
	"sync"
)

// Tests that a node embedded within a console can be started up properly and
// then terminated by closing the input stream.
var myNodeId string = "4b2f638f46c7ae5b1564ca7015d716621848a0d9be66f1d1e91d566d2a70eedc2f11e92b743acb8d97dec3fb412c1b2f66afd7fbb9399d4fb2423619eaa51411"
var (
	testNodeKey, _ = crypto.GenerateKey()
)

func testNodeConfig() *pod.Config {
	return &pod.Config{
		Name: "test node",
		P2P:  p2p.Config{PrivateKey: testNodeKey},
	}
}

type FakeEth struct {
	txPool      *core.TxPoolManager
	blockchain  *core.BlockChain
	eventMux    *event.TypeMux
	engine      consensus.Engine
	hd          *msgsend.HD            //node传进来的
	reelection  *reelection.ReElection //换届服务
	signHelper  *signhelper.SignHelper
	blockgen    *BlockGenor
	once        *sync.Once
	fetchhash   common.Hash
	fetchnum    uint64
	olConsensus *olconsensus.TopNodeService
	random      *baseinterface.Random
}

type testDPOSEngine struct {
}

func (tsdpos *testDPOSEngine) CheckSuperBlock(header *types.Header) error {

	return nil
}

func (tsdpos *testDPOSEngine) VerifyBlock(reader consensus.StateReader, header *types.Header) error {

	return nil
}

func (tsdpos *testDPOSEngine) VerifyVersion(reader consensus.StateReader, header *types.Header) error {

	return nil
}

//verify hash in current block
func (tsdpos *testDPOSEngine) VerifyHash(reader consensus.StateReader, signHash common.Hash, signs []common.Signature) ([]common.Signature, error) {
	return nil, nil
}

func (tsdpos *testDPOSEngine) VerifyHashWithVerifiedSignsAndBlock(reader consensus.StateReader, signs []*common.VerifiedSign, blockHash common.Hash) ([]common.Signature, error) {
	return nil, nil
}
func (tsdpos *testDPOSEngine) VerifyHashWithBlock(reader consensus.StateReader, signHash common.Hash, signs []common.Signature, blockHash common.Hash) ([]common.Signature, error) {
	return nil, nil
}

//VerifyHashWithStocks(signHash common.Hash, signs []common.Signature, stocks map[common.Address]uint16) ([]common.Signature, error)

func (tsdpos *testDPOSEngine) VerifyHashWithVerifiedSigns(reader consensus.StateReader, signs []*common.VerifiedSign) ([]common.Signature, error) {
	return nil, nil
}

func (s *FakeEth) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *FakeEth) TxPool() *core.TxPoolManager        { return s.txPool }
func (s *FakeEth) EventMux() *event.TypeMux           { return s.eventMux }
func (s *FakeEth) Engine() consensus.Engine           { return s.engine }
func (s *FakeEth) DPOSEngine() consensus.DPOSEngine   { return s.blockchain.DPOSEngine() }
func (s *FakeEth) SignHelper() *signhelper.SignHelper { return s.signHelper }
func (s *FakeEth) HD() *msgsend.HD                    { return s.hd }
func (s *FakeEth) ReElection() *reelection.ReElection {
	return s.reelection
}
func (s *FakeEth) FetcherNotify(hash common.Hash, number uint64, addr common.Address) {
	s.fetchhash = hash
	s.fetchnum = number
	return
}
func (s *FakeEth) OLConsensus() *olconsensus.TopNodeService {
	return s.olConsensus
}
func (s *FakeEth) Random() *baseinterface.Random {
	return s.random
}
func toBLock(g *core.Genesis, db mandb.Database) *types.Block {
	if db == nil {
		db = mandb.NewMemDatabase()
	}
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
	for addr, account := range g.Alloc {
		statedb.AddBalance(common.MainAccount, addr, account.Balance)
		///*******************************************************/
		////  应该是通过发特殊交易添加账户
		//statedb.AddBalance(common.LockAccount,addr, account.Balance)
		//statedb.AddBalance(common.EntrustAccount,addr, account.Balance)
		//statedb.AddBalance(common.FreezeAccount,addr, account.Balance)
		///*******************************************************/
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}

	root := statedb.IntermediateRoot(false)
	head := &types.Header{
		Number:            new(big.Int).SetUint64(g.Number),
		Nonce:             types.EncodeNonce(g.Nonce),
		Time:              new(big.Int).SetUint64(g.Timestamp),
		ParentHash:        g.ParentHash,
		Extra:             g.ExtraData,
		Version:           []byte(g.Version),
		VersionSignatures: g.VersionSignatures,
		VrfValue:          g.VrfValue,
		Elect:             g.NextElect,
		NetTopology:       g.NetTopology,
		Signatures:        g.Signatures,
		Leader:            g.Leader,
		GasLimit:          g.GasLimit,
		GasUsed:           g.GasUsed,
		Difficulty:        g.Difficulty,
		MixDigest:         g.Mixhash,
		Coinbase:          g.Coinbase,
		Root:              root,
	}
	if g.GasLimit == 0 {
		head.GasLimit = params.GenesisGasLimit
	} else if g.GasLimit < params.MinGasLimit {
		head.GasLimit = params.MinGasLimit
	}
	if g.Difficulty == nil {
		head.Difficulty = params.GenesisDifficulty
	}
	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, true)

	return types.NewBlock(head, nil, nil, nil)
}

const (
	veryLightScryptN = 2
	veryLightScryptP = 1
)

func tmpKeyStore(encrypted bool) (string, *keystore.KeyStore) {
	d, err := ioutil.TempDir("", "man-keystore-test")
	if err != nil {
		return d, nil
	}
	new := keystore.NewPlaintextKeyStore
	if encrypted {
		new = func(kd string) *keystore.KeyStore {
			return keystore.NewKeyStore(kd, veryLightScryptN, veryLightScryptP)
		}
	}
	return d, new(d)
}

func fakeEthNew(n int) *FakeEth {
	man := &FakeEth{once: new(sync.Once), eventMux: new(event.TypeMux), signHelper: signhelper.NewSignHelper()}

	man.once.Do(func() {
		id, _ := discover.HexID(myNodeId)

		go func() {
			//var guard *monkey.PatchGuard
			monkey.Patch(mc.PublishEvent, func(aim mc.EventCode, data interface{}) error {
				return nil
			})
			monkey.Patch(depoistInfo.GetDepositList, func(tm *big.Int, getDeposit common.RoleType) ([]vm.DepositDetail, error) {

				return nil, nil
			})
			monkey.Patch(ca.GetElectedByHeightWithdraw, func(height *big.Int) ([]vm.DepositDetail, error) {
				fmt.Println("use my GetElectedByHeightWithdraw")
				return nil, nil
			})

			monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
				fmt.Println("use my GetElectedByHeightAndRole")
				return nil, nil
			})

			monkey.Patch(ca.GetElectedByHeightWithdraw, func(tm *big.Int) ([]vm.DepositDetail, error) {

				return nil, nil
			})

			monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {

				return nil, nil
			})

			ca.Start(id, "")
		}()

		//prkey, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		//man.signHelper.SetTestMode(prkey)

		// Ensure that the AccountManager method works before the node has started.
		// We rely on this in cmd/gman.

		signHelper := signhelper.NewSignHelper()
		man.signHelper = signHelper

		hd, err := msgsend.NewHD()
		if err != nil {
			return
		}
		man.hd = hd
		db := mandb.NewMemDatabase()
		genesis := new(core.Genesis)

		block := toBLock(genesis, db)
		if nil == block {
			return
		}
		if block.Number().Sign() != 0 {
			return
		}
		rawdb.WriteTd(db, block.Hash(), block.NumberU64(), genesis.Difficulty)
		rawdb.WriteBlock(db, block)
		rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil)
		rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
		rawdb.WriteHeadBlockHash(db, block.Hash())
		rawdb.WriteHeadHeaderHash(db, block.Hash())

		config := genesis.Config
		if config == nil {
			config = params.AllManashProtocolChanges
		}
		rawdb.WriteChainConfig(db, block.Hash(), config)
		blockchain, _ := core.NewBlockChain(db, nil, params.AllManashProtocolChanges, manash.NewFaker(), vm.Config{})
		//core.NewCanonical()
		man.signHelper.SetAuthReader(blockchain)

		_, keystore := tmpKeyStore(false)
		backends := []accounts.Backend{
			keystore,
		}

		entrustValue := make(map[common.Address]string, 0)

		entrustValue[common.HexToAddress("0x6a3217d128a76e4777403e092bde8362d4117773")] = "xxx"
		man.signHelper.SetAccountManager(accounts.NewManager(backends...))
		manparams.EntrustAccountValue.SetEntrustValue(entrustValue)
		fakedpos := &testDPOSEngine{}
		blockchain.SetDposEngine(fakedpos)
		if err != nil {
			fmt.Println("failed to create pristine chain: ", err)
			return
		}
		//defer blockchain.Stop()
		man.blockchain = blockchain

		man.txPool = core.NewTxPoolManager(core.DefaultTxPoolConfig, params.TestChainConfig, blockchain, "")

		man.random, err = baseinterface.NewRandom(man)
		if err != nil {
			return
		}
		reElection, err := reelection.New(man.BlockChain(), man.random)
		if err != nil {
			return
		}
		man.reelection = reElection
		man.olConsensus = olconsensus.NewTopNodeService(man.blockchain.DPOSEngine())
		topNodeInstance := olconsensus.NewTopNodeInstance(man.signHelper, man.hd)
		man.olConsensus.SetValidatorReader(man.blockchain)
		man.olConsensus.SetStateReaderInterface(man.blockchain)
		man.olConsensus.SetTopNodeStateInterface(topNodeInstance)
		man.olConsensus.SetValidatorAccountInterface(topNodeInstance)
		man.olConsensus.SetMessageSendInterface(topNodeInstance)
		man.olConsensus.SetMessageCenterInterface(topNodeInstance)
		man.blockgen, err = New(man)
		if err != nil {
			return
		}

	})
	return man
}
