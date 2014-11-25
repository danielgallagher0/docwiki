// Copyright (c) 2014, Daniel Gallagher
// Use of this source code is covered by the MIT License, the full
// text of which can be found in the LICENSE file.

package main

import (
	"github.com/danielgallagher0/docwiki/wikilang"
	"testing"
)

func compare(t *testing.T, actual, expected string) {
	if expected != actual {
		t.Errorf("  Expected: \"%s\" (%d)", expected, len(expected))
		t.Errorf("    Actual: \"%s\" (%d)", actual, len(actual))
	}
}

type ViewTextTestData struct {
	data, expected string
}

func compareViewText(t *testing.T, allData []ViewTextTestData) {
	for _, data := range allData {
		compare(t, wikilang.WikiToHtml(data.data), data.expected)
	}
}

func TestWikiLinks(t *testing.T) {
	compareViewText(t, []ViewTextTestData{
		{data: "[LinkOnly]", expected: "<p>\n  <a href=\"/view/LinkOnly\">Link Only</a>\n</p>\n"},
		{data: "An embedded [Link] in some text", expected: "<p>\n  An embedded <a href=\"/view/Link\">Link</a> in some text\n</p>\n"},
		{data: "Multiple [WikiLinks] embedded in [LotsOfText].", expected: "<p>\n  Multiple <a href=\"/view/WikiLinks\">Wiki Links</a> embedded in <a href=\"/view/LotsOfText\"\n  >Lots Of Text</a>.\n</p>\n"},
		{data: "Spread over [Multiple\nLines]", expected: "<p>\n  Spread over <a href=\"/view/Multiple%0ALines\">Multiple Lines</a>\n</p>\n"},
		{data: "Not a[SeparateWord]", expected: "<p>\n  Not a <a href=\"/view/SeparateWord\">Separate Word</a>\n</p>\n"},
		{data: "Not [SeparateWord]s", expected: "<p>\n  Not <a href=\"/view/SeparateWord\">Separate Word</a> s\n</p>\n"},
		{data: "[Multiple Words]", expected: "<p>\n  <a href=\"/view/Multiple+Words\">Multiple Words</a>\n</p>\n"}})
}

func TestExternalLinks(t *testing.T) {
	compareViewText(t, []ViewTextTestData{
		{data: "[LinkOnly:http://www.google.com]", expected: "<p>\n  <a href=\"http://www.google.com\">Link Only</a>\n</p>\n"},
		{data: "An embedded [Link:mailto:test@example.com] in some text", expected: "<p>\n  An embedded <a href=\"mailto:test@example.com\">Link</a> in some text\n</p>\n"},
		{data: "Multiple [WikiLinks:OtherLink] embedded in [LotsOfText].", expected: "<p>\n  Multiple <a href=\"OtherLink\">Wiki Links</a> embedded in <a href=\"/view/LotsOfText\"\n  >Lots Of Text</a>.\n</p>\n"},
		{data: "Spread over [Multiple\nLines:http://www.example.com]", expected: "<p>\n  Spread over <a href=\"http://www.example.com\">Multiple Lines</a>\n</p>\n"},
		{data: "Not a[SeparateWord:invalid stuff]", expected: "<p>\n  Not a <a href=\"invalid stuff\">Separate Word</a>\n</p>\n"},
		{data: "Not [SeparateWord:chrome://history]s", expected: "<p>\n  Not <a href=\"chrome://history\">Separate Word</a> s\n</p>\n"},
		{data: "[Multiple Words:http://asdf.com]", expected: "<p>\n  <a href=\"http://asdf.com\">Multiple Words</a>\n</p>\n"}})
}

func TestParagraphs(t *testing.T) {
	compareViewText(t, []ViewTextTestData{
		{data: "One line", expected: "<p>\n  One line\n</p>\n"},
		{data: "Multiple\nLines", expected: "<p>\n  Multiple Lines\n</p>\n"},
		{data: "Empty\n\nLine", expected: "<p>\n  Empty\n</p>\n\n\n<p>\n  Line\n</p>\n"},
		{data: "Line\n   \nWith spaces", expected: "<p>\n  Line\n</p>\n\n\n<p>\n  With spaces\n</p>\n"}})
}

