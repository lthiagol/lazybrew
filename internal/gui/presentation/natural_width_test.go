package presentation

import (
	"testing"

	"charm.land/lipgloss/v2"
)

// TestNaturalWidthIsTheWidthOfTheUnwrappedDetail (cr3 AC-07): Info measures its
// left column with naturalWidth, which must give the width Detail(fields, 0)
// draws, without drawing it. Continuation lines of a multi-line value are not
// indented, so they count at their own width.
func TestNaturalWidthIsTheWidthOfTheUnwrappedDetail(t *testing.T) {
	cases := map[string][]Field{
		"none":                nil,
		"plain":               {{Label: "Name", Value: "bash"}, {Label: "Homepage", Value: "https://example.test/bash"}},
		"empty value skipped": {{Label: "Name", Value: "bash"}, {Label: "Caveats", Value: ""}},
		"multi-line":          {{Label: "Caveats", Value: "line one\nlonger line two", Prose: true}},
		"continuation widest": {{Label: "A", Value: "x\nsomething much longer than the label"}},
		"wide runes":          {{Label: "Desc", Value: "日本語 text"}},
		"tab":                 {{Label: "Cmd", Value: "a\tb"}},
	}
	for name, fields := range cases {
		want := lipgloss.Width(Detail(fields, 0))
		if got := naturalWidth(fields); got != want {
			t.Errorf("%s: naturalWidth = %d, want the drawn width %d", name, got, want)
		}
	}
}
