package gui

import (
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

// footerHints is the key row the footer shows: the keys of the open modal,
// else the help screen's keys while help is open, else the active panel's
// keys followed by the global ones.
func (m Model) footerHints() []keyHint {
	if m.activeModal != nil {
		return modalHints(m.activeModal)
	}
	if m.showHelp {
		return []keyHint{keyHelpOff}
	}
	hints := append([]keyHint{}, m.panelKeyHints()...)
	hints = append(hints, m.globalFooterHints()...)
	if m.opState != nil && m.opState.Done() {
		// A finished operation covers the main area, and Esc clears it.
		return append([]keyHint{keyClose}, hints...)
	}
	return hints
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
	return footerLine(m.footerHints(), m.width)
}

// footerLine lays the hints out on one row, w cells wide. Each hint is its key
// in the accent and its description in the subtle text, and hints are
// separated by two spaces. When the row cannot hold them all, it keeps the
// hints that fit and ends with "? more", which is always shown: the ? help
// screen lists the rest.
func footerLine(hints []keyHint, w int) string {
	const gap = "  "
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, style.AccentText.Render(h.key)+" "+style.SubtleText.Render(h.desc))
	}

	line := " " + strings.Join(parts, gap)
	if lipgloss.Width(line) > w {
		more := style.AccentText.Render("?") + " " + style.SubtleText.Render("more")
		line = " " + more
		for i := range parts {
			candidate := " " + strings.Join(parts[:i+1], gap) + gap + more
			if lipgloss.Width(candidate) > w {
				break
			}
			line = candidate
		}
	}
	return fitFrame(line, w, bottomBarRows)
}
