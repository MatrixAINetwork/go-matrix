// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package manash

import (
	crand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"runtime"
	"sync"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
)

type diffiList []*big.Int

func (v diffiList) Len() int {
	return len(v)
}

func (v diffiList) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v diffiList) Less(i, j int) bool {
	if v[i].Cmp(v[j]) == 1 {

		return true
	}
	return false
}

type minerDifficultyList struct {
	lock      sync.RWMutex
	diffiList []*big.Int
	targets   []*big.Int
}

func GetdifficultyListAndTargetList(difficultyList []*big.Int) minerDifficultyList {
	difficultyListAndTargetList := minerDifficultyList{
		diffiList: make([]*big.Int, len(difficultyList)),
		targets:   make([]*big.Int, len(difficultyList)),
		lock:      sync.RWMutex{},
	}
	copy(difficultyListAndTargetList.diffiList, difficultyList)
	var targets = make([]*big.Int, len(difficultyList))
	for i := 0; i < len(difficultyList); i++ {
		targets[i] = new(big.Int).Div(maxUint256, difficultyList[i])

	}
	copy(difficultyListAndTargetList.targets, targets)

	return difficultyListAndTargetList
}

// Seal implements consensus.Engine, attempting to find a nonce that satisfies
// the block's difficulty requirements.

func (manash *Manash) Seal(chain consensus.ChainReader, header *types.Header, stop <-chan struct{}, isBroadcastNode bool) (*types.Header, error) {
	log.INFO("seal", "挖矿", "开始", "高度", header.Number.Uint64())
	defer log.INFO("seal", "挖矿", "结束", "高度", header.Number.Uint64())

	// Create a runner and the multiple search threads it directs
	abort := make(chan struct{})
	found := make(chan *types.Header)
	manash.lock.Lock()
	curHeader := types.CopyHeader(header)
	threads := manash.threads
	if manash.rand == nil {
		seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			manash.lock.Unlock()
			return nil, err
		}
		manash.rand = rand.New(rand.NewSource(seed.Int64()))
	}
	manash.lock.Unlock()

	threads = runtime.NumCPU()
	if isBroadcastNode {
		threads = 1
	}

	var pend sync.WaitGroup
	for i := 0; i < threads; i++ {
		pend.Add(1)
		go func(id int, nonce uint64) {
			defer pend.Done()
			manash.mine(curHeader, id, nonce, abort, found, isBroadcastNode)

		}(i, uint64(manash.rand.Int63()))
	}
	// Wait until sealing is terminated or a nonce is found
	var result *types.Header
	select {
	case <-stop:
		log.INFO("SEALER", "Sealer receive stop mine, curHeader", curHeader.HashNoSignsAndNonce().TerminalString())
		// Outside abort, stop all miner threads
		close(abort)
	case result = <-found:
		// One of the threads found a block, abort all others
		close(abort)
	case <-manash.update:
		// Thread count was changed on user request, restart
		close(abort)
		pend.Wait()
		return manash.Seal(chain, curHeader, stop, isBroadcastNode)
	}

	// Wait for all miners to terminate and return the block
	pend.Wait()
	return result, nil
}

func compareDifflist(result []byte, diffList []*big.Int, targets []*big.Int) (int, bool) {
	for i := 0; i < len(diffList); i++ {
		if new(big.Int).SetBytes(result).Cmp(targets[i]) <= 0 {
			return i, true
		}
	}

	return -1, false
}

// mine is the actual proof-of-work miner that searches for a nonce starting from
// seed that results in correct final block difficulty.
func (manash *Manash) mine(header *types.Header, id int, seed uint64, abort chan struct{}, found chan *types.Header, isBroadcastNode bool) {
	// Extract some data from the header
	var (
		curHeader = types.CopyHeader(header)
		hash      = curHeader.HashNoNonce().Bytes()
		target    = new(big.Int).Div(maxUint256, header.Difficulty)
		number    = curHeader.Number.Uint64()
		dataset   = manash.dataset(number)
	)
	if isBroadcastNode {
		target = maxUint256
	}
	// Start generating random nonces until we abort or find a good one
	log.INFO("SEALER begin mine", "target", target, "isBroadcast", isBroadcastNode, "number", curHeader.Number.Uint64(), "diff", header.Difficulty.Uint64())
	defer log.INFO("SEALER stop mine", "number", curHeader.Number.Uint64(), "diff", header.Difficulty.Uint64())
	var (
		attempts = int64(0)
		nonce    = seed
	)
	logger := log.New("miner", id)
	logger.Trace("Started manash search for new nonces", "seed", seed)
	//log.INFO("SEALER", "Started manash search for new nonces seed", seed)
search:
	for {
		select {
		case <-abort:
			// Mining terminated, update stats and abort
			logger.Trace("Manash nonce search aborted", "attempts", nonce-seed)
			manash.hashrate.Mark(attempts)
			return

		default:
			// We don't have to update hash rate on every nonce, so update after after 2^X nonces
			attempts++
			if (attempts % (1 << 15)) == 0 {
				manash.hashrate.Mark(attempts)
				attempts = 0
			}
			// Compute the PoW value of this nonce
			digest, result := hashimotoFull(dataset.dataset, hash, nonce)

			//log.Info("sealer","result",new(big.Int).SetBytes(result))
			//log.Info("sealer","target",target)
			if new(big.Int).SetBytes(result).Cmp(target) <= 0 {
				// Correct nonce found, create a new header with it
				header = types.CopyHeader(header)
				header.Nonce = types.EncodeNonce(nonce)
				header.MixDigest = common.BytesToHash(digest)

				// Seal and return a block (if still needed)
				select {
				case found <- header:
					logger.Trace("Manash nonce found and reported", "attempts", nonce-seed, "nonce", nonce)
				case <-abort:
					logger.Trace("Manash nonce found but discarded", "attempts", nonce-seed, "nonce", nonce)
				}
				break search
			}
			nonce++
		}
	}
	// Datasets are unmapped in a finalizer. Ensure that the dataset stays live
	// during sealing so it's not unmapped while being read.
	runtime.KeepAlive(dataset)
}
