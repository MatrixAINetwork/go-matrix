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
package autorest

import (
	"fmt"
	"net/http"
)

const (
	// UndefinedStatusCode is used when HTTP status code is not available for an error.
	UndefinedStatusCode = 0
)

// DetailedError encloses a error with details of the package, method, and associated HTTP
// status code (if any).
type DetailedError struct {
	Original error

	// PackageType is the package type of the object emitting the error. For types, the value
	// matches that produced the the '%T' format specifier of the fmt package. For other elements,
	// such as functions, it is just the package name (e.g., "autorest").
	PackageType string

	// Method is the name of the method raising the error.
	Method string

	// StatusCode is the HTTP Response StatusCode (if non-zero) that led to the error.
	StatusCode interface{}

	// Message is the error message.
	Message string

	// Service Error is the response body of failed API in bytes
	ServiceError []byte
}

// NewError creates a new Error conforming object from the passed packageType, method, and
// message. message is treated as a format string to which the optional args apply.
func NewError(packageType string, method string, message string, args ...interface{}) DetailedError {
	return NewErrorWithError(nil, packageType, method, nil, message, args...)
}

// NewErrorWithResponse creates a new Error conforming object from the passed
// packageType, method, statusCode of the given resp (UndefinedStatusCode if
// resp is nil), and message. message is treated as a format string to which the
// optional args apply.
func NewErrorWithResponse(packageType string, method string, resp *http.Response, message string, args ...interface{}) DetailedError {
	return NewErrorWithError(nil, packageType, method, resp, message, args...)
}

// NewErrorWithError creates a new Error conforming object from the
// passed packageType, method, statusCode of the given resp (UndefinedStatusCode
// if resp is nil), message, and original error. message is treated as a format
// string to which the optional args apply.
func NewErrorWithError(original error, packageType string, method string, resp *http.Response, message string, args ...interface{}) DetailedError {
	if v, ok := original.(DetailedError); ok {
		return v
	}

	statusCode := UndefinedStatusCode
	if resp != nil {
		statusCode = resp.StatusCode
	}

	return DetailedError{
		Original:    original,
		PackageType: packageType,
		Method:      method,
		StatusCode:  statusCode,
		Message:     fmt.Sprintf(message, args...),
	}
}

// Error returns a formatted containing all available details (i.e., PackageType, Method,
// StatusCode, Message, and original error (if any)).
func (e DetailedError) Error() string {
	if e.Original == nil {
		return fmt.Sprintf("%s#%s: %s: StatusCode=%d", e.PackageType, e.Method, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("%s#%s: %s: StatusCode=%d -- Original Error: %v", e.PackageType, e.Method, e.Message, e.StatusCode, e.Original)
}
