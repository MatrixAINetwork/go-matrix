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
