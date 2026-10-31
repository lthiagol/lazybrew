package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// fetchTabContentTimeout bounds the time a tab-content fetch (brew deps --tree,
// brew uses, brew list, doctor, config) is allowed to run before it is
// cancelled and surfaced as an error in the tab body. The default is generous
// because brew deps --tree can be slow on big trees; tests override it.
var fetchTabContentTimeout = 15 * time.Second

// brewReadTimeout bounds a single brew read in the panel and status paths.
// A wedged `brew` invocation must never leave a panel spinning forever, so
// every individual read gets its own deadline; tests override it.
var brewReadTimeout = 30 * time.Second

// brewSearchTimeout bounds a single brew read in the search paths. Search is
// user-driven and interactive, so it gets a tighter cap than the passive
// panel and dashboard reads; tests override it.
var brewSearchTimeout = 15 * time.Second

// brewStatusTotalTimeout bounds the status dashboard as a whole. The
// dashboard aggregates several independent brew reads, so bounding each read
// is not enough: if every source hangs, the dashboard must still resolve.
// This budget is deliberately separate from (and larger than) the per-read
// cap — one slow source may not eat the whole dashboard's time; tests
// override it.
var brewStatusTotalTimeout = 45 * time.Second

// formatBudget renders a read budget at a readable granularity: whole
// seconds for the multi-second budgets, milliseconds below that.
func formatBudget(d time.Duration) time.Duration {
	if d >= time.Second {
		return d.Round(time.Second)
	}
	return d.Round(time.Millisecond)
}

// timeoutError converts a failed brew read into a *brew.TimeoutError when the
// read's deadline expired.
//
// The deadline is detected from the context's own error, never from the type
// of the error the brew layer returned: exec.CommandContext kills the brew
// process when the deadline fires, so the runner reports "signal: killed"
// (wrapped in a BrewExitError) rather than context.DeadlineExceeded. Reading
// the error's type instead would silently never fire.
func timeoutError(ctx context.Context, command string, budget time.Duration, err error) error {
	if err == nil {
		return nil
	}
	if ctx.Err() == context.DeadlineExceeded {
		return &brew.TimeoutError{Command: command, Timeout: budget}
	}
	return err
}

// boundedRead runs one brew read under its own deadline and turns a deadline
// overrun into a *brew.TimeoutError naming the brew command.
func boundedRead[T any](cap time.Duration, command string, read func(context.Context) (T, error)) (T, error) {
	return boundedReadIn(context.Background(), cap, command, read)
}

