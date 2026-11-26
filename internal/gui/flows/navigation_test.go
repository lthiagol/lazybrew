package flows_test

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

// TestNavigationFlows: 2 opens Formulae, and Tab from there lands on Casks. Only
// both keys in order leave the final main title on Casks.
func TestNavigationFlows(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(120, 40))
	tm.Send(tea.KeyPressMsg{Code: '2', Text: "2"})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyTab})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	view := ansi.Strip(tm.FinalModel(t).View().Content)
	if !strings.Contains(view, mainTitleMarker("Casks")) {
		t.Errorf("2 then Tab should land on Casks:\n%s", view)
	}
}

// TestSearchFlow: / opens the Search panel, and Esc takes the focus off its box,
// so 2 jumps to Formulae rather than typing into the box.
func TestSearchFlow(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(120, 40))
	tm.Send(tea.KeyPressMsg{Code: '/', Text: "/"})
	waitForScreen(t, tm, "Search packages...")
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEsc})
	tm.Send(tea.KeyPressMsg{Code: '2', Text: "2"})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	view := ansi.Strip(tm.FinalModel(t).View().Content)
	if !strings.Contains(view, mainTitleMarker("Formulae")) {
		t.Errorf("after Esc, 2 should reach Formulae; the box still had focus:\n%s", view)
	}
}

func TestHelpFlow(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(120, 40))
	tm.Send(tea.KeyPressMsg{Code: '?', Text: "?"})
	// Assert overlay-only copy before Esc/q — footer always has "? help",
	// which must not count as proof the Help overlay opened.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return strings.Contains(s, "?/esc close") ||
			(strings.Contains(s, "Help") && strings.Contains(s, "close"))
	}, teatest.WithDuration(2*time.Second), teatest.WithCheckInterval(50*time.Millisecond))
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEsc})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

// TestRefreshFlow: R re-reads the loaded Formulae list, so the row the list
// gains after 2 shows once R's read lands.
func TestRefreshFlow(t *testing.T) {
	runner := newFilterRunner("testformula")
	before := runner.ExecuteJSONFn
	after := newFilterRunner("reloaded").ExecuteJSONFn
	var reloaded atomic.Bool
	runner.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		if reloaded.Load() {
			return after(ctx, result, args...)
		}
		return before(ctx, result, args...)
	}
	tm := testutil.NewTestModelWithRunner(t, runner, teatest.WithInitialTermSize(120, 40))

	tm.Send(tea.KeyPressMsg{Code: '2', Text: "2"})
	waitForScreen(t, tm, "▸ ✔ testformula")
	reloaded.Store(true)
	tm.Send(tea.KeyPressMsg{Code: 'R', Text: "R"})
	waitForScreen(t, tm, "reloaded")
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// TestModalFlow: x on Formulae opens the uninstall confirm, and Esc closes it.
func TestModalFlow(t *testing.T) {
	runner := newUninstallRecordingRunner()
	tm := testutil.NewTestModelWithRunner(t, runner, teatest.WithInitialTermSize(120, 40))

	tm.Send(tea.KeyPressMsg{Code: '2', Text: "2"})
	waitForScreen(t, tm, "▸ ✔ testformula")
	tm.Send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	waitForScreen(t, tm, "Uninstall testformula?")
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEsc})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
	view := ansi.Strip(tm.FinalModel(t).View().Content)
	if strings.Contains(view, "Uninstall testformula?") {
		t.Errorf("Esc should close the uninstall confirm:\n%s", view)
	}
}

// mainTitleMarker is the main box's title in a frame: the sidebar boxes have
// their number before the name, so only the main box's corner sits right
// before the active panel's name.
func mainTitleMarker(panel string) string {
	return "╮╭─ " + panel + " "
}

// TestSearchTabCyclesAllPanelsFlow (lb19 AC-01, F-r1-01): with the Search input
// focused, Tab walks through all seven panels and Shift+Tab walks back, and the
// View shows the expected panel as the main title after every key. The model is
// driven directly, one Update and one View per key, so the walk needs no program.
func TestSearchTabCyclesAllPanelsFlow(t *testing.T) {
	type step struct {
		key  tea.KeyPressMsg
		want string
	}
	tab := tea.KeyPressMsg{Code: tea.KeyTab}
	shiftTab := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	steps := []step{{tea.KeyPressMsg{Code: '/', Text: "/"}, "Search"}}
	for _, p := range []string{"Status", "Formulae", "Casks", "Outdated", "Taps", "Services", "Search"} {
		steps = append(steps, step{tab, p})
	}
	for _, p := range []string{"Services", "Taps", "Outdated", "Casks", "Formulae", "Status", "Search"} {
		steps = append(steps, step{shiftTab, p})
	}

	var m tea.Model = testutil.NewModel(brewtest.NewMockRunner())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	for i, st := range steps {
		m, _ = m.Update(st.key)
		if got := ansi.Strip(m.View().Content); !strings.Contains(got, mainTitleMarker(st.want)) {
			t.Fatalf("after %d keys: main title is not %s", i+1, st.want)
		}
	}
}
