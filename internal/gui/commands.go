package gui

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/style"
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
			// The Taps panel opens the same input modal to add a tap; a
			// bare search has no pendingAction. Routing every InputResult
			// to executeSearch used to make `a` on Taps run a search
			// instead of `brew tap` (r1 finding 2).
			if m.pendingAction == "tap-add" {
				m.pendingAction = ""
				return m.executeTapAdd(r.Value)
			}
			c := m.executeSearch(r.Value)
			return m, c
		}
		m.pendingAction = ""
	case *modal.ConfirmResult:
		if r.Confirmed {
			switch m.pendingAction {
			case "uninstall", "mutation":
				return m.doMutation(m.pendingTarget)
			case "batch-upgrade":
				return m.batchUpgrade()
			case "untap":
				return m.executeUntap()
			case "untrust":
				m.pendingAction = ""
				cmd := m.executeTrustAction(1)
				return *m, cmd
			case "untrust-item":
				m.pendingAction = ""
				cmd := m.executeUntrustItem()
				return *m, cmd
			case "repair":
				return m.executeRepair()
			case "cleanup":
				m.pendingAction = ""
				return m.executeCleanup()
			case "autoremove":
				m.pendingAction = ""
				return m.executeAutoremove()
			case "services-cleanup":
				m.pendingAction = ""
				return m.executeServicesCleanup()
			case "service-stop":
				m.pendingAction = ""
				return m.runServiceTask("stop", m.pendingService)
			case "service-restart":
				m.pendingAction = ""
				return m.runServiceTask("restart", m.pendingService)
			}
		}
		m.pendingAction = ""
	case *modal.MenuResult:
		// pendingAction is consumed here whichever way the menu closes:
		// a CANCELLED menu used to leave it set, so the next unrelated
		// menu could route its pick into a trust action (r4 finding R4-1).
		pending := m.pendingAction
		m.pendingAction = ""
		if !r.Cancelled {
			switch pending {
			case "trust-menu":
				c := m.executeTrustMenuAction(r.SelectedIndex)
				return m, c
			case "trust-formula":
				c := m.executeTrustFormulaAction(r.SelectedIndex)
				return m, c
			case "trust-cask":
				c := m.executeTrustCaskAction(r.SelectedIndex)
				return m, c
			case "brewfile":
				c := m.executeBrewfileAction(r.SelectedIndex)
				return m, c
			}
		}
	}
	return m, nil
}

func closedCh() <-chan string {
	ch := make(chan string)
	close(ch)
	return ch
}

// failureText is the error toast of a failed task: its title and the error. A
// services write that needs sudo says so in its own words, which already name
// the service and the command to run (AC-04.1). A taps write says brew's first
// meaningful line (v2-10 AC-03.2).
func failureText(id, title string, err error) string {
	var needsRoot *brew.ServiceNeedsRootError
	if errors.As(err, &needsRoot) {
		return err.Error()
	}
	if isTapTask(id) {
		return tapFailureText(title, err)
	}
	return title + ": " + err.Error()
}

// brewMessage is the line of brew's output that says what went wrong: the first
// "Error:" line, else the first line that is not blank, without the prefix.
func brewMessage(stderr string) string {
	lines := strings.Split(strings.TrimSpace(stderr), "\n")
	for _, line := range lines {
		if msg, ok := strings.CutPrefix(strings.TrimSpace(line), "Error: "); ok {
			return msg
		}
	}
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}

// tabItemName is the name a tab's content is keyed by, or "" when the tab has
// nothing to fetch for yet. The Status tabs read the whole dashboard, so they
// key by its name whatever row is selected.
func tabItemName(p *panelData) string {
	if p.id == PanelStatus {
		if !p.loaded {
			return ""
		}
		return statusItemName
	}
	return p.selectedName()
}