func TestParagraphStartedByWikiWord(t *testing.T) {
	compareViewText(t, []ViewTextTestData{
		{data: "Paragraph.\n\n[WikiWord]", expected: "<p>\n  Paragraph.\n</p>\n\n\n<p>\n  <a href=\"/view/WikiWord\">Wiki Word</a>\n</p>\n"}})
}

func TestUnorderedList(t *testing.T) {
	compareViewText(t, []ViewTextTestData{
		{data: "Beginning of paragraph 1.\n" +
			"    - Data 1\n" +
			"    - Data 2\n" +
			"More stuff\n\n" +
			"New Paragraph",
			expected: "<p>\n" +
				"  Beginning of paragraph 1.\n" +
				"  <ul>\n" +
				"    \n" +
				"    <li>\n" +
				"      Data 1\n" +
				"    </li>\n" +
				"    \n" +
				"    <li>\n" +
				"      Data 2\n" +
				"    </li>\n" +
				"    \n" +
				"  </ul>\n" +
				"  More stuff\n" +
				"</p>\n" +
				"\n" +
				"\n" +
				"<p>\n" +
				"  New Paragraph\n" +
				"</p>\n"},
		{data: "Beginning of paragraph 2.\n" +
			"    - Data 1\n" +
			"    - Data 2\n\n" +
			"    More stuff\n\n" +
			"New Paragraph",
			expected: "<p>\n" +
				"  Beginning of paragraph 2.\n" +
				"  <ul>\n" +
				"    \n" +
				"    <li>\n" +
				"      Data 1\n" +
				"    </li>\n" +
				"    \n" +
				"    <li>\n" +
				"      Data 2 More stuff\n" +
				"    </li>\n" +
				"    \n" +
				"  </ul>\n" +
				"  \n" +
				"</p>\n\n\n" +
				"<p>\n" +
				"  New Paragraph\n" +
				"</p>\n"},
		{data: "Stuff.\n" +
			"    - Heading 1\n" +
			"        - Sub-list item 1\n" +
			"        - Item 2\n" +
			"    - Heading 2\n" +
			"End",
			expected: "<p>\n" +
				"  Stuff.\n" +
				"  <ul>\n" +
				"    \n" +
				"    <li>\n" +
				"      Heading 1\n" +
				"      <ul>\n" +
				"        \n" +
				"        <li>\n" +
				"          Sub-list item 1\n" +
				"        </li>\n" +
				"        \n" +
				"        <li>\n" +
				"          Item 2\n" +
				"        </li>\n" +
				"        \n" +
				"      </ul>\n" +
				"      \n" +
				"    </li>\n" +
				"    \n" +
				"    <li>\n" +
				"      Heading 2\n" +
				"    </li>\n" +
				"    \n" +
				"  </ul>\n" +
				"  End\n" +
				"</p>\n"}})
}

