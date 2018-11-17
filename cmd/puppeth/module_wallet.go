// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/matrix/go-matrix/log"
)

// walletDockerfile is the Dockerfile required to run a web wallet.
var walletDockerfile = `
FROM puppeth/wallet:latest

ADD genesis.json /genesis.json

RUN \
  echo 'node server.js &'                     > wallet.sh && \
	echo 'gman --cache 512 init /genesis.json' >> wallet.sh && \
	echo $'gman --networkid {{.NetworkID}} --port {{.NodePort}} --bootnodes {{.Bootnodes}} --manstats \'{{.Ethstats}}\' --cache=512 --rpc --rpcaddr=0.0.0.0 --rpccorsdomain "*" --rpcvhosts "*"' >> wallet.sh

RUN \
	sed -i 's/PuppethNetworkID/{{.NetworkID}}/g' dist/js/manwallet-master.js && \
	sed -i 's/PuppethNetwork/{{.Network}}/g'     dist/js/manwallet-master.js && \
	sed -i 's/PuppethDenom/{{.Denom}}/g'         dist/js/manwallet-master.js && \
	sed -i 's/PuppethHost/{{.Host}}/g'           dist/js/manwallet-master.js && \
	sed -i 's/PuppethRPCPort/{{.RPCPort}}/g'     dist/js/manwallet-master.js

ENTRYPOINT ["/bin/sh", "wallet.sh"]
`

// walletComposefile is the docker-compose.yml file required to deploy and
// maintain a web wallet.
var walletComposefile = `
version: '2'
services:
  wallet:
    build: .
    image: {{.Network}}/wallet
    ports:
      - "{{.NodePort}}:{{.NodePort}}"
      - "{{.NodePort}}:{{.NodePort}}/udp"
      - "{{.RPCPort}}:8545"{{if not .VHost}}
      - "{{.WebPort}}:80"{{end}}
    volumes:
      - {{.Datadir}}:/root/.matrix
    environment:
      - NODE_PORT={{.NodePort}}/tcp
      - STATS={{.Ethstats}}{{if .VHost}}
      - VIRTUAL_HOST={{.VHost}}
      - VIRTUAL_PORT=80{{end}}
    logging:
      driver: "json-file"
      options:
        max-size: "1m"
        max-file: "10"
    restart: always
`

// deployWallet deploys a new web wallet container to a remote machine via SSH,
// docker and docker-compose. If an instance with the specified network name
// already exists there, it will be overwritten!
func deployWallet(client *sshClient, network string, bootnodes []string, config *walletInfos, nocache bool) ([]byte, error) {
	// Generate the content to upload to the server
	workdir := fmt.Sprintf("%d", rand.Int63())
	files := make(map[string][]byte)

	dockerfile := new(bytes.Buffer)
	template.Must(template.New("").Parse(walletDockerfile)).Execute(dockerfile, map[string]interface{}{
		"Network":   strings.ToTitle(network),
		"Denom":     strings.ToUpper(network),
		"NetworkID": config.network,
		"NodePort":  config.nodePort,
		"RPCPort":   config.rpcPort,
		"Bootnodes": strings.Join(bootnodes, ","),
		"Ethstats":  config.manstats,
		"Host":      client.address,
	})
	files[filepath.Join(workdir, "Dockerfile")] = dockerfile.Bytes()

	composefile := new(bytes.Buffer)
	template.Must(template.New("").Parse(walletComposefile)).Execute(composefile, map[string]interface{}{
		"Datadir":  config.datadir,
		"Network":  network,
		"NodePort": config.nodePort,
		"RPCPort":  config.rpcPort,
		"VHost":    config.webHost,
		"WebPort":  config.webPort,
		"Ethstats": config.manstats[:strings.Index(config.manstats, ":")],
	})
	files[filepath.Join(workdir, "docker-compose.yaml")] = composefile.Bytes()

	files[filepath.Join(workdir, "genesis.json")] = config.genesis

	// Upload the deployment files to the remote server (and clean up afterwards)
	if out, err := client.Upload(files); err != nil {
		return out, err
	}
	defer client.Run("rm -rf " + workdir)

	// Build and deploy the boot or seal node service
	if nocache {
		return nil, client.Stream(fmt.Sprintf("cd %s && docker-compose -p %s build --pull --no-cache && docker-compose -p %s up -d --force-recreate", workdir, network, network))
	}
	return nil, client.Stream(fmt.Sprintf("cd %s && docker-compose -p %s up -d --build --force-recreate", workdir, network))
}

// walletInfos is returned from a web wallet status check to allow reporting
// various configuration parameters.
type walletInfos struct {
	genesis  []byte
	network  int64
	datadir  string
	manstats string
	nodePort int
	rpcPort  int
	webHost  string
	webPort  int
}

// Report converts the typed struct into a plain string->string map, containing
// most - but not all - fields for reporting to the user.
func (info *walletInfos) Report() map[string]string {
	report := map[string]string{
		"Data directory":         info.datadir,
		"Ethstats username":      info.manstats,
		"Node listener port ":    strconv.Itoa(info.nodePort),
		"RPC listener port ":     strconv.Itoa(info.rpcPort),
		"Website address ":       info.webHost,
		"Website listener port ": strconv.Itoa(info.webPort),
	}
	return report
}

// checkWallet does a health-check against web wallet server to verify whether
// it's running, and if yes, whether it's responsive.
func checkWallet(client *sshClient, network string) (*walletInfos, error) {
	// Inspect a possible web wallet container on the host
	infos, err := inspectContainer(client, fmt.Sprintf("%s_wallet_1", network))
	if err != nil {
		return nil, err
	}
	if !infos.running {
		return nil, ErrServiceOffline
	}
	// Resolve the port from the host, or the reverse proxy
	webPort := infos.portmap["80/tcp"]
	if webPort == 0 {
		if proxy, _ := checkNginx(client, network); proxy != nil {
			webPort = proxy.port
		}
	}
	if webPort == 0 {
		return nil, ErrNotExposed
	}
	// Resolve the host from the reverse-proxy and the config values
	host := infos.envvars["VIRTUAL_HOST"]
	if host == "" {
		host = client.server
	}
	// Run a sanity check to see if the devp2p and RPC ports are reachable
	nodePort := infos.portmap[infos.envvars["NODE_PORT"]]
	if err = checkPort(client.server, nodePort); err != nil {
		log.Warn(fmt.Sprintf("Wallet devp2p port seems unreachable"), "server", client.server, "port", nodePort, "err", err)
	}
	rpcPort := infos.portmap["8545/tcp"]
	if err = checkPort(client.server, rpcPort); err != nil {
		log.Warn(fmt.Sprintf("Wallet RPC port seems unreachable"), "server", client.server, "port", rpcPort, "err", err)
	}
	// Assemble and return the useful infos
	stats := &walletInfos{
		datadir:  infos.volumes["/root/.matrix"],
		nodePort: nodePort,
		rpcPort:  rpcPort,
		webHost:  host,
		webPort:  webPort,
		manstats: infos.envvars["STATS"],
	}
	return stats, nil
}
