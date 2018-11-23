// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package compiler

import (
	"errors"
	"fmt"

	"github.com/matrix/go-matrix/core/asm"
)

func Compile(fn string, src []byte, debug bool) (string, error) {
	compiler := asm.NewCompiler(debug)
	compiler.Feed(asm.Lex(fn, src, debug))

	bin, compileErrors := compiler.Compile()
	if len(compileErrors) > 0 {
		// report errors
		for _, err := range compileErrors {
			fmt.Printf("%s:%v\n", fn, err)
		}
		return "", errors.New("compiling failed")
	}
	return bin, nil
}
