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

// Command  MANIFEST update
package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/matrix/go-matrix/cmd/utils"
	"github.com/matrix/go-matrix/swarm/api"
	swarm "github.com/matrix/go-matrix/swarm/api/client"
	"gopkg.in/urfave/cli.v1"
)

const bzzManifestJSON = "application/bzz-manifest+json"

func add(ctx *cli.Context) {
	args := ctx.Args()
	if len(args) < 3 {
		utils.Fatalf("Need at least three arguments <MHASH> <path> <HASH> [<content-type>]")
	}

	var (
		mhash = args[0]
		path  = args[1]
		hash  = args[2]

		ctype        string
		wantManifest = ctx.GlobalBoolT(SwarmWantManifestFlag.Name)
		mroot        api.Manifest
	)

	if len(args) > 3 {
		ctype = args[3]
	} else {
		ctype = mime.TypeByExtension(filepath.Ext(path))
	}

	newManifest := addEntryToManifest(ctx, mhash, path, hash, ctype)
	fmt.Println(newManifest)

	if !wantManifest {
		// Print the manifest. This is the only output to stdout.
		mrootJSON, _ := json.MarshalIndent(mroot, "", "  ")
		fmt.Println(string(mrootJSON))
		return
	}
}

func update(ctx *cli.Context) {

	args := ctx.Args()
	if len(args) < 3 {
		utils.Fatalf("Need at least three arguments <MHASH> <path> <HASH>")
	}

	var (
		mhash = args[0]
		path  = args[1]
		hash  = args[2]

		ctype        string
		wantManifest = ctx.GlobalBoolT(SwarmWantManifestFlag.Name)
		mroot        api.Manifest
	)
	if len(args) > 3 {
		ctype = args[3]
	} else {
		ctype = mime.TypeByExtension(filepath.Ext(path))
	}

	newManifest := updateEntryInManifest(ctx, mhash, path, hash, ctype)
	fmt.Println(newManifest)

	if !wantManifest {
		// Print the manifest. This is the only output to stdout.
		mrootJSON, _ := json.MarshalIndent(mroot, "", "  ")
		fmt.Println(string(mrootJSON))
		return
	}
}

func remove(ctx *cli.Context) {
	args := ctx.Args()
	if len(args) < 2 {
		utils.Fatalf("Need at least two arguments <MHASH> <path>")
	}

	var (
		mhash = args[0]
		path  = args[1]

		wantManifest = ctx.GlobalBoolT(SwarmWantManifestFlag.Name)
		mroot        api.Manifest
	)

	newManifest := removeEntryFromManifest(ctx, mhash, path)
	fmt.Println(newManifest)

	if !wantManifest {
		// Print the manifest. This is the only output to stdout.
		mrootJSON, _ := json.MarshalIndent(mroot, "", "  ")
		fmt.Println(string(mrootJSON))
		return
	}
}

func addEntryToManifest(ctx *cli.Context, mhash, path, hash, ctype string) string {

	var (
		bzzapi           = strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
		client           = swarm.NewClient(bzzapi)
		longestPathEntry = api.ManifestEntry{}
	)

	mroot, err := client.DownloadManifest(mhash)
	if err != nil {
		utils.Fatalf("Manifest download failed: %v", err)
	}

	//TODO: check if the "hash" to add is valid and present in swarm
	_, err = client.DownloadManifest(hash)
	if err != nil {
		utils.Fatalf("Hash to add is not present: %v", err)
	}

	// See if we path is in this Manifest or do we have to dig deeper
	for _, entry := range mroot.Entries {
		if path == entry.Path {
			utils.Fatalf("Path %s already present, not adding anything", path)
		} else {
			if entry.ContentType == bzzManifestJSON {
				prfxlen := strings.HasPrefix(path, entry.Path)
				if prfxlen && len(path) > len(longestPathEntry.Path) {
					longestPathEntry = entry
				}
			}
		}
	}

	if longestPathEntry.Path != "" {
		// Load the child Manifest add the entry there
		newPath := path[len(longestPathEntry.Path):]
		newHash := addEntryToManifest(ctx, longestPathEntry.Hash, newPath, hash, ctype)

		// Replace the hash for parent Manifests
		newMRoot := &api.Manifest{}
		for _, entry := range mroot.Entries {
			if longestPathEntry.Path == entry.Path {
				entry.Hash = newHash
			}
			newMRoot.Entries = append(newMRoot.Entries, entry)
		}
		mroot = newMRoot
	} else {
		// Add the entry in the leaf Manifest
		newEntry := api.ManifestEntry{
			Hash:        hash,
			Path:        path,
			ContentType: ctype,
		}
		mroot.Entries = append(mroot.Entries, newEntry)
	}

	newManifestHash, err := client.UploadManifest(mroot)
	if err != nil {
		utils.Fatalf("Manifest upload failed: %v", err)
	}
	return newManifestHash

}

