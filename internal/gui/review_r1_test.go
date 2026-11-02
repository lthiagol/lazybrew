package gui

// Tests for the r1 code-review round: each test names the finding it
// locks in a comment.

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// --- Finding 1: batch toggle-off -----------------------------------------

func TestBatchToggleOffRemovesKey(t *testing.T) {
	b := newBatchState()
	b.toggle(0)
	b.toggle(1)
	b.toggle(1) // deselect bar

	if len(b.selected) != 1 {
		t.Fatalf("expected 1 selected after toggle-off, got %d entries", len(b.selected))
	}
	if _, ok := b.selected[1]; ok {
		t.Error("deselected index must be removed from the map, not stored as false")
	}
}

func TestBatchToggleOffThenUpgradeSkipsDeselected(t *testing.T) {
	m := newTestModel()
	panel := m.panels[PanelOutdated]
	panel.items = []string{"foo  1.0", "bar  2.0"}
	panel.formulae = []brew.Formula{{Name: "foo"}, {Name: "bar"}}

	m.batch.toggle(0)
	m.batch.toggle(1)
	m.batch.toggle(1) // space-deselect bar

	// The confirm message must count only actually-selected packages.
	nm, _ := m.confirmBatchUpgrade()
	mm := nm.(Model)
	// The modal needs a WindowSizeMsg before it renders (see confirm.go).
	mm.activeModal.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if view := mm.activeModal.View(); !strings.Contains(view, "1 selected") {
		t.Fatalf("confirm message must count 1 selected package, got view:\n%s", view)
	}

	nm2, _ := mm.batchUpgrade()
	rm := nm2.(Model)
	if rm.batchCount != 1 {
		t.Fatalf("expected batchCount 1 after deselect, got %d", rm.batchCount)
	}
}

// --- Finding 2: add-tap flow ---------------------------------------------

func TestTapAddFlowRunsBrewTap(t *testing.T) {
	var gotArgs [][]string
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		gotArgs = append(gotArgs, args)
		return []byte(""), nil
	}
	m := New(brew.NewClient(r), config.Default())
	m.switchPanel(PanelTaps)
	m.panels[PanelTaps].items = []string{"homebrew/core"}

	tm, _ := m.startTapAdd()
	tmm := tm.(*Model)
	if tmm.pendingAction != "tap-add" {
		t.Fatalf("expected pendingAction 'tap-add', got %q", tmm.pendingAction)
	}
	if tmm.activeModal == nil {
		t.Fatal("expected input modal for tap add")
	}

	// Mirror the Update flow: modal is cleared before the result dispatch.
	tmm.activeModal = nil
	res, cmd := tmm.handleModalResult(&modal.InputResult{Value: "user/repo"}, nil)
	nm := res.(Model)
	if cmd == nil {
		t.Fatal("expected a RunNext cmd after tap-add input")
	}
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatalf("expected TaskStartedMsg, got %T", cmd())
	}
	msg := nm.tasks.RunNext()()
	comp, ok := msg.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}
	if comp.Err != nil {
		t.Fatalf("unexpected task error: %v", comp.Err)
	}

	found := false
	for _, args := range gotArgs {
		if len(args) == 2 && args[0] == "tap" && args[1] == "user/repo" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected `brew tap user/repo` to be executed, got %v", gotArgs)
	}
}

func TestTapAddInputWithoutPendingActionStillSearches(t *testing.T) {
	m := newTestModel()
	m.panels[PanelSearch].loading = false
	res, _ := m.handleModalResult(&modal.InputResult{Value: "ripgrep"}, nil)
	nm := res.(*Model)
	if !nm.panels[PanelSearch].loading {
		t.Error("bare input result must still route to the search flow")
	}
}

// --- Finding 4: toast auto-dismiss ---------------------------------------

func TestSetToastArmsTickCmd(t *testing.T) {
	m := newTestModel()
	cmd := m.setToast("hello", modal.ToastInfo)
	if cmd == nil {
		t.Fatal("setToast must arm a tick cmd, otherwise the toast never dismisses")
	}
	if m.toast == nil || m.toast.Message != "hello" {
		t.Fatalf("expected toast to be set, got %+v", m.toast)
	}
}

