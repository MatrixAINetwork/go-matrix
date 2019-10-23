// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkmanage

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/MatrixAINetwork/go-matrix/ca"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/matrixwork"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"
)

type ManBCBlkPlug struct {
	baseInterface *ManBlkBasePlug
	preBlockHash  common.Hash
}

func NewBCBlkPlug() (*ManBCBlkPlug, error) {
	obj := new(ManBCBlkPlug)
	obj.baseInterface, _ = NewBlkBasePlug()
	return obj, nil
}

func (bd *ManBCBlkPlug) Prepare(version string, support BlKSupport, interval *mc.BCIntervalInfo, num uint64, args interface{}) (*types.Header, interface{}, error) {
	test, _ := args.([]interface{})
	for _, v := range test {
		switch v.(type) {

		case common.Hash:
			preBlockHash, ok := v.(common.Hash)
			if !ok {
				log.Error(LogManBlk, "反射失败,类型为", "")
				return nil, nil, errors.New("反射失败")
			}
			bd.baseInterface.preBlockHash = preBlockHash
		default:
			log.Warn(LogManBlk, "unkown type:", reflect.ValueOf(v).Type())
		}

	}

	originHeader := new(types.Header)
	parent, err := bd.baseInterface.setParentHash(support.BlockChain(), originHeader, num)
	if nil != err {
		log.Error(LogManBlk, "区块生成阶段", "获取父区块失败")
		return nil, nil, err
	}

	bd.setBCTimeStamp(parent, originHeader, num)
	bd.baseInterface.setLeader(originHeader)
	bd.baseInterface.setNumber(originHeader, num)
	bd.baseInterface.setGasLimit(originHeader, parent)
	bd.baseInterface.setExtra(originHeader)
	onlineConsensusResults, _ := bd.baseInterface.setTopology(support, version, parent.Hash(), originHeader, interval, num)
	bd.baseInterface.setSignatures(originHeader)
	err = bd.setBCVrf(support, parent, originHeader)
	if nil != err {
		return nil, nil, err
	}
	bd.baseInterface.setVersion(originHeader, parent, version)
	if nil != err {
		return nil, nil, err
	}
	bd.baseInterface.initBasePowers(originHeader)
	if manversion.VersionCmp(string(originHeader.Version), manversion.VersionAIMine) >= 0 {
		bd.setBCMiner(originHeader)
	}
	if err := support.BlockChain().Engine(originHeader.Version).Prepare(support.BlockChain(), originHeader); err != nil {
		log.Error(LogManBlk, "Failed to prepare header for mining", err)
		return nil, nil, err
	}
	return originHeader, onlineConsensusResults, nil
}

func (p *ManBCBlkPlug) getBCVrfValue(support BlKSupport, parent *types.Block) ([]byte, []byte, []byte, error) {
	_, preVrfValue, preVrfProof := baseinterface.NewVrf().GetVrfInfoFromHeader(parent.Header().VrfValue)
	parentMsg := VrfMsg{
		VrfProof: preVrfProof,
		VrfValue: preVrfValue,
		Hash:     parent.Hash(),
	}
	vrfmsg, err := json.Marshal(parentMsg)
	if err != nil {
		log.Error(LogManBlk, "生成vrfmsg出错", err, "parentMsg", parentMsg)
		return []byte{}, []byte{}, []byte{}, errors.New("生成vrfmsg出错")
	}
	return support.SignHelper().SignVrfByAccount(vrfmsg, ca.GetDepositAddress())
}

func (p *ManBCBlkPlug) setBCVrf(support BlKSupport, parent *types.Block, header *types.Header) error {
	account, vrfValue, vrfProof, err := p.getBCVrfValue(support, parent)
	if err != nil {
		log.Error(LogManBlk, "广播区块生成阶段 获取vrfValue失败 错误", err)
		return err
	}
	header.VrfValue = baseinterface.NewVrf().GetHeaderVrf(account, vrfValue, vrfProof)
	return nil
}

