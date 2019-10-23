package main

import (
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus/amhash"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"math/big"
)

func main() {
	aiMineEngine := amhash.New(amhash.Config{PowMode: amhash.ModeNormal, PictureStorePath: ""})

	mineHeader := &types.Header{
		Number:     big.NewInt(0),
		ParentHash: common.HexToHash("0x0123456789"), // 挖矿hash
		Difficulty: big.NewInt(200),
		Coinbase:   common.HexToAddress("0xabcdef"),
	}

	stopCh := make(chan struct{})
	resultChan := make(chan *types.Header, 2)
	go running(resultChan)

	powResult, err := aiMineEngine.SealPow(nil, mineHeader, stopCh, resultChan, false)
	if err != nil {
		fmt.Printf("err := %v", err)
		return
	}

	fmt.Printf("result x11 nonce := %d\n", powResult.Nonce.Uint64())
	fmt.Printf("result sm3 nonce := %d\n", powResult.Sm3Nonce.Uint64())
}

func running(resultChain chan *types.Header) {
	for {
		select {
		case <-resultChain:
		}
	}
}
