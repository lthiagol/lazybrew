package gui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// tapStates are the four states a tap can be in on the Taps list (v2-10
// AC-01): official, trusted, untrusted, and unread (the trust state unknown).
var tapStates = []struct {
	name       string
	tap        brew.Tap
	trustKnown bool
	keys       string
}{
	{"untrusted", brew.Tap{Name: "user/tool"}, true, "t trust · x untap · r repair"},
	{"trusted", brew.Tap{Name: "user/tool", Trusted: true}, true, "t untrust · x untap · r repair"},
	{"official", brew.Tap{Name: "homebrew/core", IsOfficial: true}, true, "r repair"},
	{"unread", brew.Tap{Name: "user/tool"}, false, "x untap · r repair"},
}

func TestTapActionHintsPerState(t *testing.T) {
	for _, c := range tapStates {
		t.Run(c.name, func(t *testing.T) {
			var got []string
			for _, h := range tapActionHints(c.tap, c.trustKnown) {
				got = append(got, h.key+" "+h.desc)
			}
			if line := strings.Join(got, " · "); line != c.keys {
				t.Errorf("hints = %q, want %q", line, c.keys)
			}
		})
	}
}

// tapsModel is the Taps panel with the given taps, selected on the first.
func tapsModel(taps []brew.Tap, trustKnown bool) *Model {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "5")
	p := m.panels[PanelTaps]
	p.loading = false
	p.loaded = true
	p.taps = taps
	p.setTapRows()
	p.items = make([]string, len(taps))
	for i, t := range taps {
		p.items[i] = t.Name
	}
	p.trustUnknown = !trustKnown
	p.selected = 0
	return m
}

// TestTapInfoKeyLineShowsTheActionsThatApply: the Tap Info tab ends with the
// key line for the selected tap's state.
func TestTapInfoKeyLineShowsTheActionsThatApply(t *testing.T) {
	for _, c := range tapStates {
		t.Run(c.name, func(t *testing.T) {
			m := tapsModel([]brew.Tap{c.tap}, c.trustKnown)
			out := ansi.Strip(renderTapInfo(m.panels[PanelTaps], 100))
			if !strings.Contains(out, c.keys) {
				t.Errorf("Tap Info should show %q:\n%s", c.keys, out)
			}
		})
	}
}

// TestTapsFooterAndInfoAgreeOnActions: the footer and the Tap Info key line are
// made from the same state, so they name the same actions.
func TestTapsFooterAndInfoAgreeOnActions(t *testing.T) {
	for _, c := range tapStates {
		t.Run(c.name, func(t *testing.T) {
			m := tapsModel([]brew.Tap{c.tap}, c.trustKnown)
			info := ansi.Strip(renderTapInfo(m.panels[PanelTaps], 100))
			for _, h := range m.panelKeyHints() {
				if h.key == "j/k" || h.key == "h/l" || h.key == "a" || h.key == "y" {
					continue
				}
				if !strings.Contains(info, h.key+" "+h.desc) {
					t.Errorf("footer offers %q %q, the Tap Info key line lacks it:\n%s", h.key, h.desc, info)
				}
			}
		})
	}
}

// TestTapsFooterLeadsWithTheTapActions: the state's actions come first, then
// add tap, which is always there. The list keys are on the arrow hint, not here.
func TestTapsFooterLeadsWithTheTapActions(t *testing.T) {
	m := tapsModel([]brew.Tap{{Name: "user/tool"}}, true)
	var got []string
	for _, h := range m.panelKeyHints() {
		got = append(got, h.key+" "+h.desc)
	}
	want := []string{"t trust", "x untap", "r repair", "a add tap"}
	if strings.Join(got[:len(want)], "|") != strings.Join(want, "|") {
		t.Fatalf("Taps footer = %v, want it to start with %v", got, want)
	}
}

