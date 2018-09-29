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

// Contains the tests associated with the Whisper protocol Envelope object.

package whisperv6

import (
	mrand "math/rand"
	"testing"

	"github.com/matrix/go-matrix/crypto"
)

func TestEnvelopeOpenAcceptsOnlyOneKeyTypeInFilter(t *testing.T) {
	symKey := make([]byte, aesKeyLength)
	mrand.Read(symKey)

	asymKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed GenerateKey with seed %d: %s.", seed, err)
	}

	params := MessageParams{
		PoW:      0.01,
		WorkTime: 1,
		TTL:      uint32(mrand.Intn(1024)),
		Payload:  make([]byte, 50),
		KeySym:   symKey,
		Dst:      nil,
	}

	mrand.Read(params.Payload)

	msg, err := NewSentMessage(&params)
	if err != nil {
		t.Fatalf("failed to create new message with seed %d: %s.", seed, err)
	}

	e, err := msg.Wrap(&params)
	if err != nil {
		t.Fatalf("Failed to Wrap the message in an envelope with seed %d: %s", seed, err)
	}

	f := Filter{KeySym: symKey, KeyAsym: asymKey}

	decrypted := e.Open(&f)
	if decrypted != nil {
		t.Fatalf("Managed to decrypt a message with an invalid filter, seed %d", seed)
	}
}
