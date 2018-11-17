// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package storage

// this is a clone of an earlier state of the matrix mandb/database
// no need for queueing/caching

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const openFileLimit = 128

type LDBDatabase struct {
	db *leveldb.DB
}

func NewLDBDatabase(file string) (*LDBDatabase, error) {
	// Open the db
	db, err := leveldb.OpenFile(file, &opt.Options{OpenFilesCacheCapacity: openFileLimit})
	if err != nil {
		return nil, err
	}

	database := &LDBDatabase{db: db}

	return database, nil
}

func (self *LDBDatabase) Put(key []byte, value []byte) {
	err := self.db.Put(key, value, nil)
	if err != nil {
		fmt.Println("Error put", err)
	}
}

func (self *LDBDatabase) Get(key []byte) ([]byte, error) {
	dat, err := self.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

func (self *LDBDatabase) Delete(key []byte) error {
	return self.db.Delete(key, nil)
}

func (self *LDBDatabase) LastKnownTD() []byte {
	data, _ := self.Get([]byte("LTD"))

	if len(data) == 0 {
		data = []byte{0x0}
	}

	return data
}

func (self *LDBDatabase) NewIterator() iterator.Iterator {
	return self.db.NewIterator(nil, nil)
}

func (self *LDBDatabase) Write(batch *leveldb.Batch) error {
	return self.db.Write(batch, nil)
}

func (self *LDBDatabase) Close() {
	// Close the leveldb database
	self.db.Close()
}
