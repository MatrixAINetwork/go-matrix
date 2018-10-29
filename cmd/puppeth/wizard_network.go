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

package main

import (
	"fmt"
	"strings"

	"github.com/matrix/go-matrix/log"
)

// manageServers displays a list of servers the user can disconnect from, and an
// option to connect to new servers.
func (w *wizard) manageServers() {
	// List all the servers we can disconnect, along with an entry to connect a new one
	fmt.Println()

	servers := w.conf.servers()
	for i, server := range servers {
		fmt.Printf(" %d. Disconnect %s\n", i+1, server)
	}
	fmt.Printf(" %d. Connect another server\n", len(w.conf.Servers)+1)

	choice := w.readInt()
	if choice < 0 || choice > len(w.conf.Servers)+1 {
		log.Error("Invalid server choice, aborting")
		return
	}
	// If the user selected an existing server, drop it
	if choice <= len(w.conf.Servers) {
		server := servers[choice-1]
		client := w.servers[server]

		delete(w.servers, server)
		if client != nil {
			client.Close()
		}
		delete(w.conf.Servers, server)
		w.conf.flush()

		log.Info("Disconnected existing server", "server", server)
		w.networkStats()
		return
	}
	// If the user requested connecting a new server, do it
	if w.makeServer() != "" {
		w.networkStats()
	}
}

// makeServer reads a single line from stdin and interprets it as a hostname to
// connect to. It tries to establish a new SSH session and also executing some
// baseline validations.
//
// If connection succeeds, the server is added to the wizards configs!
func (w *wizard) makeServer() string {
	fmt.Println()
	fmt.Println("Please enter remote server's address:")

	// Read and dial the server to ensure docker is present
	input := w.readString()

	client, err := dial(input, nil)
	if err != nil {
		log.Error("Server not ready for puppeth", "err", err)
		return ""
	}
	// All checks passed, start tracking the server
	w.servers[input] = client
	w.conf.Servers[input] = client.pubkey
	w.conf.flush()

	return input
}

// selectServer lists the user all the currnetly known servers to choose from,
// also granting the option to add a new one.
func (w *wizard) selectServer() string {
	// List the available server to the user and wait for a choice
	fmt.Println()
	fmt.Println("Which server do you want to interact with?")

	servers := w.conf.servers()
	for i, server := range servers {
		fmt.Printf(" %d. %s\n", i+1, server)
	}
	fmt.Printf(" %d. Connect another server\n", len(w.conf.Servers)+1)

	choice := w.readInt()
	if choice < 0 || choice > len(w.conf.Servers)+1 {
		log.Error("Invalid server choice, aborting")
		return ""
	}
	// If the user requested connecting to a new server, go for it
	if choice <= len(w.conf.Servers) {
		return servers[choice-1]
	}
	return w.makeServer()
}

// manageComponents displays a list of network components the user can tear down
// and an option
func (w *wizard) manageComponents() {
	// List all the componens we can tear down, along with an entry to deploy a new one
	fmt.Println()

	var serviceHosts, serviceNames []string
	for server, services := range w.services {
		for _, service := range services {
			serviceHosts = append(serviceHosts, server)
			serviceNames = append(serviceNames, service)

			fmt.Printf(" %d. Tear down %s on %s\n", len(serviceHosts), strings.Title(service), server)
		}
	}
	fmt.Printf(" %d. Deploy new network component\n", len(serviceHosts)+1)

	choice := w.readInt()
	if choice < 0 || choice > len(serviceHosts)+1 {
		log.Error("Invalid component choice, aborting")
		return
	}
	// If the user selected an existing service, destroy it
	if choice <= len(serviceHosts) {
		// Figure out the service to destroy and execute it
		service := serviceNames[choice-1]
		server := serviceHosts[choice-1]
		client := w.servers[server]

		if out, err := tearDown(client, w.network, service, true); err != nil {
			log.Error("Failed to tear down component", "err", err)
			if len(out) > 0 {
				fmt.Printf("%s\n", out)
			}
			return
		}
		// Clean up any references to it from out state
		services := w.services[server]
		for i, name := range services {
			if name == service {
				w.services[server] = append(services[:i], services[i+1:]...)
				if len(w.services[server]) == 0 {
					delete(w.services, server)
				}
			}
		}
		log.Info("Torn down existing component", "server", server, "service", service)
		return
	}
	// If the user requested deploying a new component, do it
	w.deployComponent()
}

// deployComponent displays a list of network components the user can deploy and
// guides through the process.
func (w *wizard) deployComponent() {
	// Print all the things we can deploy and wait or user choice
	fmt.Println()
	fmt.Println("What would you like to deploy? (recommended order)")
	fmt.Println(" 1. Ethstats  - Network monitoring tool")
	fmt.Println(" 2. Bootnode  - Entry point of the network")
	fmt.Println(" 3. Sealer    - Full node minting new blocks")
	fmt.Println(" 4. Explorer  - Chain analysis webservice (manash only)")
	fmt.Println(" 5. Wallet    - Browser wallet for quick sends")
	fmt.Println(" 6. Faucet    - Crypto faucet to give away funds")
	fmt.Println(" 7. Dashboard - Website listing above web-services")

	switch w.read() {
	case "1":
		w.deployEthstats()
	case "2":
		w.deployNode(true)
	case "3":
		w.deployNode(false)
	case "4":
		w.deployExplorer()
	case "5":
		w.deployWallet()
	case "6":
		w.deployFaucet()
	case "7":
		w.deployDashboard()
	default:
		log.Error("That's not something I can do")
	}
}
