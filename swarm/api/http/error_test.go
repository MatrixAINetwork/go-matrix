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

package http_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/matrix/go-matrix/swarm/testutil"
)

func TestError(t *testing.T) {

	srv := testutil.NewTestSwarmServer(t)
	defer srv.Close()

	var resp *http.Response
	var respbody []byte

	url := srv.URL + "/this_should_fail_as_no_bzz_protocol_present"
	resp, err := http.Get(url)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	respbody, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 400 && !strings.Contains(string(respbody), "Invalid URI &#34;/this_should_fail_as_no_bzz_protocol_present&#34;: unknown scheme") {
		t.Fatalf("Response body does not match, expected: %v, to contain: %v; received code %d, expected code: %d", string(respbody), "Invalid bzz URI: unknown scheme", 400, resp.StatusCode)
	}

	_, err = html.Parse(strings.NewReader(string(respbody)))
	if err != nil {
		t.Fatalf("HTML validation failed for error page returned!")
	}
}

func Test404Page(t *testing.T) {
	srv := testutil.NewTestSwarmServer(t)
	defer srv.Close()

	var resp *http.Response
	var respbody []byte

	url := srv.URL + "/bzz:/1234567890123456789012345678901234567890123456789012345678901234"
	resp, err := http.Get(url)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	respbody, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 404 || !strings.Contains(string(respbody), "404") {
		t.Fatalf("Invalid Status Code received, expected 404, got %d", resp.StatusCode)
	}

	_, err = html.Parse(strings.NewReader(string(respbody)))
	if err != nil {
		t.Fatalf("HTML validation failed for error page returned!")
	}
}

func Test500Page(t *testing.T) {
	srv := testutil.NewTestSwarmServer(t)
	defer srv.Close()

	var resp *http.Response
	var respbody []byte

	url := srv.URL + "/bzz:/thisShouldFailWith500Code"
	resp, err := http.Get(url)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	respbody, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 404 {
		t.Fatalf("Invalid Status Code received, expected 404, got %d", resp.StatusCode)
	}

	_, err = html.Parse(strings.NewReader(string(respbody)))
	if err != nil {
		t.Fatalf("HTML validation failed for error page returned!")
	}
}
func Test500PageWith0xHashPrefix(t *testing.T) {
	srv := testutil.NewTestSwarmServer(t)
	defer srv.Close()

	var resp *http.Response
	var respbody []byte

	url := srv.URL + "/bzz:/0xthisShouldFailWith500CodeAndAHelpfulMessage"
	resp, err := http.Get(url)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	respbody, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 404 {
		t.Fatalf("Invalid Status Code received, expected 404, got %d", resp.StatusCode)
	}

	if !strings.Contains(string(respbody), "The requested hash seems to be prefixed with") {
		t.Fatalf("Did not receive the expected error message")
	}

	_, err = html.Parse(strings.NewReader(string(respbody)))
	if err != nil {
		t.Fatalf("HTML validation failed for error page returned!")
	}
}

func TestJsonResponse(t *testing.T) {
	srv := testutil.NewTestSwarmServer(t)
	defer srv.Close()

	var resp *http.Response
	var respbody []byte

	url := srv.URL + "/bzz:/thisShouldFailWith500Code/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	defer resp.Body.Close()
	respbody, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 404 {
		t.Fatalf("Invalid Status Code received, expected 404, got %d", resp.StatusCode)
	}

	if !isJSON(string(respbody)) {
		t.Fatalf("Expected response to be JSON, received invalid JSON: %s", string(respbody))
	}

}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
