package gui

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// globalHints are the keys every panel answers to, listed after the panel's
// own in the footer.
var globalHints = []keyHint{
	keyNext, keyPrev, keyJump, keySearch, keyHelp, keyRefresh, keyLog, keyQuit,
}

// footerRow is the key row the footer shows, in the two groups it draws: the
// panel's own actions, then the global keys. A surface with one list of keys (a
// modal, the help screen, an Info tab) leaves global empty.
type footerRow struct {
	panel  []keyHint
	global []keyHint
}

// all is the row's keys in the order the footer lists them.
func (r footerRow) all() []keyHint {
	return append(append([]keyHint{}, r.panel...), r.global...)
}

// footerHints is the key row the footer shows: the keys of the open modal,
// else the help screen's keys while help is open, else the active panel's
// keys, then the global ones.
func (m Model) footerHints() footerRow {
	if m.activeModal != nil {
		return footerRow{panel: modalHints(m.activeModal)}
	}
	if m.showHelp {
		return footerRow{panel: []keyHint{keyHelpOff}}
	}
	row := footerRow{panel: m.panelKeyHints(), global: m.globalFooterHints()}
	if m.opState != nil && m.opState.Done() {
		// A finished operation covers the main area, and Esc clears it.
		row.panel = append([]keyHint{keyClose}, row.panel...)
	}
	return row
}

// globalFooterHints are globalHints for the active panel. On the panels the `/`
// filter applies to, the panel's `/ filter` takes the place of `/ search`.
func (m Model) globalFooterHints() []keyHint {
	if !m.panels[m.activePanel].canFilter() {
		return globalHints
	}
	var hints []keyHint
	for _, h := range globalHints {
		if h != keySearch {
			hints = append(hints, h)
		}
	}
	return hints
}

// panelKeyHints are the active panel's keys. The Services keys follow the
// selected service's state.
func (m Model) panelKeyHints() []keyHint {
	if m.activePanel == PanelServices {
		hints := servicesHints(m.panels[PanelServices])
		if m.activeTab == servicesLogsTab {
			hints = append(hints, svLogStream)
		}
		return hints
	}
	if m.activePanel == PanelTaps {
		return tapsHints(m.panels[PanelTaps])
	}
	return panelHints(m.activePanel)
}

// modalHints lists the keys a modal answers to, as the modal lists them in
// its Hints. A modal that lists none leaves the footer its close key.
func modalHints(mo modal.Modal) []keyHint {
	h, ok := mo.(modal.Hinter)
	if !ok {
		return []keyHint{keyClose}
	}
	var hints []keyHint
	for _, x := range h.Hints() {
		hints = append(hints, keyHint{key: x.Key, desc: x.Desc})
	}
	return hints
}

// renderBottomBar draws the footer: one row of key hints across the bottom.
func (m Model) renderBottomBar() string {
	row := m.footerHints()
	return footerLine(row.panel, row.global, m.width)
}

// footerDropOrder is the global keys in the order they leave a row too narrow
// for all of them: the first goes first. The arrow hint and ? help never go.
var footerDropOrder = []string{"C", "R", "1-7", "Tab", "/", "q"}

// footerMinPanel is the fewest panel actions a row keeps while a global key from
// footerDropOrder is still shown.
const footerMinPanel = 3

// footerMore is the hint that ends a row which dropped keys. It stands in for
// ? help.
var footerMore = keyHint{key: "?", desc: "more"}

// footerLine lays the panel's keys and the global keys out on one row, w cells
// wide. The panel's keys come first, then a wider gap and a dim rule, then the
// global keys. Within a group the hints are separated by a dim dot. When the row
// cannot hold them all, fitFooter drops the keys and renderFooter ends the row
// with "? more", which the ? help screen explains.
func footerLine(panel, global []keyHint, w int) string {
	row, more := fitFooter(footerRow{panel: panel, global: global}, w)
	return fitFrame(renderFooter(row, more), w, bottomBarRows)
}

// fitFooter drops keys from the row until it is w cells or fewer. The panel's
// actions go first, from the right end, but at least footerMinPanel stay while a
// droppable global key remains. Then the droppable global keys go in
// footerDropOrder. It reports whether any key went.
func fitFooter(row footerRow, w int) (footerRow, bool) {
	dropped := false
	for lipgloss.Width(renderFooter(row, dropped)) > w {
		i := nextDroppable(row.global)
		switch {
		case i < 0 && len(row.panel) == 0:
			return row, dropped
		case i >= 0 && len(row.panel) <= footerMinPanel:
			row.global = slices.Delete(slices.Clone(row.global), i, i+1)
		default:
			row.panel = row.panel[:len(row.panel)-1]
		}
		dropped = true
	}
	return row, dropped
}

// nextDroppable is the index in global of the key that goes next, or -1 when no
// droppable key is left.
func nextDroppable(global []keyHint) int {
	for _, key := range footerDropOrder {
		if i := slices.IndexFunc(global, func(h keyHint) bool { return h.key == key }); i >= 0 {
			return i
		}
	}
	return -1
}

// renderFooter draws the row. When more is set, ? help is replaced by ? more at
// the end of the global group, or of the panel group when there is no global one.
func renderFooter(row footerRow, more bool) string {
	panel, global := row.panel, row.global
	if more {
		if len(global) == 0 {
			panel = append(slices.Clone(panel), footerMore)
		} else {
			global = append(slices.DeleteFunc(slices.Clone(global), func(h keyHint) bool { return h == keyHelp }), footerMore)
		}
	}
	var groups []string
	if len(panel) > 0 {
		groups = append(groups, footerGroup(panel))
	}
	if len(global) > 0 {
		groups = append(groups, footerGroup(global))
	}
	return " " + strings.Join(groups, "  "+style.MutedText.Render("│")+"  ")
}

// footerGroup draws one group's hints, each key in the accent and its word in the
// subtle text, separated by a dim dot.
func footerGroup(hints []keyHint) string {
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, style.AccentText.Render(h.key)+" "+style.SubtleText.Render(h.desc))
	}
	return strings.Join(parts, " "+style.MutedText.Render("·")+" ")
}
