// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/tests"

	cli "gopkg.in/urfave/cli.v1"
)

var stateTestCommand = cli.Command{
	Action:    stateTestCmd,
	Name:      "statetest",
	Usage:     "executes the given state tests",
	ArgsUsage: "<file>",
}

// StatetestResult contains the execution status after running a state test, any
// error that might have occurred and a dump of the final state if requested.
type StatetestResult struct {
	Name  string      `json:"name"`
	Pass  bool        `json:"pass"`
	Fork  string      `json:"fork"`
	Error string      `json:"error,omitempty"`
	State *state.Dump `json:"state,omitempty"`
}

func stateTestCmd(ctx *cli.Context) error {
	if len(ctx.Args().First()) == 0 {
		return errors.New("path-to-test argument required")
	}
	// Configure the go-matrix logger
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.Lvl(ctx.GlobalInt(VerbosityFlag.Name)))
	log.Root().SetHandler(glogger)

	// Configure the EVM logger
	config := &vm.LogConfig{
		DisableMemory: ctx.GlobalBool(DisableMemoryFlag.Name),
		DisableStack:  ctx.GlobalBool(DisableStackFlag.Name),
	}
	var (
		tracer   vm.Tracer
		debugger *vm.StructLogger
	)
	switch {
	case ctx.GlobalBool(MachineFlag.Name):
		tracer = NewJSONLogger(config, os.Stderr)

	case ctx.GlobalBool(DebugFlag.Name):
		debugger = vm.NewStructLogger(config)
		tracer = debugger

	default:
		debugger = vm.NewStructLogger(config)
	}
	// Load the test content from the input file
	src, err := ioutil.ReadFile(ctx.Args().First())
	if err != nil {
		return err
	}
	var tests map[string]tests.StateTest
	if err = json.Unmarshal(src, &tests); err != nil {
		return err
	}
	// Iterate over all the tests, run them and aggregate the results
	cfg := vm.Config{
		Tracer: tracer,
		Debug:  ctx.GlobalBool(DebugFlag.Name) || ctx.GlobalBool(MachineFlag.Name),
	}
	results := make([]StatetestResult, 0, len(tests))
	for key, test := range tests {
		for _, st := range test.Subtests() {
			// Run the test and aggregate the result
			result := &StatetestResult{Name: key, Fork: st.Fork, Pass: true}
			state, err := test.Run(st, cfg)
			if err != nil {
				// Test failed, mark as so and dump any state to aid debugging
				result.Pass, result.Error = false, err.Error()
				if ctx.GlobalBool(DumpFlag.Name) && state != nil {
					dump := state.RawDump()
					result.State = &dump
				}
			}
			// print state root for evmlab tracing (already committed above, so no need to delete objects again
			if ctx.GlobalBool(MachineFlag.Name) && state != nil {
				fmt.Fprintf(os.Stderr, "{\"stateRoot\": \"%x\"}\n", state.IntermediateRoot(false))
			}

			results = append(results, *result)

			// Print any structured logs collected
			if ctx.GlobalBool(DebugFlag.Name) {
				if debugger != nil {
					fmt.Fprintln(os.Stderr, "#### TRACE ####")
					vm.WriteTrace(os.Stderr, debugger.StructLogs())
				}
			}
		}
	}
	out, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(out))
	return nil
}
