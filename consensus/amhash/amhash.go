// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Package amhash implements the amhash proof-of-work consensus engine.
package amhash

import (
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/metrics"
	"github.com/MatrixAINetwork/go-matrix/rpc"
	"math/big"
	"math/rand"
	"sync"
)

var (
	// maxUint256 is a big integer representing 2^256-1
	maxUint256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))
)

// Mode defines the type and amount of PoW verification an amhash engine makes.
type Mode uint

const (
	ModeNormal Mode = iota
)

// Config are the configuration parameters of the amhash.
type Config struct {
	PowMode          Mode
	PictureStorePath string
}

// Amhash is a consensus engine based on proot-of-work implementing the amhash
// algorithm.
type Amhash struct {
	config Config

	// Mining related fields
	rand     *rand.Rand    // Properly seeded random source for nonces
	threads  int           // Number of threads to mine on if mining
	update   chan struct{} // Notification channel to update mining parameters
	hashrate metrics.Meter // Meter tracking the average hashrate

	lock sync.Mutex // Ensures thread safety for the in-memory caches and mining fields
}

// New creates a full sized amhash PoW scheme.
func New(config Config) *Amhash {
	return &Amhash{
		config:   config,
		update:   make(chan struct{}),
		hashrate: metrics.NewMeter(),
	}
}

// Threads returns the number of mining threads currently enabled. This doesn't
// necessarily mean that mining is running!
func (amhash *Amhash) Threads() int {
	amhash.lock.Lock()
	defer amhash.lock.Unlock()
	return amhash.threads
}

// SetThreads updates the number of mining threads currently enabled. Calling
// this method does not start mining, only sets the thread count. If zero is
// specified, the miner will use all cores of the machine. Setting a thread
// count below zero is allowed and will cause the miner to idle, without any
// work being done.
func (amhash *Amhash) SetThreads(threads int) {
	amhash.lock.Lock()
	defer amhash.lock.Unlock()

	// Update the threads and ping any running seal to pull in any changes
	amhash.threads = threads
	select {
	case amhash.update <- struct{}{}:
	default:
	}
}

// Hashrate implements PoW, returning the measured rate of the search invocations
// per second over the last minute.
func (amhash *Amhash) Hashrate() float64 {
	return amhash.hashrate.Rate1()
}

// APIs implements consensus.Engine, returning the user facing RPC APIs. Currently
// that is empty.
func (amhash *Amhash) APIs(chain consensus.ChainReader) []rpc.API {
	return nil
}
