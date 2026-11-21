package presentation

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// TestDetailAlignsLabelsAndOmitsEmptyValues: one label column per view, and a
// field with no value is left out rather than printed as `Label:` alone.
func TestDetailAlignsLabelsAndOmitsEmptyValues(t *testing.T) {
	got := ansi.Strip(Detail([]Field{
		{Label: "Name", Value: "bash"},
		{Label: "Tap", Value: ""},
		{Label: "Homepage", Value: "https://www.gnu.org/software/bash/"},
	}, 80))
	want := "Name:      bash\nHomepage:  https://www.gnu.org/software/bash/"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestDetailCutsLongValuesWithEllipsis(t *testing.T) {
	got := ansi.Strip(Detail([]Field{
		{Label: "Homepage", Value: "https://example.test/a/very/long/path/indeed"},
	}, 24))
	if w := len([]rune(got)); w > 24 {
		t.Fatalf("row is %d cells, want at most 24: %q", w, got)
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("a cut value should end with …, got %q", got)
	}
}

// TestDetailWrapsProseWithHangingIndent: a prose value wraps under the value
// column, so every continuation line starts where the value starts.
func TestDetailWrapsProseWithHangingIndent(t *testing.T) {
	got := ansi.Strip(Detail([]Field{
		{Label: "Description", Value: "one two three four five six seven eight nine ten", Prose: true},
	}, 30))
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatalf("prose should wrap onto several rows, got %q", got)
	}
	indent := strings.Repeat(" ", len("Description:")+2)
	if !strings.HasPrefix(lines[0], "Description:  ") {
		t.Fatalf("first row should carry the label, got %q", lines[0])
	}
	for _, line := range lines[1:] {
		if !strings.HasPrefix(line, indent) {
			t.Fatalf("continuation %q should hang under the value column", line)
		}
	}
}

// TestDetailWidthZeroDoesNotCut: a caller that cuts the finished text itself
// asks for no cut, so the value is whole.
func TestDetailWidthZeroDoesNotCut(t *testing.T) {
	long := strings.Repeat("x", 120)
	got := ansi.Strip(Detail([]Field{{Label: "Prefix", Value: long}}, 0))
	if !strings.HasSuffix(got, long) {
		t.Fatalf("width 0 should keep the whole value, got %d cells", len([]rune(got)))
	}
}

// TestDetailRoleColoursTheValue: the value is drawn in its role; a plain value
// is drawn in the Text role, a URL in the Info role.
func TestDetailRoleColoursTheValue(t *testing.T) {
	plain := Detail([]Field{{Label: "Name", Value: "bash"}}, 80)
	link := Detail([]Field{{Label: "Name", Value: "bash", Role: style.InfoColor}}, 80)
	if plain == link {
		t.Fatalf("a role should change the value's colour: %q", plain)
	}
	if !strings.Contains(link, "\x1b[") {
		t.Fatalf("a coloured value should carry SGR codes: %q", link)
	}
}

func TestFormulaStatusRowsCarryIconAndWord(t *testing.T) {
	cases := []struct {
		name string
		f    brew.Formula
		want string
	}{
		{"installed", brew.Formula{Name: "bash", Version: "5.2"}, "✓ installed"},
		{"outdated", brew.Formula{Name: "bash", Version: "5.2", Outdated: true, NewVersion: "5.3"}, "↑ outdated 5.2 → 5.3"},
		{"pinned", brew.Formula{Name: "bash", Version: "5.2", Pinned: true, Outdated: true, NewVersion: "5.3"}, "• pinned"},
		{"keg-only", brew.Formula{Name: "bash", Version: "5.2", KegOnly: true}, "✓ installed · keg-only"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ansi.Strip(FormatFormulaInfo(c.f, 60, style.UnicodeIcons))
			if !strings.Contains(got, c.want) {
				t.Fatalf("want %q in:\n%s", c.want, got)
			}
		})
	}
}

func TestCaskStatusNotesAutoUpdate(t *testing.T) {
	got := ansi.Strip(FormatCaskInfo(brew.Cask{Name: "firefox", Version: "135", AutoUpdates: true}, 60, style.UnicodeIcons))
	if !strings.Contains(got, "✓ installed · auto-update") {
		t.Fatalf("want the auto-update note, got:\n%s", got)
	}
}

// TestInfoWithoutIconsDrawsNoGlyph: show_icons: false leaves the word alone.
func TestInfoWithoutIconsDrawsNoGlyph(t *testing.T) {
	got := ansi.Strip(FormatFormulaInfo(brew.Formula{Name: "bash", Version: "5.2"}, 60, style.Icons{}))
	if !strings.Contains(got, "Status:   installed") {
		t.Fatalf("no icon should leave the word alone, got:\n%s", got)
	}
}

// TestCaskInfoShowsDisplayNameBesideToken (LB-12): the Name row is the token
// brew acts on, and the display name sits under it when it reads differently.
func TestCaskInfoShowsDisplayNameBesideToken(t *testing.T) {
	got := ansi.Strip(FormatCaskInfo(brew.Cask{Name: "claude-code", DisplayName: "Claude Code", Version: "2.1.311"}, 60, style.UnicodeIcons))
	if !strings.Contains(got, "claude-code") || !strings.Contains(got, "Display") || !strings.Contains(got, "Claude Code") {
		t.Fatalf("want the token and its display name, got:\n%s", got)
	}
}

func TestCaskInfoOmitsDisplayRowWhenSameAsToken(t *testing.T) {
	got := ansi.Strip(FormatCaskInfo(brew.Cask{Name: "firefox", DisplayName: "firefox", Version: "135"}, 60, style.UnicodeIcons))
	if strings.Contains(got, "Display") {
		t.Fatalf("no Display row when the display name equals the token, got:\n%s", got)
	}
}
