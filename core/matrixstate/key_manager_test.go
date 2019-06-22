// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixstate

import (
	"github.com/MatrixAINetwork/go-matrix/log"
	"testing"
)

func Test_PrintKeys(t *testing.T) {
	log.InitLog(3)

	for key, opt := range mangerAlpha.operators {
		log.Info("key info", "key", key, "hash", opt.KeyHash().Hex())
	}
}
