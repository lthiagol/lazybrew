package gui

import (
	"context"
	"errors"
	"fmt"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/anim"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

const (
	minTerminalWidth  = 80
	minTerminalHeight = 24
)

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	if m.cfg.Brew.UpdateOnStart == config.UpdateAlways {
		cmds = append(cmds,
			func() tea.Msg { return StartUpdateMsg{} },
			m.updateTickerCmd(),
		)
	} else {
		// M12: Init only fetches the active panel (default = Status,
		// which doubles as the dashboard). With gui.preload the other
		// panels then load in the background once Status has landed
		// (preload.go); otherwise they lazy-load on first switch via
		// lazyLoadPanel().
		if m.activePanel == PanelStatus {
			cmds = append(cmds, launchRead(fetchStatusData(m.client, false)))
		} else {
			cmds = append(cmds, launchRead(fetchPanelData(m.client, m.activePanel)))
		}
		// Auto mode reads when brew last updated while the panel above loads.
		// The read decides whether `brew update` runs (lastUpdateMsg).
		if m.cfg.Brew.UpdateOnStart == config.UpdateAuto {
			cmds = append(cmds, m.lastUpdateCmd())
		}
	}
	// Init starts the one auto-refresh chain, with or without update_on_start.
	if tick := m.autoRefreshCmd(); tick != nil {
		cmds = append(cmds, tick)
	}
	cmds = append(cmds, func() tea.Msg { return m.spinner.Tick() })
	return tea.Batch(cmds...)
}

// launchRead marks the data a launch read returns, so that Update can drop it
// once a stale update has begun (DataLoadedMsg.launch).
func launchRead(read tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		msg := read()
		if data, ok := msg.(DataLoadedMsg); ok {
			data.launch = true
			return data
		}
		return msg
	}
}

// lastUpdateCmd reads when brew last updated. It is the launch check of auto
// mode; its result is a lastUpdateMsg.
func (m Model) lastUpdateCmd() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), brewReadTimeout)
		defer cancel()
		at, err := client.Diagnostics.LastUpdate(ctx)
		return lastUpdateMsg{at: at, err: err}
	}
}

// updateIsStale is the auto-mode test: brew's last update is unknown, unreadable
// or older than brew.update_stale_hours.
func (m Model) updateIsStale(msg lastUpdateMsg) bool {
	if msg.err != nil || msg.at.IsZero() {
		return true
	}
	return time.Since(msg.at) >= time.Duration(m.cfg.Brew.UpdateStaleHours)*time.Hour
}

// updatePending is true while auto mode has not decided whether to update, or
// its update is running. The preload waits for it, so the preload reads the
// metadata the update has just fetched.
func (m Model) updatePending() bool {
	return m.cfg.Brew.UpdateOnStart == config.UpdateAuto && (!m.updateChecked || m.isUpdating)
}

// preloadAfterCheck starts the preload when the launch check has found brew
// fresh, and the Status dashboard landed before the check did. A Status load
// that lands after the check starts the preload itself (DataLoadedMsg).
func (m *Model) preloadAfterCheck() tea.Cmd {
	if !m.cfg.GUI.Preload || m.preload.started || !m.panels[PanelStatus].loaded || m.updatePending() {
		return nil
	}
	return tea.Batch(m.startPreload()...)
}

func (m Model) updateTickerCmd() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return UpdateTickMsg{}
	})
}

// armUpdateTicker starts the next update tick unless one is already in flight.
func (m *Model) armUpdateTicker() tea.Cmd {
	if m.updateTicking {
		return nil
	}
	m.updateTicking = true
	return m.updateTickerCmd()
}

// armAutoRefresh starts the next auto-refresh tick unless one is already in
// flight. Only the tick and paused handlers and Init arm it, so a refresh
// cannot start a second chain.
func (m *Model) armAutoRefresh() tea.Cmd {
	if m.autoRefreshTicking {
		return nil
	}
	tick := m.autoRefreshCmd()
	m.autoRefreshTicking = tick != nil
	return tick
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
	next, cmd := m.animate(msg)
	return followLogs(next, cmd)
}

