package wikilang

import "testing"

func TestWikiCase(t *testing.T) {
	type TestData struct {
		data, expected string
	}

	for _, data := range [...]TestData{
		{data: "lowercase", expected: "lowercase"},
		{data: "Uppercase", expected: "Uppercase"},
		{data: "CamelCase", expected: "Camel Case"},
		{data: "wrongStart", expected: "wrong Start"},
		{data: "LotsOfWords", expected: "Lots Of Words"},
		{data: "Existing Spacing", expected: "Existing Spacing"},
		{data: "Mixed SpacingWords", expected: "Mixed Spacing Words"}} {

			if actual := WikiCase(data.data); actual != data.expected {
				t.Errorf("  Expected: \"%s\" (%d)", data.expected, len(data.expected))
				t.Errorf("    Actual: \"%s\" (%d)", actual, len(actual))
			}
	}
}
