package blkmanage

import (
	"encoding/json"
	"math/big"
	"reflect"
	"time"

	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/MatrixAINetwork/go-matrix/matrixwork"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"

	"github.com/pkg/errors"

	"github.com/MatrixAINetwork/go-matrix/log"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

type ManBlkBasePlug struct {
	preBlockHash common.Hash
}

func NewBlkBasePlug() (*ManBlkBasePlug, error) {
	obj := new(ManBlkBasePlug)
	return obj, nil
}

func (p *ManBlkBasePlug) getVrfValue(support BlKSupport, parent *types.Block) ([]byte, []byte, []byte, error) {
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
	return support.SignHelper().SignVrf(vrfmsg, parent.Hash())
}

func (p *ManBlkBasePlug) setVrf(support BlKSupport, parent *types.Block, header *types.Header) error {
	account, vrfValue, vrfProof, err := p.getVrfValue(support, parent)
	if err != nil {
		log.Error(LogManBlk, "区块生成阶段 获取vrfValue失败 错误", err)
		return err
	}
	header.VrfValue = baseinterface.NewVrf().GetHeaderVrf(account, vrfValue, vrfProof)
	return nil
}

func (p *ManBlkBasePlug) setVersion(header *types.Header, parent *types.Block, version string) error {
	header.Version = []byte(version)
	header.VersionSignatures = manparams.GetVersionSignature(parent, header.Version)
	if nil == header.VersionSignatures {
		log.Error(LogManBlk, "获取版本号签名错误", "")
		return errors.New("获取版本号签名错误")
	}
	return nil
}

func (p *ManBlkBasePlug) setSignatures(header *types.Header) {
	header.Signatures = make([]common.Signature, 0)
}
func (bd *ManBlkBasePlug) setTopology(support BlKSupport, parentHash common.Hash, header *types.Header, interval *mc.BCIntervalInfo, num uint64) ([]*mc.HD_OnlineConsensusVoteResultMsg, error) {
	NetTopology, onlineConsensusResults := support.ReElection().GetNetTopology(num, parentHash, interval)
	if nil == NetTopology {
		NetTopology = &common.NetTopology{common.NetTopoTypeChange, nil}
	}
	if nil == onlineConsensusResults {
		onlineConsensusResults = make([]*mc.HD_OnlineConsensusVoteResultMsg, 0)
	}
	//log.Debug(LogManBlk, "获取拓扑结果 ", NetTopology, "在线共识信息", onlineConsensusResults, "高度", num)
	header.NetTopology = *NetTopology
	return onlineConsensusResults, nil
}

func (bd *ManBlkBasePlug) setTime(header *types.Header, tstamp int64) {
	header.Time = big.NewInt(tstamp)
}

func (bd *ManBlkBasePlug) setExtra(header *types.Header) {
	header.Extra = make([]byte, 0)
}

func (bd *ManBlkBasePlug) setGasLimit(header *types.Header, parent *types.Block) {
	header.GasLimit = core.CalcGasLimit(parent)
}

func (bd *ManBlkBasePlug) setNumber(header *types.Header, num uint64) {
	header.Number = new(big.Int).SetUint64(num)
}

func (bd *ManBlkBasePlug) setLeader(header *types.Header) {
	header.Leader = ca.GetDepositAddress()
}
func (bd *ManBlkBasePlug) setTimeStamp(parent *types.Block, header *types.Header, num uint64) {
	tstart := time.Now()
	log.Info(LogManBlk, "关键时间点", "区块头开始生成", "time", tstart, "块高", num)
	tstamp := tstart.Unix()
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}
	// this will ensure we're not going off too far in the future
	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		log.Info(LogManBlk, "等待时间同步", common.PrettyDuration(wait))
		time.Sleep(wait)
	}
	bd.setTime(header, tstamp)
}

func (bd *ManBlkBasePlug) getParentBlock(chain ChainReader, num uint64) (*types.Block, error) {
	if num == 1 { // 第一个块直接返回创世区块作为父区块
		return chain.Genesis(), nil
	}

	if (bd.preBlockHash == common.Hash{}) {
		return nil, errors.Errorf("未知父区块hash[%s]", bd.preBlockHash.TerminalString())
	}

	parent := chain.GetBlockByHash(bd.preBlockHash)
	if nil == parent {
		return nil, errors.Errorf("未知的父区块[%s]", bd.preBlockHash.TerminalString())
	}

	return parent, nil
}

