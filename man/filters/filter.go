// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package filters

import (
	"context"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/bloombits"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/rpc"
)

type Backend interface {
	ChainDb() mandb.Database
	EventMux() *event.TypeMux
	HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error)
	GetReceipts(ctx context.Context, blockHash common.Hash) ([]types.CoinReceipts, error)
	GetLogs(ctx context.Context, blockHash common.Hash) ([]types.CoinLogs, error)

	SubscribeNewTxsEvent(chan core.NewTxsEvent) event.Subscription //Y
	SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription
	SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription
	SubscribeLogsEvent(ch chan<- []types.CoinLogs) event.Subscription

	BloomStatus() (uint64, uint64)
	ServiceFilter(ctx context.Context, session *bloombits.MatcherSession)
}

// Filter can be used to retrieve and filter logs.
type Filter struct {
	backend Backend

	db         mandb.Database
	begin, end int64
	addresses  []common.Address
	topics     [][]common.Hash

	matcher *bloombits.Matcher
}

// New creates a new filter which uses a bloom filter on blocks to figure out whether
// a particular block is interesting or not.
func New(backend Backend, begin, end int64, addresses []common.Address, topics [][]common.Hash) *Filter {
	// Flatten the address and topic filter clauses into a single bloombits filter
	// system. Since the bloombits are not positional, nil topics are permitted,
	// which get flattened into a nil byte slice.
	var filters [][][]byte
	if len(addresses) > 0 {
		filter := make([][]byte, len(addresses))
		for i, address := range addresses {
			filter[i] = address.Bytes()
		}
		filters = append(filters, filter)
	}
	for _, topicList := range topics {
		filter := make([][]byte, len(topicList))
		for i, topic := range topicList {
			filter[i] = topic.Bytes()
		}
		filters = append(filters, filter)
	}
	// Assemble and return the filter
	size, _ := backend.BloomStatus()

	return &Filter{
		backend:   backend,
		begin:     begin,
		end:       end,
		addresses: addresses,
		topics:    topics,
		db:        backend.ChainDb(),
		matcher:   bloombits.NewMatcher(size, filters),
	}
}

// Logs searches the blockchain for matching log entries, returning all from the
// first block that contains matches, updating the start of the filter accordingly.
func (f *Filter) Logs(ctx context.Context) ([]types.CoinLogs, error) {
	// Figure out the limits of the filter range
	header, _ := f.backend.HeaderByNumber(ctx, rpc.LatestBlockNumber)
	if header == nil {
		return nil, nil
	}
	head := header.Number.Uint64()

	if f.begin == -1 {
		f.begin = int64(head)
	}
	end := uint64(f.end)
	if f.end == -1 {
		end = head
	}
	// Gather all indexed logs, and finish with non indexed ones
	var (
		logs []types.CoinLogs
		err  error
	)
	size, sections := f.backend.BloomStatus()
	if indexed := sections * size; indexed > uint64(f.begin) {
		if indexed > end {
			logs, err = f.indexedLogs(ctx, end)
		} else {
			logs, err = f.indexedLogs(ctx, indexed-1)
		}
		if err != nil {
			return logs, err
		}
	}
	rest, err := f.unindexedLogs(ctx, end)
	logs = append(logs, rest...)
	return logs, err
}

// indexedLogs returns the logs matching the filter criteria based on the bloom
// bits indexed available locally or via the network.
func (f *Filter) indexedLogs(ctx context.Context, end uint64) ([]types.CoinLogs, error) {
	// Create a matcher session and request servicing from the backend
	matches := make(chan uint64, 64)

	session, err := f.matcher.Start(ctx, uint64(f.begin), end, matches)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	f.backend.ServiceFilter(ctx, session)

	// Iterate over the matches until exhausted or context closed
	var logs []types.CoinLogs

	for {
		select {
		case number, ok := <-matches:
			// Abort if all matches have been fulfilled
			if !ok {
				err := session.Error()
				if err == nil {
					f.begin = int64(end) + 1
				}
				return logs, err
			}
			f.begin = int64(number) + 1

			// Retrieve the suggested block and pull any truly matching logs
			header, err := f.backend.HeaderByNumber(ctx, rpc.BlockNumber(number))
			if header == nil || err != nil {
				return logs, err
			}
			found, err := f.checkMatches(ctx, header)
			if err != nil {
				return logs, err
			}
			logs = append(logs, found...)

		case <-ctx.Done():
			return logs, ctx.Err()
		}
	}
}

