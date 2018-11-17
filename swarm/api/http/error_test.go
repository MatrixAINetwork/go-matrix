// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

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
