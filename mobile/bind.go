// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Contains all the wrappers from the bind package.

package gman

import (
	"math/big"
	"strings"

	"github.com/matrix/go-matrix/accounts/abi"
	"github.com/matrix/go-matrix/accounts/abi/bind"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
)

// Signer is an interaface defining the callback when a contract requires a
// method to sign the transaction before submission.
type Signer interface {
	Sign(*Address, *Transaction) (tx *Transaction, _ error)
}

type signer struct {
	sign bind.SignerFn
}

func (s *signer) Sign(addr *Address, unsignedTx *Transaction) (signedTx *Transaction, _ error) {
	sig, err := s.sign(types.HomesteadSigner{}, addr.address, unsignedTx.tx)
	if err != nil {
		return nil, err
	}
	return &Transaction{sig}, nil
}

// CallOpts is the collection of options to fine tune a contract call request.
type CallOpts struct {
	opts bind.CallOpts
}

// NewCallOpts creates a new option set for contract calls.
func NewCallOpts() *CallOpts {
	return new(CallOpts)
}

func (opts *CallOpts) IsPending() bool    { return opts.opts.Pending }
func (opts *CallOpts) GetGasLimit() int64 { return 0 /* TODO(karalabe) */ }

// GetContext cannot be reliably implemented without identity preservation (https://github.com/golang/go/issues/16876)
// Even then it's awkward to unpack the subtleties of a Go context out to Java.
// func (opts *CallOpts) GetContext() *Context { return &Context{opts.opts.Context} }

func (opts *CallOpts) SetPending(pending bool)     { opts.opts.Pending = pending }
func (opts *CallOpts) SetGasLimit(limit int64)     { /* TODO(karalabe) */ }
func (opts *CallOpts) SetContext(context *Context) { opts.opts.Context = context.context }

// TransactOpts is the collection of authorization data required to create a
// valid Matrix transaction.
type TransactOpts struct {
	opts bind.TransactOpts
}

func (opts *TransactOpts) GetFrom() *Address    { return &Address{opts.opts.From} }
func (opts *TransactOpts) GetNonce() int64      { return opts.opts.Nonce.Int64() }
func (opts *TransactOpts) GetValue() *BigInt    { return &BigInt{opts.opts.Value} }
func (opts *TransactOpts) GetGasPrice() *BigInt { return &BigInt{opts.opts.GasPrice} }
func (opts *TransactOpts) GetGasLimit() int64   { return int64(opts.opts.GasLimit) }

// GetSigner cannot be reliably implemented without identity preservation (https://github.com/golang/go/issues/16876)
// func (opts *TransactOpts) GetSigner() Signer { return &signer{opts.opts.Signer} }

// GetContext cannot be reliably implemented without identity preservation (https://github.com/golang/go/issues/16876)
// Even then it's awkward to unpack the subtleties of a Go context out to Java.
//func (opts *TransactOpts) GetContext() *Context { return &Context{opts.opts.Context} }

func (opts *TransactOpts) SetFrom(from *Address) { opts.opts.From = from.address }
func (opts *TransactOpts) SetNonce(nonce int64)  { opts.opts.Nonce = big.NewInt(nonce) }
func (opts *TransactOpts) SetSigner(s Signer) {
	opts.opts.Signer = func(signer types.Signer, addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
		sig, err := s.Sign(&Address{addr}, &Transaction{tx})
		if err != nil {
			return nil, err
		}
		return sig.tx, nil
	}
}
func (opts *TransactOpts) SetValue(value *BigInt)      { opts.opts.Value = value.bigint }
func (opts *TransactOpts) SetGasPrice(price *BigInt)   { opts.opts.GasPrice = price.bigint }
func (opts *TransactOpts) SetGasLimit(limit int64)     { opts.opts.GasLimit = uint64(limit) }
func (opts *TransactOpts) SetContext(context *Context) { opts.opts.Context = context.context }

// BoundContract is the base wrapper object that reflects a contract on the
// Matrix network. It contains a collection of methods that are used by the
// higher level contract bindings to operate.
type BoundContract struct {
	contract *bind.BoundContract
	address  common.Address
	deployer *types.Transaction
}

// DeployContract deploys a contract onto the Matrix blockchain and binds the
// deployment address with a wrapper.
func DeployContract(opts *TransactOpts, abiJSON string, bytecode []byte, client *MatrixClient, args *Interfaces) (contract *BoundContract, _ error) {
	// Deploy the contract to the network
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}
	addr, tx, bound, err := bind.DeployContract(&opts.opts, parsed, common.CopyBytes(bytecode), client.client, args.objects...)
	if err != nil {
		return nil, err
	}
	return &BoundContract{
		contract: bound,
		address:  addr,
		deployer: tx,
	}, nil
}

// BindContract creates a low level contract interface through which calls and
// transactions may be made through.
func BindContract(address *Address, abiJSON string, client *MatrixClient) (contract *BoundContract, _ error) {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}
	return &BoundContract{
		contract: bind.NewBoundContract(address.address, parsed, client.client, client.client, client.client),
		address:  address.address,
	}, nil
}

func (c *BoundContract) GetAddress() *Address { return &Address{c.address} }
func (c *BoundContract) GetDeployer() *Transaction {
	if c.deployer == nil {
		return nil
	}
	return &Transaction{c.deployer}
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result.
func (c *BoundContract) Call(opts *CallOpts, out *Interfaces, method string, args *Interfaces) error {
	if len(out.objects) == 1 {
		result := out.objects[0]
		if err := c.contract.Call(&opts.opts, result, method, args.objects...); err != nil {
			return err
		}
		out.objects[0] = result
	} else {
		results := make([]interface{}, len(out.objects))
		copy(results, out.objects)
		if err := c.contract.Call(&opts.opts, &results, method, args.objects...); err != nil {
			return err
		}
		copy(out.objects, results)
	}
	return nil
}

// Transact invokes the (paid) contract method with params as input values.
func (c *BoundContract) Transact(opts *TransactOpts, method string, args *Interfaces) (tx *Transaction, _ error) {
	rawTx, err := c.contract.Transact(&opts.opts, method, args.objects...)
	if err != nil {
		return nil, err
	}
	return &Transaction{rawTx}, nil
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (c *BoundContract) Transfer(opts *TransactOpts) (tx *Transaction, _ error) {
	rawTx, err := c.contract.Transfer(&opts.opts)
	if err != nil {
		return nil, err
	}
	return &Transaction{rawTx}, nil
}
