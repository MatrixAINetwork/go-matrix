// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package amhashzeta

import (
	crand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"runtime"
	"sync"

	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/consensus/ai"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/pkg/errors"
)

var (
	aiPictureMaxCount = 64000 // AI图库数量
	aiPictureSize     = 16    // AI选取图片数量
)

func (amhash *Amhash) SealAI(chain consensus.ChainReader, header *types.Header, stop <-chan struct{}) (*consensus.SealResult, error) {
	log.Info("amhash zeta sealer", "AI挖矿", "开始", "高度", header.Number.Uint64())
	defer log.Info("amhash zeta sealer", "AI挖矿", "结束", "高度", header.Number.Uint64())

	curHeader := types.CopyHeader(header)
	// start ai mining first
	aiHash, stopped, err := amhash.aiMineProcess(chain, curHeader, stop)
	if err != nil {
		return nil, err
	}
	if stopped {
		return nil, nil
	}

	curHeader.AIHash = aiHash
	curHeader.Nonce = types.BlockNonce{}

	return &consensus.SealResult{consensus.SealTypeAI, curHeader}, nil
}

// Seal implements consensus.Engine, attempting to find a nonce that satisfies
// the block's difficulty requirements.
func (amhash *Amhash) SealPow(chain consensus.ChainReader, header *types.Header, stop <-chan struct{}, resultchan chan<- *consensus.SealResult, isBroadcastNode bool) (*consensus.SealResult, error) {
	log.Info("amhash zeta sealer", "POW挖矿", "开始", "高度", header.Number.Uint64())
	defer log.Info("amhash zeta sealer", "POW挖矿", "结束", "高度", header.Number.Uint64())
	curHeader := types.CopyHeader(header)
	amhash.lock.Lock()
	if amhash.rand == nil {
		seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			amhash.lock.Unlock()
			return nil, err
		}
		amhash.rand = rand.New(rand.NewSource(seed.Int64()))
	}
	amhash.lock.Unlock()

	var x11Header *types.Header
	x11Header, stopped, err := amhash.x11MineProcess(chain, curHeader, stop, resultchan)
	if err != nil {
		return nil, errors.Errorf("x11 mining err: %v", err)
	}
	if stopped {
		return nil, nil
	}
	if x11Header == nil {
		return nil, errors.New("x11 mine result is nil")
	}

	var sm3Header *types.Header
	sm3Header, stopped, err = amhash.sm3MineProcess(chain, curHeader, stop, resultchan)
	if err != nil {
		return nil, errors.Errorf("x11 mining err: %v", err)
	}
	if stopped {
		return nil, nil
	}
	if sm3Header == nil {
		return nil, errors.New("sm3 mine result is nil")
	}

	curHeader.Nonce = x11Header.Nonce
	curHeader.MixDigest = x11Header.MixDigest
	curHeader.Sm3Nonce = sm3Header.Sm3Nonce
	curHeader.AIHash = common.Hash{}
	result := &consensus.SealResult{consensus.SealTypePow, curHeader}
	resultchan <- result
	return result, nil
}

func generateMineData(header *types.Header, mixDigest common.Hash) []byte {
	data := header.ParentHash.Bytes()
	data = append(data, header.Coinbase.Bytes()...)
	for i := 0; i < 12; i++ {
		data = append(data, byte(0))
	}
	data = append(data, mixDigest[0:12]...)
	return data
}

func (amhash *Amhash) aiMineProcess(chain consensus.ChainReader, header *types.Header, stop <-chan struct{}) (common.Hash, bool, error) {
	abortCh := make(chan struct{}, 1)
	foundCh := make(chan []byte, 1)
	errCh := make(chan error, 1)

	go amhash.startAIMining(chain, header, abortCh, foundCh, errCh)

	for {
		select {
		case <-stop:
			log.Info("amhash zeta sealer", "Sealer receive stop mine msg", "ai mine stop", "parent hash", header.ParentHash)
			close(abortCh)
			return common.Hash{}, true, nil

		case <-amhash.update:
			close(abortCh)
			return amhash.aiMineProcess(chain, header, stop)

		case err := <-errCh:
			log.Warn("amhash zeta sealer", "ai mining err", err)
			return common.Hash{}, false, err

		case result := <-foundCh:
			aiHash := common.BytesToHash(result)
			log.Info("amhash zeta sealer", "aiMineProcess", "get ai digging result", "AIHash", aiHash)
			return aiHash, false, nil
		}
	}
}

