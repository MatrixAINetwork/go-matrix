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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/matrix/go-matrix/accounts/abi/bind"
	"github.com/matrix/go-matrix/common/compiler"
)

var (
	abiFlag = flag.String("abi", "", "Path to the Matrix contract ABI json to bind")
	binFlag = flag.String("bin", "", "Path to the Matrix contract bytecode (generate deploy method)")
	typFlag = flag.String("type", "", "Struct name for the binding (default = package name)")

	solFlag  = flag.String("sol", "", "Path to the Matrix contract Solidity source to build and bind")
	solcFlag = flag.String("solc", "solc", "Solidity compiler to use if source builds are requested")
	excFlag  = flag.String("exc", "", "Comma separated types to exclude from binding")

	pkgFlag  = flag.String("pkg", "", "Package name to generate the binding into")
	outFlag  = flag.String("out", "", "Output file for the generated binding (default = stdout)")
	langFlag = flag.String("lang", "go", "Destination language for the bindings (go, java, objc)")
)

func main() {
	// Parse and ensure all needed inputs are specified
	flag.Parse()

	if *abiFlag == "" && *solFlag == "" {
		fmt.Printf("No contract ABI (--abi) or Solidity source (--sol) specified\n")
		os.Exit(-1)
	} else if (*abiFlag != "" || *binFlag != "" || *typFlag != "") && *solFlag != "" {
		fmt.Printf("Contract ABI (--abi), bytecode (--bin) and type (--type) flags are mutually exclusive with the Solidity source (--sol) flag\n")
		os.Exit(-1)
	}
	if *pkgFlag == "" {
		fmt.Printf("No destination package specified (--pkg)\n")
		os.Exit(-1)
	}
	var lang bind.Lang
	switch *langFlag {
	case "go":
		lang = bind.LangGo
	case "java":
		lang = bind.LangJava
	case "objc":
		lang = bind.LangObjC
	default:
		fmt.Printf("Unsupported destination language \"%s\" (--lang)\n", *langFlag)
		os.Exit(-1)
	}
	// If the entire solidity code was specified, build and bind based on that
	var (
		abis  []string
		bins  []string
		types []string
	)
	if *solFlag != "" {
		// Generate the list of types to exclude from binding
		exclude := make(map[string]bool)
		for _, kind := range strings.Split(*excFlag, ",") {
			exclude[strings.ToLower(kind)] = true
		}
		contracts, err := compiler.CompileSolidity(*solcFlag, *solFlag)
		if err != nil {
			fmt.Printf("Failed to build Solidity contract: %v\n", err)
			os.Exit(-1)
		}
		// Gather all non-excluded contract for binding
		for name, contract := range contracts {
			if exclude[strings.ToLower(name)] {
				continue
			}
			abi, _ := json.Marshal(contract.Info.AbiDefinition) // Flatten the compiler parse
			abis = append(abis, string(abi))
			bins = append(bins, contract.Code)

			nameParts := strings.Split(name, ":")
			types = append(types, nameParts[len(nameParts)-1])
		}
	} else {
		// Otherwise load up the ABI, optional bytecode and type name from the parameters
		abi, err := ioutil.ReadFile(*abiFlag)
		if err != nil {
			fmt.Printf("Failed to read input ABI: %v\n", err)
			os.Exit(-1)
		}
		abis = append(abis, string(abi))

		bin := []byte{}
		if *binFlag != "" {
			if bin, err = ioutil.ReadFile(*binFlag); err != nil {
				fmt.Printf("Failed to read input bytecode: %v\n", err)
				os.Exit(-1)
			}
		}
		bins = append(bins, string(bin))

		kind := *typFlag
		if kind == "" {
			kind = *pkgFlag
		}
		types = append(types, kind)
	}
	// Generate the contract binding
	code, err := bind.Bind(types, abis, bins, *pkgFlag, lang)
	if err != nil {
		fmt.Printf("Failed to generate ABI binding: %v\n", err)
		os.Exit(-1)
	}
	// Either flush it out to a file or display on the standard output
	if *outFlag == "" {
		fmt.Printf("%s\n", code)
		return
	}
	if err := ioutil.WriteFile(*outFlag, []byte(code), 0600); err != nil {
		fmt.Printf("Failed to write ABI binding: %v\n", err)
		os.Exit(-1)
	}
}
