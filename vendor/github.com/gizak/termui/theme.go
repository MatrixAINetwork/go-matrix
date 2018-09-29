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
// Copyright 2017 Zack Guo <zack.y.guo@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package termui

import "strings"

/*
// A ColorScheme represents the current look-and-feel of the dashboard.
type ColorScheme struct {
	BodyBg            Attribute
	BlockBg           Attribute
	HasBorder         bool
	BorderFg          Attribute
	BorderBg          Attribute
	BorderLabelTextFg Attribute
	BorderLabelTextBg Attribute
	ParTextFg         Attribute
	ParTextBg         Attribute
	SparklineLine     Attribute
	SparklineTitle    Attribute
	GaugeBar          Attribute
	GaugePercent      Attribute
	LineChartLine     Attribute
	LineChartAxes     Attribute
	ListItemFg        Attribute
	ListItemBg        Attribute
	BarChartBar       Attribute
	BarChartText      Attribute
	BarChartNum       Attribute
	MBarChartBar      Attribute
	MBarChartText     Attribute
	MBarChartNum      Attribute
	TabActiveBg		  Attribute
}

// default color scheme depends on the user's terminal setting.
var themeDefault = ColorScheme{HasBorder: true}

var themeHelloWorld = ColorScheme{
	BodyBg:            ColorBlack,
	BlockBg:           ColorBlack,
	HasBorder:         true,
	BorderFg:          ColorWhite,
	BorderBg:          ColorBlack,
	BorderLabelTextBg: ColorBlack,
	BorderLabelTextFg: ColorGreen,
	ParTextBg:         ColorBlack,
	ParTextFg:         ColorWhite,
	SparklineLine:     ColorMagenta,
	SparklineTitle:    ColorWhite,
	GaugeBar:          ColorRed,
	GaugePercent:      ColorWhite,
	LineChartLine:     ColorYellow | AttrBold,
	LineChartAxes:     ColorWhite,
	ListItemBg:        ColorBlack,
	ListItemFg:        ColorYellow,
	BarChartBar:       ColorRed,
	BarChartNum:       ColorWhite,
	BarChartText:      ColorCyan,
	MBarChartBar:      ColorRed,
	MBarChartNum:      ColorWhite,
	MBarChartText:     ColorCyan,
	TabActiveBg:	   ColorMagenta,
}

var theme = themeDefault // global dep

// Theme returns the currently used theme.
func Theme() ColorScheme {
	return theme
}

// SetTheme sets a new, custom theme.
func SetTheme(newTheme ColorScheme) {
	theme = newTheme
}

// UseTheme sets a predefined scheme. Currently available: "hello-world" and
// "black-and-white".
func UseTheme(th string) {
	switch th {
	case "helloworld":
		theme = themeHelloWorld
	default:
		theme = themeDefault
	}
}
*/

var ColorMap = map[string]Attribute{
	"fg":           ColorWhite,
	"bg":           ColorDefault,
	"border.fg":    ColorWhite,
	"label.fg":     ColorGreen,
	"par.fg":       ColorYellow,
	"par.label.bg": ColorWhite,
}

func ThemeAttr(name string) Attribute {
	return lookUpAttr(ColorMap, name)
}

func lookUpAttr(clrmap map[string]Attribute, name string) Attribute {

	a, ok := clrmap[name]
	if ok {
		return a
	}

	ns := strings.Split(name, ".")
	for i := range ns {
		nn := strings.Join(ns[i:len(ns)], ".")
		a, ok = ColorMap[nn]
		if ok {
			break
		}
	}

	return a
}

// 0<=r,g,b <= 5
func ColorRGB(r, g, b int) Attribute {
	within := func(n int) int {
		if n < 0 {
			return 0
		}

		if n > 5 {
			return 5
		}

		return n
	}

	r, b, g = within(r), within(b), within(g)
	return Attribute(0x0f + 36*r + 6*g + b)
}
