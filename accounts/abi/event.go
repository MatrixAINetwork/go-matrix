// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package abi

import (
	"fmt"
	"strings"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/crypto"
)

// Event is an event potentially triggered by the EVM's LOG mechanism. The Event
// holds type information (inputs) about the yielded output. Anonymous events
// don't get the signature canonical representation as the first LOG topic.
type Event struct {
	Name      string
	Anonymous bool
	Inputs    Arguments
}

func (e Event) String() string {
	inputs := make([]string, len(e.Inputs))
	for i, input := range e.Inputs {
		inputs[i] = fmt.Sprintf("%v %v", input.Name, input.Type)
		if input.Indexed {
			inputs[i] = fmt.Sprintf("%v indexed %v", input.Name, input.Type)
		}
	}
	return fmt.Sprintf("e %v(%v)", e.Name, strings.Join(inputs, ", "))
}

// Id returns the canonical representation of the event's signature used by the
// abi definition to identify event names and types.
func (e Event) Id() common.Hash {
	types := make([]string, len(e.Inputs))
	i := 0
	for _, input := range e.Inputs {
		types[i] = input.Type.String()
		i++
	}
	return common.BytesToHash(crypto.Keccak256([]byte(fmt.Sprintf("%v(%v)", e.Name, strings.Join(types, ",")))))
}
