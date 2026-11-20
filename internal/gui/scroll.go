package gui

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// wrapCacheMax bounds the wrapped tabs the cache keeps: past it the cache
// starts again, so a long session over many formulae does not keep them all.
const wrapCacheMax = 16

// wrapEntry is one wrapped tab, with what it was rendered from. It is valid
// only while the tab content, the installed list it reads and the theme are
// the same.
type wrapEntry struct {
	src   string
	gen   int
	theme *style.Theme
	rows  []string
}

// wrapCache holds the wrapped rows of the Files and Used By tabs (cr3 AC-04).
type wrapCache struct {
	entries map[string]wrapEntry
	// renders counts the tab renders the cache did not have: a test hook.
	renders int
}

func newWrapCache() *wrapCache {
	return &wrapCache{entries: make(map[string]wrapEntry)}
}

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

// cachedTab reports whether the current tab is a Files or Used By tab that has
// its content, and that content. Only these tabs keep their wrapped rows: the
// others draw from data that changes without their key changing.
//
// The tab's rows are kept only while renderContent draws them. An operation over
// the main area and a loading panel draw something else under the same key, so
// there the rows are not cached: the cache would keep the overlay, or the
// spinner, after the state that drew it had gone (cr3 F-r1-01).
func (m Model) cachedTab() (string, bool) {
	if m.activePanel != PanelFormulae || (m.activeTab != formulaFilesTab && m.activeTab != formulaUsesTab) {
		return "", false
	}
	p := m.panels[PanelFormulae]
	if m.opState != nil || p.loading || tabItemName(p) == "" {
		return "", false
	}
	src, ok := m.tabContent[m.contentKey()]
	return src, ok
}

// mainRows is the main content wrapped to w cells, as the rows the main area
// scrolls through. A Files or Used By tab keeps its rows, so scrolling it does
// not render or wrap it again.
func (m Model) mainRows(w, h int) []string {
	src, cached := m.cachedTab()
	if !cached {
		return wrapContent(m.renderContent(w, h), w)
	}
	key := m.contentKey() + "|" + strconv.Itoa(w)
	gen := m.panels[PanelFormulae].installedGen
	if e, ok := m.wrap.entries[key]; ok && e.src == src && e.gen == gen && e.theme == style.CurrentTheme {
		return e.rows
	}
	m.wrap.renders++
	rows := wrapContent(m.renderContent(w, h), w)
	if len(m.wrap.entries) >= wrapCacheMax {
		clear(m.wrap.entries)
	}
	m.wrap.entries[key] = wrapEntry{src: src, gen: gen, theme: style.CurrentTheme, rows: rows}
	return rows
}

// scrolledContent is the main content cut to the window the scroll offset
// selects: h rows of w cells. The window never runs past the last row.
func (m Model) scrolledContent(w, h int) string {
	rows := m.mainRows(w, h)
	off := min(m.contentScroll[m.contentKey()], max(0, len(rows)-h))
	return fitBox(strings.Join(rows[off:min(off+h, len(rows))], "\n"), w, h)
}

// contentHeadRows is how many rows of the main content stay on screen above
// the scrolled part: the Logs tab's header, and its note when it has one.
// contentHead draws them.
func (m Model) contentHeadRows() int {
	if !m.onLogsTab() {
		return 0
	}
	if m.logsNote() {
		return 2
	}
	return 1
}

// scrollRows is how many rows of a main content area contentH rows tall scroll.
func (m Model) scrollRows(contentH int) int {
	return max(0, contentH-m.contentHeadRows())
}

// contentLast is the largest scroll offset that still fills the main content
// area: the offset at which the content's last row is on the area's bottom row.
func (m Model) contentLast() int {
	l := m.screenLayout()
	cw := contentWidthIn(l)
	h := m.scrollRows(l.contentH)
	return max(0, len(m.mainRows(cw, h))-h)
}

// pageRows is how far a page key moves the main content: one scrolled area.
func (m Model) pageRows() int {
	return max(1, m.scrollRows(m.screenLayout().contentH))
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
