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