func (m *Model) loadTabContent() tea.Cmd {
	panelID := m.activePanel
	tabIdx := m.activeTab

	if isPreviewTab(panelID, tabIdx) {
		return m.loadPreviewTab()
	}
	if needsTabFetch(panelID, tabIdx) {
		panel := m.panels[panelID]
		name := tabItemName(panel)
		if name == "" {
			return nil
		}
		key := tabKey(panelID, tabIdx, name)
		if _, ok := m.tabContent[key]; ok || m.tabFetching[key] {
			return nil
		}
		m.tabFetching[key] = true
		return fetchTabContentCmd(m.client, panelID, tabIdx, name, m.icons)
	}
	return nil
}

func fetchTabContentCmd(client *brew.Client, panel PanelID, tab int, name string, icons style.Icons) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTabContentTimeout)
		defer cancel()
		// wrapErr maps a deadline overrun to *brew.TimeoutError like every
		// other bounded read, so the tab body shows the timeout design's
		// error instead of a raw "signal: killed" (r1 finding 17).
		wrapErr := func(command string, err error) error {
			return timeoutError(ctx, command, fetchTabContentTimeout, err)
		}
		switch panel {
		case PanelStatus:
			if tab == statusConfigTab {
				cfg, err := client.Diagnostics.Config(ctx)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Err: wrapErr("config", err)}
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name,
					Content: formatConfig(cfg)}
			}
			if tab == statusDoctorTab {
				warnings, err := client.Diagnostics.Doctor(ctx)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Err: wrapErr("doctor", err)}
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name,
					Content: doctorView(warnings, icons)}
			}
			if tab == statusMissingTab {
				deps, err := client.Diagnostics.Missing(ctx)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Err: wrapErr("missing", err)}
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name,
					Content: missingView(deps, icons)}
			}
		case PanelFormulae:
			switch tab {
			case formulaUsesTab:
				uses, err := client.Formulae.Uses(ctx, name)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Err: wrapErr("uses", err)}
				}
				var result strings.Builder
				for _, u := range uses {
					result.WriteString(u + "\n")
				}
				content := result.String()
				if content == "" {
					content = noDependents
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Content: content}
			case formulaFilesTab:
				output, err := client.Runner.Execute(ctx, "list", name)
				if err != nil {
					return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Err: wrapErr("list", err)}
				}
				content := string(output)
				if content == "" {
					content = noFilesInstalled
				}
				return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Content: content}
			}
		}
		return TabContentMsg{PanelID: panel, TabIndex: tab, ItemName: name, Content: ""}
	}
}

func formatConfig(cfg *brew.BrewConfig) string {
	if cfg == nil {
		return "No config available"
	}
	// Built uncut: the Config tab cuts each row to the main width when drawn.
	return presentation.Detail([]presentation.Field{
		{Label: "Homebrew", Value: cfg.HomebrewVersion},
		{Label: "Prefix", Value: cfg.Prefix},
		{Label: "Cellar", Value: cfg.Cellar},
		{Label: "Repository", Value: cfg.Repository},
		{Label: "Core Tap", Value: cfg.CoreTap},
		{Label: "OS", Value: cfg.OS},
	}, 0)
}

