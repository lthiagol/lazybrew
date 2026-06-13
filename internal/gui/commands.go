package gui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thiago/lazybrew/internal/brew"
	"github.com/thiago/lazybrew/internal/gui/modal"
	"github.com/thiago/lazybrew/internal/gui/presentation"
	"github.com/thiago/lazybrew/internal/gui/task"
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
			switch m.pendingAction {
			case "uninstall":
				label := "Uninstall"
				if m.pendingMutType == mutZap {
					label = "Zap"
				}
				return m.doMutation(m.pendingMutType, label)
			case "untap":
				return m.executeUntap()
			case "repair":
				return m.executeRepair()
			case "cleanup":
				m.pendingAction = ""
				return m.executeCleanup()
			case "autoremove":
				m.pendingAction = ""
				return m.executeAutoremove()
			}
			if m.confirmCallback != nil {
				cb := m.confirmCallback
				m.confirmCallback = nil
				return m, cb
			}
		}
		m.confirmCallback = nil
		m.pendingAction = ""
	case *modal.MenuResult:
		if !r.Cancelled {
			switch m.pendingAction {
			case "trust-menu":
				m.pendingAction = ""
				return m, m.executeTrustMenuAction(r.SelectedIndex)
			case "trust-formula":
				m.pendingAction = ""
				return m, m.executeTrustFormulaAction(r.SelectedIndex)
			case "trust-cask":
				m.pendingAction = ""
				return m, m.executeTrustCaskAction(r.SelectedIndex)
			}
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

func (m *Model) executeTrustMenuAction(index int) tea.Cmd {
	panel := m.panels[PanelTaps]
	tapName := extractPackageName(panel.selectedItem())
	tap := panel.selectedTap()

	switch index {
	case 0:
		return m.executeTrustAction(0)
	case 1:
		return m.executeTrustAction(1)
	case 2:
		if tap != nil && len(tap.FormulaNames) > 0 {
			menuModal := modal.NewMenuModal("Trust formula in "+tapName, tap.FormulaNames)
			m.activeModal = menuModal
			m.pendingAction = "trust-formula"
			return menuModal.Init()
		}
	case 3:
		if tap != nil && len(tap.CaskNames) > 0 {
			menuModal := modal.NewMenuModal("Trust cask in "+tapName, tap.CaskNames)
			m.activeModal = menuModal
			m.pendingAction = "trust-cask"
			return menuModal.Init()
		}
	}
	return nil
}

func (m *Model) executeTrustAction(index int) tea.Cmd {
	panel := m.panels[PanelTaps]
	name := extractPackageName(panel.selectedItem())
	if name == "" {
		return nil
	}

	t := &task.Task{
		ID:    name,
		Title: "Trust: " + name,
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			var err error
			if index == 0 {
				err = m.client.TrustWrite.TrustTap(ctx, name)
			} else {
				err = m.client.TrustWrite.UntrustTap(ctx, name)
			}
			errCh := make(chan error, 1)
			errCh <- err
			return closedCh(), errCh, nil
		},
	}

	started, err := m.tasks.Enqueue(t)
	if err != nil {
		m.toast = modal.NewToast("Queue full: "+err.Error(), modal.ToastWarning)
		return nil
	}
	if !started {
		m.toast = modal.NewToast("A brew operation is already running", modal.ToastWarning)
		return nil
	}
	return m.tasks.RunNext()
}

func (m *Model) executeTrustFormulaAction(index int) tea.Cmd {
	panel := m.panels[PanelTaps]
	tap := panel.selectedTap()
	if tap == nil || index >= len(tap.FormulaNames) {
		return nil
	}
	return m.trustItemCmd(tap.FormulaNames[index], "formula")
}

func (m *Model) executeTrustCaskAction(index int) tea.Cmd {
	panel := m.panels[PanelTaps]
	tap := panel.selectedTap()
	if tap == nil || index >= len(tap.CaskNames) {
		return nil
	}
	return m.trustItemCmd(tap.CaskNames[index], "cask")
}