func TestToastTickChainKeepsRearmingUntilExpired(t *testing.T) {
	m := newTestModel()
	m.setToast("hello", modal.ToastInfo)

	// First tick: toast not yet expired → the handler must re-arm the
	// chain, otherwise the toast would stick around forever (the only
	// producer of ToastTickMsg is the chain itself).
	nm, tickCmd := m.Update(modal.ToastTickMsg{})
	mm := nm.(Model)
	if mm.toast == nil {
		t.Fatal("unexpired toast must survive the first tick")
	}
	if tickCmd == nil {
		t.Fatal("ToastTickMsg handler must re-arm the tick while the toast lives")
	}
}

// --- Finding 5: services cleanup through the task manager ----------------

func TestServicesCleanupTaskIsSerialized(t *testing.T) {
	var gotArgs [][]string
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		gotArgs = append(gotArgs, args)
		return []byte(""), nil
	}
	m := New(brew.NewClient(r), config.Default())
	m.switchPanel(PanelServices)
	m.panels[PanelServices].items = []string{"redis"}

	tm, _ := m.serviceCleanup()
	mm := tm.(Model)
	mm.activeModal = nil
	res, cmd := mm.handleModalResult(&modal.ConfirmResult{Confirmed: true}, nil)
	nm := res.(Model)
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatal("expected the cleanup to be enqueued via the task manager")
	}
	msg := nm.tasks.RunNext()()
	if comp, ok := msg.(TaskCompletedMsg); !ok || comp.Err != nil {
		t.Fatalf("expected clean completion, got %T (%v)", msg, comp)
	}
	found := false
	for _, args := range gotArgs {
		if len(args) == 2 && args[0] == "services" && args[1] == "cleanup" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected `brew services cleanup`, got %v", gotArgs)
	}
}

// --- Finding 6: queued enqueue outcome -----------------------------------

func TestEnqueueQueuedOutcomeToastsQueueing(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	m.panels[PanelFormulae].items = []string{"foo  1.0"}
	m.panels[PanelFormulae].selected = 0

	_, cmd := m.doMutation(mutInstall, "Install")
	_ = cmd() // start the first task

	nm, _ := m.doMutation(mutUpgrade, "Upgrade")
	mm := nm.(Model)
	if mm.toast == nil || !strings.Contains(mm.toast.Message, "Queued") {
		t.Fatalf("expected 'queued' toast, got %+v", mm.toast)
	}
}

// --- Finding 7: brew-update enqueue failure ------------------------------

func TestBrewUpdateEnqueueFailureResetsIsUpdating(t *testing.T) {
	m := newTestModel()
	m.cfg.Brew.UpdateOnStart = true

	// Fill the queue to force ErrQueueFull.
	for i := 0; i < task.DefaultMaxQueue; i++ {
		_, _ = m.tasks.Enqueue(&task.Task{
			ID:  "filler",
			Run: func(ctx context.Context) (<-chan string, <-chan error, error) { return nil, nil, nil },
		})
		// Start the first, queue the rest.
		if i == 0 {
			m.tasks.RunNext()()
		}
	}

	nm := updateModel(m, StartUpdateMsg{})
	if nm.isUpdating {
		t.Fatal("isUpdating must be reset when the brew-update task is dropped")
	}
	if nm.toast == nil {
		t.Fatal("expected a warning toast when the update cannot be queued")
	}
}

// --- Finding 11: Outdated info tab cask indexing --------------------------

func TestOutdatedInfoTabShowsCorrectCask(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelOutdated
	m.activeTab = 0
	p := m.panels[PanelOutdated]
	p.items = []string{"foo  1.0", "bar  2.0"}
	p.formulae = []brew.Formula{{Name: "foo", Version: "1.0", Description: "formula desc"}}
	p.casks = []brew.Cask{{Name: "bar", Version: "2.0", Description: "cask desc"}}
	p.selected = 1 // cask row

	out := m.renderContent(120, 40)
	if !strings.Contains(out, "bar") {
		t.Fatalf("expected the cask's info for index 1, got:\n%s", out)
	}
	if strings.Contains(out, "formula desc") {
		t.Errorf("formula info leaked into the cask row view:\n%s", out)
	}
}

