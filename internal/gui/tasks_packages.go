package gui

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

func (m *Model) executeSearch(query string) tea.Cmd {
	m.panels[PanelSearch].loading = true
	m.panels[PanelSearch].err = nil
	// Reset the cursor: a new search with fewer results used to leave a
	// stale out-of-range selection behind (r1 finding 30).
	m.panels[PanelSearch].selected = 0

	return func() tea.Msg {
		results, err := boundedRead(brewSearchTimeout, "search", func(ctx context.Context) ([]brew.SearchResult, error) {
			results, err := m.client.Search.Search(ctx, query)
			if err != nil {
				return nil, err
			}
			return markInstalled(ctx, m.client, results), nil
		})
		if err != nil {
			return SearchDoneMsg{Err: err}
		}
		return SearchDoneMsg{Results: searchItems(results), Raw: results}
	}
}

// searchItems is the name of each row the Search panel shows. The row itself
// is drawn from its SearchResult (searchCells), so the items hold no text.
func searchItems(results []brew.SearchResult) []string {
	items := make([]string, len(results))
	for i, r := range results {
		items[i] = r.Name
	}
	return items
}

// remarkSearch re-reads the installed lists and sets the marks on the rows the
// Search panel shows, without a new search. It runs after a write completes,
// since the write invalidated the lists it reads.
func (m *Model) remarkSearch() tea.Cmd {
	if len(m.searchResults) == 0 {
		return nil
	}
	m.remarkGen++
	gen := m.remarkGen
	rows := slices.Clone(m.searchResults)
	client := m.client
	return func() tea.Msg {
		marked, err := boundedRead(brewReadTimeout, "installed", func(ctx context.Context) ([]brew.SearchResult, error) {
			return markInstalled(ctx, client, rows), nil
		})
		if err != nil {
			return SearchMarksMsg{gen: gen}
		}
		return SearchMarksMsg{Results: marked, gen: gen}
	}
}

// markInstalled sets Installed on the search rows from the installed formula
// and cask lists. A list that cannot be read says nothing about its rows, so
// they keep the marks they have: the rows keep their types and the search
// still succeeds.
func markInstalled(ctx context.Context, client *brew.Client, results []brew.SearchResult) []brew.SearchResult {
	formulae, ferr := client.Formulae.List(ctx)
	casks, cerr := client.Casks.List(ctx)
	if ferr != nil {
		formulae = nil
	}
	if cerr != nil {
		casks = nil
	}
	marked := brew.MarkInstalled(slices.Clone(results), formulae, casks)
	for i, r := range results {
		if (ferr != nil && r.IsFormula) || (cerr != nil && r.IsCask) {
			marked[i].Installed = r.Installed
		}
	}
	return marked
}

// selectedSearchRow is the Search row under the cursor, if there is one.
func (m Model) selectedSearchRow() (brew.SearchResult, bool) {
	i := m.panels[PanelSearch].selected
	if i < 0 || i >= len(m.searchResults) {
		return brew.SearchResult{}, false
	}
	return m.searchResults[i], true
}

func (m Model) fetchSelectedSearchInfo() tea.Cmd {
	if m.activePanel != PanelSearch || m.panels[PanelSearch].selected >= len(m.searchResults) {
		return nil
	}
	r := m.searchResults[m.panels[PanelSearch].selected]
	// A name can be both a formula and a cask, so the row's own kind goes to
	// brew (LB-19). A row of unknown type keeps the no-flag call.
	args := []string{"info", "--json=v2"}
	switch {
	case r.IsCask:
		args = append(args, "--cask")
	case r.IsFormula:
		args = append(args, "--formula")
	}
	args = append(args, r.Name)
	return func() tea.Msg {
		output, err := boundedRead(brewSearchTimeout, "info", func(ctx context.Context) ([]byte, error) {
			return m.client.Runner.Execute(ctx, args...)
		})
		if err != nil {
			return SearchInfoLoadedMsg{Err: searchInfoErr(err)}
		}
		return SearchInfoLoadedMsg{Content: string(output)}
	}
}

