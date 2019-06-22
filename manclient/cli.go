// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package manclient

import (
	"context"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/p2p"
	manrpc "github.com/MatrixAINetwork/go-matrix/rpc"
)

// client defines typed wrappers for the matrix RPC API.
type client struct {
	*Client
	rpc *manrpc.Client
}

// Dial connects a client to the given URL.
func DialMan(rawurl string) (ManClient, error) {
	c, err := manrpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}
	return NewManClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewManClient(rpc *manrpc.Client) ManClient {
	return &client{
		Client: NewClient(rpc),
		rpc:    rpc,
	}
}

// Close closes an existing RPC connection.
func (c *client) Close() {
	c.rpc.Close()
}

// SendRawTransaction injects a signed transaction into the pending pool for execution.
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
func (c *client) SendRawTransaction(ctx context.Context, tx *types.Transaction) error {
	return c.SendTransaction(ctx, tx)
}

// BlockNumber returns the current block number.
func (c *client) BlockNumber(ctx context.Context) (*big.Int, error) {
	var r string
	err := c.rpc.CallContext(ctx, &r, "eth_blockNumber")
	if err != nil {
		return nil, err
	}
	return hexutil.DecodeBig(r)
}

// AddPeer connects to the given nodeURL.
func (c *client) AddPeer(ctx context.Context, nodeURL string) error {
	var r bool
	return c.rpc.CallContext(ctx, &r, "admin_addPeer", nodeURL)
}

// AdminPeers returns the number of connected peers.
func (c *client) AdminPeers(ctx context.Context) ([]*p2p.PeerInfo, error) {
	var r []*p2p.PeerInfo
	return r, c.rpc.CallContext(ctx, &r, "admin_peers")
}

// NodeInfo gathers and returns a collection of metadata known about the host.
func (c *client) NodeInfo(ctx context.Context) (*p2p.PeerInfo, error) {
	var r *p2p.PeerInfo
	return r, c.rpc.CallContext(ctx, &r, "admin_nodeInfo")
}

// StartMining starts mining operation.
func (c *client) StartMining(ctx context.Context) error {
	var r []byte
	return c.rpc.CallContext(ctx, &r, "miner_start", nil)
}

// StopMining stops mining.
func (c *client) StopMining(ctx context.Context) error {
	return c.rpc.CallContext(ctx, nil, "miner_stop", nil)
}
