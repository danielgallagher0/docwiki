// Copyright (c) 2014, Daniel Gallagher
// Use of this source code is covered by the MIT License, the full
// text of which can be found in the LICENSE file.

package wikilang

import (
	"fmt"
	"net/url"
	"strings"
)

// These values indicate the type of tag node in a parse tree.  They
// happen to correspond directly to the HTML tag used by the HTML
// generator, but that is only a convenience.
const (
	Paragraph     = "p"   // Container tag for a paragraph
	OrderedList   = "ol"  // Container tag for an ordered list
	UnorderedList = "ul"  // Container tag for an unordered list
	ListItem      = "li"  // An item in a list
	Link          = "a"   // A link
	Bold          = "b"   // Bold text
	Emphasis      = "em"  // Emphasized text
	Literal       = "tt"  // Literal text embedded in a single line
	Preformatted  = "pre" // Multi-line literal text
)

// A Visitor allows the parse tree to call a function for each node.
// Parse tree implements the visitor pattern, and the visitor must
// implement this interface.  The specific type of node is already
// called out, so there is no need for double dispatch.  Tag nodes are
// visited when they are entered and exited, and text nodes are
// visited once.
type Visitor interface {
	// VisitTagBegin is called on a tag node when the visitor is about
	// to visit the tag node's internal parse tree.
	VisitTagBegin(n TagNode)

	// VisitTagEnd is called on a tag node when the visitor has
	// finished visiting the tag node's internal parse tree.
	VisitTagEnd(n TagNode)

	// VisitText is called when a text node is encountered in a parse
	// tree.
	VisitText(n TextNode)
}

// A ParseNode provides the interface that other nodes must implement
// in order to be printed and visited.
type ParseNode interface {
	fmt.Stringer

	// Visit causes the node to be visited by the provided Visitor.
	Visit(v Visitor)
}

// A ParseTree provides a list of parse nodes.  Tag nodes may contain
// more parse trees (thus the name is tree, not list).
type ParseTree struct {
	Nodes []ParseNode
}

// A TagNode represents a parse tree that should be surrounded by a
// tag.  This corresponds directly to HTML tags, but could be used to
// generate other types of output
type TagNode struct {
	Tag        string            // Tag surronding the inner tree
	Attributes map[string]string // Attributes associated with this tag
	Tree       ParseTree         // Contents of the tag
}

// A TextNode is the atomic value inside a parse tree.  It represents
// a single block of text.
type TextNode struct {
	Text string // Value of the text block
}

// A Parser contains the channels that the parsing algorithm uses to
// communicate with the lexer and the generator.  It reads tokens in,
// and emits full parse trees, one per top-level paragraph.
type Parser struct {
	In  chan Token     // Channel to read tokens from
	Out chan ParseTree // Channel to write fully-formed parse trees to

	nextToken *Token
}

// NewParser creates a parser that uses the given channels to
// communicate to the lexer and generator.
func NewParser(i chan Token, o chan ParseTree) Parser {
	return Parser{i, o, nil}
}

// String converts a ParseTree to its string representation.  The
// entire parse tree is surrounded by square brackets, and the nodes
// all have spaces between them.
func (t *ParseTree) String() string {
	s := "["
	for _, node := range t.Nodes {
		s = s + node.String() + " "
	}

	s = s + "]"

	return s
}

// String converts a TagNode to its string representation.  The entire
// tag is surrounded by curly brackets, followed by the name of the
// tag, its attributes in parentheses, then its inner ParseTree.
func (n TagNode) String() string {
	s := "{Tag " + n.Tag + " ("
	for key, value := range n.Attributes {
		s = s + key + "=" + value + " "
	}
	s = s + ") " + n.Tree.String() + "}"

	return s
}

// String converts a TextNode to its string representation.  The data
// is surrounded by curly brackets and denotes that it is a text
// node.
func (n TextNode) String() string {
	return "{Text: " + n.Text + "}"
}

// Parse reads tokens and outputs full ParseTrees.  The tokens are
// read by paragraph, and one ParseTree is emitted per paragraph.  A
// "paragraph" in these terms is a top-level paragraph (all at the
// same indentation).  Paragraphs are separated by at least one blank
// line.
func (p *Parser) Parse() {
	for {
		tokens, end := p.readParagraph()
		for _, par := range parseParagraph(combineTokens(indentTokens(tokens))) {
			p.Out <- par
		}

		if end {
			p.Out <- ParseTree{}
			break
		}
	}
}

