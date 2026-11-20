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

// scrolledContent is the main content cut to the window the scroll offset
// selects: h rows of w cells. The window never runs past the last row.
func (m Model) scrolledContent(content string, w, h int) string {
	rows := wrapContent(content, w)
	off := min(m.contentScroll[m.contentKey()], max(0, len(rows)-h))
	return fitBox(strings.Join(rows[off:], "\n"), w, h)
}

// contentLast is the largest scroll offset that still fills the main content
// area: the offset at which the content's last row is on the area's bottom row.
func (m Model) contentLast() int {
	l := m.screenLayout()
	cw := m.mainContentWidth()
	rows := wrapContent(m.renderContent(cw, l.contentH), cw)
	return max(0, len(rows)-l.contentH)
}

// pageRows is how far a page key moves the main content: one content area.
func (m Model) pageRows() int {
	return max(1, m.screenLayout().contentH)
}

// scrollContentTo sets the main content's scroll offset to row, held between the
// first row and the last one that fills the content area.
func (m Model) scrollContentTo(row int) Model {
	key := m.contentKey()
	m.contentScroll[key] = min(max(0, row), m.contentLast())
	return m
}

// scrollContent moves the main content by delta rows, within its own length.
func (m Model) scrollContent(delta int) Model {
	return m.scrollContentTo(m.contentScroll[m.contentKey()] + delta)
}
