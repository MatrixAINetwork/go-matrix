// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// +build !linux,!darwin,!freebsd

package fuse

import (
	"errors"
)

var errNoFUSE = errors.New("FUSE is not supported on this platform")

func isFUSEUnsupportedError(err error) bool {
	return err == errNoFUSE
}

type MountInfo struct {
	MountPoint     string
	StartManifest  string
	LatestManifest string
}

func (self *SwarmFS) Mount(mhash, mountpoint string) (*MountInfo, error) {
	return nil, errNoFUSE
}

func (self *SwarmFS) Unmount(mountpoint string) (bool, error) {
	return false, errNoFUSE
}

func (self *SwarmFS) Listmounts() ([]*MountInfo, error) {
	return nil, errNoFUSE
}

func (self *SwarmFS) Stop() error {
	return nil
}
