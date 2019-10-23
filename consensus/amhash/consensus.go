// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package amhash

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"runtime"
	"time"

	"github.com/MatrixAINetwork/go-matrix/mc"

	"github.com/MatrixAINetwork/go-matrix/common"
	mtxmath "github.com/MatrixAINetwork/go-matrix/common/math"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/consensus/misc"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
	"gopkg.in/fatih/set.v0"
)

// amhash proof-of-work protocol constants.
var (
	maxUncles              = 2                // Maximum number of uncles allowed in a single block
	allowedFutureBlockTime = 15 * time.Second // Max time from current time allowed for blocks, before they're considered future blocks
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	errLargeBlockTime         = errors.New("timestamp too big")
	errBCBlockTimeStamp       = errors.New("broadcast block timestamp err")
	errBlockTimeInterval      = errors.New("timestamp is below min block interval")
	errTooManyUncles          = errors.New("too many uncles")
	errDuplicateUncle         = errors.New("duplicate uncle")
	errUncleIsAncestor        = errors.New("uncle is ancestor")
	errDanglingUncle          = errors.New("uncle's parent is not ancestor")
	errInvalidDifficulty      = errors.New("non-positive difficulty")
	errInvalidMixDigest       = errors.New("invalid mix digest")
	errX11InvalidPoW          = errors.New("invalid x11 proof-of-work")
	errSm3InvalidPoW          = errors.New("invalid sm3 proof-of-work")
	errInvalidBasePow         = errors.New("invalid base proof-of-work")
	errCoinbase               = errors.New("invalid coinbase")
	errAICoinbase             = errors.New("invalid ai coinbase")
	errInvalidAIMine          = errors.New("invalid AI Mine Result")
	errNoPowBlockHasPow       = errors.New("no pow block has pow info")
	errNoAIBlockHasAI         = errors.New("no ai block has ai info")
	errFirstReelectBlockHasAI = errors.New("the first block of reelect period hash ai info")
)

// Author implements consensus.Engine, returning the header's coinbase as the
// proof-of-work verified author of the block.
func (amhash *Amhash) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules of the
// stock Matrix amhash engine.
func (amhash *Amhash) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool, ai bool) error {
	// Short circuit if the header is known, or it's parent not
	number := header.Number.Uint64()
	if chain.GetHeader(header.Hash(), number) != nil {
		return nil
	}
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	// Sanity checks passed, do a proper verification
	return amhash.verifyHeader(chain, header, parent, false, seal, ai)
}

// VerifyHeader checks whether a header conforms to the consensus rules of the
// stock Matrix amhash engine.
func (amhash *Amhash) VerifySignatures(signature []common.Signature) (bool, error) {
	// If we're running a full engine faking, accept any input as valid
	return true, nil
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications.
func (amhash *Amhash) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool, ais []bool) (chan<- struct{}, <-chan error) {
	// If we're running a full engine faking, accept any input as valid
	if len(headers) == 0 {
		abort, results := make(chan struct{}), make(chan error, len(headers))
		for i := 0; i < len(headers); i++ {
			results <- nil
		}
		return abort, results
	}

	// Spawn as many workers as allowed threads
	workers := runtime.GOMAXPROCS(0)
	if len(headers) < workers {
		workers = len(headers)
	}

	// Create a task channel and spawn the verifiers
	var (
		inputs = make(chan int)
		done   = make(chan int, workers)
		errors = make([]error, len(headers))
		abort  = make(chan struct{})
	)
	for i := 0; i < workers; i++ {
		go func() {
			for index := range inputs {
				errors[index] = amhash.verifyHeaderWorker(chain, headers, seals, ais, index)
				done <- index
			}
		}()
	}

	errorsOut := make(chan error, len(headers))
	go func() {
		defer close(inputs)
		var (
			in, out = 0, 0
			checked = make([]bool, len(headers))
			inputs  = inputs
		)
		for {
			select {
			case inputs <- in:
				if in++; in == len(headers) {
					// Reached end of headers. Stop sending to workers.
					inputs = nil
				}
			case index := <-done:
				for checked[index] = true; checked[out]; out++ {
					errorsOut <- errors[out]
					if out == len(headers)-1 {
						return
					}
				}
			case <-abort:
				return
			}
		}
	}()
	return abort, errorsOut
}

