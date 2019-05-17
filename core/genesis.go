// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/MatrixAINetwork/go-matrix/base58"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"github.com/MatrixAINetwork/go-matrix/common/math"
	"github.com/MatrixAINetwork/go-matrix/core/rawdb"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"sort"
)

//go:generate gencodec -type Genesis -field-override genesisSpecMarshaling -out gen_genesis.go
//go:generate gencodec -type GenesisAccount -field-override genesisAccountMarshaling -out gen_genesis_account.go

var errGenesisNoConfig = errors.New("genesis has no chain configuration")
var errGenGenesisBlockNoConfig = errors.New("no genesis cfg and no genesis block")
var errGenesisLostChainCfg = errors.New("genesis block lost chaincfg")

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type Genesis struct {
	Config            *params.ChainConfig `json:"config,omitempty"`
	Nonce             uint64              `json:"nonce"`
	Timestamp         uint64              `json:"timestamp"    gencodec:"required"`
	ExtraData         []byte              `json:"extraData"`
	Version           string              `json:"version"    gencodec:"required"`
	VersionSignatures []common.Signature  `json:"versionSignatures"    gencodec:"required"`
	VrfValue          []byte              `json:"vrfvalue"`
	Leader            common.Address      `json:"leader"`
	NextElect         []common.Elect      `json:"nextElect"    gencodec:"required"`
	NetTopology       common.NetTopology  `json:"nettopology"       gencodec:"required"`
	Signatures        []common.Signature  `json:"signatures" gencodec:"required"`

	GasLimit   uint64         `json:"gasLimit"   gencodec:"required"`
	Difficulty *big.Int       `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash    `json:"mixHash"`
	Coinbase   common.Address `json:"coinbase"`
	Alloc      GenesisAlloc   `json:"alloc"      gencodec:"required"`
	MState     *GenesisMState `json:"mstate"    gencodec:"required"`
	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64                        `json:"number"`
	GasUsed    uint64                        `json:"gasUsed"`
	ParentHash common.Hash                   `json:"parentHash"`
	Roots      []common.CoinRoot             `json:"stateRoot"        gencodec:"required"`
	Sharding   []common.Coinbyte             `json:"sharding,omitempty"`
	Currencys  map[string][]Genesiscurrencys `json:"currencys"`
}

type Genesiscurrencys struct {
	Account string
	Quant   *big.Int
}

func (gc Genesiscurrencys) MarshalJSON() ([]byte, error) {
	type Genesiscurrencys struct {
		Account string                `json:"Account" gencodec:"required"`
		Quant   *math.HexOrDecimal256 `json:"Quant" gencodec:"required"`
	}
	var enc Genesiscurrencys
	enc.Account = gc.Account
	enc.Quant = (*math.HexOrDecimal256)(gc.Quant)
	return json.Marshal(&enc)
}
func (g *Genesiscurrencys) UnmarshalJSON(input []byte) error {
	type Genesiscurrencys struct {
		Account string                `json:"Account" gencodec:"required"`
		Quant   *math.HexOrDecimal256 `json:"Quant" gencodec:"required"`
	}
	var dec Genesiscurrencys
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	g.Account = dec.Account
	g.Quant = (*big.Int)(dec.Quant)
	return nil
}

// GenesisAlloc specifies the initial state that is part of the genesis block.
type GenesisAlloc map[common.Address]GenesisAccount

func (ga *GenesisAlloc) UnmarshalJSON(data []byte) error {
	m := make(map[common.UnprefixedAddress]GenesisAccount)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*ga = make(GenesisAlloc)
	for addr, a := range m {
		(*ga)[common.Address(addr)] = a
	}
	return nil
}

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Code       []byte                      `json:"code,omitempty"`
	Storage    map[common.Hash]common.Hash `json:"storage,omitempty"`
	Balance    *big.Int                    `json:"balance" gencodec:"required"`
	Nonce      uint64                      `json:"nonce,omitempty"`
	PrivateKey []byte                      `json:"secretKey,omitempty"` // for tests
}

// field type overrides for gencodec
type genesisSpecMarshaling struct {
	Nonce      math.HexOrDecimal64
	Timestamp  math.HexOrDecimal64
	ExtraData  hexutil.Bytes
	GasLimit   math.HexOrDecimal64
	GasUsed    math.HexOrDecimal64
	Number     math.HexOrDecimal64
	Difficulty *math.HexOrDecimal256
	Alloc      map[common.UnprefixedAddress]GenesisAccount
}

type genesisAccountMarshaling struct {
	Code       hexutil.Bytes
	Balance    *math.HexOrDecimal256
	Nonce      math.HexOrDecimal64
	Storage    map[storageJSON]storageJSON
	PrivateKey hexutil.Bytes
}

// storageJSON represents a 256 bit byte array, but allows less than 256 bits when
// unmarshaling from hex.
type storageJSON common.Hash

func (h *storageJSON) UnmarshalText(text []byte) error {
	text = bytes.TrimPrefix(text, []byte("0x"))
	if len(text) > 64 {
		return fmt.Errorf("too many hex characters in storage key/value %q", text)
	}
	offset := len(h) - len(text)/2 // pad on the left
	if _, err := hex.Decode(h[offset:], text); err != nil {
		fmt.Println(err)
		return fmt.Errorf("invalid hex storage key/value %q", text)
	}
	return nil
}

func (h storageJSON) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New common.Hash
}

func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database already contains an incompatible genesis block (have %x, new %x)", e.Stored[:8], e.New[:8])
}

// SetupGenesisBlock writes or updates the genesis block in db.
// The block that will be used is:
//
//                          genesis == nil       genesis != nil
//                       +------------------------------------------
//     db has no genesis |  main-net default  |  genesis
//     db has genesis    |  from DB           |  genesis (if compatible)
//
// The stored chain configuration will be updated if it is compatible (i.e. does not
// specify a fork block below the local head block). In case of a conflict, the
// error is a *params.ConfigCompatError and the new, unwritten config is returned.
//
// The returned chain configuration is never nil.
func SetupGenesisBlock(db mandb.Database, genesis *Genesis) (*params.ChainConfig, common.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return params.AllManashProtocolChanges, common.Hash{}, errGenesisNoConfig
	}

	// Just commit the new block if there is no stored genesis block.
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		if genesis == nil {
			log.Error("Without GenesisBlock and GenesisCfg")
			return nil, common.Hash{}, errGenGenesisBlockNoConfig
		}

		log.Info("Writing custom genesis block")
		block, err := genesis.Commit(db)
		if err != nil {
			return nil, common.Hash{}, errGenGenesisBlockNoConfig
		}
		return genesis.Config, block.Hash(), err
	}

	// Check whether the genesis block is already written.
	if genesis != nil {
		block, err := genesis.ToBlock(nil)
		if err != nil {
			return nil, common.Hash{}, err
		}
		if block.Hash() != stored {
			return genesis.Config, block.Hash(), &GenesisMismatchError{stored, block.Hash()}
		}
	}

	// Get the existing chain configuration.
	newcfg := genesis.configOrDefault(stored)
	storedcfg := rawdb.ReadChainConfig(db, stored)
	if storedcfg == nil {
		log.Warn("Genesis Block Lost Cfg")
		return newcfg, stored, errGenesisLostChainCfg
	}
	if genesis == nil {
		return storedcfg, stored, nil
	}
	// Check config compatibility and write the config. Compatibility errors
	// are returned to the caller unless we're already at block zero.
	height := rawdb.ReadHeaderNumber(db, rawdb.ReadHeadHeaderHash(db))
	if height == nil {
		return newcfg, stored, fmt.Errorf("missing block number for head header hash")
	}
	compatErr := storedcfg.CheckCompatible(newcfg, *height)
	if compatErr != nil && *height != 0 && compatErr.RewindTo != 0 {
		return newcfg, stored, compatErr
	}
	rawdb.WriteChainConfig(db, stored, newcfg)
	return newcfg, stored, nil
}

func (g *Genesis) configOrDefault(ghash common.Hash) *params.ChainConfig {
	switch {
	case g != nil:
		return g.Config
	default:
		return params.AllManashProtocolChanges
	}
}
func sortMapByString(tmpMap map[string][]Genesiscurrencys) []string {
	clist := make([]string, 0, len(tmpMap))
	for k, _ := range tmpMap {
		clist = append(clist, k)
	}
	sort.Strings(clist)
	return clist
}

// ToBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func (g *Genesis) ToBlock(db mandb.Database) (*types.Block, error) {
	if db == nil {
		db = mandb.NewMemDatabase()
	}
	roots := make([]common.CoinRoot, 0, len(g.Currencys)+1)
	roots = append(roots, common.CoinRoot{Cointyp: params.MAN_COIN, Root: common.Hash{}})
	var coinlist []string
	var coincfglist []common.CoinConfig
	if len(g.Currencys) > 0 {
		for _, coinname := range sortMapByString(g.Currencys) {
			roots = append(roots, common.CoinRoot{Cointyp: coinname, Root: common.Hash{}})
			coinlist = append(coinlist, coinname)
			allAmount := new(big.Int)
			for _, cuff := range g.Currencys[coinname] {
				allAmount = new(big.Int).Add(allAmount, cuff.Quant)
			}
			coincfglist = append(coincfglist, common.CoinConfig{CoinType: coinname, CoinTotal: (*hexutil.Big)(allAmount),
				PackNum: params.CallTxPachNum, CoinUnit: (*hexutil.Big)(new(big.Int).SetUint64(params.CoinTypeUnit)),
				CoinAddress: common.TxGasRewardAddress, CoinRange: coinname})
		}
	}
	statedb, _ := state.NewStateDBManage(roots, db, state.NewDatabase(db))
	for addr, account := range g.Alloc {
		statedb.AddBalance(params.MAN_COIN, common.MainAccount, addr, account.Balance)
		statedb.SetCode(params.MAN_COIN, addr, account.Code)
		statedb.SetNonce(params.MAN_COIN, addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(params.MAN_COIN, addr, key, value)
		}
	}
	key := types.RlpHash(params.COIN_NAME)
	coinby, _ := json.Marshal(coinlist)
	statedb.SetMatrixData(key, coinby)
	coinCfgbs, _ := json.Marshal(coincfglist)
	statedb.SetMatrixData(types.RlpHash(common.COINPREFIX+mc.MSCurrencyConfig), coinCfgbs)
	for _, cname := range sortMapByString(g.Currencys) {
		for _, cuff := range g.Currencys[cname] {
			addr, err := base58.Base58DecodeToAddress(cuff.Account)
			if err != nil {
				continue
			}
			//tmpval,_:=new(big.Int).SetString(cuff.Quant,0)
			statedb.AddBalance(cname, common.MainAccount, addr, cuff.Quant)
		}
	}
	if nil == g.MState {
		log.Error("genesis", "设置matrix状态树错误", "g.MState = nil")
		return nil, errors.New("MState of genesis is nil")
	}
	if err := g.MState.setMatrixState(statedb, g.NetTopology, g.NextElect, g.Version, g.Version, g.Number); err != nil {
		log.Error("genesis", "MState.setMatrixState err", err)
		return nil, err
	}

	if err := g.MState.SetSuperBlkToState(statedb, g.ExtraData, g.Number); err != nil {
		log.Error("genesis", "MState.SetSuperBlkToState err", err)
		return nil, err
	}
	roots, sharding := statedb.IntermediateRoot(false)
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
		Roots:             make([]common.CoinRoot, len(roots)),
		Sharding:          make([]common.Coinbyte, len(sharding)),
	}
	copy(head.Roots, roots)
	copy(head.Sharding, sharding)
	if g.GasLimit == 0 {
		head.GasLimit = params.GenesisGasLimit
	} else if g.GasLimit < params.MinGasLimit {
		head.GasLimit = params.MinGasLimit
	}
	if g.Difficulty == nil {
		head.Difficulty = params.GenesisDifficulty
	}
	statedb.Commit(false)
	statedb.Database().TrieDB().CommitRoots(roots, true)

	return types.NewBlock(head, nil, nil), nil
}

func (g *Genesis) GenSuperBlock(parentHeader *types.Header, mdb mandb.Database, sdb state.Database, chainCfg *params.ChainConfig) *types.Block {
	if nil == parentHeader || nil == sdb {
		log.ERROR("genesis super block", "param err", "nil")
		return nil
	}

	stateDB, err := state.NewStateDBManage(parentHeader.Roots, mdb, sdb)
	if err != nil {
		log.Error("genesis super block", "get parent state db err", err)
		return nil
	}

	if nil != g.MState {
		if err := g.MState.setMatrixState(stateDB, g.NetTopology, g.NextElect, g.Version, string(parentHeader.Version), g.Number); err != nil {
			log.Error("genesis super block", "设置matrix状态树错误", err)
			return nil
		}
	} else {
		mState := new(GenesisMState)
		if err := mState.setMatrixState(stateDB, g.NetTopology, g.NextElect, g.Version, string(parentHeader.Version), g.Number); err != nil {
			log.Error("genesis super block", "mstate参数为nil时, 设置matrix状态树错误", err)
			return nil
		}
	}
	if err := g.MState.SetSuperBlkToState(stateDB, g.ExtraData, g.Number); err != nil {
		log.Error("genesis", "设置matrix状态树错误", err)
		return nil
	}
	head := &types.Header{
		Number:            new(big.Int).SetUint64(g.Number),
		Nonce:             types.EncodeNonce(g.Nonce),
		Time:              new(big.Int).SetUint64(g.Timestamp),
		ParentHash:        g.ParentHash,
		Extra:             g.ExtraData,
		Version:           []byte(g.Version),
		VersionSignatures: g.VersionSignatures,
		Elect:             g.NextElect,
		NetTopology:       g.NetTopology,
		Signatures:        g.Signatures,
		Leader:            g.Leader,
		GasLimit:          g.GasLimit,
		GasUsed:           g.GasUsed,
		Difficulty:        g.Difficulty,
		MixDigest:         g.Mixhash,
		Coinbase:          g.Coinbase,
	}

	head.Roots, head.Sharding = stateDB.IntermediateRoot(chainCfg.IsEIP158(head.Number))
	if g.GasLimit == 0 {
		head.GasLimit = params.GenesisGasLimit
	}
	if g.Difficulty == nil {
		head.Difficulty = params.GenesisDifficulty
	}

	// 创建超级区块交易
	txs := make([]types.SelfTransaction, 0)
	g.Alloc = make(GenesisAlloc)
	data, err := json.Marshal(g.Alloc)
	if err != nil {
		log.ERROR("genesis super block", "marshal alloc info err", err)
		return nil
	}
	tx0 := types.NewTransaction(g.Number, common.Address{}, nil, 0, nil, data, nil, nil, nil, common.ExtraSuperBlockTx, 0, params.MAN_COIN, 0)
	if tx0 == nil {
		log.ERROR("genesis super block", "create super block tx err", "NewTransaction return nil")
		return nil
	}
	txs = append(txs, tx0)

	var msData []byte = nil
	if nil != g.MState {
		msData, err = json.Marshal(g.MState)
		if err != nil {
			log.ERROR("genesis super block", "marshal alloc info err", err)
			return nil
		}
	}
	txMState := types.NewTransaction(g.Number, common.Address{}, nil, 1, nil, msData, nil, nil, nil, common.ExtraSuperBlockTx, 0, params.MAN_COIN, 0)
	if txMState == nil {
		log.ERROR("genesis super block", "create super block matrix state tx err", "NewTransaction return nil")
		return nil
	}
	txs = append(txs, txMState)

	var cts []types.CoinSelfTransaction
	ct := types.CoinSelfTransaction{params.MAN_COIN, txs}
	cts = append(cts, ct)

	return types.NewBlock(head, types.MakeCurencyBlock(cts, nil, nil), nil)
}
func (g *Genesis) ToSuperBlock() *types.Block {
	head := &types.Header{
		Number:            new(big.Int).SetUint64(g.Number),
		Nonce:             types.EncodeNonce(g.Nonce),
		Time:              new(big.Int).SetUint64(g.Timestamp),
		ParentHash:        g.ParentHash,
		Extra:             g.ExtraData,
		Version:           []byte(g.Version),
		VersionSignatures: g.VersionSignatures,
		Elect:             g.NextElect,
		NetTopology:       g.NetTopology,
		Signatures:        g.Signatures,
		Leader:            g.Leader,
		GasLimit:          g.GasLimit,
		GasUsed:           g.GasUsed,
		Difficulty:        g.Difficulty,
		MixDigest:         g.Mixhash,
		Coinbase:          g.Coinbase,
		Roots:             g.Roots,
		Sharding:          g.Sharding,
	}

	// 创建超级区块交易
	txs := make([]types.SelfTransaction, 0)
	g.Alloc = make(GenesisAlloc)
	data, err := json.Marshal(g.Alloc)
	if err != nil {
		log.ERROR("genesis super block", "marshal alloc info err", err)
		return nil
	}
	tx0 := types.NewTransaction(g.Number, common.Address{}, nil, 0, nil, data, nil, nil, nil, common.ExtraSuperBlockTx, 0, params.MAN_COIN, 0)
	if tx0 == nil {
		log.ERROR("genesis super block", "create super block tx err", "NewTransaction return nil")
		return nil
	}
	txs = append(txs, tx0)

	var msData []byte = nil
	if nil != g.MState {
		msData, err = json.Marshal(g.MState)
		if err != nil {
			log.ERROR("genesis super block", "marshal alloc info err", err)
			return nil
		}
	}
	txMState := types.NewTransaction(g.Number, common.Address{}, nil, 1, nil, msData, nil, nil, nil, common.ExtraSuperBlockTx, 0, params.MAN_COIN, 0)
	if txMState == nil {
		log.ERROR("genesis super block", "create super block matrix state tx err", "NewTransaction return nil")
		return nil
	}
	txs = append(txs, txMState)

	var cts []types.CoinSelfTransaction
	ct := types.CoinSelfTransaction{params.MAN_COIN, txs}
	cts = append(cts, ct)

	return types.NewBlock(head, types.MakeCurencyBlock(cts, nil, nil), nil)
}

// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db mandb.Database) (*types.Block, error) {
	block, err := g.ToBlock(db)
	if err != nil || nil == block {
		return nil, fmt.Errorf("can't create genesis block, err = %v", err)
	}
	if block.Number().Sign() != 0 {
		return nil, fmt.Errorf("can't commit genesis block with number > 0")
	}
	rawdb.WriteTd(db, block.Hash(), block.NumberU64(), g.Difficulty)
	rawdb.WriteBlock(db, block)
	rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash())
	rawdb.WriteHeadHeaderHash(db, block.Hash())

	config := g.Config
	if config == nil {
		config = params.AllManashProtocolChanges
	}
	rawdb.WriteChainConfig(db, block.Hash(), config)
	return block, nil
}

// MustCommit writes the genesis block and state to db, panicking on error.
// The block is committed as the canonical head block.
func (g *Genesis) MustCommit(db mandb.Database) *types.Block {
	block, err := g.Commit(db)
	if err != nil {
		panic(err)
	}
	return block
}

// GenesisBlockForTesting creates and writes a block in which addr has the given wei balance.
func GenesisBlockForTesting(db mandb.Database, addr common.Address, balance *big.Int) *types.Block {
	g := Genesis{Alloc: GenesisAlloc{addr: {Balance: balance}}}
	return g.MustCommit(db)
}

// DefaultGenesisBlock returns the Matrix main net genesis block.
func DefaultGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params.MainnetChainConfig,
		Nonce:      66,
		ExtraData:  hexutil.MustDecode("0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa"),
		GasLimit:   5000,
		Difficulty: big.NewInt(17179869184),
		Alloc:      decodePrealloc(mainnetAllocData),
	}
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
func DefaultTestnetGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params.TestnetChainConfig,
		Nonce:      66,
		ExtraData:  hexutil.MustDecode("0x3535353535353535353535353535353535353535353535353535353535353535"),
		GasLimit:   16777216,
		Difficulty: big.NewInt(1048576),
		Alloc:      decodePrealloc(testnetAllocData),
	}
}

// DefaultRinkebyGenesisBlock returns the Rinkeby network genesis block.
func DefaultRinkebyGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params.RinkebyChainConfig,
		Timestamp:  1492009146,
		ExtraData:  hexutil.MustDecode("0x52657370656374206d7920617574686f7269746168207e452e436172746d616e42eb768f2244c8811c63729a21a3569731535f067ffc57839b00206d1ad20c69a1981b489f772031b279182d99e65703f0076e4812653aab85fca0f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		GasLimit:   4700000,
		Difficulty: big.NewInt(1),
		Alloc:      decodePrealloc(rinkebyAllocData),
	}
}

// DeveloperGenesisBlock returns the 'gman --dev' genesis block. Note, this must
// be seeded with the
func DeveloperGenesisBlock(period uint64, faucet common.Address) *Genesis {
	// Override the default period to the user requested one
	config := *params.AllCliqueProtocolChanges
	//config.Clique.Period = period
	//todo 把clique设为空，默认启动ethash引擎挖矿
	config.Clique = nil

	// Assemble and return the genesis with the precompiles and faucet pre-funded
	return &Genesis{
		Config:     &config,
		ExtraData:  append(append(make([]byte, 32), faucet[:]...), make([]byte, 65)...),
		GasLimit:   6283185,
		Difficulty: big.NewInt(1000000),
		Alloc: map[common.Address]GenesisAccount{
			common.BytesToAddress([]byte{1}): {Balance: big.NewInt(1)}, // ECRecover
			common.BytesToAddress([]byte{2}): {Balance: big.NewInt(1)}, // SHA256
			common.BytesToAddress([]byte{3}): {Balance: big.NewInt(1)}, // RIPEMD
			common.BytesToAddress([]byte{4}): {Balance: big.NewInt(1)}, // Identity
			common.BytesToAddress([]byte{5}): {Balance: big.NewInt(1)}, // ModExp
			common.BytesToAddress([]byte{6}): {Balance: big.NewInt(1)}, // ECAdd
			common.BytesToAddress([]byte{7}): {Balance: big.NewInt(1)}, // ECScalarMul
			common.BytesToAddress([]byte{8}): {Balance: big.NewInt(1)}, // ECPairing
			faucet:                           {Balance: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))},
		},
	}
}

func decodePrealloc(data string) GenesisAlloc {
	var p []struct{ Addr, Balance *big.Int }
	if err := rlp.NewStream(strings.NewReader(data), 0).Decode(&p); err != nil {
		panic(err)
	}
	ga := make(GenesisAlloc, len(p))
	for _, account := range p {
		ga[common.BigToAddress(account.Addr)] = GenesisAccount{Balance: account.Balance}
	}
	return ga
}
