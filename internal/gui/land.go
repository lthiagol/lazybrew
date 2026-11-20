package gui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// landData applies a panel's fresh data to the model and returns the commands
// the landing starts. preloaded is set for a preload read.
func (m *Model) landData(msg DataLoadedMsg, preloaded bool) tea.Cmd {
	if int(msg.PanelID) < len(m.panels) {
		p := m.panels[msg.PanelID]
		// A reload keeps the cursor on the same service by name, not by
		// index, so a list that changed under it does not move the
		// selection (AC-03.3).
		selectedService := ""
		if s := p.selectedService(); msg.PanelID == PanelServices && s != nil {
			selectedService = s.Name
		}
		// A reload keeps the cursor on the same tap by name (v2-10 AC-04.1).
		selectedTap := ""
		if t := p.selectedTap(); msg.PanelID == PanelTaps && t != nil {
			selectedTap = t.Name
		}
		p.loading = false
		p.loaded = true
		p.err = msg.Err
		if msg.Err == nil {
			p.items = msg.Items
			p.formulae = msg.Formulae
			p.casks = msg.Casks
			p.taps = msg.Taps
			p.trustUnknown = msg.TrustUnknown
			p.services = msg.Services
			p.serviceInfo = msg.ServiceInfo
			if msg.PanelID == PanelStatus {
				m.status = msg.Summary
				if m.cfg.GUI.Preload {
					m.status.fillUnknown(m.landed)
					p.items = statusRows(m.status)
				}
			}
			selectNamed(p, p.services, func(s brew.Service) string { return s.Name }, selectedService)
			selectNamed(p, p.taps, func(t brew.Tap) string { return t.Name }, selectedTap)
		}
		p.indexInstalled()
		p.normalize()
	}
	// The Outdated panel was just rewritten: an index-based batch
	// selection would silently remap onto different rows and make a
	// later batch upgrade target the wrong packages (r4 finding
	// R4-2). A failed fetch keeps the old items, so the selection is
	// only dropped when the list actually changed.
	if msg.PanelID == PanelOutdated && msg.Err == nil {
		m.batch.selected = make(map[int]bool)
	}
	m.clearPanelTabContent(msg.PanelID)
	// A preload read is not one a refresh waits for, so it does not count.
	if m.refreshing > 0 && !preloaded {
		m.refreshing--
		if m.refreshing == 0 {
			return tea.Batch(
				m.setToast("Data refreshed", modal.ToastSuccess),
				m.loadTabContent(),
			)
		}
	}
	if msg.PanelID == m.activePanel && needsTabFetch(m.activePanel, m.activeTab) {
		return m.loadTabContent()
	}
	return nil
}
