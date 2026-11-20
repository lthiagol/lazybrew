package gui

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// A sidebar error is cut to the box width by display cells, never by bytes: a
// byte cut through a multibyte character leaves invalid UTF-8 on screen (C12).
func TestSidebarErrorCutKeepsValidUTF8(t *testing.T) {
	// The first byte cut at width-5 = 15 lands inside the three-byte エ.
	p := &panelData{err: errors.New("a日本語のエラーメッセージ")}
	const width = 20

	out := ansi.Strip(p.renderSidebarContent(width, 5, nil, style.Icons{}, true))
	if !utf8.ValidString(out) {
		t.Fatalf("error cut split a rune:\n%q", out)
	}
	if !strings.Contains(out, "...") {
		t.Errorf("a long error should be cut with an ellipsis:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if w := ansi.StringWidth(line); w > width {
			t.Errorf("line is %d cells wide, want at most %d: %q", w, width, line)
		}
	}
}
