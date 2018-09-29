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
package storage

import (
	"fmt"
	"net/http"
	"net/url"
)

// ServiceProperties represents the storage account service properties
type ServiceProperties struct {
	Logging       *Logging
	HourMetrics   *Metrics
	MinuteMetrics *Metrics
	Cors          *Cors
}

// Logging represents the Azure Analytics Logging settings
type Logging struct {
	Version         string
	Delete          bool
	Read            bool
	Write           bool
	RetentionPolicy *RetentionPolicy
}

// RetentionPolicy indicates if retention is enabled and for how many days
type RetentionPolicy struct {
	Enabled bool
	Days    *int
}

// Metrics provide request statistics.
type Metrics struct {
	Version         string
	Enabled         bool
	IncludeAPIs     *bool
	RetentionPolicy *RetentionPolicy
}

// Cors includes all the CORS rules
type Cors struct {
	CorsRule []CorsRule
}

// CorsRule includes all settings for a Cors rule
type CorsRule struct {
	AllowedOrigins  string
	AllowedMethods  string
	MaxAgeInSeconds int
	ExposedHeaders  string
	AllowedHeaders  string
}

func (c Client) getServiceProperties(service string, auth authentication) (*ServiceProperties, error) {
	query := url.Values{
		"restype": {"service"},
		"comp":    {"properties"},
	}
	uri := c.getEndpoint(service, "", query)
	headers := c.getStandardHeaders()

	resp, err := c.exec(http.MethodGet, uri, headers, nil, auth)
	if err != nil {
		return nil, err
	}
	defer resp.body.Close()

	if err := checkRespCode(resp.statusCode, []int{http.StatusOK}); err != nil {
		return nil, err
	}

	var out ServiceProperties
	err = xmlUnmarshal(resp.body, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (c Client) setServiceProperties(props ServiceProperties, service string, auth authentication) error {
	query := url.Values{
		"restype": {"service"},
		"comp":    {"properties"},
	}
	uri := c.getEndpoint(service, "", query)

	// Ideally, StorageServiceProperties would be the output struct
	// This is to avoid golint stuttering, while generating the correct XML
	type StorageServiceProperties struct {
		Logging       *Logging
		HourMetrics   *Metrics
		MinuteMetrics *Metrics
		Cors          *Cors
	}
	input := StorageServiceProperties{
		Logging:       props.Logging,
		HourMetrics:   props.HourMetrics,
		MinuteMetrics: props.MinuteMetrics,
		Cors:          props.Cors,
	}

	body, length, err := xmlMarshal(input)
	if err != nil {
		return err
	}

	headers := c.getStandardHeaders()
	headers["Content-Length"] = fmt.Sprintf("%v", length)

	resp, err := c.exec(http.MethodPut, uri, headers, body, auth)
	if err != nil {
		return err
	}
	defer readAndCloseBody(resp.body)

	return checkRespCode(resp.statusCode, []int{http.StatusAccepted})
}
