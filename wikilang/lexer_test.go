package wikilang

import (
	"strings"
	"testing"
)

func compareTokens(t *testing.T, actual, expected Token) {
	if actual.Type != expected.Type {
		t.Errorf("Expected token type %d", expected.Type)
		t.Errorf("  Actual token type %d", actual.Type)
	}

	if actual.TextValue != expected.TextValue {
		t.Errorf("Expected token text \"%s\"", expected.TextValue)
		t.Errorf("  Actual token text \"%s\"", actual.TextValue)
	}

	if actual.IntValue != expected.IntValue {
		t.Errorf("Expected token int value %d", expected.IntValue)
		t.Errorf("  Actual token int value %d", actual.IntValue)
	}
}

func TestSingleToken(t *testing.T) {
	for _, data := range [...]struct {
		data  string
		token Token
	}{
		{"abcdef", Token{Text, "abcdef", 0}},
		{"*", Token{BoldDelimeter, "*", 0}},
		{"/", Token{EmphasisDelimeter, "/", 0}},
		{"-", Token{UnorderedListItem, "-", 0}},
		{"#", Token{OrderedListItem, "#", 0}},
		{"{Some stuff that may *contain* /other/ tokens}",
			Token{LiteralText, "Some stuff that may *contain* /other/ tokens", 0}},
		{"{Multi-level {Literal Text}",
			Token{LiteralText, "Multi-level {Literal Text}", 0}},
		{"{Multi-level {Literal Text} in a sentence}",
			Token{LiteralText, "Multi-level {Literal Text} in a sentence", 0}},
		{"[WikiLink]", Token{WikiLink, "WikiLink", 0}},
		{"[ComplexLink:http://othersite.com]",
			Token{WikiLink, "ComplexLink:http://othersite.com", 0}},
		{"<a href=\"http://othersite.com\">",
			Token{Tag, "<a href=\"http://othersite.com\">", 0}},
		{"\n", Token{NewLine, "", 0}},
		{"\n    ", Token{NewLine, "", 4}},
	} {
		in := make(chan byte)
		out := make(chan Token)
		done := make(chan int)

		lexer := &Lexer{in, out}

		reader := strings.NewReader(data.data)
		go func() {
			ch, err := reader.ReadByte()
			for err == nil {
				in <- ch
				ch, err = reader.ReadByte()
			}

			in <- 0
		}()
		go lexer.Lex()

		go func() {
			actual := <-out
			compareTokens(t, actual, data.token)

			actual = <-out
			compareTokens(t, actual, Token{EndOfFile, "", 0})

			done <- 1
		}()

		<-done
	}
}

func TestMultiToken(t *testing.T) {
	for _, data := range [...]struct {
		data   string
		tokens *[]Token
	}{
		{"abc def", &[]Token{
			Token{Text, "abc", 0},
			Token{Text, "def", 0}}},
		{"*Some Bold Text*", &[]Token{
			Token{BoldDelimeter, "*", 0},
			Token{Text, "Some", 0},
			Token{Text, "Bold", 0},
			Token{Text, "Text", 0},
			Token{BoldDelimeter, "*", 0}}},
		{"Emphatic /text/", &[]Token{
			Token{Text, "Emphatic", 0},
			Token{EmphasisDelimeter, "/", 0},
			Token{Text, "text", 0},
			Token{EmphasisDelimeter, "/", 0}}},
		{"\n    - A new entry", &[]Token{
			Token{NewLine, "", 4},
			Token{UnorderedListItem, "-", 0},
			Token{Text, "A", 0},
			Token{Text, "new", 0},
			Token{Text, "entry", 0}}},
		{"\n    # Entry\n        # List\n    # Entry\n", &[]Token{
			Token{NewLine, "", 4},
			Token{OrderedListItem, "#", 0},
			Token{Text, "Entry", 0},
			Token{NewLine, "", 8},
			Token{OrderedListItem, "#", 0},
			Token{Text, "List", 0},
			Token{NewLine, "", 4},
			Token{OrderedListItem, "#", 0},
			Token{Text, "Entry", 0},
			Token{NewLine, "", 0}}},
		{"{Some stuff that may *contain* /other/ tokens}", &[]Token{
			Token{LiteralText, "Some stuff that may *contain* /other/ tokens", 0}}},
		{"Inside {Multi-level {Literal Text}}", &[]Token{
			Token{Text, "Inside", 0},
			Token{LiteralText, "Multi-level {Literal Text}", 0}}},
		{"Inter-word Mul{ti-level {Literal Text} in a sentence}", &[]Token{
			Token{Text, "Inter-word", 0},
			Token{Text, "Mul", 0},
			Token{LiteralText, "ti-level {Literal Text} in a sentence", 0}}},
		{"A [WikiLink]", &[]Token{
			Token{Text, "A", 0},
			Token{WikiLink, "WikiLink", 0}}},
		{"A more[ComplexLink:http://othersite.com]", &[]Token{
			Token{Text, "A", 0},
			Token{Text, "more", 0},
			Token{WikiLink, "ComplexLink:http://othersite.com", 0}}},
		{"Lots of    \tspaces: <a href=\"http://othersite.com\">", &[]Token{
			Token{Text, "Lots", 0},
			Token{Text, "of", 0},
			Token{Text, "spaces:", 0},
			Token{Tag, "<a href=\"http://othersite.com\">", 0}}},
	} {
		in := make(chan byte)
		out := make(chan Token)

		lexer := &Lexer{in, out}

		reader := strings.NewReader(data.data)
		go func() {
			ch, err := reader.ReadByte()
			for err == nil {
				in <- ch
				ch, err = reader.ReadByte()
			}

			in <- 0
		}()
		go lexer.Lex()

		for _, expected := range *data.tokens {
			actual := <-out
			compareTokens(t, actual, expected)
		}
		actual := <-out
		compareTokens(t, actual, Token{EndOfFile, "", 0})
	}
}
