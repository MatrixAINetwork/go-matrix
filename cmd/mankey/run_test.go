// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/docker/docker/pkg/reexec"
	"github.com/matrix/go-matrix/internal/cmdtest"
)

type testEthkey struct {
	*cmdtest.TestCmd
}

// spawns mankey with the given command line args.
func runEthkey(t *testing.T, args ...string) *testEthkey {
	tt := new(testEthkey)
	tt.TestCmd = cmdtest.NewTestCmd(t, tt)
	tt.Run("mankey-test", args...)
	return tt
}

func TestMain(m *testing.M) {
	// Run the app if we've been exec'd as "mankey-test" in runEthkey.
	reexec.Register("mankey-test", func() {
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	})
	// check if we have been reexec'd
	if reexec.Init() {
		return
	}
	os.Exit(m.Run())
}
