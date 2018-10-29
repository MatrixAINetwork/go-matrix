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

package ens

//go:generate abigen --sol contract/ENS.sol --exc contract/AbstractENS.sol:AbstractENS --pkg contract --out contract/ens.go
//go:generate abigen --sol contract/FIFSRegistrar.sol --exc contract/AbstractENS.sol:AbstractENS --pkg contract --out contract/fifsregistrar.go
//go:generate abigen --sol contract/PublicResolver.sol --exc contract/AbstractENS.sol:AbstractENS --pkg contract --out contract/publicresolver.go

import (
	"strings"

	"github.com/matrix/go-matrix/accounts/abi/bind"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/contracts/ens/contract"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
)

var (
	MainNetAddress = common.HexToAddress("0x314159265dD8dbb310642f98f50C066173C1259b")
	TestNetAddress = common.HexToAddress("0x112234455c3a32fd11230c42e7bccd4a84e02010")
)

// swarm domain name registry and resolver
type ENS struct {
	*contract.ENSSession
	contractBackend bind.ContractBackend
}

// NewENS creates a struct exposing convenient high-level operations for interacting with
// the Matrix Name Service.
func NewENS(transactOpts *bind.TransactOpts, contractAddr common.Address, contractBackend bind.ContractBackend) (*ENS, error) {
	ens, err := contract.NewENS(contractAddr, contractBackend)
	if err != nil {
		return nil, err
	}

	return &ENS{
		&contract.ENSSession{
			Contract:     ens,
			TransactOpts: *transactOpts,
		},
		contractBackend,
	}, nil
}

// DeployENS deploys an instance of the ENS nameservice, with a 'first-in, first-served' root registrar.
func DeployENS(transactOpts *bind.TransactOpts, contractBackend bind.ContractBackend) (common.Address, *ENS, error) {
	// Deploy the ENS registry.
	ensAddr, _, _, err := contract.DeployENS(transactOpts, contractBackend)
	if err != nil {
		return ensAddr, nil, err
	}

	ens, err := NewENS(transactOpts, ensAddr, contractBackend)
	if err != nil {
		return ensAddr, nil, err
	}

	// Deploy the registrar.
	regAddr, _, _, err := contract.DeployFIFSRegistrar(transactOpts, contractBackend, ensAddr, [32]byte{})
	if err != nil {
		return ensAddr, nil, err
	}
	// Set the registrar as owner of the ENS root.
	if _, err = ens.SetOwner([32]byte{}, regAddr); err != nil {
		return ensAddr, nil, err
	}

	return ensAddr, ens, nil
}

func ensParentNode(name string) (common.Hash, common.Hash) {
	parts := strings.SplitN(name, ".", 2)
	label := crypto.Keccak256Hash([]byte(parts[0]))
	if len(parts) == 1 {
		return [32]byte{}, label
	} else {
		parentNode, parentLabel := ensParentNode(parts[1])
		return crypto.Keccak256Hash(parentNode[:], parentLabel[:]), label
	}
}

func ensNode(name string) common.Hash {
	parentNode, parentLabel := ensParentNode(name)
	return crypto.Keccak256Hash(parentNode[:], parentLabel[:])
}

func (self *ENS) getResolver(node [32]byte) (*contract.PublicResolverSession, error) {
	resolverAddr, err := self.Resolver(node)
	if err != nil {
		return nil, err
	}

	resolver, err := contract.NewPublicResolver(resolverAddr, self.contractBackend)
	if err != nil {
		return nil, err
	}

	return &contract.PublicResolverSession{
		Contract:     resolver,
		TransactOpts: self.TransactOpts,
	}, nil
}

func (self *ENS) getRegistrar(node [32]byte) (*contract.FIFSRegistrarSession, error) {
	registrarAddr, err := self.Owner(node)
	if err != nil {
		return nil, err
	}

	registrar, err := contract.NewFIFSRegistrar(registrarAddr, self.contractBackend)
	if err != nil {
		return nil, err
	}

	return &contract.FIFSRegistrarSession{
		Contract:     registrar,
		TransactOpts: self.TransactOpts,
	}, nil
}

// Resolve is a non-transactional call that returns the content hash associated with a name.
func (self *ENS) Resolve(name string) (common.Hash, error) {
	node := ensNode(name)

	resolver, err := self.getResolver(node)
	if err != nil {
		return common.Hash{}, err
	}

	ret, err := resolver.Content(node)
	if err != nil {
		return common.Hash{}, err
	}

	return common.BytesToHash(ret[:]), nil
}

// Register registers a new domain name for the caller, making them the owner of the new name.
// Only works if the registrar for the parent domain implements the FIFS registrar protocol.
func (self *ENS) Register(name string) (*types.Transaction, error) {
	parentNode, label := ensParentNode(name)
	registrar, err := self.getRegistrar(parentNode)
	if err != nil {
		return nil, err
	}
	return registrar.Contract.Register(&self.TransactOpts, label, self.TransactOpts.From)
}

// SetContentHash sets the content hash associated with a name. Only works if the caller
// owns the name, and the associated resolver implements a `setContent` function.
func (self *ENS) SetContentHash(name string, hash common.Hash) (*types.Transaction, error) {
	node := ensNode(name)

	resolver, err := self.getResolver(node)
	if err != nil {
		return nil, err
	}

	opts := self.TransactOpts
	opts.GasLimit = 200000
	return resolver.Contract.SetContent(&opts, node, hash)
}
