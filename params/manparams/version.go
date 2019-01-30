package manparams

import (
	"bytes"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
)

const (
	VersionAlpha = "1.0.0.0"
	//增加版本号示例
	//VersionBeta          = "1.0.0.1"
	//VersionSignatureBeta = "0xc3a8b3c887e2a896cca7a3d86997ac458d4f2e1ac0472fbc37290ee131eb82400cde214d72427dcf83ad22eb5b98a269311c1589fab14d0eeeee632617714cc000"
	//VersionNumBeta       = uint64(32)
)

var VersionList [][]byte
var VersionSignatureMap map[string][]common.Signature

func init() {
	VersionList = [][]byte{[]byte(VersionAlpha) /*[]byte(VersionBeta)*/}
	VersionSignatureMap = make(map[string][]common.Signature)
	//VersionSignatureMap[VersionBeta] = []common.Signature{common.BytesToSignature(common.FromHex(VersionSignatureBeta))}
}

func IsCorrectVersion(version []byte) bool {
	if len(version) == 0 {
		return false
	}
	for _, item := range VersionList {
		if bytes.Equal(version, item) {
			return true
		}
	}
	return false
}

func GetVersionSignature(parentBlock *types.Block, version []byte) []common.Signature {
	if len(version) == 0 {
		return nil
	}
	if string(version) == string(parentBlock.Version()) {
		return parentBlock.VersionSignature()
	}
	if sig, ok := VersionSignatureMap[string(version)]; ok {
		return sig
	}

	return nil
}
