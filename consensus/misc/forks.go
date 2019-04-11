// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package misc

import (
	"fmt"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/params"
)

// VerifyForkHashes verifies that blocks conforming to network hard-forks do have
// the correct hashes, to avoid clients going off on different chains. This is an
// optional feature.
func VerifyForkHashes(config *params.ChainConfig, header *types.Header, uncle bool) error {
	// We don't care about uncles
	if uncle {
		return nil
	}
	// If the homestead reprice hash is set, validate it
	if config.EIP150Block != nil && config.EIP150Block.Cmp(header.Number) == 0 {
		if config.EIP150Hash != (common.Hash{}) && config.EIP150Hash != header.Hash() {
			return fmt.Errorf("homestead gas reprice fork: have 0x%x, want 0x%x", header.Hash(), config.EIP150Hash)
		}
	}
	// All ok, return
	return nil
}
