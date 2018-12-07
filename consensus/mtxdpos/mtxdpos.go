// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package mtxdpos

import (
	"math"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/pkg/errors"
)

const (
	DPOSTargetSignCountRatio   = 0.66667
	DPOSTargetStockRatio       = 0 // 暂时关闭股权的要求
	DPOSMinStockCount          = 3
	DPOSFullSignThreshold      = 7
	SuperNodeFullSignThreshold = 3
)

var (
	errInputHeaderErr = errors.New("input param header err")

	errSignHashLenErr = errors.New("hash is required to be exactly 32 bytes")

	errStockCountErr = errors.New("stock count is less then min count")

	errSignCountErr = errors.New("sign count is not enough")

	errSignStockErr = errors.New("sign stock is not enough")

	errDisagreeCountErr = errors.New("consensus failed, cause to disagree count is too much")

	errDisagreeStockErr = errors.New("consensus failed, cause to disagree stock is too much")

	errBroadcastSignCount = errors.New("broadcast block's sign count err, not one")

	errBroadcastVerifySign = errors.New("broadcast block's sign is not from broadcast node")

	errBroadcastVerifySignFalse = errors.New("broadcast block's sign is false")
)

type dposTarget struct {
	totalCount       int
	totalStock       uint64
	targetCount      int
	targetStock      uint64
	maxDisagreeCount int
	maxDisagreeStock uint64
}

type MtxDPOS struct {
}

func NewMtxDPOS() *MtxDPOS {
	return &MtxDPOS{}
}
func (md *MtxDPOS) VerifyVersion(reader consensus.ValidatorReader, header *types.Header) error {
	targetCount := md.calcSuperNodeTarget(len(manparams.SuperVersionNodes))

	if len(header.Version) < targetCount {
		log.ERROR("共识引擎", "版本号签名数量不足 size", len(header.Version), "target", targetCount)
		return errSignCountErr
	}

	verifiedVersion := md.verifyHashWithSuperNodes(common.BytesToHash([]byte(header.Version)), header.VersionSignatures, manparams.SuperVersionNodes)
	log.INFO("共识引擎", "版本", string(header.Version), "签名", header.VersionSignatures[0].Bytes(), "版本节点", manparams.SuperVersionNodes[0].Address.String())
	if len(verifiedVersion) < targetCount {
		log.ERROR("共识引擎", "验证版本,验证后的签名数量不足 size", len(verifiedVersion), "target", targetCount)
		return errSignCountErr
	}
	return nil
}

func (md *MtxDPOS) calcSuperNodeTarget(totalCount int) int {
	targetCount := 0
	if totalCount <= SuperNodeFullSignThreshold {
		targetCount = totalCount
	} else {
		targetCount = int(math.Ceil(float64(totalCount) * DPOSTargetSignCountRatio))
	}
	return targetCount
}

func (md *MtxDPOS) checkSuperBlock(header *types.Header) error {
	targetCount := md.calcSuperNodeTarget(len(manparams.SuperRollbackNodes))
	if len(header.Signatures) < targetCount {
		log.ERROR("共识引擎", "版本号签名数量不足 size", len(header.Version), "target", targetCount)
		return errSignCountErr
	}
	verifiedSigh := md.verifyHashWithSuperNodes(header.HashNoSigns(), header.Signatures, manparams.SuperRollbackNodes)
	if len(verifiedSigh) < targetCount {
		log.ERROR("共识引擎", "验证版本,验证后的签名数量不足 size", len(verifiedSigh), "target", targetCount, "hash", header.HashNoSigns().TerminalString())
		return errSignCountErr
	}
	return nil
}