func (amhash *Amhash) verifyHeaderWorker(chain consensus.ChainReader, headers []*types.Header, seals []bool, ais []bool, index int) error {
	var parent *types.Header
	if index == 0 {
		parent = chain.GetHeader(headers[0].ParentHash, headers[0].Number.Uint64()-1)
	} else if headers[index-1].Hash() == headers[index].ParentHash {
		parent = headers[index-1]
	}
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	if chain.GetHeader(headers[index].Hash(), headers[index].Number.Uint64()) != nil {
		return nil // known block
	}
	return amhash.verifyHeader(chain, headers[index], parent, false, seals[index], ais[index])
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of the stock Matrix amhash engine.
func (amhash *Amhash) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	// Verify that there are at most 2 uncles included in this block
	if len(block.Uncles()) > maxUncles {
		return errTooManyUncles
	}
	// Gather the set of past uncles and ancestors
	uncles, ancestors := set.New(), make(map[common.Hash]*types.Header)

	number, parent := block.NumberU64()-1, block.ParentHash()
	for i := 0; i < 7; i++ {
		ancestor := chain.GetBlock(parent, number)
		if ancestor == nil {
			break
		}
		ancestors[ancestor.Hash()] = ancestor.Header()
		for _, uncle := range ancestor.Uncles() {
			uncles.Add(uncle.Hash())
		}
		parent, number = ancestor.ParentHash(), number-1
	}
	ancestors[block.Hash()] = block.Header()
	uncles.Add(block.Hash())

	// Verify each of the uncles that it's recent, but not an ancestor
	for _, uncle := range block.Uncles() {
		// Make sure every uncle is rewarded only once
		hash := uncle.Hash()
		if uncles.Has(hash) {
			return errDuplicateUncle
		}
		uncles.Add(hash)

		// Make sure the uncle has a valid ancestry
		if ancestors[hash] != nil {
			return errUncleIsAncestor
		}
		if ancestors[uncle.ParentHash] == nil || uncle.ParentHash == block.ParentHash() {
			return errDanglingUncle
		}
		if err := amhash.verifyHeader(chain, uncle, ancestors[uncle.ParentHash], true, true, false); err != nil {
			return err
		}
	}
	return nil
}

// verifyHeader checks whether a header conforms to the consensus rules of the
// stock Matrix amhash engine.
// See YP section 4.3.4. "Block Header Validity"
func (amhash *Amhash) verifyHeader(chain consensus.ChainReader, header, parent *types.Header, uncle bool, seal bool, ai bool) error {
	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
	}
	// Verify the header's timestamp
	if err := amhash.verifyHeaderTime(chain, header, parent, uncle); err != nil {
		return err
	}
	// super header don't verify difficulty
	if header.IsSuperHeader() == false {
		// Verify the block's difficulty based in it's timestamp and parent's difficulty
		expected, err := amhash.CalcDifficulty(chain, string(header.Version), header.Time.Uint64(), parent)
		if err != nil {
			return fmt.Errorf("calc difficulty err : %v", err)
		}
		if expected.Cmp(header.Difficulty) != 0 {
			return fmt.Errorf("invalid difficulty: have %v, want %v", header.Difficulty, expected)
		}
	}
	// Verify that the gas limit is <= 2^63-1
	cap := uint64(0x7fffffffffffffff)
	if header.GasLimit > cap {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, cap)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}

	// Verify that the gas limit remains within allowed bounds
	diff := int64(parent.GasLimit) - int64(header.GasLimit)
	if diff < 0 {
		diff *= -1
	}
	limit := parent.GasLimit / params.GasLimitBoundDivisor

	if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
		return fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parent.GasLimit, limit)
	}
	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return consensus.ErrInvalidNumber
	}
	// Verify the engine specific ai seal securing the block
	if ai {
		if err := amhash.VerifyAISeal(chain, header); err != nil {
			return err
		}
	}
	// Verify the engine specific seal securing the block
	if seal {
		if err := amhash.VerifySeal(chain, header); err != nil {
			return err
		}
	}
	// If all checks passed, validate any special fields for hard forks
	if err := misc.VerifyDAOHeaderExtraData(chain.Config(), header); err != nil {
		return err
	}
	if err := misc.VerifyForkHashes(chain.Config(), header, uncle); err != nil {
		return err
	}
	return nil
}

