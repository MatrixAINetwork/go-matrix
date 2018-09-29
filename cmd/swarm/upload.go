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

// Command bzzup uploads files to the swarm HTTP API.
package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/matrix/go-matrix/cmd/utils"
	swarm "github.com/matrix/go-matrix/swarm/api/client"
	"gopkg.in/urfave/cli.v1"
)

func upload(ctx *cli.Context) {

	args := ctx.Args()
	var (
		bzzapi       = strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
		recursive    = ctx.GlobalBool(SwarmRecursiveUploadFlag.Name)
		wantManifest = ctx.GlobalBoolT(SwarmWantManifestFlag.Name)
		defaultPath  = ctx.GlobalString(SwarmUploadDefaultPath.Name)
		fromStdin    = ctx.GlobalBool(SwarmUpFromStdinFlag.Name)
		mimeType     = ctx.GlobalString(SwarmUploadMimeType.Name)
		client       = swarm.NewClient(bzzapi)
		file         string
	)

	if len(args) != 1 {
		if fromStdin {
			tmp, err := ioutil.TempFile("", "swarm-stdin")
			if err != nil {
				utils.Fatalf("error create tempfile: %s", err)
			}
			defer os.Remove(tmp.Name())
			n, err := io.Copy(tmp, os.Stdin)
			if err != nil {
				utils.Fatalf("error copying stdin to tempfile: %s", err)
			} else if n == 0 {
				utils.Fatalf("error reading from stdin: zero length")
			}
			file = tmp.Name()
		} else {
			utils.Fatalf("Need filename as the first and only argument")
		}
	} else {
		file = expandPath(args[0])
	}

	if !wantManifest {
		f, err := swarm.Open(file)
		if err != nil {
			utils.Fatalf("Error opening file: %s", err)
		}
		defer f.Close()
		hash, err := client.UploadRaw(f, f.Size)
		if err != nil {
			utils.Fatalf("Upload failed: %s", err)
		}
		fmt.Println(hash)
		return
	}

	stat, err := os.Stat(file)
	if err != nil {
		utils.Fatalf("Error opening file: %s", err)
	}

	// define a function which either uploads a directory or single file
	// based on the type of the file being uploaded
	var doUpload func() (hash string, err error)
	if stat.IsDir() {
		doUpload = func() (string, error) {
			if !recursive {
				return "", errors.New("Argument is a directory and recursive upload is disabled")
			}
			return client.UploadDirectory(file, defaultPath, "")
		}
	} else {
		doUpload = func() (string, error) {
			f, err := swarm.Open(file)
			if err != nil {
				return "", fmt.Errorf("error opening file: %s", err)
			}
			defer f.Close()
			if mimeType == "" {
				mimeType = detectMimeType(file)
			}
			f.ContentType = mimeType
			return client.Upload(f, "")
		}
	}
	hash, err := doUpload()
	if err != nil {
		utils.Fatalf("Upload failed: %s", err)
	}
	fmt.Println(hash)
}

// Expands a file path
// 1. replace tilde with users home dir
// 2. expands embedded environment variables
// 3. cleans the path, e.g. /a/b/../c -> /a/c
// Note, it has limitations, e.g. ~someuser/tmp will not be expanded
func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home := homeDir(); home != "" {
			p = home + p[1:]
		}
	}
	return path.Clean(os.ExpandEnv(p))
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func detectMimeType(file string) string {
	if ext := filepath.Ext(file); ext != "" {
		return mime.TypeByExtension(ext)
	}
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()
	buf := make([]byte, 512)
	if n, _ := f.Read(buf); n > 0 {
		return http.DetectContentType(buf)
	}
	return ""
}
