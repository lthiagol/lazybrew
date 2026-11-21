package flows_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/gui"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

// newFilterRunner is the uninstall runner with the installed formulae the given
// names, in that order.
func newFilterRunner(names ...string) *uninstallRecordingRunner {
	rr := newUninstallRecordingRunner()
	var formulae []map[string]string
	for _, n := range names {
		formulae = append(formulae, map[string]string{"name": n, "full_name": n, "desc": "", "tap": "homebrew/core"})
	}
	raw, _ := json.Marshal(map[string]any{"formulae": formulae})
	rr.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		if len(args) >= 2 && args[0] == "info" {
			return json.Unmarshal(raw, result)
		}
		return nil
	}
	return rr
}

// TestFilterUninstallsTheFilteredRow: on a kept filter, x asks about the row the
// cursor is on, and the brew call names that row.
func TestFilterUninstallsTheFilteredRow(t *testing.T) {
	// Unfiltered, the cursor's first row is ab; filtered by f, it is fd, and one
	// Down is fzf. So the prompt names fzf only if the filter is applied.
	runner := newFilterRunner("ab", "fd", "fzf")
	tm := testutil.NewTestModelWithRunner(t, runner, teatest.WithInitialTermSize(120, 40))

	tm.Send(tea.KeyPressMsg{Code: '2', Text: "2"})
	waitForScreen(t, tm, "fzf")
	tm.Send(tea.KeyPressMsg{Code: '/', Text: "/"})
	tm.Send(tea.KeyPressMsg{Code: 'f', Text: "f"})
	waitForScreen(t, tm, "/ f")
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyDown})
	tm.Send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	waitForScreen(t, tm, "Uninstall fzf?")
	tm.Send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	waitForCall(t, runner.getStreamCalls, "uninstall", "fzf")
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// TestFilterEscClearsTheList: Esc on a kept filter clears it, and the final
// screen lists every formula again.
func TestFilterEscClearsTheList(t *testing.T) {
	runner := newFilterRunner("ab", "fd", "fzf")
	tm := testutil.NewTestModelWithRunner(t, runner, teatest.WithInitialTermSize(120, 40))

	tm.Send(tea.KeyPressMsg{Code: '2', Text: "2"})
	waitForScreen(t, tm, "fzf")
	tm.Send(tea.KeyPressMsg{Code: '/', Text: "/"})
	tm.Send(tea.KeyPressMsg{Code: 'f', Text: "f"})
	waitForScreen(t, tm, "/ f")
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEsc})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))

	final := tm.FinalModel(t)
	fm, ok := final.(gui.Model)
	if !ok {
		t.Fatalf("final model is %T, want gui.Model", final)
	}
	view := ansi.Strip(fm.View().Content)
	// The footer's "/ filter" also starts "/ f", so the filter row is "/ f " with
	// its trailing space.
	if strings.Contains(view, "2/3") || strings.Contains(view, "/ f ") {
		t.Fatalf("Esc should clear the filter; the final screen still shows it:\n%s", view)
	}
	for _, name := range []string{"ab", "fd", "fzf"} {
		if !strings.Contains(view, name) {
			t.Errorf("the cleared list should show %s:\n%s", name, view)
		}
	}
}
