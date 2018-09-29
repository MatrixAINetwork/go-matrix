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
// Package boot :net search
package boot

import (
	"time"

	"github.com/matrix/go-matrix/core"

	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/p2p"
	"github.com/matrix/go-matrix/p2p/discover"
	"github.com/matrix/go-matrix/params"
)

//Boots HandleMessage and some channal
type Boots struct {
	blockchain *core.BlockChain

	ChanPing chan LocalPongInfo

	BootReturn REboot
	PublicSeq  uint64
	NeedAck    []uint64
	LocalID    string

	MyRecvChan    chan p2p.Custsend
	HandleMessage map[int]func(p2p.Custsend)
}

//REboot return result
type REboot struct {
	NetFlag int
	Height  uint64
	//	MainList []election.NodeInfo
}

const (
	// TimeOutLimit wait data time
	TimeOutLimit = 2 * time.Second
	//FindConnStatusInterval sleep time
	FindConnStatusInterval = 5
	// MinLiveBootCount min live boot count
	MinLiveBootCount = 12
	Module           = "BOOT"
	ChanSize         = 100
)

//CheckIPStatusAndCount Check IP Status And count
func (TBoot *Boots) CheckIPStatusAndCount(ListID []string) (int, []string) {
	count := 0
	ListIDConnStatus := TBoot.GetPingPong(ListID)
	tt := make([]string, 0)
	for i := 0; i < len(ListIDConnStatus); i++ {
		if count >= MinLiveBootCount {
			break
		}
		count++
		tt = append(tt, ListIDConnStatus[i].IP)
	}
	log.INFO(Module, "CheckIPStatusAndCount count", count, "ListIDConnStatus", ListIDConnStatus)
	return count, tt
}

func (TBoot *Boots) FindOneBootNode(ListID []string) bool {
	for {
		count, _ := TBoot.CheckIPStatusAndCount(ListID)

		if count >= MinLiveBootCount {
			return true
		}
		log.INFO(Module, "FindOneBootNode sleep end", "go on checkIPStatus")
		time.Sleep(FindConnStatusInterval * time.Second)
	}
}

//New new Boots
func New(bc *core.BlockChain, nodeID string) *Boots {
	boot := &Boots{
		blockchain: bc,

		ChanPing: make(chan LocalPongInfo, ChanSize),

		PublicSeq:     0,
		LocalID:       nodeID,
		MyRecvChan:    make(chan p2p.Custsend),
		HandleMessage: make(map[int]func(p2p.Custsend)),
	}

	boot.HandleMessage[Getpingreq] = boot.HandleGetPingPongReq
	boot.HandleMessage[Getpongrsp] = boot.HandleGetPingPongRsp
	return boot
}

//IsBootNode check if it is boot node
func (TBoot *Boots) IsBootNode(AimID string) (bool, []string) {
	LocalIPStatus := false
	needfindid := make([]string, 0)
	for _, url := range params.MainnetBootnodes {
		if bootNode, err := discover.ParseNode(url); err == nil {
			if bootNode.ID.String() == AimID {
				LocalIPStatus = true
			} else {
				needfindid = append(needfindid, bootNode.ID.String())
			}
		}
	}
	log.INFO(Module, "AimID", AimID, "LocalIPStatus", LocalIPStatus, "needfindid", needfindid)
	return LocalIPStatus, needfindid
}

//Run boot main function
func (TBoot *Boots) Run() {
	/*
		go TBoot.HandleP2PMessage()
		go TBoot.ReadRecvChanfromP2P()
		_, needfindid := TBoot.IsBootNode(TBoot.LocalID)
		TBoot.FindOneBootNode(needfindid)
	*/

	time.Sleep(3 * time.Second)
	mc.PublishEvent(mc.NewBlockMessage, TBoot.blockchain.GetBlockByNumber(0))
	log.INFO("MAIN", "创世区块插入消息已发送", TBoot.blockchain.GetBlockByNumber(0))
	log.INFO("Peer总量", "len", p2p.ServerP2p.PeerCount())

}