// packageWriter is the write side brew has for one kind of package. Formulae
// and casks both have it, so a mutation picks its writer once.
type packageWriter interface {
	Install(ctx context.Context, name string) (<-chan string, <-chan error)
	Uninstall(ctx context.Context, name string) (<-chan string, <-chan error)
	Reinstall(ctx context.Context, name string) (<-chan string, <-chan error)
	Upgrade(ctx context.Context, name string) (<-chan string, <-chan error)
}

// mutTarget is the package a mutation acts on, captured when the action is
// asked. The confirm runs on it whatever the cursor, panel or lists did while
// the dialog (or the dependency check before it) was up (LB-47).
type mutTarget struct {
	mutType mutationType
	// label is the verb the task is titled with.
	label  string
	name   string
	isCask bool
	panel  PanelID
}

// selectedTarget is the target of mutType on the selected row of the active
// panel; ok is false when no row is selected.
func (m Model) selectedTarget(mutType mutationType, label string) (mutTarget, bool) {
	name := m.panels[m.activePanel].selectedName()
	if name == "" {
		return mutTarget{}, false
	}
	// Casks must route to the cask writer whether the action starts from
	// Search or from the Casks panel itself (r2 finding NEW-4). Search
	// results can be casks too (r1 finding 12): the row's own flag decides.
	searchIsCask := false
	if m.activePanel == PanelSearch {
		if r, ok := m.selectedSearchRow(); ok {
			searchIsCask = r.IsCask
		}
	}
	return mutTarget{
		mutType: mutType,
		label:   label,
		name:    name,
		isCask:  m.activePanel == PanelCasks || (m.activePanel == PanelSearch && searchIsCask),
		panel:   m.activePanel,
	}, true
}

func (m Model) doMutation(tg mutTarget) (tea.Model, tea.Cmd) {
	mutType, name := tg.mutType, tg.name
	if name == "" {
		return m, nil
	}

	title := tg.label + " " + name
	pkgName := name
	if mutType == mutUpgradeAll {
		title = "Upgrading all packages"
		pkgName = ""
	}

	// Task IDs must be unique: the batch counter keys on them, and two
	// concurrent tasks sharing a bare package name would cross-talk
	// (r2 finding NEW-3). Prefix every mutation with its op type.
	taskID := mutationID(mutType, name)

	var writer packageWriter = m.client.FormulaeWrite
	if tg.isCask {
		writer = m.client.CasksWrite
	}
	client := m.client

	t := streamTask(taskID, title, func(ctx context.Context) (<-chan string, <-chan error) {
		var ch <-chan string
		var errCh <-chan error

		switch mutType {
		case mutInstall:
			ch, errCh = writer.Install(ctx, pkgName)
		case mutUninstall:
			ch, errCh = writer.Uninstall(ctx, pkgName)
		case mutReinstall:
			ch, errCh = writer.Reinstall(ctx, pkgName)
		case mutUpgrade:
			ch, errCh = writer.Upgrade(ctx, pkgName)
		case mutUpgradeAll:
			ch, errCh = client.FormulaeWrite.Upgrade(ctx, "")
		case mutZap:
			ch, errCh = client.CasksWrite.Zap(ctx, pkgName)
		case mutFetch:
			ch, errCh = client.Runner.ExecuteStream(ctx, "fetch", pkgName)
		}
		return ch, errCh
	})

	c := m.enqueue(t)
	return m, c
}

