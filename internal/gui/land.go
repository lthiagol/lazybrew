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
		// A reload keeps the cursor on the same tap, or trusted item, by name
		// (v2-10 AC-04.1, lb30 AC-02).
		selectedRow := ""
		onTrusted := false
		if msg.PanelID == PanelTaps {
			selectedRow = p.selectedName()
			onTrusted = p.selectedTrusted() != nil
		}
		p.loading = false
		p.loaded = true
		p.err = msg.Err
		if msg.Err == nil {
			// A filter on the list is applied again to the fresh rows, with the
			// cursor kept on the same name (lb26).
			if p.narrowed() {
				name := p.selectedName()
				p.unfilteredItems = msg.Items
				p.unfilteredFormulae = msg.Formulae
				p.unfilteredCasks = msg.Casks
				p.narrow(name)
			} else {
				p.items = msg.Items
				p.formulae = msg.Formulae
				p.casks = msg.Casks
			}
			p.taps = msg.Taps
			p.trustUnknown = msg.TrustUnknown
			p.trusted = msg.Trusted
			if msg.PanelID == PanelTaps {
				// The Taps rows, and their names, come from the taps and trusted
				// entries, not from msg.Items.
				p.setTapRows()
			}
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
			if msg.PanelID == PanelTaps {
				selectTapRow(p, selectedRow, onTrusted)
			}
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
	// A failed trust read is toasted once: a refresh that fails the same way
	// stays quiet, and the trusted rows are simply absent (lb30 AC-02).
	var trustToast tea.Cmd
	if msg.PanelID == PanelTaps && msg.Err == nil {
		p := m.panels[PanelTaps]
		if msg.TrustErr != nil && !p.trustFailed {
			trustToast = m.setToast("Could not read the trust list: "+msg.TrustErr.Error(), modal.ToastError)
		}
		p.trustFailed = msg.TrustErr != nil
	}
	m.clearPanelTabContent(msg.PanelID)
	// A preload read is not one a refresh waits for, so it does not count.
	if m.refreshing > 0 && !preloaded {
		m.refreshing--
		if m.refreshing == 0 {
			return tea.Batch(
				trustToast,
				m.setToast("Data refreshed", modal.ToastSuccess),
				m.loadTabContent(),
			)
		}
	}
	if msg.PanelID == m.activePanel && needsTabFetch(m.activePanel, m.activeTab) {
		return tea.Batch(trustToast, m.loadTabContent())
	}
	return trustToast
}