func (m *Model) trustItemCmd(fullName, itemType string) tea.Cmd {
	t := &task.Task{
		ID:    fullName,
		Title: "Trust: " + fullName,
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			var err error
			if itemType == "formula" {
				err = m.client.TrustWrite.TrustFormula(ctx, fullName)
			} else {
				err = m.client.TrustWrite.TrustCask(ctx, fullName)
			}
			errCh := make(chan error, 1)
			errCh <- err
			return closedCh(), errCh, nil
		},
	}

	started, err := m.tasks.Enqueue(t)
	if err != nil {
		m.toast = modal.NewToast("Queue full: "+err.Error(), modal.ToastWarning)
		return nil
	}
	if !started {
		m.toast = modal.NewToast("A brew operation is already running", modal.ToastWarning)
		return nil
	}
	return m.tasks.RunNext()
}

func (m Model) executeUntap() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelTaps]
	name := extractPackageName(panel.selectedItem())

	t := &task.Task{
		ID:    name,
		Title: "Untap " + name,
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			errCh := make(chan error, 1)
			errCh <- m.client.TapsWrite.Untap(ctx, name)
			return closedCh(), errCh, nil
		},
	}

	started, err := m.tasks.Enqueue(t)
	if err != nil {
		m.toast = modal.NewToast("Queue full: "+err.Error(), modal.ToastWarning)
		return m, nil
	}
	if !started {
		m.toast = modal.NewToast("A brew operation is already running", modal.ToastWarning)
		return m, nil
	}
	return m, m.tasks.RunNext()
}

func (m Model) executeRepair() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelTaps]
	name := extractPackageName(panel.selectedItem())

	t := &task.Task{
		ID:    name,
		Title: "Repair " + name,
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			errCh := make(chan error, 1)
			errCh <- m.client.TapsWrite.Repair(ctx, name)
			return closedCh(), errCh, nil
		},
	}

	started, err := m.tasks.Enqueue(t)
	if err != nil {
		m.toast = modal.NewToast("Queue full: "+err.Error(), modal.ToastWarning)
		return m, nil
	}
	if !started {
		m.toast = modal.NewToast("A brew operation is already running", modal.ToastWarning)
		return m, nil
	}
	return m, m.tasks.RunNext()
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
		ch, errCh := m.client.Runner.ExecuteStream(cancelCtx, "bundle", action)
		if ch != nil {
			for line := range ch {
				m.program.Send(ProgressLineMsg{Line: line})
			}
		}
		var err error
		if errCh != nil {
			err = <-errCh
		}
		return ProgressCompleteMsg{Err: err, Name: "brewfile " + action}
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
	pkgName := name
	if mutType == mutUpgradeAll {
		title = "Upgrading all packages"
		pkgName = ""
	}

	t := &task.Task{
		ID:    name,
		Title: title,
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			var ch <-chan string
			var errCh <-chan error

			switch mutType {
			case mutInstall:
				ch, errCh = m.client.FormulaeWrite.Install(ctx, pkgName)
			case mutUninstall:
				if m.activePanel == PanelCasks {
					ch, errCh = m.client.CasksWrite.Uninstall(ctx, pkgName)
				} else {
					ch, errCh = m.client.FormulaeWrite.Uninstall(ctx, pkgName)
				}
			case mutReinstall:
				if m.activePanel == PanelCasks {
					ch, errCh = m.client.CasksWrite.Reinstall(ctx, pkgName)
				} else {
					ch, errCh = m.client.FormulaeWrite.Reinstall(ctx, pkgName)
				}
			case mutUpgrade:
				if m.activePanel == PanelCasks {
					ch, errCh = m.client.CasksWrite.Upgrade(ctx, pkgName)
				} else {
					ch, errCh = m.client.FormulaeWrite.Upgrade(ctx, pkgName)
				}
			case mutUpgradeAll:
				ch, errCh = m.client.FormulaeWrite.Upgrade(ctx, "")
			case mutZap:
				ch, errCh = m.client.CasksWrite.Zap(ctx, pkgName)
			case mutFetch:
				ch, errCh = m.client.Runner.ExecuteStream(ctx, "fetch", pkgName)
			}

			if ch == nil {
				ch = closedCh()
			}
			return ch, errCh, nil
		},
	}

	started, err := m.tasks.Enqueue(t)
	if err != nil {
		m.toast = modal.NewToast("Queue full: "+err.Error(), modal.ToastWarning)
		return m, nil
	}
	if !started {
		m.toast = modal.NewToast("A brew operation is already running", modal.ToastWarning)
		return m, nil
	}

	return m, m.tasks.RunNext()
}

