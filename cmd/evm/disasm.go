// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/matrix/go-matrix/core/asm"
	cli "gopkg.in/urfave/cli.v1"
)

var disasmCommand = cli.Command{
	Action:    disasmCmd,
	Name:      "disasm",
	Usage:     "disassembles evm binary",
	ArgsUsage: "<file>",
}

func disasmCmd(ctx *cli.Context) error {
	if len(ctx.Args().First()) == 0 {
		return errors.New("filename required")
	}

	fn := ctx.Args().First()
	in, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	code := strings.TrimSpace(string(in[:]))
	fmt.Printf("%v\n", code)
	return asm.PrintDisassembled(code)
}
