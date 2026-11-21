package gui

import (
	"github.com/lthiagol/lazybrew/internal/config"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// handleKey answers a key press. The open modal takes it first; otherwise the
// search box, the help screen and the panel's keys do, and a key typed on the
// Search panel goes to the search box.
func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.activeModal != nil {
		updated, cmd := m.activeModal.Update(msg)
		next, ok := updated.(modal.Modal)
		if !ok {
			// Not a modal any more, so there is no result to act on.
			m.activeModal = nil
			return m, cmd
		}
		m.activeModal = next
		if next.Done() || next.Cancelled() {
			result := next.Result()
			m.activeModal = nil
			return m.handleModalResult(result, cmd)
		}
		return m, cmd
	}

	if m.searchInput.Focused() {
		switch msg.String() {
		case "esc":
			m.searchInput.Blur()
			return m, nil
		case "enter":
			q := strings.TrimSpace(m.searchInput.Value())
			if q != "" {
				m.searchInput.Blur()
				c := m.executeSearch(q)
				return m, c
			}
			return m, nil
		case "q", "ctrl+c":
			if m.searchInput.Value() == "" {
				return m, tea.Quit
			}
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		default:
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}
	}

	// While help is open its own keys scroll it or close it; the rest
	// pass through as they always have.
	if m.showHelp {
		if next, handled := m.helpKey(msg.String()); handled {
			return next, nil
		}
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.opState != nil && m.opState.Running() {
			m.tasks.CancelCurrent()
			return m, nil
		}
		if m.opState != nil && m.opState.Done() {
			m.opState = nil
			return m, nil
		}
		if m.showHelp {
			m.showHelp = false
		}
		return m, nil
	case "R":
		if m.cfg.Brew.UpdateOnStart == config.UpdateAlways && !m.isUpdating {
			m.isUpdating = true
			m.updateOutput = nil
			// The update task and the load-flagged refresh travel
			// together: R is an explicit user request, so the
			// outdated count must be loaded even when R also runs
			// `brew update`. The flag stays on the R path only —
			// setting it in the brew-update completion handler
			// would also load the count at startup (M9).
			return m, tea.Batch(
				m.startBrewUpdate(),
				func() tea.Msg { return RefreshMsg{LoadOutdated: true, ReloadDiagnostics: true} },
			)
		}
		// R is the only path allowed to shell out to `brew outdated`:
		// the dashboard's outdated count is loaded on demand.
		return m, func() tea.Msg { return RefreshMsg{LoadOutdated: true, ReloadDiagnostics: true} }

	case "C":
		m.logCollapsed = !m.logCollapsed
		// The toggle changes the main list height (logLayoutHeights)
		// without a resize, so re-clamp the active panel's
		// cursor/offset against the new effective viewport — the
		// smaller of the sidebar and main-list windows, which the
		// shared offset must satisfy (r3 finding R3-4).
		vc := m.panels[m.activePanel].visibleCount()
		mainH := m.screenLayout().contentH
		m.panels[m.activePanel].normalizeToHeight(min(vc, mainH))
		return m, nil

	case "tab":
		cmd := m.nextPanel()
		return m, cmd
	case "shift+tab":
		cmd := m.prevPanel()
		return m, cmd
	case "1", "2", "3", "4", "5", "6", "7":
		// The digits follow the panel order: 1 is the first panel.
		cmd := m.switchPanel(PanelID(msg.String()[0] - '1'))
		return m, cmd

	case "j", "down":
		m.panels[m.activePanel].down()
		if needsTabFetch(m.activePanel, m.activeTab) {
			c := m.loadTabContent()
			return m, c
		}
		return m, nil
	case "k", "up":
		m.panels[m.activePanel].up()
		if needsTabFetch(m.activePanel, m.activeTab) {
			c := m.loadTabContent()
			return m, c
		}
		return m, nil

	case "pgup":
		return m.scrollContent(-m.pageRows()), nil
	case "pgdown":
		return m.scrollContent(m.pageRows()), nil
	case "home":
		return m.scrollContentTo(0), nil
	case "end":
		return m.scrollContentTo(m.contentLast()), nil
	case "g", "G":
		// On Search these letters are typed into the search box, as before.
		// Elsewhere they are the top and bottom keys.
		if m.activePanel != PanelSearch {
			if msg.String() == "g" {
				return m.scrollContentTo(0), nil
			}
			return m.scrollContentTo(m.contentLast()), nil
		}

	case "[", "left":
		c := m.prevTab()
		return m, c
	case "]", "right":
		c := m.nextTab()
		return m, c

	case "/":
		m.switchPanel(PanelSearch)
		return m, m.searchInput.Focus()

	case "enter":
		if m.activePanel == PanelSearch && len(m.searchResults) > 0 {
			return m, m.fetchSelectedSearchInfo()
		}

	case "?":
		m.showHelp = !m.showHelp
		m.helpScroll = 0
		return m, nil

	case "i":
		if m.activePanel == PanelSearch || m.activePanel == PanelFormulae || m.activePanel == PanelCasks {
			return m.confirmMutation(mutInstall, "Install")
		}
	case "I":
		if m.activePanel == PanelSearch {
			return m.confirmMutation(mutInstall, "Install")
		}
	case "x":
		if m.activePanel == PanelFormulae || m.activePanel == PanelCasks {
			return m.confirmUninstall(mutUninstall)
		}
		if m.activePanel == PanelTaps {
			return m.confirmUntap()
		}
	case "X":
		if m.activePanel == PanelCasks {
			return m.confirmUninstall(mutZap)
		}
	case "r":
		if m.activePanel == PanelFormulae || m.activePanel == PanelCasks {
			return m.confirmMutation(mutReinstall, "Reinstall")
		}
		if m.activePanel == PanelTaps {
			return m.confirmRepair()
		}
		if m.activePanel == PanelServices {
			return m.serviceAction("restart")
		}
	case "F":
		if m.activePanel == PanelSearch || m.activePanel == PanelFormulae || m.activePanel == PanelCasks {
			return m.doMutation(mutFetch, "Fetch")
		}
	case "u":
		if m.activePanel == PanelOutdated {
			if len(m.batch.selected) > 0 {
				return m.confirmBatchUpgrade()
			}
			return m.confirmMutation(mutUpgrade, "Upgrade")
		}
		if m.activePanel == PanelFormulae || m.activePanel == PanelCasks {
			return m.confirmMutation(mutUpgrade, "Upgrade")
		}
	case "U":
		// Upgrade-all only from package panels: pressing U on
		// Services/Taps/Search silently queued a global formula
		// upgrade from an unrelated panel (r1 finding 25).
		if m.activePanel == PanelFormulae || m.activePanel == PanelCasks || m.activePanel == PanelOutdated {
			return m.confirmMutation(mutUpgradeAll, "Upgrade All")
		}
	case "space":
		if m.activePanel == PanelOutdated {
			p := m.panels[PanelOutdated]
			m.batch.toggle(p.selected)
		}
	case "a":
		if m.activePanel == PanelTaps {
			m, cmd := m.startTapAdd()
			return m, cmd
		}
		if m.activePanel == PanelOutdated {
			p := m.panels[PanelOutdated]
			if len(m.batch.selected) == len(p.items) {
				m.batch.selected = make(map[int]bool)
			} else {
				for i := range p.items {
					m.batch.selected[i] = true
				}
			}
			return m, nil
		}
	case "t":
		if m.activePanel == PanelTaps {
			m, cmd := m.toggleTrust()
			return m, cmd
		}
	case "T":
		if m.activePanel == PanelTaps {
			m, cmd := m.startTrustMenu()
			return m, cmd
		}
	case "s":
		if m.activePanel == PanelServices {
			return m.serviceAction("start")
		}
	case "S":
		if m.activePanel == PanelServices {
			return m.serviceAction("stop")
		}
	case "f":
		if m.activePanel == PanelServices {
			return m.serviceAction("run")
		}
	case "e":
		if m.activePanel == PanelServices && m.activeTab == servicesLogsTab {
			m.toggleLogStream()
			return m, nil
		}
	case "c":
		if m.activePanel == PanelServices {
			return m.serviceCleanup()
		}
		if m.activePanel == PanelStatus {
			return m.brewCleanup()
		}
	case "d":
		if m.activePanel == PanelStatus {
			return m.runDoctor()
		}
	case "L":
		if m.activePanel == PanelFormulae {
			return m.toggleLeaves()
		}
	case "A":
		if m.activePanel == PanelStatus {
			return m.runAutoremove()
		}
	case "B":
		if m.activePanel == PanelStatus {
			return m.brewfileMenu()
		}
	case "v":
		if m.activePanel == PanelStatus {
			return m.runVulns()
		}
	case "m":
		if m.activePanel == PanelStatus {
			return m.runMissing()
		}
	case "p":
		// Pin is formula-only: `brew pin` has no cask counterpart, and
		// the old cask path read the pinned flag from a misindexed
		// cask (r1 finding 13).
		if m.activePanel == PanelFormulae {
			return m.togglePin()
		}
	}

	if m.activePanel == PanelSearch && !m.searchInput.Focused() {
		if msg.Text != "" && msg.Code != tea.KeySpace {
			m.searchInput.Focus()
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}
	}

	if m.activeModal != nil {
		updated, cmd := m.activeModal.Update(msg)
		if modal, ok := updated.(modal.Modal); ok {
			m.activeModal = modal
		}
		return m, cmd
	}

	return m, nil
}