// followLogs brings the Logs view in step with the model once an update is
// done, and adds the read that step starts to the update's own command.
func followLogs(next tea.Model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch nm := next.(type) {
	case *Model:
		c := nm.syncLogs()
		return nm, tea.Batch(cmd, c)
	case Model:
		c := nm.syncLogs()
		return nm, tea.Batch(cmd, c)
	}
	return next, cmd
}

// animate runs the update with the motion layer around the message switch.
func (m Model) animate(msg tea.Msg) (tea.Model, tea.Cmd) {
	// With the motion off nothing is animated: the update is the message
	// switch, with no copy of the animated values and no motion work.
	if !m.motionOn {
		return m.update(msg)
	}
	// Each update works on its own copy of the animated values, so a model
	// kept from before the update is not moved by it.
	m.motion = m.motion.Clone()
	if tick, ok := msg.(anim.TickMsg); ok {
		return m.stepMotion(tick.At)
	}
	next, cmd := m.update(msg)
	// A handler on *Model returns the pointer, so the settled model goes back
	// into it and the returned type stays the same.
	if p, ok := next.(*Model); ok {
		var nm Model
		nm, cmd = p.settleMotion(cmd)
		*p = nm
		return p, cmd
	}
	return next.(Model).settleMotion(cmd)
}

