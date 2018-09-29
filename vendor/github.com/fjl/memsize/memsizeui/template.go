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
package memsizeui

import (
	"html/template"
	"strconv"
	"sync"

	"github.com/fjl/memsize"
)

var (
	base         *template.Template // the "base" template
	baseInitOnce sync.Once
)

func baseInit() {
	base = template.Must(template.New("base").Parse(`<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>memsize</title>
		<style>
		body {
			 font-family: sans-serif;
		}
		button, .button {
			 display: inline-block;
			 font-weight: bold;
			 color: black;
			 text-decoration: none;
			 font-size: inherit;
			 padding: 3pt;
			 margin: 3pt;
			 background-color: #eee;
			 border: 1px solid #999;
			 border-radius: 2pt;
		}
		form.inline {
			display: inline-block;
		}
		</style>
	</head>
	<body>
		{{template "content" .}}
	</body>
</html>`))

	base.Funcs(template.FuncMap{
		"quote":     strconv.Quote,
		"humansize": memsize.HumanSize,
	})

	template.Must(base.New("rootbuttons").Parse(`
<a class="button" href="{{$.Link ""}}">Overview</a>
{{- range $root := .Roots -}}
<form class="inline" method="POST" action="{{$.Link "scan?root=" $root}}">
	<button type="submit">Scan {{quote $root}}</button>
</form>
{{- end -}}`))
}

func contentTemplate(source string) *template.Template {
	baseInitOnce.Do(baseInit)
	t := template.Must(base.Clone())
	template.Must(t.New("content").Parse(source))
	return t
}

var rootTemplate = contentTemplate(`
<h1>Memsize</h1>
{{template "rootbuttons" .}}
<hr/>
<h3>Reports</h3>
<ul>
	{{range .Reports}}
		<li><a href="{{printf "%d" | $.Link "report/"}}">{{quote .RootName}} @ {{.Date}}</a></li>
	{{else}}
		No reports yet, hit a scan button to create one.
	{{end}}
</ul>
`)

var notFoundTemplate = contentTemplate(`
<h1>{{.Data}}</h1>
{{template "rootbuttons" .}}
`)

var reportTemplate = contentTemplate(`
{{- $report := .Data -}}
<h1>Memsize Report {{$report.ID}}</h1>
<form method="POST" action="{{$.Link "scan?root=" $report.RootName}}">
	<a class="button" href="{{$.Link ""}}">Overview</a>
	<button type="submit">Scan Again</button>
</form>
<pre>
Root: {{quote $report.RootName}}
Date: {{$report.Date}}
Duration: {{$report.Duration}}
Bitmap Size: {{$report.Sizes.BitmapSize | humansize}}
Bitmap Utilization: {{$report.Sizes.BitmapUtilization}}
</pre>
<hr/>
<pre>
{{$report.Sizes.Report}}
</pre>
`)
