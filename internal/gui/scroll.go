package gui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// contentKey names the content the main area shows: the panel, the tab and the
// item the tab is about. A scroll position belongs to one such key, so moving
// the selection starts the next item at its top.
func (m Model) contentKey() string {
	return tabKey(m.activePanel, m.activeTab, tabItemName(m.panels[m.activePanel]))
}

// wrapContent wraps content to w cells and splits it into rows, the rows the
// main area can scroll through.
func wrapContent(content string, w int) []string {
	return strings.Split(lipgloss.NewStyle().Width(w).Render(content), "\n")
}

// scrolledContent is the main content cut to the window the wheel has scrolled
// it to: h rows of w cells. The window never runs past the last row.
func (m Model) scrolledContent(content string, w, h int) string {
	rows := wrapContent(content, w)
	off := min(m.contentScroll[m.contentKey()], max(0, len(rows)-h))
	return fitBox(strings.Join(rows[off:], "\n"), w, h)
}

// scrollContent moves the main content by delta rows. The offset stops at the
// first row and at the last row that fills the content area.
func (m Model) scrollContent(delta int) Model {
	l := m.screenLayout()
	innerW := max(1, l.mainW-2)
	rows := wrapContent(m.renderContent(innerW, l.contentH), innerW)
	last := max(0, len(rows)-l.contentH)
	key := m.contentKey()
	m.contentScroll[key] = min(max(0, m.contentScroll[key]+delta), last)
	return m
}
