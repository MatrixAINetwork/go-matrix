// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mtxdpos

import (
	"math"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
)

const (
	DPOSDefStock = 1 // 默认股权值
)

type Config struct {
	TargetSignCountRatio       float64
	TargetStockRatio           float64
	MinStockCount              int
	FullSignThreshold          int
	SuperNodeFullSignThreshold int
}

var defaultConfig = Config{
	TargetSignCountRatio:       0.66667,
	TargetStockRatio:           0,
	MinStockCount:              3,
	FullSignThreshold:          7,
	SuperNodeFullSignThreshold: 3,
}

var simpleConfig = Config{
	TargetSignCountRatio:       0.66667,
	TargetStockRatio:           0,
	MinStockCount:              1,
	FullSignThreshold:          7,
	SuperNodeFullSignThreshold: 3,
}

var (
	errSignHashLenErr = errors.New("hash is required to be exactly 32 bytes")

	errStockCountErr = errors.New("stock count is less then min count")

	errSignCountErr = errors.New("sign count is not enough")

	errSignStockErr = errors.New("sign stock is not enough")

	errDisagreeCountErr = errors.New("consensus failed, cause to disagree count is too much")

	errDisagreeStockErr = errors.New("consensus failed, cause to disagree stock is too much")

	errBroadcastSignCount = errors.New("broadcast block's sign count err, not one")

	errBroadcastVerifySign = errors.New("broadcast block's sign is not from broadcast node")

	errBroadcastVerifySignFalse = errors.New("broadcast block's sign is false")

	errVersionSignCount = errors.New("block's version sign count err, not one")

	errVersionVerifySign = errors.New("block's version sign is not from super version account")

	errSuperBlockSignCount = errors.New("super block sign count err, not one")

	errSuperBlockVerifySign = errors.New("super block sign is not from super super block account")
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
	config Config
}

func NewMtxDPOS(simpleMode bool) *MtxDPOS {
	if simpleMode {
		return &MtxDPOS{config: simpleConfig}
	} else {
		return &MtxDPOS{config: defaultConfig}
	}
}

func (md *MtxDPOS) VerifyVersionSigns(reader consensus.StateReader, header *types.Header) error {
	var blockHash common.Hash
	number := header.Number.Uint64()
	if 0 == number {
		blockHash = header.Hash()
	} else {
		blockHash = header.ParentHash
	}
	accounts, err := reader.GetVersionSuperAccounts(blockHash)
	if err != nil || accounts == nil {
		return errors.Errorf("get super version account from state err(%s)", err)
	}

	targetCount := md.calcSuperNodeTarget(len(accounts))

	if len(header.VersionSignatures) < targetCount {
		log.ERROR("共识引擎", "版本号签名数量不足 size", len(header.Version), "target", targetCount)
		return errVersionSignCount
	}

	verifiedVersion := md.verifyHashWithSuperNodes(common.BytesToHash([]byte(header.Version)), header.VersionSignatures, accounts)
	//log.Debug("共识引擎", "版本", string(header.Version))
	if len(verifiedVersion) < targetCount {
		log.ERROR("共识引擎", "验证版本号签名,验证后的签名数量不足 size", len(verifiedVersion), "target", targetCount)
		return errVersionVerifySign
	}
	return nil
}

func (md *MtxDPOS) calcSuperNodeTarget(totalCount int) int {
	targetCount := 0
	if totalCount <= md.config.SuperNodeFullSignThreshold {
		targetCount = totalCount
	} else {
		targetCount = int(math.Ceil(float64(totalCount) * md.config.TargetSignCountRatio))
	}
	return targetCount
}

