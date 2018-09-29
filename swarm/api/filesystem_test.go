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

package api

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/swarm/storage"
)

var testDownloadDir, _ = ioutil.TempDir(os.TempDir(), "bzz-test")

func testFileSystem(t *testing.T, f func(*FileSystem)) {
	testApi(t, func(api *Api) {
		f(NewFileSystem(api))
	})
}

func readPath(t *testing.T, parts ...string) string {
	file := filepath.Join(parts...)
	content, err := ioutil.ReadFile(file)

	if err != nil {
		t.Fatalf("unexpected error reading '%v': %v", file, err)
	}
	return string(content)
}

func TestApiDirUpload0(t *testing.T) {
	testFileSystem(t, func(fs *FileSystem) {
		api := fs.api
		bzzhash, err := fs.Upload(filepath.Join("testdata", "test0"), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		content := readPath(t, "testdata", "test0", "index.html")
		resp := testGet(t, api, bzzhash, "index.html")
		exp := expResponse(content, "text/html; charset=utf-8", 0)
		checkResponse(t, resp, exp)

		content = readPath(t, "testdata", "test0", "index.css")
		resp = testGet(t, api, bzzhash, "index.css")
		exp = expResponse(content, "text/css", 0)
		checkResponse(t, resp, exp)

		key := storage.Key(common.Hex2Bytes(bzzhash))
		_, _, _, err = api.Get(key, "")
		if err == nil {
			t.Fatalf("expected error: %v", err)
		}

		downloadDir := filepath.Join(testDownloadDir, "test0")
		defer os.RemoveAll(downloadDir)
		err = fs.Download(bzzhash, downloadDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		newbzzhash, err := fs.Upload(downloadDir, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bzzhash != newbzzhash {
			t.Fatalf("download %v reuploaded has incorrect hash, expected %v, got %v", downloadDir, bzzhash, newbzzhash)
		}
	})
}

func TestApiDirUploadModify(t *testing.T) {
	testFileSystem(t, func(fs *FileSystem) {
		api := fs.api
		bzzhash, err := fs.Upload(filepath.Join("testdata", "test0"), "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		key := storage.Key(common.Hex2Bytes(bzzhash))
		key, err = api.Modify(key, "index.html", "", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		index, err := ioutil.ReadFile(filepath.Join("testdata", "test0", "index.html"))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		wg := &sync.WaitGroup{}
		hash, err := api.Store(bytes.NewReader(index), int64(len(index)), wg)
		wg.Wait()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		key, err = api.Modify(key, "index2.html", hash.Hex(), "text/html; charset=utf-8")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		key, err = api.Modify(key, "img/logo.png", hash.Hex(), "text/html; charset=utf-8")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		bzzhash = key.String()

		content := readPath(t, "testdata", "test0", "index.html")
		resp := testGet(t, api, bzzhash, "index2.html")
		exp := expResponse(content, "text/html; charset=utf-8", 0)
		checkResponse(t, resp, exp)

		resp = testGet(t, api, bzzhash, "img/logo.png")
		exp = expResponse(content, "text/html; charset=utf-8", 0)
		checkResponse(t, resp, exp)

		content = readPath(t, "testdata", "test0", "index.css")
		resp = testGet(t, api, bzzhash, "index.css")
		exp = expResponse(content, "text/css", 0)
		checkResponse(t, resp, exp)

		_, _, _, err = api.Get(key, "")
		if err == nil {
			t.Errorf("expected error: %v", err)
		}
	})
}

func TestApiDirUploadWithRootFile(t *testing.T) {
	testFileSystem(t, func(fs *FileSystem) {
		api := fs.api
		bzzhash, err := fs.Upload(filepath.Join("testdata", "test0"), "index.html")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		content := readPath(t, "testdata", "test0", "index.html")
		resp := testGet(t, api, bzzhash, "")
		exp := expResponse(content, "text/html; charset=utf-8", 0)
		checkResponse(t, resp, exp)
	})
}

func TestApiFileUpload(t *testing.T) {
	testFileSystem(t, func(fs *FileSystem) {
		api := fs.api
		bzzhash, err := fs.Upload(filepath.Join("testdata", "test0", "index.html"), "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		content := readPath(t, "testdata", "test0", "index.html")
		resp := testGet(t, api, bzzhash, "index.html")
		exp := expResponse(content, "text/html; charset=utf-8", 0)
		checkResponse(t, resp, exp)
	})
}

func TestApiFileUploadWithRootFile(t *testing.T) {
	testFileSystem(t, func(fs *FileSystem) {
		api := fs.api
		bzzhash, err := fs.Upload(filepath.Join("testdata", "test0", "index.html"), "index.html")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		content := readPath(t, "testdata", "test0", "index.html")
		resp := testGet(t, api, bzzhash, "")
		exp := expResponse(content, "text/html; charset=utf-8", 0)
		checkResponse(t, resp, exp)
	})
}
