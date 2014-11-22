package wikilang

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	Paragraph     = "p"
	OrderedList   = "ol"
	UnorderedList = "ul"
	ListItem      = "li"
	Link          = "a"
	Bold          = "b"
	Emphasis      = "em"
	Literal       = "tt"
	Preformatted  = "pre"
)

type Visitor interface {
	VisitTagBegin(n TagNode)
	VisitTagEnd(n TagNode)
	VisitText(n TextNode)
}

type ParseNode interface {
	fmt.Stringer

	Visit(v Visitor)
}

type ParseTree struct {
	nodes []ParseNode
}

type TagNode struct {
	tag        string
	attributes map[string]string
	tree       ParseTree
}

type TextNode struct {
	text string
}

type NewLineNode struct {
	indentation int
}

type Parser struct {
	in  chan Token
	out chan ParseTree

	nextToken *Token
}

func NewParser(i chan Token, o chan ParseTree) Parser {
	return Parser{i, o, nil}
}

// Parse Tree

func (t *ParseTree) String() string {
	s := "["
	for _, node := range t.nodes {
		s = s + node.String() + " "
	}

	s = s + "]"

	return s
}

// Tag Node

func (n TagNode) String() string {
	s := "{Tag " + n.tag + " ("
	for key, value := range n.attributes {
		s = s + key + "=" + value + " "
	}
	s = s + ") " + n.tree.String() + "}"

	return s
}

// Text Node

func (n TextNode) String() string {
	return "{Text: " + n.text + "}"
}

// New line node

func (n NewLineNode) String() string {
	return "\n"
}

// Parser

func (p *Parser) Parse() {
	for {
		tokens, end := p.ReadParagraph()
		for _, par := range ParseParagraph(combineTokens(indentTokens(tokens))) {
			p.out <- par
		}

		if end {
			p.out <- ParseTree{}
			break
		}
	}
}

func (p *Parser) ReadParagraph() (tokens []Token, end bool) {
	hasContent := false

	end = false

	for {
		var token Token
		if p.nextToken != nil {
			token = *p.nextToken
			p.nextToken = nil
		} else {
			token = <-p.in
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

				token = <-p.in
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

func WikiCase(link string) string {
	addSpaces := regexp.MustCompile("[[:upper:]]")
	splitString := strings.Trim(addSpaces.ReplaceAllStringFunc(link, func(s string) string {
		return " " + s
	}), " ")

	consolidateSpaces := regexp.MustCompile("[[:space:]]+")
	return consolidateSpaces.ReplaceAllString(splitString, " ")
}

func wikiWordUrl(s string) string {
	parts := strings.SplitN(s, ":", 3)
	switch len(parts) {
	case 1:
		return "/view/" + url.QueryEscape(parts[0])

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

	case NewLine:
		return NewLineNode{t.IntValue}
	}

	return TextNode{""}
}

func BuildTag(tokens []Token, indentation, prevIndentation int, endPredicates []func(Token) (bool, bool)) (ParseTree, int) {
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
			subtree, next := BuildTag(tokens[i:], tokens[i].IntValue, indentation, endPredicates)
			innerTree.nodes = append(innerTree.nodes, subtree.nodes...)
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
			subtree, next := BuildTag(tokens[i+1:], tokens[i].IntValue, indentation, nextPredicates)
			if len(subtree.nodes) > 0 {
				subtag := subtree.nodes[0].(TagNode)
				boldType := Bold;
				if tokens[i].Type == EmphasisDelimeter {
					boldType = Emphasis
				}
				boldTree := TagNode{boldType, map[string]string{}, subtag.tree}
				innerTree.nodes = append(innerTree.nodes, boldTree)
			}
			i += next

		case UnorderedListItem, OrderedListItem:
			nextPredicates := endPredicates
			nextPredicates = append(nextPredicates, func(token Token) (bool, bool) {
				return ( token.Type == UnorderedListItem ||
					token.Type == OrderedListItem ) &&
					token.IntValue <= indentation, false
			})

			endType := tokens[i].Type
			currentIndent := tokens[i].IntValue
			for i < len(tokens) && ((tokens[i].Type == endType && tokens[i].IntValue == currentIndent) || tokens[i].IntValue > currentIndent) {
				if tokens[i].IntValue > indentation {
					subtree, next := BuildTag(tokens[i:], tokens[i].IntValue, indentation, endPredicates)
					innerTree.nodes = append(innerTree.nodes, subtree.nodes...)
					i += next + 1
				} else {
					subtree, next := BuildTag(tokens[i+1:], tokens[i].IntValue, indentation, nextPredicates)
					if len(subtree.nodes) > 0 {
						firstSubNode := subtree.nodes[0].(TagNode)
						if firstSubNode.tag == Paragraph {
							subtree = firstSubNode.tree
						}

						listItem := TagNode{ListItem, map[string]string{}, subtree}
						innerTree.nodes = append(innerTree.nodes, listItem)
					}

					i += next + 1
				}
			}

			if (i < len(tokens) && tokens[i].Type == endType) ||
				(i < len(tokens) && tokens[i-1].IntValue > tokens[i].IntValue) ||
				(i + 1 < len(tokens) && tokens[i].IntValue <= tokens[i+1].IntValue){
				i--
			}

			break ParagraphParser

		default:
			innerTree.nodes = append(innerTree.nodes, tokens[i].ToNode())
		}
	}

	if tagType == "" {
		return ParseTree{}, i
	}

	return ParseTree{[]ParseNode{TagNode{tagType, map[string]string{}, innerTree}}}, i
}

func ParseParagraph(tokens []Token) []ParseTree {
	trees := []ParseTree{}
	start := 0

	for start < len(tokens) {
		tree, next := BuildTag(tokens[start:], 0, 0, []func(Token) (bool, bool){})
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

// Visiting tree

func (t ParseTree) Visit(v Visitor) {
	for _, node := range t.nodes {
		node.Visit(v)
	}
}

func (n TagNode) Visit(v Visitor) {
	v.VisitTagBegin(n)
	n.tree.Visit(v)
	v.VisitTagEnd(n)
}

func (n TextNode) Visit(v Visitor) {
	v.VisitText(n)
}

func (n NewLineNode) Visit(v Visitor) {
}
