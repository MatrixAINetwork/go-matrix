// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
//

package storage

import (
	"fmt"
)

type Storage interface {
	// Put stores a value by key. 0-length keys results in no-op
	Put(key, value string)
	// Get returns the previously stored value, or the empty string if it does not exist or key is of 0-length
	Get(key string) string
}

// EphemeralStorage is an in-memory storage that does
// not persist values to disk. Mainly used for testing
type EphemeralStorage struct {
	data      map[string]string
	namespace string
}

func (s *EphemeralStorage) Put(key, value string) {
	if len(key) == 0 {
		return
	}
	fmt.Printf("storage: put %v -> %v\n", key, value)
	s.data[key] = value
}

func (s *EphemeralStorage) Get(key string) string {
	if len(key) == 0 {
		return ""
	}
	fmt.Printf("storage: get %v\n", key)
	if v, exist := s.data[key]; exist {
		return v
	}
	return ""
}

func NewEphemeralStorage() Storage {
	s := &EphemeralStorage{
		data: make(map[string]string),
	}
	return s
}
