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
package otto

import (
	"strconv"
)

func (self *_runtime) cmpl_evaluate_nodeProgram(node *_nodeProgram, eval bool) Value {
	if !eval {
		self.enterGlobalScope()
		defer func() {
			self.leaveScope()
		}()
	}
	self.cmpl_functionDeclaration(node.functionList)
	self.cmpl_variableDeclaration(node.varList)
	self.scope.frame.file = node.file
	return self.cmpl_evaluate_nodeStatementList(node.body)
}

func (self *_runtime) cmpl_call_nodeFunction(function *_object, stash *_fnStash, node *_nodeFunctionLiteral, this Value, argumentList []Value) Value {

	indexOfParameterName := make([]string, len(argumentList))
	// function(abc, def, ghi)
	// indexOfParameterName[0] = "abc"
	// indexOfParameterName[1] = "def"
	// indexOfParameterName[2] = "ghi"
	// ...

	argumentsFound := false
	for index, name := range node.parameterList {
		if name == "arguments" {
			argumentsFound = true
		}
		value := Value{}
		if index < len(argumentList) {
			value = argumentList[index]
			indexOfParameterName[index] = name
		}
		// strict = false
		self.scope.lexical.setValue(name, value, false)
	}

	if !argumentsFound {
		arguments := self.newArgumentsObject(indexOfParameterName, stash, len(argumentList))
		arguments.defineProperty("callee", toValue_object(function), 0101, false)
		stash.arguments = arguments
		// strict = false
		self.scope.lexical.setValue("arguments", toValue_object(arguments), false)
		for index, _ := range argumentList {
			if index < len(node.parameterList) {
				continue
			}
			indexAsString := strconv.FormatInt(int64(index), 10)
			arguments.defineProperty(indexAsString, argumentList[index], 0111, false)
		}
	}

	self.cmpl_functionDeclaration(node.functionList)
	self.cmpl_variableDeclaration(node.varList)

	result := self.cmpl_evaluate_nodeStatement(node.body)
	if result.kind == valueResult {
		return result
	}

	return Value{}
}

func (self *_runtime) cmpl_functionDeclaration(list []*_nodeFunctionLiteral) {
	executionContext := self.scope
	eval := executionContext.eval
	stash := executionContext.variable

	for _, function := range list {
		name := function.name
		value := self.cmpl_evaluate_nodeExpression(function)
		if !stash.hasBinding(name) {
			stash.createBinding(name, eval == true, value)
		} else {
			// TODO 10.5.5.e
			stash.setBinding(name, value, false) // TODO strict
		}
	}
}

func (self *_runtime) cmpl_variableDeclaration(list []string) {
	executionContext := self.scope
	eval := executionContext.eval
	stash := executionContext.variable

	for _, name := range list {
		if !stash.hasBinding(name) {
			stash.createBinding(name, eval == true, Value{}) // TODO strict?
		}
	}
}
