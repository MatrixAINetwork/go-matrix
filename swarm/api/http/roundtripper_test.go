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

package http

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRoundTripper(t *testing.T) {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/plain")
			http.ServeContent(w, r, "", time.Unix(0, 0), strings.NewReader(r.RequestURI))
		} else {
			http.Error(w, "Method "+r.Method+" is not supported.", http.StatusMethodNotAllowed)
		}
	})

	srv := httptest.NewServer(serveMux)
	defer srv.Close()

	host, port, _ := net.SplitHostPort(srv.Listener.Addr().String())
	rt := &RoundTripper{Host: host, Port: port}
	trans := &http.Transport{}
	trans.RegisterProtocol("bzz", rt)
	client := &http.Client{Transport: trans}
	resp, err := client.Get("bzz://test.com/path")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
		return
	}

	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
		return
	}
	if string(content) != "/HTTP/1.1:/test.com/path" {
		t.Errorf("incorrect response from http server: expected '%v', got '%v'", "/HTTP/1.1:/test.com/path", string(content))
	}

}
