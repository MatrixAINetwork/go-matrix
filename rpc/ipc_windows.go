// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// +build windows

package rpc

import (
	"context"
	"net"
	"time"

	"gopkg.in/natefinch/npipe.v2"
)

// This is used if the dialing context has no deadline. It is much smaller than the
// defaultDialTimeout because named pipes are local and there is no need to wait so long.
const defaultPipeDialTimeout = 2 * time.Second

// ipcListen will create a named pipe on the given endpoint.
func ipcListen(endpoint string) (net.Listener, error) {
	return npipe.Listen(endpoint)
}

// newIPCConnection will connect to a named pipe with the given endpoint as name.
func newIPCConnection(ctx context.Context, endpoint string) (net.Conn, error) {
	timeout := defaultPipeDialTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = deadline.Sub(time.Now())
		if timeout < 0 {
			timeout = 0
		}
	}
	return npipe.DialTimeout(endpoint, timeout)
}