func (amhash *Amhash) startAIMining(chain consensus.ChainReader, header *types.Header, abort chan struct{}, found chan []byte, errCh chan error) {
	// get seed
	vrf := baseinterface.NewVrf()
	_, vrfValue, _ := vrf.GetVrfInfoFromHeader(header.VrfValue)
	seed := big.NewInt(0).Add(types.RlpHash(vrfValue).Big(), header.AICoinbase.Big()).Int64()
	log.Info("amhash zeta sealer", "start ai mining", seed, "vrf", types.RlpHash(vrfValue).Hex(), "coinbase", header.AICoinbase.Hex())
	ai.Mining(seed, abort, found, errCh)
}

func (amhash *Amhash) x11MineProcess(chain consensus.ChainReader, header *types.Header, stop <-chan struct{}, resultChan chan<- *consensus.SealResult) (*types.Header, bool, error) {
	// Create a runner and the multiple search threads it directs
	log.Info("amhash zeta sealer", "x11 mine process", "begin", "number", header.Number)
	defer log.Info("amhash zeta sealer", "x11 mine process", "end", "number", header.Number)
	abort := make(chan struct{})
	found := make(chan *consensus.SealResult)
	/*amhash.lock.Lock()
	threads := amhash.threads
	if amhash.rand == nil {
		seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			amhash.lock.Unlock()
			return nil, false, err
		}
		amhash.rand = rand.New(rand.NewSource(seed.Int64()))
	}
	amhash.lock.Unlock()*/
	threads := runtime.NumCPU()

	var pend sync.WaitGroup
	for i := 0; i < threads; i++ {
		pend.Add(1)
		go func(id int, nonce uint64, abortCh chan struct{}) {
			defer pend.Done()
			amhash.x11Mine(header, id, nonce, abortCh, found)
		}(i, uint64(amhash.rand.Int63()), abort)
	}
	// Wait until sealing is terminated or a nonce is found
	var result *consensus.SealResult
	var isStop = false
x11seal:
	for {
		select {
		case <-stop:
			log.Info("amhash zeta sealer", "x11 process", "receive stop mine", "curHeader", header.ParentHash.TerminalString())
			// Outside abort, stop all miner threads
			if nil != abort {
				close(abort)
				abort = nil
			}
			isStop = true
			pend.Wait()
			return nil, isStop, nil
		case result = <-found:
			log.Info("amhash zeta sealer", "x11 process", "receive result", "type", result.Type, "number", result.Header.Number)
			// One of the threads found a block, abort all others
			if result.Type == consensus.SealTypeBasePow {
				resultChan <- result
			} else {
				log.Info("amhash zeta sealer", "x11 process", "Successfully sealed new x11 result", "nonce", result.Header.Nonce)
				if nil != abort {
					close(abort)
					abort = nil
				}
				break x11seal
			}

		case <-amhash.update:
			// Thread count was changed on user request, restart
			log.Info("amhash zeta sealer", "x11 process", "receive update", "number", header.Number, "cur difficulty", header.Difficulty)
			if nil != abort {
				close(abort)
				abort = nil
			}
			pend.Wait()
			return amhash.x11MineProcess(chain, header, stop, resultChan)
		}
	}
	// Wait for all miners to terminate and return the block
	pend.Wait()
	return result.Header, isStop, nil
}

