package gui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

const testPrefix = "/home/linuxbrew/.linuxbrew"

// previewPreview is a cleanup preview of three paths, one with no size, and two
// unused formulae, in the shape brew prints.
func previewPreview() brew.CleanupPreview {
	return brew.CleanupPreview{
		Total: "950MB",
		Items: []brew.CleanupItem{
			{Path: testPrefix + "/Cellar/old-formula/1.0", Files: 3, Size: "230.3MB", Bytes: 230300000},
			{Path: "/home/user/.cache/Homebrew/downloads/0a1b-foo-1.0.tar.gz", Files: 1, Size: "48.2MB", Bytes: 48200000},
			{Path: testPrefix + "/bin/gone", Kind: "broken link"},
		},
		Unused: []string{"cmake", "openssl@3"},
	}
}

// TestCleanupViewTable: the header is brew's total, the rows run biggest first,
// the paths are shortened to <prefix>/ and ~/, and the note names the unused.
func TestCleanupViewTable(t *testing.T) {
	out := ansi.Strip(cleanupView(previewPreview(), 100, style.UnicodeIcons, "/home/user", testPrefix))
	lines := strings.Split(out, "\n")

	if lines[0] != "Would free ~950MB" {
		t.Fatalf("header = %q, want %q", lines[0], "Would free ~950MB")
	}
	wantRows := []string{
		"230.3MB  3 files      <prefix>/Cellar/old-formula/1.0",
		" 48.2MB  1 file       ~/.cache/Homebrew/downloads/0a1b-foo-1.0.tar.gz",
		"      —  broken link  <prefix>/bin/gone",
	}
	for i, want := range wantRows {
		if got := lines[3+i]; got != want {
			t.Errorf("row %d:\n got %q\nwant %q", i, got, want)
		}
	}
	if last := lines[len(lines)-1]; last != "Also removes 2 unused formulae (Unused tab)" {
		t.Errorf("note = %q", last)
	}
}

// The table keeps its shape at any width: no line is wider than the tab.
func TestCleanupViewFitsWidth(t *testing.T) {
	for _, w := range []int{30, 40, 60} {
		for _, line := range strings.Split(ansi.Strip(cleanupView(previewPreview(), w, style.UnicodeIcons, "/home/user", testPrefix)), "\n") {
			if got := lipglossWidth(line); got > w {
				t.Errorf("width %d: line %q is %d cells", w, line, got)
			}
		}
	}
}

// Empty and raw-fallback: nothing parsed shows brew's own text, nothing shows the
// empty line.
func TestCleanupViewEmptyAndRaw(t *testing.T) {
	if got := ansi.Strip(cleanupView(brew.CleanupPreview{}, 80, style.UnicodeIcons, "", "")); got != "✔ Nothing to clean up" {
		t.Errorf("empty = %q", got)
	}
	raw := brew.CleanupPreview{Raw: []string{"Something brew says in a new shape"}}
	if got := cleanupView(raw, 80, style.UnicodeIcons, "", ""); got != "Something brew says in a new shape" {
		t.Errorf("raw = %q", got)
	}
}

func TestUnusedView(t *testing.T) {
	out := ansi.Strip(unusedView(brew.AutoremovePreview{Formulae: []string{"cmake", "openssl@3"}}, style.UnicodeIcons))
	if want := "2 unused formulae would be removed\ncmake\nopenssl@3"; out != want {
		t.Errorf("unused = %q, want %q", out, want)
	}
	one := ansi.Strip(unusedView(brew.AutoremovePreview{Formulae: []string{"foo"}}, style.UnicodeIcons))
	if !strings.HasPrefix(one, "1 unused formula would be removed") {
		t.Errorf("singular = %q", one)
	}
	if got := ansi.Strip(unusedView(brew.AutoremovePreview{}, style.UnicodeIcons)); got != "✔ Nothing unused" {
		t.Errorf("empty = %q", got)
	}
}

