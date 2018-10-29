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
package librato

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const Operations = "operations"
const OperationsShort = "ops"

type LibratoClient struct {
	Email, Token string
}

// property strings
const (
	// display attributes
	Color             = "color"
	DisplayMax        = "display_max"
	DisplayMin        = "display_min"
	DisplayUnitsLong  = "display_units_long"
	DisplayUnitsShort = "display_units_short"
	DisplayStacked    = "display_stacked"
	DisplayTransform  = "display_transform"
	// special gauge display attributes
	SummarizeFunction = "summarize_function"
	Aggregate         = "aggregate"

	// metric keys
	Name        = "name"
	Period      = "period"
	Description = "description"
	DisplayName = "display_name"
	Attributes  = "attributes"

	// measurement keys
	MeasureTime = "measure_time"
	Source      = "source"
	Value       = "value"

	// special gauge keys
	Count      = "count"
	Sum        = "sum"
	Max        = "max"
	Min        = "min"
	SumSquares = "sum_squares"

	// batch keys
	Counters = "counters"
	Gauges   = "gauges"

	MetricsPostUrl = "https://metrics-api.librato.com/v1/metrics"
)

type Measurement map[string]interface{}
type Metric map[string]interface{}

type Batch struct {
	Gauges      []Measurement `json:"gauges,omitempty"`
	Counters    []Measurement `json:"counters,omitempty"`
	MeasureTime int64         `json:"measure_time"`
	Source      string        `json:"source"`
}

func (c *LibratoClient) PostMetrics(batch Batch) (err error) {
	var (
		js   []byte
		req  *http.Request
		resp *http.Response
	)

	if len(batch.Counters) == 0 && len(batch.Gauges) == 0 {
		return nil
	}

	if js, err = json.Marshal(batch); err != nil {
		return
	}

	if req, err = http.NewRequest("POST", MetricsPostUrl, bytes.NewBuffer(js)); err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.Email, c.Token)

	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		var body []byte
		if body, err = ioutil.ReadAll(resp.Body); err != nil {
			body = []byte(fmt.Sprintf("(could not fetch response body for error: %s)", err))
		}
		err = fmt.Errorf("Unable to post to Librato: %d %s %s", resp.StatusCode, resp.Status, string(body))
	}
	return
}
