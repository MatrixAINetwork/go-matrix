// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package main

import (
	"fmt"

	"github.com/matrix/go-matrix/log"
)

// ensureVirtualHost checks whether a reverse-proxy is running on the specified
// host machine, and if yes requests a virtual host from the user to host a
// specific web service on. If no proxy exists, the method will offer to deploy
// one.
//
// If the user elects not to use a reverse proxy, an empty hostname is returned!
func (w *wizard) ensureVirtualHost(client *sshClient, port int, def string) (string, error) {
	proxy, _ := checkNginx(client, w.network)
	if proxy != nil {
		// Reverse proxy is running, if ports match, we need a virtual host
		if proxy.port == port {
			fmt.Println()
			fmt.Printf("Shared port, which domain to assign? (default = %s)\n", def)
			return w.readDefaultString(def), nil
		}
	}
	// Reverse proxy is not running, offer to deploy a new one
	fmt.Println()
	fmt.Println("Allow sharing the port with other services (y/n)? (default = yes)")
	if w.readDefaultString("y") == "y" {
		nocache := false
		if proxy != nil {
			fmt.Println()
			fmt.Printf("Should the reverse-proxy be rebuilt from scratch (y/n)? (default = no)\n")
			nocache = w.readDefaultString("n") != "n"
		}
		if out, err := deployNginx(client, w.network, port, nocache); err != nil {
			log.Error("Failed to deploy reverse-proxy", "err", err)
			if len(out) > 0 {
				fmt.Printf("%s\n", out)
			}
			return "", err
		}
		// Reverse proxy deployed, ask again for the virtual-host
		fmt.Println()
		fmt.Printf("Proxy deployed, which domain to assign? (default = %s)\n", def)
		return w.readDefaultString(def), nil
	}
	// Reverse proxy not requested, deploy as a standalone service
	return "", nil
}
