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

package whisperv5

import (
	"testing"

	"github.com/matrix/go-matrix/crypto"
)

func BenchmarkDeriveKeyMaterial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deriveKeyMaterial([]byte("test"), 0)
	}
}

func BenchmarkEncryptionSym(b *testing.B) {
	InitSingleTest()

	params, err := generateMessageParams()
	if err != nil {
		b.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
	}

	for i := 0; i < b.N; i++ {
		msg, _ := NewSentMessage(params)
		_, err := msg.Wrap(params)
		if err != nil {
			b.Errorf("failed Wrap with seed %d: %s.", seed, err)
			b.Errorf("i = %d, len(msg.Raw) = %d, params.Payload = %d.", i, len(msg.Raw), len(params.Payload))
			return
		}
	}
}

func BenchmarkEncryptionAsym(b *testing.B) {
	InitSingleTest()

	params, err := generateMessageParams()
	if err != nil {
		b.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
	}
	key, err := crypto.GenerateKey()
	if err != nil {
		b.Fatalf("failed GenerateKey with seed %d: %s.", seed, err)
	}
	params.KeySym = nil
	params.Dst = &key.PublicKey

	for i := 0; i < b.N; i++ {
		msg, _ := NewSentMessage(params)
		_, err := msg.Wrap(params)
		if err != nil {
			b.Fatalf("failed Wrap with seed %d: %s.", seed, err)
		}
	}
}

func BenchmarkDecryptionSymValid(b *testing.B) {
	InitSingleTest()

	params, err := generateMessageParams()
	if err != nil {
		b.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
	}
	msg, _ := NewSentMessage(params)
	env, err := msg.Wrap(params)
	if err != nil {
		b.Fatalf("failed Wrap with seed %d: %s.", seed, err)
	}
	f := Filter{KeySym: params.KeySym}

	for i := 0; i < b.N; i++ {
		msg := env.Open(&f)
		if msg == nil {
			b.Fatalf("failed to open with seed %d.", seed)
		}
	}
}

func BenchmarkDecryptionSymInvalid(b *testing.B) {
	InitSingleTest()

	params, err := generateMessageParams()
	if err != nil {
		b.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
	}
	msg, _ := NewSentMessage(params)
	env, err := msg.Wrap(params)
	if err != nil {
		b.Fatalf("failed Wrap with seed %d: %s.", seed, err)
	}
	f := Filter{KeySym: []byte("arbitrary stuff here")}

	for i := 0; i < b.N; i++ {
		msg := env.Open(&f)
		if msg != nil {
			b.Fatalf("opened envelope with invalid key, seed: %d.", seed)
		}
	}
}

func BenchmarkDecryptionAsymValid(b *testing.B) {
	InitSingleTest()

	params, err := generateMessageParams()
	if err != nil {
		b.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
	}
	key, err := crypto.GenerateKey()
	if err != nil {
		b.Fatalf("failed GenerateKey with seed %d: %s.", seed, err)
	}
	f := Filter{KeyAsym: key}
	params.KeySym = nil
	params.Dst = &key.PublicKey
	msg, _ := NewSentMessage(params)
	env, err := msg.Wrap(params)
	if err != nil {
		b.Fatalf("failed Wrap with seed %d: %s.", seed, err)
	}

	for i := 0; i < b.N; i++ {
		msg := env.Open(&f)
		if msg == nil {
			b.Fatalf("fail to open, seed: %d.", seed)
		}
	}
}

func BenchmarkDecryptionAsymInvalid(b *testing.B) {
	InitSingleTest()

	params, err := generateMessageParams()
	if err != nil {
		b.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
	}
	key, err := crypto.GenerateKey()
	if err != nil {
		b.Fatalf("failed GenerateKey with seed %d: %s.", seed, err)
	}
	params.KeySym = nil
	params.Dst = &key.PublicKey
	msg, _ := NewSentMessage(params)
	env, err := msg.Wrap(params)
	if err != nil {
		b.Fatalf("failed Wrap with seed %d: %s.", seed, err)
	}

	key, err = crypto.GenerateKey()
	if err != nil {
		b.Fatalf("failed GenerateKey with seed %d: %s.", seed, err)
	}
	f := Filter{KeyAsym: key}

	for i := 0; i < b.N; i++ {
		msg := env.Open(&f)
		if msg != nil {
			b.Fatalf("opened envelope with invalid key, seed: %d.", seed)
		}
	}
}

func increment(x []byte) {
	for i := 0; i < len(x); i++ {
		x[i]++
		if x[i] != 0 {
			break
		}
	}
}

func BenchmarkPoW(b *testing.B) {
	InitSingleTest()

	params, err := generateMessageParams()
	if err != nil {
		b.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
	}
	params.Payload = make([]byte, 32)
	params.PoW = 10.0
	params.TTL = 1

	for i := 0; i < b.N; i++ {
		increment(params.Payload)
		msg, _ := NewSentMessage(params)
		_, err := msg.Wrap(params)
		if err != nil {
			b.Fatalf("failed Wrap with seed %d: %s.", seed, err)
		}
	}
}