func (amhash *Amhash) verifyHeaderTime(chain consensus.ChainReader, header, parent *types.Header, uncle bool) error {
	if uncle {
		if header.Time.Cmp(mtxmath.MaxBig256) > 0 {
			return errLargeBlockTime
		}
	} else {
		if header.Time.Cmp(big.NewInt(time.Now().Add(allowedFutureBlockTime).Unix())) > 0 {
			return consensus.ErrFutureBlock
		}
	}

	bcInterval, err := chain.GetBroadcastIntervalByHash(header.ParentHash)
	if err != nil {
		return fmt.Errorf("get broadcast interval err: %v", err)
	}

	if bcInterval.IsBroadcastNumber(header.Number.Uint64()) {
		if !header.IsSuperHeader() {
			targetTime := big.NewInt(0).Add(parent.Time, big.NewInt(1))
			if header.Time.Cmp(targetTime) != 0 {
				return errBCBlockTimeStamp
			}
		}
	} else {
		minBlockTime := big.NewInt(0).Add(parent.Time, params.MinBlockInterval)
		if header.Time.Cmp(minBlockTime) < 0 {
			return errBlockTimeInterval
		}
	}
	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func (amhash *Amhash) CalcDifficulty(chain consensus.ChainReader, _ string, curTime uint64, parent *types.Header) (*big.Int, error) {
	if parent == nil {
		return nil, errors.New("父区块为空")
	}
	if curTime < parent.Time.Uint64() {
		return nil, errors.New("当前时间戳错误")
	}
	if chain == nil {
		return nil, errors.New("传入的chain 为空")
	}

	bcInterval, err := chain.GetBroadcastIntervalByHash(parent.Hash())
	if err != nil {
		return nil, err
	}

	curNumber := parent.Number.Uint64() + 1
	if !params.IsAIBlock(curNumber, bcInterval.BCInterval) {
		return common.Big0, nil
	}

	reelectionDifficulty, err := chain.GetReelectionDifficulty(parent.Hash())
	if err != nil {
		return nil, err
	}

	if amhash.ShoudEqualReelectionDifficulty(bcInterval, parent) {
		log.Info("CalcDifficulty", "换届区块,难度等于换届难度", reelectionDifficulty)

		return reelectionDifficulty, nil
	}

	minDifficulty, err := chain.GetMinDifficulty(parent.Hash())
	if err != nil {
		return nil, err
	}

	maxDifficulty, err := chain.GetMaxDifficulty(parent.Hash())
	if err != nil {
		return nil, err
	}

	innerMiners, err := chain.GetInnerMinerAccounts(parent.Hash())
	isTimeout := IsPowTimeout(parent.Coinbase, innerMiners)
	mineHeader, _, err := amhash.getMineHeader(parent.Number.Uint64(), parent.Hash(), bcInterval, chain)
	if err != nil {
		return nil, err
	}
	durationLimit := amhash.getDurationLimit()
	mineHeaderDuration, err := chain.GetBlockDurationStatus(mineHeader.Hash())
	if err != nil || len(mineHeaderDuration.Status) == 0 {
		log.Warn("难度计算", "获取难度状态出错", "")
		mineHeaderDuration = &mc.BlockDurationStatus{[]uint8{0}}
	}
	difficultyInfors, err := amhash.getDifficultyInfors(parent.Hash(), curNumber, curTime, bcInterval, chain)
	if err != nil {
		return nil, err
	}
	return FastTrackEMAAlg(bcInterval.GetReElectionInterval(), curNumber, difficultyInfors, mineHeader, minDifficulty, maxDifficulty, isTimeout, curTime, amhash.getBlockMinDuration(), durationLimit, mineHeaderDuration)
}

func (amhash *Amhash) getDifficultyInfors(parentHash common.Hash, curNumber, curTime uint64, bcInterval *mc.BCIntervalInfo, chain consensus.ChainReader) ([]DifficultyInfo, error) {
	headers := make([]*types.Header, 0)

	if curNumber%bcInterval.GetReElectionInterval() < params.PowBlockPeriod*params.VersionAIDifficultyQuickTraceNum+1 {
		return nil, nil
	}

	for i := curNumber - 1; i > bcInterval.LastReelectNumber; i-- {

		if params.VersionAIDifficultyAvgLength == uint64(len(headers)) {
			break
		}
		header := chain.GetHeaderByHash(parentHash)
		if nil == header {
			return nil, errors.New("难度计算，获取区块为空")
		}
		if header.IsAIHeader(bcInterval.BCInterval) {
			headers = append(headers, header)
		}
		parentHash = header.ParentHash
	}
	difficultyInfors := make([]DifficultyInfo, 0)
	for i := 1; i < len(headers); i++ {
		if headers[i].Difficulty.Uint64() != 0 {
			difficultyInfo := []DifficultyInfo{{difficulty: headers[i].Difficulty, Duration: headers[i-1].Time.Uint64() - headers[i].Time.Uint64()}}
			log.Trace("难度调整算法跟踪", "高度", headers[i].Number, "Difficulty", headers[i].Difficulty, "Cost", headers[i-1].Time.Uint64()-headers[i].Time.Uint64())
			difficultyInfors = append(difficultyInfo, difficultyInfors...)
		}
	}
	difficultyInfors = append(difficultyInfors, DifficultyInfo{headers[0].Difficulty, curTime - headers[0].Time.Uint64()})
	return difficultyInfors, nil
}

func (amhash *Amhash) isHasSuperBLock(parent *types.Header, grandParent *types.Header) bool {
	var isHasSuperBLock bool
	if parent.IsSuperHeader() || grandParent.IsSuperHeader() {
		isHasSuperBLock = true
	}
	return isHasSuperBLock
}

func (amhash *Amhash) isHasBroadcast(bcInterval *mc.BCIntervalInfo, parent *types.Header) bool {
	var isHasBroadCast bool
	if bcInterval.IsBroadcastNumber(parent.Number.Uint64()) {
		isHasBroadCast = true
	}
	return isHasBroadCast
}

func (amhash *Amhash) ShoudEqualReelectionDifficulty(bcInterval *mc.BCIntervalInfo, parent *types.Header) bool {
	var isHasReelection bool
	if bcInterval.IsReElectionNumber(parent.Number.Uint64()) {
		isHasReelection = true
	}
	return isHasReelection
}

func (amhash *Amhash) getDurationLimit() *big.Int {
	return new(big.Int).Mul(params.VersionAIDurationLimit, new(big.Int).SetUint64(params.PowBlockPeriod))
}
func (amhash *Amhash) getBlockMinDuration() uint64 {
	return new(big.Int).Mul(params.MinBlockInterval, new(big.Int).SetUint64(params.PowBlockPeriod)).Uint64()
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func FastTrackEMAAlg(reelectionNumber, curNumber uint64, difficultyInfors []DifficultyInfo, parentAIHeader *types.Header, minimumDifficulty, maxDifficulty *big.Int, isTimeout bool, curTime, blockMinInterval uint64, durationLimit *big.Int, mineHeaderDuration *mc.BlockDurationStatus) (x *big.Int, err error) {

	defer func() {
		if x.Cmp(minimumDifficulty) < 0 {
			x = minimumDifficulty
		}
		if x.Cmp(maxDifficulty) > 0 {
			x = maxDifficulty
		}
	}()

	bigParentDifficulty := new(big.Int).Set(parentAIHeader.Difficulty)

	//连续两次超时
	if isTimeout && mineHeaderDuration.Status[0] == 2 {
		y := new(big.Int)
		y.Div(bigParentDifficulty, params.VersionAIDifficultySecondBoundDivisor)
		y.Mul(y, big37)
		x = new(big.Int).Sub(bigParentDifficulty, y)
		return x, nil
	}

	if isTimeout {
		y := new(big.Int)
		y.Div(bigParentDifficulty, params.VersionAIDifficultyFirstBoundDivisor)
		y.Mul(y, big2)
		x = new(big.Int).Sub(bigParentDifficulty, y)
		return x, nil
	}

	if (curNumber % reelectionNumber) < params.PowBlockPeriod*params.VersionAIDifficultyQuickTraceNum+1 {
		//连续两次小于pos时间
		if isLessBlockMinInterval(curTime, parentAIHeader, blockMinInterval) && mineHeaderDuration.Status[0] == 1 {
			y := new(big.Int)
			y.Div(bigParentDifficulty, params.VersionAIDifficultySecondBoundDivisor)
			y.Mul(y, big60)
			x = new(big.Int).Add(bigParentDifficulty, y)
			return x, nil
		}

		//小于最小出块间隔固定加1/10
		if isLessBlockMinInterval(curTime, parentAIHeader, blockMinInterval) {
			y := new(big.Int).Div(bigParentDifficulty, params.VersionAIDifficultyFirstBoundDivisor)
			x = new(big.Int).Add(bigParentDifficulty, y)
			return x, nil
		}

		return computeDifficultyQuickTrace(curTime, parentAIHeader, durationLimit), nil
	}

	return EMAAlg(difficultyInfors, durationLimit)
}

// Some weird constants to avoid constant memory allocs for them.
var (
	expDiffPeriod = big.NewInt(100000)
	big1          = big.NewInt(1)
	big2          = big.NewInt(2)
	big9          = big.NewInt(9)
	big60         = big.NewInt(60)
	big37         = big.NewInt(37)
	big10         = big.NewInt(10)
	bigMinus99    = big.NewInt(-99)
	big2999999    = big.NewInt(2999999)
)

// VerifySeal implements consensus.Engine, checking whether the given block satisfies
// the PoW difficulty requirements.
func (amhash *Amhash) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	bcInterval, err := chain.GetBroadcastIntervalByHash(header.ParentHash)
	if err != nil {
		return fmt.Errorf("get broadcast interval err: %v", err)
	}

	if header.IsPowHeader(bcInterval.GetBroadcastInterval()) == false {
		if (header.Coinbase != common.Address{}) {
			return errNoPowBlockHasPow
		}
		return nil
	}

	if (header.Coinbase == common.Address{}) {
		// block no pow info
		return errCoinbase
	}

	mineHeader, mineHash, err := amhash.getMineHeader(header.Number.Uint64(), header.ParentHash, bcInterval, chain)
	if err != nil {
		return consensus.ErrUnknownAncestor
	}

	// check miner role
	role, err := amhash.verifyCoinbaseRole(chain, header.Coinbase, mineHeader.ParentHash)
	if err != nil {
		return err
	}

	// check base power info
	for _, v := range header.BasePowers {
		if err := amhash.verifyBasePow(chain, v, mineHeader, mineHash); err != nil {
			return err
		}
	}

	verifiedHeader := &types.Header{
		ParentHash: mineHash,
		Difficulty: mineHeader.Difficulty,
		VrfValue:   mineHeader.VrfValue,
		Coinbase:   header.Coinbase,
	}
	if role == common.RoleInnerMiner {
		verifiedHeader.Difficulty = params.InnerMinerDifficulty
	}

	// Ensure that we have a valid difficulty for the block
	if verifiedHeader.Difficulty.Sign() <= 0 {
		return errInvalidDifficulty
	}

	verifyData := generateMineData(verifiedHeader)

	// verify x11 pow
	x11Result := x11PowHash(verifyData, header.Nonce.Uint64())
	x11Target := new(big.Int).Div(maxUint256, verifiedHeader.Difficulty)
	if new(big.Int).SetBytes(Reverse(x11Result)).Cmp(x11Target) > 0 {
		return errX11InvalidPoW
	}

	// verify sm3 pow
	sm3Difficulty := big.NewInt(int64(math.Ceil(float64(verifiedHeader.Difficulty.Uint64()) * params.Sm3DifficultyRatio)))
	sm3Target := new(big.Int).Div(maxUint256, sm3Difficulty)
	sm3Result := sm3PowHash(verifyData, header.Sm3Nonce.Uint64())
	if new(big.Int).SetBytes(sm3Result).Cmp(sm3Target) > 0 {
		return errSm3InvalidPoW
	}

	return nil
}

func (amhash *Amhash) VerifyAISeal(chain consensus.ChainReader, header *types.Header) error {
	bcInterval, err := chain.GetBroadcastIntervalByHash(header.ParentHash)
	if err != nil {
		return fmt.Errorf("get broadcast interval err: %v", err)
	}

	if header.IsAIHeader(bcInterval.GetBroadcastInterval()) == false {
		if (header.AICoinbase != common.Address{}) {
			return errNoAIBlockHasAI
		}
		return nil
	}

	number := header.Number.Uint64()
	if number > 0 && bcInterval.IsReElectionNumber(number-1) {
		if (header.AICoinbase != common.Address{}) {
			return errFirstReelectBlockHasAI
		}
		return nil
	}

	if (header.AICoinbase == common.Address{}) {
		return errAICoinbase
	}

	aiMineHeader, aiMineHash, err := amhash.getMineHeader(header.Number.Uint64()-1, header.ParentHash, bcInterval, chain)
	if err != nil {
		return consensus.ErrUnknownAncestor
	}
	if _, err := amhash.verifyCoinbaseRole(chain, header.AICoinbase, aiMineHeader.ParentHash); err != nil {
		return fmt.Errorf("verify coinbase role err: %v", err)
	}

	verifiedHeader := &types.Header{
		ParentHash: aiMineHash,
		VrfValue:   aiMineHeader.VrfValue,
		AICoinbase: header.AICoinbase,
	}

	aiHash, _, err := amhash.aiMineProcess(chain, verifiedHeader, make(chan struct{}))
	if err != nil {
		return fmt.Errorf("ai mine process err: %v", err)
	}
	if aiHash != header.AIHash {
		return errInvalidAIMine
	}
	return nil
}

func (amhash *Amhash) getMineHeader(number uint64, sonHash common.Hash, bcInterval *mc.BCIntervalInfo, chain consensus.ChainReader) (*types.Header, common.Hash, error) {
	mineHashNumber := params.GetCurAIBlockNumber(number, bcInterval.GetBroadcastInterval())
	mineHeaderHash, err := chain.GetAncestorHash(sonHash, mineHashNumber)
	if err != nil {
		return nil, common.Hash{}, fmt.Errorf("get mine header hash err: %v", err)
	}
	mineHeader := chain.GetHeaderByHash(mineHeaderHash)
	if mineHeader == nil {
		return nil, common.Hash{}, fmt.Errorf("get mine header err")
	}

	return mineHeader, mineHeader.HashNoSignsAndNonce(), nil
}

func (amhash *Amhash) VerifyBasePow(chain consensus.ChainReader, header *types.Header, basePower types.BasePowers) error {
	bcInterval, err := chain.GetBroadcastIntervalByHash(header.ParentHash)
	if err != nil {
		return fmt.Errorf("get broadcast interval err: %v", err)
	}
	mineHeader, mineHash, err := amhash.getMineHeader(header.Number.Uint64(), header.ParentHash, bcInterval, chain)
	if err != nil {
		return consensus.ErrUnknownAncestor
	}
	return amhash.verifyBasePow(chain, basePower, mineHeader, mineHash)
}

func (amhash *Amhash) verifyBasePow(chain consensus.ChainReader, basePower types.BasePowers, mineHeader *types.Header, mineHash common.Hash) error {
	if (basePower.Miner == common.Address{}) {
		// block no pow info
		return errors.New("no basepower miner")
	}

	if _, err := amhash.verifyCoinbaseRole(chain, basePower.Miner, mineHeader.ParentHash); err != nil {
		return err
	}

	verifiedHeader := &types.Header{
		ParentHash: mineHash,
		Difficulty: params.BasePowerDifficulty,
		VrfValue:   mineHeader.VrfValue,
		Coinbase:   basePower.Miner,
	}

	result := x11PowHash(generateMineData(verifiedHeader), basePower.Nonce.Uint64())
	target := new(big.Int).Div(maxUint256, params.BasePowerDifficulty)
	if new(big.Int).SetBytes(Reverse(result)).Cmp(target) > 0 {
		return errInvalidBasePow
	}
	return nil
}

// Prepare implements consensus.Engine, initializing the difficulty field of a
// header to conform to the amhash protocol. The changes are done inline.
func (amhash *Amhash) Prepare(chain consensus.ChainReader, header *types.Header) error {
	parent := chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	difficulty, err := amhash.CalcDifficulty(chain, string(header.Version), header.Time.Uint64(), parent)
	if err != nil {
		return fmt.Errorf("calc difficulty err : %v", err)
	}
	header.Difficulty = difficulty
	return nil
}

// Finalize implements consensus.Engine, accumulating the block and uncle rewards,
// setting the final state and assembling the block.
func (manash *Amhash) GenOtherCurrencyBlock(chain consensus.ChainReader, header *types.Header, state *state.StateDBManage, uncles []*types.Header, currencyBlock []types.CurrencyBlock) (*types.Block, error) {
	// Accumulate any block and uncle rewards and commit the final state root
	//	accumulateRewards(chain.Config(), state, header, uncles)
	header.Roots, header.Sharding = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))

	// Header seems complete, assemble into a block and return
	return types.NewBlockCurrency(header, currencyBlock, uncles), nil
}