func updateEntryInManifest(ctx *cli.Context, mhash, path, hash, ctype string) string {

	var (
		bzzapi           = strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
		client           = swarm.NewClient(bzzapi)
		newEntry         = api.ManifestEntry{}
		longestPathEntry = api.ManifestEntry{}
	)

	mroot, err := client.DownloadManifest(mhash)
	if err != nil {
		utils.Fatalf("Manifest download failed: %v", err)
	}

	//TODO: check if the "hash" with which to update is valid and present in swarm

	// See if we path is in this Manifest or do we have to dig deeper
	for _, entry := range mroot.Entries {
		if path == entry.Path {
			newEntry = entry
		} else {
			if entry.ContentType == bzzManifestJSON {
				prfxlen := strings.HasPrefix(path, entry.Path)
				if prfxlen && len(path) > len(longestPathEntry.Path) {
					longestPathEntry = entry
				}
			}
		}
	}

	if longestPathEntry.Path == "" && newEntry.Path == "" {
		utils.Fatalf("Path %s not present in the Manifest, not setting anything", path)
	}

	if longestPathEntry.Path != "" {
		// Load the child Manifest add the entry there
		newPath := path[len(longestPathEntry.Path):]
		newHash := updateEntryInManifest(ctx, longestPathEntry.Hash, newPath, hash, ctype)

		// Replace the hash for parent Manifests
		newMRoot := &api.Manifest{}
		for _, entry := range mroot.Entries {
			if longestPathEntry.Path == entry.Path {
				entry.Hash = newHash
			}
			newMRoot.Entries = append(newMRoot.Entries, entry)

		}
		mroot = newMRoot
	}

	if newEntry.Path != "" {
		// Replace the hash for leaf Manifest
		newMRoot := &api.Manifest{}
		for _, entry := range mroot.Entries {
			if newEntry.Path == entry.Path {
				myEntry := api.ManifestEntry{
					Hash:        hash,
					Path:        entry.Path,
					ContentType: ctype,
				}
				newMRoot.Entries = append(newMRoot.Entries, myEntry)
			} else {
				newMRoot.Entries = append(newMRoot.Entries, entry)
			}
		}
		mroot = newMRoot
	}

	newManifestHash, err := client.UploadManifest(mroot)
	if err != nil {
		utils.Fatalf("Manifest upload failed: %v", err)
	}
	return newManifestHash
}

func removeEntryFromManifest(ctx *cli.Context, mhash, path string) string {

	var (
		bzzapi           = strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
		client           = swarm.NewClient(bzzapi)
		entryToRemove    = api.ManifestEntry{}
		longestPathEntry = api.ManifestEntry{}
	)

	mroot, err := client.DownloadManifest(mhash)
	if err != nil {
		utils.Fatalf("Manifest download failed: %v", err)
	}

	// See if we path is in this Manifest or do we have to dig deeper
	for _, entry := range mroot.Entries {
		if path == entry.Path {
			entryToRemove = entry
		} else {
			if entry.ContentType == bzzManifestJSON {
				prfxlen := strings.HasPrefix(path, entry.Path)
				if prfxlen && len(path) > len(longestPathEntry.Path) {
					longestPathEntry = entry
				}
			}
		}
	}

	if longestPathEntry.Path == "" && entryToRemove.Path == "" {
		utils.Fatalf("Path %s not present in the Manifest, not removing anything", path)
	}

	if longestPathEntry.Path != "" {
		// Load the child Manifest remove the entry there
		newPath := path[len(longestPathEntry.Path):]
		newHash := removeEntryFromManifest(ctx, longestPathEntry.Hash, newPath)

		// Replace the hash for parent Manifests
		newMRoot := &api.Manifest{}
		for _, entry := range mroot.Entries {
			if longestPathEntry.Path == entry.Path {
				entry.Hash = newHash
			}
			newMRoot.Entries = append(newMRoot.Entries, entry)
		}
		mroot = newMRoot
	}

	if entryToRemove.Path != "" {
		// remove the entry in this Manifest
		newMRoot := &api.Manifest{}
		for _, entry := range mroot.Entries {
			if entryToRemove.Path != entry.Path {
				newMRoot.Entries = append(newMRoot.Entries, entry)
			}
		}
		mroot = newMRoot
	}

	newManifestHash, err := client.UploadManifest(mroot)
	if err != nil {
		utils.Fatalf("Manifest upload failed: %v", err)
	}
	return newManifestHash
}
