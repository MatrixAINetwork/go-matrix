// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package rawdb

import "github.com/MatrixAINetwork/go-matrix/mandb"
// DatabaseReader wraps the Has and Get method of a backing data store.
type DatabaseReader interface {
	mandb.Putter
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Close()
	NewBatch() mandb.Batch
}

// DatabaseWriter wraps the Put method of a backing data store.
type DatabaseWriter interface {
	Put(key []byte, value []byte) error
}

// DatabaseDeleter wraps the Delete method of a backing data store.
type DatabaseDeleter interface {
	Delete(key []byte) error
}
