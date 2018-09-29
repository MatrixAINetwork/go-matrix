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
package p2p

import (
	"net"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/p2p/discover"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/rlp"
)

func UdpStart() {
	addr, err := net.ResolveUDPAddr("udp", ":30000")
	if err != nil {
		log.Error("Can't resolve address: ", "p2p udp", err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Error("Error listening:", "p2p udp", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, params.MaxUdpBuf)

	for {
		var mxtxs []*types.Transaction_Mx
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Error("UDP read error", "err", err)
			return
		}

		err = rlp.DecodeBytes(buf[:n], &mxtxs)
		if err != nil {
			log.Error("rlp decode error", "err", err)
			continue
		}
		mc.PublishEvent(mc.SendUdpTx, mxtxs)
	}
}

func UdpSend(data interface{}) {
	bytes, err := rlp.EncodeToBytes(data)
	if err != nil {
		log.Error("error", "p2p udp", err)
		return
	}

	ids := make([]discover.NodeID, 0)
	if ca.InDuration() {
		ids = ca.GetRolesByGroupOnlyBackup(common.RoleValidator)
	} else {
		ids = ca.GetRolesByGroup(common.RoleValidator)
	}
	if len(ids) <= 2 {
		for _, id := range ids {
			send(id, bytes)
		}
		return
	}

	is := Random(len(ids), 2)
	for _, i := range is {
		send(ids[i], bytes)
	}
}

func send(id discover.NodeID, data []byte) {
	node := ServerP2p.ntab.Resolve(id)
	if node == nil {
		log.Error("buckets nodes", "p2p", id)
		return
	}

	addr, err := net.ResolveUDPAddr("udp", node.IP.String()+":30000")
	if err != nil {
		log.Error("Can't resolve address: ", "p2p udp", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Error("Can't dial: ", "p2p udp", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		log.Error("failed:", "p2p udp", err)
		return
	}
}
