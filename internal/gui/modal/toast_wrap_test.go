package modal

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// toastText is the message as the box shows it: the rows of the box, borders
// and icon included, joined and with runs of spaces made one.
func toastText(view string) (text string, rows int) {
	lines := strings.Split(ansi.Strip(view), "\n")
	var inner []string
	for _, line := range lines[1 : len(lines)-1] {
		inner = append(inner, strings.Trim(line, "│ "))
	}
	return strings.Join(strings.Fields(strings.Join(inner, " ")), " "), len(inner)
}

// A needs-sudo message is wider than a toast's cap: it wraps, and the command it
// names stays whole.
func TestToastWrapsALongMessageWhole(t *testing.T) {
	msg := `moshi-hook: needs sudo — run "sudo brew services restart moshi-hook" in a terminal`
	text, rows := toastText((&Toast{Message: msg, Type: ToastError}).View("✗", 60))

	if !strings.Contains(text, msg) {
		t.Errorf("the box shows %q, want the whole message %q", text, msg)
	}
	if rows != 2 {
		t.Errorf("rows = %d, want 2", rows)
	}
}

// A message that fits draws as one row, as it always did.
func TestToastShortMessageIsOneRow(t *testing.T) {
	text, rows := toastText((&Toast{Message: "Install completed", Type: ToastSuccess}).View("✓", 60))

	if text != "✓ Install completed" || rows != 1 {
		t.Errorf("box = %q in %d rows, want one row %q", text, rows, "✓ Install completed")
	}
}

// Past toastMaxRows the last row is cut with an ellipsis, so the box never grows
// without limit.
func TestToastLongMessageIsCutAfterThreeRows(t *testing.T) {
	msg := strings.Repeat("word ", 60)
	text, rows := toastText((&Toast{Message: msg, Type: ToastInfo}).View("", 20))

	if rows != toastMaxRows {
		t.Errorf("rows = %d, want %d", rows, toastMaxRows)
	}
	if !strings.HasSuffix(text, "…") {
		t.Errorf("the cut message %q does not end with an ellipsis", text)
	}
}
