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
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/matrix/go-matrix/accounts/keystore"
	"github.com/matrix/go-matrix/cmd/utils"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/crypto"
	"gopkg.in/urfave/cli.v1"
)

type outputSign struct {
	Signature string
}

var msgfileFlag = cli.StringFlag{
	Name:  "msgfile",
	Usage: "file containing the message to sign/verify",
}

var commandSignMessage = cli.Command{
	Name:      "signmessage",
	Usage:     "sign a message",
	ArgsUsage: "<keyfile> <message>",
	Description: `
Sign the message with a keyfile.

To sign a message contained in a file, use the --msgfile flag.
`,
	Flags: []cli.Flag{
		passphraseFlag,
		jsonFlag,
		msgfileFlag,
	},
	Action: func(ctx *cli.Context) error {
		message := getMessage(ctx, 1)

		// Load the keyfile.
		keyfilepath := ctx.Args().First()
		keyjson, err := ioutil.ReadFile(keyfilepath)
		if err != nil {
			utils.Fatalf("Failed to read the keyfile at '%s': %v", keyfilepath, err)
		}

		// Decrypt key with passphrase.
		passphrase := getPassPhrase(ctx, false)
		key, err := keystore.DecryptKey(keyjson, passphrase)
		if err != nil {
			utils.Fatalf("Error decrypting key: %v", err)
		}

		signature, err := crypto.Sign(signHash(message), key.PrivateKey)
		if err != nil {
			utils.Fatalf("Failed to sign message: %v", err)
		}
		out := outputSign{Signature: hex.EncodeToString(signature)}
		if ctx.Bool(jsonFlag.Name) {
			mustPrintJSON(out)
		} else {
			fmt.Println("Signature:", out.Signature)
		}
		return nil
	},
}

type outputVerify struct {
	Success            bool
	RecoveredAddress   string
	RecoveredPublicKey string
}

var commandVerifyMessage = cli.Command{
	Name:      "verifymessage",
	Usage:     "verify the signature of a signed message",
	ArgsUsage: "<address> <signature> <message>",
	Description: `
Verify the signature of the message.
It is possible to refer to a file containing the message.`,
	Flags: []cli.Flag{
		jsonFlag,
		msgfileFlag,
	},
	Action: func(ctx *cli.Context) error {
		addressStr := ctx.Args().First()
		signatureHex := ctx.Args().Get(1)
		message := getMessage(ctx, 2)

		if !common.IsHexAddress(addressStr) {
			utils.Fatalf("Invalid address: %s", addressStr)
		}
		address := common.HexToAddress(addressStr)
		signature, err := hex.DecodeString(signatureHex)
		if err != nil {
			utils.Fatalf("Signature encoding is not hexadecimal: %v", err)
		}

		recoveredPubkey, err := crypto.SigToPub(signHash(message), signature)
		if err != nil || recoveredPubkey == nil {
			utils.Fatalf("Signature verification failed: %v", err)
		}
		recoveredPubkeyBytes := crypto.FromECDSAPub(recoveredPubkey)
		recoveredAddress := crypto.PubkeyToAddress(*recoveredPubkey)
		success := address == recoveredAddress

		out := outputVerify{
			Success:            success,
			RecoveredPublicKey: hex.EncodeToString(recoveredPubkeyBytes),
			RecoveredAddress:   recoveredAddress.Hex(),
		}
		if ctx.Bool(jsonFlag.Name) {
			mustPrintJSON(out)
		} else {
			if out.Success {
				fmt.Println("Signature verification successful!")
			} else {
				fmt.Println("Signature verification failed!")
			}
			fmt.Println("Recovered public key:", out.RecoveredPublicKey)
			fmt.Println("Recovered address:", out.RecoveredAddress)
		}
		return nil
	},
}

func getMessage(ctx *cli.Context, msgarg int) []byte {
	if file := ctx.String("msgfile"); file != "" {
		if len(ctx.Args()) > msgarg {
			utils.Fatalf("Can't use --msgfile and message argument at the same time.")
		}
		msg, err := ioutil.ReadFile(file)
		if err != nil {
			utils.Fatalf("Can't read message file: %v", err)
		}
		return msg
	} else if len(ctx.Args()) == msgarg+1 {
		return []byte(ctx.Args().Get(msgarg))
	}
	utils.Fatalf("Invalid number of arguments: want %d, got %d", msgarg+1, len(ctx.Args()))
	return nil
}
