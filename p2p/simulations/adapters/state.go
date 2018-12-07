// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package adapters

type SimStateStore struct {
	m map[string][]byte
}

func (st *SimStateStore) Load(s string) ([]byte, error) {
	return st.m[s], nil
}

func (st *SimStateStore) Save(s string, data []byte) error {
	st.m[s] = data
	return nil
}

func NewSimStateStore() *SimStateStore {
	return &SimStateStore{
		make(map[string][]byte),
	}
}
