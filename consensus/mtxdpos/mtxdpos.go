// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package mtxdpos

import (
	"errors"
	"math"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
)

const (
	DPOSTargetSignCountRatio = 0.66667
	DPOSTargetStockRatio     = 0 // 暂时关闭股权的要求
	DPOSMinStockCount        = 3
	DPOSFullSignThreshold    = 7
)

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
	chain consensus.ChainReader
}

func NewMtxDPOS(chain consensus.ChainReader) *MtxDPOS {
	return &MtxDPOS{
		chain: chain,
	}
}

func (md *MtxDPOS) VerifyBlock(header *types.Header) error {
	if common.IsBroadcastNumber(header.Number.Uint64()) {
		return md.verifyBroadcastBlock(header)
	}

	stocks, err := md.getValidatorStocks(header.Number.Uint64())
	if err != nil {
		return err
	}

	hash := header.HashNoSignsAndNonce()
	log.INFO("共识引擎", "VerifyBlock, 签名总数", len(header.Signatures), "hash", hash)

	_, err = md.VerifyHashWithStocks(hash, header.Signatures, stocks)
	return err
}

func (md *MtxDPOS) VerifyHash(signHash common.Hash, signs []common.Signature) ([]common.Signature, error) {
	return md.VerifyHashWithNumber(signHash, signs, md.chain.CurrentHeader().Number.Uint64())
}

func (md *MtxDPOS) VerifyHashWithNumber(signHash common.Hash, signs []common.Signature, number uint64) ([]common.Signature, error) {
	stocks, err := md.getValidatorStocks(number)
	if err != nil {
		return nil, err
	}

	return md.VerifyHashWithStocks(signHash, signs, stocks)
}

func (md *MtxDPOS) VerifyHashWithStocks(signHash common.Hash, signs []common.Signature, stocks map[common.Address]uint16) ([]common.Signature, error) {
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
		log.ERROR("共识引擎", "验证后的签名数量不足 size", len(signs), "target", target.targetCount)
		return nil, errSignCountErr
	}

	return md.verifyDPOS(verifiedSigns, target)
}

func (md *MtxDPOS) VerifyHashWithVerifiedSigns(signs []*common.VerifiedSign) ([]common.Signature, error) {
	return md.VerifyHashWithVerifiedSignsAndNumber(signs, md.chain.CurrentHeader().Number.Uint64())
}

func (md *MtxDPOS) VerifyHashWithVerifiedSignsAndNumber(signs []*common.VerifiedSign, number uint64) ([]common.Signature, error) {
	stocks, err := md.getValidatorStocks(number)
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

	if md.isBroadcastRole(from) == false {
		return errBroadcastVerifySign
	}

	if result == false {
		return errBroadcastVerifySignFalse
	}

	return nil
}

func (md *MtxDPOS) getValidatorStocks(number uint64) (map[common.Address]uint16, error) {
	parentNumber := number
	if parentNumber != 0 {
		parentNumber--
	}

	graphInfo, err := ca.GetTopologyByNumber(common.RoleType(common.RoleValidator), parentNumber)
	if err != nil {
		return nil, err
	}

	stocks := make(map[common.Address]uint16)
	for _, validator := range graphInfo.NodeList {
		if _, exist := stocks[validator.Account]; exist {
			continue
		}

		stocks[validator.Account] = validator.Stock
		//log.Info("DPOS引擎", "验证者", validator.Account, "股权", validator.Stock, "高度", number)
	}

	return stocks, nil
}

func (md *MtxDPOS) isBroadcastRole(address common.Address) bool {
	for _, b := range params.BroadCastNodes {
		if b.Address == address {
			return true
		}
	}
	return false
}
