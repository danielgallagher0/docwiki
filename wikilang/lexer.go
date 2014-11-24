// Copyright (c) 2014, Daniel Gallagher
// Use of this source code is covered by the MIT License, the full
// text of which can be found in the LICENSE file.

package wikilang

import (
	"bytes"
	"unicode"
)

// These are the types of tokens that the lexer may generate.
const (
	Text = iota       // Standard text
	BoldDelimeter     // Beginning or end of boldness
	EmphasisDelimeter // Beginning or end of emphasis
	UnorderedListItem // Beginning of an item in an unordered list
	OrderedListItem   // Beginning of an item in an ordered list
	LiteralText       // Literal, preformatted text
	WikiLink          // Wiki markup for links
	Tag               // Embedded HTML markup
	NewLine           // New line (with indentation of next line)
	EndOfFile         // Special token for the end of list
)

// These are the special bytes that indicate switches in token types.
// Text tokens are additionally separated by spaces.
const (
	BOLD_MARK                = '*' // Denote boldness
	EMPHASIS_MARK            = '/' // Denote emphasis
	UNORDERED_LIST_ITEM_MARK = '-' // Denote item in unordered list
	ORDERED_LIST_ITEM_MARK   = '#' // Denote item in ordered list
	WIKILINK_OPEN            = '[' // Begin wiki markup
	WIKILINK_CLOSE           = ']' // End wiki markup
	LITERAL_TEXT_OPEN        = '{' // Begin literal text markup
	LITERAL_TEXT_CLOSE       = '}' // End literal text markup
	TAG_OPEN                 = '<' // Begin embedded HTML
	TAG_CLOSE                = '>' // End embedded HTML
)

// A Token is an atomic element of the wiki language.  Each token has
// a type, which the parser uses to structure the end result.  The
// Token also knows its own value, which is usually text, but may be
// an integer (representing indentation).
type Token struct {
	Type      int    // Categorization of the token
	TextValue string // Text value of the token
	IntValue  int    // Indentation of the token
}

// A Lexer converts a stream of bytes into a stream of tokens.  The
// end of the byte stream must be denoted by the special NUL character
// (0), which must not, therefore, appear in the main byte stream.
// When the end character is read, a special EndOfFile token is
// generated.
type Lexer struct {
	In  chan byte  // Channel to read bytes from
	Out chan Token // Channel to write tokens to
}

// NewLexer returns a lexer that communicates over the provided
// channels.
func NewLexer(i chan byte, o chan Token) Lexer {
	return Lexer{i, o}
}

var interrupters []byte

var lexers map[byte]func(byte, chan byte) (Token, bool)
var pending byte
var hasPending bool

func singleByteToken(t int) func(byte, chan byte) (Token, bool) {
	return func(c byte, ch chan byte) (Token, bool) {
		return Token{t, string(c), 0}, false
	}
}

func defaultToken(b byte, ch chan byte) (Token, bool) {
	value := string(b)
	b = <-ch
	for b != 0 && !unicode.IsSpace(rune(b)) && bytes.IndexByte(interrupters, b) < 0 {
		value = value + string(b)
		b = <-ch
	}

	if bytes.IndexByte(interrupters, b) >= 0 {
		pending = b
		hasPending = true
	}

	return Token{Text, value, 0}, b == 0
}

func init() {
	interrupters = []byte{
		WIKILINK_OPEN,
		LITERAL_TEXT_OPEN,
		BOLD_MARK,
		EMPHASIS_MARK,
		TAG_OPEN,
		'\n'}

	lexers = make(map[byte]func(byte, chan byte) (Token, bool))

	lexers[BOLD_MARK] = singleByteToken(BoldDelimeter)
	lexers[EMPHASIS_MARK] = singleByteToken(EmphasisDelimeter)
	lexers[UNORDERED_LIST_ITEM_MARK] = singleByteToken(UnorderedListItem)
	lexers[ORDERED_LIST_ITEM_MARK] = singleByteToken(OrderedListItem)

	lexers[LITERAL_TEXT_OPEN] = func(c byte, ch chan byte) (Token, bool) {
		nesting := 0
		value := ""
		b := <-ch
		for (nesting > 0 || b != LITERAL_TEXT_CLOSE) && b != 0 {
			if b == LITERAL_TEXT_OPEN {
				nesting++
			}

			if b == LITERAL_TEXT_CLOSE {
				nesting--
			}

			value = value + string(b)
			b = <-ch
		}

		return Token{LiteralText, value, 0}, b == 0
	}
	lexers[WIKILINK_OPEN] = func(c byte, ch chan byte) (Token, bool) {
		value := ""
		b := <-ch
		for b != WIKILINK_CLOSE && b != 0 {
			value = value + string(b)
			b = <-ch
		}

		return Token{WikiLink, value, 0}, b == 0
	}
	lexers[TAG_OPEN] = func(c byte, ch chan byte) (Token, bool) {
		value := string(c)
		b := <-ch
		for b != TAG_CLOSE && b != 0 {
			value = value + string(b)
			b = <-ch
		}
		if b != 0 {
			value = value + string(TAG_CLOSE)
		}

		return Token{Tag, value, 0}, b == 0
	}

	lexers['\n'] = func(b byte, ch chan byte) (Token, bool) {
		indent := 0
		c := <-ch
		for c == ' ' {
			indent++
			c = <-ch
		}

		if c != 0 {
			pending = c
			hasPending = true
		}

		return Token{NewLine, "", indent}, c == 0
	}

	pending = 0
	hasPending = false
}

// Lex runs a Lexer.  It reads all bytes in until it reaches the end
// mark, and writes out tokens.  When it is complete, it writes out
// an EndOfFile token.
//
// Text tokens are separated by whitespace.  Bold, emphasis, and list
// item delimeters are all single-byte tokens, only consuming the
// single byte.  Literal text, wiki markup, and embedded HTML tags
// consume all bytes from the open byte to the close byte.
func (l *Lexer) Lex() {
	eof := false
	b := <-l.In
	for !eof && b != 0 {
		if b == '\n' || !unicode.IsSpace(rune(b)) {
			f, ok := lexers[b]
			if ok {
				t, end := f(b, l.In)
				eof = end
				l.Out <- t
			} else {
				t, end := defaultToken(b, l.In)
				eof = end
				l.Out <- t
			}
		}

		if hasPending {
			b = pending
			hasPending = false
		} else if !eof {
			b = <-l.In
		}
	}

	l.Out <- Token{EndOfFile, "", 0}
}