func (m Model) batchUpgrade() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelOutdated]
	if len(m.batch.selected) == 0 {
		// The selection can drain while the confirm modal is open: a
		// refresh delivered in between clears it (r4 finding R4-2), and
		// confirming then used to end in a silent dead-end after the
		// modal promised N upgrades (r5 finding R5-1).
		c := m.setToast("No selected packages are still in the Outdated list", modal.ToastWarning)
		return m, c
	}

	names := make([]struct {
		name   string
		isCask bool
	}, 0, len(m.batch.selected))

	// The map only ever holds true entries (toggle deletes on deselect),
	// but check the value anyway so a stale false entry can never sneak a
	// package into the batch (r1 finding 1).
	for idx := range m.batch.selected {
		if !m.batch.selected[idx] {
			continue
		}
		if idx >= len(panel.items) {
			continue
		}
		itemName := panel.nameAt(idx)
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

	m.batch.selected = make(map[int]bool)
	m.batchIDs = make(map[string]int, len(names))

	// Batch task IDs carry a per-batch sequence number: a task queued
	// from a superseded batch used to share "batch-upgrade:<name>" with
	// the new batch and decrement its counter early (r4 finding R4-4).
	m.batchSeq++
	batchPrefix := fmt.Sprintf("batch-upgrade:%d:", m.batchSeq)

	// Enqueue errors must surface: silently dropping tasks while still
	// counting them made the progress counter lie about the batch
	// (r1 finding 14).
	enqueued := 0
	queueFull := false
	// started is the first Enqueue's result. Only EnqueueStarted lets this
	// batch call RunNext: behind a running task, or one done but not yet
	// handled (FC2), the batch waits for that task's completion to start it
	// (D1). RunNext here would clear the done task before its completion lands.
	started := false
	client := m.client
	for _, n := range names {
		name := n.name
		isCask := n.isCask

		t := streamTask(batchPrefix+name, "Upgrade "+name, func(ctx context.Context) (<-chan string, <-chan error) {
			if isCask {
				return client.CasksWrite.Upgrade(ctx, name)
			}
			return client.FormulaeWrite.Upgrade(ctx, name)
		})

		res, err := m.tasks.Enqueue(t)
		if err != nil {
			queueFull = true
			break
		}
		if enqueued == 0 {
			started = res == task.EnqueueStarted
		}
		enqueued++
		// Track the batch by the task's (prefixed) ID, counting
		// occurrences: a formula and a cask can share a name, and a bool
		// set collapsed them into one entry, stranding the counter above
		// zero when the second duplicate completed (r3 finding R3-2).
		m.batchIDs[batchPrefix+name]++
	}
	m.batchCount = enqueued
	m.batchTotal = enqueued

	if enqueued == 0 {
		if queueFull {
			// The very first enqueue hit a full queue: without a toast
			// the batch silently did nothing — no feedback, no RunNext
			// (r3 finding R3-1).
			c := m.setToast("Queue full: no batch tasks could be enqueued", modal.ToastWarning)
			return m, c
		}
		// Every selected index fell out of range (a refresh rewrote the
		// Outdated list between selection and confirm): the modal
		// promised N upgrades, so silence would be a dead-end
		// (r4 finding R4-2).
		c := m.setToast("No selected packages are still in the Outdated list", modal.ToastWarning)
		return m, c
	}
	if queueFull {
		// The queued batch tasks are parked waiting for a RunNext: without
		// it the whole batch silently stalls until the user triggers some
		// other mutation (r2 finding NEW-1). Behind a running task, the
		// completion's RunNext starts them (D1).
		c := m.setToast(
			fmt.Sprintf("Queue full: batch limited to %d of %d packages", enqueued, len(names)),
			modal.ToastWarning,
		)
		if !started {
			return m, c
		}
		return m, tea.Batch(
			c,
			m.tasks.RunNext(),
		)
	}

	if !started {
		c := m.setToast("Queued behind running task", modal.ToastInfo)
		return m, c
	}
	return m, m.tasks.RunNext()
}

type pkgInfo struct {
	Name        string
	Version     string
	Type        string
	Bottled     bool
	Installed   bool
	InstallPath string
	License     string
	Description string
	Homepage    string
	// Groups are the dependency lists of the package, each under its heading.
	Groups  []depGroup
	Caveats string
}

// searchInfoErr is the error the search pane shows for a failed `brew info`.
// brew's exit error carries only the exit code, so the error is the "Error:"
// line of its stderr, as brew printed it.
func searchInfoErr(err error) error {
	var exit *brew.BrewExitError
	if !errors.As(err, &exit) {
		return err
	}
	for _, line := range strings.Split(exit.Stderr, "\n") {
		if line = strings.TrimSpace(line); strings.HasPrefix(line, "Error:") {
			return errors.New(line)
		}
	}
	return err
}

// infoParseError is a search info payload that brew sent and that did not
// decode. The pane shows it as a parse error; brew's own errors show as brew
// wrote them.
type infoParseError struct{ err error }

func (e infoParseError) Error() string { return "Parse error: " + e.err.Error() }

// parseSearchInfo decodes the info payload of a search result.
func parseSearchInfo(raw string) (*pkgInfo, error) {
	info, err := parsePackageInfo(raw)
	if err != nil {
		return nil, infoParseError{err}
	}
	return info, nil
}

