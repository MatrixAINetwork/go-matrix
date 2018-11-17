// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package api

import (
	"reflect"
	"testing"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/crypto"
)

func TestConfig(t *testing.T) {

	var hexprvkey = "65138b2aa745041b372153550584587da326ab440576b2a1191dd95cee30039c"

	prvkey, err := crypto.HexToECDSA(hexprvkey)
	if err != nil {
		t.Fatalf("failed to load private key: %v", err)
	}

	one := NewDefaultConfig()
	two := NewDefaultConfig()

	if equal := reflect.DeepEqual(one, two); !equal {
		t.Fatal("Two default configs are not equal")
	}

	one.Init(prvkey)

	//the init function should set the following fields
	if one.BzzKey == "" {
		t.Fatal("Expected BzzKey to be set")
	}
	if one.PublicKey == "" {
		t.Fatal("Expected PublicKey to be set")
	}

	//the Init function should append subdirs to the given path
	if one.Swap.PayProfile.Beneficiary == (common.Address{}) {
		t.Fatal("Failed to correctly initialize SwapParams")
	}

	if one.SyncParams.RequestDbPath == one.Path {
		t.Fatal("Failed to correctly initialize SyncParams")
	}

	if one.HiveParams.KadDbPath == one.Path {
		t.Fatal("Failed to correctly initialize HiveParams")
	}

	if one.StoreParams.ChunkDbPath == one.Path {
		t.Fatal("Failed to correctly initialize StoreParams")
	}
}
