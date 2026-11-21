package gui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// missingRunner answers `brew missing` with out and err, and counts the calls.
func missingRunner(out string, err error, calls *int) *brewtest.MockRunner {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 && args[0] == "missing" {
			*calls++
			return []byte(out), err
		}
		return []byte{}, nil
	}
	return r
}

// pressStatus presses key on a loaded Status panel and lands what it starts,
// as the program does.
func pressStatus(m *Model, key string) *Model {
	nm, cmd := m.Update(tea.KeyPressMsg{Code: []rune(key)[0], Text: key})
	return landCmd(modelOf(nm), cmd)
}

// landCmd runs cmd, feeds each message back through Update, and runs the
// commands those produce, to a small depth.
func landCmd(m *Model, cmd tea.Cmd) *Model {
	for depth := 4; depth > 0; depth-- {
		msgs := drainCmds(cmd)
		cmd = nil
		for _, msg := range msgs {
			nm, next := m.Update(msg)
			m = modelOf(nm)
			cmd = tea.Batch(cmd, next)
		}
	}
	return m
}

// loadedStatus is the Status panel after its dashboard landed, at size w x h.
func loadedStatus(w, h int) *Model {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
	return updateModel(m, statusLoaded())
}

// TestMissingViewRows: each missing dependency is one row, and a clean run says
// the dependencies are satisfied in the success style.
func TestMissingViewRows(t *testing.T) {
	deps := []brew.MissingDep{
		{Formula: "openssl@3", Missing: "ca-certificates"},
		{Formula: "node", Missing: "icu4c"},
	}
	out := ansi.Strip(missingView(deps, style.UnicodeIcons))
	want := "⚠ openssl@3: missing ca-certificates\n⚠ node: missing icu4c"
	if out != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out, want)
	}
	clean := ansi.Strip(missingView(nil, style.UnicodeIcons))
	if clean != "✔ All dependencies satisfied" {
		t.Fatalf("a clean run should say so, got %q", clean)
	}
}

// TestMissingTabFetchesRows: opening the Missing tab runs `brew missing` and
// shows its rows; exit 1 with output is the warning case, not an error.
func TestMissingTabFetchesRows(t *testing.T) {
	calls := 0
	client := brew.NewClient(missingRunner("openssl@3: ca-certificates\n", &brew.BrewExitError{Command: "missing", ExitCode: 1}, &calls))

	cmd := fetchTabContentCmd(client, PanelStatus, statusMissingTab, statusItemName, style.Icons{})
	tMsg, ok := cmd().(TabContentMsg)
	if !ok {
		t.Fatalf("expected TabContentMsg")
	}
	if tMsg.Err != nil {
		t.Fatalf("exit=1 must not be Err on the Missing tab, got %v", tMsg.Err)
	}
	if !strings.Contains(tMsg.Content, "openssl@3: missing ca-certificates") {
		t.Errorf("expected the missing row, got %q", tMsg.Content)
	}
}

// TestMissingTabRealErrorShowsError: a failure that is not exit 1 reaches the
// tab as an error, like the Doctor tab.
func TestMissingTabRealErrorShowsError(t *testing.T) {
	calls := 0
	client := brew.NewClient(missingRunner("", &brew.BrewExitError{Command: "missing", ExitCode: 2}, &calls))

	tMsg := fetchTabContentCmd(client, PanelStatus, statusMissingTab, statusItemName, style.Icons{})().(TabContentMsg)
	if tMsg.Err == nil {
		t.Fatal("expected Err on the Missing tab for exit=2")
	}
	m := updateModel(loadedStatus(140, 40), tMsg)
	if key := tabKey(PanelStatus, statusMissingTab, statusItemName); !strings.Contains(m.tabContent[key], "Error") {
		t.Errorf("expected Error in the tab body, got %q", m.tabContent[key])
	}
}

// TestMissingTabLoadsOnceAndStays: the tab loads when opened, is kept while the
// panel is switched away and back, and reloads only on R.
func TestMissingTabLoadsOnceAndStays(t *testing.T) {
	calls := 0
	m := loadedStatus(140, 40)
	m.client = brew.NewClient(missingRunner("wget: openssl@3\n", &brew.BrewExitError{Command: "missing", ExitCode: 1}, &calls))

	m = pressStatus(m, "l")
	m = pressStatus(m, "l")
	m = pressStatus(m, "l")
	if m.activeTab != statusMissingTab {
		t.Fatalf("three l presses should reach the Missing tab, got tab %d", m.activeTab)
	}
	if calls != 1 {
		t.Fatalf("opening the tab ran brew missing %d times, want 1", calls)
	}

	m = pressStatus(m, "2")
	m = pressStatus(m, "1")
	m = pressStatus(m, "l")
	m = pressStatus(m, "l")
	m = pressStatus(m, "l")
	if calls != 1 {
		t.Fatalf("returning to the tab ran brew missing again (%d calls), want the cached 1", calls)
	}
	if out := ansi.Strip(m.View().Content); !strings.Contains(out, "wget: missing openssl@3") {
		t.Fatalf("the cached rows should still show:\n%s", out)
	}

	nm, cmd := m.Update(RefreshMsg{LoadOutdated: true, ReloadDiagnostics: true})
	landCmd(modelOf(nm), cmd)
	if calls != 2 {
		t.Fatalf("R should reload the Missing tab once, ran brew missing %d times", calls)
	}
}

// TestMissingTabFrames: the Missing tab draws a frame of the terminal's size in
// both themes, with its rows or the satisfied line.
func TestMissingTabFrames(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	cases := []struct {
		name string
		out  string
		err  error
		want string
	}{
		{"rows", "wget: openssl@3\n", &brew.BrewExitError{Command: "missing", ExitCode: 1}, "wget: missing openssl@3"},
		{"satisfied", "", nil, "All dependencies satisfied"},
	}
	for _, size := range []struct{ w, h int }{{80, 24}, {140, 40}} {
		for theme, th := range map[string]*style.Theme{"dark": style.DarkTheme(), "light": style.LightTheme()} {
			style.ApplyTheme(th)
			for _, c := range cases {
				t.Run(fmt.Sprintf("%s/%s/%dx%d", c.name, theme, size.w, size.h), func(t *testing.T) {
					calls := 0
					m := loadedStatus(size.w, size.h)
					m.client = brew.NewClient(missingRunner(c.out, c.err, &calls))
					m = pressStatus(m, "l")
					m = pressStatus(m, "l")
					m = pressStatus(m, "l")
					view := m.View().Content
					assertFrameSize(t, view, size.w, size.h)
					if !strings.Contains(ansi.Strip(view), c.want) {
						t.Errorf("want %q in the frame:\n%s", c.want, ansi.Strip(view))
					}
				})
			}
		}
	}
}