// --- Finding 12: install from Search respects the result type -------------

func TestSearchInstallUsesCaskWriterForCasks(t *testing.T) {
	var gotArgs [][]string
	r := brew.NewMockRunner()
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		gotArgs = append(gotArgs, args)
		ch := make(chan string)
		close(ch)
		errCh := make(chan error, 1)
		close(errCh)
		return ch, errCh
	}
	m := New(brew.NewClient(r), config.Default())
	m.activePanel = PanelSearch
	m.searchResults = []brew.SearchResult{{Name: "firefox", IsCask: true}}
	p := m.panels[PanelSearch]
	p.items = []string{"firefox  cask"}
	p.selected = 0

	nm, _ := m.confirmMutation(mutInstall, "Install")
	mm := nm.(Model)
	mm.activeModal = nil
	res, cmd := mm.handleModalResult(&modal.ConfirmResult{Confirmed: true}, nil)
	nm2 := res.(Model)
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatal("expected the install to be enqueued")
	}
	if msg := nm2.tasks.RunNext()(); msg.(TaskCompletedMsg).Err != nil {
		t.Fatal("unexpected task error")
	}
	found := false
	for _, args := range gotArgs {
		if len(args) == 3 && args[0] == "install" && args[1] == "--cask" && args[2] == "firefox" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected `brew install --cask firefox`, got %v", gotArgs)
	}
}

func TestSearchInstallUsesFormulaWriterForFormulae(t *testing.T) {
	var gotArgs [][]string
	r := brew.NewMockRunner()
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		gotArgs = append(gotArgs, args)
		ch := make(chan string)
		close(ch)
		errCh := make(chan error, 1)
		close(errCh)
		return ch, errCh
	}
	m := New(brew.NewClient(r), config.Default())
	m.activePanel = PanelSearch
	m.searchResults = []brew.SearchResult{{Name: "ripgrep", IsFormula: true}}
	p := m.panels[PanelSearch]
	p.items = []string{"ripgrep  formula"}
	p.selected = 0

	nm, _ := m.confirmMutation(mutInstall, "Install")
	mm := nm.(Model)
	mm.activeModal = nil
	res, cmd := mm.handleModalResult(&modal.ConfirmResult{Confirmed: true}, nil)
	nm2 := res.(Model)
	_, _ = cmd(), nm2.tasks.RunNext()()
	found := false
	for _, args := range gotArgs {
		if len(args) == 2 && args[0] == "install" && args[1] == "ripgrep" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected `brew install ripgrep` (formula), got %v", gotArgs)
	}
}

// --- Finding 13: pin decision by name, formula-only -----------------------

func TestTogglePinLooksUpPinnedStateByName(t *testing.T) {
	var gotArgs [][]string
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		gotArgs = append(gotArgs, args)
		return []byte(""), nil
	}
	m := New(brew.NewClient(r), config.Default())
	m.activePanel = PanelFormulae
	p := m.panels[PanelFormulae]
	// Leaves filter: items show only "bar", but formulae keeps both —
	// the old index-based lookup read foo's Pinned flag for bar.
	p.items = []string{"bar  2.0  bottled"}
	p.formulae = []brew.Formula{
		{Name: "foo", Pinned: false},
		{Name: "bar", Pinned: true},
	}
	p.selected = 0

	tm, cmd := m.togglePin()
	tmm := tm.(Model)
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatal("expected TaskStartedMsg from togglePin")
	}
	msg := tmm.tasks.RunNext()()
	if comp, ok := msg.(TaskCompletedMsg); !ok || comp.Err != nil {
		t.Fatalf("expected clean completion, got %T (%v)", msg, comp)
	}
	found := false
	for _, args := range gotArgs {
		if len(args) == 2 && args[0] == "unpin" && args[1] == "bar" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected `brew unpin bar` (bar is pinned), got %v", gotArgs)
	}
}

