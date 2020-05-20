// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package miner

import (
	"errors"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"crypto/sha256"
	"github.com/MatrixAINetwork/go-matrix/base58"
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
	"math"
	"strings"
)

var (
	Submitx11Nonce string = "x11Nonce"
	SubmitAiHash   string = "AiHash"
	SubmitSM3Nonce string = "SM3Nonce"
)

type hashrate struct {
	ping time.Time
	rate uint64
}

type RemoteAgent struct {
	mu sync.Mutex

	quitCh   chan struct{}
	workCh   chan *Work
	returnCh chan<- *consensus.SealResult

	chain       consensus.ChainReader
	engine      map[string]consensus.Engine
	currentWork *Work
	work        map[common.Hash]*Work

	hashrateMu sync.RWMutex
	hashrate   map[common.Hash]hashrate

	running int32 // running indicates whether the agent is active. Call atomically
	workid  int64
}

func NewRemoteAgent(chain consensus.ChainReader, engine map[string]consensus.Engine) *RemoteAgent {
	return &RemoteAgent{
		chain:    chain,
		engine:   engine,
		work:     make(map[common.Hash]*Work),
		hashrate: make(map[common.Hash]hashrate),
		workid:   0,
	}
}

func (a *RemoteAgent) SubmitHashrate(id common.Hash, rate uint64) {
	a.hashrateMu.Lock()
	defer a.hashrateMu.Unlock()

	a.hashrate[id] = hashrate{time.Now(), rate}
}

func (a *RemoteAgent) Work() chan<- *Work {
	return a.workCh
}

func (a *RemoteAgent) SetReturnCh(returnCh chan<- *consensus.SealResult) {
	a.returnCh = returnCh
}

func (a *RemoteAgent) Start() {
	if !atomic.CompareAndSwapInt32(&a.running, 0, 1) {
		return
	}
	a.quitCh = make(chan struct{})
	a.workCh = make(chan *Work, 1)
	go a.loop(a.workCh, a.quitCh)
}

func (a *RemoteAgent) Stop() {
	if !atomic.CompareAndSwapInt32(&a.running, 1, 0) {
		return
	}
	close(a.quitCh)
	close(a.workCh)
}

// GetHashRate returns the accumulated hashrate of all identifier combined
func (a *RemoteAgent) GetHashRate() (tot int64) {
	a.hashrateMu.RLock()
	defer a.hashrateMu.RUnlock()

	// this could overflow
	for _, hashrate := range a.hashrate {
		tot += int64(hashrate.rate)
	}
	return
}

func calcSm3Difficulty(x11Difficulty *big.Int) *big.Int {
	headerDifficulty := big.NewInt(int64(math.Ceil(float64(x11Difficulty.Uint64()) * params.Sm3DifficultyRatio)))
	if headerDifficulty.Cmp(params.ZetaSM3MaxDifficulty) > 0 {
		// headerDifficulty > 	params.ZetaSM3MaxDifficulty
		return params.ZetaSM3MaxDifficulty
	} else {
		return headerDifficulty
	}
}

func (a *RemoteAgent) GetWork() ([6]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var res [6]string

	if a.currentWork != nil {
		block := a.currentWork.header

		res[0] = block.ParentHash.Hex()

		vrf := baseinterface.NewVrf()
		_, vrfValue, _ := vrf.GetVrfInfoFromHeader(block.VrfValue)
		seed := types.RlpHash(vrfValue).Hex()

		res[1] = seed
		n := block.Difficulty

		res[2] = common.BytesToHash(n.Bytes()).Hex()
		res[3] = a.currentWork.mineType.String()
		res[4] = block.Coinbase.Hex()
		res[5] = common.BytesToHash(calcSm3Difficulty(n).Bytes()).Hex() //sm3
		log.Info("GetWork", "work type", res[3], "Coinbase", res[4], "vrf", seed, "parentHash", block.ParentHash.Hex())
		//a.work[block.HashNoNonce()] = a.currentWork
		a.work[block.ParentHash] = a.currentWork

		return res, nil
	}
	return res, errors.New("No work available yet, don't panic.")
}

