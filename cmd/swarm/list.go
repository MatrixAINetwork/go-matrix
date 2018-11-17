// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/matrix/go-matrix/cmd/utils"
	swarm "github.com/matrix/go-matrix/swarm/api/client"
	"gopkg.in/urfave/cli.v1"
)

func list(ctx *cli.Context) {
	args := ctx.Args()

	if len(args) < 1 {
		utils.Fatalf("Please supply a manifest reference as the first argument")
	} else if len(args) > 2 {
		utils.Fatalf("Too many arguments - usage 'swarm ls manifest [prefix]'")
	}
	manifest := args[0]

	var prefix string
	if len(args) == 2 {
		prefix = args[1]
	}

	bzzapi := strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
	client := swarm.NewClient(bzzapi)
	list, err := client.List(manifest, prefix)
	if err != nil {
		utils.Fatalf("Failed to generate file and directory list: %s", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	fmt.Fprintln(w, "HASH\tCONTENT TYPE\tPATH")
	for _, prefix := range list.CommonPrefixes {
		fmt.Fprintf(w, "%s\t%s\t%s\n", "", "DIR", prefix)
	}
	for _, entry := range list.Entries {
		fmt.Fprintf(w, "%s\t%s\t%s\n", entry.Hash, entry.ContentType, entry.Path)
	}
}