// --- Finding 14: batch enqueue queue-full ---------------------------------

func TestBatchUpgradeQueueFullSurfacesAndAdjustsCount(t *testing.T) {
	m := newTestModel()
	panel := m.panels[PanelOutdated]
	for i := 0; i < task.DefaultMaxQueue+2; i++ {
		panel.items = append(panel.items, "pkg"+string(rune('a'+i))+"  1.0")
		panel.formulae = append(panel.formulae, brew.Formula{Name: "pkg" + string(rune('a'+i))})
		m.batch.selected[i] = true
	}

	nm, _ := m.batchUpgrade()
	rm := nm.(Model)
	if rm.batchCount != task.DefaultMaxQueue {
		t.Fatalf("expected batchCount to be capped at the queue size (%d), got %d",
			task.DefaultMaxQueue, rm.batchCount)
	}
	if rm.toast == nil || !strings.Contains(rm.toast.Message, "Queue full") {
		t.Fatalf("expected a queue-full warning toast, got %+v", rm.toast)
	}
}

// --- Finding 25: U restricted to package panels ---------------------------

func TestUpgradeAllRestrictedToPackagePanels(t *testing.T) {
	for _, panel := range []PanelID{PanelServices, PanelTaps, PanelSearch, PanelStatus} {
		m := newTestModel()
		m.activePanel = panel
		m.panels[panel].items = []string{"pkg  1.0"}
		m.panels[panel].selected = 0
		m = sendKey(m, "U")
		if m.activeModal != nil {
			t.Errorf("panel %v: U must not open the upgrade-all confirm", panel)
		}
	}
	for _, panel := range []PanelID{PanelFormulae, PanelCasks, PanelOutdated} {
		m := newTestModel()
		m.activePanel = panel
		m.panels[panel].items = []string{"pkg  1.0"}
		m.panels[panel].selected = 0
		m = sendKey(m, "U")
		if m.activeModal == nil {
			t.Errorf("panel %v: U must open the upgrade-all confirm", panel)
		}
	}
}

// --- Finding 27: batch counter only counts batch completions --------------

func TestBatchCountOnlyDecrementsBatchTasks(t *testing.T) {
	m := newTestModel()
	m.batchCount = 2
	m.batchIDs = map[string]int{"batch-upgrade:foo": 1, "batch-upgrade:bar": 1}

	m = updateModel(m, TaskCompletedMsg{ID: "doctor", Title: "Doctor"})
	if m.batchCount != 2 {
		t.Fatalf("unrelated completion must not skew batchCount, got %d", m.batchCount)
	}

	m = updateModel(m, TaskCompletedMsg{ID: "batch-upgrade:foo", Title: "Upgrade foo"})
	if m.batchCount != 1 {
		t.Fatalf("batch completion must decrement batchCount, got %d", m.batchCount)
	}
}

// --- Finding 28: unloaded status tabs show an explicit state --------------

func TestStatusTabsShowNotLoadedInsteadOfEternalSpinner(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelStatus
	m.activeTab = 1
	m.panels[PanelStatus].items = nil

	out := m.renderContent(120, 40)
	if strings.Contains(out, "Loading...") {
		t.Fatalf("unloaded status tab must not show an eternal spinner, got:\n%s", out)
	}
	if !strings.Contains(out, "not loaded yet") {
		t.Fatalf("expected an explicit not-loaded state, got:\n%s", out)
	}
}

func TestFormulaeTabsShowNoPackageSelectedWithoutItems(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	m.activeTab = 2
	m.panels[PanelFormulae].items = nil

	out := m.renderContent(120, 40)
	if strings.Contains(out, "Loading...") {
		t.Fatalf("empty formulae deps tab must not show an eternal spinner, got:\n%s", out)
	}
}

// --- Finding 29: normalize in Update, not render --------------------------

