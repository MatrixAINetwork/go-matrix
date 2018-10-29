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
package random

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/event"
)

type ElectionSeed struct {
	randomSeedReqCh  chan *mc.RandomRequest
	randomSeedReqSub event.Subscription
	msgcenter        *mc.Center
}

func newElectionSeed(msgcenter *mc.Center) (*ElectionSeed, error) {
	electionSeed := &ElectionSeed{
		randomSeedReqCh: make(chan *mc.RandomRequest, 10),
		msgcenter:       msgcenter,
	}
	err := electionSeed.initSubscribeEvent()
	if err != nil {
		return nil, err
	}
	go electionSeed.update()
	return electionSeed, nil
}

func (self *ElectionSeed) initSubscribeEvent() error {
	var err error

	self.randomSeedReqSub, err = mc.SubscribeEvent(mc.ReElec_TopoSeedReq, self.randomSeedReqCh)

	if err != nil {
		return err
	}
	log.INFO(ModuleSeed, "订阅成功", "nil")
	return nil
}

func (self *ElectionSeed) update() {
	defer func() {
		if self.randomSeedReqSub != nil {

			self.randomSeedReqSub.Unsubscribe()
		}
	}()

	for {
		select {
		case randomdata := <-self.randomSeedReqCh:
			log.INFO(ModuleSeed, "randomdata", randomdata)
			self.randomSeedReqHandle(randomdata)
		}
	}
}
func (self *ElectionSeed) randomSeedReqHandle(data *mc.RandomRequest) error {

	ans := compareMap(data.PrivateMap, data.PublicMap)
	ans.Add(ans, data.MinHash.Big())

	err := mc.PublishEvent(mc.Random_TopoSeedRsp, &mc.ElectionEvent{Seed: ans})
	if err != nil {
		log.WARN(ModuleSeed, "Random_TopoSeedRsp:err", err)
	}
	log.INFO(ModuleSeed, "Random_TopoSeedRsp Seed", ans)
	return nil
}

func compareMap(private map[common.Address][]byte, public map[common.Address][]byte) *big.Int {
	if len(private) > len(public) {
		return rangePrivate(private, public)
	}
	ans := rangePublic(private, public)
	log.INFO(ModuleSeed, "隨機數map匹配的公私鑰 data", ans)
	return ans
}

func rangePrivate(privateMap map[common.Address][]byte, publicMap map[common.Address][]byte) *big.Int {
	ans := big.NewInt(0)
	for address, privateV := range privateMap {
		publicV, ok := publicMap[address]
		if false == ok {
			continue
		}
		if compare(privateV, publicV) {
			anst := common.BytesToHash(privateV).Big()
			ans.Add(ans, anst)
		}
	}
	return ans

}
func rangePublic(privateMap map[common.Address][]byte, publicMap map[common.Address][]byte) *big.Int {
	ans := big.NewInt(0)
	for adress, publicV := range publicMap {
		PrivateV, ok := privateMap[adress]

		if false == ok {
			continue
		}
		if compare(PrivateV, publicV) {
			anst := common.BytesToHash(PrivateV).Big()
			ans.Add(ans, anst)
		}
	}
	return ans
}
func compare(private []byte, public []byte) bool {
	curve := btcec.S256()
	pk1, err := btcec.ParsePubKey(public, curve)
	if err != nil {
		return false
	}

	pk1_1 := (*ecdsa.PublicKey)(pk1)
	xx, yy := pk1_1.Curve.ScalarBaseMult(private)
	if xx.Cmp(pk1_1.X) != 0 {
		return false
	}
	if yy.Cmp(pk1_1.Y) != 0 {
		return false
	}
	return true
}
