// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package rpc_test

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/MatrixAINetwork/go-matrix/rpc"
)

// In this example, our client whishes to track the latest 'block number'
// known to the server. The server supports two methods:
//
// man_getBlockByNumber("latest", {})
//    returns the latest block object.
//
// man_subscribe("newBlocks")
//    creates a subscription which fires block objects when new blocks arrive.

type Block struct {
	Number *big.Int
}

func ExampleClientSubscription() {
	// Connect the client.
	client, _ := rpc.Dial("ws://127.0.0.1:8485")
	subch := make(chan Block)

	// Ensure that subch receives the latest block.
	go func() {
		for i := 0; ; i++ {
			if i > 0 {
				time.Sleep(2 * time.Second)
			}
			subscribeBlocks(client, subch)
		}
	}()

	// Print events from the subscription as they arrive.
	for block := range subch {
		fmt.Println("latest block:", block.Number)
	}
}

// subscribeBlocks runs in its own goroutine and maintains
// a subscription for new blocks.
func subscribeBlocks(client *rpc.Client, subch chan Block) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Subscribe to new blocks.
	sub, err := client.ManSubscribe(ctx, subch, "newBlocks")
	if err != nil {
		fmt.Println("subscribe error:", err)
		return
	}

	// The connection is established now.
	// Update the channel with the current block.
	var lastBlock Block
	if err := client.CallContext(ctx, &lastBlock, "man_getBlockByNumber", "latest"); err != nil {
		fmt.Println("can't get latest block:", err)
		return
	}
	subch <- lastBlock

	// The subscription will deliver events to the channel. Wait for the
	// subscription to end for any reason, then loop around to re-establish
	// the connection.
	fmt.Println("connection lost: ", <-sub.Err())
}