func TestNormalizeClampsAfterDataShrink(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	p := m.panels[PanelFormulae]
	p.items = []string{"a", "b", "c"}
	p.selected = 2
	p.offset = 2

	// Data shrinks underneath the cursor.
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: []string{"a"}})
	if m.panels[PanelFormulae].selected != 0 {
		t.Fatalf("expected selected clamped to 0, got %d", m.panels[PanelFormulae].selected)
	}
}

func TestSidebarRenderDoesNotMutateCursor(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	p := m.panels[PanelFormulae]
	p.items = []string{"a", "b", "c"}
	p.selected = 5 // out of range on purpose
	rowsBefore := p.visibleRows

	_ = m.renderSidebar()
	if p.selected != 5 {
		t.Fatalf("render must not mutate selected (r1 finding 29), got %d", p.selected)
	}
	// visibleRows is model state too: render must not write it back
	// (r2 finding NEW-8).
	if p.visibleRows != rowsBefore {
		t.Fatalf("render must not mutate visibleRows (r2 finding NEW-8): before %d, after %d",
			rowsBefore, p.visibleRows)
	}
}

// TestWindowSizeSyncsVisibleRows locks r2 finding NEW-8: visibleRows is
// derived from the window geometry in the Update loop, and re-derived on
// panel switches (the sidebar layout depends on the active panel).
func TestWindowSizeSyncsVisibleRows(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	for i := range m.panels {
		want := max(0, m.computeContentHeights()[i]-1)
		if m.panels[i].visibleRows != want {
			t.Errorf("panel %d: visibleRows = %d, want %d", i, m.panels[i].visibleRows, want)
		}
	}

	// Switching the active panel changes the height split: the newly
	// active panel must get its recomputed viewport without a resize.
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	heights := m.computeContentHeights()
	for i := range m.panels {
		want := max(0, heights[i]-1)
		if m.panels[i].visibleRows != want {
			t.Errorf("after switch, panel %d: visibleRows = %d, want %d", i, m.panels[i].visibleRows, want)
		}
	}
}

// --- Finding 30: search cursor reset --------------------------------------

func TestExecuteSearchResetsSelection(t *testing.T) {
	m := newTestModel()
	m.panels[PanelSearch].selected = 5
	m.executeSearch("ripgrep")
	if m.panels[PanelSearch].selected != 0 {
		t.Fatalf("expected search cursor reset to 0, got %d", m.panels[PanelSearch].selected)
	}
}

// --- Finding 22: search info installed detection --------------------------

func TestParsePackageInfoInstalledDetection(t *testing.T) {
	// Formula installed but neither installed_on_request nor
	// installed_as_dependency set — the old heuristic said "not installed".
	formulaJSON := `{"formulae":[{"name":"foo","installed":[{"version":"1.0"}]}],"casks":[]}`
	info, err := parsePackageInfo(formulaJSON)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Installed {
		t.Error("formula with installed entries must be reported installed")
	}

	caskJSON := `{"formulae":[],"casks":[{"name":"bar","installed":[{"version":"2.0"}]}]}`
	info, err = parsePackageInfo(caskJSON)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Installed {
		t.Error("cask with installed entries must be reported installed")
	}

	notInstalled := `{"formulae":[],"casks":[{"name":"baz"}]}`
	info, err = parsePackageInfo(notInstalled)
	if err != nil {
		t.Fatal(err)
	}
	if info.Installed {
		t.Error("cask without installed entries must be reported not installed")
	}
}

// --- Finding 24: documented keys have handlers ----------------------------

func TestServicesRestartKeyHandler(t *testing.T) {
	var gotArgs [][]string
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		gotArgs = append(gotArgs, args)
		return []byte(""), nil
	}
	m := New(brew.NewClient(r), config.Default())
	m.switchPanel(PanelServices)
	p := m.panels[PanelServices]
	p.items = []string{"redis  started"}
	p.selected = 0

	tm, cmd := m.serviceAction("restart")
	tmm := tm.(Model)
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatal("expected TaskStartedMsg for restart")
	}
	msg := tmm.tasks.RunNext()()
	if comp, ok := msg.(TaskCompletedMsg); !ok || comp.Err != nil {
		t.Fatalf("expected clean completion, got %T (%v)", msg, comp)
	}
	found := false
	for _, args := range gotArgs {
		if len(args) == 3 && args[0] == "services" && args[1] == "restart" && args[2] == "redis" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected `brew services restart redis`, got %v", gotArgs)
	}
}