func closedCh() <-chan string {
	ch := make(chan string)
	close(ch)
	return ch
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
	tap := panel.selectedTap()
	formulaCount := 0
	caskCount := 0
	if tap != nil {
		formulaCount = len(tap.FormulaNames)
		caskCount = len(tap.CaskNames)
	}

	menuItems := []string{
		"Trust entire tap: " + tapName,
		"Untrust tap: " + tapName,
	}
	if formulaCount > 0 {
		menuItems = append(menuItems, "Trust specific formula...")
	}
	if caskCount > 0 {
		menuItems = append(menuItems, "Trust specific cask...")
	}

	m.pendingAction = "trust-menu"
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

	t := &task.Task{
		ID:    name,
		Title: action + " " + name,
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			errCh := make(chan error, 1)
			switch action {
			case "start":
				errCh <- m.client.ServicesWrite.Start(ctx, name)
			case "stop":
				errCh <- m.client.ServicesWrite.Stop(ctx, name)
			case "restart":
				errCh <- m.client.ServicesWrite.Restart(ctx, name)
			case "run":
				errCh <- m.client.ServicesWrite.Run(ctx, name)
			}
			return closedCh(), errCh, nil
		},
	}

	started, err := m.tasks.Enqueue(t)
	if err != nil {
		m.toast = modal.NewToast("Queue full: "+err.Error(), modal.ToastWarning)
		return m, nil
	}
	if !started {
		m.toast = modal.NewToast("A brew operation is already running", modal.ToastWarning)
		return m, nil
	}
	return m, m.tasks.RunNext()
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

	t := &task.Task{
		ID:    name,
		Title: "Toggle pin " + name,
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			var err error
			if m.activePanel == PanelCasks {
				err = m.client.CasksWrite.Unpin(ctx, name)
				if err != nil {
					err = m.client.CasksWrite.Pin(ctx, name)
				}
			} else {
				err = m.client.FormulaeWrite.Unpin(ctx, name)
				if err != nil {
					err = m.client.FormulaeWrite.Pin(ctx, name)
				}
			}
			errCh := make(chan error, 1)
			errCh <- err
			return closedCh(), errCh, nil
		},
	}

	started, err := m.tasks.Enqueue(t)
	if err != nil {
		m.toast = modal.NewToast("Queue full: "+err.Error(), modal.ToastWarning)
		return m, nil
	}
	if !started {
		m.toast = modal.NewToast("A brew operation is already running", modal.ToastWarning)
		return m, nil
	}
	return m, m.tasks.RunNext()
}

func (m Model) serviceCleanup() (tea.Model, tea.Cmd) {
	m.activeModal = modal.NewConfirmModal("Cleanup Services", "Remove stale service files?")
	m.confirmCallback = func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := m.client.ServicesWrite.Cleanup(ctx)
		return MutationResultMsg{Name: "services cleanup", Type: mutInstall, Err: err}
	}
	return m, m.activeModal.Init()
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
		output, err := m.client.Diagnostics.Vulns(cancelCtx)
		if err != nil {
			return ProgressCompleteMsg{Err: err, Name: "vulns"}
		}
		if output == "" {
			output = "No vulnerabilities found"
		}
		m.program.Send(ProgressLineMsg{Line: output})
		return ProgressCompleteMsg{Err: nil, Name: "vulns"}
	}
}

