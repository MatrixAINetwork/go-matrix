// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package msgsend

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
)

// AlgorithmMsg
type AlgorithmMsg struct {
	Account common.Address
	Data    NetData
}

//NetData
type NetData struct {
	SubCode uint32
	Msg     []byte
}

type BroadcastRspMsgForMarshal struct {
	Header *types.Header
	Txs    []*types.Transaction_Mx
}
