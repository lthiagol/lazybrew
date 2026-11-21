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

// footerRowAt140 quits the flow and returns the footer row of the final frame at
// 140 columns, with its escape codes stripped.
func footerRowAt140(t *testing.T, tm *teatest.TestModel) string {
	t.Helper()
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	view := ansi.Strip(tm.FinalModel(t).View().Content)
	rows := strings.Split(view, "\n")
	return strings.TrimRight(rows[len(rows)-1], " ")
}

// TestFooterFlowStatusAt140 (lb29 AC-03, r2 F-r1-04): Status at 140 columns shows
// its own keys, the dim rule, then the global keys with / search and R refresh once.
func TestFooterFlowStatusAt140(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(140, 40))
	got := footerRowAt140(t, tm)
	for _, want := range []string{
		" c cleanup · d doctor · A autoremove · B brewfile  │  ",
		"←↑↓→ navigate · tab next · 1-7 jump · / search · R refresh · C log · q quit · ? more",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Status footer at 140 columns = %q, want it to contain %q", got, want)
		}
	}
}

// TestFooterFlowFormulaeAt140 (lb29 AC-03): Formulae at 140 columns shows its own
// actions, the dim rule, then the global keys with / filter where / search is.
func TestFooterFlowFormulaeAt140(t *testing.T) {
	tm := testutil.NewTestModel(t, teatest.WithInitialTermSize(140, 40))
	tm.Send(tea.KeyPressMsg{Code: '2', Text: "2"})
	got := footerRowAt140(t, tm)
	for _, want := range []string{
		" i install · x remove · r reinstall · u upgrade  │  ",
		"←↑↓→ navigate · tab next · 1-7 jump · / filter · R refresh · C log · q quit · ? more",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Formulae footer at 140 columns = %q, want it to contain %q", got, want)
		}
	}
}
