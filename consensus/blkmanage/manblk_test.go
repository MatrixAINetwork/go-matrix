// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkmanage

import (
	"testing"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/params/manparams"
)

func TestManBlkBasePlug_Prepare(t *testing.T) {
	test, _ := New(nil)
	base, _ := NewBlkBasePlug()
	test.RegisterManBLkPlugs("common", manparams.VersionAlpha, base)

	test.Prepare("common", manparams.VersionAlpha, 0, nil, common.Hash{1})
}
