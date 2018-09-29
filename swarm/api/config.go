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

package api

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"path/filepath"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/contracts/ens"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/node"
	"github.com/matrix/go-matrix/swarm/network"
	"github.com/matrix/go-matrix/swarm/services/swap"
	"github.com/matrix/go-matrix/swarm/storage"
)

const (
	DefaultHTTPListenAddr = "127.0.0.1"
	DefaultHTTPPort       = "8500"
)

// separate bzz directories
// allow several bzz nodes running in parallel
type Config struct {
	// serialised/persisted fields
	*storage.StoreParams
	*storage.ChunkerParams
	*network.HiveParams
	Swap *swap.SwapParams
	*network.SyncParams
	Contract    common.Address
	EnsRoot     common.Address
	EnsAPIs     []string
	Path        string
	ListenAddr  string
	Port        string
	PublicKey   string
	BzzKey      string
	NetworkId   uint64
	SwapEnabled bool
	SyncEnabled bool
	SwapApi     string
	Cors        string
	BzzAccount  string
	BootNodes   string
}

//create a default config with all parameters to set to defaults
func NewDefaultConfig() (self *Config) {

	self = &Config{
		StoreParams:   storage.NewDefaultStoreParams(),
		ChunkerParams: storage.NewChunkerParams(),
		HiveParams:    network.NewDefaultHiveParams(),
		SyncParams:    network.NewDefaultSyncParams(),
		Swap:          swap.NewDefaultSwapParams(),
		ListenAddr:    DefaultHTTPListenAddr,
		Port:          DefaultHTTPPort,
		Path:          node.DefaultDataDir(),
		EnsAPIs:       nil,
		EnsRoot:       ens.TestNetAddress,
		NetworkId:     network.NetworkId,
		SwapEnabled:   false,
		SyncEnabled:   true,
		SwapApi:       "",
		BootNodes:     "",
	}

	return
}

//some config params need to be initialized after the complete
//config building phase is completed (e.g. due to overriding flags)
func (self *Config) Init(prvKey *ecdsa.PrivateKey) {

	address := crypto.PubkeyToAddress(prvKey.PublicKey)
	self.Path = filepath.Join(self.Path, "bzz-"+common.Bytes2Hex(address.Bytes()))
	err := os.MkdirAll(self.Path, os.ModePerm)
	if err != nil {
		log.Error(fmt.Sprintf("Error creating root swarm data directory: %v", err))
		return
	}

	pubkey := crypto.FromECDSAPub(&prvKey.PublicKey)
	pubkeyhex := common.ToHex(pubkey)
	keyhex := crypto.Keccak256Hash(pubkey).Hex()

	self.PublicKey = pubkeyhex
	self.BzzKey = keyhex

	self.Swap.Init(self.Contract, prvKey)
	self.SyncParams.Init(self.Path)
	self.HiveParams.Init(self.Path)
	self.StoreParams.Init(self.Path)
}