func (bd *ManBlkBasePlug) setParentHash(chain ChainReader, header *types.Header, num uint64) (*types.Block, error) {
	parent, err := bd.getParentBlock(chain, num)
	if err != nil {
		return nil, err
	}

	header.ParentHash = parent.Hash()
	return parent, nil
}

func (bd *ManBlkBasePlug) setElect(support BlKSupport, stateDB *state.StateDBManage, header *types.Header) error {
	// 运行完状态树后，才能获取elect
	Elect := support.ReElection().GenElection(stateDB, header.ParentHash)
	if Elect == nil {
		return errors.New("生成elect信息错误")
	}
	//log.Debug(LogManBlk, "获取选举结果 ", Elect, "高度", header.Number.Uint64())
	header.Elect = Elect
	return nil
}

func (bd *ManBlkBasePlug) Prepare(version string, support BlKSupport, interval *mc.BCIntervalInfo, num uint64, args interface{}) (*types.Header, interface{}, error) {
	test, _ := args.([]interface{})
	for _, v := range test {
		switch v.(type) {

		case common.Hash:
			preBlockHash, ok := v.(common.Hash)
			if !ok {
				log.Error(LogManBlk, "反射失败,类型为", "")
				return nil, nil, errors.New("反射失败")
			}
			bd.preBlockHash = preBlockHash
		default:
			log.Error(LogManBlk, "unkown type", reflect.ValueOf(v).Type())
		}

	}

	originHeader := new(types.Header)
	parent, err := bd.setParentHash(support.BlockChain(), originHeader, num)
	if nil != err {
		log.ERROR(LogManBlk, "区块生成阶段", "获取父区块失败")
		return nil, nil, err
	}

	bd.setTimeStamp(parent, originHeader, num)
	bd.setLeader(originHeader)
	bd.setNumber(originHeader, num)
	bd.setGasLimit(originHeader, parent)
	bd.setExtra(originHeader)
	onlineConsensusResults, _ := bd.setTopology(support, parent.Hash(), originHeader, interval, num)
	bd.setSignatures(originHeader)
	err = bd.setVrf(support, parent, originHeader)
	if nil != err {
		return nil, nil, err
	}
	bd.setVersion(originHeader, parent, version)
	if nil != err {
		return nil, nil, err
	}
	if err := support.BlockChain().Engine(originHeader.Version).Prepare(support.BlockChain(), originHeader); err != nil {
		log.ERROR(LogManBlk, "Failed to prepare header for mining", err)
		return nil, nil, err
	}
	return originHeader, onlineConsensusResults, nil
}

func (bd *ManBlkBasePlug) ProcessState(support BlKSupport, header *types.Header, args interface{}) ([]*common.RetCallTxN, *state.StateDBManage, []types.CoinReceipts, []types.CoinSelfTransaction, []types.CoinSelfTransaction, interface{}, error) {
	work, err := matrixwork.NewWork(support.BlockChain().Config(), support.BlockChain(), nil, header)
	if err != nil {
		log.ERROR(LogManBlk, "区块验证请求生成,交易部分", "Work创建失败", "err", err)
		return nil, nil, nil, nil, nil, nil, err
	}
	if err = support.BlockChain().ProcessStateVersion(header.Version, work.State); err != nil {
		log.ERROR(LogManBlk, "状态树更新版本号失败", err, "高度", header.Number.Uint64())
		return nil, nil, nil, nil, nil, nil, err
	}
	upTimeMap, err := support.BlockChain().ProcessUpTime(work.State, header)
	if err != nil {
		log.ERROR(LogManBlk, "执行uptime错误", err, "高度", header.Number)
		return nil, nil, nil, nil, nil, nil, err
	}
	err = support.BlockChain().ProcessBlockGProduceSlash(work.State, header)
	if err != nil {
		log.ERROR(LogManBlk, "执行区块惩罚处理错误", err, "高度", header.Number)
		return nil, nil, nil, nil, nil, nil, err
	}
	txsCode, originalTxs, finalTxs := work.ProcessTransactions(support.EventMux(), support.TxPool(), upTimeMap)

	block := types.NewBlock(header, types.MakeCurencyBlock(types.GetCoinTX(finalTxs), work.Receipts, nil), nil)
	log.Debug(LogManBlk, "区块验证请求生成，交易部分,完成 tx hash", types.TxHashList(finalTxs))
	parent := support.BlockChain().GetBlockByHash(header.ParentHash)
	err = support.BlockChain().ProcessMatrixState(block, string(parent.Version()), work.State)
	if err != nil {
		log.Error(LogManBlk, "运行matrix状态树失败", err)
		return nil, nil, nil, nil, nil, nil, err
	}

	return txsCode, work.State, work.Receipts, types.GetCoinTX(originalTxs), types.GetCoinTX(finalTxs), nil, nil
}

