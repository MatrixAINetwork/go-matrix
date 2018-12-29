// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package enr

import (
	"crypto/ecdsa"
	"math/big"
	"testing"
)

// Checks that failure to sign leaves the record unmodified.
func TestSignError(t *testing.T) {
	invalidKey := &ecdsa.PrivateKey{D: new(big.Int), PublicKey: *pubkey}

	var r Record
	if err := SignV4(&r, invalidKey); err == nil {
		t.Fatal("expected error from SignV4")
	}
	if len(r.pairs) > 0 {
		t.Fatal("expected empty record, have", r.pairs)
	}
}
