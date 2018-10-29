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

// Contains all the wrappers from the go-matrix root package.

package gman

import (
	"errors"

	matrix "github.com/matrix/go-matrix"
	"github.com/matrix/go-matrix/common"
)

// Subscription represents an event subscription where events are
// delivered on a data channel.
type Subscription struct {
	sub matrix.Subscription
}

// Unsubscribe cancels the sending of events to the data channel
// and closes the error channel.
func (s *Subscription) Unsubscribe() {
	s.sub.Unsubscribe()
}

// CallMsg contains parameters for contract calls.
type CallMsg struct {
	msg matrix.CallMsg
}

// NewCallMsg creates an empty contract call parameter list.
func NewCallMsg() *CallMsg {
	return new(CallMsg)
}

func (msg *CallMsg) GetFrom() *Address    { return &Address{msg.msg.From} }
func (msg *CallMsg) GetGas() int64        { return int64(msg.msg.Gas) }
func (msg *CallMsg) GetGasPrice() *BigInt { return &BigInt{msg.msg.GasPrice} }
func (msg *CallMsg) GetValue() *BigInt    { return &BigInt{msg.msg.Value} }
func (msg *CallMsg) GetData() []byte      { return msg.msg.Data }
func (msg *CallMsg) GetTo() *Address {
	if to := msg.msg.To; to != nil {
		return &Address{*msg.msg.To}
	}
	return nil
}

func (msg *CallMsg) SetFrom(address *Address)  { msg.msg.From = address.address }
func (msg *CallMsg) SetGas(gas int64)          { msg.msg.Gas = uint64(gas) }
func (msg *CallMsg) SetGasPrice(price *BigInt) { msg.msg.GasPrice = price.bigint }
func (msg *CallMsg) SetValue(value *BigInt)    { msg.msg.Value = value.bigint }
func (msg *CallMsg) SetData(data []byte)       { msg.msg.Data = common.CopyBytes(data) }
func (msg *CallMsg) SetTo(address *Address) {
	if address == nil {
		msg.msg.To = nil
	}
	msg.msg.To = &address.address
}

// SyncProgress gives progress indications when the node is synchronising with
// the Matrix network.
type SyncProgress struct {
	progress matrix.SyncProgress
}

func (p *SyncProgress) GetStartingBlock() int64 { return int64(p.progress.StartingBlock) }
func (p *SyncProgress) GetCurrentBlock() int64  { return int64(p.progress.CurrentBlock) }
func (p *SyncProgress) GetHighestBlock() int64  { return int64(p.progress.HighestBlock) }
func (p *SyncProgress) GetPulledStates() int64  { return int64(p.progress.PulledStates) }
func (p *SyncProgress) GetKnownStates() int64   { return int64(p.progress.KnownStates) }

// Topics is a set of topic lists to filter events with.
type Topics struct{ topics [][]common.Hash }

// NewTopics creates a slice of uninitialized Topics.
func NewTopics(size int) *Topics {
	return &Topics{
		topics: make([][]common.Hash, size),
	}
}

// NewTopicsEmpty creates an empty slice of Topics values.
func NewTopicsEmpty() *Topics {
	return NewTopics(0)
}

// Size returns the number of topic lists inside the set
func (t *Topics) Size() int {
	return len(t.topics)
}

// Get returns the topic list at the given index from the slice.
func (t *Topics) Get(index int) (hashes *Hashes, _ error) {
	if index < 0 || index >= len(t.topics) {
		return nil, errors.New("index out of bounds")
	}
	return &Hashes{t.topics[index]}, nil
}

// Set sets the topic list at the given index in the slice.
func (t *Topics) Set(index int, topics *Hashes) error {
	if index < 0 || index >= len(t.topics) {
		return errors.New("index out of bounds")
	}
	t.topics[index] = topics.hashes
	return nil
}

// Append adds a new topic list to the end of the slice.
func (t *Topics) Append(topics *Hashes) {
	t.topics = append(t.topics, topics.hashes)
}

// FilterQuery contains options for contact log filtering.
type FilterQuery struct {
	query matrix.FilterQuery
}

// NewFilterQuery creates an empty filter query for contact log filtering.
func NewFilterQuery() *FilterQuery {
	return new(FilterQuery)
}

func (fq *FilterQuery) GetFromBlock() *BigInt    { return &BigInt{fq.query.FromBlock} }
func (fq *FilterQuery) GetToBlock() *BigInt      { return &BigInt{fq.query.ToBlock} }
func (fq *FilterQuery) GetAddresses() *Addresses { return &Addresses{fq.query.Addresses} }
func (fq *FilterQuery) GetTopics() *Topics       { return &Topics{fq.query.Topics} }

func (fq *FilterQuery) SetFromBlock(fromBlock *BigInt)    { fq.query.FromBlock = fromBlock.bigint }
func (fq *FilterQuery) SetToBlock(toBlock *BigInt)        { fq.query.ToBlock = toBlock.bigint }
func (fq *FilterQuery) SetAddresses(addresses *Addresses) { fq.query.Addresses = addresses.addresses }
func (fq *FilterQuery) SetTopics(topics *Topics)          { fq.query.Topics = topics.topics }
