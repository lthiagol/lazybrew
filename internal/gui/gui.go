package gui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/brew"
	"github.com/thiago/lazybrew/internal/config"
	"github.com/thiago/lazybrew/internal/gui/modal"
	"github.com/thiago/lazybrew/internal/gui/style"
	"github.com/thiago/lazybrew/internal/gui/task"
)

const (
	minTerminalWidth  = 80
	minTerminalHeight = 24
)

type Model struct {
	client      *brew.Client
	cfg         *config.Config
	width       int
	height      int
	terminalTooSmall bool
	activePanel PanelID
	panels      []*panelData
	activeTab   int
	tabs        []tabInfo
	viewport    viewport.Model
	ready       bool

	activeModal     modal.Modal
	toast           *modal.Toast
	batch           *batchState
	showHelp        bool
	tabContent      map[string]string
	program         *tea.Program
	confirmCallback func() tea.Msg
	pendingAction   string
	pendingMutType  mutationType
	batchCount      int
	lastUpdate      time.Time
	isUpdating      bool
	updateOutput    []string

	searchResults      []brew.SearchResult
	searchInfoContent  string

	tasks      *task.Manager
	commandLog *CommandLog
	spinner    spinner.Model
}

func (m *Model) Cfg() *config.Config { return m.cfg }

func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

func New(client *brew.Client, cfg *config.Config) *Model {
	panels := initPanels()
	cl := NewCommandLog(20)
	s := spinner.New(spinner.WithStyle(style.SubtleText))
	return &Model{
		client:      client,
		cfg:         cfg,
		activePanel: PanelStatus,
		panels:      panels,
		activeTab:   0,
		tabs:        panelTabs[PanelStatus],
		batch:       newBatchState(),
		tabContent:  make(map[string]string),
		tasks:       task.NewManager(task.DefaultMaxQueue),
		commandLog:  cl,
		spinner:     s,
	}
}

func (m *Model) CommandLogCallback() brew.CommandCallback {
	cl := m.commandLog
	return func(args []string, err error) {
		cmd := brewCommandString(args)
		cl.Append(cmd)
		if err != nil {
			cl.SetStatus(cmd, CommandError)
		} else {
			cl.SetStatus(cmd, CommandSuccess)
		}
	}
}

func (m *Model) CommandLogStartCallback() brew.CommandCallback {
	cl := m.commandLog
	return func(args []string, _ error) {
		cl.Append(brewCommandString(args))
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	if m.cfg.Brew.UpdateOnStart {
		cmds = append(cmds,
			func() tea.Msg { return StartUpdateMsg{} },
			m.updateTickerCmd(),
		)
	} else {
		cmds = append(cmds,
			fetchPanelData(m.client, PanelFormulae),
			fetchPanelData(m.client, PanelCasks),
			fetchPanelData(m.client, PanelOutdated),
			fetchPanelData(m.client, PanelTaps),
			fetchPanelData(m.client, PanelServices),
			fetchStatusData(m.client),
		)
		if tick := m.autoRefreshCmd(); tick != nil {
			cmds = append(cmds, tick)
		}
	}
	cmds = append(cmds, func() tea.Msg { return m.spinner.Tick() })
	return tea.Batch(cmds...)
}



func (m Model) updateTickerCmd() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return UpdateTickMsg{}
	})
}

