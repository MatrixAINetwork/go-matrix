// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Contains all the wrappers from the accounts package to support client side enode
// management on mobile platforms.

package gman

import (
	"errors"

	"github.com/matrix/go-matrix/p2p/discv5"
)

// Enode represents a host on the network.
type Enode struct {
	node *discv5.Node
}

// NewEnode parses a node designator.
//
// There are two basic forms of node designators
//   - incomplete nodes, which only have the public key (node ID)
//   - complete nodes, which contain the public key and IP/Port information
//
// For incomplete nodes, the designator must look like one of these
//
//    enode://<hex node id>
//    <hex node id>
//
// For complete nodes, the node ID is encoded in the username portion
// of the URL, separated from the host by an @ sign. The hostname can
// only be given as an IP address, DNS domain names are not allowed.
// The port in the host name section is the TCP listening port. If the
// TCP and UDP (discovery) ports differ, the UDP port is specified as
// query parameter "discport".
//
// In the following example, the node URL describes
// a node with IP address 10.3.58.6, TCP listening port 30303
// and UDP discovery port 30301.
//
//    enode://<hex node id>@10.3.58.6:30303?discport=30301
func NewEnode(rawurl string) (enode *Enode, _ error) {
	node, err := discv5.ParseNode(rawurl)
	if err != nil {
		return nil, err
	}
	return &Enode{node}, nil
}

// Enodes represents a slice of accounts.
type Enodes struct{ nodes []*discv5.Node }

// NewEnodes creates a slice of uninitialized enodes.
func NewEnodes(size int) *Enodes {
	return &Enodes{
		nodes: make([]*discv5.Node, size),
	}
}

// NewEnodesEmpty creates an empty slice of Enode values.
func NewEnodesEmpty() *Enodes {
	return NewEnodes(0)
}

// Size returns the number of enodes in the slice.
func (e *Enodes) Size() int {
	return len(e.nodes)
}

// Get returns the enode at the given index from the slice.
func (e *Enodes) Get(index int) (enode *Enode, _ error) {
	if index < 0 || index >= len(e.nodes) {
		return nil, errors.New("index out of bounds")
	}
	return &Enode{e.nodes[index]}, nil
}

// Set sets the enode at the given index in the slice.
func (e *Enodes) Set(index int, enode *Enode) error {
	if index < 0 || index >= len(e.nodes) {
		return errors.New("index out of bounds")
	}
	e.nodes[index] = enode.node
	return nil
}

// Append adds a new enode element to the end of the slice.
func (e *Enodes) Append(enode *Enode) {
	e.nodes = append(e.nodes, enode.node)
}
