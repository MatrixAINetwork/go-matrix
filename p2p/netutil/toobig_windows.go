// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

//+build windows

package netutil

import (
	"net"
	"os"
	"syscall"
)

const _WSAEMSGSIZE = syscall.Errno(10040)

// isPacketTooBig reports whether err indicates that a UDP packet didn't
// fit the receive buffer. On Windows, WSARecvFrom returns
// code WSAEMSGSIZE and no data if this happens.
func isPacketTooBig(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		if scErr, ok := opErr.Err.(*os.SyscallError); ok {
			return scErr.Err == _WSAEMSGSIZE
		}
		return opErr.Err == _WSAEMSGSIZE
	}
	return false
}
