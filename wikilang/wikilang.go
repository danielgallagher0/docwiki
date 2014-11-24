// Copyright (c) 2014, Daniel Gallagher
// Use of this source code is covered by the MIT License, the full
// text of which can be found in the LICENSE file.

// Package wikilang contains the lexer, parser, and HTML generator for
// the docwiki language.  The language is similar to reStructuredtext
// in that you write some ASCII and the HTML appears to follow the
// same structure.
//
// For the general case, just use WikiToHtml(string)string to convert
// a body of wiki text to the corresponding HTML.  If you need finer
// control over the steps of translation, the package also exposes the
// Lexer, Parser, and HtmlGen types.
//
// The language allows bold and emphasized text, links (internal wiki
// links, external links, and links to Doxygen), and lists.  For a
// full description, see the DocWikiLang page in the installed and
// running wiki.
//
// The Lexer, Parser, and Generator are all designed to work in
// independent goroutines, communicating over a set of channels.
package wikilang

import (
	"regexp"
	"strings"
)

// WikiToHtml converts a string of wiki text into a string of
// equivalent HTML.
func WikiToHtml(body string) string {
	data := make(chan byte)
	tokens := make(chan Token)
	trees := make(chan ParseTree)
	result := make(chan string)

	lexer := NewLexer(data, tokens)
	parser := NewParser(tokens, trees)
	gen := NewHtmlGen(trees, result)

	go func() {
		for _, c := range []byte(body) {
			data <- c
		}
		data <- 0
	}()
	go lexer.Lex()
	go parser.Parse()
	go gen.Generate()

	return <-result
}

// WikiCase converts a string from PascalCase into a list of words
// separated by spaces.
//     WikiCase("PascalCase") => "Pascal Case"
func WikiCase(link string) string {
	addSpaces := regexp.MustCompile("[[:upper:]]")
	splitString := strings.Trim(addSpaces.ReplaceAllStringFunc(link, func(s string) string {
		return " " + s
	}), " ")

	consolidateSpaces := regexp.MustCompile("[[:space:]]+")
	return consolidateSpaces.ReplaceAllString(splitString, " ")
}
