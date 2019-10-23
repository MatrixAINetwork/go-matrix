// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkmanage

import (
	"errors"

	"github.com/MatrixAINetwork/go-matrix/accounts/signhelper"
	"github.com/MatrixAINetwork/go-matrix/reelection"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"
)

type MANBLK interface {
	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(types string, version string, num uint64, interval *mc.BCIntervalInfo, args ...interface{}) (*types.Header, interface{}, error)
	ProcessState(types string, version string, header *types.Header, args ...interface{}) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, interface{}, error)
	Finalize(types string, version string, header *types.Header, state *state.StateDB, txs []types.SelfTransaction, uncles []*types.Header, receipts []*types.Receipt, args interface{}) (*types.Block, interface{}, error)
	VerifyHeader(types string, version string, header *types.Header, args ...interface{}) (interface{}, error)
	VerifyTxsAndState(types string, version string, header *types.Header, Txs types.SelfTransactions, args ...interface{}) (*state.StateDB, types.SelfTransactions, []*types.Receipt, interface{}, error)
}

type ChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeader retrieves a block header from the database by hash and number.
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	// GetBlock retrieves a block from the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block

	Genesis() *types.Block

	GetBlockByHash(hash common.Hash) *types.Block

	GetCurrentHash() common.Hash
	GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error)
	GetVersionSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBlockSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBroadcastInterval() (*mc.BCIntervalInfo, error)

	ProcessUpTime(state *state.StateDBManage, header *types.Header) (map[common.Address]uint64, error)
	StateAt(root []common.CoinRoot) (*state.StateDBManage, error)
	Engine(version []byte) consensus.Engine
	DPOSEngine(version []byte) consensus.DPOSEngine
	Processor(version []byte) core.Processor
}

type MANBLKPlUGS interface {
	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(version string, support BlKSupport, interval *mc.BCIntervalInfo, num uint64, args interface{}) (*types.Header, interface{}, error)
	ProcessState(support BlKSupport, header *types.Header, args interface{}) ([]*common.RetCallTxN, *state.StateDBManage, []types.CoinReceipts, []types.CoinSelfTransaction, []types.CoinSelfTransaction, interface{}, error)
	Finalize(support BlKSupport, header *types.Header, state *state.StateDBManage, txs []types.CoinSelfTransaction, uncles []*types.Header, receipts []types.CoinReceipts, args interface{}) (*types.Block, interface{}, error)
	VerifyHeader(version string, support BlKSupport, header *types.Header, args interface{}) (interface{}, error)
	VerifyTxsAndState(support BlKSupport, header *types.Header, Txs []types.CoinSelfTransaction, args interface{}) (*state.StateDBManage, []types.CoinSelfTransaction, []types.CoinReceipts, interface{}, error)
}

type TopNodeService interface {
	GetConsensusOnlineResults() []*mc.HD_OnlineConsensusVoteResultMsg
}

type Reelection interface {
	VerifyNetTopology(header *types.Header, onlineConsensusResults []*mc.HD_OnlineConsensusVoteResultMsg) error
	VerifyElection(header *types.Header, state *state.StateDB) error
	GetNetTopology(num uint64, parentHash common.Hash, bcInterval *mc.BCIntervalInfo) (*common.NetTopology, []*mc.HD_OnlineConsensusVoteResultMsg)
	GenElection(state *state.StateDB, preBlockHash common.Hash) []common.Elect
	VerifyVrf(header *types.Header) error
}
type SignHelper interface {
	SignVrf(msg []byte, blkHash common.Hash) ([]byte, []byte, []byte, error)
}
type txPool interface {
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[string]map[common.Address]types.SelfTransactions, error)
	GetAllSpecialTxs() (reqVal map[common.Address][]types.SelfTransaction)
}

type Mux interface {
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	EventMux() *event.TypeMux
}

type BlKSupport interface {
	BlockChain() *core.BlockChain
	TxPool() *core.TxPoolManager
	SignHelper() *signhelper.SignHelper
	EventMux() *event.TypeMux
	ReElection() *reelection.ReElection
}
type VrfMsg struct {
	VrfValue []byte
	VrfProof []byte
	Hash     common.Hash
}

var (
	LogManBlk    = "区块生成验证引擎"
	CommonBlk    = "common"
	BroadcastBlk = "broadcast"
)

type ManBlkManage struct {
	support        BlKSupport
	mapManBlkPlugs map[string]MANBLKPlUGS
}

