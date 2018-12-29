// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package tests

import (
	"testing"
)

func TestRLP(t *testing.T) {
	t.Parallel()
	tm := new(testMatcher)
	tm.walk(t, rlpTestDir, func(t *testing.T, name string, test *RLPTest) {
		if err := tm.checkFailure(t, name, test.Run()); err != nil {
			t.Error(err)
		}
	})
}
