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

type HtmlGen struct {
	in  chan ParseTree
	out chan string
}

func NewHtmlGen(i chan ParseTree, o chan string) HtmlGen {
	return HtmlGen{i, o}
}

type HtmlGenVisitor struct {
	writer CodeWriter

	indent      int
	lastWasText bool
}

func (g HtmlGen) Generate() {
	g.out <- g.generateString()
}

func (g HtmlGen) generateString() string {
	var buf bytes.Buffer

	tree := <-g.in
	if len(tree.nodes) > 0 {
		buf.WriteString(generateTreeString(tree))

		for tree = <-g.in; len(tree.nodes) > 0; tree = <-g.in {
			buf.WriteString("\n\n")
			buf.WriteString(generateTreeString(tree))
		}
	}

	return buf.String()
}

func generateTreeString(tree ParseTree) string {
	var buf bytes.Buffer

	visitor := HtmlGenVisitor{NewCodeWriter(&buf), 0, false}
	tree.Visit(&visitor)

	return buf.String()
}

func (v *HtmlGenVisitor) VisitTagBegin(n TagNode) {
	isParagraph := strings.Contains(paragraphTags, ","+n.tag+",")
	if isParagraph {
		v.writer.FreshLine()
	} else if v.lastWasText {
		fmt.Fprintf(v.writer, " ")
	}

	fmt.Fprintf(v.writer, "<%s", n.tag)
	for key, value := range n.attributes {
		fmt.Fprintf(v.writer, " %s=\"%s\"", key, value)
	}
	fmt.Fprintf(v.writer, ">")
	if isParagraph {
		v.writer.ChangeIndentation(2)
		v.writer.FreshLine()
		if n.tag == Preformatted {
			v.writer.LiteralText(true)
		}
	}
	v.lastWasText = false
}

func (v *HtmlGenVisitor) VisitTagEnd(n TagNode) {
	isParagraph := strings.Contains(paragraphTags, ","+n.tag+",")
	if isParagraph {
		if n.tag == Preformatted {
			v.writer.LiteralText(false)
		}
		v.writer.ChangeIndentation(-2)
		v.writer.FreshLine()
	}
	fmt.Fprintf(v.writer, "</%s>", n.tag)
	if isParagraph {
		v.writer.NewLine()
		v.lastWasText = false
	}
}

func (v *HtmlGenVisitor) VisitText(n TextNode) {
	if len(n.text) > 0 {
		if v.lastWasText && !unicode.IsPunct(rune(n.text[0])) {
			fmt.Fprintf(v.writer, " ")
		}

		fmt.Fprintf(v.writer, "%s", n.text)
		v.lastWasText = !unicode.IsPunct(rune(n.text[len(n.text)-1]))
	}
}
