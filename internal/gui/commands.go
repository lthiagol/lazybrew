package gui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thiago/lazybrew/internal/brew"
	"github.com/thiago/lazybrew/internal/gui/modal"
	"github.com/thiago/lazybrew/internal/gui/presentation"
)

func (m *Model) startSearch() (tea.Model, tea.Cmd) {
	inputModal := modal.NewInputModal("Search:")
	m.activeModal = inputModal
	return m, inputModal.Init()
}

func (m *Model) handleModalResult(result interface{}, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch r := result.(type) {
	case *modal.InputResult:
		if !r.Cancelled && r.Value != "" {
			return m, m.executeSearch(r.Value)
		}
	case *modal.ConfirmResult:
		if r.Confirmed {
			return m, cmd
		}
	case *modal.MenuResult:
		if !r.Cancelled {
			return m, m.executeMenuAction(r.SelectedIndex)
		}
	}
	return m, nil
}

func (m *Model) executeMenuAction(selectedIndex int) tea.Cmd {
	switch {
	case m.activePanel == PanelTaps && selectedIndex >= 0 && selectedIndex <= 1:
		return m.executeTrustAction(selectedIndex)
	default:
		return m.executeBrewfileAction(selectedIndex)
	}
}

func (m *Model) executeTrustAction(index int) tea.Cmd {
	panel := m.panels[PanelTaps]
	name := extractPackageName(panel.selectedItem())
	if name == "" {
		return nil
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	m.activeModal = modal.NewProgressModal("Trust: "+name, cancel)

	return func() tea.Msg {
		var err error
		if index == 0 {
			err = m.client.Trust.TrustTap(cancelCtx, name)
		} else {
			err = m.client.Trust.UntrustTap(cancelCtx, name)
		}
		return MutationResultMsg{Name: name, Type: mutInstall, Err: err}
	}
}

func (m *Model) executeBrewfileAction(index int) tea.Cmd {
	actions := []string{"dump", "install", "cleanup", "check", "list"}
	if index < 0 || index >= len(actions) {
		return nil
	}
	action := actions[index]
	cancelCtx, cancel := context.WithCancel(context.Background())
	title := "Brewfile " + action
	m.activeModal = modal.NewProgressModal(title, cancel)

	return func() tea.Msg {
		_, err := m.client.Diagnostics.Doctor(cancelCtx)
		_ = err
		return MutationResultMsg{Name: "brewfile " + action, Type: mutInstall, Err: nil}
	}
}

func (m *Model) executeSearch(query string) tea.Cmd {
	m.panels[PanelSearch].loading = true
	m.panels[PanelSearch].err = nil

	return func() tea.Msg {
		results, err := m.client.Search.Search(context.Background(), query)
		if err != nil {
			return SearchDoneMsg{Err: err}
		}

		items := make([]string, len(results))
		for i, r := range results {
			typ := ""
			if r.IsFormula {
				typ = "formula"
			} else {
				typ = "cask"
			}
			installed := ""
			if r.Installed {
				installed = "installed"
			}
			items[i] = r.Name + "  " + typ + "  " + installed + "  " + r.Description
		}
		return SearchDoneMsg{Results: items}
	}
}

func (m Model) doMutation(mutType mutationType, label string) (tea.Model, tea.Cmd) {
	panel := m.panels[m.activePanel]
	if panel.selected >= len(panel.items) {
		return m, nil
	}

	name := extractPackageName(panel.items[panel.selected])
	if name == "" {
		return m, nil
	}

	title := label + " " + name
	if mutType == mutUpgradeAll {
		title = "Upgrading all packages"
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	m.activeModal = modal.NewProgressModal(title, cancel)

	if m.program == nil {
		return m, nil
	}

	go func() {
		var ch <-chan string
		var errCh <-chan error

		switch mutType {
		case mutInstall:
			ch, errCh = m.client.FormulaeWrite.Install(cancelCtx, name)
		case mutUninstall:
			if m.activePanel == PanelCasks {
				ch, errCh = m.client.CasksWrite.Uninstall(cancelCtx, name)
			} else {
				ch, errCh = m.client.FormulaeWrite.Uninstall(cancelCtx, name)
			}
		case mutReinstall:
			if m.activePanel == PanelCasks {
				ch, errCh = m.client.CasksWrite.Zap(cancelCtx, name)
			} else {
				ch, errCh = m.client.FormulaeWrite.Reinstall(cancelCtx, name)
			}
		case mutUpgrade:
			if m.activePanel == PanelCasks {
				ch, errCh = m.client.CasksWrite.Upgrade(cancelCtx, name)
			} else {
				ch, errCh = m.client.FormulaeWrite.Upgrade(cancelCtx, name)
			}
		case mutUpgradeAll:
			ch, errCh = m.client.FormulaeWrite.Upgrade(cancelCtx, "")
		case mutZap:
			ch, errCh = m.client.CasksWrite.Zap(cancelCtx, name)
		}

		if ch != nil {
			for line := range ch {
				m.program.Send(ProgressLineMsg{Line: line})
			}
		}
		var err error
		if errCh != nil {
			err = <-errCh
		}
		m.program.Send(ProgressCompleteMsg{Err: err, Name: name})
	}()

	return m, nil
}

func (m *Model) startTapAdd() (tea.Model, tea.Cmd) {
	inputModal := modal.NewInputModal("Tap repository (user/repo):")
	m.activeModal = inputModal
	return m, inputModal.Init()
}

func (m *Model) startTrustMenu() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelTaps]
	if panel.selected >= len(panel.items) {
		return m, nil
	}
	tapName := extractPackageName(panel.items[panel.selected])
	menuItems := []string{
		"Trust entire tap: " + tapName,
		"Untrust tap: " + tapName,
	}
	menuModal := modal.NewMenuModal("Trust: "+tapName, menuItems)
	m.activeModal = menuModal
	return m, menuModal.Init()
}