// mine is the actual proof-of-work miner that searches for a nonce starting from
// seed that results in correct final block difficulty.
func (amhash *Amhash) x11Mine(header *types.Header, id int, seed uint64, abort chan struct{}, found chan *consensus.SealResult) {
	// Extract some data from the header
	var (
		curHeader         = types.CopyHeader(header)
		mixDigest         = types.RlpHash(seed)
		mineData          = generateMineData(curHeader, mixDigest)
		target            = new(big.Int).Div(maxUint256, header.Difficulty)
		basePowerTarget   = new(big.Int).Div(maxUint256, params.BasePowerDifficulty)
		basePowerFindFlag = false
		number            = curHeader.Number.Uint64()
	)
	curHeader.MixDigest = mixDigest

	logger := log.New("x11 miner", id)
	// Start generating random nonces until we abort or find a good one
	logger.Info("amhash zeta sealer", "begin", number, "target", target, "diff", header.Difficulty.Uint64())
	defer logger.Info("amhash zeta sealer", "end", number, "diff", header.Difficulty.Uint64())
	var (
		attempts = int64(0)
		nonce    = seed
	)
	logger.Trace("amhash zeta sealer", "Started x11 mine search for new nonces, seed", seed)
x11search:
	for {
		select {
		case <-abort:
			// Mining terminated, update stats and abort
			logger.Trace("amhash zeta sealer", "x11 mine nonce search aborted, attempts", nonce-seed)
			amhash.hashrate.Mark(attempts)
			return

		default:
			// We don't have to update hash rate on every nonce, so update after after 2^X nonces
			attempts++
			if (attempts % (1 << 15)) == 0 {
				amhash.hashrate.Mark(attempts)
				attempts = 0
			}
			// Compute the PoW value of this nonce
			result := x11PowHash(mineData, nonce)
			resultBigInt := new(big.Int).SetBytes(Reverse(result))
			if resultBigInt.Cmp(target) <= 0 {
				// Correct nonce found, create a new header with it
				resultHeader := types.CopyHeader(curHeader)
				resultHeader.Nonce = types.EncodeNonce(nonce)
				// Seal and return a block (if still needed)
				select {
				case found <- &consensus.SealResult{consensus.SealTypePow, resultHeader}:
					logger.Trace("amhash zeta sealer", "x11 nonce found and reported, attempts", nonce-seed, "nonce", nonce)
				case <-abort:
					logger.Trace("amhash zeta sealer", "x11 nonce found but discarded, attempts", nonce-seed, "nonce", nonce)
				}
				break x11search
			} else if resultBigInt.Cmp(basePowerTarget) <= 0 && !basePowerFindFlag {
				// Correct nonce found, create a new header with it
				baseHeader := types.CopyHeader(curHeader)
				baseHeader.Nonce = types.EncodeNonce(nonce)
				baseHeader.Difficulty = params.BasePowerDifficulty
				baseHeader.AIHash = common.Hash{}
				// Seal and return a block (if still needed)
				select {
				case found <- &consensus.SealResult{consensus.SealTypeBasePow, baseHeader}:
					basePowerFindFlag = true
					logger.Trace("amhash zeta sealer", "x11 base power nonce found and reported, attempts", nonce-seed, "nonce", nonce)
				case <-abort:
					logger.Trace("amhash zeta sealer", "x11 base power nonce found but discarded, attempts", nonce-seed, "nonce", nonce)
					break x11search
				}
			}
			nonce++
		}
	}
}

