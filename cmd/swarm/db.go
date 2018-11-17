// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/matrix/go-matrix/cmd/utils"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/swarm/storage"
	"gopkg.in/urfave/cli.v1"
)

func dbExport(ctx *cli.Context) {
	args := ctx.Args()
	if len(args) != 2 {
		utils.Fatalf("invalid arguments, please specify both <chunkdb> (path to a local chunk database) and <file> (path to write the tar archive to, - for stdout)")
	}

	store, err := openDbStore(args[0])
	if err != nil {
		utils.Fatalf("error opening local chunk database: %s", err)
	}
	defer store.Close()

	var out io.Writer
	if args[1] == "-" {
		out = os.Stdout
	} else {
		f, err := os.Create(args[1])
		if err != nil {
			utils.Fatalf("error opening output file: %s", err)
		}
		defer f.Close()
		out = f
	}

	count, err := store.Export(out)
	if err != nil {
		utils.Fatalf("error exporting local chunk database: %s", err)
	}

	log.Info(fmt.Sprintf("successfully exported %d chunks", count))
}

func dbImport(ctx *cli.Context) {
	args := ctx.Args()
	if len(args) != 2 {
		utils.Fatalf("invalid arguments, please specify both <chunkdb> (path to a local chunk database) and <file> (path to read the tar archive from, - for stdin)")
	}

	store, err := openDbStore(args[0])
	if err != nil {
		utils.Fatalf("error opening local chunk database: %s", err)
	}
	defer store.Close()

	var in io.Reader
	if args[1] == "-" {
		in = os.Stdin
	} else {
		f, err := os.Open(args[1])
		if err != nil {
			utils.Fatalf("error opening input file: %s", err)
		}
		defer f.Close()
		in = f
	}

	count, err := store.Import(in)
	if err != nil {
		utils.Fatalf("error importing local chunk database: %s", err)
	}

	log.Info(fmt.Sprintf("successfully imported %d chunks", count))
}

func dbClean(ctx *cli.Context) {
	args := ctx.Args()
	if len(args) != 1 {
		utils.Fatalf("invalid arguments, please specify <chunkdb> (path to a local chunk database)")
	}

	store, err := openDbStore(args[0])
	if err != nil {
		utils.Fatalf("error opening local chunk database: %s", err)
	}
	defer store.Close()

	store.Cleanup()
}

func openDbStore(path string) (*storage.DbStore, error) {
	if _, err := os.Stat(filepath.Join(path, "CURRENT")); err != nil {
		return nil, fmt.Errorf("invalid chunkdb path: %s", err)
	}
	hash := storage.MakeHashFunc("SHA3")
	return storage.NewDbStore(path, hash, 10000000, 0)
}
