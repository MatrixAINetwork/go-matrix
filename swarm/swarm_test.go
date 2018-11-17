// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package swarm

import (
	"testing"

	"github.com/matrix/go-matrix/common"
)

func TestParseEnsAPIAddress(t *testing.T) {
	for _, x := range []struct {
		description string
		value       string
		tld         string
		endpoint    string
		addr        common.Address
	}{
		{
			description: "IPC endpoint",
			value:       "/data/testnet/gman.ipc",
			endpoint:    "/data/testnet/gman.ipc",
		},
		{
			description: "HTTP endpoint",
			value:       "http://127.0.0.1:1234",
			endpoint:    "http://127.0.0.1:1234",
		},
		{
			description: "WS endpoint",
			value:       "ws://127.0.0.1:1234",
			endpoint:    "ws://127.0.0.1:1234",
		},
		{
			description: "IPC Endpoint and TLD",
			value:       "test:/data/testnet/gman.ipc",
			endpoint:    "/data/testnet/gman.ipc",
			tld:         "test",
		},
		{
			description: "HTTP endpoint and TLD",
			value:       "test:http://127.0.0.1:1234",
			endpoint:    "http://127.0.0.1:1234",
			tld:         "test",
		},
		{
			description: "WS endpoint and TLD",
			value:       "test:ws://127.0.0.1:1234",
			endpoint:    "ws://127.0.0.1:1234",
			tld:         "test",
		},
		{
			description: "IPC Endpoint and contract address",
			value:       "314159265dD8dbb310642f98f50C066173C1259b@/data/testnet/gman.ipc",
			endpoint:    "/data/testnet/gman.ipc",
			addr:        common.HexToAddress("314159265dD8dbb310642f98f50C066173C1259b"),
		},
		{
			description: "HTTP endpoint and contract address",
			value:       "314159265dD8dbb310642f98f50C066173C1259b@http://127.0.0.1:1234",
			endpoint:    "http://127.0.0.1:1234",
			addr:        common.HexToAddress("314159265dD8dbb310642f98f50C066173C1259b"),
		},
		{
			description: "WS endpoint and contract address",
			value:       "314159265dD8dbb310642f98f50C066173C1259b@ws://127.0.0.1:1234",
			endpoint:    "ws://127.0.0.1:1234",
			addr:        common.HexToAddress("314159265dD8dbb310642f98f50C066173C1259b"),
		},
		{
			description: "IPC Endpoint, TLD and contract address",
			value:       "test:314159265dD8dbb310642f98f50C066173C1259b@/data/testnet/gman.ipc",
			endpoint:    "/data/testnet/gman.ipc",
			addr:        common.HexToAddress("314159265dD8dbb310642f98f50C066173C1259b"),
			tld:         "test",
		},
		{
			description: "HTTP endpoint, TLD and contract address",
			value:       "man:314159265dD8dbb310642f98f50C066173C1259b@http://127.0.0.1:1234",
			endpoint:    "http://127.0.0.1:1234",
			addr:        common.HexToAddress("314159265dD8dbb310642f98f50C066173C1259b"),
			tld:         "man",
		},
		{
			description: "WS endpoint, TLD and contract address",
			value:       "man:314159265dD8dbb310642f98f50C066173C1259b@ws://127.0.0.1:1234",
			endpoint:    "ws://127.0.0.1:1234",
			addr:        common.HexToAddress("314159265dD8dbb310642f98f50C066173C1259b"),
			tld:         "man",
		},
	} {
		t.Run(x.description, func(t *testing.T) {
			tld, endpoint, addr := parseEnsAPIAddress(x.value)
			if endpoint != x.endpoint {
				t.Errorf("expected Endpoint %q, got %q", x.endpoint, endpoint)
			}
			if addr != x.addr {
				t.Errorf("expected ContractAddress %q, got %q", x.addr.String(), addr.String())
			}
			if tld != x.tld {
				t.Errorf("expected TLD %q, got %q", x.tld, tld)
			}
		})
	}
}
