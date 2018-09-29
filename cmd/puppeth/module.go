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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/matrix/go-matrix/log"
)

var (
	// ErrServiceUnknown is returned when a service container doesn't exist.
	ErrServiceUnknown = errors.New("service unknown")

	// ErrServiceOffline is returned when a service container exists, but it is not
	// running.
	ErrServiceOffline = errors.New("service offline")

	// ErrServiceUnreachable is returned when a service container is running, but
	// seems to not respond to communication attempts.
	ErrServiceUnreachable = errors.New("service unreachable")

	// ErrNotExposed is returned if a web-service doesn't have an exposed port, nor
	// a reverse-proxy in front of it to forward requests.
	ErrNotExposed = errors.New("service not exposed, nor proxied")
)

// containerInfos is a heavily reduced version of the huge inspection dataset
// returned from docker inspect, parsed into a form easily usable by puppeth.
type containerInfos struct {
	running bool              // Flag whether the container is running currently
	envvars map[string]string // Collection of environmental variables set on the container
	portmap map[string]int    // Port mapping from internal port/proto combos to host binds
	volumes map[string]string // Volume mount points from container to host directories
}

// inspectContainer runs docker inspect against a running container
func inspectContainer(client *sshClient, container string) (*containerInfos, error) {
	// Check whether there's a container running for the service
	out, err := client.Run(fmt.Sprintf("docker inspect %s", container))
	if err != nil {
		return nil, ErrServiceUnknown
	}
	// If yes, extract various configuration options
	type inspection struct {
		State struct {
			Running bool
		}
		Mounts []struct {
			Source      string
			Destination string
		}
		Config struct {
			Env []string
		}
		HostConfig struct {
			PortBindings map[string][]map[string]string
		}
	}
	var inspects []inspection
	if err = json.Unmarshal(out, &inspects); err != nil {
		return nil, err
	}
	inspect := inspects[0]

	// Infos retrieved, parse the above into something meaningful
	infos := &containerInfos{
		running: inspect.State.Running,
		envvars: make(map[string]string),
		portmap: make(map[string]int),
		volumes: make(map[string]string),
	}
	for _, envvar := range inspect.Config.Env {
		if parts := strings.Split(envvar, "="); len(parts) == 2 {
			infos.envvars[parts[0]] = parts[1]
		}
	}
	for portname, details := range inspect.HostConfig.PortBindings {
		if len(details) > 0 {
			port, _ := strconv.Atoi(details[0]["HostPort"])
			infos.portmap[portname] = port
		}
	}
	for _, mount := range inspect.Mounts {
		infos.volumes[mount.Destination] = mount.Source
	}
	return infos, err
}

// tearDown connects to a remote machine via SSH and terminates docker containers
// running with the specified name in the specified network.
func tearDown(client *sshClient, network string, service string, purge bool) ([]byte, error) {
	// Tear down the running (or paused) container
	out, err := client.Run(fmt.Sprintf("docker rm -f %s_%s_1", network, service))
	if err != nil {
		return out, err
	}
	// If requested, purge the associated docker image too
	if purge {
		return client.Run(fmt.Sprintf("docker rmi %s/%s", network, service))
	}
	return nil, nil
}

// resolve retrieves the hostname a service is running on either by returning the
// actual server name and port, or preferably an nginx virtual host if available.
func resolve(client *sshClient, network string, service string, port int) (string, error) {
	// Inspect the service to get various configurations from it
	infos, err := inspectContainer(client, fmt.Sprintf("%s_%s_1", network, service))
	if err != nil {
		return "", err
	}
	if !infos.running {
		return "", ErrServiceOffline
	}
	// Container online, extract any environmental variables
	if vhost := infos.envvars["VIRTUAL_HOST"]; vhost != "" {
		return vhost, nil
	}
	return fmt.Sprintf("%s:%d", client.server, port), nil
}

// checkPort tries to connect to a remote host on a given
func checkPort(host string, port int) error {
	log.Trace("Verifying remote TCP connectivity", "server", host, "port", port)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
