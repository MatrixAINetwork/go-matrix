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
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package packet

import (
	"crypto"
	"encoding/binary"
	"golang.org/x/crypto/openpgp/errors"
	"golang.org/x/crypto/openpgp/s2k"
	"io"
	"strconv"
)

// OnePassSignature represents a one-pass signature packet. See RFC 4880,
// section 5.4.
type OnePassSignature struct {
	SigType    SignatureType
	Hash       crypto.Hash
	PubKeyAlgo PublicKeyAlgorithm
	KeyId      uint64
	IsLast     bool
}

const onePassSignatureVersion = 3

func (ops *OnePassSignature) parse(r io.Reader) (err error) {
	var buf [13]byte

	_, err = readFull(r, buf[:])
	if err != nil {
		return
	}
	if buf[0] != onePassSignatureVersion {
		err = errors.UnsupportedError("one-pass-signature packet version " + strconv.Itoa(int(buf[0])))
	}

	var ok bool
	ops.Hash, ok = s2k.HashIdToHash(buf[2])
	if !ok {
		return errors.UnsupportedError("hash function: " + strconv.Itoa(int(buf[2])))
	}

	ops.SigType = SignatureType(buf[1])
	ops.PubKeyAlgo = PublicKeyAlgorithm(buf[3])
	ops.KeyId = binary.BigEndian.Uint64(buf[4:12])
	ops.IsLast = buf[12] != 0
	return
}

// Serialize marshals the given OnePassSignature to w.
func (ops *OnePassSignature) Serialize(w io.Writer) error {
	var buf [13]byte
	buf[0] = onePassSignatureVersion
	buf[1] = uint8(ops.SigType)
	var ok bool
	buf[2], ok = s2k.HashToHashId(ops.Hash)
	if !ok {
		return errors.UnsupportedError("hash type: " + strconv.Itoa(int(ops.Hash)))
	}
	buf[3] = uint8(ops.PubKeyAlgo)
	binary.BigEndian.PutUint64(buf[4:12], ops.KeyId)
	if ops.IsLast {
		buf[12] = 1
	}

	if err := serializeHeader(w, packetTypeOnePassSignature, len(buf)); err != nil {
		return err
	}
	_, err := w.Write(buf[:])
	return err
}
