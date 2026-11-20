package textutil

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestClip(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		w     int
		want  string
		width int
	}{
		{"fits", "hello", 5, "hello", 5},
		{"plain cut", "hello world", 6, "hello…", 6},
		{"wide runes cut on a cell boundary", "日本語テキスト", 5, "日本…", 5},
		{"wide rune that would split is dropped", "日本語", 3, "日…", 3},
		{"width zero", "hello", 0, "", 0},
		{"negative width", "hello", -2, "", 0},
		{"width one is the ellipsis", "hello", 1, "…", 1},
		{"width one fits a single cell", "h", 1, "h", 1},
		{"empty input", "", 4, "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Clip(tc.in, tc.w)
			if got != tc.want {
				t.Fatalf("Clip(%q, %d) = %q, want %q", tc.in, tc.w, got, tc.want)
			}
			if width := ansi.StringWidth(got); width != tc.width {
				t.Fatalf("Clip(%q, %d) is %d cells, want %d", tc.in, tc.w, width, tc.width)
			}
		})
	}
}

func TestClipKeepsStyles(t *testing.T) {
	const red = "\x1b[31m"
	const reset = "\x1b[0m"
	in := red + "hello world" + reset
	got := Clip(in, 6)
	if w := ansi.StringWidth(got); w != 6 {
		t.Fatalf("Clip styled = %q, %d cells, want 6", got, w)
	}
	if !strings.Contains(got, red) {
		t.Fatalf("Clip styled = %q, lost the red style", got)
	}
	if plain := ansi.Strip(got); plain != "hello…" {
		t.Fatalf("Clip styled text = %q, want hello…", plain)
	}
}

func TestFit(t *testing.T) {
	cases := []struct {
		name string
		in   string
		w    int
		want string
	}{
		{"pads a narrow line", "ab", 5, "ab   "},
		{"cuts a wide line", "abcdef", 4, "abc…"},
		{"exact width is unchanged", "abcd", 4, "abcd"},
		{"wide runes pad to the cell", "日本語", 4, "日… "},
		{"width zero", "abc", 0, ""},
		{"width one cuts to the ellipsis", "abc", 1, "…"},
		{"width one pads an empty line", "", 1, " "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Fit(tc.in, tc.w)
			if got != tc.want {
				t.Fatalf("Fit(%q, %d) = %q, want %q", tc.in, tc.w, got, tc.want)
			}
		})
	}
}

func TestFitStyledIsExactWidth(t *testing.T) {
	in := "\x1b[1mbold\x1b[0m and more text"
	for w := 0; w <= 12; w++ {
		if got := ansi.StringWidth(Fit(in, w)); got != w {
			t.Fatalf("Fit(%q, %d) is %d cells, want %d", in, w, got, w)
		}
	}
}
