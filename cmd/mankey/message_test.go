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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMessageSignVerify(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "mankey-test")
	if err != nil {
		t.Fatal("Can't create temporary directory:", err)
	}
	defer os.RemoveAll(tmpdir)

	keyfile := filepath.Join(tmpdir, "the-keyfile")
	message := "test message"

	// Create the key.
	generate := runEthkey(t, "generate", keyfile)
	generate.Expect(`
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "foobar"}}
Repeat passphrase: {{.InputLine "foobar"}}
`)
	_, matches := generate.ExpectRegexp(`Address: (0x[0-9a-fA-F]{40})\n`)
	address := matches[1]
	generate.ExpectExit()

	// Sign a message.
	sign := runEthkey(t, "signmessage", keyfile, message)
	sign.Expect(`
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "foobar"}}
`)
	_, matches = sign.ExpectRegexp(`Signature: ([0-9a-f]+)\n`)
	signature := matches[1]
	sign.ExpectExit()

	// Verify the message.
	verify := runEthkey(t, "verifymessage", address, signature, message)
	_, matches = verify.ExpectRegexp(`
Signature verification successful!
Recovered public key: [0-9a-f]+
Recovered address: (0x[0-9a-fA-F]{40})
`)
	recovered := matches[1]
	verify.ExpectExit()

	if recovered != address {
		t.Error("recovered address doesn't match generated key")
	}
}