// Finalize implements consensus.Engine, accumulating the block and uncle rewards,
// setting the final state and assembling the block.
func (manash *Amhash) GenManBlock(chain consensus.ChainReader, header *types.Header, state *state.StateDBManage, uncles []*types.Header, currencyBlock []types.CurrencyBlock) (*types.Block, error) {
	// Accumulate any block and uncle rewards and commit the final state root
	//	accumulateRewards(chain.Config(), state, header, uncles)
	header.Roots, header.Sharding = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))

	// Header seems complete, assemble into a block and return
	return types.NewBlockMan(header, currencyBlock, uncles), nil
}

// Finalize implements consensus.Engine, accumulating the block and uncle rewards,
// setting the final state and assembling the block.
func (amhash *Amhash) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDBManage, uncles []*types.Header, currencyBlock []types.CurrencyBlock) (*types.Block, error) {
	// Accumulate any block and uncle rewards and commit the final state root
	//	accumulateRewards(chain.Config(), state, header, uncles)
	header.Roots, header.Sharding = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))

	// Header seems complete, assemble into a block and return
	return types.NewBlock(header, currencyBlock, uncles), nil
}

// Some weird constants to avoid constant memory allocs for them.
var (
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)
)

func (amhash *Amhash) verifyCoinbaseRole(chain consensus.ChainReader, coinbase common.Address, baseHash common.Hash) (common.RoleType, error) {
	//log.DEBUG("seal coinbase", "开始验证coinbase", header.Coinbase.Hex(), "高度", header.Number, "hash", header.Hash().Hex())
	innerMiners, err := chain.GetInnerMinerAccounts(baseHash)
	if err == nil {
		for _, account := range innerMiners {
			if account == coinbase {
				return common.RoleInnerMiner, nil
			}
		}
	} else {
		log.Error("seal coinbase", "get inner miner accounts err", err)
	}

	preTopology, _, err := chain.GetGraphByHash(baseHash)
	if err == nil {
		if preTopology.CheckAccountRole(coinbase, common.RoleMiner) {
			return common.RoleMiner, nil
		}
	} else {
		log.Error("seal coinbase", "get pre topology graph err", err)
	}

	return common.RoleNil, errCoinbase
}