func (m Model) serviceAction(action string) (tea.Model, tea.Cmd) {
	panel := m.panels[PanelServices]
	if panel.selected >= len(panel.items) {
		return m, nil
	}
	name := extractPackageName(panel.items[panel.selected])
	if name == "" {
		return m, nil
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	m.activeModal = modal.NewProgressModal(action+" "+name, cancel)

	return m, func() tea.Msg {
		var err error
		switch action {
		case "start":
			err = m.client.Services.Start(cancelCtx, name)
		case "stop":
			err = m.client.Services.Stop(cancelCtx, name)
		case "restart":
			err = m.client.Services.Restart(cancelCtx, name)
		}
		return MutationResultMsg{Name: name, Type: mutInstall, Err: err}
	}
}

func (m Model) togglePin(mutType mutationType) (tea.Model, tea.Cmd) {
	panel := m.panels[m.activePanel]
	if panel.selected >= len(panel.items) {
		return m, nil
	}
	name := extractPackageName(panel.items[panel.selected])
	if name == "" {
		return m, nil
	}

	return m, func() tea.Msg {
		var err error
		if m.activePanel == PanelCasks {
			err = m.client.CasksWrite.Unpin(context.Background(), name)
			if err != nil {
				err = m.client.CasksWrite.Pin(context.Background(), name)
			}
		} else {
			err = m.client.FormulaeWrite.Unpin(context.Background(), name)
			if err != nil {
				err = m.client.FormulaeWrite.Pin(context.Background(), name)
			}
		}
		return MutationResultMsg{Name: name, Type: mutInstall, Err: err}
	}
}

func (m Model) serviceCleanup() (tea.Model, tea.Cmd) {
	m.activeModal = modal.NewConfirmModal("Cleanup Services", "Remove stale service files?")
	return m, nil
}

func (m Model) brewfileMenu() (tea.Model, tea.Cmd) {
	menuItems := []string{
		"Export to Brewfile (brew bundle dump)",
		"Install from Brewfile (brew bundle install)",
		"Cleanup (brew bundle cleanup)",
		"Check (brew bundle check)",
		"List entries (brew bundle list)",
	}
	menuModal := modal.NewMenuModal("Brewfile", menuItems)
	m.activeModal = menuModal
	return m, menuModal.Init()
}

func (m Model) runVulns() (tea.Model, tea.Cmd) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	m.activeModal = modal.NewProgressModal("Vulnerability Check", cancel)

	return m, func() tea.Msg {
		warnings, err := m.client.Diagnostics.Doctor(cancelCtx)
		var result string
		if err != nil {
			return MutationResultMsg{Name: "vulns", Type: mutInstall, Err: err}
		}
		if len(warnings) == 0 {
			result = "No vulnerabilities found"
		} else {
			for _, w := range warnings {
				result += w.Title + "\n" + w.Details + "\n\n"
			}
		}
		key := tabKey(PanelStatus, 2)
		m.tabContent[key] = result
		return MutationResultMsg{Name: "vulns", Type: mutInstall, Err: nil}
	}
}

func (m Model) runMissing() (tea.Model, tea.Cmd) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	m.activeModal = modal.NewProgressModal("Missing Dependencies", cancel)

	return m, func() tea.Msg {
		missing, err := m.client.Diagnostics.Missing(cancelCtx)
		var result string
		if err != nil {
			return MutationResultMsg{Name: "missing", Type: mutInstall, Err: err}
		}
		if len(missing) == 0 {
			result = "All dependencies satisfied"
		} else {
			for _, m := range missing {
				result += m.Formula + ": " + m.Missing + "\n"
			}
		}
		key := tabKey(PanelStatus, 2)
		m.tabContent[key] = result
		return MutationResultMsg{Name: "missing", Type: mutInstall, Err: nil}
	}
}

func (m *Model) loadTabContent() tea.Cmd {
	needsFetch := map[PanelID]map[int]bool{
		PanelStatus:   {1: true},
		PanelFormulae: {1: true, 2: true, 4: true},
	}

	panelID := m.activePanel
	tabIdx := m.activeTab

	if needsFetch[panelID] != nil && needsFetch[panelID][tabIdx] {
		key := tabKey(panelID, tabIdx)
		if _, ok := m.tabContent[key]; ok {
			return nil
		}
		panel := m.panels[panelID]
		if panel.selected >= len(panel.items) {
			return nil
		}
		name := extractPackageName(panel.selectedItem())
		if name == "" {
			return nil
		}
		return fetchTabContentCmd(m.client, panelID, tabIdx, name)
	}
	return nil
}

