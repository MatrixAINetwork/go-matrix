// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

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