func TestOrderedList(t *testing.T) {
	compareViewText(t, []ViewTextTestData{
		{data: "Beginning of paragraph 1.\n" +
			"    # Data 1\n" +
			"    # Data 2\n" +
			"More stuff\n\n" +
			"New Paragraph",
			expected: "<p>\n" +
				"  Beginning of paragraph 1.\n" +
				"  <ol>\n" +
				"    \n" +
				"    <li>\n" +
				"      Data 1\n" +
				"    </li>\n" +
				"    \n" +
				"    <li>\n" +
				"      Data 2\n" +
				"    </li>\n" +
				"    \n" +
				"  </ol>\n" +
				"  More stuff\n" +
				"</p>\n" +
				"\n" +
				"\n" +
				"<p>\n" +
				"  New Paragraph\n" +
				"</p>\n"},
		{data: "Beginning of paragraph 2.\n" +
			"    # Data 1\n" +
			"    # Data 2\n\n" +
			"    More stuff\n\n" +
			"New Paragraph",
			expected: "<p>\n" +
				"  Beginning of paragraph 2.\n" +
				"  <ol>\n" +
				"    \n" +
				"    <li>\n" +
				"      Data 1\n" +
				"    </li>\n" +
				"    \n" +
				"    <li>\n" +
				"      Data 2 More stuff\n" +
				"    </li>\n" +
				"    \n" +
				"  </ol>\n" +
				"  \n" +
				"</p>\n\n\n" +
				"<p>\n" +
				"  New Paragraph\n" +
				"</p>\n"},
		{data: "Stuff.\n" +
			"    # Heading 1\n" +
			"        # Sub-list item 1\n" +
			"        # Item 2\n" +
			"    # Heading 2\n" +
			"End",
			expected: "<p>\n" +
				"  Stuff.\n" +
				"  <ol>\n" +
				"    \n" +
				"    <li>\n" +
				"      Heading 1\n" +
				"      <ol>\n" +
				"        \n" +
				"        <li>\n" +
				"          Sub-list item 1\n" +
				"        </li>\n" +
				"        \n" +
				"        <li>\n" +
				"          Item 2\n" +
				"        </li>\n" +
				"        \n" +
				"      </ol>\n" +
				"      \n" +
				"    </li>\n" +
				"    \n" +
				"    <li>\n" +
				"      Heading 2\n" +
				"    </li>\n" +
				"    \n" +
				"  </ol>\n" +
				"  End\n" +
				"</p>\n"}})
}

func TestEmphasis(t *testing.T) {
	compareViewText(t, []ViewTextTestData{
		{data: "*Bold Only*", expected: "<p>\n  <b>Bold Only</b>\n</p>\n"},
		{data: "An embedded *Bold* in some text", expected: "<p>\n  An embedded <b>Bold</b> in some text\n</p>\n"},
		{data: "Multiple *Bold elements* embedded in *Lots Of Text*.", expected: "<p>\n  Multiple <b>Bold elements</b> embedded in <b>Lots Of Text</b>.\n</p>\n"},
		{data: "Spread over *Multiple\nLines*", expected: "<p>\n  Spread over <b>Multiple Lines</b>\n</p>\n"},
		{data: "Still a*Separate Word*", expected: "<p>\n  Still a <b>Separate Word</b>\n</p>\n"},
		{data: "Still *Separate Word*s", expected: "<p>\n  Still <b>Separate Word</b> s\n</p>\n"},
		{data: "/Bold Only/", expected: "<p>\n  <em>Bold Only</em>\n</p>\n"},
		{data: "An embedded /Bold/ in some text", expected: "<p>\n  An embedded <em>Bold</em> in some text\n</p>\n"},
		{data: "Multiple /Bold elements/ embedded in /Lots Of Text/.", expected: "<p>\n  Multiple <em>Bold elements</em> embedded in <em>Lots Of Text</em>.\n</p>\n"},
		{data: "Spread over /Multiple\nLines/", expected: "<p>\n  Spread over <em>Multiple Lines</em>\n</p>\n"},
		{data: "Still a/Separate Word/", expected: "<p>\n  Still a <em>Separate Word</em>\n</p>\n"},
		{data: "Still /Separate Word/s", expected: "<p>\n  Still <em>Separate Word</em> s\n</p>\n"},
		{data: "Mixed *bold and /emphasis/* or /emphasis and *bold*/", expected: "<p>\n  Mixed <b>bold and <em>emphasis</em></b> or <em>emphasis and <b>bold</b></em>\n</p>\n"},
		{data: "Improper *mixing /of emphasis* and bold/", expected: "<p>\n  Improper <b>mixing <em>of emphasis</em> and bold</b>\n</p>\n"},
	})
}