func (amhash *Amhash) sm3MineProcess(chain consensus.ChainReader, header *types.Header, stop <-chan struct{}, resultChan chan<- *consensus.SealResult) (*types.Header, bool, error) {
	log.Info("amhash zeta sealer", "sm3 mine process", "begin", "number", header.Number)
	defer log.Info("amhash zeta sealer", "sm3 mine process", "end", "number", header.Number)
	// Create a runner and the multiple search threads it directs
	abort := make(chan struct{})
	found := make(chan *types.Header)
	threads := runtime.NumCPU()

	var pend sync.WaitGroup
	for i := 0; i < threads; i++ {
		pend.Add(1)
		go func(id int, nonce uint64, abortCh chan struct{}) {
			defer pend.Done()
			amhash.sm3Mine(header, id, nonce, abortCh, found)
		}(i, uint64(amhash.rand.Int63()), abort)
	}
	// Wait until sealing is terminated or a nonce is found
	var result *types.Header
	var isStop = false
sm3seal:
	for {
		select {
		case <-stop:
			log.Info("amhash zeta sealer", "sm3 sealer receive stop mine", header.Number, "parent hash", header.ParentHash.TerminalString())
			// Outside abort, stop all miner threads
			if nil != abort {
				close(abort)
				abort = nil
			}
			isStop = true
			break sm3seal
		case result = <-found:
			// One of the threads found a block, abort all others
			log.Info("amhash zeta sealer", "successfully sealed new sm3 result", result.Sm3Nonce, "number", result.Number)
			if nil != abort {
				close(abort)
				abort = nil
			}
			break sm3seal

		case <-amhash.update:
			// Thread count was changed on user request, restart
			if nil != abort {
				close(abort)
				abort = nil
			}
			pend.Wait()
			return amhash.sm3MineProcess(chain, header, stop, resultChan)
		}
	}
	// Wait for all miners to terminate and return the block
	pend.Wait()
	return result, isStop, nil
}

func calcSm3Difficulty(header *types.Header) *big.Int {
	headerDifficulty := big.NewInt(int64(math.Ceil(float64(header.Difficulty.Uint64()) * params.Sm3DifficultyRatio)))
	if headerDifficulty.Cmp(params.ZetaSM3MaxDifficulty) > 0 {
		// headerDifficulty > 	params.ZetaSM3MaxDifficulty 使用最大难度值
		return params.ZetaSM3MaxDifficulty
	} else {
		return headerDifficulty
	}
}

func (amhash *Amhash) sm3Mine(header *types.Header, id int, seed uint64, abort chan struct{}, found chan *types.Header) {
	// Extract some data from the header
	var (
		curHeader     = types.CopyHeader(header)
		mineData      = generateMineData(curHeader, common.Hash{})
		sm3Difficulty = calcSm3Difficulty(header)
		target        = new(big.Int).Div(maxUint256, sm3Difficulty)
		number        = curHeader.Number.Uint64()
	)
	logger := log.New("sm3 miner", id)
	// Start generating random nonces until we abort or find a good one
	logger.Info("amhash zeta sealer", "begin sm3 mine", number, "target", target, "diff", sm3Difficulty.Uint64(), "header diff", header.Difficulty)
	defer logger.Info("amhash zeta sealer", "stop sm3 mine", number, "diff", sm3Difficulty.Uint64())
	var (
		attempts = int64(0)
		nonce    = seed
	)
	logger.Trace("amhash zeta sealer", "Started sm3 mine search for new nonces, seed", seed)
sm3search:
	for {
		select {
		case <-abort:
			// Mining terminated, update stats and abort
			logger.Trace("amhash zeta sealer", "pow mine nonce search aborted, attempts", nonce-seed)
			amhash.hashrate.Mark(attempts)
			return

		default:
			// We don't have to update hash rate on every nonce, so update after after 2^X nonces
			attempts++
			if (attempts % (1 << 15)) == 0 {
				amhash.hashrate.Mark(attempts)
				attempts = 0
			}
			// Compute the PoW value of this nonce
			result := sm3PowHash(mineData, nonce)

			if new(big.Int).SetBytes(result).Cmp(target) <= 0 {
				// Correct nonce found, create a new header with it
				resultHeader := types.CopyHeader(curHeader)
				resultHeader.Sm3Nonce = types.EncodeNonce(nonce)
				// Seal and return a block (if still needed)
				select {
				case found <- resultHeader:
					logger.Trace("amhash zeta sealer", "sm3 nonce found and reported, attempts", nonce-seed, "nonce", nonce)
				case <-abort:
					logger.Trace("amhash zeta sealer", "sm3 nonce found but discarded, attempts", nonce-seed, "nonce", nonce)
				}
				break sm3search
			}
			nonce++
		}
	}
}
