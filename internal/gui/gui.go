package gui

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

const (
	minTerminalWidth  = 80
	minTerminalHeight = 24
)

type Model struct {
	client           *brew.Client
	cfg              *config.Config
	width            int
	height           int
	terminalTooSmall bool
	activePanel      PanelID
	panels           []*panelData
	activeTab        int
	tabs             []tabInfo
	viewport         viewport.Model
	ready            bool

	activeModal    modal.Modal
	toast          *modal.Toast
	batch          *batchState
	showHelp       bool
	tabContent     map[string]string
	pendingAction  string
	pendingMutType mutationType
	batchCount     int
	// batchIDs counts the occurrences of each task ID enqueued by the
	// current batch upgrade, so batchCount is only decremented by batch
	// completions, never by unrelated ones (r1 finding 27). A formula and
	// a cask can share a name in one batch, yielding two tasks with the
	// same ID: counting occurrences (instead of a bool set) keeps the
	// counter honest until the last duplicate completes (r3 finding R3-2).
	batchIDs map[string]int
	// batchTotal is the enqueued size of the current batch; together with
	// batchCount it surfaces "n/total done" in the Operation overlay so
	// the counter machinery has a user-visible effect (r3 finding R3-5).
	batchTotal int
	// batchSeq numbers each batch so its task IDs
	// ("batch-upgrade:<seq>:<name>") cannot collide with a task still
	// queued from a superseded batch (r4 finding R4-4).
	batchSeq int
	// pendingTrustFormulaIdx / pendingTrustCaskIdx record the positions
	// of the submenu entries in the trust menu built by startTrustMenu;
	// they exist only when the tap has formulae/casks, so their
	// positions vary and must not be hardcoded (r4 finding R4-3).
	pendingTrustFormulaIdx int
	pendingTrustCaskIdx    int
	lastUpdate             time.Time
	isUpdating             bool
	updateOutput           []string

	searchResults     []brew.SearchResult
	searchInfoContent string
	searchInput       textinput.Model

	tasks      *task.Manager
	commandLog *CommandLog
	spinner    spinner.Model
	opState    *Operation
	opDoneAt   time.Time
	// taskID was removed (r3 finding R3-6): nothing read it, and stale
	// output filtering belongs to the TaskManager, not the model.
	refreshing int

	// logCollapsed hides the command log pane (Shift+C / "C"). Default false.
	logCollapsed bool

	// lastKeyAt tracks the last time the user pressed a key. The auto
	// refresh tick consults this to pause refreshes during interaction
	// (M12 AC-03) so typing/navigation is not janked by a full
	// refresh storm.
	lastKeyAt time.Time
	// autoRefreshPaused is set by the tick handler when lastKeyAt is
	// within the pause window. The next tick will re-evaluate; the
	// pending refresh is rescheduled for one AutoRefreshSeconds later.
	autoRefreshPaused bool
}

func (m *Model) Cfg() *config.Config { return m.cfg }

// setToast shows a toast and arms its auto-dismiss tick. Every toast
// producer must go through this: the only producer of ToastTickMsg is
// Toast.Update itself, so without an armed first tick a toast would stay
// on screen forever (r1 finding 4).
func (m *Model) setToast(message string, t modal.ToastType) tea.Cmd {
	m.toast = modal.NewToast(message, t)
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return modal.ToastTickMsg{}
	})
}

func New(client *brew.Client, cfg *config.Config) *Model {
	panels := initPanels()
	cl := NewCommandLog(20)
	s := spinner.New(spinner.WithStyle(style.AccentText))
	s.Spinner = spinner.Dot
	ti := textinput.New()
	ti.Placeholder = "Search packages..."
	ti.Width = 40
	m := &Model{
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
		searchInput: ti,
	}
	// -1 marks "no such menu entry" until startTrustMenu records the
	// real positions (r4 finding R4-3).
	m.pendingTrustFormulaIdx = -1
	m.pendingTrustCaskIdx = -1
	return m
}

