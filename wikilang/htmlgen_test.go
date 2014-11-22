package wikilang

import (
	"testing"
)

func runHtmlGenTest(t *testing.T, trees []ParseTree, expected string) {
	in := make(chan ParseTree)
	out := make(chan string)

	gen := HtmlGen{in, out}

	go func() {
		for _, tree := range trees {
			in <- tree
		}
	}()
	go gen.Generate()

	actual := <-out
	if actual != expected {
		t.Errorf("Expected: \"%s\"", expected)
		t.Errorf("  Actual: \"%s\"", actual)
	}
}

func TestNoParseTrees(t *testing.T) {
	runHtmlGenTest(t, []ParseTree{ParseTree{}}, "")
}

func TestBasicParagraphs(t *testing.T) {
	for _, data := range [...]struct {
		trees []ParseTree
		text  string
	}{
		{
			[]ParseTree{
				ParseTree{
					[]ParseNode{
						TagNode{
							Paragraph,
							map[string]string{},
							ParseTree{
								[]ParseNode{
									TextNode{
										"This is a very simple paragraph.",
									},
								},
							},
						},
					},
				},
				ParseTree{},
			},
			"<p>\n" +
				"  This is a very simple paragraph.\n" +
				"</p>\n",
		},
		{
			[]ParseTree{
				ParseTree{
					[]ParseNode{
						TagNode{
							Paragraph,
							map[string]string{},
							ParseTree{
								[]ParseNode{
									TextNode{
										"This is a paragraph that contains much " +
											"longer text than the previous one. It " +
											"is broken up into sentences and so the " +
											"generator should wrap it to multiple " +
											"lines so we can read it easier.",
									},
								},
							},
						},
					},
				},
				ParseTree{},
			},
			"<p>\n" +
				"  This is a paragraph that contains much longer text than the previous one. It is broken\n" +
				"  up into sentences and so the generator should wrap it to multiple lines so we can\n" +
				"  read it easier.\n" +
				"</p>\n",
		},
	} {
		runHtmlGenTest(t, data.trees, data.text)
	}
}
