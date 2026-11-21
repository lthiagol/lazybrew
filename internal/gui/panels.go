package gui

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// panelCacheGroups names the brew cache group each panel reads. The groups are
// brew's (see brew.Cache.InvalidateFor), so the keys are kept in one place.
//
// The status group does not include the outdated entries: the dashboard only
// reads them, and invalidating them here would throw away a valid 30-minute
// result. The Outdated panel and the mutation-level invalidations own them.
var panelCacheGroups = map[PanelID]string{
	PanelFormulae: "formulae",
	PanelCasks:    "casks",
	PanelOutdated: "outdated",
	PanelTaps:     "taps",
	PanelServices: "services",
	PanelStatus:   "status",
}

// invalidatePanelCache drops the brew cache group a panel reads, so the next
// fetch re-shells even when the per-class TTL has not expired (AC-04 R).
func (m *Model) invalidatePanelCache(id PanelID) {
	m.invalidateCacheGroup(panelCacheGroups[id])
}

// invalidateCacheGroup drops the brew cache keys of one group.
func (m *Model) invalidateCacheGroup(group string) {
	if m.client == nil || m.client.Cache == nil {
		return
	}
	m.client.Cache.InvalidateFor(group)
}

// lazyLoadPanel returns fetchPanelData(id) iff the panel is empty, not
// already loading. It is the single source of truth for M9/M12's lazy
// policy: callers that activate any non-Status panel (switchPanel,
// nextPanel, prevPanel) delegate here. Clearing p.err on retry ensures
// a stale error from a previous failed fetch doesn't block the new
// attempt. A panel the preload has queued is not fetched a second time:
// it moves to the front of the preload queue instead.
func (m *Model) lazyLoadPanel(id PanelID) tea.Cmd {
	if id == PanelStatus || id == PanelSearch {
		return nil
	}
	p := m.panels[id]
	if p.items == nil && !p.loading {
		p.loading = true
		p.err = nil
		return fetchPanelData(m.client, id)
	}
	if m.promotePreload(id) {
		return tea.Batch(m.nextPreloadReads()...)
	}
	return nil
}

// nextPanel / prevPanel cycle through panels. When cycling lands on a
// not-yet-loaded panel, the lazy path triggers a fetch — same policy
// as switchPanel. They delegate to switchPanel so every panel switch
// re-derives visibleRows: the sidebar layout depends on the active
// panel, and a stale viewport paginated the wrong rows in cramped
// terminals (r3 finding R3-3).
func (m *Model) nextPanel() tea.Cmd {
	return m.switchPanel(PanelID((int(m.activePanel) + 1) % len(m.panels)))
}

func (m *Model) prevPanel() tea.Cmd {
	return m.switchPanel(PanelID((int(m.activePanel) - 1 + len(m.panels)) % len(m.panels)))
}

// syncVisibleRows derives each panel's sidebar viewport height from the
// current window geometry. It runs in the Update loop — on WindowSizeMsg
// and on every panel switch, since the sidebar layout depends on which
// panel is active — so rendering never mutates model state (r1 finding
// 29, r2 finding NEW-8).
func (m *Model) syncVisibleRows() {
	heights := m.computeContentHeights()
	for i := range m.panels {
		m.panels[i].visibleRows = heights[i]
	}
}

// switchPanel activates the panel and returns any lazy-fetch command.
// Returns nil for panels that don't lazy-load (Status is preloaded on
// Init; Search is user-input-driven). All other panels lazy-load on
// first visit so Init doesn't stampede `brew …` and tiered refresh
// (M12) can re-fetch them later.
func (m *Model) switchPanel(id PanelID) tea.Cmd {
	if int(id) >= len(m.panels) {
		return nil
	}
	// Moving to another panel clears a finished operation, whose view covers the
	// main area. A running one stays. A key or click on the panel already shown
	// is not a move, so it clears nothing.
	if id != m.activePanel && m.opState != nil && m.opState.Done() {
		m.opState = nil
	}
	m.activePanel = id
	// The sidebar layout depends on the active panel: re-derive the
	// per-panel viewport heights (r2 finding NEW-8).
	m.syncVisibleRows()
	m.activeTab = 0
	m.tabs = panelTabs[m.activePanel]
	if id == PanelSearch {
		if len(m.searchResults) == 0 {
			m.searchInput.Focus()
		}
	} else {
		m.searchInput.Blur()
	}
	return m.lazyLoadPanel(id)
}

func (m *Model) nextTab() tea.Cmd {
	if len(m.tabs) > 0 {
		m.activeTab = (m.activeTab + 1) % len(m.tabs)
	}
	return m.loadTabContent()
}

func (m *Model) prevTab() tea.Cmd {
	if len(m.tabs) > 0 {
		m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
	}
	return m.loadTabContent()
}

func tabKey(panel PanelID, tab int, itemName string) string {
	return strconv.Itoa(int(panel)) + ":" + strconv.Itoa(tab) + ":" + itemName
}

// needsTabFetch is whether a tab's content comes from a brew read, not from the
// data the panel already holds. It runs on every j, k and data landing, so it
// builds nothing.
func needsTabFetch(panelID PanelID, tabIdx int) bool {
	switch panelID {
	case PanelStatus:
		return tabIdx == statusConfigTab || tabIdx == statusDoctorTab
	case PanelFormulae:
		return tabIdx == formulaUsesTab || tabIdx == formulaFilesTab
	}
	return false
}

func (m *Model) clearTabContent() {
	m.tabContent = make(map[string]string)
}

func (m *Model) clearPanelTabContent(pid PanelID) {
	prefix := strconv.Itoa(int(pid)) + ":"
	for k := range m.tabContent {
		if strings.HasPrefix(k, prefix) {
			delete(m.tabContent, k)
		}
	}
}