// SubmitWork tries to inject a pow solution into the remote agent, returning
// whether the solution was accepted or not (not can be both a bad pow as well as
// any other error, like no work pending).
func (a *RemoteAgent) SubmitWork(strNonce, strAIHah, strHash, strMinerAddr, seed string, dataType string, strWorkid string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	log.Info("SubmitWork", "nonce", strNonce, "AIHah", strAIHah, "header hash", strHash, "addr", strMinerAddr, "dataType", dataType, "strWorkid", strWorkid)
	var hash, AIHah common.Hash
	nonce := types.BlockNonce{}
	var err error

	// Make sure the work submitted is present
	//work := a.currentWork //a.work[hash]

	hash = common.HexToHash(strHash)
	log.Info("SubmitWork", "parentHash", hash.Hex())

	work := a.work[hash]
	if work == nil {
		log.Error("Work submitted but none pending", "err ", "work is nil")
		return false
	}

	tmpHead := work.header
	addr := common.HexToAddress(strMinerAddr)
	if addr == (common.Address{}) {
		addr, err = base58.Base58DecodeToAddress(strMinerAddr)
		if err != nil {
			log.Error("SubmitWork submitted but miner address", "err", err, "miner address", strMinerAddr)
			return false
		}
	}

	//workid, _ := strconv.ParseInt(strWorkid, 10, 64)
	//result := work.header
	tmpHead.Coinbase = addr
	switch dataType {
	case Submitx11Nonce:
		if (strings.Index(strHash, "0x") > 0 && len(strHash) != 66) || (strings.Index(strHash, "0x") < 0 && len(strHash) != 64) {
			log.Error("SubmitWork", "SubmitWork err, hash length wrong. nonce", strHash, "dataType", dataType)
			return false
		}
		hash = common.HexToHash(strHash)
		if len(strNonce) > 10 {
			log.Error("SubmitWork", "SubmitWork err, nonce too long ", strNonce, "hash", hash)
			return false
		}
		if len(strNonce) != 0 {
			nonce, err = reverseToNonce(common.FromHex(strNonce))

			if err != nil {
				log.Error("SubmitWork", "SubmitWork err, nonce is illegal ", strNonce)
				return false
			}
		}
		/*if a.workid > workid || hash != work.header.ParentHash {
			log.Error("SubmitWork", "recv x11 Workid or ParentHash mismatch", "a.workid", a.workid, "param workid", workid,
				"submit hash", hash.Hex(), "ParentHash", work.header.ParentHash.Hex(), "block number", work.header.Number)
			return false
		} else if a.workid <= workid {
			a.workid = workid
		}*/
		tmpHead.Nonce = nonce
		mixDigest := sha256.Sum256([]byte(seed))
		tmpHead.MixDigest.SetBytes(mixDigest[:])
		//result.Nonce = nonce
	case SubmitAiHash:
		//if len(strAIHah) != 66 {
		if (strings.Index(strAIHah, "0x") > 0 && len(strAIHah) != 66) || (strings.Index(strAIHah, "0x") < 0 && len(strAIHah) != 64) {
			log.Error("SubmitWork", "SubmitWork err, AIHash length wrong. AIHah", strAIHah, "dataType", dataType)
			return false
		}
		AIHah = common.HexToHash(strAIHah)
		tmpHead.AICoinbase = addr
		tmpHead.AIHash = AIHah
		//engine, exist := a.engine[string(tmpHead.Version)]
		//if exist == false {
		//	log.Warn("SubmitWork", "SubmitWork err", "can`t get engine by version", "version", string(tmpHead.Version))
		//	return false
		//}
		//if err := engine.VerifyAISeal(a.chain, tmpHead); err != nil {
		//	log.Warn("SubmitWork", "Invalid ai work submitted", hash, "err", err)
		//	return false
		//}
		a.returnCh <- &consensus.SealResult{consensus.SealTypeAI, tmpHead}
		return true

	case SubmitSM3Nonce:
		if len(strNonce) > 10 {
			log.Error("SubmitWork", "SubmitWork err, nonce too long ", strNonce)
			return false
		}
		if len(strNonce) != 0 {
			noncebig, _ := new(big.Int).SetString(strings.TrimPrefix(strNonce, "0x"), 16)
			nonce = types.EncodeNonce(noncebig.Uint64())
			log.Info("SubmitWork", "sm3 nonce", nonce.Uint64(), "strnonce", strNonce)
			if err != nil {
				log.Info("SubmitWork", "submitWork,recv nonce len less 8", err, "datatype", dataType)
				return false
			}
		}
		/*hash = common.HexToHash(strHash)
		if a.workid > workid || hash != work.header.ParentHash {
			log.Error("SubmitWork", "submitWork,recv sm3 workid too low or ParentHash mismatch", a.workid, "param workid", workid,
				"submit hash", hash.Hex(), "ParentHash", work.header.ParentHash.Hex(), "block number", work.header.Number)
			return false
		} else if a.workid <= workid {
			a.workid = workid
		}
		a.workid = workid*/
		if tmpHead.Sm3Nonce.Uint64() <= 0 {
			tmpHead.Sm3Nonce = nonce
		}
		//result.Sm3Nonce = nonce
	default:
		log.Error("SubmitWork", "SubmitWork err, unknow datatype", dataType)
		return false
	}

	// Make sure the Engine solutions is indeed valid

	log.Info("SubmitWork", "verify header", tmpHead.Hash(), "hash", strHash, "number", tmpHead.Number, "SM3 nonce", tmpHead.Sm3Nonce.Uint64(), "x11 nonce", tmpHead.Nonce.Uint64())

	if tmpHead.Sm3Nonce.Uint64() > 0 && tmpHead.Nonce.Uint64() > 0 {
		//engine, exist := a.engine[string(tmpHead.Version)]
		//if exist == false {
		//	log.Warn("SubmitWork", "SubmitWork err", "can`t get engine by version", "version", string(tmpHead.Version))
		//	return false
		//}
		//if err := engine.VerifySeal(a.chain, tmpHead); err != nil {
		//	log.Warn("SubmitWork", "Invalid pow work submitted", hash, "err", err)
		//	return false
		//}
		a.returnCh <- &consensus.SealResult{consensus.SealTypePow, tmpHead}
		a.workid = 0
		log.Info("YYYYYYYYYSubmitWorkYYYYYYYYYYY", "information", "pow ok send head")
		delete(a.work, hash)
	}
	//block := work.Block.WithSeal(result)

	//delete(a.work, hash)

	return true
}