func (md *MtxDPOS) CheckSuperBlock(reader consensus.StateReader, header *types.Header) error {

	accounts, err := reader.GetBlockSuperAccounts(reader.GetCurrentHash())
	if err != nil || accounts == nil {
		return errors.Errorf("get super block account from state err(%s)", err)
	}

	targetCount := md.calcSuperNodeTarget(len(accounts))
	if len(header.Signatures) < targetCount {
		log.Error("共识引擎", "超级区块签名数量不足 size", len(header.Signatures), "target", targetCount)
		return errSuperBlockSignCount
	}
	verifiedSigh := md.verifyHashWithSuperNodes(header.HashNoSigns(), header.Signatures, accounts)
	if len(verifiedSigh) < targetCount {
		log.Error("共识引擎", "验证超级区块,验证后的签名数量不足 size", len(verifiedSigh), "target", targetCount, "hash", header.HashNoSigns().TerminalString())
		return errSuperBlockVerifySign
	}
	return nil
}

func (md *MtxDPOS) verifyHashWithSuperNodes(hash common.Hash, signatures []common.Signature, superNodes []common.Address) map[common.Address]byte {
	verifiedSigh := make(map[common.Address]byte, 0)
	for _, sigh := range signatures {
		account, _, err := crypto.VerifySignWithValidate(hash.Bytes(), sigh.Bytes())
		if nil != err {
			log.ERROR("共识引擎", "验证版本错误", err)
			continue
		}
		findFlag := 0
		for _, superAccount := range superNodes {
			if account == superAccount {
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

func (md *MtxDPOS) VerifyBlock(reader consensus.StateReader, header *types.Header) error {
	if nil == header {
		return errors.New("header is nil")
	}
	if err := md.VerifyVersionSigns(reader, header); err != nil {
		log.INFO("共识引擎", "验证版本号签名失败", err)
		return err
	}

	if header.IsSuperHeader() {
		return md.CheckSuperBlock(reader, header)
	}

	bcInterval, err := reader.GetBroadcastIntervalByHash(header.ParentHash)
	if err != nil {
		return errors.Errorf("get broadcast interval from reader err: %v", err)
	}

	number := header.Number.Uint64()
	if bcInterval.IsBroadcastNumber(number) {
		return md.verifyBroadcastBlock(reader, header)
	}

	stocks, err := md.getValidatorStocks(reader, header.ParentHash)
	if err != nil {
		return err
	}

	hash := header.HashNoSignsAndNonce()
	log.Trace("共识引擎", "VerifyBlock, 签名总数", len(header.Signatures), "hash", hash, "txhash:", header.Roots)
	_, err = md.VerifyHashWithStocks(reader, hash, header.Signatures, stocks, header.ParentHash)
	return err
}

func (md *MtxDPOS) VerifyHash(reader consensus.StateReader, signHash common.Hash, signs []common.Signature) ([]common.Signature, error) {
	return md.VerifyHashWithBlock(reader, signHash, signs, reader.GetCurrentHash())
}

func (md *MtxDPOS) VerifyHashWithBlock(reader consensus.StateReader, signHash common.Hash, signs []common.Signature, blockHash common.Hash) ([]common.Signature, error) {
	stocks, err := md.getValidatorStocks(reader, blockHash)
	if err != nil {
		return nil, err
	}

	return md.VerifyHashWithStocks(reader, signHash, signs, stocks, blockHash)
}

func (md *MtxDPOS) VerifyHashWithStocks(reader consensus.StateReader, signHash common.Hash, signs []common.Signature, stocks map[common.Address]uint16, blockHash common.Hash) ([]common.Signature, error) {
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

	verifiedSigns := md.verifySigns(reader, signHash, signs, stocks, blockHash)
	if len(verifiedSigns) < target.targetCount {
		log.ERROR("共识引擎", "验证后的签名数量不足 size", len(verifiedSigns), "target", target.targetCount)
		return nil, errSignCountErr
	}

	return md.verifyDPOS(verifiedSigns, target)
}

func (md *MtxDPOS) VerifyHashWithVerifiedSigns(reader consensus.StateReader, signs []*common.VerifiedSign) ([]common.Signature, error) {
	return md.VerifyHashWithVerifiedSignsAndBlock(reader, signs, reader.GetCurrentHash())
}

func (md *MtxDPOS) VerifyHashWithVerifiedSignsAndBlock(reader consensus.StateReader, signs []*common.VerifiedSign, blockHash common.Hash) ([]common.Signature, error) {
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
	if totalCount < md.config.MinStockCount {
		return nil, errStockCountErr
	}

	//calculate total stock
	var totalStock uint64 = 0
	for _, stock := range stocks {
		totalStock += uint64(stock)
	}

	//calculate target
	target := &dposTarget{totalCount: totalCount, totalStock: totalStock}
	if totalCount <= md.config.FullSignThreshold {
		target.targetCount = totalCount
		target.targetStock = totalStock
	} else {
		target.targetCount = int(math.Ceil(float64(totalCount) * md.config.TargetSignCountRatio))
		target.targetStock = uint64(math.Ceil(float64(totalStock) * md.config.TargetStockRatio))
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

func (md *MtxDPOS) verifySigns(reader consensus.StateReader, signHash common.Hash, signs []common.Signature, stocks map[common.Address]uint16, blockHash common.Hash) map[common.Address]*common.VerifiedSign {
	verifiedSign := make(map[common.Address]*common.VerifiedSign)
	signCount := len(signs)
	for i := 0; i < signCount; i++ {
		sign := signs[i]
		signAccount, signValidate, err := crypto.VerifySignWithValidate(signHash.Bytes(), sign.Bytes())
		if err != nil {
			log.ERROR("共识引擎", "验证签名 错误", err)
			continue
		}

		accountA0, _, err := reader.GetA0AccountFromAnyAccount(signAccount, blockHash)
		if err != nil {
			log.ERROR("共识引擎", "get auth account err", err)
			continue
		}

		stock, findStock := stocks[accountA0]
		if findStock == false {
			// can't find in stock, discard
			log.ERROR("共识引擎", "验证签名 股权未找到 node", accountA0.Hex(), "签名：", signHash)
			continue
		}

		if existData, exist := verifiedSign[accountA0]; exist {
			log.ERROR("共识引擎", "验证签名 重复签名 node", accountA0.Hex())
			//already exist, replace "disagree" sign with "agree" sign
			if existData.Validate == false && signValidate == true {
				existData.Sign = sign
				existData.Account = accountA0
				existData.Validate = signValidate
				existData.Stock = stock
			}
		} else {
			verifiedSign[accountA0] = &common.VerifiedSign{Sign: sign, Account: accountA0, Validate: signValidate, Stock: stock}
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

func (md *MtxDPOS) verifyBroadcastBlock(reader consensus.StateReader, header *types.Header) error {
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
	if result == false {
		return errBroadcastVerifySignFalse
	}

	broadcasts, err := reader.GetBroadcastAccounts(header.ParentHash)
	if err != nil || len(broadcasts) == 0 {
		return errors.Errorf("get broadcast account from state err(%s)", err)
	}
	for _, bc := range broadcasts {
		if from == bc {
			return nil
		}
	}
	return errBroadcastVerifySign
}

func (md *MtxDPOS) getValidatorStocks(reader consensus.StateReader, hash common.Hash) (map[common.Address]uint16, error) {
	topologyInfo, electInfo, err := reader.GetGraphByHash(hash)
	if err != nil {
		return nil, err
	}
	return md.graph2ValidatorStocks(topologyInfo, electInfo), nil
}

func (md *MtxDPOS) graph2ValidatorStocks(topologyInfo *mc.TopologyGraph, electInfo *mc.ElectGraph) map[common.Address]uint16 {
	stocks := make(map[common.Address]uint16)
	for _, node := range topologyInfo.NodeList {
		if node.Type != common.RoleValidator {
			continue
		}
		if _, exist := stocks[node.Account]; exist {
			continue
		}
		stocks[node.Account] = md.findStockInElect(node.Account, electInfo)
		//log.Info("DPOS引擎", "验证者", validator.Account, "股权", validator.Stock, "高度", graph.Number)
	}
	return stocks
}

func (md *MtxDPOS) findStockInElect(node common.Address, electInfo *mc.ElectGraph) uint16 {
	for _, elect := range electInfo.ElectList {
		if elect.Account == node {
			return elect.Stock
		}
	}
	return DPOSDefStock
}
