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
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

func handleFusermountStderr(errCh chan<- error) func(line string) (ignore bool) {
	return func(line string) (ignore bool) {
		if line == `fusermount: failed to open /etc/fuse.conf: Permission denied` {
			// Silence this particular message, it occurs way too
			// commonly and isn't very relevant to whether the mount
			// succeeds or not.
			return true
		}

		const (
			noMountpointPrefix = `fusermount: failed to access mountpoint `
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

// isBoringFusermountError returns whether the Wait error is
// uninteresting; exit status 1 is.
func isBoringFusermountError(err error) bool {
	if err, ok := err.(*exec.ExitError); ok && err.Exited() {
		if status, ok := err.Sys().(syscall.WaitStatus); ok && status.ExitStatus() == 1 {
			return true
		}
	}
	return false
}

func mount(dir string, conf *mountConfig, ready chan<- struct{}, errp *error) (fusefd *os.File, err error) {
	// linux mount is never delayed
	close(ready)

	fds, err := syscall.Socketpair(syscall.AF_FILE, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, fmt.Errorf("socketpair error: %v", err)
	}

	writeFile := os.NewFile(uintptr(fds[0]), "fusermount-child-writes")
	defer writeFile.Close()

	readFile := os.NewFile(uintptr(fds[1]), "fusermount-parent-reads")
	defer readFile.Close()

	cmd := exec.Command(
		"fusermount",
		"-o", conf.getOptions(),
		"--",
		dir,
	)
	cmd.Env = append(os.Environ(), "_FUSE_COMMFD=3")

	cmd.ExtraFiles = []*os.File{writeFile}

	var wg sync.WaitGroup
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("setting up fusermount stderr: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("setting up fusermount stderr: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("fusermount: %v", err)
	}
	helperErrCh := make(chan error, 1)
	wg.Add(2)
	go lineLogger(&wg, "mount helper output", neverIgnoreLine, stdout)
	go lineLogger(&wg, "mount helper error", handleFusermountStderr(helperErrCh), stderr)
	wg.Wait()
	if err := cmd.Wait(); err != nil {
		// see if we have a better error to report
		select {
		case helperErr := <-helperErrCh:
			// log the Wait error if it's not what we expected
			if !isBoringFusermountError(err) {
				log.Printf("mount helper failed: %v", err)
			}
			// and now return what we grabbed from stderr as the real
			// error
			return nil, helperErr
		default:
			// nope, fall back to generic message
		}

		return nil, fmt.Errorf("fusermount: %v", err)
	}

	c, err := net.FileConn(readFile)
	if err != nil {
		return nil, fmt.Errorf("FileConn from fusermount socket: %v", err)
	}
	defer c.Close()

	uc, ok := c.(*net.UnixConn)
	if !ok {
		return nil, fmt.Errorf("unexpected FileConn type; expected UnixConn, got %T", c)
	}

	buf := make([]byte, 32) // expect 1 byte
	oob := make([]byte, 32) // expect 24 bytes
	_, oobn, _, _, err := uc.ReadMsgUnix(buf, oob)
	scms, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return nil, fmt.Errorf("ParseSocketControlMessage: %v", err)
	}
	if len(scms) != 1 {
		return nil, fmt.Errorf("expected 1 SocketControlMessage; got scms = %#v", scms)
	}
	scm := scms[0]
	gotFds, err := syscall.ParseUnixRights(&scm)
	if err != nil {
		return nil, fmt.Errorf("syscall.ParseUnixRights: %v", err)
	}
	if len(gotFds) != 1 {
		return nil, fmt.Errorf("wanted 1 fd; got %#v", gotFds)
	}
	f := os.NewFile(uintptr(gotFds[0]), "/dev/fuse")
	return f, nil
}