/*func conversType(str string) (types.BlockNonce, error) {
	noncebig, _ := new(big.Int).SetString(strings.TrimPrefix(str, "0x"), 16)
	retByte := reversebyBytes(noncebig.Bytes())
	noncebig2 := new(big.Int).SetBytes(retByte)
	blockNonce := types.EncodeNonce(noncebig2.Uint64())
	return blockNonce, nil
}*/
func reverseToNonce(s []byte) (types.BlockNonce, error) {
	var ret types.BlockNonce

	if 4 != len(s) {
		//log.Error("reverseToNonce, nonce length != 4")
		return types.BlockNonce{}, errors.New("reverseToNonce, nonce length != 4")
	}

	for i := 0; i < len(s); i++ {
		ret[4+i] = s[len(s)-1-i]
	}
	return ret, nil
}

// loop monitors mining events on the work and quit channels, updating the internal
// state of the remote miner until a termination is requested.
//
// Note, the reason the work and quit channels are passed as parameters is because
// RemoteAgent.Start() constantly recreates these channels, so the loop code cannot
// assume data stability in these member fields.
func (a *RemoteAgent) loop(workCh chan *Work, quitCh chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-quitCh:
			return
		case work := <-workCh:
			a.mu.Lock()
			a.currentWork = work
			a.mu.Unlock()
		case <-ticker.C:
			// cleanup
			a.mu.Lock()
			for hash, work := range a.work {
				log.Info("5s loop", "hash", hash, "createdAt", work.createdAt, "elapsed time", time.Since(work.createdAt))
				if time.Since(work.createdAt) > 7*(12*time.Second) {
					delete(a.work, hash)
				}
			}
			a.mu.Unlock()

			a.hashrateMu.Lock()
			for id, hashrate := range a.hashrate {
				if time.Since(hashrate.ping) > 10*time.Second {
					delete(a.hashrate, id)
				}
			}
			a.hashrateMu.Unlock()
		}
	}
}
