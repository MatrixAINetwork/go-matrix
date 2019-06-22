// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package snapshot

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/MatrixAINetwork/go-matrix/common"
)

const SNAPDIR = "./snapdir"

func init() {
	_, e := os.Stat(SNAPDIR)
	if e != nil {
		os.Mkdir(SNAPDIR, os.ModePerm)
	}
}

type Config struct {
	DataDir string
	Name    string
}

var isOldGmanResource = map[string]bool{
	"chaindata":          true,
	"nodes":              true,
	"nodekey":            true,
	"static-nodes.json":  true,
	"trusted-nodes.json": true,
}

var DefaultConfig = Config{
	DataDir: "",
	Name:    "snapshot.db",
}

func (c *Config) name() string {
	if c.Name == "" {
		progname := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if progname == "" {
			panic("empty executable name, set Config.Name")
		}
		return progname
	}
	return c.Name
}

func (c *Config) instanceDir() string {
	if c.DataDir == "" {
		return ""
	}
	return filepath.Join(c.DataDir, c.name())
}

func (c *Config) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if c.DataDir == "" {
		return ""
	}
	// Backwards-compatibility: ensure that data directory files created
	// by gman 1.4 are used if they exist.
	if c.name() == "gman" && isOldGmanResource[path] {
		oldpath := ""
		if c.Name == "gman" {
			oldpath = filepath.Join(c.DataDir, path)
		}
		if oldpath != "" && common.FileExist(oldpath) {
			// TODO: print warning
			return oldpath
		}
	}
	return filepath.Join(c.instanceDir(), path)
}