func TestSearchInstallKeyIHasHandler(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelSearch
	p := m.panels[PanelSearch]
	p.items = []string{"ripgrep  formula"}
	m.searchResults = []brew.SearchResult{{Name: "ripgrep", IsFormula: true}}
	p.selected = 0

	m = sendKey(m, "I")
	if m.activeModal == nil {
		t.Fatal("I on Search must open the install confirm (r1 finding 24)")
	}
}

func TestServicesRestartHintKeyHandled(t *testing.T) {
	var gotArgs [][]string
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		gotArgs = append(gotArgs, args)
		return []byte(""), nil
	}
	m := New(brew.NewClient(r), config.Default())
	m.switchPanel(PanelServices)
	p := m.panels[PanelServices]
	p.items = []string{"redis  stopped"}
	p.selected = 0

	nm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	mm := nm.(Model)
	if cmd == nil {
		t.Fatal("r on Services must enqueue a restart task (r1 finding 24)")
	}
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatal("expected TaskStartedMsg for restart")
	}
	msg := mm.tasks.RunNext()()
	if comp, ok := msg.(TaskCompletedMsg); !ok || comp.Err != nil {
		t.Fatalf("expected clean completion, got %T (%v)", msg, comp)
	}
	found := false
	for _, args := range gotArgs {
		if len(args) == 3 && args[0] == "services" && args[1] == "restart" && args[2] == "redis" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected `brew services restart redis`, got %v", gotArgs)
	}
}

// --- Finding 16: unknown service action resolves the task -----------------

func TestServiceActionUnknownResolvesWithError(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelServices
	p := m.panels[PanelServices]
	p.items = []string{"redis  started"}
	p.selected = 0

	tm, cmd := m.serviceAction("bogus")
	tmm := tm.(Model)
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatal("expected TaskStartedMsg")
	}
	msg := tmm.tasks.RunNext()()
	comp, ok := msg.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T — the task must not hang", msg)
	}
	if comp.Err == nil || !strings.Contains(comp.Err.Error(), "unknown service action") {
		t.Fatalf("expected unknown-action error, got %v", comp.Err)
	}
}

// --- Finding 3/26: command log is race-free and id-keyed ------------------

func TestCommandLogConcurrentAppendSetStatus(t *testing.T) {
	cl := NewCommandLog(20)
	done := make(chan struct{})
	for i := 0; i < 4; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 100; j++ {
				id := cl.Append("list")
				cl.SetStatus(id, CommandSuccess)
				_ = cl.Entries()
			}
		}()
	}
	for i := 0; i < 4; i++ {
		<-done
	}
	// Run under -race in CI; a plain pass here plus the race detector is
	// the assertion.
	if got := len(cl.Entries()); got != 20 {
		t.Fatalf("expected ring buffer capped at 20, got %d", got)
	}
}

// --- Finding 6/14 helpers sanity: toast tick scheduling -------------------

func TestToastTickMsgEventuallyClearsToast(t *testing.T) {
	m := newTestModel()
	m.setToast("hello", modal.ToastInfo)
	// Simulate the expiry path at the gui level by ticking until the
	// toast ages out — bounded by a real ~3s wait; the modal-level test
	// covers the boundary arithmetic itself.
	deadline := time.After(5 * time.Second)
	var cur Model = *m
	for cur.toast != nil {
		select {
		case <-deadline:
			t.Fatal("toast never dismissed via tick chain")
		default:
		}
		nm, cmd := cur.Update(modal.ToastTickMsg{})
		cur = nm.(Model)
		if cur.toast != nil && cmd == nil {
			t.Fatal("tick chain broke: no re-arm cmd while toast alive")
		}
		time.Sleep(300 * time.Millisecond)
	}
}
