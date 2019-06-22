// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
// +build !linux

package log

import (
	"errors"
)

func SetNetLogHandler(net, addr, tag string, fmtr Format) (Handler, error) {
	return nil, errors.New("window not support")

}