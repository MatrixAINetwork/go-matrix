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
package cli

// CommandCategories is a slice of *CommandCategory.
type CommandCategories []*CommandCategory

// CommandCategory is a category containing commands.
type CommandCategory struct {
	Name     string
	Commands Commands
}

func (c CommandCategories) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

func (c CommandCategories) Len() int {
	return len(c)
}

func (c CommandCategories) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// AddCommand adds a command to a category.
func (c CommandCategories) AddCommand(category string, command Command) CommandCategories {
	for _, commandCategory := range c {
		if commandCategory.Name == category {
			commandCategory.Commands = append(commandCategory.Commands, command)
			return c
		}
	}
	return append(c, &CommandCategory{Name: category, Commands: []Command{command}})
}

// VisibleCommands returns a slice of the Commands with Hidden=false
func (c *CommandCategory) VisibleCommands() []Command {
	ret := []Command{}
	for _, command := range c.Commands {
		if !command.Hidden {
			ret = append(ret, command)
		}
	}
	return ret
}
