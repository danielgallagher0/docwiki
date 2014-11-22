package wikilang

import (
	"strings"
	"unicode"
)

const (
	Text = iota
	BoldDelimeter
	EmphasisDelimeter
	UnorderedListItem
	OrderedListItem
	LiteralText
	WikiLink
	Tag
	NewLine
	EndOfFile
)

const (
	BOLD_DELIMETER           = '*'
	EMPHASIS_DELIMETER       = '/'
	UNORDERED_LIST_ITEM_MARK = '-'
	ORDERED_LIST_ITEM_MARK   = '#'
	WIKILINK_OPEN            = '['
	WIKILINK_CLOSE           = ']'
	LITERAL_TEXT_OPEN        = '{'
	LITERAL_TEXT_CLOSE       = '}'
	TAG_OPEN                 = '<'
	TAG_CLOSE                = '>'
)

type Token struct {
	Type      int
	TextValue string
	IntValue  int
}

type Lexer struct {
	in  chan byte
	out chan Token
}

func NewLexer(i chan byte, o chan Token) Lexer {
	return Lexer{i, o}
}

const interrupters = "[{*/<\n"

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
	for b != 0 && !unicode.IsSpace(rune(b)) && !strings.Contains(interrupters, string(b)) {
		value = value + string(b)
		b = <-ch
	}

	if strings.Contains(interrupters, string(b)) {
		pending = b
		hasPending = true
	}

	return Token{Text, value, 0}, b == 0
}

func init() {
	lexers = make(map[byte]func(byte, chan byte) (Token, bool))

	lexers['*'] = singleByteToken(BoldDelimeter)
	lexers['/'] = singleByteToken(EmphasisDelimeter)
	lexers['-'] = singleByteToken(UnorderedListItem)
	lexers['#'] = singleByteToken(OrderedListItem)

	lexers['{'] = func(c byte, ch chan byte) (Token, bool) {
		nesting := 0
		value := ""
		b := <-ch
		for (nesting > 0 || b != '}') && b != 0 {
			if b == '{' {
				nesting++
			}

			if b == '}' {
				nesting--
			}

			value = value + string(b)
			b = <-ch
		}

		return Token{LiteralText, value, 0}, b == 0
	}
	lexers['['] = func(c byte, ch chan byte) (Token, bool) {
		value := ""
		b := <-ch
		for b != ']' && b != 0 {
			value = value + string(b)
			b = <-ch
		}

		return Token{WikiLink, value, 0}, b == 0
	}
	lexers['<'] = func(c byte, ch chan byte) (Token, bool) {
		value := string(c)
		b := <-ch
		for b != '>' && b != 0 {
			value = value + string(b)
			b = <-ch
		}
		if b != 0 {
			value = value + string('>')
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

func (l *Lexer) Lex() {
	eof := false
	b := <-l.in
	for !eof && b != 0 {
		if b == '\n' || !unicode.IsSpace(rune(b)) {
			f, ok := lexers[b]
			if ok {
				t, end := f(b, l.in)
				eof = end
				l.out <- t
			} else {
				t, end := defaultToken(b, l.in)
				eof = end
				l.out <- t
			}
		}

		if hasPending {
			b = pending
			hasPending = false
		} else if !eof {
			b = <-l.in
		}
	}

	l.out <- Token{EndOfFile, "", 0}
}