// TestTapTrustTabNamesTheKeysThatAct (v2-10 F-r1-01): the Trust tab, as drawn,
// names t for the tap and T for its formulae and casks, and only where those
// keys act. An official tap names neither; an unread one names R.
func TestTapTrustTabNamesTheKeysThatAct(t *testing.T) {
	const tapKeys = "Press t to trust or untrust the tap, and T to trust one of its formulae or casks."
	for _, c := range []struct {
		name       string
		tap        brew.Tap
		trustKnown bool
		want       string
		absent     []string
	}{
		{"untrusted", brew.Tap{Name: "user/tool"}, true, tapKeys, []string{"Press R"}},
		{"trusted", brew.Tap{Name: "user/tool", Trusted: true}, true, tapKeys, []string{"Press R"}},
		{"official", brew.Tap{Name: "homebrew/core", IsOfficial: true}, true, "", []string{"Press t", "Press T", "Press R"}},
		{"unread", brew.Tap{Name: "user/tool"}, false, "Press R to read the trust state again.", []string{"Press t", "Press T"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := tapsModel([]brew.Tap{c.tap}, c.trustKnown)
			m = sendKey(m, "l")
			out := ansi.Strip(m.View().Content)
			if !strings.Contains(out, "Trust:") {
				t.Fatalf("the Trust tab is not on screen:\n%s", out)
			}
			if c.want != "" && !strings.Contains(strings.Join(strings.Fields(out), " "), c.want) {
				t.Errorf("Trust tab should say %q:\n%s", c.want, out)
			}
			for _, bad := range c.absent {
				if strings.Contains(out, bad) {
					t.Errorf("Trust tab should not say %q:\n%s", bad, out)
				}
			}
		})
	}
}

// tapAddInput sends an add-tap input through the Taps flow with a fake runner.
// It returns the brew arguments that ran and whether the task manager is idle
// (nothing running, nothing queued).
func tapAddInput(t *testing.T, value string) ([][]string, bool) {
	t.Helper()
	var ran [][]string
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		ran = append(ran, args)
		return []byte(""), nil
	}
	m := New(brew.NewClient(r), testConfig())
	m.pendingAction = "tap-add"
	res, cmd := m.handleModalResult(&modal.InputResult{Value: value}, nil)
	nm := res.(Model)
	if cmd != nil {
		if _, ok := cmd().(TaskStartedMsg); !ok {
			t.Fatalf("expected TaskStartedMsg, got %T", cmd())
		}
		if _, ok := nm.tasks.RunNext()().(TaskCompletedMsg); !ok {
			t.Fatal("expected TaskCompletedMsg")
		}
	}
	return ran, nm.tasks.Idle()
}

// TestTapAddTrimsTrailingSlashAndSpaces: a typed `user/repo/` or a value with
// surrounding spaces runs `brew tap user/repo`, since brew rejects the slash.
func TestTapAddTrimsTrailingSlashAndSpaces(t *testing.T) {
	for _, value := range []string{"anomalyco/tap/", " anomalyco/tap/ ", "anomalyco/tap//"} {
		t.Run(value, func(t *testing.T) {
			ran, _ := tapAddInput(t, value)
			if len(ran) != 1 || strings.Join(ran[0], " ") != "tap anomalyco/tap" {
				t.Fatalf("brew calls = %v, want [tap anomalyco/tap]", ran)
			}
		})
	}
}

// TestTapAddEnqueuesNothingWhenEmptyAfterTrim: a value that is only slashes or
// whitespace runs no brew command and leaves the task manager idle.
func TestTapAddEnqueuesNothingWhenEmptyAfterTrim(t *testing.T) {
	for _, value := range []string{"/", "//", "   ", " / "} {
		t.Run(value, func(t *testing.T) {
			ran, idle := tapAddInput(t, value)
			if len(ran) != 0 {
				t.Errorf("brew calls = %v, want none", ran)
			}
			if !idle {
				t.Error("the task manager has a task, want none enqueued")
			}
		})
	}
}
