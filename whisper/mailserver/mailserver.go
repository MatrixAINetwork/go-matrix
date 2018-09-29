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

package mailserver

import (
	"encoding/binary"
	"fmt"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/rlp"
	whisper "github.com/matrix/go-matrix/whisper/whisperv6"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type WMailServer struct {
	db  *leveldb.DB
	w   *whisper.Whisper
	pow float64
	key []byte
}

type DBKey struct {
	timestamp uint32
	hash      common.Hash
	raw       []byte
}

func NewDbKey(t uint32, h common.Hash) *DBKey {
	const sz = common.HashLength + 4
	var k DBKey
	k.timestamp = t
	k.hash = h
	k.raw = make([]byte, sz)
	binary.BigEndian.PutUint32(k.raw, k.timestamp)
	copy(k.raw[4:], k.hash[:])
	return &k
}

func (s *WMailServer) Init(shh *whisper.Whisper, path string, password string, pow float64) error {
	var err error
	if len(path) == 0 {
		return fmt.Errorf("DB file is not specified")
	}

	if len(password) == 0 {
		return fmt.Errorf("password is not specified")
	}

	s.db, err = leveldb.OpenFile(path, nil)
	if err != nil {
		return fmt.Errorf("open DB file: %s", err)
	}

	s.w = shh
	s.pow = pow

	MailServerKeyID, err := s.w.AddSymKeyFromPassword(password)
	if err != nil {
		return fmt.Errorf("create symmetric key: %s", err)
	}
	s.key, err = s.w.GetSymKey(MailServerKeyID)
	if err != nil {
		return fmt.Errorf("save symmetric key: %s", err)
	}
	return nil
}

func (s *WMailServer) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *WMailServer) Archive(env *whisper.Envelope) {
	key := NewDbKey(env.Expiry-env.TTL, env.Hash())
	rawEnvelope, err := rlp.EncodeToBytes(env)
	if err != nil {
		log.Error(fmt.Sprintf("rlp.EncodeToBytes failed: %s", err))
	} else {
		err = s.db.Put(key.raw, rawEnvelope, nil)
		if err != nil {
			log.Error(fmt.Sprintf("Writing to DB failed: %s", err))
		}
	}
}

func (s *WMailServer) DeliverMail(peer *whisper.Peer, request *whisper.Envelope) {
	if peer == nil {
		log.Error("Whisper peer is nil")
		return
	}

	ok, lower, upper, bloom := s.validateRequest(peer.ID(), request)
	if ok {
		s.processRequest(peer, lower, upper, bloom)
	}
}

func (s *WMailServer) processRequest(peer *whisper.Peer, lower, upper uint32, bloom []byte) []*whisper.Envelope {
	ret := make([]*whisper.Envelope, 0)
	var err error
	var zero common.Hash
	kl := NewDbKey(lower, zero)
	ku := NewDbKey(upper, zero)
	i := s.db.NewIterator(&util.Range{Start: kl.raw, Limit: ku.raw}, nil)
	defer i.Release()

	for i.Next() {
		var envelope whisper.Envelope
		err = rlp.DecodeBytes(i.Value(), &envelope)
		if err != nil {
			log.Error(fmt.Sprintf("RLP decoding failed: %s", err))
		}

		if whisper.BloomFilterMatch(bloom, envelope.Bloom()) {
			if peer == nil {
				// used for test purposes
				ret = append(ret, &envelope)
			} else {
				err = s.w.SendP2PDirect(peer, &envelope)
				if err != nil {
					log.Error(fmt.Sprintf("Failed to send direct message to peer: %s", err))
					return nil
				}
			}
		}
	}

	err = i.Error()
	if err != nil {
		log.Error(fmt.Sprintf("Level DB iterator error: %s", err))
	}

	return ret
}

func (s *WMailServer) validateRequest(peerID []byte, request *whisper.Envelope) (bool, uint32, uint32, []byte) {
	if s.pow > 0.0 && request.PoW() < s.pow {
		return false, 0, 0, nil
	}

	f := whisper.Filter{KeySym: s.key}
	decrypted := request.Open(&f)
	if decrypted == nil {
		log.Warn(fmt.Sprintf("Failed to decrypt p2p request"))
		return false, 0, 0, nil
	}

	src := crypto.FromECDSAPub(decrypted.Src)
	if len(src)-len(peerID) == 1 {
		src = src[1:]
	}

	// if you want to check the signature, you can do it here. e.g.:
	// if !bytes.Equal(peerID, src) {
	if src == nil {
		log.Warn(fmt.Sprintf("Wrong signature of p2p request"))
		return false, 0, 0, nil
	}

	var bloom []byte
	payloadSize := len(decrypted.Payload)
	if payloadSize < 8 {
		log.Warn(fmt.Sprintf("Undersized p2p request"))
		return false, 0, 0, nil
	} else if payloadSize == 8 {
		bloom = whisper.MakeFullNodeBloom()
	} else if payloadSize < 8+whisper.BloomFilterSize {
		log.Warn(fmt.Sprintf("Undersized bloom filter in p2p request"))
		return false, 0, 0, nil
	} else {
		bloom = decrypted.Payload[8 : 8+whisper.BloomFilterSize]
	}

	lower := binary.BigEndian.Uint32(decrypted.Payload[:4])
	upper := binary.BigEndian.Uint32(decrypted.Payload[4:8])
	return true, lower, upper, bloom
}