func (p *ManBCBlkPlug) setBCTimeStamp(parent *types.Block, header *types.Header, num uint64) {
	nowTime := time.Now()
	// 广播区块时间戳默认为父区块+1s， 保证所有广播节点出块的时间戳一致
	tsTamp := parent.Time().Int64() + 1
	log.Info(LogManBlk, "关键时间点", "广播区块头开始生成", "cur time", nowTime, "header time", tsTamp, "块高", num)
	// this will ensure we're not going off too far in the future
	if now := time.Now().Unix(); tsTamp > now+1 {
		wait := time.Duration(tsTamp-now) * time.Second
		log.Info(LogManBlk, "等待时间同步", common.PrettyDuration(wait))
		time.Sleep(wait)
	}
	p.baseInterface.setTime(header, tsTamp)
}

func (p *ManBCBlkPlug) setBCMiner(header *types.Header) {
	address := ca.GetDepositAddress()
	header.Coinbase = address
	header.AICoinbase = address
}

func (bd *ManBCBlkPlug) ProcessState(support BlKSupport, header *types.Header, args interface{}) ([]*common.RetCallTxN, *state.StateDBManage, []types.CoinReceipts, []types.CoinSelfTransaction, []types.CoinSelfTransaction, interface{}, error) {

	work, err := matrixwork.NewWork(support.BlockChain().Config(), support.BlockChain(), nil, header)
	if err != nil {
		log.Error(LogManBlk, "NewWork!", err, "高度", header.Number.Uint64())
		return nil, nil, nil, nil, nil, nil, err
	}

	if err = support.BlockChain().ProcessStateVersion(header.Version, work.State); err != nil {
		log.Error(LogManBlk, "状态树更新版本号失败", err, "高度", header.Number.Uint64())
		return nil, nil, nil, nil, nil, nil, err
	}

	if err = support.BlockChain().ProcessStateVersionSwitch(header.Number.Uint64(), header.Time.Uint64(), work.State); err != nil {
		log.Error(LogManBlk, "状态树版本号切换更新状态树", err, "高度", header.Number.Uint64())
		return nil, nil, nil, nil, nil, nil, err
	}

	err = support.BlockChain().BasePowerGProduceSlash(string(header.Version), work.State, header)
	if err != nil {
		log.Error(LogManBlk, "执行算力检测处理错误", err, "高度", header.Number)
		return nil, nil, nil, nil, nil, nil, err
	}

	mapTxs := support.TxPool().GetAllSpecialTxs()
	Txs := make([]types.SelfTransaction, 0)
	for _, txs := range mapTxs {
		for _, tx := range txs {
			log.Trace(LogManBlk, "交易数据", tx)
		}
		Txs = append(Txs, txs...)
	}
	work.ProcessBroadcastTransactions(support.EventMux(), types.GetCoinTX(Txs))
	log.Info(LogManBlk, "关键时间点", "开始执行MatrixState", "time", time.Now(), "块高", header.Number.Uint64())

	block := types.NewBlock(header, types.MakeCurencyBlock(work.GetTxs(), work.Receipts, nil), nil)
	parent := support.BlockChain().GetBlockByHash(header.ParentHash)
	if parent == nil {
		log.Error(LogManBlk, "获取父区块失败", "is nil")
		return nil, nil, nil, nil, nil, nil, errors.New("父区块为nil")
	}
	err = support.BlockChain().ProcessMatrixState(block, string(parent.Version()), work.State)
	if err != nil {
		log.Error(LogManBlk, "运行matrix状态树失败", err)
		return nil, nil, nil, nil, nil, nil, err
	}
	return nil, work.State, work.Receipts, types.GetCoinTX(Txs), work.GetTxs(), nil, nil
}

func (bd *ManBCBlkPlug) Finalize(support BlKSupport, header *types.Header, state *state.StateDBManage, txs []types.CoinSelfTransaction, uncles []*types.Header, receipts []types.CoinReceipts, args interface{}) (*types.Block, interface{}, error) {
	block, _, err := bd.baseInterface.Finalize(support, header, state, txs, uncles, receipts, nil)
	if err != nil {
		log.Error(LogManBlk, "最终finalize错误", err)
		return nil, nil, err
	}
	return block, nil, nil
}