func parsePackageInfo(rawJSON string) (*pkgInfo, error) {
	formulae, casks, err := brew.ParseInfo([]byte(rawJSON))
	if err != nil {
		return nil, err
	}
	if len(formulae) > 0 {
		f := formulae[0]
		return &pkgInfo{
			Name:    f.Name,
			Version: f.Version,
			Type:    "formula",
			Bottled: f.Bottled,
			// parseFormula sets InstalledOn only from an installed entry, so a
			// set time means installed even when neither install flag is.
			Installed:   !f.InstalledOn.IsZero(),
			InstallPath: f.InstallPath,
			License:     f.License,
			Description: f.Description,
			Homepage:    f.Homepage,
			Groups: []depGroup{
				{heading: "Runtime", names: f.Dependencies, check: depFormula},
				{heading: "Build", names: f.BuildDeps, check: depFormula},
			},
			Caveats: f.Caveats,
		}, nil
	}
	if len(casks) > 0 {
		c := casks[0]
		return &pkgInfo{
			Name:        c.Name,
			Version:     c.Version,
			Type:        "cask",
			Installed:   len(c.Installed) > 0,
			Description: c.Description,
			Homepage:    c.Homepage,
			Groups: []depGroup{
				{heading: "Formulae", names: c.DependsOnFormulae, check: depFormula},
				{heading: "Casks", names: c.DependsOnCasks, check: depCask},
				{heading: "Requirements", names: c.Requirements, check: depNone},
			},
		}, nil
	}
	return nil, fmt.Errorf("no package info found")
}

// togglePin pins/unpins the selected formula. The pinned state is looked
// up by NAME, not by index: the leaves filter changes panel.items without
// touching panel.formulae, so raw indices desync and the decision could
// target the wrong package (r1 finding 13). Casks are not offered pinning
// — `brew pin` is formula-only.
func (m Model) togglePin() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelFormulae]
	name := panel.selectedName()
	if name == "" {
		return m, nil
	}

	pinned := false
	for _, f := range panel.formulae {
		if f.Name == name {
			pinned = f.Pinned
			break
		}
	}

	client := m.client
	t := errTask("pin:"+name, "Toggle pin "+name, func(ctx context.Context) error {
		if pinned {
			return client.FormulaeWrite.Unpin(ctx, name)
		}
		return client.FormulaeWrite.Pin(ctx, name)
	})

	c := m.enqueue(t)
	return m, c
}

func (m Model) confirmUninstall(mutType mutationType) (tea.Model, tea.Cmd) {
	label := "Uninstall"
	if mutType == mutZap {
		label = "Zap"
	}
	target, ok := m.selectedTarget(mutType, label)
	if !ok {
		return m, nil
	}
	name := target.name
	if m.activePanel == PanelSearch {
		r, ok := m.selectedSearchRow()
		if !ok || mutType == mutZap && !r.IsCask {
			return m, nil
		}
		if !r.Installed {
			cmd := m.setToast(name+" is not installed", modal.ToastInfo)
			return m, cmd
		}
	}

	message := "Uninstall " + name + "? This cannot be undone."
	if mutType == mutZap {
		message = "Zap " + name + "? This will remove ALL associated files and preferences."
	}
	// Only formulae have dependents, so a cask's confirm does not ask brew for
	// them (Q4): brew uses has no answer for a cask.
	isCask := target.isCask
	client := m.client

	return m, func() tea.Msg {
		var uses []string
		var err error
		if !isCask {
			uses, err = boundedRead(brewReadTimeout, "uses", func(ctx context.Context) ([]string, error) {
				return client.Formulae.Uses(ctx, name)
			})
		}
		text := message
		if len(uses) > 0 {
			var b strings.Builder
			b.WriteString(text + "\n\nThe following depend on " + name + ":")
			for _, u := range uses {
				b.WriteString("\n  " + u)
			}
			text = b.String()
		}
		// A timed-out dependency check must not look like "nothing depends
		// on this package": the user still gets the confirm dialog, but it
		// says the check did not complete and the list may be incomplete.
		// Any other error keeps today's behaviour — no dependents, no
		// complaint.
		var timeoutErr *brew.TimeoutError
		if errors.As(err, &timeoutErr) {
			text += "\n\nDependency check timed out after " + timeoutErr.Timeout.String() +
				" — the list of packages that depend on " + name + " may be incomplete."
		}
		return DepCheckMsg{
			Target:  target,
			Message: text,
			Kind:    mutationKind(mutType),
			Detail:  removalDetail(name, mutType == mutZap, uses, err),
		}
	}
}

