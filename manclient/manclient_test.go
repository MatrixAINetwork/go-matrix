// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package manclient

import "github.com/MatrixAINetwork/go-matrix"

// Verify that Client implements the matrix interfaces.
var (
	_ = matrix.ChainReader(&Client{})
	_ = matrix.TransactionReader(&Client{})
	_ = matrix.ChainStateReader(&Client{})
	_ = matrix.ChainSyncReader(&Client{})
	_ = matrix.ContractCaller(&Client{})
	_ = matrix.GasEstimator(&Client{})
	_ = matrix.GasPricer(&Client{})
	_ = matrix.LogFilterer(&Client{})
	_ = matrix.PendingStateReader(&Client{})
	// _ = matrix.PendingStateEventer(&Client{})
	_ = matrix.PendingContractCaller(&Client{})
)
