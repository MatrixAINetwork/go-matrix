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
package fuse

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

func handleMountFusefsStderr(errCh chan<- error) func(line string) (ignore bool) {
	return func(line string) (ignore bool) {
		const (
			noMountpointPrefix = `mount_fusefs: `
			noMountpointSuffix = `: No such file or directory`
		)
		if strings.HasPrefix(line, noMountpointPrefix) && strings.HasSuffix(line, noMountpointSuffix) {
			// re-extract it from the error message in case some layer
			// changed the path
			mountpoint := line[len(noMountpointPrefix) : len(line)-len(noMountpointSuffix)]
			err := &MountpointDoesNotExistError{
				Path: mountpoint,
			}
			select {
			case errCh <- err:
				return true
			default:
				// not the first error; fall back to logging it
				return false
			}
		}

		return false
	}
}

// isBoringMountFusefsError returns whether the Wait error is
// uninteresting; exit status 1 is.
func isBoringMountFusefsError(err error) bool {
	if err, ok := err.(*exec.ExitError); ok && err.Exited() {
		if status, ok := err.Sys().(syscall.WaitStatus); ok && status.ExitStatus() == 1 {
			return true
		}
	}
	return false
}

func mount(dir string, conf *mountConfig, ready chan<- struct{}, errp *error) (*os.File, error) {
	for k, v := range conf.options {
		if strings.Contains(k, ",") || strings.Contains(v, ",") {
			// Silly limitation but the mount helper does not
			// understand any escaping. See TestMountOptionCommaError.
			return nil, fmt.Errorf("mount options cannot contain commas on FreeBSD: %q=%q", k, v)
		}
	}

	f, err := os.OpenFile("/dev/fuse", os.O_RDWR, 0000)
	if err != nil {
		*errp = err
		return nil, err
	}

	cmd := exec.Command(
		"/sbin/mount_fusefs",
		"--safe",
		"-o", conf.getOptions(),
		"3",
		dir,
	)
	cmd.ExtraFiles = []*os.File{f}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("setting up mount_fusefs stderr: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("setting up mount_fusefs stderr: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("mount_fusefs: %v", err)
	}
	helperErrCh := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(2)
	go lineLogger(&wg, "mount helper output", neverIgnoreLine, stdout)
	go lineLogger(&wg, "mount helper error", handleMountFusefsStderr(helperErrCh), stderr)
	wg.Wait()
	if err := cmd.Wait(); err != nil {
		// see if we have a better error to report
		select {
		case helperErr := <-helperErrCh:
			// log the Wait error if it's not what we expected
			if !isBoringMountFusefsError(err) {
				log.Printf("mount helper failed: %v", err)
			}
			// and now return what we grabbed from stderr as the real
			// error
			return nil, helperErr
		default:
			// nope, fall back to generic message
		}
		return nil, fmt.Errorf("mount_fusefs: %v", err)
	}

	close(ready)
	return f, nil
}
