// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixstate

import (
	"encoding/binary"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"github.com/pkg/errors"
)

func encodeAccount(account common.Address) ([]byte, error) {
	data, err := rlp.EncodeToBytes(account)
	if err != nil {
		return nil, errors.Errorf("rlp encode failed: %s", err)
	}
	return data, nil
}

func decodeAccount(data []byte) (common.Address, error) {
	msg := common.Address{}
	err := rlp.DecodeBytes(data, &msg)
	if err != nil {
		return common.Address{}, errors.Errorf("rlp decode failed: %s", err)
	}
	return msg, nil
}

func encodeAccounts(accounts []common.Address) ([]byte, error) {
	data, err := rlp.EncodeToBytes(accounts)
	if err != nil {
		return nil, errors.Errorf("rlp encode failed: %s", err)
	}
	return data, nil
}

func decodeAccounts(data []byte) ([]common.Address, error) {
	msg := make([]common.Address, 0)
	err := rlp.DecodeBytes(data, &msg)
	if err != nil {
		return nil, errors.Errorf("rlp decode failed: %s", err)
	}
	return msg, nil
}

func encodeString(str string) ([]byte, error) {
	data, err := rlp.EncodeToBytes(str)
	if err != nil {
		return nil, errors.Errorf("rkp encdoe failed: %s", err)
	}
	return data, nil
}

func decodeString(data []byte) (string, error) {
	var msg string
	err := rlp.DecodeBytes(data, &msg)
	if err != nil {
		return msg, errors.Errorf("rlp decode failed: %s", err)
	}
	return msg, nil
}

func encodeUint64(num uint64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, num)
	return data
}

func decodeUint64(data []byte) (uint64, error) {
	if len(data) < 8 {
		log.Error(logInfo, "decode uint64 failed", "data size is not enough", "size", len(data))
		return 0, ErrDataSize
	}
	return binary.BigEndian.Uint64(data[:8]), nil
}