// removalDetail is the subtle line under an uninstall or zap question: what
// goes, and how many installed packages depend on it. A timed-out check says
// the count is unknown rather than zero.
func removalDetail(name string, zap bool, uses []string, err error) string {
	what := "removes " + name
	if zap {
		what = "removes " + name + " and all its files and preferences"
	}
	var timeoutErr *brew.TimeoutError
	switch {
	case errors.As(err, &timeoutErr):
		return what + "; dependents unknown (check timed out)"
	case len(uses) == 1:
		return what + "; 1 package depends on it"
	case len(uses) > 1:
		return fmt.Sprintf("%s; %d packages depend on it", what, len(uses))
	}
	return what
}

// mutationDetail is the consequence line under a package mutation's question.
func mutationDetail(mutType mutationType, name string) string {
	switch mutType {
	case mutInstall:
		return "Installs " + name + " from Homebrew."
	case mutReinstall:
		return "Installs " + name + " again at its current version."
	case mutUpgrade:
		return "Installs the newest version of " + name + "."
	case mutUpgradeAll:
		return "Installs the newest version of every outdated package."
	default:
		return ""
	}
}

func (m Model) confirmMutation(mutType mutationType, label string) (tea.Model, tea.Cmd) {
	target, ok := m.selectedTarget(mutType, label)
	if !ok {
		return m, nil
	}
	name := target.name

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
	m.pendingTarget = target
	modal := modal.NewConfirmModal(mutationKind(mutType), actionLabel, message).
		WithDetail(mutationDetail(mutType, name))
	m.activeModal = modal
	return m, m.activeModal.Init()
}

// mutationKind is the action kind of a package mutation's confirm dialog.
func mutationKind(mutType mutationType) modal.ActionKind {
	switch mutType {
	case mutUninstall, mutZap:
		return modal.ActionDanger
	case mutInstall:
		return modal.ActionInstall
	case mutUpgrade, mutUpgradeAll, mutReinstall:
		return modal.ActionUpgrade
	default:
		return modal.ActionNeutral
	}
}

func (m Model) confirmBatchUpgrade() (tea.Model, tea.Cmd) {
	// Count only actually-selected entries: deselection removes the key,
	// but stay defensive against stale false entries (r1 finding 1).
	n := 0
	for _, sel := range m.batch.selected {
		if sel {
			n++
		}
	}
	message := "Upgrade " + strconv.Itoa(n) + " selected packages?"

	m.pendingAction = "batch-upgrade"
	modal := modal.NewConfirmModal(modal.ActionUpgrade, "Batch Upgrade", message).
		WithDetail("Installs the newest version of each selected package.")
	m.activeModal = modal
	return m, m.activeModal.Init()
}

func (m Model) toggleLeaves() (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		leaves, err := boundedRead(brewReadTimeout, "leaves", func(ctx context.Context) ([]string, error) {
			return m.client.Formulae.Leaves(ctx)
		})
		if err != nil {
			return MutationResultMsg{Name: "leaves", Err: err}
		}
		return MutationResultMsg{Name: "leaves", Leaves: leaves}
	}
}

// packagesReload is the quiet refresh and the Search re-mark after a package
// write fails or is cancelled: it may have changed what is installed (an
// upgrade that fails on one package after upgrading the rest), so the lists
// are read again, with no "Data refreshed" toast over the error (FB4).
func (m *Model) packagesReload(id string) tea.Cmd {
	if !isPackageWrite(id) {
		return nil
	}
	return tea.Batch(
		func() tea.Msg { return RefreshMsg{Quiet: true} },
		m.remarkSearch(),
	)
}

// isPackageWrite is whether a task changes which packages are installed.
func isPackageWrite(id string) bool {
	switch id {
	case "upgrade-all", "cleanup", "autoremove":
		return true
	}
	for _, prefix := range []string{"install:", "uninstall:", "reinstall:", "upgrade:", "zap:", "batch-upgrade:"} {
		if strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}
