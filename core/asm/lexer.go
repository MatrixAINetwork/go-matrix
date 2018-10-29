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

package asm

import (
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// stateFn is used through the lifetime of the
// lexer to parse the different values at the
// current state.
type stateFn func(*lexer) stateFn

// token is emitted when the lexer has discovered
// a new parsable token. These are delivered over
// the tokens channels of the lexer
type token struct {
	typ    tokenType
	lineno int
	text   string
}

// tokenType are the different types the lexer
// is able to parse and return.
type tokenType int

const (
	eof              tokenType = iota // end of file
	lineStart                         // emitted when a line starts
	lineEnd                           // emitted when a line ends
	invalidStatement                  // any invalid statement
	element                           // any element during element parsing
	label                             // label is emitted when a label is found
	labelDef                          // label definition is emitted when a new label is found
	number                            // number is emitted when a number is found
	stringValue                       // stringValue is emitted when a string has been found

	Numbers            = "1234567890"                                           // characters representing any decimal number
	HexadecimalNumbers = Numbers + "aAbBcCdDeEfF"                               // characters representing any hexadecimal
	Alpha              = "abcdefghijklmnopqrstuwvxyzABCDEFGHIJKLMNOPQRSTUWVXYZ" // characters representing alphanumeric
)

// String implements stringer
func (it tokenType) String() string {
	if int(it) > len(stringtokenTypes) {
		return "invalid"
	}
	return stringtokenTypes[it]
}

var stringtokenTypes = []string{
	eof:              "EOF",
	invalidStatement: "invalid statement",
	element:          "element",
	lineEnd:          "end of line",
	lineStart:        "new line",
	label:            "label",
	labelDef:         "label definition",
	number:           "number",
	stringValue:      "string",
}

// lexer is the basic construct for parsing
// source code and turning them in to tokens.
// Tokens are interpreted by the compiler.
type lexer struct {
	input string // input contains the source code of the program

	tokens chan token // tokens is used to deliver tokens to the listener
	state  stateFn    // the current state function

	lineno            int // current line number in the source file
	start, pos, width int // positions for lexing and returning value

	debug bool // flag for triggering debug output
}

// lex lexes the program by name with the given source. It returns a
// channel on which the tokens are delivered.
func Lex(name string, source []byte, debug bool) <-chan token {
	ch := make(chan token)
	l := &lexer{
		input:  string(source),
		tokens: ch,
		state:  lexLine,
		debug:  debug,
	}
	go func() {
		l.emit(lineStart)
		for l.state != nil {
			l.state = l.state(l)
		}
		l.emit(eof)
		close(l.tokens)
	}()

	return ch
}

// next returns the next rune in the program's source.
func (l *lexer) next() (rune rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return 0
	}
	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return rune
}

// backup backsup the last parsed element (multi-character)
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns the next rune but does not advance the seeker
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// ignore advances the seeker and ignores the value
func (l *lexer) ignore() {
	l.start = l.pos
}

// Accepts checks whether the given input matches the next rune
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}

	l.backup()

	return false
}

// acceptRun will continue to advance the seeker until valid
// can no longer be met.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// acceptRunUntil is the inverse of acceptRun and will continue
// to advance the seeker until the rune has been found.
func (l *lexer) acceptRunUntil(until rune) bool {
	// Continues running until a rune is found
	for i := l.next(); !strings.ContainsRune(string(until), i); i = l.next() {
		if i == 0 {
			return false
		}
	}

	return true
}

// blob returns the current value
func (l *lexer) blob() string {
	return l.input[l.start:l.pos]
}

// Emits a new token on to token channel for processing
func (l *lexer) emit(t tokenType) {
	token := token{t, l.lineno, l.blob()}

	if l.debug {
		fmt.Fprintf(os.Stderr, "%04d: (%-20v) %s\n", token.lineno, token.typ, token.text)
	}

	l.tokens <- token
	l.start = l.pos
}

// lexLine is state function for lexing lines
func lexLine(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == '\n':
			l.emit(lineEnd)
			l.ignore()
			l.lineno++

			l.emit(lineStart)
		case r == ';' && l.peek() == ';':
			return lexComment
		case isSpace(r):
			l.ignore()
		case isLetter(r) || r == '_':
			return lexElement
		case isNumber(r):
			return lexNumber
		case r == '@':
			l.ignore()
			return lexLabel
		case r == '"':
			return lexInsideString
		default:
			return nil
		}
	}
}

// lexComment parses the current position until the end
// of the line and discards the text.
func lexComment(l *lexer) stateFn {
	l.acceptRunUntil('\n')
	l.ignore()

	return lexLine
}

// lexLabel parses the current label, emits and returns
// the lex text state function to advance the parsing
// process.
func lexLabel(l *lexer) stateFn {
	l.acceptRun(Alpha + "_")

	l.emit(label)

	return lexLine
}

// lexInsideString lexes the inside of a string until
// until the state function finds the closing quote.
// It returns the lex text state function.
func lexInsideString(l *lexer) stateFn {
	if l.acceptRunUntil('"') {
		l.emit(stringValue)
	}

	return lexLine
}

func lexNumber(l *lexer) stateFn {
	acceptance := Numbers
	if l.accept("0") || l.accept("xX") {
		acceptance = HexadecimalNumbers
	}
	l.acceptRun(acceptance)

	l.emit(number)

	return lexLine
}

func lexElement(l *lexer) stateFn {
	l.acceptRun(Alpha + "_" + Numbers)

	if l.peek() == ':' {
		l.emit(labelDef)

		l.accept(":")
		l.ignore()
	} else {
		l.emit(element)
	}
	return lexLine
}

func isLetter(t rune) bool {
	return unicode.IsLetter(t)
}

func isSpace(t rune) bool {
	return unicode.IsSpace(t)
}

func isNumber(t rune) bool {
	return unicode.IsNumber(t)
}
