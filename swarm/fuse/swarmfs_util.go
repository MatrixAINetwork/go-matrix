// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// +build linux darwin freebsd

package fuse

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/matrix/go-matrix/log"
)

func externalUnmount(mountPoint string) error {
	ctx, cancel := context.WithTimeout(context.Background(), unmountTimeout)
	defer cancel()

	// Try generic umount.
	if err := exec.CommandContext(ctx, "umount", mountPoint).Run(); err == nil {
		return nil
	}
	// Try FUSE-specific commands if umount didn't work.
	switch runtime.GOOS {
	case "darwin":
		return exec.CommandContext(ctx, "diskutil", "umount", "force", mountPoint).Run()
	case "linux":
		return exec.CommandContext(ctx, "fusermount", "-u", mountPoint).Run()
	default:
		return fmt.Errorf("unmount: unimplemented")
	}
}

func addFileToSwarm(sf *SwarmFile, content []byte, size int) error {
	fkey, mhash, err := sf.mountInfo.swarmApi.AddFile(sf.mountInfo.LatestManifest, sf.path, sf.name, content, true)
	if err != nil {
		return err
	}

	sf.lock.Lock()
	defer sf.lock.Unlock()
	sf.key = fkey
	sf.fileSize = int64(size)

	sf.mountInfo.lock.Lock()
	defer sf.mountInfo.lock.Unlock()
	sf.mountInfo.LatestManifest = mhash

	log.Info("Added new file:", "fname", sf.name, "New Manifest hash", mhash)
	return nil
}

func removeFileFromSwarm(sf *SwarmFile) error {
	mkey, err := sf.mountInfo.swarmApi.RemoveFile(sf.mountInfo.LatestManifest, sf.path, sf.name, true)
	if err != nil {
		return err
	}

	sf.mountInfo.lock.Lock()
	defer sf.mountInfo.lock.Unlock()
	sf.mountInfo.LatestManifest = mkey

	log.Info("Removed file:", "fname", sf.name, "New Manifest hash", mkey)
	return nil
}

func removeDirectoryFromSwarm(sd *SwarmDir) error {
	if len(sd.directories) == 0 && len(sd.files) == 0 {
		return nil
	}

	for _, d := range sd.directories {
		err := removeDirectoryFromSwarm(d)
		if err != nil {
			return err
		}
	}

	for _, f := range sd.files {
		err := removeFileFromSwarm(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func appendToExistingFileInSwarm(sf *SwarmFile, content []byte, offset int64, length int64) error {
	fkey, mhash, err := sf.mountInfo.swarmApi.AppendFile(sf.mountInfo.LatestManifest, sf.path, sf.name, sf.fileSize, content, sf.key, offset, length, true)
	if err != nil {
		return err
	}

	sf.lock.Lock()
	defer sf.lock.Unlock()
	sf.key = fkey
	sf.fileSize = sf.fileSize + int64(len(content))

	sf.mountInfo.lock.Lock()
	defer sf.mountInfo.lock.Unlock()
	sf.mountInfo.LatestManifest = mhash

	log.Info("Appended file:", "fname", sf.name, "New Manifest hash", mhash)
	return nil
}
