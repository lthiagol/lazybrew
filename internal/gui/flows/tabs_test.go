package flows_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

// TestTabsFlow: 2 opens Formulae, and two l presses move its tabs on to Used By.
// The main title names the tab, so the final View shows where the keys led.
func TestTabsFlow(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(120, 40))
	tm.Send(tea.KeyPressMsg{Code: '2', Text: "2"})
	tm.Send(tea.KeyPressMsg{Code: 'l', Text: "l"})
	tm.Send(tea.KeyPressMsg{Code: 'l', Text: "l"})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	view := ansi.Strip(tm.FinalModel(t).View().Content)
	if !strings.Contains(view, mainTitleMarker("Formulae › Used By")) {
		t.Errorf("2 then two l presses should reach Formulae › Used By:\n%s", view)
	}
}

// AC-03 Deps load regression is covered by gui.TestDepsTabLoads (deterministic
// Update/View path). teatest FinalOutput is cumulative/cleared on quit and was
// a flaky host for the same assertion (see M13 F-05).
