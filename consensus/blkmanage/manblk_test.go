package blkmanage

import (
	"testing"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

func TestManBlkBasePlug_Prepare(t *testing.T) {
	test, _ := New(nil)
	base, _ := NewBlkBasePlug()
	test.RegisterManBLkPlugs("common", manparams.VersionAlpha, base)

	test.Prepare("common", manparams.VersionAlpha, 0, nil, common.Hash{1})
}
