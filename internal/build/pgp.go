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

// signFile reads the contents of an input file and signs it (in armored format)
// with the key provided, placing the signature into the output file.

package build

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/crypto/openpgp"
)

// PGPSignFile parses a PGP private key from the specified string and creates a
// signature file into the output parameter of the input file.
//
// Note, this method assumes a single key will be container in the pgpkey arg,
// furthermore that it is in armored format.
func PGPSignFile(input string, output string, pgpkey string) error {
	// Parse the keyring and make sure we only have a single private key in it
	keys, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(pgpkey))
	if err != nil {
		return err
	}
	if len(keys) != 1 {
		return fmt.Errorf("key count mismatch: have %d, want %d", len(keys), 1)
	}
	// Create the input and output streams for signing
	in, err := os.Open(input)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()

	// Generate the signature and return
	return openpgp.ArmoredDetachSign(out, keys[0], in, nil)
}

// PGPKeyID parses an armored key and returns the key ID.
func PGPKeyID(pgpkey string) (string, error) {
	keys, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(pgpkey))
	if err != nil {
		return "", err
	}
	if len(keys) != 1 {
		return "", fmt.Errorf("key count mismatch: have %d, want %d", len(keys), 1)
	}
	return keys[0].PrimaryKey.KeyIdString(), nil
}