// indexedLogs returns the logs matching the filter criteria based on raw block
// iteration and bloom matching.
func (f *Filter) unindexedLogs(ctx context.Context, end uint64) ([]types.CoinLogs, error) {
	var logs []types.CoinLogs

	for ; f.begin <= int64(end); f.begin++ {
		header, err := f.backend.HeaderByNumber(ctx, rpc.BlockNumber(f.begin))
		if header == nil || err != nil {
			return logs, err
		}
		for _, cr := range header.Roots {
			if bloomFilter(cr.Bloom, f.addresses, f.topics) {
				found, err := f.checkMatches(ctx, header)
				if err != nil {
					return logs, err
				}
				logs = append(logs, found...)
			}
		}

	}
	return logs, nil
}

// checkMatches checks if the receipts belonging to the given header contain any log events that
// match the filter criteria. This function is called when the bloom filter signals a potential match.
func (f *Filter) checkMatches(ctx context.Context, header *types.Header) (logs []types.CoinLogs, err error) {
	// Get the logs of the block
	logsList, err := f.backend.GetLogs(ctx, header.Hash())
	if err != nil {
		return nil, err
	}
	var unfiltered []types.CoinLogs
	for _, logs := range logsList {
		unfiltered = append(unfiltered, logs)
	}
	logs = filterLogs(unfiltered, nil, nil, f.addresses, f.topics)
	for _, ls := range logs {
		if len(ls.Logs) > 0 {
			// We have matching logs, check if we need to resolve full logs via the light client
			if ls.Logs[0].TxHash == (common.Hash{}) {
				receipts, err := f.backend.GetReceipts(ctx, header.Hash())
				if err != nil {
					return nil, err
				}
				unfiltered = unfiltered[:0]
				for _, receipt := range receipts {
					if receipt.CoinType == ls.CoinType {
						for _, r := range receipt.Receiptlist {
							unfiltered = append(unfiltered, types.CoinLogs{ls.CoinType, r.Logs})
						}
					}
				}
			}

		}
	}
	logs = filterLogs(unfiltered, nil, nil, f.addresses, f.topics)
	return logs, nil
}

func includes(addresses []common.Address, a common.Address) bool {
	for _, addr := range addresses {
		if addr == a {
			return true
		}
	}

	return false
}

// filterLogs creates a slice of logs matching the given criteria.
func filterLogs(logs []types.CoinLogs, fromBlock, toBlock *big.Int, addresses []common.Address, topics [][]common.Hash) (ret []types.CoinLogs) {
	for _, l := range logs {
		var r []*types.Log
	Logs:
		for _, log := range l.Logs {
			if fromBlock != nil && fromBlock.Int64() >= 0 && fromBlock.Uint64() > log.BlockNumber {
				continue

			}
			if toBlock != nil && toBlock.Int64() >= 0 && toBlock.Uint64() < log.BlockNumber {
				continue
			}

			if len(addresses) > 0 && !includes(addresses, log.Address) {
				continue
			}
			// If the to filtered topics is greater than the amount of topics in logs, skip.
			if len(topics) > len(log.Topics) {
				continue Logs
			}
			for i, topics := range topics {
				match := len(topics) == 0 // empty rule set == wildcard
				for _, topic := range topics {
					if log.Topics[i] == topic {
						match = true
						break
					}
				}
				if !match {
					continue Logs
				}
			}
			r = append(r, log)

		}
		ret = append(ret, types.CoinLogs{l.CoinType, r})
	}
	return
}

func bloomFilter(bloom types.Bloom, addresses []common.Address, topics [][]common.Hash) bool {
	if len(addresses) > 0 {
		var included bool
		for _, addr := range addresses {
			if types.BloomLookup(bloom, addr) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, sub := range topics {
		included := len(sub) == 0 // empty rule set == wildcard
		for _, topic := range sub {
			if types.BloomLookup(bloom, topic) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}
	return true
}
