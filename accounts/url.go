// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package accounts

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// URL represents the canonical identification URL of a wallet or account.
//
// It is a simplified version of url.URL, with the important limitations (which
// are considered features here) that it contains value-copyable components only,
// as well as that it doesn't do any URL encoding/decoding of special characters.
//
// The former is important to allow an account to be copied without leaving live
// references to the original version, whereas the latter is important to ensure
// one single canonical form opposed to many allowed ones by the RFC 3986 spec.
//
// As such, these URLs should not be used outside of the scope of an Matrix
// wallet or account.
type URL struct {
	Scheme string // Protocol scheme to identify a capable account backend
	Path   string // Path for the backend to identify a unique entity
}

// parseURL converts a user supplied URL into the accounts specific structure.
func parseURL(url string) (URL, error) {
	parts := strings.Split(url, "://")
	if len(parts) != 2 || parts[0] == "" {
		return URL{}, errors.New("protocol scheme missing")
	}
	return URL{
		Scheme: parts[0],
		Path:   parts[1],
	}, nil
}

// String implements the stringer interface.
func (u URL) String() string {
	if u.Scheme != "" {
		return fmt.Sprintf("%s://%s", u.Scheme, u.Path)
	}
	return u.Path
}

// TerminalString implements the log.TerminalStringer interface.
func (u URL) TerminalString() string {
	url := u.String()
	if len(url) > 32 {
		return url[:31] + "…"
	}
	return url
}

// MarshalJSON implements the json.Marshaller interface.
func (u URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

// UnmarshalJSON parses url.
func (u *URL) UnmarshalJSON(input []byte) error {
	var textUrl string
	err := json.Unmarshal(input, &textUrl)
	if err != nil {
		return err
	}
	url, err := parseURL(textUrl)
	if err != nil {
		return err
	}
	u.Scheme = url.Scheme
	u.Path = url.Path
	return nil
}

// Cmp compares x and y and returns:
//
//   -1 if x <  y
//    0 if x == y
//   +1 if x >  y
//
func (u URL) Cmp(url URL) int {
	if u.Scheme == url.Scheme {
		return strings.Compare(u.Path, url.Path)
	}
	return strings.Compare(u.Scheme, url.Scheme)
}
