// Copyright (c) 2014, Daniel Gallagher
// Use of this source code is covered by the MIT License, the full
// text of which can be found in the LICENSE file.

package wikilang

import (
	"testing"
)

func compareFlattenedParseTrees(t *testing.T, actual, expected string) {
	if actual != expected {
		t.Errorf("Expected: \"%s\"", expected)
		t.Errorf("  Actual: \"%s\"", actual)
	}
}

func compareParseTrees(t *testing.T, actual, expected ParseTree) {
	compareFlattenedParseTrees(t, actual.String(), expected.String())
}

func runParserTest(t *testing.T, tokens []Token, expected string) {
	in := make(chan Token)
	out := make(chan ParseTree)
	done := make(chan int)

	parser := Parser{in, out, nil}

	go func() {
		for _, token := range tokens {
			in <- token
		}
	}()
	go parser.Parse()

	go func() {
		actual := ""
		count := 0
		for {
			count++
			tree := <-out
			if len(tree.Nodes) == 0 {
				break
			}
			actual = actual + tree.String()
		}
		if actual != expected {
			t.Errorf("Expected: \"%s\"", expected)
			t.Errorf("  Actual: \"%s\"", actual)
		}

		done <- 1
	}()

	<-done
}

func TestEmptyParseTree(t *testing.T) {
	token := Token{EndOfFile, "", 0}

	in := make(chan Token)
	out := make(chan ParseTree)
	done := make(chan int)

	parser := Parser{in, out, nil}

	go func() {
		in <- token
	}()
	go parser.Parse()

	go func() {
		actual := <-out
		compareParseTrees(t, actual, ParseTree{})

		done <- 1
	}()

	<-done
}

func TestSimpleParseTree(t *testing.T) {
	for _, data := range [...]struct {
		tokens        []Token
		flattenedTree string
	}{
		{
			[]Token{
				Token{Text, "abcdef", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: abcdef} ]} ]",
		},
		{
			[]Token{
				Token{Tag, "<a href=\"/Something\">", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: <a href=\"/Something\">} ]} ]",
		},
		{
			[]Token{
				Token{WikiLink, "WikiLink", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Tag a (href=/view/WikiLink ) [{Text: Wiki Link} ]} ]} ]",
		},
	} {
		runParserTest(t, data.tokens, data.flattenedTree)
	}
}

