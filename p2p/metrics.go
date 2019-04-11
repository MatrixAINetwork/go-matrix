// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Contains the meters and timers used by the networking layer.

package p2p

import (
	"net"

	"github.com/MatrixAINetwork/go-matrix/metrics"
)

var (
	ingressConnectMeter = metrics.NewRegisteredMeter("p2p/InboundConnects", nil)
	ingressTrafficMeter = metrics.NewRegisteredMeter("p2p/InboundTraffic", nil)
	egressConnectMeter  = metrics.NewRegisteredMeter("p2p/OutboundConnects", nil)
	egressTrafficMeter  = metrics.NewRegisteredMeter("p2p/OutboundTraffic", nil)
)

// meteredConn is a wrapper around a network TCP connection that meters both the
// inbound and outbound network traffic.
type meteredConn struct {
	*net.TCPConn // Network connection to wrap with metering
}

// newMeteredConn creates a new metered connection, also bumping the ingress or
// egress connection meter. If the metrics system is disabled, this function
// returns the original object.
func newMeteredConn(conn net.Conn, ingress bool) net.Conn {
	// Short circuit if metrics are disabled
	if !metrics.Enabled {
		return conn
	}
	// Otherwise bump the connection counters and wrap the connection
	if ingress {
		ingressConnectMeter.Mark(1)
	} else {
		egressConnectMeter.Mark(1)
	}
	return &meteredConn{conn.(*net.TCPConn)}
}

// Read delegates a network read to the underlying connection, bumping the ingress
// traffic meter along the way.
func (c *meteredConn) Read(b []byte) (n int, err error) {
	n, err = c.TCPConn.Read(b)
	ingressTrafficMeter.Mark(int64(n))
	return
}

// Write delegates a network write to the underlying connection, bumping the
// egress traffic meter along the way.
func (c *meteredConn) Write(b []byte) (n int, err error) {
	n, err = c.TCPConn.Write(b)
	egressTrafficMeter.Mark(int64(n))
	return
}
