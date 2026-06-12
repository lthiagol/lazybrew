package gui

import (
	"strconv"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/brew"
	"github.com/thiago/lazybrew/internal/config"
	"github.com/thiago/lazybrew/internal/gui/modal"
	"github.com/thiago/lazybrew/internal/gui/style"
)

type Model struct {
	client      *brew.Client
	cfg         *config.Config
	width       int
	height      int
	activePanel PanelID
	panels      []*panelData
	activeTab   int
	tabs        []tabInfo
	viewport    viewport.Model
	ready       bool

	activeModal modal.Modal
	toast       *modal.Toast
	batch       *batchState
	showHelp    bool
	tabContent  map[string]string
	program     *tea.Program
}

func (m *Model) Cfg() *config.Config { return m.cfg }

func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

func New(client *brew.Client, cfg *config.Config) *Model {
	panels := initPanels()
	return &Model{
		client:      client,
		cfg:         cfg,
		activePanel: PanelStatus,
		panels:      panels,
		activeTab:   0,
		tabs:        panelTabs[PanelStatus],
		batch:       newBatchState(),
		tabContent:  make(map[string]string),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchPanelData(m.client, PanelFormulae),
		fetchPanelData(m.client, PanelCasks),
		fetchPanelData(m.client, PanelOutdated),
		fetchPanelData(m.client, PanelTaps),
		fetchPanelData(m.client, PanelServices),
		fetchStatusData(m.client),
	)
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

	case SearchDoneMsg:
		p := m.panels[PanelSearch]
		p.loading = false
		if msg.Err != nil {
			p.err = msg.Err
		} else {
			p.items = msg.Results
		}
		m.switchPanel(PanelSearch)
		return m, nil

	case ProgressLineMsg:
		if m.activeModal != nil {
			if p, ok := m.activeModal.(*modal.ProgressModal); ok {
				p.AppendLine(msg.Line)
			}
		}
		return m, nil

	case ProgressCompleteMsg:
		if m.activeModal != nil {
			if p, ok := m.activeModal.(*modal.ProgressModal); ok {
				p.SetDone(msg.Err)
			}
		}
		if msg.Err == nil {
			return m, func() tea.Msg { return RefreshMsg{} }
		}
		return m, nil

	case MutationResultMsg:
		m.activeModal = nil
		if msg.Err == nil {
			m.toast = modal.NewToast(msg.Name+" completed", modal.ToastSuccess)
		} else {
			m.toast = modal.NewToast(msg.Name+": "+msg.Err.Error(), modal.ToastError)
		}
		return m, func() tea.Msg { return RefreshMsg{} }

	case TabContentMsg:
		if msg.Err != nil && msg.Content == "" {
			return m, nil
		}
		key := tabKey(msg.PanelID, msg.TabIndex)
		m.tabContent[key] = msg.Content
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
				p.rawData = msg.RawData
			}
		}
		return m, nil

	case RefreshMsg:
		return m, tea.Batch(
			fetchPanelData(m.client, PanelFormulae),
			fetchPanelData(m.client, PanelCasks),
			fetchPanelData(m.client, PanelOutdated),
			fetchPanelData(m.client, PanelTaps),
			fetchPanelData(m.client, PanelServices),
			fetchStatusData(m.client),
		)

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
		case "k", "up":
			m.panels[m.activePanel].up()

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
				return m.doMutation(mutUninstall, "Uninstall")
			}
		case "X":
			if m.activePanel == PanelCasks {
				return m.doMutation(mutZap, "Zap")
			}
		case "r":
			if m.activePanel == PanelFormulae || m.activePanel == PanelCasks {
				return m.doMutation(mutReinstall, "Reinstall")
			}
		case "u":
			if m.activePanel == PanelOutdated || m.activePanel == PanelFormulae || m.activePanel == PanelCasks {
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

func tabKey(panel PanelID, tab int) string {
	return strconv.Itoa(int(panel)) + ":" + strconv.Itoa(tab)
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