func IsPowTimeout(coinbase common.Address, innerMiners []common.Address) bool {
	/*	if coinbase.Equal(common.Address{}) {
		return true
	}*/
	for _, v := range innerMiners {
		if coinbase.Equal(v) {
			return true
		}
	}
	return false
}
func isLeaderReelection(chain consensus.ChainReader, grandparent, parent *types.Header, bcInterval *mc.BCIntervalInfo, topologyGraph *mc.TopologyGraph) (bool, error) {

	if bcInterval.IsBroadcastNumber(parent.Number.Uint64()) || bcInterval.IsReElectionNumber(parent.Number.Uint64()-1) {
		return false, nil
	}

	cmpHeader := grandparent
	cmpNumber := parent.Number.Uint64() - 1
	for bcInterval.IsBroadcastNumber(uint64(cmpNumber)) || cmpHeader.IsSuperHeader() {
		if cmpNumber == 0 {
			return false, errors.New("无对比区块")
		}
		cmpNumber--
		cmpHeader = chain.GetHeaderByHash(cmpHeader.ParentHash)
		if cmpHeader == nil {
			return false, errors.New("获取不到父区块")
		}
	}
	nextLeader := topologyGraph.FindNextValidator(cmpHeader.Leader)
	return nextLeader != parent.Leader, nil

}