// runConcurrently runs each read in its own goroutine and returns when they
// have all finished. The reads only fill in the caller's local variables, as
// in statusDashboard; no Bubble Tea state is touched from them.
func runConcurrently(reads ...func()) {
	var wg sync.WaitGroup
	for _, read := range reads {
		wg.Add(1)
		go func() {
			defer wg.Done()
			read()
		}()
	}
	wg.Wait()
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
				items[i] = f.Name
			}
			return DataLoadedMsg{PanelID: panel, Items: items, Formulae: formulae}

		case PanelCasks:
			casks, err := boundedRead(brewReadTimeout, "list", client.Casks.List)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			items := make([]string, len(casks))
			for i, c := range casks {
				items[i] = c.Name
			}
			return DataLoadedMsg{PanelID: panel, Items: items, Casks: casks}

		case PanelOutdated:
			// Each outdated read gets its own deadline: one slow source
			// must not eat the whole budget of the other. The two reads run
			// at the same time.
			var (
				formulae []brew.Formula
				casks    []brew.Cask
				fErr     error
				cErr     error
			)
			runConcurrently(
				func() { formulae, fErr = boundedRead(brewReadTimeout, "outdated", client.Formulae.Outdated) },
				func() { casks, cErr = boundedRead(brewReadTimeout, "outdated", client.Casks.Outdated) },
			)
			if fErr != nil {
				return DataLoadedMsg{PanelID: panel, Err: fErr}
			}
			if cErr != nil {
				return DataLoadedMsg{PanelID: panel, Err: cErr}
			}
			items := make([]string, 0, len(formulae)+len(casks))
			for _, f := range formulae {
				items = append(items, f.Name)
			}
			for _, c := range casks {
				items = append(items, c.Name)
			}
			return DataLoadedMsg{PanelID: panel, Items: items, Formulae: formulae, Casks: casks}

		case PanelTaps:
			// The trust list is its own read: a failure there does not fail the
			// taps, it only leaves the trusted rows out (lb30 AC-02).
			var (
				list     brew.TapList
				err      error
				trusted  []brew.TrustEntry
				trustErr error
			)
			runConcurrently(
				func() { list, err = boundedRead(brewReadTimeout, "taps", client.Taps.List) },
				func() { trusted, trustErr = boundedRead(brewReadTimeout, "trust", client.Trust.ListTrusted) },
			)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			return DataLoadedMsg{
				PanelID:      panel,
				Taps:         list.Taps,
				TrustUnknown: list.TrustUnknown,
				Trusted:      trustedRows(trusted),
				TrustErr:     trustErr,
			}

		case PanelServices:
			var (
				services []brew.Service
				info     []brew.ServiceInfo
				err      error
			)
			runConcurrently(
				func() { services, err = boundedRead(brewReadTimeout, "services", client.Services.List) },
				// A failed info read is not a failed panel: the Info view
				// falls back to the list's fields (AC-02.3), so its error is
				// dropped.
				func() { info, _ = boundedRead(brewReadTimeout, "services", client.Services.Info) },
			)
			if err != nil {
				return DataLoadedMsg{PanelID: panel, Err: err}
			}
			items := make([]string, len(services))
			for i, s := range services {
				items[i] = s.Name
			}
			return DataLoadedMsg{PanelID: panel, Items: items, Services: services, ServiceInfo: info}

		default:
			// PanelStatus and PanelSearch are not fetched here. Init and
			// RefreshMsg use fetchStatusData for Status; Search is
			// user-driven. Do not route Status through this default —
			// an empty Items payload would wipe the dashboard.
			return DataLoadedMsg{PanelID: panel, Items: []string{}}
		}
	}
}

// fetchStatusData renders the status dashboard. loadOutdated is false for
// every automatic refresh (Init, the auto-refresh tick, install/uninstall,
// task completion): those read the outdated entries from the cache only and
// never shell out to `brew outdated`. Only an explicit user request (the R
// key) passes true.
func fetchStatusData(client *brew.Client, loadOutdated bool) tea.Cmd {
	return func() tea.Msg {
		summary := statusDashboard(client, loadOutdated)
		return DataLoadedMsg{PanelID: PanelStatus, Items: statusRows(summary), Summary: summary, Err: nil}
	}
}

// fetchDoctor is the Health card's brew doctor read, capped at
// brewReadTimeout like each dashboard read. gen is the generation Update
// issued it under (Model.doctorGen).
func fetchDoctor(client *brew.Client, gen int) tea.Cmd {
	return func() tea.Msg {
		warnings, err := boundedReadIn(context.Background(), brewReadTimeout, "doctor", client.Diagnostics.Doctor)
		return doctorLoadedMsg{gen: gen, warnings: len(warnings), err: err}
	}
}