func (md *MtxDPOS) verifyHashWithSuperNodes(hash common.Hash, signatures []common.Signature, superNodes []manparams.NodeInfo) map[common.Address]byte {
	verifiedSigh := make(map[common.Address]byte, 0)
	for _, sigh := range signatures {
		account, _, err := crypto.VerifySignWithValidate(hash.Bytes(), sigh.Bytes())
		if nil != err {
			log.ERROR("共识引擎", "验证版本错误", err)
			continue
		}
		findFlag := 0
		for _, superAccount := range superNodes {
			if account == superAccount.Address {
				findFlag = 1
				break
			}
		}
		if 0 == findFlag {
			log.WARN("共识引擎", "验证版本 账户未找到 node", account.Hex(), "签名：", sigh)
			continue
		}
		if _, ok := verifiedSigh[account]; !ok {
			verifiedSigh[account] = 0
		}
	}
	return verifiedSigh
}
func (md *MtxDPOS) VerifyBlock(reader consensus.ValidatorReader, header *types.Header) error {
	if nil == header {
		return errors.New("header is nil")
	}
	if err := md.VerifyVersion(reader, header); err != nil {
		log.INFO("MtxDPOS", "验证区块阶段 ", "验证版本", "版本号不正确 err", "err")
		return err
	}

	if header.IsSuperHeader() {
		return md.checkSuperBlock(header)
	}

	if common.IsBroadcastNumber(header.Number.Uint64()) {
		return md.verifyBroadcastBlock(header)
	}

	stocks, err := md.getValidatorStocks(reader, header.ParentHash)
	if err != nil {
		return err
	}

	hash := header.HashNoSignsAndNonce()
	log.INFO("共识引擎", "VerifyBlock, 签名总数", len(header.Signatures), "hash", hash, "txhash:", header.TxHash.TerminalString())

	_, err = md.VerifyHashWithStocks(reader, hash, header.Signatures, stocks)
	return err
}

func (md *MtxDPOS) VerifyBlocks(reader consensus.ValidatorReader, headers []*types.Header) error {
	if len(headers) <= 0 {
		return errInputHeaderErr
	}

	var (
		preGraph *mc.TopologyGraph = nil
		err      error
	)
	for _, header := range headers {
		if nil == preGraph {
			preGraph, err = md.getValidatorGraph(reader, header.ParentHash)
			if err != nil {
				return err
			}
		}

		if err := md.VerifyVersion(reader, header); err != nil {
			log.INFO("MtxDPOS", "VerifyBlocks ", "验证版本", "版本号不正确 err", err)
			return err
		}

		hash := header.HashNoSignsAndNonce()
		number := header.Number.Uint64()

		if header.IsSuperHeader() {
			err = md.checkSuperBlock(header)
			if err != nil {
				return errors.Errorf("header(hash:%s, number:%d) verify Broadcast Block err: %v", hash.Hex(), number, err)
			}
		} else {
			if common.IsBroadcastNumber(number) {
				err = md.verifyBroadcastBlock(header)
				if err != nil {
					return errors.Errorf("header(hash:%s, number:%d) verify Broadcast Block err: %v", hash.Hex(), number, err)
				}
			} else {
				stocks := md.graph2ValidatorStocks(preGraph)
				_, err = md.VerifyHashWithStocks(reader, hash, header.Signatures, stocks)
				if err != nil {
					return errors.Errorf("header(hash:%s, number:%d) dpos verify err: %v", hash.Hex(), number, err)
				}
			}
		}

		preGraph, err = preGraph.Transfer2NextGraph(header.Number.Uint64(), &header.NetTopology, nil)
		if err != nil {
			return errors.Errorf("header(hash:%s, number:%d) gen next topology err: %v", hash.Hex(), number, err)
		}
	}

	return nil
}

func (md *MtxDPOS) VerifyHash(reader consensus.ValidatorReader, signHash common.Hash, signs []common.Signature) ([]common.Signature, error) {
	return md.VerifyHashWithBlock(reader, signHash, signs, reader.GetCurrentHash())
}

func (md *MtxDPOS) VerifyHashWithBlock(reader consensus.ValidatorReader, signHash common.Hash, signs []common.Signature, blockHash common.Hash) ([]common.Signature, error) {
	stocks, err := md.getValidatorStocks(reader, blockHash)
	if err != nil {
		return nil, err
	}

	return md.VerifyHashWithStocks(reader, signHash, signs, stocks)
}