func isLessBlockMinInterval(curTime uint64, parentAIHeader *types.Header, blockMinInterval uint64) bool {
	return uint64(curTime)-parentAIHeader.Time.Uint64() <= blockMinInterval
}

type DifficultyInfo struct {
	difficulty *big.Int
	Duration   uint64
}

func computeDifficultyQuickTrace(curTime uint64, parent *types.Header, durationLimit *big.Int) *big.Int {
	x := new(big.Int)
	defer log.Info("cal Diff", "x", x)
	bigParentDifficulty := new(big.Int).Set(parent.Difficulty)
	parentTime := parent.Time.Int64()
	duration := int64(curTime) - parentTime
	diffTime := duration - durationLimit.Int64()

	if diffTime > 0 {
		bigDiffTime := big.NewInt(diffTime)
		delta := new(big.Int).Div(new(big.Int).Mul(bigParentDifficulty, bigDiffTime), new(big.Int).SetInt64(duration))
		delta = delta.Mul(delta, params.VersionAIDifficultyFallFac)
		delta = delta.Div(delta, params.VersionAIDifficultyNormalDivisor)
		//根据目标相差得倍数，调整下降系数
		rollFac := new(big.Int).Div(big.NewInt(duration), durationLimit)
		delta = delta.Mul(delta, rollFac)
		x = new(big.Int).Sub(bigParentDifficulty, delta)
	} else {
		diffTime = -diffTime
		bigDiffTime := big.NewInt(diffTime)
		delta := new(big.Int).Div(new(big.Int).Mul(bigParentDifficulty, bigDiffTime), new(big.Int).SetInt64(duration))
		delta = delta.Mul(delta, params.VersionAIDifficultyRiseFac)
		delta = delta.Div(delta, params.VersionAIDifficultyNormalDivisor)
		x = new(big.Int).Add(bigParentDifficulty, delta)
	}

	return x
}