func (m *Model) CommandLogCallback() brew.CommandCallback {
	cl := m.commandLog
	return func(args []string, err error) {
		id := cl.Append(brewCommandString(args))
		if err != nil {
			cl.SetStatus(id, CommandError)
		} else {
			cl.SetStatus(id, CommandSuccess)
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
		// M12: Init only fetches the active panel (default = Status,
		// which doubles as the dashboard). Other panels lazy-load on
		// first switch via lazyLoadPanel().
		if m.activePanel == PanelStatus {
			cmds = append(cmds, fetchStatusData(m.client, false))
		} else {
			cmds = append(cmds, fetchPanelData(m.client, m.activePanel))
		}
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

// autoRefreshCmd schedules the next auto-refresh tick. The tick delivers
// autoRefreshTickMsg so Update can evaluate pause-on-interaction against
// the live model (M12 AC-03) — never capture Model in the Tick closure.
func (m Model) autoRefreshCmd() tea.Cmd {
	if m.cfg.GUI.AutoRefreshSeconds <= 0 {
		return nil
	}
	d := time.Duration(m.cfg.GUI.AutoRefreshSeconds) * time.Second
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return autoRefreshTickMsg{at: t}
	})
}

// evaluateAutoRefresh decides whether to pause or refresh. Pure w.r.t.
// model fields; callers apply the returned msg via Update.
func (m Model) evaluateAutoRefresh(now time.Time) tea.Msg {
	pause := m.cfg.GUI.AutoRefreshPause
	if pause > 0 && !m.lastKeyAt.IsZero() && now.Sub(m.lastKeyAt) < pause {
		return autoRefreshPausedMsg{}
	}
	return RefreshMsg{}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case modal.ToastTickMsg:
		if m.toast != nil {
			updated, tickCmd := m.toast.Update(msg)
			m.toast = updated
			if m.toast.Dismissed() {
				m.toast = nil
			} else if tickCmd != nil {
				// Re-arm the dismissal chain — without this the first
				// tick would be the last and the toast would stick.
				return m, tickCmd
			}
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case autoRefreshTickMsg:
		// Evaluate pause against the *current* model (lastKeyAt may have
		// advanced since the tick was scheduled).
		next := m.evaluateAutoRefresh(msg.at)
		if _, paused := next.(autoRefreshPausedMsg); paused {
			m.autoRefreshPaused = true
		} else {
			m.autoRefreshPaused = false
		}
		return m, func() tea.Msg { return next }

	case autoRefreshPausedMsg:
		// Reschedule a full tick so pause can re-apply if the user is
		// still typing when the next interval elapses.
		m.autoRefreshPaused = true
		return m, m.autoRefreshCmd()

	case SearchDoneMsg:
		p := m.panels[PanelSearch]
		p.loading = false
		if msg.Err != nil {
			p.err = msg.Err
		} else {
			p.items = msg.Results
			m.searchResults = msg.Raw
		}
		p.normalize()
		cmd := m.switchPanel(PanelSearch)
		return m, cmd

	case SearchInfoLoadedMsg:
		if msg.Err != nil {
			m.searchInfoContent = "Error: " + msg.Err.Error()
		} else {
			m.searchInfoContent = msg.Content
		}
		return m, nil

	case ProgressLineMsg:
		if m.opState != nil {
			m.opState.AppendLine(msg.Line)
		}
		return m, nil

	case TaskStartedMsg:
		title := msg.Title
		// Surface the batch counter in the Operation overlay: batchCount
		// used to be write-only state with no user-visible effect (r3
		// finding R3-5).
		if _, ok := m.batchIDs[msg.ID]; ok && m.batchTotal > 0 {
			title = fmt.Sprintf("%s (%d/%d done)", msg.Title, m.batchTotal-m.batchCount, m.batchTotal)
		}
		m.opState = &Operation{
			Title:  title,
			Status: opRunning,
		}
		m.opDoneAt = time.Time{}
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
		// A dropped enqueue would leave isUpdating stuck true and block
		// every later R update, so surface the failure and reset the flag
		// (r1 finding 7).
		if _, err := m.tasks.Enqueue(t); err != nil {
			m.isUpdating = false
			return m, m.setToast("Queue full: "+err.Error(), modal.ToastWarning)
		}
		return m, m.tasks.RunNext()

	case TaskOutputMsg:
		if m.isUpdating {
			m.updateOutput = append(m.updateOutput, msg.Line)
		}
		if m.opState != nil {
			m.opState.AppendLine(msg.Line)
		}
		return m, m.tasks.RunNext()

	case TaskCompletedMsg:
		if msg.ID == "brew-update" {
			m.isUpdating = false
			m.lastUpdate = time.Now()
			m.opState = nil
			var toastTick tea.Cmd
			if msg.Err != nil {
				toastTick = m.setToast("Update: "+msg.Err.Error(), modal.ToastError)
			} else {
				summary := parseUpdateSummary(m.updateOutput)
				if summary != "" {
					toastTick = m.setToast(summary, modal.ToastSuccess)
				}
			}
			cmds := []tea.Cmd{
				m.tasks.RunNext(),
				func() tea.Msg { return RefreshMsg{} },
			}
			if toastTick != nil {
				cmds = append(cmds, toastTick)
			}
			return m, tea.Batch(cmds...)
		}
		if m.opState != nil {
			if msg.Err != nil {
				if errors.Is(msg.Err, context.Canceled) {
					m.opState.Status = opCancelled
				} else {
					m.opState.Status = opError
					m.opState.Err = msg.Err
				}
			} else {
				m.opState.Status = opSuccess
			}
			m.opDoneAt = time.Now()
		}
		// Batch bookkeeping applies to success, failure and cancel alike:
		// decrementing only on the success path left failed batch tasks
		// stuck in batchIDs with a permanently inflated batchCount, and a
		// later unrelated completion of the same ID could corrupt a future
		// batch's counter (r2 finding NEW-2).
		if n := m.batchIDs[msg.ID]; n > 0 {
			n--
			if n == 0 {
				delete(m.batchIDs, msg.ID)
			} else {
				m.batchIDs[msg.ID] = n
			}
			if m.batchCount > 0 {
				m.batchCount--
			}
			if m.batchCount == 0 {
				m.batchTotal = 0
			}
		}
		if msg.Err != nil && !errors.Is(msg.Err, context.Canceled) {
			return m, tea.Batch(
				m.setToast(msg.Title+": "+msg.Err.Error(), modal.ToastError),
				m.tasks.RunNext(),
			)
		}
		if errors.Is(msg.Err, context.Canceled) {
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
		var toastTick tea.Cmd
		if msg.Title != "" {
			toastTick = m.setToast(msg.Title+" completed", modal.ToastSuccess)
		}
		cmds := []tea.Cmd{
			m.tasks.RunNext(),
			func() tea.Msg { return RefreshMsg{} },
		}
		if toastTick != nil {
			cmds = append(cmds, toastTick)
		}
		return m, tea.Batch(cmds...)

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
			p.normalize()
		}
		m.activeModal = nil
		if msg.Err == nil && msg.Leaves == nil {
			return m, tea.Batch(
				m.setToast(msg.Name+" completed", modal.ToastSuccess),
				func() tea.Msg { return RefreshMsg{} },
			)
		} else if msg.Err != nil {
			return m, tea.Batch(
				m.setToast(msg.Name+": "+msg.Err.Error(), modal.ToastError),
				func() tea.Msg { return RefreshMsg{} },
			)
		}
		return m, func() tea.Msg { return RefreshMsg{} }

	case DepCheckMsg:
		m.pendingAction = "uninstall"
		m.pendingMutType = msg.MutType
		m.activeModal = modal.NewConfirmModal(msg.Label+" "+msg.Name, msg.Message)
		return m, m.activeModal.Init()

	case TabContentMsg:
		key := tabKey(msg.PanelID, msg.TabIndex, msg.ItemName)
		if msg.Err != nil {
			m.tabContent[key] = "Error: " + msg.Err.Error()
		} else if msg.Content != "" {
			m.tabContent[key] = msg.Content
		} else {
			m.tabContent[key] = "No data"
		}
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
		// Re-derive the per-panel viewport heights once per resize, then
		// clamp cursor/offset per panel here in the Update loop instead
		// of during render (r1 finding 29, r3 finding R3-7).
		m.syncVisibleRows()
		for i := range m.panels {
			m.panels[i].width = sidebarWidth(m.cfg, m.width)
			m.panels[i].height = m.height - 4
			m.panels[i].normalize()
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
		if m.refreshing > 0 {
			m.refreshing--
			if m.refreshing == 0 {
				return m, tea.Batch(
					m.setToast("Data refreshed", modal.ToastSuccess),
					m.loadTabContent(),
				)
			}
		}
		if msg.PanelID == m.activePanel && needsTabFetch(m.activePanel, m.activeTab) {
			return m, m.loadTabContent()
		}
		return m, nil

	case RefreshMsg:
		m.clearTabContent()
		// M12 tiered refresh:
		//   * Status always refreshes via fetchStatusData (never
		//     fetchPanelData — that default-case returns an empty stub).
		//   * Active panel always refreshes (AC-04: R forces it even
		//     if TTL not expired — cache keys invalidated below).
		//   * Other loaded panels refresh; unloaded stay lazy.
		// LoadOutdated is set only by the R key: the status refresh below
		// loads the outdated count only when the user asked for it.
		// Bust cache for force-refreshed classes so per-class TTLs do
		// not short-circuit the shell (AC-04).
		m.invalidatePanelCache(PanelStatus)

		m.panels[PanelStatus].loading = true
		cmds := []tea.Cmd{fetchStatusData(m.client, msg.LoadOutdated)}
		if m.activePanel != PanelStatus && m.activePanel != PanelSearch {
			m.invalidatePanelCache(m.activePanel)
			m.panels[m.activePanel].loading = true
			cmds = append(cmds, fetchPanelData(m.client, m.activePanel))
		}
		for _, p := range m.panels {
			if p.id == PanelStatus || p.id == PanelSearch {
				continue
			}
			if p.id == m.activePanel {
				continue
			}
			if p.items == nil {
				continue // never loaded; honour lazy policy.
			}
			m.invalidatePanelCache(p.id)
			p.loading = true
			cmds = append(cmds, fetchPanelData(m.client, p.id))
		}
		// Count only DataLoadedMsg producers so the "Data refreshed" toast
		// fires when the last panel lands.
		m.refreshing = len(cmds)
		if tick := m.autoRefreshCmd(); tick != nil {
			cmds = append(cmds, tick)
		}
		return m, tea.Batch(cmds...)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		// M12 AC-03: every key press resets the auto-refresh pause
		// window so the next tick defers its work.
		m.lastKeyAt = time.Now()
		m.autoRefreshPaused = false
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

		if m.searchInput.Focused() {
			switch msg.String() {
			case "esc":
				m.searchInput.Blur()
				return m, nil
			case "enter":
				q := strings.TrimSpace(m.searchInput.Value())
				if q != "" {
					m.searchInput.Blur()
					return m, m.executeSearch(q)
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
			if m.searchInput.Focused() {
				m.searchInput.Blur()
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
				// A dropped enqueue would leave isUpdating stuck true and
				// permanently block the R update path (r1 finding 7).
				if _, err := m.tasks.Enqueue(t); err != nil {
					m.isUpdating = false
					return m, tea.Batch(
						m.setToast("Queue full: "+err.Error(), modal.ToastWarning),
						func() tea.Msg { return RefreshMsg{LoadOutdated: true} },
					)
				}
				// The update task and the load-flagged refresh travel
				// together: R is an explicit user request, so the
				// outdated count must be loaded even when R also runs
				// `brew update`. The flag stays on the R path only —
				// setting it in the brew-update completion handler
				// would also load the count at startup (M9).
				return m, tea.Batch(
					m.tasks.RunNext(),
					func() tea.Msg { return RefreshMsg{LoadOutdated: true} },
				)
			}
			if m.cfg.Brew.UpdateOnStart {
				return m, tea.Batch(
					m.updateTickerCmd(),
					func() tea.Msg { return RefreshMsg{LoadOutdated: true} },
				)
			}
			// R is the only path allowed to shell out to `brew outdated`:
			// the dashboard's outdated count is loaded on demand.
			return m, func() tea.Msg { return RefreshMsg{LoadOutdated: true} }

		case "C":
			m.logCollapsed = !m.logCollapsed
			// The toggle changes the main list height (logLayoutHeights)
			// without a resize, so re-clamp the active panel's
			// cursor/offset against the new effective viewport — the
			// smaller of the sidebar and main-list windows, which the
			// shared offset must satisfy (r3 finding R3-4).
			vc := m.panels[m.activePanel].visibleCount()
			_, mainH := m.logLayoutHeights(m.height - 4)
			m.panels[m.activePanel].normalizeToHeight(min(vc, mainH))
			return m, nil

		case "tab":
			cmd := m.nextPanel()
			return m, cmd
		case "shift+tab":
			cmd := m.prevPanel()
			return m, cmd
		case "1":
			cmd := m.switchPanel(PanelStatus)
			return m, cmd
		case "2":
			cmd := m.switchPanel(PanelFormulae)
			return m, cmd
		case "3":
			cmd := m.switchPanel(PanelCasks)
			return m, cmd
		case "4":
			cmd := m.switchPanel(PanelOutdated)
			return m, cmd
		case "5":
			cmd := m.switchPanel(PanelTaps)
			return m, cmd
		case "6":
			cmd := m.switchPanel(PanelServices)
			return m, cmd
		case "7":
			cmd := m.switchPanel(PanelSearch)
			return m, cmd

		case "j", "down":
			m.panels[m.activePanel].down()
			if needsTabFetch(m.activePanel, m.activeTab) {
				return m, m.loadTabContent()
			}
			return m, nil
		case "k", "up":
			m.panels[m.activePanel].up()
			if needsTabFetch(m.activePanel, m.activeTab) {
				return m, m.loadTabContent()
			}
			return m, nil

		case "[":
			return m, m.prevTab()
		case "]":
			return m, m.nextTab()

		case "left":
			return m, m.prevTab()
		case "right":
			return m, m.nextTab()

		case "/":
			m.switchPanel(PanelSearch)
			m.searchInput.Focus()
			return m, m.searchInput.Cursor.BlinkCmd()

		case "enter":
			if m.activePanel == PanelSearch && m.searchInput.Focused() {
				m.searchInput.Blur()
				return m, m.executeSearch(m.searchInput.Value())
			}
			if m.activePanel == PanelSearch && len(m.searchResults) > 0 {
				return m, m.fetchSelectedSearchInfo()
			}

		case "?":
			m.showHelp = !m.showHelp
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
			// Pin is formula-only: `brew pin` has no cask counterpart, and
			// the old cask path read the pinned flag from a misindexed
			// cask (r1 finding 13).
			if m.activePanel == PanelFormulae {
				return m.togglePin()
			}
		}
	}

	if m.activePanel == PanelSearch && !m.searchInput.Focused() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyRunes && len(keyMsg.Runes) > 0 {
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

// invalidatePanelCache drops brew cache entries for a panel so the next
// fetch re-shells even when the per-class TTL has not expired (AC-04 R).
func (m *Model) invalidatePanelCache(id PanelID) {
	if m.client == nil || m.client.Cache == nil {
		return
	}
	c := m.client.Cache
	switch id {
	case PanelFormulae:
		c.Invalidate(brew.KeyFormulaeList)
	case PanelCasks:
		c.Invalidate(brew.KeyCasksList)
	case PanelOutdated:
		c.Invalidate(brew.KeyOutdatedFormulae, brew.KeyOutdatedCasks)
	case PanelTaps:
		c.Invalidate(brew.KeyTapsList)
	case PanelServices:
		c.Invalidate(brew.KeyServicesList)
	case PanelStatus:
		// Status dashboard reads formulae/casks/taps/services/doctor from
		// the cache. It does not own the outdated class: the dashboard
		// reads the outdated entries from the cache, so invalidating them
		// here would throw away a valid 30-minute result. The Outdated
		// panel and the mutation-level invalidations still own them.
		c.Invalidate(
			brew.KeyFormulaeList, brew.KeyCasksList,
			brew.KeyTapsList, brew.KeyServicesList,
			brew.KeyDoctorResult, brew.KeyConfig,
		)
	}
}

// lazyLoadPanel returns fetchPanelData(id) iff the panel is empty, not
// already loading. It is the single source of truth for M9/M12's lazy
// policy: callers that activate any non-Status panel (switchPanel,
// nextPanel, prevPanel) delegate here. Clearing p.err on retry ensures
// a stale error from a previous failed fetch doesn't block the new
// attempt.
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
		m.panels[i].visibleRows = max(0, heights[i]-1)
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
	m.panels[m.activePanel].active = false
	m.activePanel = id
	m.panels[m.activePanel].active = true
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

func needsTabFetch(panelID PanelID, tabIdx int) bool {
	needsFetch := map[PanelID]map[int]bool{
		PanelStatus:   {1: true, 2: true},
		PanelFormulae: {2: true, 3: true, 5: true},
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
