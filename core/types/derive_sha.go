// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package types

import (
	"bytes"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"github.com/MatrixAINetwork/go-matrix/trie"
)

type DerivableList interface {
	Len() int
	GetRlp(i int) []byte
}

func DeriveSha(list DerivableList) common.Hash {
	keybuf := new(bytes.Buffer)
	trie := new(trie.Trie)
	//log.Info("DeriveSha Empty Hash", "hash", trie.Hash())
	//	log.Info("DeriveSha Trie Root Type", "Type Name",trie.Root())
	for i := 0; i < list.Len(); i++ {
		keybuf.Reset()
		rlp.Encode(keybuf, uint(i))
		trie.Update(keybuf.Bytes(), list.GetRlp(i))
	}
	log.Info("DeriveSha Result Hash", "hash", trie.Hash())
	return trie.Hash()
}
func DeriveShaHash(list []common.Hash) common.Hash {
	if len(list) == 0 {
		return EmptyRootHash
	}
	trie := new(trie.Trie)
	for i := 0; i < len(list); i++ {
		buff, _ := rlp.EncodeUint(uint64(i))
		hash1 := list[i]
		trie.Update(buff, hash1[:])
	}
	hash := trie.Hash()
	return hash
}
