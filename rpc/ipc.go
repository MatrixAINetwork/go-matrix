// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package rpc

import (
	"context"
	"net"

	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/p2p/netutil"
)

// ServeListener accepts connections on l, serving JSON-RPC on them.
func (srv *Server) ServeListener(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if netutil.IsTemporaryError(err) {
			log.Warn("RPC accept error", "err", err)
			continue
		} else if err != nil {
			return err
		}
		log.Trace("Accepted connection", "addr", conn.RemoteAddr())
		go srv.ServeCodec(NewJSONCodec(conn), OptionMethodInvocation|OptionSubscriptions)
	}
}

// DialIPC create a new IPC client that connects to the given endpoint. On Unix it assumes
// the endpoint is the full path to a unix socket, and Windows the endpoint is an
// identifier for a named pipe.
//
// The context is used for the initial connection establishment. It does not
// affect subsequent interactions with the client.
func DialIPC(ctx context.Context, endpoint string) (*Client, error) {
	return newClient(ctx, func(ctx context.Context) (net.Conn, error) {
		return newIPCConnection(ctx, endpoint)
	})
}
