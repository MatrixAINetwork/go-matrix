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

package metrics

import (
	"time"

	"github.com/matrix/go-matrix/cmd/utils"
	"github.com/matrix/go-matrix/log"
	gmanmetrics "github.com/matrix/go-matrix/metrics"
	"github.com/matrix/go-matrix/metrics/influxdb"
	"gopkg.in/urfave/cli.v1"
)

var (
	metricsEnableInfluxDBExportFlag = cli.BoolFlag{
		Name:  "metrics.influxdb.export",
		Usage: "Enable metrics export/push to an external InfluxDB database",
	}
	metricsInfluxDBEndpointFlag = cli.StringFlag{
		Name:  "metrics.influxdb.endpoint",
		Usage: "Metrics InfluxDB endpoint",
		Value: "http://127.0.0.1:8086",
	}
	metricsInfluxDBDatabaseFlag = cli.StringFlag{
		Name:  "metrics.influxdb.database",
		Usage: "Metrics InfluxDB database",
		Value: "metrics",
	}
	metricsInfluxDBUsernameFlag = cli.StringFlag{
		Name:  "metrics.influxdb.username",
		Usage: "Metrics InfluxDB username",
		Value: "",
	}
	metricsInfluxDBPasswordFlag = cli.StringFlag{
		Name:  "metrics.influxdb.password",
		Usage: "Metrics InfluxDB password",
		Value: "",
	}
	// The `host` tag is part of every measurement sent to InfluxDB. Queries on tags are faster in InfluxDB.
	// It is used so that we can group all nodes and average a measurement across all of them, but also so
	// that we can select a specific node and inspect its measurements.
	// https://docs.influxdata.com/influxdb/v1.4/concepts/key_concepts/#tag-key
	metricsInfluxDBHostTagFlag = cli.StringFlag{
		Name:  "metrics.influxdb.host.tag",
		Usage: "Metrics InfluxDB `host` tag attached to all measurements",
		Value: "localhost",
	}
)

// Flags holds all command-line flags required for metrics collection.
var Flags = []cli.Flag{
	utils.MetricsEnabledFlag,
	metricsEnableInfluxDBExportFlag,
	metricsInfluxDBEndpointFlag, metricsInfluxDBDatabaseFlag, metricsInfluxDBUsernameFlag, metricsInfluxDBPasswordFlag, metricsInfluxDBHostTagFlag,
}

func Setup(ctx *cli.Context) {
	if gmanmetrics.Enabled {
		log.Info("Enabling swarm metrics collection")
		var (
			enableExport = ctx.GlobalBool(metricsEnableInfluxDBExportFlag.Name)
			endpoint     = ctx.GlobalString(metricsInfluxDBEndpointFlag.Name)
			database     = ctx.GlobalString(metricsInfluxDBDatabaseFlag.Name)
			username     = ctx.GlobalString(metricsInfluxDBUsernameFlag.Name)
			password     = ctx.GlobalString(metricsInfluxDBPasswordFlag.Name)
			hosttag      = ctx.GlobalString(metricsInfluxDBHostTagFlag.Name)
		)

		if enableExport {
			log.Info("Enabling swarm metrics export to InfluxDB")
			go influxdb.InfluxDBWithTags(gmanmetrics.DefaultRegistry, 10*time.Second, endpoint, database, username, password, "swarm.", map[string]string{
				"host": hosttag,
			})
		}
	}
}