func (bd *ManBCBlkPlug) VerifyHeader(version string, support BlKSupport, header *types.Header, args interface{}) (interface{}, error) {
	if err := support.BlockChain().Engine([]byte(version)).VerifyHeader(support.BlockChain(), header, true, true); err != nil {
		log.Error(LogManBlk, "预验证头信息失败", err, "高度", header.Number.Uint64())
		return nil, err
	}

	if err := support.BlockChain().DPOSEngine(header.Version).VerifyBlock(support.BlockChain(), header); err != nil {
		log.Warn(LogManBlk, "验证广播挖矿结果", "结果异常", "err", err)
		return nil, err
	}
	onlineConsensusResults := make([]*mc.HD_OnlineConsensusVoteResultMsg, 0)

	if err := support.ReElection().VerifyNetTopology(version, header, onlineConsensusResults); err != nil {
		log.Error(LogManBlk, "验证拓扑信息失败", err, "高度", header.Number.Uint64())
		return nil, err
	}

	//verify vrf
	if err := support.ReElection().VerifyVrf(header); err != nil {
		log.Error(LogManBlk, "验证vrf失败", err, "高度", header.Number.Uint64())
		return nil, err
	}
	//log.Info(LogManBlk, "验证vrf成功 高度", header.Number.Uint64())

	return nil, nil
}

func (bd *ManBCBlkPlug) VerifyTxsAndState(support BlKSupport, verifyHeader *types.Header, verifyTxs []types.CoinSelfTransaction, args interface{}) (*state.StateDBManage, []types.CoinSelfTransaction,
	[]types.CoinReceipts, interface{}, error) {
	parent := support.BlockChain().GetBlockByHash(verifyHeader.ParentHash)
	if parent == nil {
		log.Warn(LogManBlk, "广播挖矿结果验证", "获取父区块错误!")
		return nil, nil, nil, nil, errors.New("广播挖矿结果验证,获取父区块错误!")
	}

	work, err := matrixwork.NewWork(support.BlockChain().Config(), support.BlockChain(), nil, verifyHeader)
	if err != nil {
		log.Warn(LogManBlk, "广播挖矿结果验证, 创建worker错误", err)
		return nil, nil, nil, nil, err
	}
	if err = support.BlockChain().ProcessStateVersion(verifyHeader.Version, work.State); err != nil {
		log.Error(LogManBlk, "状态树更新版本号失败", err, "高度", verifyHeader.Number.Uint64())
		return nil, nil, nil, nil, err
	}
	if err = support.BlockChain().ProcessStateVersionSwitch(verifyHeader.Number.Uint64(), verifyHeader.Time.Uint64(), work.State); err != nil {
		log.Error(LogManBlk, "状态树版本号切换更新状态树", err, "高度", verifyHeader.Number.Uint64())
		return nil, nil, nil, nil, err
	}
	err = support.BlockChain().BasePowerGProduceSlash(string(verifyHeader.Version), work.State, verifyHeader)
	if err != nil {
		log.Error(LogManBlk, "执行算力检测处理错误", err, "高度", verifyHeader.Number)
		return nil, nil, nil, nil, err
	}
	//执行交易
	work.ProcessBroadcastTransactions(support.EventMux(), verifyTxs)

	retTxs := work.GetTxs()
	currblock := types.MakeCurencyBlock(retTxs, work.Receipts, nil)
	// 运行matrix状态树
	block := types.NewBlock(verifyHeader, currblock, nil)
	if err := support.BlockChain().ProcessMatrixState(block, string(parent.Version()), work.State); err != nil {
		log.Error(LogManBlk, "广播挖矿结果验证, matrix 状态树运行错误", err)
		return nil, nil, nil, nil, err
	}

	err = support.BlockChain().UpdateCurrencyHeaderState(work.State, string(verifyHeader.Version), verifyHeader.Roots[1:], verifyHeader.Sharding[1:])
	if err != nil {
		log.Error(LogManBlk, "设置选举信息失败", err)
		return nil, nil, nil, nil, err
	}
	localBlock, err := support.BlockChain().Engine(verifyHeader.Version).Finalize(support.BlockChain(), block.Header(), work.State, nil, currblock)
	if err != nil {
		log.Error(LogManBlk, "Failed to finalize block for sealing", err)
		return nil, nil, nil, nil, err
	}
	for _, root := range localBlock.Root() {
		for _, headRoot := range verifyHeader.Roots {
			if root.Cointyp == headRoot.Cointyp {
				if root.Root != headRoot.Root {
					log.Error(LogManBlk, "广播挖矿结果验证", "root验证错误, 不匹配", "localRoot", root.Root.String(), "remote root", headRoot.Root.String(), "coin type ", root.Cointyp)
					return nil, nil, nil, nil, errors.New("root不一致")
				}
			}
		}
	}

	return work.State, retTxs, work.Receipts, nil, nil
}