func TestShortPath(t *testing.T) {
	cases := []struct{ path, home, prefix, want string }{
		{testPrefix + "/Cellar/x/1.0", "/home/user", testPrefix, "<prefix>/Cellar/x/1.0"},
		{"/home/user/.cache/Homebrew/api", "/home/user", testPrefix, "~/.cache/Homebrew/api"},
		{"/home/user", "/home/user", testPrefix, "/home/user"},
		{"/home/usernames/x", "/home/user", testPrefix, "/home/usernames/x"},
		{"/var/tmp/x", "", "", "/var/tmp/x"},
	}
	for _, tc := range cases {
		if got := shortPath(tc.path, tc.home, tc.prefix); got != tc.want {
			t.Errorf("shortPath(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

// cutMiddle keeps both ends of a path and cuts on a separator where one is near.
func TestCutMiddle(t *testing.T) {
	const path = "~/Library/Caches/Homebrew/downloads/0a1b2c3d-some-long-formula-1.2.3.tar.gz"
	if got := cutMiddle(path, 200); got != path {
		t.Errorf("a path that fits is returned whole: %q", got)
	}
	// The cut moves to a "/" only within cutSnap cells of where the width ends.
	// Otherwise it falls at the width, so "Caches" is kept and "~/Library/" is not
	// the whole head.
	cases := map[int]string{
		33: "~/Library/Caches…ula-1.2.3.tar.gz",
		20: "~/Library…2.3.tar.gz",
	}
	for w, want := range cases {
		if got := cutMiddle(path, w); got != want {
			t.Errorf("cutMiddle(w=%d) = %q, want %q", w, got, want)
		}
	}
	if got := cutMiddle(path, 1); got != "…" && got != "~" {
		t.Errorf("cutMiddle(w=1) = %q", got)
	}
}

// A cut that already falls after a "/" keeps the whole last segment: at 22 cells
// the tail is "llvm/17.0.6", not "/17.0.6".
func TestCutMiddleKeepsWholeSegmentAtCut(t *testing.T) {
	if got, want := cutMiddle("<prefix>/Cellar/llvm/17.0.6", 22), "<prefix>/…llvm/17.0.6"; got != want {
		t.Errorf("cutMiddle = %q, want %q", got, want)
	}
}

// At 80x24 the path column is 22 cells wide (content 48, the Files column 15 for
// "empty directory"). The rows keep the part that tells them apart where it fits:
// the version of a formula, and the formula name when it is short enough.
func TestCleanupViewRowsAt80x24(t *testing.T) {
	width := contentWidthIn(loadedStatus(80, 24).screenLayout())
	p := brew.CleanupPreview{Items: []brew.CleanupItem{
		{Path: testPrefix + "/Cellar/python@3.12/3.12.4_1", Files: 1700, Size: "64.4MB", Bytes: 64400000},
		{Path: testPrefix + "/Cellar/llvm/17.0.6", Kind: "empty directory"},
	}}
	rows := cleanupTable(p.Items, width, "/Users/alice", testPrefix)
	if len(rows) != 3 {
		t.Fatalf("rows = %q, want a header and two rows", rows)
	}
	// Biggest first: the python row is the first item row.
	if want := "<prefix>/…/3.12.4_1"; !strings.HasSuffix(ansi.Strip(rows[1]), want) {
		t.Errorf("python row %q does not end with %q", rows[1], want)
	}
	if want := "<prefix>/…llvm/17.0.6"; !strings.HasSuffix(ansi.Strip(rows[2]), want) {
		t.Errorf("llvm row %q does not end with %q", rows[2], want)
	}
}

// Two download-cache bottles differ only in their hash and formula, and their names
// are wider than the path column at 80x24 (23 cells). The directories go, and the
// name keeps its start (the hash that tells the rows apart) and its end (the bottle
// suffix). A Cellar row whose last segment fits is cut as before.
func TestCleanupViewBottleRowsAt80x24(t *testing.T) {
	width := contentWidthIn(loadedStatus(80, 24).screenLayout())
	const dir = "/home/user/Library/Caches/Homebrew/downloads/"
	p := brew.CleanupPreview{Items: []brew.CleanupItem{
		{Path: dir + "0a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f--llvm-17.0.6.arm64_sonoma.bottle.tar.gz", Files: 1, Size: "48.2MB", Bytes: 48200000},
		{Path: dir + "9f8e7d6c5b4a39281706f5e4d3c2b1a09f8e7d6c5b4a39281706f5e4d3c2b1a0--python@3.12-3.12.4_1.arm64_sonoma.bottle.tar.gz", Files: 1, Size: "47.9MB", Bytes: 47900000},
		{Path: testPrefix + "/Cellar/llvm/17.0.6", Kind: "empty directory"},
	}}
	rows := cleanupTable(p.Items, width, "/home/user", testPrefix)
	if len(rows) != 4 {
		t.Fatalf("rows = %q, want a header and three rows", rows)
	}
	want := []string{
		"0a1b2c3d4e5…ttle.tar.gz",
		"9f8e7d6c5b4…ttle.tar.gz",
		"<prefix>/…llvm/17.0.6",
	}
	for i, w := range want {
		if got := ansi.Strip(rows[i+1]); !strings.HasSuffix(got, w) {
			t.Errorf("row %d %q does not end with %q", i+1, got, w)
		}
	}
}

// The Status tab row holds all seven names whole at 140x40, in both themes. At
// 80x24 it does not fit (the names need 60 cells, the bar has about 50), so that
// size is an open question, not a passing check: see Q-lb31-tabs-80.
func TestStatusTabRowWhole140x40(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	for _, theme := range []*style.Theme{style.DarkTheme(), style.LightTheme()} {
		style.ApplyTheme(theme)
		m := loadedStatus(140, 40)
		row := ""
		for _, line := range strings.Split(ansi.Strip(m.View().Content), "\n") {
			if strings.Contains(line, "Config") {
				row = line
				break
			}
		}
		for _, name := range []string{"Dashboard", "Config", "Doctor", "Missing", "Vulns", "Cleanup", "Unused"} {
			if !strings.Contains(row, name) {
				t.Errorf("tab row at 140x40 lacks %q: %q", name, row)
			}
		}
	}
}

// The frames: each of the two tabs, with data, empty and failed, at 80x24 and
// 140x40, in both themes. The copy of a tab is the same text, uncut.
func TestPreviewTabFrames(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	states := []struct {
		name string
		set  func(m *Model)
		want map[int]string
	}{
		{
			name: "data",
			set: func(m *Model) {
				p := previewPreview()
				m.previews.cleanup = &p
				u := brew.AutoremovePreview{Formulae: []string{"cmake"}}
				m.previews.unused = &u
			},
			want: map[int]string{statusCleanupTab: "Would free ~950MB", statusUnusedTab: "1 unused formula would be removed"},
		},
		{
			name: "empty",
			set: func(m *Model) {
				m.previews.cleanup = &brew.CleanupPreview{}
				m.previews.unused = &brew.AutoremovePreview{}
			},
			want: map[int]string{statusCleanupTab: "Nothing to clean up", statusUnusedTab: "Nothing unused"},
		},
		{
			name: "error",
			set: func(m *Model) {
				m.previews.cleanupErr = fmt.Errorf("brew cleanup: no network")
				m.previews.unusedErr = fmt.Errorf("brew autoremove: no network")
			},
			want: map[int]string{statusCleanupTab: "Error: brew cleanup: no network", statusUnusedTab: "Error: brew autoremove: no network"},
		},
	}
	for _, theme := range []*style.Theme{style.DarkTheme(), style.LightTheme()} {
		for _, size := range [][2]int{{80, 24}, {140, 40}} {
			for _, st := range states {
				for tab, want := range st.want {
					name := fmt.Sprintf("%s/%dx%d/%s/tab%d", map[bool]string{true: "dark", false: "light"}[theme == style.DarkTheme()], size[0], size[1], st.name, tab)
					t.Run(name, func(t *testing.T) {
						style.ApplyTheme(theme)
						m := loadedStatus(size[0], size[1])
						m.status.Prefix = testPrefix
						st.set(m)
						m.activeTab = tab
						if out := ansi.Strip(m.View().Content); !strings.Contains(out, want) {
							t.Fatalf("frame lacks %q:\n%s", want, out)
						}
					})
				}
			}
		}
	}
}

// y copies the Cleanup tab uncut: the long path is whole, not cut to the tab.
func TestCleanupTabCopyIsUncut(t *testing.T) {
	long := testPrefix + "/Cellar/a-formula-with-a-very-long-name-that-would-be-cut/1.2.3/share/some/deep/directory/file.txt"
	p := brew.CleanupPreview{Items: []brew.CleanupItem{{Path: long, Files: 1, Size: "1KB", Bytes: 1000}}}
	m := loadedStatus(80, 24)
	m.status.Prefix = testPrefix
	m.previews.cleanup = &p
	m.activeTab = statusCleanupTab
	want := "<prefix>" + long[len(testPrefix):]
	if got := m.copyText(); !strings.Contains(got, want) {
		t.Fatalf("copy cut the path:\n%s", got)
	}
}

// R drops both previews, so the tabs read brew again when drawn.
func TestClearTabContentDropsPreviews(t *testing.T) {
	m := loadedStatus(80, 24)
	p := previewPreview()
	m.previews.cleanup = &p
	m.clearTabContent()
	if m.previews.cleanupLoaded() {
		t.Fatal("clearTabContent should drop the Cleanup preview")
	}
}

// TestCleanupTabReadsOnOpen: opening the Cleanup tab runs `brew cleanup -n` once;
// moving away and back shows the read again without running it again.
func TestCleanupTabReadsOnOpen(t *testing.T) {
	calls := map[string]int{}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		calls[args[0]]++
		if args[0] == "cleanup" {
			return []byte("Would remove: /x/old/1.0 (3 files, 230.3MB)\n" +
				"==> This operation would free approximately 230.3MB of disk space.\n"), nil
		}
		return []byte{}, nil
	}
	m := loadedStatus(140, 40)
	m.client = brew.NewClient(r)
	for range 5 {
		m = pressStatus(m, "l")
	}
	if m.activeTab != statusCleanupTab {
		t.Fatalf("five l presses should reach the Cleanup tab, got tab %d", m.activeTab)
	}
	if out := ansi.Strip(m.View().Content); !strings.Contains(out, "Would free ~230.3MB") {
		t.Fatalf("the Cleanup tab should show the preview:\n%s", out)
	}
	m = pressStatus(m, "h")
	m = pressStatus(m, "l")
	if m.activeTab != statusCleanupTab {
		t.Fatalf("h then l should come back to the Cleanup tab, got tab %d", m.activeTab)
	}
	if n := calls["cleanup"]; n != 1 {
		t.Fatalf("brew cleanup ran %d times, want 1", n)
	}
}

// lipglossWidth is the cell width of a line, ignoring any escape codes.
func lipglossWidth(s string) int {
	return ansi.StringWidth(s)
}
