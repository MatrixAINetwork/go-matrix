// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// +build VERIFY_EVM_INTEGER_POOL

package vm

import "fmt"

const verifyPool = true

func verifyIntegerPool(ip *intPool) {
	for i, item := range ip.pool.data {
		if item.Cmp(checkVal) != 0 {
			panic(fmt.Sprintf("%d'th item failed aggressive pool check. Value was modified", i))
		}
	}
}