func EMAAlg(difficultyInfos []DifficultyInfo, durationLimit *big.Int) (*big.Int, error) {

	if uint64(len(difficultyInfos)) != params.VersionAIDifficultyAvgLength {
		log.Error("难度调整算法跟踪", "传入数据长度有错误", len(difficultyInfos))
		return big.NewInt(0), errors.New("获取的长度不等于滑动窗口长度")
	}

	filtedTarget := new(big.Int).SetUint64(0)

	for i := uint64(0); i < params.VersionAIDifficultyAvgLength; i++ {
		if difficultyInfos[i].difficulty.Cmp(big1) == -1 {
			log.Error("难度调整算法跟踪", "获取区块difficulty错误", difficultyInfos[i].difficulty)
			return big.NewInt(0), errors.New("获取区块difficulty错误")
		}
		if 0 == i {
			filtedTarget = new(big.Int).Div(maxUint256, difficultyInfos[i].difficulty)
		} else {
			filtedTarget = filtedTarget.Mul(filtedTarget, big.NewInt(int64(i+1)))
			filtedTarget.Add(filtedTarget, new(big.Int).Div(maxUint256, difficultyInfos[i].difficulty))
			filtedTarget = filtedTarget.Div(filtedTarget, big.NewInt(int64(i+2)))
		}
	}

	actualTime := uint64(0)
	for i := uint64(0); i < params.VersionAIDifficultyAvgLength; i++ {
		actualTime += difficultyInfos[i].Duration
	}
	targetTime := durationLimit.Uint64() * params.VersionAIDifficultyAvgLength

	if actualTime*3 < targetTime {
		actualTime = targetTime / 3
	}
	if actualTime > 3*targetTime {
		actualTime = 3 * targetTime
	}

	filtedTarget = filtedTarget.Mul(filtedTarget, new(big.Int).SetUint64(actualTime))
	filtedTarget = filtedTarget.Div(filtedTarget, new(big.Int).SetUint64(targetTime))

	if filtedTarget.Cmp(big1) == -1 {
		return big.NewInt(0), nil
	}
	x := new(big.Int).Div(maxUint256, filtedTarget)
	log.Trace("难度调整算法跟踪", "下一个区块难度", x.Uint64())
	return x, nil
}
