// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package downloader

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/rawdb"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mandb"
)

// FakePeer is a mock downloader peer that operates on a local database instance
// instead of being an actual live node. It's useful for testing and to implement
// sync commands from an existing local database.
type FakePeer struct {
	id string
	db mandb.Database
	hc *core.HeaderChain
	dl *Downloader
}

// NewFakePeer creates a new mock downloader peer with the given data sources.
func NewFakePeer(id string, db mandb.Database, hc *core.HeaderChain, dl *Downloader) *FakePeer {
	return &FakePeer{id: id, db: db, hc: hc, dl: dl}
}

// Head implements downloader.Peer, returning the current head hash and number
// of the best known header.
func (p *FakePeer) Head() (common.Hash, *big.Int, uint64, uint64, uint64, uint64) {
	header := p.hc.CurrentHeader()
	return header.Hash(), header.Number, header.SuperBlockSeq(), 0, 0, 0
}

// RequestHeadersByHash implements downloader.Peer, returning a batch of headers
// defined by the origin hash and the associated query parameters.
func (p *FakePeer) RequestHeadersByHash(hash common.Hash, amount int, skip int, reverse bool) error {
	var (
		headers []*types.Header
		unknown bool
	)
	for !unknown && len(headers) < amount {
		origin := p.hc.GetHeaderByHash(hash)
		if origin == nil {
			break
		}
		number := origin.Number.Uint64()
		headers = append(headers, origin)
		if reverse {
			for i := 0; i <= skip; i++ {
				if header := p.hc.GetHeader(hash, number); header != nil {
					hash = header.ParentHash
					number--
				} else {
					unknown = true
					break
				}
			}
		} else {
			var (
				current = origin.Number.Uint64()
				next    = current + uint64(skip) + 1
			)
			if header := p.hc.GetHeaderByNumber(next); header != nil {
				if p.hc.GetBlockHashesFromHash(header.Hash(), uint64(skip+1))[skip] == hash {
					hash = header.Hash()
				} else {
					unknown = true
				}
			} else {
				unknown = true
			}
		}
	}
	p.dl.DeliverHeaders(p.id, headers)
	return nil
}

// RequestHeadersByNumber implements downloader.Peer, returning a batch of headers
// defined by the origin number and the associated query parameters.
func (p *FakePeer) RequestHeadersByNumber(number uint64, amount int, skip int, reverse bool) error {
	var (
		headers []*types.Header
		unknown bool
	)
	for !unknown && len(headers) < amount {
		origin := p.hc.GetHeaderByNumber(number)
		if origin == nil {
			break
		}
		if reverse {
			if number >= uint64(skip+1) {
				number -= uint64(skip + 1)
			} else {
				unknown = true
			}
		} else {
			number += uint64(skip + 1)
		}
		headers = append(headers, origin)
	}
	p.dl.DeliverHeaders(p.id, headers)
	return nil
}

// RequestBodies implements downloader.Peer, returning a batch of block bodies
// corresponding to the specified block hashes.
func (p *FakePeer) RequestBodies(hashes []common.Hash) error {
	var (
		txs    [][]types.CurrencyBlock
		uncles [][]*types.Header
	)
	for _, hash := range hashes {
		block := rawdb.ReadBlock(p.db, hash, *p.hc.GetBlockNumber(hash))
		txs = append(txs, block.Currencies())
		uncles = append(uncles, block.Uncles())
	}
	p.dl.DeliverBodies(p.id, txs, uncles)
	return nil
}

// RequestReceipts implements downloader.Peer, returning a batch of transaction
// receipts corresponding to the specified block hashes.
func (p *FakePeer) RequestReceipts(hashes []common.Hash) error {
	var receipts [][]types.CoinReceipts
	for _, hash := range hashes {
		receipts = append(receipts, rawdb.ReadReceipts(p.db, hash, *p.hc.GetBlockNumber(hash)))
	}
	p.dl.DeliverReceipts(p.id, receipts)
	return nil
}

// RequestNodeData implements downloader.Peer, returning a batch of state trie
// nodes corresponding to the specified trie hashes.
func (p *FakePeer) RequestNodeData(hashes []common.Hash) error {
	var data [][]byte
	for _, hash := range hashes {
		if entry, err := p.db.Get(hash.Bytes()); err == nil {
			data = append(data, entry)
		}
	}
	p.dl.DeliverNodeData(p.id, data)
	return nil
}
