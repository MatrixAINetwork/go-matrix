// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

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