func TestSingleParagraphParseTree(t *testing.T) {
	for _, data := range [...]struct {
		tokens        []Token
		flattenedTree string
	}{
		{
			[]Token{
				Token{Text, "abc", 0},
				Token{Text, "def", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: abc def} ]} ]",
		},
		{
			[]Token{
				Token{Text, "Manual", 0},
				Token{Text, "Link:", 0},
				Token{Tag, "<a href=\"/Something\">", 0},
				Token{Text, "To", 0},
				Token{Text, "Here", 0},
				Token{Tag, "</a>", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: Manual Link: <a href=\"/Something\"> To Here </a>} ]} ]",
		},
		{
			[]Token{
				Token{Text, "Talking", 0},
				Token{Text, "about", 0},
				Token{Text, "a", 0},
				Token{WikiLink, "WikiLink", 0},
				Token{Text, "here", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: Talking about a} " +
				"{Tag a (href=/view/WikiLink ) [{Text: Wiki Link} ]} " +
				"{Text: here} ]} ]",
		},
		{
			[]Token{
				Token{Text, "Some", 0},
				Token{BoldDelimeter, "*", 0},
				Token{Text, "bold", 0},
				Token{Text, "text", 0},
				Token{BoldDelimeter, "*", 0},
				Token{Text, "in", 0},
				Token{EmphasisDelimeter, "/", 0},
				Token{Text, "a", 0},
				Token{Text, "paragraph", 0},
				Token{Text, "with", 0},
				Token{Text, "emphasized", 0},
				Token{EmphasisDelimeter, "/", 0},
				Token{Text, "text", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: Some} " +
				"{Tag b () [{Text: bold text} ]} " +
				"{Text: in} " +
				"{Tag em () [{Text: a paragraph with emphasized} ]} " +
				"{Text: text} ]} ]",
		},
		{
			[]Token{
				Token{Text, "Some", 0},
				Token{NewLine, "", 0},
				Token{BoldDelimeter, "*", 0},
				Token{Text, "bold", 0},
				Token{NewLine, "", 0},
				Token{Text, "text", 0},
				Token{BoldDelimeter, "*", 0},
				Token{Text, "in", 0},
				Token{EmphasisDelimeter, "/", 0},
				Token{Text, "a", 0},
				Token{Text, "paragraph", 0},
				Token{NewLine, "", 0},
				Token{Text, "with", 0},
				Token{Text, "emphasized", 0},
				Token{EmphasisDelimeter, "/", 0},
				Token{Text, "text", 0},
				Token{NewLine, "", 0},
				Token{Text, "over", 0},
				Token{Text, "multiple", 0},
				Token{Text, "lines", 0},
				Token{NewLine, "", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: Some} " +
				"{Tag b () [{Text: bold text} ]} " +
				"{Text: in} " +
				"{Tag em () [{Text: a paragraph with emphasized} ]} " +
				"{Text: text over multiple lines} ]} ]",
		},
	} {
		runParserTest(t, data.tokens, data.flattenedTree)
	}
}

func TestListParseTree(t *testing.T) {
	for _, data := range [...]struct {
		tokens        []Token
		flattenedTree string
	}{
		{
			[]Token{
				Token{UnorderedListItem, "-", 0},
				Token{Text, "abc", 0},
				Token{NewLine, "", 0},
				Token{UnorderedListItem, "-", 0},
				Token{Text, "def", 0},
				Token{NewLine, "", 0},
				Token{Text, "ghi", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag ul () [{Tag li () [{Text: abc} ]} " +
				"{Tag li () [{Text: def ghi} ]} ]} ]",
		},
		{
			[]Token{
				Token{OrderedListItem, "#", 0},
				Token{Text, "abc", 0},
				Token{NewLine, "", 0},
				Token{OrderedListItem, "#", 0},
				Token{Text, "def", 0},
				Token{NewLine, "", 0},
				Token{Text, "ghi", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag ol () [{Tag li () [{Text: abc} ]} " +
				"{Tag li () [{Text: def ghi} ]} ]} ]",
		},
	} {
		runParserTest(t, data.tokens, data.flattenedTree)
	}
}

func TestMultipleListParseTree(t *testing.T) {
	for _, data := range [...]struct {
		tokens        []Token
		flattenedTree string
	}{
		{
			[]Token{
				Token{UnorderedListItem, "-", 0},
				Token{Text, "abc", 0},
				Token{NewLine, "", 0},
				Token{OrderedListItem, "#", 0},
				Token{Text, "def", 0},
				Token{NewLine, "", 0},
				Token{Text, "ghi", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag ul () [{Tag li () [{Text: abc} ]} ]} ]" +
				"[{Tag ol () [{Tag li () [{Text: def ghi} ]} ]} ]",
		},
		{
			[]Token{
				Token{OrderedListItem, "#", 0},
				Token{Text, "abc", 0},
				Token{NewLine, "", 0},
				Token{UnorderedListItem, "-", 0},
				Token{Text, "def", 0},
				Token{NewLine, "", 0},
				Token{Text, "ghi", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag ol () [{Tag li () [{Text: abc} ]} ]} ]" +
				"[{Tag ul () [{Tag li () [{Text: def ghi} ]} ]} ]",
		},
	} {
		runParserTest(t, data.tokens, data.flattenedTree)
	}
}

func TestNestedParagraph(t *testing.T) {
	for _, data := range [...]struct {
		tokens        []Token
		flattenedTree string
	}{
		{
			[]Token{
				Token{Text, "Introductory", 0},
				Token{Text, "sentence", 0},
				Token{Text, "for", 0},
				Token{Text, "a", 0},
				Token{Text, "list:", 0},
				Token{NewLine, "", 4},
				Token{UnorderedListItem, "-", 0},
				Token{Text, "abc", 0},
				Token{NewLine, "", 4},
				Token{UnorderedListItem, "-", 0},
				Token{Text, "def", 0},
				Token{NewLine, "", 4},
				Token{Text, "ghi", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: Introductory sentence for a list:} " +
				"{Tag ul () [{Tag li () [{Text: abc} ]} " +
				"{Tag li () [{Text: def ghi} ]} ]} ]} ]",
		},
		{
			[]Token{
				Token{Text, "Beginning", 0},
				Token{Text, "paragraph", 0},
				Token{NewLine, "", 4},
				Token{Text, "Nested", 0},
				Token{Text, "paragraph", 0},
				Token{NewLine, "", 0},
				Token{Text, "Resumption", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: Beginning paragraph} " +
				"{Tag p () [{Text: Nested paragraph} ]} " +
				"{Text: Resumption} ]} ]",
		},
	} {
		runParserTest(t, data.tokens, data.flattenedTree)
	}
}

func TestMultipleBlankLines(t *testing.T) {
	for _, data := range [...]struct {
		tokens        []Token
		flattenedTree string
	}{
		{
			[]Token{
				Token{Text, "Introductory", 0},
				Token{Text, "sentence", 0},
				Token{Text, "for", 0},
				Token{Text, "a", 0},
				Token{Text, "list:", 0},
				Token{NewLine, "", 0},
				Token{NewLine, "", 4},
				Token{NewLine, "", 0},
				Token{NewLine, "", 0},
				Token{NewLine, "", 4},
				Token{UnorderedListItem, "-", 0},
				Token{Text, "abc", 0},
				Token{NewLine, "", 0},
				Token{NewLine, "", 0},
				Token{NewLine, "", 4},
				Token{UnorderedListItem, "-", 0},
				Token{Text, "def", 0},
				Token{NewLine, "", 15},
				Token{NewLine, "", 9},
				Token{NewLine, "", 4},
				Token{Text, "ghi", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: Introductory sentence for a list:} " +
				"{Tag ul () [{Tag li () [{Text: abc} ]} " +
				"{Tag li () [{Text: def ghi} ]} ]} ]} ]",
		},
		{
			[]Token{
				Token{Text, "Beginning", 0},
				Token{Text, "paragraph", 0},
				Token{NewLine, "", 0},
				Token{NewLine, "", 182},
				Token{NewLine, "", 81},
				Token{NewLine, "", 3},
				Token{NewLine, "", 4},
				Token{Text, "Nested", 0},
				Token{Text, "paragraph", 0},
				Token{NewLine, "", 0},
				Token{Text, "Resumption", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: Beginning paragraph} " +
				"{Tag p () [{Text: Nested paragraph} ]} " +
				"{Text: Resumption} ]} ]",
		},
	} {
		runParserTest(t, data.tokens, data.flattenedTree)
	}
}

func TestMultipleParagraphs(t *testing.T) {
	for _, data := range [...]struct {
		tokens        []Token
		flattenedTree string
	}{
		{
			[]Token{
				Token{Text, "Introductory", 0},
				Token{Text, "sentence", 0},
				Token{Text, "for", 0},
				Token{Text, "a", 0},
				Token{Text, "list:", 0},
				Token{NewLine, "", 0},
				Token{NewLine, "", 4},
				Token{NewLine, "", 0},
				Token{NewLine, "", 0},
				Token{NewLine, "", 4},
				Token{UnorderedListItem, "-", 0},
				Token{Text, "abc", 0},
				Token{NewLine, "", 0},
				Token{NewLine, "", 0},
				Token{NewLine, "", 4},
				Token{UnorderedListItem, "-", 0},
				Token{Text, "def", 0},
				Token{NewLine, "", 15},
				Token{NewLine, "", 9},
				Token{NewLine, "", 4},
				Token{Text, "ghi", 0},
				Token{NewLine, "", 1500128},
				Token{NewLine, "", 0},
				Token{Text, "New", 0},
				Token{Text, "paragraph", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: Introductory sentence for a list:} " +
				"{Tag ul () [{Tag li () [{Text: abc} ]} " +
				"{Tag li () [{Text: def ghi} ]} ]} ]} ]" +
				"[{Tag p () [{Text: New paragraph} ]} ]",
		},
		{
			[]Token{
				Token{Text, "Beginning", 0},
				Token{Text, "paragraph", 0},
				Token{NewLine, "", 0},
				Token{NewLine, "", 182},
				Token{NewLine, "", 81},
				Token{NewLine, "", 3},
				Token{NewLine, "", 4},
				Token{Text, "Nested", 0},
				Token{Text, "paragraph", 0},
				Token{NewLine, "", 4},
				Token{NewLine, "", 20},
				Token{NewLine, "", 0},
				Token{NewLine, "", 0},
				Token{Text, "different", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Text: Beginning paragraph} " +
				"{Tag p () [{Text: Nested paragraph} ]} ]} ]" +
				"[{Tag p () [{Text: different} ]} ]",
		},
	} {
		runParserTest(t, data.tokens, data.flattenedTree)
	}
}

func TestLiteralText(t *testing.T) {
	for _, data := range [...]struct {
		tokens        []Token
		flattenedTree string
	}{
		{
			[]Token{
				Token{LiteralText, "Literal Text", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Tag tt () [{Text: Literal Text} ]} ]} ]",
		},
		{
			[]Token{
				Token{LiteralText, "Literal\nText on\nmultiple lines", 0},
				Token{EndOfFile, "", 0},
			},
			"[{Tag p () [{Tag pre () [{Text: Literal\nText on\nmultiple lines} ]} ]} ]",
		},
	} {
		runParserTest(t, data.tokens, data.flattenedTree)
	}
}

func TestParagraphAfterList(t *testing.T) {
	// Beginning of paragraph 1.
	//     - Data 1
	//     - Data 2
	// More stuff
	//
	// New Paragraph
	runParserTest(t, []Token{
		Token{Text, "Beginning", 0},
		Token{Text, "of", 0},
		Token{Text, "paragraph", 0},
		Token{Text, "1.", 0},
		Token{NewLine, "", 4},
		Token{UnorderedListItem, "-", 0},
		Token{Text, "Data", 0},
		Token{Text, "1", 0},
		Token{NewLine, "", 4},
		Token{UnorderedListItem, "-", 0},
		Token{Text, "Data", 0},
		Token{Text, "2", 0},
		Token{NewLine, "", 0},
		Token{Text, "More", 0},
		Token{Text, "stuff", 0},
		Token{NewLine, "", 0},
		Token{NewLine, "", 0},
		Token{Text, "New", 0},
		Token{Text, "Paragraph", 0},
		Token{EndOfFile, "", 0}},
		"[{Tag p () [{Text: Beginning of paragraph 1.} "+
			"{Tag ul () [{Tag li () [{Text: Data 1} ]} "+
			"{Tag li () [{Text: Data 2} ]} ]} "+
			"{Text: More stuff} ]} ]"+
			"[{Tag p () [{Text: New Paragraph} ]} ]")
}

func TestTotallyEmphasized(t *testing.T) {
	// *Bold only*
	runParserTest(t, []Token{
		Token{BoldDelimeter, "*", 0},
		Token{Text, "Bold", 0},
		Token{Text, "Only", 0},
		Token{BoldDelimeter, "*", 0},
		Token{EndOfFile, "", 0}},
		"[{Tag p () [{Tag b () [{Text: Bold Only} ]} ]} ]")

	// /Emphasis only/
	runParserTest(t, []Token{
		Token{EmphasisDelimeter, "/", 0},
		Token{Text, "Emphasis", 0},
		Token{Text, "Only", 0},
		Token{EmphasisDelimeter, "/", 0},
		Token{EndOfFile, "", 0}},
		"[{Tag p () [{Tag em () [{Text: Emphasis Only} ]} ]} ]")
}

func TestSubList(t *testing.T) {
	// Stuff.
	//     - Heading 1
	//         - Sub-list item 1
	//         - Item 2
	//     - Heading 2
	// End
	runParserTest(t, []Token{
		Token{Text, "Stuff.", 0},
		Token{NewLine, "", 4},
		Token{UnorderedListItem, "-", 0},
		Token{Text, "Heading", 0},
		Token{Text, "1", 0},
		Token{NewLine, "", 8},
		Token{UnorderedListItem, "-", 0},
		Token{Text, "Sub-list", 0},
		Token{Text, "item", 0},
		Token{Text, "1", 0},
		Token{NewLine, "", 8},
		Token{UnorderedListItem, "-", 0},
		Token{Text, "Item", 0},
		Token{Text, "2", 0},
		Token{NewLine, "", 4},
		Token{UnorderedListItem, "-", 0},
		Token{Text, "Heading", 0},
		Token{Text, "2", 0},
		Token{NewLine, "", 0},
		Token{Text, "End", 0},
		Token{EndOfFile, "", 0}},
		"[{Tag p () [{Text: Stuff.} "+
			"{Tag ul () [{Tag li () [{Text: Heading 1} "+
			"{Tag ul () [{Tag li () [{Text: Sub-list item 1} ]} "+
			"{Tag li () [{Text: Item 2} ]} ]} ]} "+
			"{Tag li () [{Text: Heading 2} ]} ]} "+
			"{Text: End} ]} ]")
}

func TestMultiplyNestedListFalloff(t *testing.T) {
	// Stuff.
	//     - Heading 1
	//         - Sub-list item 1
	//             - Sub-sub-list
	//                 - Sub-sub-sub-list
	//     - Heading 2
	// End
	runParserTest(t, []Token{
		Token{Text, "Stuff.", 0},
		Token{NewLine, "", 4},
		Token{UnorderedListItem, "-", 0},
		Token{Text, "Heading", 0},
		Token{Text, "1", 0},
		Token{NewLine, "", 8},
		Token{UnorderedListItem, "-", 0},
		Token{Text, "Sub-list", 0},
		Token{Text, "item", 0},
		Token{Text, "1", 0},
		Token{NewLine, "", 12},
		Token{UnorderedListItem, "-", 0},
		Token{Text, "Sub-sub-list", 0},
		Token{NewLine, "", 16},
		Token{UnorderedListItem, "-", 0},
		Token{Text, "Sub-sub-sub-list", 0},
		Token{NewLine, "", 4},
		Token{UnorderedListItem, "-", 0},
		Token{Text, "Heading", 0},
		Token{Text, "2", 0},
		Token{NewLine, "", 0},
		Token{Text, "End", 0},
		Token{EndOfFile, "", 0}},
		"[{Tag p () [{Text: Stuff.} "+
			"{Tag ul () [{Tag li () [{Text: Heading 1} "+
			"{Tag ul () [{Tag li () [{Text: Sub-list item 1} "+
			"{Tag ul () [{Tag li () [{Text: Sub-sub-list} "+
			"{Tag ul () [{Tag li () [{Text: Sub-sub-sub-list} ]} ]} ]} ]} ]} ]} ]} "+
			"{Tag li () [{Text: Heading 2} ]} ]} "+
			"{Text: End} ]} ]")
}
