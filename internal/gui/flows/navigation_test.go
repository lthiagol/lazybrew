package flows_test

import (
	"io"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

func readOutput(t *testing.T, r io.Reader) string {
	t.Helper()
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestNavigationFlows(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(120, 40))
	tm.Send(tea.KeyPressMsg{Code: tea.KeyTab})
	tm.Send(tea.KeyPressMsg{Code: '2', Text: "2"})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	out := readOutput(t, tm.FinalOutput(t))
	if !strings.Contains(out, "Formulae") {
		t.Error("expected Formulae in output")
	}
}

func TestSearchFlow(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(120, 40))
	tm.Send(tea.KeyPressMsg{Code: '/', Text: "/"})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEsc})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
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

func TestRefreshFlow(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(120, 40))
	tm.Send(tea.KeyPressMsg{Code: 'R', Text: "R"})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

func TestModalFlow(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(120, 40))
	tm.Send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEsc})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

// TestSearchTabCycleFlow (lb19 AC-01): Tab and Shift+Tab keep moving between
// panels with the Search input focused, and the View shows each panel's name.
func TestSearchTabCycleFlow(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(120, 40))
	tm.Send(tea.KeyPressMsg{Code: '/', Text: "/"})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyTab})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	out := readOutput(t, tm.FinalOutput(t))
	if !strings.Contains(out, "Services") {
		t.Error("Shift+Tab from Status should reach Services; the View never showed it")
	}
}

// mainTitleMarker is the main box's title in a frame: the sidebar boxes have
// their number before the name, so only the main box's corner sits right
// before the active panel's name.
func mainTitleMarker(panel string) string {
	return "╮╭─ " + panel + " "
}

// finalView runs the program over keys, quits, and returns the View of the
// model it ended with. The renderer writes cursor-positioned diffs, not whole
// frames, so the state after each key is read from the model's View instead.
func finalView(t *testing.T, keys ...tea.KeyPressMsg) string {
	t.Helper()
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(120, 40))
	for _, k := range keys {
		tm.Send(k)
	}
	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	return ansi.Strip(tm.FinalModel(t, teatest.WithFinalTimeout(2*time.Second)).View().Content)
}

// TestSearchTabCyclesAllPanelsFlow (lb19 AC-01, F-r1-01): with the Search input
// focused, Tab walks through all seven panels and Shift+Tab walks back, and the
// View shows the expected panel as the main title after every key.
func TestSearchTabCyclesAllPanelsFlow(t *testing.T) {
	tab := tea.KeyPressMsg{Code: tea.KeyTab}
	shiftTab := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	keys := []tea.KeyPressMsg{{Code: '/', Text: "/"}}
	panels := []string{"Search"}
	for _, p := range []string{"Status", "Formulae", "Casks", "Outdated", "Taps", "Services", "Search"} {
		keys, panels = append(keys, tab), append(panels, p)
	}
	for _, p := range []string{"Services", "Taps", "Outdated", "Casks", "Formulae", "Status", "Search"} {
		keys, panels = append(keys, shiftTab), append(panels, p)
	}
	for i, want := range panels {
		if got := finalView(t, keys[:i+1]...); !strings.Contains(got, mainTitleMarker(want)) {
			t.Fatalf("after %d keys: main title is not %s", i+1, want)
		}
	}
}