// update is the message switch. Update wraps it, so the motion targets are
// settled after every message, whichever branch handled it.
func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case modal.ToastTickMsg:
		// Drop the expired toasts; while any are left, re-arm the chain, or
		// the last one would stick.
		m.toasts = liveToasts(m.toasts, time.Now())
		if len(m.toasts) == 0 {
			m.toastTicking = false
			return m, nil
		}
		return m, modal.ToastTick(toastTickInterval)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case autoRefreshTickMsg:
		m.autoRefreshTicking = false
		// Evaluate pause against the *current* model (lastKeyAt may have
		// advanced since the tick was scheduled).
		next := m.evaluateAutoRefresh(msg.at)
		if _, paused := next.(autoRefreshPausedMsg); paused {
			m.autoRefreshPaused = true
		} else {
			m.autoRefreshPaused = false
		}
		// The tick arms the next one, so the chain runs at one tick per
		// interval whatever the refresh does. Armed before m is returned:
		// the order of a call and a variable read in one return is unspecified.
		arm := m.armAutoRefresh()
		return m, tea.Batch(func() tea.Msg { return next }, arm)

	case autoRefreshPausedMsg:
		// Reschedule a full tick so pause can re-apply if the user is
		// still typing when the next interval elapses. The tick that sent
		// this message has usually armed the next one already.
		m.autoRefreshPaused = true
		arm := m.armAutoRefresh()
		return m, arm

	case SearchDoneMsg:
		p := m.panels[PanelSearch]
		p.loading = false
		p.loaded = true
		if msg.Err != nil {
			p.err = msg.Err
		} else {
			p.items = msg.Results
			p.results = msg.Raw
			m.searchResults = msg.Raw
		}
		p.normalize()
		cmd := m.switchPanel(PanelSearch)
		return m, cmd

	case SearchInfoLoadedMsg:
		m.searchInfo, m.searchInfoErr = nil, msg.Err
		if msg.Err == nil && msg.Content != "" {
			m.searchInfo, m.searchInfoErr = parseSearchInfo(msg.Content)
		}
		return m, nil

	case progress.FrameMsg:
		// A frame of the progress bar's own spring: the bar asks for the next
		// one until it is at rest.
		var cmd tea.Cmd
		m.opBar, cmd = m.opBar.Update(msg)
		return m, cmd

	case opAutoCloseMsg:
		// Only the operation the timer names, and only once it is done with no
		// batch still in flight. A newer or running operation is left alone.
		if m.opState != nil && m.opState == msg.op && m.opState.Done() && m.batchCount == 0 {
			m.opState = nil
		}
		return m, nil

	case TaskStartedMsg:
		title := msg.Title
		op := &Operation{
			Icon:    m.actionIcon(msg.ID),
			Status:  opRunning,
			Started: time.Now(),
		}
		// Surface the batch counter in the Operation overlay: batchCount
		// used to be write-only state with no user-visible effect (r3
		// finding R3-5). The title and the bar both carry it.
		if _, ok := m.batchIDs[msg.ID]; ok && m.batchTotal > 0 {
			title = fmt.Sprintf("%s (%d/%d done)", msg.Title, m.batchTotal-m.batchCount, m.batchTotal)
			op.BatchDone = m.batchTotal - m.batchCount
			op.BatchTotal = m.batchTotal
		}
		// Upgrade all is one brew run: its progress is the packages brew
		// has begun upgrading, out of the total in brew's own header.
		if msg.ID == "upgrade-all" {
			op.Upgrades = true
		}
		op.Title = title
		m.opState = op
		return m, m.tasks.RunNext()

	case UpdateTickMsg:
		m.updateTicking = false
		arm := m.armUpdateTicker()
		return m, arm

	case StartUpdateMsg:
		if m.isUpdating || m.cfg.Brew.UpdateOnStart == config.UpdateNever {
			return m, nil
		}
		cmd := m.beginUpdate()
		return m, cmd

	case lastUpdateMsg:
		m.updateChecked = true
		if !m.updateIsStale(msg) {
			m.lastUpdate = msg.at
			cmd := m.preloadAfterCheck()
			return m, cmd
		}
		m.launchStale = true
		if m.isUpdating {
			return m, nil
		}
		cmd := m.beginUpdate()
		return m, cmd

	case TaskOutputMsg:
		if m.isUpdating {
			m.updateOutput = append(m.updateOutput, msg.Lines...)
		}
		if m.opState != nil {
			for _, line := range msg.Lines {
				m.opState.AppendLine(line)
			}
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
			// The update changed brew's metadata, so the outdated count is
			// read again here: it would be stale otherwise.
			cmds := []tea.Cmd{
				m.tasks.RunNext(),
				func() tea.Msg { return RefreshMsg{LoadOutdated: true, ReloadDiagnostics: true} },
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
			m.opState.Ended = time.Now()
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
			c := m.setToast(failureText(msg.ID, msg.Title, msg.Err), modal.ToastError)
			return m, tea.Batch(
				c,
				m.tasks.RunNext(),
				serviceReload(msg.ID),
				tapsReload(msg.ID),
			)
		}
		if errors.Is(msg.Err, context.Canceled) {
			return m, tea.Batch(m.tasks.RunNext(), serviceReload(msg.ID), tapsReload(msg.ID))
		}
		closeCmd := m.autoCloseCmd()
		switch msg.ID {
		case "cleanup-preview":
			m.pendingAction = "cleanup"
			m.activeModal = modal.NewConfirmModal(modal.ActionDanger, "Cleanup",
				"Run brew cleanup?").WithDetail("This will remove old versions.")
			return m, tea.Batch(m.tasks.RunNext(), m.activeModal.Init(), closeCmd)
		case "autoremove-preview":
			m.pendingAction = "autoremove"
			m.activeModal = modal.NewConfirmModal(modal.ActionDanger, "Autoremove",
				"Remove orphaned dependencies?").
				WithDetail("Removes dependencies that no installed package needs any more.")
			return m, tea.Batch(m.tasks.RunNext(), m.activeModal.Init(), closeCmd)
		}
		var toastTick tea.Cmd
		if msg.Title != "" {
			toastTick = m.setToast(msg.Title+" completed", modal.ToastSuccess)
		}
		cmds := []tea.Cmd{
			m.tasks.RunNext(),
			func() tea.Msg { return RefreshMsg{} },
			closeCmd,
		}
		if toastTick != nil {
			cmds = append(cmds, toastTick)
		}
		return m, tea.Batch(cmds...)

	case MutationResultMsg:
		if msg.Leaves != nil && m.activePanel == PanelFormulae {
			m.panels[PanelFormulae].toggleLeaves(msg.Leaves)
		}
		m.activeModal = nil
		// A leaves result is not a finished mutation, so it has no toast.
		var toast tea.Cmd
		switch {
		case msg.Err != nil:
			toast = m.setToast(msg.Name+": "+msg.Err.Error(), modal.ToastError)
		case msg.Leaves == nil:
			toast = m.setToast(msg.Name+" completed", modal.ToastSuccess)
		}
		return m, tea.Batch(toast, func() tea.Msg { return RefreshMsg{} })

	case DepCheckMsg:
		m.pendingAction = "uninstall"
		m.pendingMutType = msg.MutType
		m.pendingLabel = msg.Label
		m.activeModal = modal.NewConfirmModal(msg.Kind, msg.Label, msg.Message).WithDetail(msg.Detail)
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
		clear(m.wrap.entries)
		m.terminalTooSmall = msg.Width < minTerminalWidth || msg.Height < minTerminalHeight
		m.ready = true
		// Re-derive the per-panel viewport heights once per resize, then
		// clamp cursor/offset per panel here in the Update loop instead
		// of during render (r1 finding 29, r3 finding R3-7).
		m.syncVisibleRows()
		for i := range m.panels {
			m.panels[i].height = m.height - 4
			m.panels[i].normalize()
		}
		return m, nil

	case DataLoadedMsg:
		// A launch read that lands after a stale update began holds the data the
		// update replaces. The update's refresh reads the panels again, so the
		// stale read is dropped rather than applied over it.
		if msg.launch && m.launchStale {
			return m, nil
		}
		cmd := m.landData(msg, false)
		if msg.PanelID != PanelStatus {
			m.absorbLanded(msg)
		}
		// The first Status load starts the preload. With update_on_start that
		// is the load after the brew update, so the preload reads fresh data.
		// In auto mode it waits for the launch check (updatePending).
		if msg.PanelID == PanelStatus && m.cfg.GUI.Preload && !m.preload.started && !m.updatePending() {
			cmd = tea.Batch(cmd, tea.Batch(m.startPreload()...))
		}
		return m, cmd

	case preloadedMsg:
		m.preload.inFlight--
		cmd := m.landData(DataLoadedMsg(msg), true)
		m.absorbLanded(DataLoadedMsg(msg))
		c := tea.Batch(cmd, tea.Batch(m.nextPreloadReads()...))
		return m, c

	case logsReadMsg:
		// A read for a view this model no longer has is stale, and it ends
		// the loop: nothing is applied and no next read is scheduled.
		if msg.gen != m.logs.gen {
			return m, nil
		}
		if !m.logsShown() {
			m.logs.started = false
			return m, nil
		}
		m.applyLogs(msg)
		return m, logsTick(msg.gen)

	case logsTickMsg:
		// The loop runs only while its view is on screen. A stale tick is
		// dropped: a newer view has its own loop.
		if msg.gen != m.logs.gen {
			return m, nil
		}
		if !m.logsShown() {
			m.logs.started = false
			return m, nil
		}
		c := m.readLogsCmd()
		return m, c

	case servicesReloadMsg:
		m.invalidatePanelCache(PanelServices)
		m.panels[PanelServices].loading = true
		return m, fetchPanelData(m.client, PanelServices)

	case tapsReloadMsg:
		m.invalidatePanelCache(PanelTaps)
		m.panels[PanelTaps].loading = true
		return m, fetchPanelData(m.client, PanelTaps)

	case RefreshMsg:
		m.clearTabContent()
		// M12 tiered refresh:
		//   * Status always refreshes via fetchStatusData (never
		//     fetchPanelData — that default-case returns an empty stub).
		//   * Active panel always refreshes (AC-04: R forces it even
		//     if TTL not expired — cache keys invalidated below).
		//   * Other loaded panels refresh; unloaded stay lazy.
		// LoadOutdated is set by the R key and by the end of a brew update:
		// the status refresh below loads the outdated count only then.
		// Bust cache for force-refreshed classes so per-class TTLs do
		// not short-circuit the shell (AC-04).
		m.invalidatePanelCache(PanelStatus)
		if msg.ReloadDiagnostics {
			m.invalidateCacheGroup("diagnostics")
		}

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
		return m, tea.Batch(cmds...)

	case tea.PasteMsg:
		// v1 delivered a paste as a KeyMsg, so it reached the active modal,
		// the focused search box, or started the search box from its panel.
		// v2 sends a PasteMsg instead; route it the same way.
		m.lastKeyAt = time.Now()
		m.autoRefreshPaused = false
		if m.activeModal != nil {
			updated, cmd := m.activeModal.Update(msg)
			next, ok := updated.(modal.Modal)
			if !ok {
				m.activeModal = nil
				return m, cmd
			}
			m.activeModal = next
			return m, cmd
		}
		if m.activePanel == PanelSearch && !m.searchInput.Focused() {
			m.searchInput.Focus()
		}
		if m.searchInput.Focused() {
			m.searchInput, cmd = m.searchInput.Update(msg)
		}
		return m, cmd

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyPressMsg:
		// M12 AC-03: every key press resets the auto-refresh pause
		// window so the next tick defers its work.
		m.lastKeyAt = time.Now()
		m.autoRefreshPaused = false
		return m.handleKey(msg)
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

// View declares the terminal setup the program needs — the alternate screen
// and, when the GUI mouse option is on, cell-motion mouse reporting — and
// wraps the rendered screen. These used to be program options in main.go.
func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.WindowTitle = "lazybrew"
	// The app owns its canvas: the theme's base and text, as change-gate does.
	v.BackgroundColor = style.BgColor
	v.ForegroundColor = style.TextColor
	if m.cfg != nil && m.cfg.GUI.Mouse {
		v.MouseMode = tea.MouseModeCellMotion
	}
	return v
}