// boundedReadIn runs one brew read under its own deadline, derived from
// parent, and turns a deadline overrun into a *brew.TimeoutError naming the
// brew command.
//
// The read's own deadline is capped by cap and by whatever budget the parent
// has left: a parent deadline trims the read, so an aggregate can bound its
// total time without every read knowing about it.
func boundedReadIn[T any](parent context.Context, cap time.Duration, command string, read func(context.Context) (T, error)) (T, error) {
	var zero T

	// The budget the user is told about: the read's own cap, unless the
	// parent's deadline cuts it shorter.
	budget := cap
	var remaining time.Duration
	hasDeadline := false
	if parentDeadline, ok := parent.Deadline(); ok {
		remaining = time.Until(parentDeadline)
		hasDeadline = true
		if remaining < budget {
			budget = remaining
		}
	}

	// The parent's deadline has already passed: it has no budget left and
	// the remaining time is negative. The read has no chance, so it is not
	// started at all — the brew client is not invoked with a dead context —
	// and the timeout is reported right away. The reported budget is
	// clamped to zero, so the message never reads "timed out after -1s".
	if hasDeadline && remaining <= 0 {
		return zero, &brew.TimeoutError{Command: command, Timeout: formatBudget(0)}
	}

	// The clamp: the budget handed to the user is never negative.
	if budget < 0 {
		budget = 0
	}

	ctx, cancel := context.WithTimeout(parent, cap)
	defer cancel()

	res, err := read(ctx)
	if err != nil {
		err = timeoutError(ctx, command, formatBudget(budget), err)
	}
	return res, err
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
			case "mutation":
				labels := map[mutationType]string{
					mutInstall:    "Install",
					mutReinstall:  "Reinstall",
					mutUpgrade:    "Upgrade",
					mutUpgradeAll: "Upgrade All",
				}
				label := labels[m.pendingMutType]
				if label == "" {
					label = "Install"
				}
				return m.doMutation(m.pendingMutType, label)
			case "batch-upgrade":
				return m.batchUpgrade()
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
	title := "Brewfile " + action

	t := &task.Task{
		ID:    "brewfile-" + action,
		Title: title,
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			ch, errCh := m.client.Runner.ExecuteStream(ctx, "bundle", action)
			if ch == nil {
				ch = closedCh()
			}
			return ch, errCh, nil
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

func (m *Model) executeSearch(query string) tea.Cmd {
	m.panels[PanelSearch].loading = true
	m.panels[PanelSearch].err = nil

	return func() tea.Msg {
		results, err := boundedRead(brewSearchTimeout, "search", func(ctx context.Context) ([]brew.SearchResult, error) {
			return m.client.Search.Search(ctx, query)
		})
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
		return SearchDoneMsg{Results: items, Raw: results}
	}
}

func (m Model) fetchSelectedSearchInfo() tea.Cmd {
	if m.activePanel != PanelSearch || m.panels[PanelSearch].selected >= len(m.searchResults) {
		return nil
	}
	r := m.searchResults[m.panels[PanelSearch].selected]
	return func() tea.Msg {
		output, err := boundedRead(brewSearchTimeout, "info", func(ctx context.Context) ([]byte, error) {
			return m.client.Runner.Execute(ctx, "info", "--json=v2", r.Name)
		})
		if err != nil {
			return SearchInfoLoadedMsg{Err: err}
		}
		return SearchInfoLoadedMsg{Content: string(output)}
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

func (m Model) batchUpgrade() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelOutdated]
	if len(m.batch.selected) == 0 {
		return m, nil
	}

	names := make([]struct {
		name   string
		isCask bool
	}, 0, len(m.batch.selected))

	for idx := range m.batch.selected {
		if idx >= len(panel.items) {
			continue
		}
		itemName := extractPackageName(panel.items[idx])
		if itemName == "" {
			continue
		}
		isCask := false
		if idx >= len(panel.formulae) && idx-len(panel.formulae) < len(panel.casks) {
			isCask = true
		}
		names = append(names, struct {
			name   string
			isCask bool
		}{itemName, isCask})
	}

	m.batchCount = len(names)
	m.batch.selected = make(map[int]bool)

	for _, n := range names {
		name := n.name
		isCask := n.isCask

		t := &task.Task{
			ID:    name,
			Title: "Upgrade " + name,
			Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
				var ch <-chan string
				var errCh <-chan error
				if isCask {
					ch, errCh = m.client.CasksWrite.Upgrade(ctx, name)
				} else {
					ch, errCh = m.client.FormulaeWrite.Upgrade(ctx, name)
				}
				if ch == nil {
					ch = closedCh()
				}
				return ch, errCh, nil
			},
		}

		m.tasks.Enqueue(t)
	}

	return m, m.tasks.RunNext()
}

type pkgInfo struct {
	Name         string
	Version      string
	Type         string
	Bottled      bool
	Installed    bool
	InstallPath  string
	License      string
	Description  string
	Homepage     string
	Dependencies []string
	Caveats      string
}

func parsePackageInfo(rawJSON string) (*pkgInfo, error) {
	var result struct {
		Formulae []brew.Formula `json:"formulae"`
		Casks    []brew.Cask    `json:"casks"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return nil, err
	}
	if len(result.Formulae) > 0 {
		f := result.Formulae[0]
		return &pkgInfo{
			Name:         f.Name,
			Version:      f.Version,
			Type:         "formula",
			Bottled:      f.Bottled,
			Installed:    f.InstalledOnReq || f.InstalledAsDep,
			InstallPath:  f.InstallPath,
			License:      f.License,
			Description:  f.Description,
			Homepage:     f.Homepage,
			Dependencies: append(f.Dependencies, f.BuildDeps...),
			Caveats:      f.Caveats,
		}, nil
	}
	if len(result.Casks) > 0 {
		c := result.Casks[0]
		return &pkgInfo{
			Name:         c.Name,
			Version:      c.Version,
			Type:         "cask",
			Installed:    false,
			Description:  c.Description,
			Homepage:     c.Homepage,
			Dependencies: c.DependsOn,
		}, nil
	}
	return nil, fmt.Errorf("no package info found")
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
				c := panel.selectedCask()
				if c != nil && c.Pinned {
					err = m.client.CasksWrite.Unpin(ctx, name)
				} else {
					err = m.client.CasksWrite.Pin(ctx, name)
				}
			} else {
				f := panel.selectedFormula()
				if f != nil && f.Pinned {
					err = m.client.FormulaeWrite.Unpin(ctx, name)
				} else {
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
	t := &task.Task{
		ID:    "vulns",
		Title: "Vulnerability Check",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			out := make(chan string)
			errCh := make(chan error, 1)
			go func() {
				output, err := m.client.Diagnostics.Vulns(ctx)
				if err != nil {
					errCh <- err
					close(out)
					return
				}
				if output == "" {
					output = "No vulnerabilities found"
				}
				out <- output
				errCh <- nil
				close(out)
			}()
			return out, errCh, nil
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

func (m Model) runMissing() (tea.Model, tea.Cmd) {
	t := &task.Task{
		ID:    "missing",
		Title: "Missing Dependencies",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			out := make(chan string)
			errCh := make(chan error, 1)
			go func() {
				missing, err := m.client.Diagnostics.Missing(ctx)
				if err != nil {
					errCh <- err
					close(out)
					return
				}
				if len(missing) == 0 {
					out <- "All dependencies satisfied"
				} else {
					for _, d := range missing {
						out <- d.Formula + ": missing " + d.Missing
					}
				}
				errCh <- nil
				close(out)
			}()
			return out, errCh, nil
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

func (m *Model) loadTabContent() tea.Cmd {
	needsFetch := map[PanelID]map[int]bool{
		PanelStatus:   {1: true, 2: true},
		PanelFormulae: {2: true, 3: true, 5: true},
	}

	panelID := m.activePanel
	tabIdx := m.activeTab

	if needsFetch[panelID] != nil && needsFetch[panelID][tabIdx] {
		panel := m.panels[panelID]
		if panel.selected >= len(panel.items) {
			return nil
		}
		name := extractPackageName(panel.selectedItem())
		if name == "" {
			return nil
		}
		key := tabKey(panelID, tabIdx, name)
		if _, ok := m.tabContent[key]; ok {
			return nil
		}
		if content, ok := renderCachedTabContent(panelID, tabIdx, panel); ok {
			m.tabContent[key] = content
			return nil
		}
		return fetchTabContentCmd(m.client, panelID, tabIdx, name)
	}
	return nil
}

// renderCachedTabContent returns a body string for tabs whose content is
// already on the model (avoids shelling out for data we already have).
// Returns ok=false when no fast path applies so the caller falls through to
// fetchTabContentCmd. The boolean lets the caller distinguish "rendered
// something terminal" from "no fast path" — empty strings still count as
// a terminal state (e.g. "No dependencies").
func renderCachedTabContent(panelID PanelID, tabIdx int, panel *panelData) (string, bool) {
	if panelID == PanelFormulae && tabIdx == 2 {
		f := panel.selectedFormula()
		if f == nil {
			return "", false
		}
		return renderFormulaDeps(f), true
	}
	return "", false
}

// renderFormulaDeps formats the flat dependency list already on the
// Formula model (runtime + build deps). Returns "No dependencies" when the
// formula declares none, so the tab always reaches a terminal state without
// shelling out.
func renderFormulaDeps(f *brew.Formula) string {
	runtime := f.Dependencies
	build := f.BuildDeps
	if len(runtime) == 0 && len(build) == 0 {
		return "No dependencies"
	}
	var b strings.Builder
	if len(runtime) > 0 {
		b.WriteString("Runtime:\n")
		for _, d := range runtime {
			b.WriteString("  " + d + "\n")
		}
	}
	if len(build) > 0 {
		b.WriteString("Build:\n")
		for _, d := range build {
			b.WriteString("  " + d + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func fetchTabContentCmd(client *brew.Client, panel PanelID, tab int, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTabContentTimeout)
		defer cancel()
		switch panel {
		case PanelStatus:
			if tab == 1 {
				cfg, err := client.Diagnostics.Config(ctx)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Err: err}
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name,
					Content: formatConfig(cfg)}
			}
			if tab == 2 {
				warnings, err := client.Diagnostics.Doctor(ctx)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Err: err}
				}
				if len(warnings) == 0 {
					return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name,
						Content: "Your system is ready to brew."}
				}
				content := ""
				for _, w := range warnings {
					content += w.Title + "\n" + w.Details + "\n\n"
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Content: content}
			}
		case PanelFormulae:
			switch tab {
			case 2: // Deps (List=0, Info=1)
				deps, err := client.Formulae.Deps(ctx, name)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Err: err}
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Content: deps}
			case 3: // Used By
				uses, err := client.Formulae.Uses(ctx, name)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Err: err}
				}
				result := ""
				for _, u := range uses {
					result += u + "\n"
				}
				if result == "" {
					result = "No dependents"
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Content: result}
			case 5: // Files
				output, err := client.Runner.Execute(ctx, "list", name)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Err: err}
				}
				content := string(output)
				if content == "" {
					content = "No files installed"
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Content: content}
			}
		}
		return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Content: ""}
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

func (m Model) confirmMutation(mutType mutationType, label string) (tea.Model, tea.Cmd) {
	panel := m.panels[m.activePanel]
	if panel.selected >= len(panel.items) {
		return m, nil
	}
	name := extractPackageName(panel.items[panel.selected])
	if name == "" {
		return m, nil
	}

	actionLabel := label
	confirmLabel := label + " " + name
	message := confirmLabel + "?"
	if mutType == mutUpgradeAll {
		actionLabel = "Upgrade All"
		message = "Upgrade all packages?"
	} else if mutType == mutReinstall {
		message = "Reinstall " + name + "?"
	}

	m.pendingAction = "mutation"
	m.pendingMutType = mutType
	modal := modal.NewConfirmModal("Confirm "+actionLabel, message)
	m.activeModal = modal
	return m, m.activeModal.Init()
}

func (m Model) confirmBatchUpgrade() (tea.Model, tea.Cmd) {
	n := len(m.batch.selected)
	message := "Upgrade " + strconv.Itoa(n) + " selected packages?"

	m.pendingAction = "batch-upgrade"
	modal := modal.NewConfirmModal("Confirm Batch Upgrade", message)
	m.activeModal = modal
	return m, m.activeModal.Init()
}

func (m Model) runDoctor() (tea.Model, tea.Cmd) {
	t := &task.Task{
		ID:    "doctor",
		Title: "Doctor",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			out := make(chan string)
			errCh := make(chan error, 1)
			go func() {
				warnings, err := m.client.Diagnostics.Doctor(ctx)
				if err != nil {
					errCh <- err
					close(out)
					return
				}
				if len(warnings) == 0 {
					out <- "Your system is ready to brew."
				} else {
					for _, w := range warnings {
						out <- w.Title
						if w.Details != "" {
							out <- "  " + w.Details
						}
					}
				}
				errCh <- nil
				close(out)
			}()
			return out, errCh, nil
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
	t := &task.Task{
		ID:    "cleanup-preview",
		Title: "Cleanup Preview",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			ch, errCh := m.client.DiagnosticsWrite.Cleanup(ctx, true)
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

func (m Model) runAutoremove() (tea.Model, tea.Cmd) {
	t := &task.Task{
		ID:    "autoremove-preview",
		Title: "Autoremove Preview",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			ch, errCh := m.client.DiagnosticsWrite.Autoremove(ctx, true)
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

func (m Model) executeCleanup() (tea.Model, tea.Cmd) {
	t := &task.Task{
		ID:    "cleanup",
		Title: "Cleanup",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			ch, errCh := m.client.DiagnosticsWrite.Cleanup(ctx, false)
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

func (m Model) executeAutoremove() (tea.Model, tea.Cmd) {
	t := &task.Task{
		ID:    "autoremove",
		Title: "Autoremove",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			ch, errCh := m.client.DiagnosticsWrite.Autoremove(ctx, false)
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
		switch panel {
		case PanelFormulae:
			formulae, err := boundedRead(brewReadTimeout, "list", client.Formulae.List)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			sortFormulae(formulae)
			items := make([]string, len(formulae))
			for i, f := range formulae {
				items[i] = presentation.FormatFormula(f, 0)
			}
			return DataLoadedMsg{PanelID: panel, Items: items, Formulae: formulae}

		case PanelCasks:
			casks, err := boundedRead(brewReadTimeout, "list", client.Casks.List)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			items := make([]string, len(casks))
			for i, c := range casks {
				items[i] = presentation.FormatCask(c, 0)
			}
			return DataLoadedMsg{PanelID: panel, Items: items, Casks: casks}

		case PanelOutdated:
			// Each outdated read gets its own deadline: one slow source
			// must not eat the whole budget of the other.
			formulae, fErr := boundedRead(brewReadTimeout, "outdated", client.Formulae.Outdated)
			casks, cErr := boundedRead(brewReadTimeout, "outdated", client.Casks.Outdated)
			if fErr != nil {
				return DataLoadedMsg{PanelID: panel, Err: fErr}
			}
			if cErr != nil {
				return DataLoadedMsg{PanelID: panel, Err: cErr}
			}
			items := make([]string, 0, len(formulae)+len(casks))
			for _, f := range formulae {
				items = append(items, presentation.FormatOutdatedFormula(f))
			}
			for _, c := range casks {
				items = append(items, presentation.FormatOutdatedCask(c))
			}
			return DataLoadedMsg{PanelID: panel, Items: items, Formulae: formulae, Casks: casks}

		case PanelTaps:
			taps, err := boundedRead(brewReadTimeout, "taps", client.Taps.List)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			items := make([]string, len(taps))
			for i, t := range taps {
				items[i] = presentation.FormatTap(t, 0)
			}
			return DataLoadedMsg{PanelID: panel, Items: items, Taps: taps}

		case PanelServices:
			services, err := boundedRead(brewReadTimeout, "services", client.Services.List)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			items := make([]string, len(services))
			for i, s := range services {
				items[i] = presentation.FormatService(s, 0)
			}
			return DataLoadedMsg{PanelID: panel, Items: items, Services: services}

		default:
			// PanelStatus and PanelSearch are not fetched here. Init and
			// RefreshMsg use fetchStatusData for Status; Search is
			// user-driven. Do not route Status through this default —
			// an empty Items payload would wipe the dashboard.
			return DataLoadedMsg{PanelID: panel, Items: []string{}}
		}
	}
}

func fetchStatusData(client *brew.Client) tea.Cmd {
	return func() tea.Msg {
		items, _ := statusDashboard(client)
		return DataLoadedMsg{PanelID: PanelStatus, Items: items, Err: nil}
	}
}

// statusDashboard reads every source the status dashboard aggregates and
// renders the dashboard lines.
//
// The dashboard aggregates five independent sources plus brew doctor. Each
// read is capped on its own (brewReadTimeout), and the aggregation as a
// whole is bounded by a separate total budget (brewStatusTotalTimeout):
// bounding the reads alone would still let five slow sources add up to
// minutes. One slow source can therefore consume at most its own cap of the
// budget, while the other sources keep theirs.
//
// The reads are independent, so they run concurrently. The goroutines only
// fill in local results and are joined before this function returns; no
// Bubble Tea state is touched from them.
//
// The returned errors are the aggregated per-source failures — the carrier
// through which a timed-out read reaches the user as a ⚠ dashboard line.
func statusDashboard(client *brew.Client) ([]string, []error) {
	totalCtx, cancelTotal := context.WithTimeout(context.Background(), brewStatusTotalTimeout)
	defer cancelTotal()

	var (
		formulae []brew.Formula
		casks    []brew.Cask
		taps     []brew.Tap
		services []brew.Service
		cfg      *brew.BrewConfig
		warnings []brew.DoctorWarning

		fErr, cErr, tErr, sErr, docErr error
	)

	var wg sync.WaitGroup
	read := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	read(func() {
		formulae, fErr = boundedReadIn(totalCtx, brewReadTimeout, "list", client.Formulae.List)
	})
	read(func() {
		casks, cErr = boundedReadIn(totalCtx, brewReadTimeout, "list", client.Casks.List)
	})
	read(func() {
		taps, tErr = boundedReadIn(totalCtx, brewReadTimeout, "taps", client.Taps.List)
	})
	read(func() {
		services, sErr = boundedReadIn(totalCtx, brewReadTimeout, "services", client.Services.List)
	})
	read(func() {
		// Config errors stay ignored: the dashboard has always tolerated a
		// missing config. The deadline still applies so it cannot hang.
		cfg, _ = boundedReadIn(totalCtx, brewReadTimeout, "config", client.Diagnostics.Config)
	})
	read(func() {
		warnings, docErr = boundedReadIn(totalCtx, brewReadTimeout, "doctor", client.Diagnostics.Doctor)
	})
	wg.Wait()

	// Outdated is lazy (M9): read from cache only. The cache fills when
	// the user first visits the Outdated panel; until then we report 0
	// outdated in the dashboard and the user is not surprised by a
	// `brew outdated` shell invocation on every Init / Refresh.
	var outdatedFormulae []brew.Formula
	if cached, ok := client.Cache.Get(brew.KeyOutdatedFormulae); ok {
		if f, ok := cached.([]brew.Formula); ok {
			outdatedFormulae = f
		}
	}
	var outdatedCasks []brew.Cask
	if cached, ok := client.Cache.Get(brew.KeyOutdatedCasks); ok {
		if c, ok := cached.([]brew.Cask); ok {
			outdatedCasks = c
		}
	}

	// The aggregated errors carry the per-source failures. Each entry keeps
	// the source label the dashboard shows next to the message.
	type sourceErr struct {
		label string
		err   error
	}
	var sources []sourceErr
	if fErr != nil {
		sources = append(sources, sourceErr{"formulae", fErr})
	}
	if cErr != nil {
		sources = append(sources, sourceErr{"casks", cErr})
	}
	if tErr != nil {
		sources = append(sources, sourceErr{"taps", tErr})
	}
	if sErr != nil {
		sources = append(sources, sourceErr{"services", sErr})
	}
	if docErr != nil {
		sources = append(sources, sourceErr{"doctor", docErr})
	}

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

	items = append(items, "")
	items = append(items, presentation.FormatDoctorStatus(warnings, docErr))

	errs := make([]error, 0, len(sources))
	for _, s := range sources {
		items = append(items, "⚠ "+s.label+": "+s.err.Error())
		errs = append(errs, s.err)
	}

	return items, errs
}

func parseUpdateSummary(lines []string) string {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "Already up-to-date") {
			return "Already up to date"
		}
		if strings.HasPrefix(line, "Updated") {
			return strings.TrimRight(line, ".")
		}
		if strings.HasPrefix(line, "Error:") {
			return ""
		}
	}
	return ""
}