func (md *MtxDPOS) VerifyHashWithStocks(reader consensus.ValidatorReader, signHash common.Hash, signs []common.Signature, stocks map[common.Address]uint16) ([]common.Signature, error) {
	if len(signHash) != 32 {
		return nil, errSignHashLenErr
	}

	target, err := md.calculateDPOSTarget(stocks)
	if err != nil {
		return nil, err
	}

	// check whether sign count is enough
	if len(signs) < target.targetCount {
		log.ERROR("共识引擎", "签名数量不足 size", len(signs), "target", target.targetCount)
		return nil, errSignCountErr
	}

	verifiedSigns := md.verifySigns(signHash, signs, stocks)
	if len(verifiedSigns) < target.targetCount {
		log.ERROR("共识引擎", "验证后的签名数量不足 size", len(verifiedSigns), "target", target.targetCount)
		return nil, errSignCountErr
	}

	return md.verifyDPOS(verifiedSigns, target)
}

func (md *MtxDPOS) VerifyHashWithVerifiedSigns(reader consensus.ValidatorReader, signs []*common.VerifiedSign) ([]common.Signature, error) {
	return md.VerifyHashWithVerifiedSignsAndBlock(reader, signs, reader.GetCurrentHash())
}

func (md *MtxDPOS) VerifyHashWithVerifiedSignsAndBlock(reader consensus.ValidatorReader, signs []*common.VerifiedSign, blockHash common.Hash) ([]common.Signature, error) {
	stocks, err := md.getValidatorStocks(reader, blockHash)
	if err != nil {
		return nil, err
	}

	target, err := md.calculateDPOSTarget(stocks)
	if err != nil {
		return nil, err
	}

	// check whether sign count is enough
	if len(signs) < target.targetCount {
		return nil, errSignCountErr
	}

	verifiedSigns := md.parseVerifiedSigns(signs, stocks)
	if len(verifiedSigns) < target.targetCount {
		return nil, errSignCountErr
	}

	return md.verifyDPOS(verifiedSigns, target)
}

func (md *MtxDPOS) calculateDPOSTarget(stocks map[common.Address]uint16) (*dposTarget, error) {
	totalCount := len(stocks)
	//check total count
	if totalCount < DPOSMinStockCount {
		return nil, errStockCountErr
	}

	//calculate total stock
	var totalStock uint64 = 0
	for _, stock := range stocks {
		totalStock += uint64(stock)
	}

	//calculate target
	target := &dposTarget{totalCount: totalCount, totalStock: totalStock}
	if totalCount <= DPOSFullSignThreshold {
		target.targetCount = totalCount
		target.targetStock = totalStock
	} else {
		target.targetCount = int(math.Ceil(float64(totalCount) * DPOSTargetSignCountRatio))
		target.targetStock = uint64(math.Ceil(float64(totalStock) * DPOSTargetStockRatio))
	}
	target.maxDisagreeCount = target.totalCount - target.targetCount
	target.maxDisagreeStock = target.totalStock - target.targetStock
	return target, nil
}

func (md *MtxDPOS) parseVerifiedSigns(verifiedSigns []*common.VerifiedSign, stocks map[common.Address]uint16) map[common.Address]*common.VerifiedSign {
	verifiedSign := make(map[common.Address]*common.VerifiedSign)
	signCount := len(verifiedSigns)
	for i := 0; i < signCount; i++ {
		sign := verifiedSigns[i]
		stock, findStock := stocks[sign.Account]
		if findStock == false {
			// can't find in stock, discard
			continue
		}

		if existData, exist := verifiedSign[sign.Account]; exist {
			//already exist, replace "disagree" sign with "agree" sign
			if existData.Validate == false && sign.Validate == true {
				existData.Sign = sign.Sign
				existData.Account = sign.Account
				existData.Validate = sign.Validate
				existData.Stock = stock
			}
		} else {
			verifiedSign[sign.Account] = &common.VerifiedSign{Sign: sign.Sign, Account: sign.Account, Validate: sign.Validate, Stock: stock}
		}
	}

	return verifiedSign
}

