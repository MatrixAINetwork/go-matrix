// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package boot

import (
	"fmt"
	"testing"
	"time"

	"github.com/matrix/go-matrix/p2p"

	"github.com/matrix/go-matrix/core"
)

func TestNewBoot(t *testing.T) {
	go p2p.Receiveudp()
	go p2p.CustSend()
	var bc *core.BlockChain
	ans := New(bc, "asdddd")
	ans.Run()

	time.Sleep(100 * time.Second)

	fmt.Println(ans)
}