func fetchTabContentCmd(client *brew.Client, panel PanelID, tab int, name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		switch panel {
		case PanelStatus:
			if tab == 1 {
				cfg, err := client.Diagnostics.Config(ctx)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, Err: err}
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab,
					Content: formatConfig(cfg)}
			}
		case PanelFormulae:
			switch tab {
			case 1:
				deps, err := client.Formulae.Deps(ctx, name)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, Err: err}
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, Content: deps}
			case 2:
				uses, err := client.Formulae.Uses(ctx, name)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, Err: err}
				}
				result := ""
				for _, u := range uses {
					result += u + "\n"
				}
				if result == "" {
					result = "No dependents"
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, Content: result}
			case 4:
				return TabContentMsg{PanelID: panel, TabIndex: tab,
					Content: "Files for " + name + " (use 'brew list " + name + "')"}
			}
		}
		return TabContentMsg{PanelID: panel, TabIndex: tab, Content: ""}
	}
}

func formatConfig(cfg *brew.BrewConfig) string {
	if cfg == nil {
		return "No config available"
	}
	return "Homebrew: " + cfg.HomebrewVersion + "\n" +
		"Prefix: " + cfg.Prefix + "\n" +
		"Cellar: " + cfg.Cellar + "\n" +
		"Repository: " + cfg.Repository + "\n" +
		"Core Tap: " + cfg.CoreTap + "\n" +
		"OS: " + cfg.OS + "\n"
}

func fetchPanelData(client *brew.Client, panel PanelID) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		switch panel {
		case PanelFormulae:
			formulae, err := client.Formulae.List(ctx)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			items := make([]string, len(formulae))
			for i, f := range formulae {
				items[i] = presentation.FormatFormula(f, 0)
			}
			return DataLoadedMsg{PanelID: panel, Items: items, RawData: formulae}

		case PanelCasks:
			casks, err := client.Casks.List(ctx)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			items := make([]string, len(casks))
			for i, c := range casks {
				items[i] = presentation.FormatCask(c, 0)
			}
			return DataLoadedMsg{PanelID: panel, Items: items, RawData: casks}

		case PanelOutdated:
			formulae, _ := client.Formulae.Outdated(ctx)
			casks, _ := client.Casks.Outdated(ctx)
			items := make([]string, 0, len(formulae)+len(casks))
			for _, f := range formulae {
				items = append(items, presentation.FormatOutdatedFormula(f))
			}
			for _, c := range casks {
				items = append(items, presentation.FormatOutdatedCask(c))
			}
			return DataLoadedMsg{PanelID: panel, Items: items}

		case PanelTaps:
			taps, err := client.Taps.List(ctx)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			items := make([]string, len(taps))
			for i, t := range taps {
				items[i] = presentation.FormatTap(t, 0)
			}
			return DataLoadedMsg{PanelID: panel, Items: items, RawData: taps}

		case PanelServices:
			services, err := client.Services.List(ctx)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			items := make([]string, len(services))
			for i, s := range services {
				items[i] = presentation.FormatService(s, 0)
			}
			return DataLoadedMsg{PanelID: panel, Items: items, RawData: services}

		default:
			return DataLoadedMsg{PanelID: panel, Items: []string{}}
		}
	}
}

func fetchStatusData(client *brew.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		formulae, _ := client.Formulae.List(ctx)
		casks, _ := client.Casks.List(ctx)
		outdatedFormulae, _ := client.Formulae.Outdated(ctx)
		outdatedCasks, _ := client.Casks.Outdated(ctx)
		taps, _ := client.Taps.List(ctx)
		services, _ := client.Services.List(ctx)
		cfg, _ := client.Diagnostics.Config(ctx)

		brewVersion := ""
		prefix := ""
		if cfg != nil {
			brewVersion = cfg.HomebrewVersion
			prefix = cfg.Prefix
		}

		officialTaps := 0
		thirdPartyTaps := 0
		for _, t := range taps {
			if t.IsOfficial {
				officialTaps++
			} else {
				thirdPartyTaps++
			}
		}

		servicesStarted := 0
		for _, s := range services {
			if s.Status == brew.ServiceStarted {
				servicesStarted++
			}
		}

		outdatedCount := len(outdatedFormulae) + len(outdatedCasks)
		items := presentation.FormatStatusDashboard(
			len(formulae), len(casks), outdatedCount,
			len(taps), officialTaps, thirdPartyTaps,
			len(services), servicesStarted,
			brewVersion, prefix,
		)

		doctorWarnings, _ := client.Diagnostics.Doctor(ctx)
		items = append(items, "")
		items = append(items, presentation.FormatDoctorStatus(doctorWarnings))

		return DataLoadedMsg{PanelID: PanelStatus, Items: items}
	}
}