func (md *MtxDPOS) verifySigns(signHash common.Hash, signs []common.Signature, stocks map[common.Address]uint16) map[common.Address]*common.VerifiedSign {
	verifiedSign := make(map[common.Address]*common.VerifiedSign)
	signCount := len(signs)
	for i := 0; i < signCount; i++ {
		sign := signs[i]
		account, signValidate, err := crypto.VerifySignWithValidate(signHash.Bytes(), sign.Bytes())
		if err != nil {
			log.ERROR("共识引擎", "验证签名 错误", err)
			continue
		}

		stock, findStock := stocks[account]
		if findStock == false {
			// can't find in stock, discard
			log.ERROR("共识引擎", "验证签名 股权未找到 node", account.Hex(), "签名：", signHash)
			continue
		}

		if existData, exist := verifiedSign[account]; exist {
			log.ERROR("共识引擎", "验证签名 重复签名 node", account.Hex())
			//already exist, replace "disagree" sign with "agree" sign
			if existData.Validate == false && signValidate == true {
				existData.Sign = sign
				existData.Account = account
				existData.Validate = signValidate
				existData.Stock = stock
			}
		} else {
			verifiedSign[account] = &common.VerifiedSign{Sign: sign, Account: account, Validate: signValidate, Stock: stock}
		}
	}

	return verifiedSign
}

func (md *MtxDPOS) verifyDPOS(verifiedSigns map[common.Address]*common.VerifiedSign, target *dposTarget) ([]common.Signature, error) {
	var agreeCount, disagreeCount int = 0, 0
	var agreeStock, disagreeStock uint64 = 0, 0

	rightSigns := make([]common.Signature, 0)

	for _, signInfo := range verifiedSigns {
		if signInfo.Validate == true {
			agreeCount++
			agreeStock += uint64(signInfo.Stock)
			rightSigns = append(rightSigns, signInfo.Sign)
		} else {
			disagreeCount++
			disagreeStock += uint64(signInfo.Stock)
			if disagreeCount > target.maxDisagreeCount {
				return nil, errDisagreeCountErr
			}
			if disagreeStock > target.maxDisagreeStock {
				return nil, errDisagreeStockErr
			}
		}
	}

	if agreeCount < target.targetCount {
		return nil, errSignCountErr
	}
	if agreeStock < target.targetStock {
		return nil, errSignStockErr
	}

	return rightSigns, nil
}

func (md *MtxDPOS) verifyBroadcastBlock(header *types.Header) error {
	if len(header.Signatures) != 1 {
		return errBroadcastSignCount
	}
	from, result, err := crypto.VerifySignWithValidate(header.HashNoSignsAndNonce().Bytes(), header.Signatures[0].Bytes())
	if err != nil {
		return err
	}
	if from != header.Leader {
		return errors.Errorf("broadcast block's sign account(%s) is not block leader(%s)", from.Hex(), header.Leader.Hex())
	}
	if md.isBroadcastRole(from) == false {
		return errBroadcastVerifySign
	}
	if result == false {
		return errBroadcastVerifySignFalse
	}
	return nil
}

func (md *MtxDPOS) getValidatorStocks(reader consensus.ValidatorReader, hash common.Hash) (map[common.Address]uint16, error) {
	graphInfo, err := md.getValidatorGraph(reader, hash)
	if err != nil {
		return nil, err
	}
	return md.graph2ValidatorStocks(graphInfo), nil
}

func (md *MtxDPOS) getValidatorGraph(reader consensus.ValidatorReader, hash common.Hash) (*mc.TopologyGraph, error) {
	graphInfo, err := reader.GetValidatorByHash(hash)
	if err != nil {
		return nil, err
	}
	return graphInfo, nil
}

func (md *MtxDPOS) graph2ValidatorStocks(graph *mc.TopologyGraph) map[common.Address]uint16 {
	stocks := make(map[common.Address]uint16)
	for _, node := range graph.NodeList {
		if node.Type != common.RoleValidator {
			continue
		}
		if _, exist := stocks[node.Account]; exist {
			continue
		}
		stocks[node.Account] = node.Stock
		//log.Info("DPOS引擎", "验证者", validator.Account, "股权", validator.Stock, "高度", graph.Number)
	}
	return stocks
}

func (md *MtxDPOS) isBroadcastRole(address common.Address) bool {
	for _, b := range manparams.BroadCastNodes {
		if b.Address == address {
			return true
		}
	}
	return false
}