func (bd *ManBlkBasePlug) Finalize(support BlKSupport, header *types.Header, state *state.StateDBManage, txs []types.CoinSelfTransaction, uncles []*types.Header, receipts []types.CoinReceipts, args interface{}) (*types.Block, interface{}, error) {
	err := bd.setElect(support, state, header)
	if err != nil {
		log.Error(LogManBlk, "设置选举信息失败", err)
		return nil, nil, err
	}

	block, err := support.BlockChain().Engine(header.Version).Finalize(support.BlockChain(), header, state, uncles, types.MakeCurencyBlock(txs, receipts, nil))
	if err != nil {
		log.Error(LogManBlk, "最终finalize错误", err)
		return nil, nil, err
	}
	return block, nil, nil
}

func (bd *ManBlkBasePlug) VerifyHeader(version string, support BlKSupport, header *types.Header, args interface{}) (interface{}, error) {
	if err := support.BlockChain().VerifyHeader(header); err != nil {
		log.ERROR(LogManBlk, "预验证头信息失败", err, "高度", header.Number.Uint64())
		return nil, err
	}

	// verify net topology info
	onlineConsensusResults := make([]*mc.HD_OnlineConsensusVoteResultMsg, 0)
	test, _ := args.([]interface{})
	for _, v := range test {
		switch v.(type) {

		case []*mc.HD_OnlineConsensusVoteResultMsg:
			data, ok := v.([]*mc.HD_OnlineConsensusVoteResultMsg)
			if !ok {
				log.Error(LogManBlk, "反射顶点配置失败", "")
				return nil, errors.New("反射失败")
			}
			onlineConsensusResults = data
		default:
			log.Warn(LogManBlk, "unkown type:", reflect.ValueOf(v).Type())
		}

	}
	if err := support.ReElection().VerifyNetTopology(header, onlineConsensusResults); err != nil {
		log.ERROR(LogManBlk, "验证拓扑信息失败", err, "高度", header.Number.Uint64())
		return nil, err
	}

	if err := support.BlockChain().DPOSEngine([]byte(version)).VerifyVersionSigns(support.BlockChain(), header); err != nil {
		log.ERROR(LogManBlk, "验证版本号签名失败", err, "高度", header.Number.Uint64())
		return nil, err
	}

	//verify vrf
	if err := support.ReElection().VerifyVrf(header); err != nil {
		log.Error(LogManBlk, "验证vrf失败", err, "高度", header.Number.Uint64())
		return nil, err
	}
	//log.INFO(LogManBlk, "验证vrf成功 高度", header.Number.Uint64())

	return nil, nil
}
func (bd *ManBlkBasePlug) VerifyTxsAndState(support BlKSupport, verifyHeader *types.Header, verifyTxs []types.CoinSelfTransaction, args interface{}) (*state.StateDBManage, []types.CoinSelfTransaction,
	[]types.CoinReceipts, interface{}, error) {
	//log.INFO(LogManBlk, "开始交易验证, 数量", len(verifyTxs), "高度", verifyHeader.Number.Uint64())

	//跑交易交易验证， Root TxHash ReceiptHash Bloom GasLimit GasUsed
	localHeader := types.CopyHeader(verifyHeader)
	localHeader.GasUsed = 0
	verifyHeaderHash := verifyHeader.HashNoSignsAndNonce()
	work, err := matrixwork.NewWork(support.BlockChain().Config(), support.BlockChain(), nil, localHeader)
	if err != nil {
		log.ERROR(LogManBlk, "交易验证，创建work失败!", err, "高度", verifyHeader.Number.Uint64())
		return nil, nil, nil, nil, err
	}
	if err = support.BlockChain().ProcessStateVersion(verifyHeader.Version, work.State); err != nil {
		log.ERROR(LogManBlk, "状态树更新版本号失败", err, "高度", verifyHeader.Number.Uint64())
		return nil, nil, nil, nil, err
	}
	uptimeMap, err := support.BlockChain().ProcessUpTime(work.State, localHeader)
	if err != nil {
		log.Error(LogManBlk, "uptime处理错误", err)
		return nil, nil, nil, nil, err
	}
	err = support.BlockChain().ProcessBlockGProduceSlash(work.State, localHeader)
	if err != nil {
		log.Error(LogManBlk, "区块生产惩罚处理错误", err)
		return nil, nil, nil, nil, err
	}
	err = work.ConsensusTransactions(support.EventMux(), verifyTxs, uptimeMap)
	if err != nil {
		log.ERROR(LogManBlk, "交易验证，共识执行交易出错!", err, "高度", verifyHeader.Number.Uint64())
		return nil, nil, nil, nil, err
	}
	finalTxs := work.GetTxs()

	localBlock := types.NewBlock(localHeader, types.MakeCurencyBlock(finalTxs, work.Receipts, nil), nil)
	// process matrix state
	parent := support.BlockChain().GetBlockByHash(verifyHeader.ParentHash)
	if parent == nil {
		log.Error(LogManBlk, "获取父区块失败", "is nil")
		return nil, nil, nil, nil, errors.New("父区块为nil")
	}
	err = support.BlockChain().ProcessMatrixState(localBlock, string(parent.Version()), work.State)
	if err != nil {
		log.ERROR(LogManBlk, "matrix状态验证,错误", "运行matrix状态出错", "err", err)
		return nil, nil, nil, nil, err
	}

	// 运行完matrix state后，生成root
	localBlock, err = support.BlockChain().Engine(verifyHeader.Version).Finalize(support.BlockChain(), localHeader, work.State, nil, types.MakeCurencyBlock(finalTxs, work.Receipts, nil))
	if err != nil {
		log.ERROR(LogManBlk, "matrix状态验证,错误", "Failed to finalize block for sealing", "err", err)
		return nil, nil, nil, nil, err
	}
	for _, curr := range localBlock.Header().Roots {
		for _, he := range verifyHeader.Roots {
			if curr.Cointyp == he.Cointyp {
				if !curr.TxHash.Equal(he.TxHash) {
					log.WARN(LogManBlk, "共识后的交易本地hash", curr.TxHash.String(), "共识后的交易远程hash", he.TxHash.String(), "coin type", curr.Cointyp)
				}
				if !curr.Root.Equal(he.Root) {
					log.WARN(LogManBlk, "finalize root", curr.Root.Hex(), "remote root", he.Root.Hex(), "coin type", curr.Cointyp)
				}
				break
			}
		}
	}

	// verify election info
	if err := support.ReElection().VerifyElection(verifyHeader, work.State); err != nil {
		log.ERROR(LogManBlk, "验证选举信息失败", err, "高度", verifyHeader.Number.Uint64())
		return nil, nil, nil, nil, err
	}

	//localBlock check
	localHeader = localBlock.Header()
	localHash := localHeader.HashNoSignsAndNonce()

	if localHash != verifyHeaderHash {
		log.ERROR(LogManBlk, "交易验证及状态，错误", "block hash不匹配",
			"local hash", localHash.TerminalString(), "remote hash", verifyHeaderHash.TerminalString(),
			"local root", localHeader.Roots, "remote root", verifyHeader.Roots,
			//"local txHash", localHeader.TxHash.TerminalString(), "remote txHash", verifyHeader.TxHash.TerminalString(),
			//"local ReceiptHash", localHeader.ReceiptHash.TerminalString(), "remote ReceiptHash", verifyHeader.ReceiptHash.TerminalString(),
			//"local Bloom", localHeader.Bloom.Big(), "remote Bloom", verifyHeader.Bloom.Big(),
			"local GasLimit", localHeader.GasLimit, "remote GasLimit", verifyHeader.GasLimit,
			"local GasUsed", localHeader.GasUsed, "remote GasUsed", verifyHeader.GasUsed)
		return nil, nil, nil, nil, errors.New("hash 不一致")
	}
	return work.State, finalTxs, work.Receipts, nil, nil
}