func (m Model) runMissing() (tea.Model, tea.Cmd) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	m.activeModal = modal.NewProgressModal("Missing Dependencies", cancel)

	return m, func() tea.Msg {
		missing, err := m.client.Diagnostics.Missing(cancelCtx)
		if err != nil {
			return ProgressCompleteMsg{Err: err, Name: "missing"}
		}
		if len(missing) == 0 {
			m.program.Send(ProgressLineMsg{Line: "All dependencies satisfied"})
		} else {
			for _, d := range missing {
				m.program.Send(ProgressLineMsg{Line: d.Formula + ": missing " + d.Missing})
			}
		}
		return ProgressCompleteMsg{Err: nil, Name: "missing"}
	}
}

func (m *Model) loadTabContent() tea.Cmd {
	needsFetch := map[PanelID]map[int]bool{
		PanelStatus:   {1: true, 2: true},
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
			if tab == 2 {
				warnings, err := client.Diagnostics.Doctor(ctx)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, Err: err}
				}
				if len(warnings) == 0 {
					return TabContentMsg{PanelID: panel, TabIndex: tab,
						Content: "Your system is ready to brew."}
				}
				content := ""
				for _, w := range warnings {
					content += w.Title + "\n" + w.Details + "\n\n"
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, Content: content}
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
			output, err := client.Runner.Execute(ctx, "list", name)
			if err != nil {
				return TabContentMsg{PanelID: panel, TabIndex: tab, Err: err}
			}
			content := string(output)
			if content == "" {
				content = "No files installed"
			}
			return TabContentMsg{PanelID: panel, TabIndex: tab, Content: content}
			}
		}
		return TabContentMsg{PanelID: panel, TabIndex: tab, Content: ""}
	}
}

func (m Model) confirmUntap() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelTaps]
	if panel.selected >= len(panel.items) {
		return m, nil
	}
	name := extractPackageName(panel.items[panel.selected])
	if name == "" {
		return m, nil
	}
	if isOfficialTap(name) {
		m.toast = modal.NewToast("Cannot untap official taps", modal.ToastWarning)
		return m, nil
	}
	m.pendingAction = "untap"
	m.activeModal = modal.NewConfirmModal("Untap "+name, "Remove tap "+name+"? Installed formulae from this tap will become unavailable for updates.")
	return m, m.activeModal.Init()
}

func isOfficialTap(name string) bool {
	return strings.HasPrefix(name, "homebrew/")
}

func (m Model) confirmRepair() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelTaps]
	if panel.selected >= len(panel.items) {
		return m, nil
	}
	name := extractPackageName(panel.items[panel.selected])
	if name == "" {
		return m, nil
	}
	m.pendingAction = "repair"
	m.activeModal = modal.NewConfirmModal("Repair "+name, "Repair tap "+name+"?")
	return m, m.activeModal.Init()
}

func (m Model) confirmUninstall(mutType mutationType) (tea.Model, tea.Cmd) {
	panel := m.panels[m.activePanel]
	if panel.selected >= len(panel.items) {
		return m, nil
	}
	name := extractPackageName(panel.items[panel.selected])
	if name == "" {
		return m, nil
	}

	label := "Uninstall"
	message := "Uninstall " + name + "? This cannot be undone."
	if mutType == mutZap {
		label = "Zap"
		message = "Zap " + name + "? This will remove ALL associated files and preferences."
	}

	return m, func() tea.Msg {
		ctx := context.Background()
		uses, _ := m.client.Formulae.Uses(ctx, name)
		if len(uses) > 0 {
			message += "\n\nThe following depend on " + name + ":"
			for _, u := range uses {
				message += "\n  " + u
			}
		}
		return DepCheckMsg{MutType: mutType, Name: name, Label: label, Message: message}
	}
}