func (m Model) autoRefreshCmd() tea.Cmd {
	if m.cfg.GUI.AutoRefreshSeconds <= 0 {
		return nil
	}
	return tea.Tick(time.Duration(m.cfg.GUI.AutoRefreshSeconds)*time.Second, func(t time.Time) tea.Msg {
		return RefreshMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case modal.ToastTickMsg:
		if m.toast != nil {
			updated, _ := m.toast.Update(msg)
			m.toast = updated
			if m.toast.Dismissed() {
				m.toast = nil
			}
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case SearchDoneMsg:
		p := m.panels[PanelSearch]
		p.loading = false
		if msg.Err != nil {
			p.err = msg.Err
		} else {
			p.items = msg.Results
			m.searchResults = msg.Raw
		}
		m.switchPanel(PanelSearch)
		return m, m.fetchSelectedSearchInfo()

	case SearchInfoLoadedMsg:
		if msg.Err != nil {
			m.searchInfoContent = "Error: " + msg.Err.Error()
		} else {
			m.searchInfoContent = msg.Content
		}
		return m, nil

	case ProgressLineMsg:
		if m.activeModal != nil {
			if p, ok := m.activeModal.(*modal.ProgressModal); ok {
				p.AppendLine(msg.Line)
			}
		}
		return m, nil

	case TaskStartedMsg:
		m.activeModal = modal.NewProgressModal(msg.Title, m.tasks.CancelCurrent)
		return m, m.tasks.RunNext()

	case UpdateTickMsg:
		return m, m.updateTickerCmd()

	case StartUpdateMsg:
		if m.isUpdating || !m.cfg.Brew.UpdateOnStart {
			return m, nil
		}
		m.isUpdating = true
		m.updateOutput = nil
		t := &task.Task{
			ID:    "brew-update",
			Title: "brew update",
			Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
				ch, errCh := m.client.Runner.ExecuteStream(ctx, "update")
				if ch == nil {
					ch = closedCh()
				}
				return ch, errCh, nil
			},
		}
		m.tasks.Enqueue(t)
		return m, m.tasks.RunNext()

	case TaskOutputMsg:
		if m.isUpdating {
			m.updateOutput = append(m.updateOutput, msg.Line)
		}
		if m.activeModal != nil {
			if p, ok := m.activeModal.(*modal.ProgressModal); ok {
				p.AppendLine(msg.Line)
			}
		}
		return m, m.tasks.RunNext()

	case TaskCompletedMsg:
		if msg.ID == "brew-update" {
			m.isUpdating = false
			m.lastUpdate = time.Now()
			return m, tea.Batch(
				m.tasks.RunNext(),
				func() tea.Msg { return RefreshMsg{} },
			)
		}
		if m.activeModal != nil {
			if p, ok := m.activeModal.(*modal.ProgressModal); ok {
				p.SetDone(msg.Err)
			}
		}
		if msg.Err != nil {
			m.toast = modal.NewToast(msg.Title+": "+msg.Err.Error(), modal.ToastError)
			return m, m.tasks.RunNext()
		}
		switch msg.ID {
		case "cleanup-preview":
			m.pendingAction = "cleanup"
			m.activeModal = modal.NewConfirmModal("Confirm Cleanup",
				"Run brew cleanup? This will remove old versions.")
			return m, tea.Batch(m.tasks.RunNext(), m.activeModal.Init())
		case "autoremove-preview":
			m.pendingAction = "autoremove"
			m.activeModal = modal.NewConfirmModal("Confirm Autoremove",
				"Remove orphaned dependencies?")
			return m, tea.Batch(m.tasks.RunNext(), m.activeModal.Init())
		}
		if msg.Title != "" {
			m.toast = modal.NewToast(msg.Title+" completed", modal.ToastSuccess)
		}
		if m.batchCount > 0 {
			m.batchCount--
		}
		return m, tea.Batch(
			m.tasks.RunNext(),
			func() tea.Msg { return RefreshMsg{} },
		)

	case TaskRejectedMsg:
		m.toast = modal.NewToast(msg.Reason, modal.ToastWarning)
		return m, nil

	case MutationResultMsg:
		if msg.Leaves != nil && m.activePanel == PanelFormulae {
			p := m.panels[PanelFormulae]
			if !p.leavesActive {
				p.unfilteredItems = p.items
				leavesSet := make(map[string]bool, len(msg.Leaves))
				for _, l := range msg.Leaves {
					leavesSet[l] = true
				}
				filtered := make([]string, 0, len(msg.Leaves))
				for _, item := range p.items {
					name := extractPackageName(item)
					if leavesSet[name] {
						filtered = append(filtered, item)
					}
				}
				p.items = filtered
				p.leavesActive = true
			} else {
				p.items = p.unfilteredItems
				p.unfilteredItems = nil
				p.leavesActive = false
			}
			if p.selected >= len(p.items) {
				p.selected = max(0, len(p.items)-1)
			}
		}
		m.activeModal = nil
		if msg.Err == nil && msg.Leaves == nil {
			m.toast = modal.NewToast(msg.Name+" completed", modal.ToastSuccess)
		} else if msg.Err != nil {
			m.toast = modal.NewToast(msg.Name+": "+msg.Err.Error(), modal.ToastError)
		}
		return m, func() tea.Msg { return RefreshMsg{} }

	case CleanupPreviewMsg:
		m.activeModal = nil
		m.pendingAction = "cleanup"
		m.activeModal = modal.NewConfirmModal("Confirm Cleanup",
			"Run brew cleanup? This will remove old versions.")
		return m, m.activeModal.Init()

	case AutoremovePreviewMsg:
		m.activeModal = nil
		m.pendingAction = "autoremove"
		m.activeModal = modal.NewConfirmModal("Confirm Autoremove",
			"Remove orphaned dependencies?")
		return m, m.activeModal.Init()

	case DepCheckMsg:
		m.pendingAction = "uninstall"
		m.pendingMutType = msg.MutType
		m.activeModal = modal.NewConfirmModal(msg.Label+" "+msg.Name, msg.Message)
		return m, m.activeModal.Init()

	case TabContentMsg:
		if msg.Err != nil && msg.Content == "" {
			return m, nil
		}
		key := tabKey(msg.PanelID, msg.TabIndex, msg.ItemName)
		m.tabContent[key] = msg.Content
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.terminalTooSmall = msg.Width < minTerminalWidth || msg.Height < minTerminalHeight
		if !m.ready {
			m.viewport = viewport.New(max(10, msg.Width-4), max(10, msg.Height-6))
			m.ready = true
		} else {
			m.viewport.Width = max(10, msg.Width-4)
			m.viewport.Height = max(10, msg.Height-6)
		}
		for i := range m.panels {
			m.panels[i].width = sidebarWidth(m.cfg, m.width)
			m.panels[i].height = m.height - 4
		}
		return m, nil

	case DataLoadedMsg:
		if int(msg.PanelID) < len(m.panels) {
			p := m.panels[msg.PanelID]
			p.loading = false
			p.err = msg.Err
			if msg.Err == nil {
				p.items = msg.Items
				p.formulae = msg.Formulae
				p.casks = msg.Casks
				p.taps = msg.Taps
				p.services = msg.Services
			}
		}
		m.clearPanelTabContent(msg.PanelID)
		return m, nil

	case RefreshMsg:
		m.clearTabContent()
		cmds := []tea.Cmd{
			fetchPanelData(m.client, PanelFormulae),
			fetchPanelData(m.client, PanelCasks),
			fetchPanelData(m.client, PanelOutdated),
			fetchPanelData(m.client, PanelTaps),
			fetchPanelData(m.client, PanelServices),
			fetchStatusData(m.client),
		}
		if tick := m.autoRefreshCmd(); tick != nil {
			cmds = append(cmds, tick)
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		if m.activeModal != nil {
			updated, cmd := m.activeModal.Update(msg)
			if modal, ok := updated.(modal.Modal); ok {
				m.activeModal = modal
			} else {
				m.activeModal = nil
			}
			if m.activeModal.Done() || m.activeModal.Cancelled() {
				result := m.activeModal.Result()
				m.activeModal = nil
				return m.handleModalResult(result, cmd)
			}
			return m, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.showHelp {
				m.showHelp = false
			}
			return m, nil
		case "R":
			if m.cfg.Brew.UpdateOnStart && !m.isUpdating {
				m.isUpdating = true
				m.updateOutput = nil
				t := &task.Task{
					ID:    "brew-update",
					Title: "brew update",
					Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
						ch, errCh := m.client.Runner.ExecuteStream(ctx, "update")
						if ch == nil {
							ch = closedCh()
						}
						return ch, errCh, nil
					},
				}
				m.tasks.Enqueue(t)
				return m, m.tasks.RunNext()
			}
			if m.cfg.Brew.UpdateOnStart {
				return m, tea.Batch(
					m.updateTickerCmd(),
					func() tea.Msg { return RefreshMsg{} },
				)
			}
			return m, func() tea.Msg { return RefreshMsg{} }

		case "tab":
			m.nextPanel()
		case "shift+tab":
			m.prevPanel()
		case "1":
			m.switchPanel(PanelStatus)
		case "2":
			m.switchPanel(PanelFormulae)
		case "3":
			m.switchPanel(PanelCasks)
		case "4":
			m.switchPanel(PanelOutdated)
		case "5":
			m.switchPanel(PanelTaps)
		case "6":
			m.switchPanel(PanelServices)
		case "7":
			m.switchPanel(PanelSearch)

		case "j", "down":
			m.panels[m.activePanel].down()
			if m.activePanel == PanelSearch {
				return m, m.fetchSelectedSearchInfo()
			}
			if needsTabFetch(m.activePanel, m.activeTab) {
				return m, m.loadTabContent()
			}
		case "k", "up":
			m.panels[m.activePanel].up()
			if m.activePanel == PanelSearch {
				return m, m.fetchSelectedSearchInfo()
			}
			if needsTabFetch(m.activePanel, m.activeTab) {
				return m, m.loadTabContent()
			}

		case "[":
			cmd = m.prevTab()
		case "]":
			cmd = m.nextTab()

		case "/":
			m, cmd := m.startSearch()
			return m, cmd

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "i":
			if m.activePanel == PanelSearch {
				return m.doMutation(mutInstall, "Install")
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
				return m.doMutation(mutReinstall, "Reinstall")
			}
			if m.activePanel == PanelTaps {
				return m.confirmRepair()
			}
		case "F":
			if m.activePanel == PanelSearch || m.activePanel == PanelFormulae || m.activePanel == PanelCasks {
				return m.doMutation(mutFetch, "Fetch")
			}
		case "u":
			if m.activePanel == PanelOutdated {
				if len(m.batch.selected) > 0 {
					return m.batchUpgrade()
				}
				return m.doMutation(mutUpgrade, "Upgrade")
			}
			if m.activePanel == PanelFormulae || m.activePanel == PanelCasks {
				return m.doMutation(mutUpgrade, "Upgrade")
			}
		case "U":
			return m.doMutation(mutUpgradeAll, "Upgrade All")
		case " ":
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
			if m.activePanel == PanelFormulae || m.activePanel == PanelCasks {
				return m.togglePin(mutInstall)
			}
		}
	}

	if m.activeModal != nil {
		updated, cmd := m.activeModal.Update(msg)
		if modal, ok := updated.(modal.Modal); ok {
			m.activeModal = modal
		}
		return m, cmd
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.terminalTooSmall {
		warning := fmt.Sprintf("Terminal too small (%dx%d). Minimum: %dx%d",
			m.width, m.height, minTerminalWidth, minTerminalHeight)
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(style.ErrorBadge.Render(warning))
	}

	if m.showHelp {
		helpContent := renderHelp(m.width)
		helpView := lipgloss.NewStyle().
			Width(m.width-4).
			Height(m.height-4).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(style.AccentColor).
			Render(helpContent)

		helpView += "\n" + style.SubtleText.Render("  Press ? or Esc to close")
		return helpView
	}

	sidebar := m.renderSidebar()
	mainContent := m.renderMainPanel()
	bottomBar := m.renderBottomBar()

	toastView := ""
	if m.toast != nil {
		toastView = m.toast.View()
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
	var full string
	if toastView != "" {
		full = lipgloss.JoinVertical(lipgloss.Top, body, bottomBar, toastView)
	} else {
		full = lipgloss.JoinVertical(lipgloss.Top, body, bottomBar)
	}

	if m.activeModal != nil {
		modalView := m.activeModal.View()
		overlay := style.ActiveBorder.Render(modalView)
		full = lipgloss.JoinVertical(lipgloss.Center, full, "  ", overlay)
	}

	return full
}

func (m *Model) nextPanel() {
	m.panels[m.activePanel].active = false
	m.activePanel = PanelID((int(m.activePanel) + 1) % len(m.panels))
	m.panels[m.activePanel].active = true
	m.activeTab = 0
	m.tabs = panelTabs[m.activePanel]
}

func (m *Model) prevPanel() {
	m.panels[m.activePanel].active = false
	m.activePanel = PanelID((int(m.activePanel) - 1 + len(m.panels)) % len(m.panels))
	m.panels[m.activePanel].active = true
	m.activeTab = 0
	m.tabs = panelTabs[m.activePanel]
}

func (m *Model) switchPanel(id PanelID) {
	if int(id) >= len(m.panels) {
		return
	}
	m.panels[m.activePanel].active = false
	m.activePanel = id
	m.panels[m.activePanel].active = true
	m.activeTab = 0
	m.tabs = panelTabs[m.activePanel]
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

func needsTabFetch(panelID PanelID, tabIdx int) bool {
	needsFetch := map[PanelID]map[int]bool{
		PanelStatus:   {1: true, 2: true},
		PanelFormulae: {1: true, 2: true, 4: true},
	}
	return needsFetch[panelID] != nil && needsFetch[panelID][tabIdx]
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

func selectedItemName(p *panelData) string {
	if p.selected >= len(p.items) {
		return ""
	}
	return extractPackageName(p.items[p.selected])
}

func extractPackageName(item string) string {
	for i, r := range item {
		if r == ' ' || r == '\t' {
			return item[:i]
		}
	}
	return item
}

func sidebarWidth(cfg *config.Config, totalWidth int) int {
	pct := cfg.GUI.SidebarWidth
	if pct < 15 {
		pct = 15
	}
	if pct > 50 {
		pct = 50
	}
	w := totalWidth * pct / 100
	if w < 20 {
		w = 20
	}
	if w > 40 {
		w = 40
	}
	return w
}
