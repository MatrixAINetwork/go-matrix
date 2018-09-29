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
/*
Package autorest implements an HTTP request pipeline suitable for use across multiple go-routines
and provides the shared routines relied on by AutoRest (see https://github.com/Azure/autorest/)
generated Go code.

The package breaks sending and responding to HTTP requests into three phases: Preparing, Sending,
and Responding. A typical pattern is:

  req, err := Prepare(&http.Request{},
    token.WithAuthorization())

  resp, err := Send(req,
    WithLogging(logger),
    DoErrorIfStatusCode(http.StatusInternalServerError),
    DoCloseIfError(),
    DoRetryForAttempts(5, time.Second))

  err = Respond(resp,
    ByDiscardingBody(),
    ByClosing())

Each phase relies on decorators to modify and / or manage processing. Decorators may first modify
and then pass the data along, pass the data first and then modify the result, or wrap themselves
around passing the data (such as a logger might do). Decorators run in the order provided. For
example, the following:

  req, err := Prepare(&http.Request{},
    WithBaseURL("https://microsoft.com/"),
    WithPath("a"),
    WithPath("b"),
    WithPath("c"))

will set the URL to:

  https://microsoft.com/a/b/c

Preparers and Responders may be shared and re-used (assuming the underlying decorators support
sharing and re-use). Performant use is obtained by creating one or more Preparers and Responders
shared among multiple go-routines, and a single Sender shared among multiple sending go-routines,
all bound together by means of input / output channels.

Decorators hold their passed state within a closure (such as the path components in the example
above). Be careful to share Preparers and Responders only in a context where such held state
applies. For example, it may not make sense to share a Preparer that applies a query string from a
fixed set of values. Similarly, sharing a Responder that reads the response body into a passed
struct (e.g., ByUnmarshallingJson) is likely incorrect.

Lastly, the Swagger specification (https://swagger.io) that drives AutoRest
(https://github.com/Azure/autorest/) precisely defines two date forms: date and date-time. The
github.com/Azure/go-autorest/autorest/date package provides time.Time derivations to ensure
correct parsing and formatting.

Errors raised by autorest objects and methods will conform to the autorest.Error interface.

See the included examples for more detail. For details on the suggested use of this package by
generated clients, see the Client described below.
*/
package autorest

import (
	"net/http"
	"time"
)

const (
	// HeaderLocation specifies the HTTP Location header.
	HeaderLocation = "Location"

	// HeaderRetryAfter specifies the HTTP Retry-After header.
	HeaderRetryAfter = "Retry-After"
)

// ResponseHasStatusCode returns true if the status code in the HTTP Response is in the passed set
// and false otherwise.
func ResponseHasStatusCode(resp *http.Response, codes ...int) bool {
	return containsInt(codes, resp.StatusCode)
}

// GetLocation retrieves the URL from the Location header of the passed response.
func GetLocation(resp *http.Response) string {
	return resp.Header.Get(HeaderLocation)
}

// GetRetryAfter extracts the retry delay from the Retry-After header of the passed response. If
// the header is absent or is malformed, it will return the supplied default delay time.Duration.
func GetRetryAfter(resp *http.Response, defaultDelay time.Duration) time.Duration {
	retry := resp.Header.Get(HeaderRetryAfter)
	if retry == "" {
		return defaultDelay
	}

	d, err := time.ParseDuration(retry + "s")
	if err != nil {
		return defaultDelay
	}

	return d
}

// NewPollingRequest allocates and returns a new http.Request to poll for the passed response.
func NewPollingRequest(resp *http.Response, cancel <-chan struct{}) (*http.Request, error) {
	location := GetLocation(resp)
	if location == "" {
		return nil, NewErrorWithResponse("autorest", "NewPollingRequest", resp, "Location header missing from response that requires polling")
	}

	req, err := Prepare(&http.Request{Cancel: cancel},
		AsGet(),
		WithBaseURL(location))
	if err != nil {
		return nil, NewErrorWithError(err, "autorest", "NewPollingRequest", nil, "Failure creating poll request to %s", location)
	}

	return req, nil
}