func (m Model) runDoctor() (tea.Model, tea.Cmd) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	m.activeModal = modal.NewProgressModal("Doctor", cancel)

	return m, func() tea.Msg {
		warnings, err := m.client.Diagnostics.Doctor(cancelCtx)
		if err != nil {
			return ProgressCompleteMsg{Err: err, Name: "doctor"}
		}
		if len(warnings) == 0 {
			m.program.Send(ProgressLineMsg{Line: "Your system is ready to brew."})
		} else {
			for _, w := range warnings {
				m.program.Send(ProgressLineMsg{Line: w.Title})
				if w.Details != "" {
					m.program.Send(ProgressLineMsg{Line: "  " + w.Details})
				}
			}
		}
		return ProgressCompleteMsg{Err: nil, Name: "doctor"}
	}
}

func (m Model) toggleLeaves() (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		ctx := context.Background()
		leaves, err := m.client.Formulae.Leaves(ctx)
		if err != nil {
			return MutationResultMsg{Name: "leaves", Type: mutInstall, Err: err}
		}
		return MutationResultMsg{Name: "leaves", Type: mutInstall, Leaves: leaves}
	}
}

func (m Model) brewCleanup() (tea.Model, tea.Cmd) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	m.activeModal = modal.NewProgressModal("Cleanup Preview", cancel)

	return m, func() tea.Msg {
		ch, errCh := m.client.DiagnosticsWrite.Cleanup(cancelCtx, true)
		var lines []string
		if ch != nil {
			for line := range ch {
				lines = append(lines, line)
				m.program.Send(ProgressLineMsg{Line: line})
			}
		}
		if errCh != nil {
			err := <-errCh
			if err != nil {
				return ProgressCompleteMsg{Err: err, Name: "cleanup"}
			}
		}
		return CleanupPreviewMsg{Lines: lines}
	}
}

func (m Model) runAutoremove() (tea.Model, tea.Cmd) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	m.activeModal = modal.NewProgressModal("Autoremove Preview", cancel)

	return m, func() tea.Msg {
		ch, errCh := m.client.DiagnosticsWrite.Autoremove(cancelCtx, true)
		var lines []string
		if ch != nil {
			for line := range ch {
				lines = append(lines, line)
				m.program.Send(ProgressLineMsg{Line: line})
			}
		}
		if errCh != nil {
			err := <-errCh
			if err != nil {
				return ProgressCompleteMsg{Err: err, Name: "autoremove"}
			}
		}
		return AutoremovePreviewMsg{Lines: lines}
	}
}

func (m Model) executeCleanup() (tea.Model, tea.Cmd) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	m.activeModal = modal.NewProgressModal("Cleanup", cancel)

	return m, func() tea.Msg {
		ch, errCh := m.client.DiagnosticsWrite.Cleanup(cancelCtx, false)
		if ch != nil {
			for line := range ch {
				m.program.Send(ProgressLineMsg{Line: line})
			}
		}
		var err error
		if errCh != nil {
			err = <-errCh
		}
		return ProgressCompleteMsg{Err: err, Name: "cleanup"}
	}
}

func (m Model) executeAutoremove() (tea.Model, tea.Cmd) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	m.activeModal = modal.NewProgressModal("Autoremove", cancel)

	return m, func() tea.Msg {
		ch, errCh := m.client.DiagnosticsWrite.Autoremove(cancelCtx, false)
		if ch != nil {
			for line := range ch {
				m.program.Send(ProgressLineMsg{Line: line})
			}
		}
		var err error
		if errCh != nil {
			err = <-errCh
		}
		return ProgressCompleteMsg{Err: err, Name: "autoremove"}
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
			return DataLoadedMsg{PanelID: panel, Items: items, Formulae: formulae}

		case PanelCasks:
			casks, err := client.Casks.List(ctx)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			items := make([]string, len(casks))
			for i, c := range casks {
				items[i] = presentation.FormatCask(c, 0)
			}
			return DataLoadedMsg{PanelID: panel, Items: items, Casks: casks}

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
			return DataLoadedMsg{PanelID: panel, Items: items, Taps: taps}

		case PanelServices:
			services, err := client.Services.List(ctx)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			items := make([]string, len(services))
			for i, s := range services {
				items[i] = presentation.FormatService(s, 0)
			}
			return DataLoadedMsg{PanelID: panel, Items: items, Services: services}

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