func (p *Parser) readParagraph() (tokens []Token, end bool) {
	hasContent := false

	end = false

	for {
		var token Token
		if p.nextToken != nil {
			token = *p.nextToken
			p.nextToken = nil
		} else {
			token = <-p.In
		}

		switch token.Type {
		case Text, BoldDelimeter, EmphasisDelimeter, UnorderedListItem, OrderedListItem, LiteralText, WikiLink, Tag:
			tokens = append(tokens, token)
			hasContent = true

		case NewLine:
			consecutiveNewLines := 0
			shouldBreak := false

			for token.Type == NewLine {
				tokens = append(tokens, token)
				consecutiveNewLines++

				shouldBreak = (consecutiveNewLines > 1 && token.IntValue == 0)

				token = <-p.In
			}
			p.nextToken = &token

			if hasContent && shouldBreak {
				hasContent = false
				return
			}

		case EndOfFile:
			end = true
			return
		}
	}
}

func combineTokens(tokens []Token) []Token {
	combined := []Token{}

	currentToken := Token{EndOfFile, "", 0}
	currentlyText := false

	for _, token := range tokens {
		if token.IntValue != currentToken.IntValue {
			if currentToken.Type != EndOfFile {
				combined = append(combined, currentToken)
			}
			currentToken = Token{EndOfFile, "", 0}
			currentlyText = false
		} else if currentlyText && token.Type != Text && token.Type != Tag {
			if currentToken.Type != EndOfFile {
				combined = append(combined, currentToken)
			}
			currentToken = Token{EndOfFile, "", 0}
		}

		switch token.Type {
		case Text, Tag:
			if currentlyText {
				currentToken.TextValue = currentToken.TextValue + " " + token.TextValue
			} else {
				currentToken.Type = Text
				currentToken.TextValue = token.TextValue
			}
			currentToken.IntValue = token.IntValue
			currentlyText = true

		default:
			combined = append(combined, token)
			currentToken = Token{EndOfFile, "", 0}
			currentlyText = false
		}
	}

	if currentlyText {
		combined = append(combined, currentToken)
	}

	return combined
}

func indentTokens(tokens []Token) []Token {
	indented := []Token{}
	currentIndentation := 0

	for _, token := range tokens {
		if token.Type == NewLine {
			currentIndentation = token.IntValue
		} else {
			token.IntValue = currentIndentation
			indented = append(indented, token)
		}
	}

	return indented
}

func wikiWordUrl(s string) string {
	parts := strings.SplitN(s, ":", 3)
	switch len(parts) {
	case 1:
		return proxyRoot() + "/view/" + url.QueryEscape(parts[0])

	case 2:
		return parts[1]

	case 3:
		if parts[0] == "doc" {
			return DocLink(parts[1], parts[2])
		}

		return parts[1] + ":" + parts[2]
	}

	return ""
}

func wikiWordText(s string) string {
	parts := strings.SplitN(s, ":", 3)
	switch len(parts) {
	case 1, 2:
		return WikiCase(parts[0])

	case 3:
		if parts[0] == "doc" {
			return parts[2]
		}

		return WikiCase(parts[0])
	}

	return ""
}

// ToNode converts a token to its corresponding parse node.  For
// tokens that denote TagNodes, the inner parse trees are empty.
// Embedded HTML is treated as text, and is passed straight to the
// output.
func (t Token) ToNode() ParseNode {
	switch t.Type {
	case Text, Tag:
		return TextNode{
			t.TextValue,
		}

	case WikiLink:
		return TagNode{
			Link,
			map[string]string{
				"href": wikiWordUrl(t.TextValue),
			},
			ParseTree{
				[]ParseNode{
					TextNode{
						wikiWordText(t.TextValue),
					}}}}

	case BoldDelimeter:
		return TagNode{Bold, map[string]string{}, ParseTree{}}

	case EmphasisDelimeter:
		return TagNode{Emphasis, map[string]string{}, ParseTree{}}

	case OrderedListItem, UnorderedListItem:
		return TagNode{ListItem, map[string]string{}, ParseTree{}}

	case LiteralText:
		tagType := Literal
		if strings.Contains(t.TextValue, "\n") {
			tagType = Preformatted
		}

		return TagNode{tagType, map[string]string{},
			ParseTree{[]ParseNode{TextNode{t.TextValue}}}}
	}

	return TextNode{""}
}

