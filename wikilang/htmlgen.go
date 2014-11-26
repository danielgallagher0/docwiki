// Copyright (c) 2014, Daniel Gallagher
// Use of this source code is covered by the MIT License, the full
// text of which can be found in the LICENSE file.

package wikilang

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

const paragraphTags = ",p,pre,ul,li,ol,"

// A HtmlGen converts wiki a set of ParseTrees to HTML output.
type HtmlGen struct {
	In  chan ParseTree // Channel to read ParseTrees from
	Out chan string    // Channel to write final output to
}

// NewHtmlGen creates an HtmlGen that communicates over the provided
// channels.
func NewHtmlGen(i chan ParseTree, o chan string) HtmlGen {
	return HtmlGen{i, o}
}

type htmlGenVisitor struct {
	writer CodeWriter

	indent      int
	lastWasText bool
}

// Generate reads all of the parse trees from the input channel and
// writes the total output to the output channel.  The output channel
// is only written to once.
func (g HtmlGen) Generate() {
	g.Out <- g.generateString()
}

func (g HtmlGen) generateString() string {
	var buf bytes.Buffer

	tree := <-g.In
	if len(tree.Nodes) > 0 {
		buf.WriteString(generateTreeString(tree))

		for tree = <-g.In; len(tree.Nodes) > 0; tree = <-g.In {
			buf.WriteString("\n\n")
			buf.WriteString(generateTreeString(tree))
		}
	}

	return buf.String()
}

func generateTreeString(tree ParseTree) string {
	var buf bytes.Buffer

	visitor := htmlGenVisitor{NewCodeWriter(&buf), 0, false}
	tree.Visit(&visitor)

	return buf.String()
}

func (v *htmlGenVisitor) VisitTagBegin(n TagNode) {
	isParagraph := strings.Contains(paragraphTags, ","+n.Tag+",")
	if isParagraph {
		v.writer.FreshLine()
	} else if v.lastWasText {
		fmt.Fprintf(v.writer, " ")
	}

	fmt.Fprintf(v.writer, "<%s", n.Tag)
	for key, value := range n.Attributes {
		fmt.Fprintf(v.writer, " %s=\"%s\"", key, value)
	}
	fmt.Fprintf(v.writer, ">")
	if isParagraph {
		v.writer.ChangeIndentation(2)
		v.writer.FreshLine()
		if n.Tag == Preformatted {
			v.writer.LiteralText(true)
		}
	}
	v.lastWasText = false
}

func (v *htmlGenVisitor) VisitTagEnd(n TagNode) {
	isParagraph := strings.Contains(paragraphTags, ","+n.Tag+",")
	if isParagraph {
		if n.Tag == Preformatted {
			v.writer.LiteralText(false)
		}
		v.writer.ChangeIndentation(-2)
		v.writer.FreshLine()
	}
	fmt.Fprintf(v.writer, "</%s>", n.Tag)
	if isParagraph {
		v.writer.NewLine()
		v.lastWasText = false
	}
}

func (v *htmlGenVisitor) VisitText(n TextNode) {
	if len(n.Text) > 0 {
		if v.lastWasText && !unicode.IsPunct(rune(n.Text[0])) {
			fmt.Fprintf(v.writer, " ")
		}

		fmt.Fprintf(v.writer, "%s", n.Text)
		v.lastWasText = !unicode.IsPunct(rune(n.Text[len(n.Text)-1]))
	}
}