func New(support BlKSupport) (*ManBlkManage, error) {
	obj := new(ManBlkManage)
	obj.support = support

	obj.mapManBlkPlugs = make(map[string]MANBLKPlUGS)
	manCommonplug, err := NewBlkBasePlug()
	if err != nil {
		return nil, err
	}

	aiMinePlug, err := NewBlkV2Plug()
	if err != nil {
		return nil, err
	}

	obj.RegisterManBLkPlugs(CommonBlk, manversion.VersionAlpha, manCommonplug)

	manBcplug, err := NewBCBlkPlug()

	obj.RegisterManBLkPlugs(BroadcastBlk, manversion.VersionAlpha, manBcplug)

	obj.RegisterManBLkPlugs(CommonBlk, manversion.VersionBeta, manCommonplug)

	obj.RegisterManBLkPlugs(BroadcastBlk, manversion.VersionBeta, manBcplug)

	obj.RegisterManBLkPlugs(CommonBlk, manversion.VersionGamma, manCommonplug)

	obj.RegisterManBLkPlugs(BroadcastBlk, manversion.VersionGamma, manBcplug)

	obj.RegisterManBLkPlugs(CommonBlk, manversion.VersionDelta, manCommonplug)

	obj.RegisterManBLkPlugs(BroadcastBlk, manversion.VersionDelta, manBcplug)

	obj.RegisterManBLkPlugs(CommonBlk, manversion.VersionAIMine, aiMinePlug)

	obj.RegisterManBLkPlugs(BroadcastBlk, manversion.VersionAIMine, manBcplug)

	return obj, nil
}

func (bd *ManBlkManage) RegisterManBLkPlugs(types string, version string, plug MANBLKPlUGS) {
	bd.mapManBlkPlugs[types+version] = plug
}

func (bd *ManBlkManage) ProduceBlockVersion(num uint64, preVersion string) string {

	switch num {
	case manversion.VersionNumGamma:
		return manversion.VersionGamma
	case manversion.VersionNumDelta:
		return manversion.VersionDelta
	case manversion.VersionNumAIMine:
		return manversion.VersionAIMine
	default:
		return preVersion
	}
}

func (bd *ManBlkManage) VerifyBlockVersion(num uint64, curVersion string, preVersion string) error {

	switch num {
	case manversion.VersionNumGamma:
		if manversion.VersionCmp(curVersion, manversion.VersionGamma) != 0 {
			return errors.New("版本号异常")
		} else {
			return nil
		}
	case manversion.VersionNumDelta:
		if manversion.VersionCmp(curVersion, manversion.VersionDelta) != 0 {
			return errors.New("版本号异常")
		} else {
			return nil
		}
	case manversion.VersionNumAIMine:
		if manversion.VersionCmp(curVersion, manversion.VersionAIMine) != 0 {
			return errors.New("版本号异常")
		} else {
			return nil
		}
	default:
		if curVersion != preVersion {
			return errors.New("版本号异常,不等于父区块版本号")
		} else {
			return nil
		}

	}
}

func (bd *ManBlkManage) Prepare(types string, version string, num uint64, interval *mc.BCIntervalInfo, args ...interface{}) (*types.Header, interface{}, error) {
	plug, ok := bd.mapManBlkPlugs[types+version]
	if !ok {
		log.Error(LogManBlk, "获取插件失败", "")
		return nil, nil, errors.New("获取插件失败")
	}
	return plug.Prepare(version, bd.support, interval, num, args)
}

func (bd *ManBlkManage) ProcessState(types string, version string, header *types.Header, args ...interface{}) ([]*common.RetCallTxN, *state.StateDBManage, []types.CoinReceipts, []types.CoinSelfTransaction, []types.CoinSelfTransaction, interface{}, error) {
	plug, ok := bd.mapManBlkPlugs[types+version]
	if !ok {
		log.Error(LogManBlk, "获取插件失败", "")
		return nil, nil, nil, nil, nil, nil, errors.New("获取插件失败")
	}
	return plug.ProcessState(bd.support, header, args)
}

func (bd *ManBlkManage) Finalize(types string, version string, header *types.Header, state *state.StateDBManage, txs []types.CoinSelfTransaction, uncles []*types.Header, receipts []types.CoinReceipts, args ...interface{}) (*types.Block, interface{}, error) {
	plug, ok := bd.mapManBlkPlugs[types+version]
	if !ok {
		log.Error(LogManBlk, "获取插件失败", "")
		return nil, nil, errors.New("获取插件失败")
	}
	return plug.Finalize(bd.support, header, state, txs, uncles, receipts, args)
}

func (bd *ManBlkManage) VerifyHeader(types string, version string, header *types.Header, args ...interface{}) (interface{}, error) {
	plug, ok := bd.mapManBlkPlugs[types+version]
	if !ok {
		log.Error(LogManBlk, "获取插件失败", "")
		return nil, errors.New("获取插件失败")
	}
	return plug.VerifyHeader(version, bd.support, header, args)
}

func (bd *ManBlkManage) VerifyTxsAndState(types string, version string, header *types.Header, Txs []types.CoinSelfTransaction, args ...interface{}) (*state.StateDBManage, []types.CoinSelfTransaction, []types.CoinReceipts, interface{}, error) {
	plug, ok := bd.mapManBlkPlugs[types+version]
	if !ok {
		log.Error(LogManBlk, "获取插件失败", "")
		return nil, nil, nil, nil, errors.New("获取插件失败")
	}
	return plug.VerifyTxsAndState(bd.support, header, Txs, args)
}
