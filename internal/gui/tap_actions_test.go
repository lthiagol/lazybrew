package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// tapStates are the four states a tap can be in on the Taps list (v2-10
// AC-01): official, trusted, untrusted, and unread (the trust state unknown).
var tapStates = []struct {
	name       string
	tap        brew.Tap
	trustKnown bool
	keys       string
}{
	{"untrusted", brew.Tap{Name: "user/tool"}, true, "t trust  x untap  r repair"},
	{"trusted", brew.Tap{Name: "user/tool", Trusted: true}, true, "t untrust  x untap  r repair"},
	{"official", brew.Tap{Name: "homebrew/core", IsOfficial: true}, true, "r repair"},
	{"unread", brew.Tap{Name: "user/tool"}, false, "x untap  r repair"},
}

func TestTapActionHintsPerState(t *testing.T) {
	for _, c := range tapStates {
		t.Run(c.name, func(t *testing.T) {
			var got []string
			for _, h := range tapActionHints(c.tap, c.trustKnown) {
				got = append(got, h.key+" "+h.desc)
			}
			if line := strings.Join(got, "  "); line != c.keys {
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
				if h.key == "j/k" || h.key == "a" {
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
// navigate, then add tap, which is always there.
func TestTapsFooterLeadsWithTheTapActions(t *testing.T) {
	m := tapsModel([]brew.Tap{{Name: "user/tool"}}, true)
	var got []string
	for _, h := range m.panelKeyHints() {
		got = append(got, h.key+" "+h.desc)
	}
	want := []string{"t trust", "x untap", "r repair", "j/k navigate", "a add tap"}
	if strings.Join(got[:len(want)], "|") != strings.Join(want, "|") {
		t.Fatalf("Taps footer = %v, want it to start with %v", got, want)
	}
}
