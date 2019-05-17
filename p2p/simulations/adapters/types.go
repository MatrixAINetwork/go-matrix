// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package adapters

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/p2p"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
	"github.com/MatrixAINetwork/go-matrix/pod"
	"github.com/MatrixAINetwork/go-matrix/rpc"
	"github.com/docker/docker/pkg/reexec"
)

// Node represents a node in a simulation network which is created by a
// NodeAdapter, for example:
//
// * SimNode    - An in-memory node
// * ExecNode   - A child process node
// * DockerNode - A Docker container node
//
type Node interface {
	// Addr returns the node's address (e.g. an Enode URL)
	Addr() []byte

	// Client returns the RPC client which is created once the node is
	// up and running
	Client() (*rpc.Client, error)

	// ServeRPC serves RPC requests over the given connection
	ServeRPC(net.Conn) error

	// Start starts the node with the given snapshots
	Start(snapshots map[string][]byte) error

	// Stop stops the node
	Stop() error

	// NodeInfo returns information about the node
	NodeInfo() *p2p.NodeInfo

	// Snapshots creates snapshots of the running services
	Snapshots() (map[string][]byte, error)
}

// NodeAdapter is used to create Nodes in a simulation network
type NodeAdapter interface {
	// Name returns the name of the adapter for logging purposes
	Name() string

	// NewNode creates a new node with the given configuration
	NewNode(config *NodeConfig) (Node, error)
}

// NodeConfig is the configuration used to start a node in a simulation
// network
type NodeConfig struct {
	// ID is the node's ID which is used to identify the node in the
	// simulation network
	ID discover.NodeID

	// PrivateKey is the node's private key which is used by the devp2p
	// stack to encrypt communications
	PrivateKey *ecdsa.PrivateKey

	// Enable peer events for Msgs
	EnableMsgEvents bool

	// Name is a human friendly name for the node like "node01"
	Name string

	// Services are the names of the services which should be run when
	// starting the node (for SimNodes it should be the names of services
	// contained in SimAdapter.services, for other nodes it should be
	// services registered by calling the RegisterService function)
	Services []string

	// function to sanction or prevent suggesting a peer
	Reachable func(id discover.NodeID) bool
}

// nodeConfigJSON is used to encode and decode NodeConfig as JSON by encoding
// all fields as strings
type nodeConfigJSON struct {
	ID         string   `json:"id"`
	PrivateKey string   `json:"private_key"`
	Name       string   `json:"name"`
	Services   []string `json:"services"`
}

// MarshalJSON implements the json.Marshaler interface by encoding the config
// fields as strings
func (n *NodeConfig) MarshalJSON() ([]byte, error) {
	confJSON := nodeConfigJSON{
		ID:       n.ID.String(),
		Name:     n.Name,
		Services: n.Services,
	}
	if n.PrivateKey != nil {
		confJSON.PrivateKey = hex.EncodeToString(crypto.FromECDSA(n.PrivateKey))
	}
	return json.Marshal(confJSON)
}

// UnmarshalJSON implements the json.Unmarshaler interface by decoding the json
// string values into the config fields
func (n *NodeConfig) UnmarshalJSON(data []byte) error {
	var confJSON nodeConfigJSON
	if err := json.Unmarshal(data, &confJSON); err != nil {
		return err
	}

	if confJSON.ID != "" {
		nodeID, err := discover.HexID(confJSON.ID)
		if err != nil {
			return err
		}
		n.ID = nodeID
	}

	if confJSON.PrivateKey != "" {
		key, err := hex.DecodeString(confJSON.PrivateKey)
		if err != nil {
			return err
		}
		privKey, err := crypto.ToECDSA(key)
		if err != nil {
			return err
		}
		n.PrivateKey = privKey
	}

	n.Name = confJSON.Name
	n.Services = confJSON.Services

	return nil
}

// RandomNodeConfig returns node configuration with a randomly generated ID and
// PrivateKey
func RandomNodeConfig() *NodeConfig {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic("unable to generate key")
	}
	var id discover.NodeID
	pubkey := crypto.FromECDSAPub(&key.PublicKey)
	copy(id[:], pubkey[1:])
	return &NodeConfig{
		ID:         id,
		PrivateKey: key,
	}
}

// ServiceContext is a collection of options and methods which can be utilised
// when starting services
type ServiceContext struct {
	RPCDialer

	NodeContext *node.ServiceContext
	Config      *NodeConfig
	Snapshot    []byte
}

// RPCDialer is used when initialising services which need to connect to
// other nodes in the network (for example a simulated Swarm node which needs
// to connect to a Gman node to resolve ENS names)
type RPCDialer interface {
	DialRPC(id discover.NodeID) (*rpc.Client, error)
}

// Services is a collection of services which can be run in a simulation
type Services map[string]ServiceFunc

// ServiceFunc returns a node.Service which can be used to boot a devp2p node
type ServiceFunc func(ctx *ServiceContext) (node.Service, error)

// serviceFuncs is a map of registered services which are used to boot devp2p
// nodes
var serviceFuncs = make(Services)

// RegisterServices registers the given Services which can then be used to
// start devp2p nodes using either the Exec or Docker adapters.
//
// It should be called in an init function so that it has the opportunity to
// execute the services before main() is called.
func RegisterServices(services Services) {
	for name, f := range services {
		if _, exists := serviceFuncs[name]; exists {
			panic(fmt.Sprintf("node service already exists: %q", name))
		}
		serviceFuncs[name] = f
	}

	// now we have registered the services, run reexec.Init() which will
	// potentially start one of the services if the current binary has
	// been exec'd with argv[0] set to "p2p-node"
	if reexec.Init() {
		os.Exit(0)
	}
}