func (m Model) render() string {
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
		return m.renderHelpScreen()
	}

	l := m.screenLayout()
	sidebar := m.renderSidebar()
	mainContent := m.renderMainPanelIn(l)
	bottomBar := m.renderBottomBar()

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
	base := fitFrame(lipgloss.JoinVertical(lipgloss.Top, m.renderHeader(), body, bottomBar), m.width, m.height)

	// Overlays float over the finished frame, in stacking order: the toast
	// on the last body row, then the modal centred above it. Neither adds
	// a row, so the screen behind them keeps its size and position. While a
	// modal is open the frame behind it is muted first (dimBackdrop).
	var overlays []*lipgloss.Layer
	overlays = append(overlays, m.toastLayers()...)
	if m.activeModal != nil {
		base = dimBackdrop(base)
		overlays = append(overlays, centred(m.activeModal.View().Content, m.width, m.height, m.modalDropAt()))
	}
	if len(overlays) == 0 {
		return base
	}
	return compose(base, m.width, m.height, overlays...)
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
	if w > 60 {
		w = 60
	}
	return w
}

// searchStyles draws the search input from the theme roles: the prompt and the
// cursor in the accent, typed text in the text role, the placeholder and
// suggestion muted. The bubbles defaults are a fixed palette, which the app's
// own canvas leaves unreadable on a light theme.
func searchStyles() textinput.Styles {
	focused := textinput.StyleState{
		Text:        style.TextStyle,
		Placeholder: style.MutedText,
		Suggestion:  style.MutedText,
		Prompt:      lipgloss.NewStyle().Foreground(style.AccentColor),
	}
	blurred := focused
	blurred.Text = style.SubtleText
	return textinput.Styles{
		Focused: focused,
		Blurred: blurred,
		Cursor:  textinput.CursorStyle{Color: style.AccentColor, Shape: tea.CursorBlock, Blink: true},
	}
}