// statusDashboard reads every source the status dashboard aggregates and
// returns the typed summary the cards and the sidebar are drawn from.
//
// The dashboard aggregates five independent sources. brew doctor is not one
// of them: it is slow, so it is its own read (fetchDoctor) and the dashboard
// lands without waiting for it. Each
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
//
// The outdated count has two modes. The default (loadOutdated=false) reads
// the outdated entries from the client cache only, so the dashboard never
// shells out to `brew outdated` on its own; the count is known only when
// both outdated cache entries are hits, because a partial hit would be an
// incomplete count. loadOutdated=true is the explicit user request (the R
// key): the outdated entries are then read through the brew reader — bounded
// like every other read and part of the same concurrent fan-out — and the
// count becomes known. A failed or timed-out outdated read joins the ⚠
// lines like any other source and the count falls back to "not known", never
// to a misleading 0.
func statusDashboard(client *brew.Client, loadOutdated bool) statusSummary {
	totalCtx, cancelTotal := context.WithTimeout(context.Background(), brewStatusTotalTimeout)
	defer cancelTotal()

	var (
		formulae []brew.Formula
		casks    []brew.Cask
		taps     brew.TapList
		services []brew.Service
		cfg      *brew.BrewConfig

		fErr, cErr, tErr, sErr error

		outdatedFormulae []brew.Formula
		outdatedCasks    []brew.Cask
		outdatedFErr     error
		outdatedCErr     error
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
	if loadOutdated {
		// Explicit request (R): the outdated entries are read through the
		// brew reader, which serves a still-fresh cached result and only
		// shells out when the data is stale.
		read(func() {
			outdatedFormulae, outdatedFErr = boundedReadIn(totalCtx, brewReadTimeout, "outdated", client.Formulae.Outdated)
		})
		read(func() {
			outdatedCasks, outdatedCErr = boundedReadIn(totalCtx, brewReadTimeout, "outdated", client.Casks.Outdated)
		})
	}
	wg.Wait()

	// Cache-only mode (the default): the outdated entries come from the
	// client cache, which the Outdated panel fills. The count is known only
	// when both entries are hits — a partial hit would be an incomplete
	// count. A cache hit holding an empty slice is a known zero: the reader
	// caches an empty result when `brew outdated` exits 1.
	outdatedKnown := false
	if loadOutdated {
		outdatedKnown = outdatedFErr == nil && outdatedCErr == nil
	} else {
		formulaeHit := false
		if cached, ok := client.Cache.Get(brew.KeyOutdatedFormulae); ok {
			if f, ok := cached.([]brew.Formula); ok {
				outdatedFormulae = f
				formulaeHit = true
			}
		}
		casksHit := false
		if cached, ok := client.Cache.Get(brew.KeyOutdatedCasks); ok {
			if c, ok := cached.([]brew.Cask); ok {
				outdatedCasks = c
				casksHit = true
			}
		}
		outdatedKnown = formulaeHit && casksHit
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
	if outdatedFErr != nil {
		sources = append(sources, sourceErr{"outdated", outdatedFErr})
	}
	if outdatedCErr != nil {
		sources = append(sources, sourceErr{"outdated", outdatedCErr})
	}

	summary := statusSummary{
		Formulae: knownCount{N: len(formulae), Known: fErr == nil},
		Casks:    knownCount{N: len(casks), Known: cErr == nil},
		Taps:     knownCount{N: len(taps.Taps), Known: tErr == nil},
		Services: knownCount{N: len(services), Known: sErr == nil},
		Outdated: knownCount{N: len(outdatedFormulae) + len(outdatedCasks), Known: outdatedKnown},
	}
	if cfg != nil {
		summary.Version = cfg.HomebrewVersion
		summary.Prefix = cfg.Prefix
	}

	officialTaps := 0
	for _, t := range taps.Taps {
		if t.IsOfficial {
			officialTaps++
		}
	}

	servicesStarted := 0
	for _, s := range services {
		if s.Status == brew.ServiceStarted {
			servicesStarted++
		}
	}
	summary.OfficialTaps = officialTaps
	summary.ServicesStarted = servicesStarted

	// The failed sources are ⚠ lines.
	for _, s := range sources {
		summary.Problems = append(summary.Problems, "⚠ "+s.label+": "+s.err.Error())
	}

	return summary
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
