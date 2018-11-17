// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// +build linux darwin freebsd

package fuse

import (
	"bazil.org/fuse/fs"
)

var (
	_ fs.Node = (*SwarmDir)(nil)
)

type SwarmRoot struct {
	root *SwarmDir
}

func (filesystem *SwarmRoot) Root() (fs.Node, error) {
	return filesystem.root, nil
}
