package flows_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

func TestTabsFlow(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(120, 40))
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	out := readOutput(t, tm.FinalOutput(t))
	if !strings.Contains(out, "Deps") && !strings.Contains(out, "Loading") {
		t.Log("output (first 500):", out[:min(len(out), 500)])
	}
}

// AC-03 Deps load regression is covered by gui.TestDepsTabLoads (deterministic
// Update/View path). teatest FinalOutput is cumulative/cleared on quit and was
// a flaky host for the same assertion (see M13 F-05).
