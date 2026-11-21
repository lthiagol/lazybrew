package gui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// trustedItemTabs is the tab bar of a trusted item on the Taps list: its Info
// tab alone. The tab's index is tapInfoTab, the first of the Taps tabs.
var trustedItemTabs = []tabInfo{{"Info", tapInfoTab}}

// trustedRows is the rows of the Taps list below the taps: every trusted
// formula, cask and command, in that order, alphabetical within each kind. A tap
// is a row of its own and is not repeated here.
func trustedRows(entries []brew.TrustEntry) []brew.TrustEntry {
	rows := make([]brew.TrustEntry, 0, len(entries))
	for _, kind := range []brew.TrustType{brew.TrustTypeFormula, brew.TrustTypeCask, brew.TrustTypeCommand} {
		start := len(rows)
		for _, e := range entries {
			if e.Type == kind {
				rows = append(rows, e)
			}
		}
		slices.SortFunc(rows[start:], func(a, b brew.TrustEntry) int {
			return strings.Compare(a.Name, b.Name)
		})
	}
	return rows
}

// selectTapRow moves the Taps cursor to the tap, or the trusted item, called name
// in the list the cursor was on. A name the reload no longer has leaves the cursor
// where it was; normalize then clamps it to the rows that are left.
func selectTapRow(p *panelData, name string, trusted bool) {
	if name == "" {
		return
	}
	if !trusted {
		selectNamed(p, p.taps, func(t brew.Tap) string { return t.Name }, name)
		return
	}
	for i, e := range p.trusted {
		if e.Name == name {
			p.selected = len(p.taps) + i
			return
		}
	}
}

// followTaps brings the Taps tab bar in step with the cursor once an update is
// done: a trusted item has the Info tab alone. It runs before the motion settles,
// so the tab underline aims at the bar that is drawn.
func followTaps(next tea.Model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch nm := next.(type) {
	case *Model:
		nm.syncTapTabs()
	case Model:
		nm.syncTapTabs()
		return nm, cmd
	}
	return next, cmd
}

// syncTapTabs sets the tab bar to the Taps tabs, or to the Info tab alone when the
// cursor is on a trusted item. A tab index the shorter bar no longer has goes to Info.
func (m *Model) syncTapTabs() {
	if m.activePanel != PanelTaps {
		return
	}
	m.tabs = panelTabs[PanelTaps]
	if m.panels[PanelTaps].selectedTrusted() != nil {
		m.tabs = trustedItemTabs
	}
	if m.activeTab >= len(m.tabs) {
		m.activeTab = 0
	}
}

// renderTrustedInfo is the Info tab of a trusted item: its name, kind and tap,
// whether that tap is trusted, and what trust means for the item alone. The footer
// line names the key that untrusts it.
func renderTrustedInfo(p *panelData, e brew.TrustEntry, width int) string {
	tap := e.Tap
	if tap == "" {
		tap = "—"
	}
	fields := []presentation.Field{
		{Label: "Name", Value: e.Name},
		{Label: "Type", Value: string(e.Type)},
		{Label: "Tap", Value: tap},
		{Label: "Tap trusted", Value: p.tapTrustWord(e.Tap)},
	}
	note := style.NormalItem.Render("Trusted on its own: Homebrew may load it even if its tap is not trusted.")
	return presentation.Detail(fields, width) + "\n\n" + note + "\n\n" + footerLine([]keyHint{tUntrust}, nil, width)
}

// tapTrustWord is the trust of the tap a trusted item belongs to: yes, no, or
// unknown when the taps were read without tap-info, the item has no tap, or the
// tap is not on the list.
func (p *panelData) tapTrustWord(tap string) string {
	if p.trustUnknown || tap == "" {
		return "unknown"
	}
	for _, t := range p.taps {
		if t.Name == tap {
			if t.Trusted || t.IsOfficial {
				return "yes"
			}
			return "no"
		}
	}
	return "unknown"
}