func buildTag(tokens []Token, indentation, prevIndentation int, endPredicates []func(Token) (bool, bool)) (ParseTree, int) {
	tagType := ""

	innerTree := ParseTree{}

	i := 0
ParagraphParser:
	for ; i < len(tokens); i++ {
		if tokens[i].IntValue < indentation {
			if tokens[i].IntValue >= prevIndentation {
				i--
			}
			break
		}

		for _, f := range endPredicates {
			if end, consumed := f(tokens[i]); end {
				if consumed {
					i++
				}
				break ParagraphParser
			}
		}

		if tokens[i].IntValue > indentation {
			subtree, next := buildTag(tokens[i:], tokens[i].IntValue, indentation, endPredicates)
			innerTree.Nodes = append(innerTree.Nodes, subtree.Nodes...)
			i += next
			continue
		}

		if tagType == "" {
			tagType = getWrapperTag(tokens[i])
		}

		switch tokens[i].Type {
		case BoldDelimeter, EmphasisDelimeter:
			nextPredicates := endPredicates
			nextPredicates = append(nextPredicates, func(token Token) (bool, bool) {
				return token.Type == tokens[i].Type, true
			})
			subtree, next := buildTag(tokens[i+1:], tokens[i].IntValue, indentation, nextPredicates)
			if len(subtree.Nodes) > 0 {
				subtag := subtree.Nodes[0].(TagNode)
				boldType := Bold
				if tokens[i].Type == EmphasisDelimeter {
					boldType = Emphasis
				}
				boldTree := TagNode{boldType, map[string]string{}, subtag.Tree}
				innerTree.Nodes = append(innerTree.Nodes, boldTree)
			}
			i += next

		case UnorderedListItem, OrderedListItem:
			nextPredicates := endPredicates
			nextPredicates = append(nextPredicates, func(token Token) (bool, bool) {
				return (token.Type == UnorderedListItem ||
					token.Type == OrderedListItem) &&
					token.IntValue <= indentation, false
			})

			endType := tokens[i].Type
			currentIndent := tokens[i].IntValue
			for i < len(tokens) && ((tokens[i].Type == endType && tokens[i].IntValue == currentIndent) || tokens[i].IntValue > currentIndent) {
				if tokens[i].IntValue > indentation {
					subtree, next := buildTag(tokens[i:], tokens[i].IntValue, indentation, endPredicates)
					innerTree.Nodes = append(innerTree.Nodes, subtree.Nodes...)
					i += next + 1
				} else {
					subtree, next := buildTag(tokens[i+1:], tokens[i].IntValue, indentation, nextPredicates)
					if len(subtree.Nodes) > 0 {
						firstSubNode := subtree.Nodes[0].(TagNode)
						if firstSubNode.Tag == Paragraph {
							subtree = firstSubNode.Tree
						}

						listItem := TagNode{ListItem, map[string]string{}, subtree}
						innerTree.Nodes = append(innerTree.Nodes, listItem)
					}

					i += next + 1
				}
			}

			if (i < len(tokens) && tokens[i].Type == endType) ||
				(i < len(tokens) && tokens[i-1].IntValue > tokens[i].IntValue) ||
				(i+1 < len(tokens) && tokens[i].IntValue <= tokens[i+1].IntValue) {
				i--
			}

			break ParagraphParser

		default:
			innerTree.Nodes = append(innerTree.Nodes, tokens[i].ToNode())
		}
	}

	if tagType == "" {
		return ParseTree{}, i
	}

	return ParseTree{[]ParseNode{TagNode{tagType, map[string]string{}, innerTree}}}, i
}

func parseParagraph(tokens []Token) []ParseTree {
	trees := []ParseTree{}
	start := 0

	for start < len(tokens) {
		tree, next := buildTag(tokens[start:], 0, 0, []func(Token) (bool, bool){})
		trees = append(trees, tree)
		start += next + 1
	}

	return trees
}

func getWrapperTag(token Token) string {
	wrapperTag := Paragraph
	switch token.Type {
	case UnorderedListItem:
		wrapperTag = UnorderedList

	case OrderedListItem:
		wrapperTag = OrderedList
	}

	return wrapperTag
}

// Visit visits each of the nodes in the ParseTree.
func (t ParseTree) Visit(v Visitor) {
	for _, node := range t.Nodes {
		node.Visit(v)
	}
}

// Visit visits the tag node and all of the nodes in its ParseTree.
// The tag node is visited before and after its ParseTree.
func (n TagNode) Visit(v Visitor) {
	v.VisitTagBegin(n)
	n.Tree.Visit(v)
	v.VisitTagEnd(n)
}

// Visit visits the text node exactly once.
func (n TextNode) Visit(v Visitor) {
	v.VisitText(n)
}
