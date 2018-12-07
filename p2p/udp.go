// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package p2p

import (
	"net"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
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

	signAddr := make([]common.Address, 0)
	if ca.InDuration() {
		signAddr = ca.GetRolesByGroupOnlyNextElect(common.RoleValidator | common.RoleBackupValidator)
	} else {
		signAddr = ca.GetRolesByGroup(common.RoleValidator | common.RoleBackupValidator)
	}
	if len(signAddr) <= 2 {
		for _, id := range signAddr {
			send(id, bytes)
		}
		return
	}

	is := Random(len(signAddr), 2)
	for _, i := range is {
		send(signAddr[i], bytes)
	}
}

func send(address common.Address, data []byte) {
	node := ServerP2p.ntab.GetAllAddress()
	n, ok := node[address]
	if !ok {
		log.Error("can't send udp to", "addr", address)
		return
	}

	addr, err := net.ResolveUDPAddr("udp", n.IP.String()+":30000")
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
