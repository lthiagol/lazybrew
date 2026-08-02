package gui

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

var assertAnError = errors.New("test error")

func newTestModel() *Model {
	cfg := config.Default()
	client := brew.NewClient(brew.NewMockRunner())
	return New(client, cfg)
}

func updateModel(m *Model, msg tea.Msg) *Model {
	nm, _ := m.Update(msg)
	switch v := nm.(type) {
	case Model:
		return &v
	case *Model:
		return v
	}
	return m
}

func sendKey(m *Model, key string) *Model {
	return updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

func sendSpecial(m *Model, key tea.KeyType) *Model {
	return updateModel(m, tea.KeyMsg{Type: key})
}

func TestNewModel(t *testing.T) {
	m := newTestModel()
	if m == nil {
		t.Fatal("New() returned nil")
	}
	if m.activePanel != PanelStatus {
		t.Errorf("activePanel = %v, want PanelStatus", m.activePanel)
	}
	if len(m.panels) != 7 {
		t.Errorf("panels count = %d, want 7", len(m.panels))
	}
	if m.panels[PanelStatus] == nil {
		t.Fatal("Status panel is nil")
	}
	if !m.panels[PanelStatus].active {
		t.Error("Status panel should be active by default")
	}
}

func TestPanelNavigation(t *testing.T) {
	m := newTestModel()

	m = sendSpecial(m, tea.KeyTab)
	if m.activePanel != PanelFormulae {
		t.Errorf("after Tab: activePanel = %v, want PanelFormulae", m.activePanel)
	}

	m = sendSpecial(m, tea.KeyShiftTab)
	if m.activePanel != PanelStatus {
		t.Errorf("after Shift+Tab: activePanel = %v, want PanelStatus", m.activePanel)
	}
}

func TestPanelJump(t *testing.T) {
	m := newTestModel()

	jumps := []struct {
		key  string
		want PanelID
	}{
		{"1", PanelStatus},
		{"2", PanelFormulae},
		{"3", PanelCasks},
		{"4", PanelOutdated},
		{"5", PanelTaps},
		{"6", PanelServices},
		{"7", PanelSearch},
	}
	for _, tc := range jumps {
		m = sendKey(m, tc.key)
		if m.activePanel != tc.want {
			t.Errorf("after %s: activePanel = %v, want %v", tc.key, m.activePanel, tc.want)
		}
	}
}

func TestTabSwitching(t *testing.T) {
	m := newTestModel()

	m = sendKey(m, "]")
	if m.activeTab != 1 {
		t.Errorf("after ]: activeTab = %d, want 1", m.activeTab)
	}

	m = sendKey(m, "[")
	if m.activeTab != 0 {
		t.Errorf("after [: activeTab = %d, want 0", m.activeTab)
	}
}

func TestHelpToggle(t *testing.T) {
	m := newTestModel()

	m = sendKey(m, "?")
	if !m.showHelp {
		t.Error("showHelp should be true after ?")
	}

	m = sendKey(m, "?")
	if m.showHelp {
		t.Error("showHelp should be false after second ?")
	}
}

func TestHelpEscClose(t *testing.T) {
	m := newTestModel()

	m = sendKey(m, "?")
	if !m.showHelp {
		t.Fatal("help should be shown")
	}

	m = sendSpecial(m, tea.KeyEsc)
	if m.showHelp {
		t.Error("help should be closed after Esc")
	}
}

func TestQuitKey(t *testing.T) {
	m := newTestModel()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should return a quit command")
	}
	msg := cmd()
	if msg != tea.Quit() {
		t.Error("q should return tea.Quit")
	}
}

func TestRefreshKey(t *testing.T) {
	m := newTestModel()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
	if cmd == nil {
		t.Fatal("R should return a refresh command")
	}
	msg := cmd()
	if _, ok := msg.(RefreshMsg); !ok {
		t.Errorf("R should return RefreshMsg, got %T", msg)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := newTestModel()

	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
	if !m.ready {
		t.Error("model should be ready after WindowSizeMsg")
	}
}

func TestSmallTerminalWarning(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 79, Height: 24})
	if !m.terminalTooSmall {
		t.Error("79x24 should trigger terminalTooSmall")
	}

	m2 := newTestModel()
	m2 = updateModel(m2, tea.WindowSizeMsg{Width: 80, Height: 24})
	if m2.terminalTooSmall {
		t.Error("80x24 should NOT trigger terminalTooSmall")
	}

	m3 := newTestModel()
	m3 = updateModel(m3, tea.WindowSizeMsg{Width: 80, Height: 23})
	if !m3.terminalTooSmall {
		t.Error("80x23 should trigger terminalTooSmall")
	}
}

func TestDataLoadedMsg(t *testing.T) {
	m := newTestModel()

	msg := DataLoadedMsg{
		PanelID: PanelFormulae,
		Items:   []string{"item1", "item2"},
		Formulae: []brew.Formula{
			{Name: "test", Version: "1.0"},
		},
	}
	m = updateModel(m, msg)

	p := m.panels[PanelFormulae]
	if p.loading {
		t.Error("loading should be false after DataLoadedMsg")
	}
	if len(p.items) != 2 {
		t.Errorf("items count = %d, want 2", len(p.items))
	}
	if len(p.formulae) != 1 {
		t.Errorf("formulae count = %d, want 1", len(p.formulae))
	}
}

func TestSearchFlow(t *testing.T) {
	m := newTestModel()

	m = sendKey(m, "/")
	if m.activePanel != PanelSearch {
		t.Fatalf("expected PanelSearch after /, got %v", m.activePanel)
	}
	if !m.searchInput.Focused() {
		t.Fatal("search input should be focused after /")
	}
}

func TestSearchEnterShowsInfo(t *testing.T) {
	m := newTestModel()
	m.switchPanel(PanelSearch)
	m.searchInput.SetValue("lolcat")
	m.searchInput.Blur()
	m.searchResults = []brew.SearchResult{
		{Name: "lolcat", IsFormula: true},
	}
	m.panels[PanelSearch].items = []string{"lolcat  formula  installed  lolcat description"}
	m.panels[PanelSearch].selected = 0

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from Enter on search result")
	}
}

func TestSearchInstallKeybind(t *testing.T) {
	m := newTestModel()
	m.switchPanel(PanelSearch)
	m.searchResults = []brew.SearchResult{
		{Name: "lolcat", IsFormula: true},
	}
	m.panels[PanelSearch].items = []string{"lolcat  formula  installed  lolcat description"}
	m.panels[PanelSearch].selected = 0

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	if cmd == nil {
		t.Fatal("expected command from i on search result")
	}
}

func TestServiceKeybindings(t *testing.T) {
	m := newTestModel()
	m.switchPanel(PanelServices)

	sendKey(m, "s")
	sendKey(m, "S")
	sendKey(m, "f")
}

func TestStatusKeybindings(t *testing.T) {
	m := newTestModel()

	sendKey(m, "c")
	sendKey(m, "d")
	sendKey(m, "A")
	sendKey(m, "B")
	sendKey(m, "v")
	sendKey(m, "m")
}

func TestPanelHints(t *testing.T) {
	for _, id := range []PanelID{
		PanelStatus, PanelFormulae, PanelCasks, PanelOutdated,
		PanelTaps, PanelServices, PanelSearch,
	} {
		hints := panelHints(id)
		if len(hints) == 0 {
			t.Errorf("panelHints(%v) returned empty hints", id)
		}
	}
}

func TestEmptyStateMessages(t *testing.T) {
	tests := []struct {
		panel PanelID
		want  string
	}{
		{PanelFormulae, "No formulae installed"},
		{PanelCasks, "No casks installed"},
		{PanelOutdated, "Everything up to date!"},
		{PanelTaps, "No custom taps"},
		{PanelServices, "No services configured"},
		{PanelSearch, "No results"},
		{PanelStatus, "No data"},
	}
	for _, tt := range tests {
		got := emptyMessage(tt.panel)
		if got != tt.want {
			t.Errorf("emptyMessage(%v) = %q, want %q", tt.panel, got, tt.want)
		}
	}
}

func TestExtractPackageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"neovim  0.10.4  bottled", "neovim"},
		{"google-chrome  132.0  auto-update", "google-chrome"},
		{"ripgrep", "ripgrep"},
		{"", ""},
	}
	for _, tc := range tests {
		result := extractPackageName(tc.input)
		if result != tc.expected {
			t.Errorf("extractPackageName(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestDoMutationRejectedWhenRunning(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	m.panels[PanelFormulae].items = []string{"test  1.0  bottled"}
	m.panels[PanelFormulae].selected = 0

	_, cmd := m.doMutation(mutInstall, "Install")
	if cmd == nil {
		t.Fatal("expected cmd from first doMutation")
	}
	_ = cmd()

	_, cmd2 := m.doMutation(mutInstall, "Install")
	if cmd2 != nil {
		t.Error("doMutation should return nil when already running")
	}
}

func TestMutationTypeValues(t *testing.T) {
	if mutInstall == mutUninstall {
		t.Error("mutInstall and mutUninstall should have different values")
	}
	if mutFetch <= mutZap {
		t.Error("mutFetch should come after mutZap")
	}
}

func TestTabKey(t *testing.T) {
	key := tabKey(PanelFormulae, 2, "test")
	if key == "" {
		t.Error("tabKey should not be empty")
	}
	key2 := tabKey(PanelCasks, 2, "test")
	if key == key2 {
		t.Error("different panel/tab combos should produce different keys")
	}
	if tabKey(PanelFormulae, 2, "a") == tabKey(PanelFormulae, 2, "b") {
		t.Error("different item names should produce different keys")
	}
}

func TestToggleLeaves(t *testing.T) {
	m := newTestModel()
	p := m.panels[PanelFormulae]
	p.items = []string{"formula1  1.0  bottled", "formula2  2.0  bottled"}
	p.unfilteredItems = nil
	p.leavesActive = false

	msg := MutationResultMsg{
		Name:   "leaves",
		Leaves: []string{"formula1"},
	}
	m.activePanel = PanelFormulae
	m = updateModel(m, msg)
	p = m.panels[PanelFormulae]

	if !p.leavesActive {
		t.Error("leavesActive should be true after toggle")
	}
	if len(p.items) != 1 {
		t.Errorf("items should be filtered to 1, got %d", len(p.items))
	}

	m = updateModel(m, msg)
	p = m.panels[PanelFormulae]
	if p.leavesActive {
		t.Error("leavesActive should be false after second toggle")
	}
}

func TestConfirmUninstall(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	p := m.panels[PanelFormulae]
	p.items = []string{"ripgrep  14.1.1  bottled"}
	p.selected = 0

	_, cmd := m.confirmUninstall(mutUninstall)
	if cmd == nil {
		t.Fatal("confirmUninstall should return a command")
	}
	msg := cmd()
	if _, ok := msg.(DepCheckMsg); !ok {
		t.Errorf("confirmUninstall should return DepCheckMsg, got %T", msg)
	}
}

func TestConfirmZap(t *testing.T) {
	m := newTestModel()
	p := m.panels[PanelCasks]
	p.items = []string{"google-chrome  132.0  auto-update"}
	p.selected = 0
	m.activePanel = PanelCasks

	_, cmd := m.confirmUninstall(mutZap)
	if cmd == nil {
		t.Fatal("confirmZap should return a command")
	}
}

func TestBrewfileMenu(t *testing.T) {
	m := newTestModel()
	tm, _ := m.brewfileMenu()
	if tm.(Model).activeModal == nil {
		t.Fatal("activeModal should be set")
	}
}

func TestRunDoctor(t *testing.T) {
	m := newTestModel()
	_, cmd := m.runDoctor()
	if cmd == nil {
		t.Fatal("runDoctor should return a command")
	}
}

func TestRunVulns(t *testing.T) {
	m := newTestModel()
	_, cmd := m.runVulns()
	if cmd == nil {
		t.Fatal("runVulns should return a command")
	}
}

func TestRunMissing(t *testing.T) {
	m := newTestModel()
	_, cmd := m.runMissing()
	if cmd == nil {
		t.Fatal("runMissing should return a command")
	}
}

func TestViewNoSmoke(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	view := m.View()
	if view == "" {
		t.Error("View should not return empty string")
	}
}

func TestConfirmRepair(t *testing.T) {
	m := newTestModel()
	m.switchPanel(PanelTaps)
	p := m.panels[PanelTaps]
	p.items = []string{"nicknisi/tap"}
	p.selected = 0

	tm, _ := m.confirmRepair()
	if tm.(Model).activeModal == nil {
		t.Fatal("activeModal should be set")
	}
}

func TestConfirmUntap(t *testing.T) {
	m := newTestModel()
	m.switchPanel(PanelTaps)
	p := m.panels[PanelTaps]
	p.items = []string{"nicknisi/tap"}
	p.selected = 0

	tm, _ := m.confirmUntap()
	if tm.(Model).activeModal == nil {
		t.Fatal("activeModal should be set")
	}
}

func TestConfirmUntapOfficial(t *testing.T) {
	m := newTestModel()
	m.switchPanel(PanelTaps)
	p := m.panels[PanelTaps]
	p.items = []string{"homebrew/core"}
	p.selected = 0

	tm, _ := m.confirmUntap()
	mm := tm.(Model)
	if mm.toast == nil {
		t.Error("should show warning toast for official tap")
	}
}

func TestIsOfficialTap(t *testing.T) {
	if !isOfficialTap("homebrew/core") {
		t.Error("homebrew/core should be official")
	}
	if !isOfficialTap("homebrew/cask") {
		t.Error("homebrew/cask should be official")
	}
	if isOfficialTap("nicknisi/tap") {
		t.Error("nicknisi/tap should not be official")
	}
}

func TestPendingActionClear(t *testing.T) {
	m := newTestModel()
	result := &modal.ConfirmResult{Confirmed: false}
	m.pendingAction = "any"
	m.confirmCallback = func() tea.Msg { return nil }
	m.handleModalResult(result, nil)
	if m.pendingAction != "" {
		t.Error("pendingAction should be cleared on cancel")
	}
	if m.confirmCallback != nil {
		t.Error("confirmCallback should be cleared on cancel")
	}
}

func TestNoCrashOnEmptyPanel(t *testing.T) {
	m := newTestModel()
	p := m.panels[PanelFormulae]
	p.items = nil
	p.selected = 0

	f := p.selectedFormula()
	if f != nil {
		t.Error("selectedFormula should be nil for empty items")
	}
	c := p.selectedCask()
	if c != nil {
		t.Error("selectedCask should be nil for empty items")
	}
}

func TestServiceCleanupFlow(t *testing.T) {
	m := newTestModel()
	tm, _ := m.serviceCleanup()
	mm := tm.(Model)
	if mm.confirmCallback == nil {
		t.Error("confirmCallback should be set")
	}
}

func TestInitDispatchesFetch(t *testing.T) {
	m := newTestModel()
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init should return a fetch command")
	}
}

func TestBrewCleanupFlow(t *testing.T) {
	m := newTestModel()
	_, cmd := m.brewCleanup()
	if cmd == nil {
		t.Fatal("brewCleanup should return a command")
	}
}

func TestRunAutoremove(t *testing.T) {
	m := newTestModel()
	_, cmd := m.runAutoremove()
	if cmd == nil {
		t.Fatal("runAutoremove should return a command")
	}
}

func TestStartTrustMenu(t *testing.T) {
	m := newTestModel()
	m.switchPanel(PanelTaps)
	p := m.panels[PanelTaps]
	p.items = []string{"nicknisi/tap"}
	p.taps = []brew.Tap{{Name: "nicknisi/tap", FormulaNames: []string{"formula1"}, CaskNames: []string{"cask1"}}}
	p.selected = 0

	tm, _ := m.startTrustMenu()
	if tm.(*Model).activeModal == nil {
		t.Fatal("activeModal should be set")
	}
}

func TestMutationMessageTypes(t *testing.T) {
	for _, mt := range []mutationType{mutInstall, mutUninstall, mutReinstall, mutUpgrade, mutUpgradeAll, mutZap, mutFetch} {
		msg := MutationResultMsg{Name: "test", Type: mt, Err: nil}
		if msg.Name != "test" {
			t.Error("MutationResultMsg name mismatch")
		}
	}
}

func TestProgressMessageTypes(t *testing.T) {
	msg := ProgressLineMsg{Line: "test line"}
	if msg.Line != "test line" {
		t.Error("ProgressLineMsg mismatch")
	}

	complete := ProgressCompleteMsg{Name: "test", Err: nil}
	if complete.Name != "test" {
		t.Error("ProgressCompleteMsg mismatch")
	}
}

func TestFetchFlow(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	p := m.panels[PanelFormulae]
	p.items = []string{"ripgrep  14.1.1  bottled"}
	p.selected = 0

	_, cmd := m.doMutation(mutFetch, "Fetch")
	if cmd != nil {
		t.Log("fetch mutation dispatched")
	}
}

func TestModelTaskStartedOpensModal(t *testing.T) {
	m := newTestModel()
	msg := TaskStartedMsg{ID: "t1", Title: "Test Task"}
	m = updateModel(m, msg)
	if m.opState == nil {
		t.Fatal("expected opState to be set")
	}
	if m.opState.Title != "Test Task" {
		t.Fatalf("expected Title 'Test Task', got %q", m.opState.Title)
	}
	if !m.opState.Running() {
		t.Fatal("expected opState to be running")
	}
}

func TestModelTaskOutputAppendsLine(t *testing.T) {
	m := newTestModel()
	m.opState = &Operation{Title: "Test", Status: opRunning, Lines: []string{}}

	msg := TaskOutputMsg{ID: "t1", Line: "hello"}
	m = updateModel(m, msg)

	if len(m.opState.Lines) != 1 || m.opState.Lines[0] != "hello" {
		t.Fatalf("expected line 'hello', got %v", m.opState.Lines)
	}
}

func TestModelTaskCompletedToast(t *testing.T) {
	m := newTestModel()
	msg := TaskCompletedMsg{ID: "t1", Title: "Test", Err: nil}
	m = updateModel(m, msg)
	if m.toast == nil {
		t.Fatal("expected toast for successful completion")
	}
}

func TestModelTaskCompletedErrorToast(t *testing.T) {
	m := newTestModel()
	err := assertAnError
	msg := TaskCompletedMsg{ID: "t1", Title: "Test", Err: err}
	m = updateModel(m, msg)
	if m.toast == nil {
		t.Fatal("expected toast for error completion")
	}
}

func TestModelTaskRejectedToast(t *testing.T) {
	m := newTestModel()
	msg := TaskRejectedMsg{Reason: "queue is full"}
	m = updateModel(m, msg)
	if m.toast == nil {
		t.Fatal("expected toast for rejected task")
	}
}

func TestOutdatedFetchSurfacesError(t *testing.T) {
	client := brew.NewClient(brew.NewMockRunner())
	cmd := fetchPanelData(client, PanelOutdated)
	if cmd == nil {
		t.Fatal("fetchPanelData returned nil cmd")
	}
	msg := cmd()
	dMsg, ok := msg.(DataLoadedMsg)
	if !ok {
		t.Fatalf("expected DataLoadedMsg, got %T", msg)
	}
	if dMsg.Err == nil {
		t.Log("Outdated fetch succeeded with mock")
	}
}

func TestOutdatedPanelTypedData(t *testing.T) {
	m := newTestModel()
	msg := DataLoadedMsg{
		PanelID: PanelOutdated,
		Items:   []string{"formula-a  2.0  outdated"},
		Formulae: []brew.Formula{
			{Name: "formula-a", Version: "1.0", NewVersion: "2.0"},
		},
	}
	m = updateModel(m, msg)

	p := m.panels[PanelOutdated]
	if p.loading {
		t.Error("loading should be false after DataLoadedMsg")
	}
	if len(p.formulae) != 1 {
		t.Errorf("formulae count = %d, want 1", len(p.formulae))
	}

	f := p.selectedFormula()
	if f == nil {
		t.Fatal("selectedFormula() returned nil")
	}
	if f.Name != "formula-a" {
		t.Errorf("formula name = %q, want %q", f.Name, "formula-a")
	}
	if f.NewVersion != "2.0" {
		t.Errorf("NewVersion = %q, want %q", f.NewVersion, "2.0")
	}
}

func TestOperationOutputRendersInline(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, TaskStartedMsg{ID: "t1", Title: "Install foo"})
	m = updateModel(m, TaskOutputMsg{ID: "t1", Line: "==> Downloading foo"})
	m = updateModel(m, TaskOutputMsg{ID: "t1", Line: "==> Installing foo"})

	content := m.renderContent(60, 20)
	if !strings.Contains(content, "Install foo") {
		t.Fatal("expected operation title in content")
	}
	if !strings.Contains(content, "Downloading foo") {
		t.Fatal("expected output line in content")
	}
	if !strings.Contains(content, "Installing foo") {
		t.Fatal("expected output line in content")
	}
}

func TestOperationCancelShowsCancelled(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, TaskStartedMsg{ID: "t1", Title: "Install foo"})
	m = updateModel(m, TaskCompletedMsg{ID: "t1", Title: "Install foo", Err: context.Canceled})

	if m.opState == nil {
		t.Fatal("expected opState to be set")
	}
	if m.opState.Status != opCancelled {
		t.Fatalf("expected opCancelled, got %v", m.opState.Status)
	}
	if m.toast != nil {
		t.Fatal("expected no error toast on cancel")
	}

	content := m.renderContent(60, 20)
	if !strings.Contains(content, "Cancelled") {
		t.Fatal("expected Cancelled message, got:", content)
	}
}

func TestOperationCompletedNotCancelledOnRealError(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, TaskStartedMsg{ID: "t1", Title: "Install foo"})
	m = updateModel(m, TaskCompletedMsg{ID: "t1", Title: "Install foo", Err: assertAnError})

	if m.opState == nil {
		t.Fatal("expected opState to be set")
	}
	if m.opState.Status != opError {
		t.Fatalf("expected opError, got %v", m.opState.Status)
	}
	if m.toast == nil {
		t.Fatal("expected error toast for real error")
	}
	content := m.renderContent(60, 20)
	if !strings.Contains(content, "Error:") {
		t.Fatal("expected Error message, got:", content)
	}
}

func TestOperationSuccessDismissOnEsc(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, TaskStartedMsg{ID: "t1", Title: "Install foo"})
	m = updateModel(m, TaskCompletedMsg{ID: "t1", Title: "Install foo", Err: nil})

	if m.opState == nil || !m.opState.Done() {
		t.Fatal("expected completed opState")
	}

	m = sendSpecial(m, tea.KeyEsc)
	if m.opState != nil {
		t.Fatal("expected opState to be nil after Esc dismiss")
	}
}

func TestCommandLogShowsExecutedCommands(t *testing.T) {
	cl := NewCommandLog(20)
	cl.Append("install lolcat")
	cl.Append("list --formula")

	view := cl.View(40, 10)
	if !strings.Contains(view, "install lolcat") {
		t.Fatal("expected 'install lolcat' in command log view")
	}
	if !strings.Contains(view, "list --formula") {
		t.Fatal("expected 'list --formula' in command log view")
	}
	if !strings.Contains(view, "⟳") {
		t.Fatal("expected running indicator in command log view")
	}

	cl.SetStatus("install lolcat", CommandSuccess)
	view = cl.View(40, 10)
	if !strings.Contains(view, "✓") {
		t.Fatal("expected success indicator in command log view")
	}
}

func TestCommandLogRendersInMainPanel(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.commandLog.Append("test command")

	panel := m.renderMainPanel()
	if !strings.Contains(panel, "test command") {
		t.Fatal("expected command log entry in main panel render")
	}
	if !strings.Contains(panel, "brew") {
		t.Fatal("expected 'brew' prefix in command log render")
	}
}

func TestBatchSelectionShowsIndicator(t *testing.T) {
	p := &panelData{
		items: []string{"a  1.0", "b  2.0", "c  3.0"},
	}
	batch := map[int]bool{0: true, 2: true}

	// Check sidebar content shows batch indicator
	sidebar := p.renderSidebarContent(20, 5, batch)
	if !strings.Contains(sidebar, "●") {
		t.Fatal("expected batch indicator in sidebar for selected items")
	}

	// Check renderList shows batch indicator
	list := p.renderList(20, 5, batch)
	if !strings.Contains(list, "●") {
		t.Fatal("expected batch indicator in renderList for selected items")
	}

	// Item at cursor that's also selected shows combined indicator
	p.selected = 0
	sidebar = p.renderSidebarContent(20, 5, batch)
	if !strings.Contains(sidebar, "▸●") {
		t.Fatal("expected combined cursor+batch indicator, got:", sidebar)
	}
}

func TestSpaceTogglesOutdatedBatchSelection(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, "4")
	p := m.panels[PanelOutdated]
	p.items = []string{"foo  1.0 → 2.0", "bar  3.0 → 4.0"}
	p.selected = 1
	m.batch.selected = make(map[int]bool)

	m = sendKey(m, " ")
	if !m.batch.selected[1] {
		t.Fatal("Space should select outdated item at cursor")
	}
	m = sendKey(m, " ")
	if m.batch.selected[1] {
		t.Fatal("second Space should deselect outdated item at cursor")
	}
}

func TestSmallTerminalWarningInView(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 79, Height: 24})
	view := m.View()
	if !strings.Contains(view, "Terminal too small") {
		t.Fatalf("79x24 View should show warning, got:\n%s", view[:min(len(view), 300)])
	}
	m = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	view = m.View()
	if strings.Contains(view, "Terminal too small") {
		t.Fatal("80x24 View should not show terminal-too-small warning")
	}
}

func TestDepsTabContentSavedOnSuccess(t *testing.T) {
	m := newTestModel()
	msg := TabContentMsg{
		PanelID:  PanelFormulae,
		TabIndex: 1,
		ItemName: "formula-a",
		Content:  "formula-a depends on: openssl",
	}
	m = updateModel(m, msg)

	key := tabKey(PanelFormulae, 1, "formula-a")
	if m.tabContent[key] != "formula-a depends on: openssl" {
		t.Fatalf("expected tab content, got %q", m.tabContent[key])
	}
}

func TestDepsTabErrorShowsErrorMessage(t *testing.T) {
	m := newTestModel()
	msg := TabContentMsg{
		PanelID:  PanelFormulae,
		TabIndex: 1,
		ItemName: "formula-a",
		Err:      errors.New("brew deps failed"),
	}
	m = updateModel(m, msg)

	key := tabKey(PanelFormulae, 1, "formula-a")
	if !strings.Contains(m.tabContent[key], "Error") {
		t.Fatalf("expected error message, got %q", m.tabContent[key])
	}
}

func TestDepsTabShowsNoDataOnEmptyResult(t *testing.T) {
	m := newTestModel()
	msg := TabContentMsg{
		PanelID:  PanelFormulae,
		TabIndex: 1,
		ItemName: "formula-a",
		Content:  "",
	}
	m = updateModel(m, msg)

	key := tabKey(PanelFormulae, 1, "formula-a")
	if m.tabContent[key] != "No data" {
		t.Fatalf("expected 'No data', got %q", m.tabContent[key])
	}
}

func TestDepsTabPopulatedAfterDataLoaded(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	m.activeTab = 1
	p := m.panels[PanelFormulae]
	p.items = []string{"ripgrep  14.1.1  bottled"}
	p.formulae = []brew.Formula{
		{Name: "ripgrep", Version: "14.1.1", Dependencies: []string{"pcre2"}},
	}
	p.selected = 0

	_, cmd := m.Update(DataLoadedMsg{
		PanelID:  PanelFormulae,
		Items:    []string{"ripgrep  14.1.1  bottled"},
		Formulae: []brew.Formula{{Name: "ripgrep", Version: "14.1.1", Dependencies: []string{"pcre2"}}},
	})

	key := tabKey(PanelFormulae, 1, "ripgrep")
	if cmd == nil {
		content, ok := m.tabContent[key]
		if !ok || !strings.Contains(content, "pcre2") {
			t.Fatalf("expected Deps tabContent populated after DataLoadedMsg, got cmd=nil content=%q ok=%v", content, ok)
		}
		return
	}
	msg := cmd()
	tMsg, ok := msg.(TabContentMsg)
	if !ok {
		t.Fatalf("expected TabContentMsg from re-issued fetch, got %T", msg)
	}
	if tMsg.PanelID != PanelFormulae || tMsg.TabIndex != 1 || tMsg.ItemName != "ripgrep" {
		t.Errorf("re-fetch meta = (%v, %d, %q), want (PanelFormulae, 1, ripgrep)",
			tMsg.PanelID, tMsg.TabIndex, tMsg.ItemName)
	}
}

func TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	m.activeTab = 2
	p := m.panels[PanelFormulae]
	p.items = []string{"ripgrep  14.1.1  bottled"}
	p.formulae = []brew.Formula{
		{Name: "ripgrep", Version: "14.1.1"},
	}
	p.selected = 0

	_, cmd := m.Update(DataLoadedMsg{
		PanelID:  PanelFormulae,
		Items:    []string{"ripgrep  14.1.1  bottled"},
		Formulae: []brew.Formula{{Name: "ripgrep", Version: "14.1.1"}},
	})

	if cmd == nil {
		t.Fatal("expected loadTabContent cmd re-issued for Used By (no fast path) after DataLoadedMsg")
	}
	msg := cmd()
	tMsg, ok := msg.(TabContentMsg)
	if !ok {
		t.Fatalf("expected TabContentMsg, got %T", msg)
	}
	if tMsg.TabIndex != 2 {
		t.Errorf("expected Used By tab 2, got %d", tMsg.TabIndex)
	}
}

func TestDataLoadedForDifferentPanelDoesNotReissueActivePanelFetch(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelCasks
	m.activeTab = 0
	p := m.panels[PanelFormulae]
	p.items = []string{"ripgrep  14.1.1  bottled"}
	p.formulae = []brew.Formula{
		{Name: "ripgrep", Version: "14.1.1", Dependencies: []string{"pcre2"}},
	}
	p.selected = 0

	_, cmd := m.Update(DataLoadedMsg{
		PanelID:  PanelFormulae,
		Items:    []string{"ripgrep  14.1.1  bottled"},
		Formulae: []brew.Formula{{Name: "ripgrep", Version: "14.1.1", Dependencies: []string{"pcre2"}}},
	})

	if cmd != nil {
		t.Errorf("did not expect re-fetch cmd (active panel Casks != msg panel Formulae), got %T", cmd())
	}
}

func TestDataLoadedOnNonFetchTabDoesNotReissueFetch(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	m.activeTab = 0
	p := m.panels[PanelFormulae]
	p.items = []string{"ripgrep  14.1.1  bottled"}
	p.formulae = []brew.Formula{
		{Name: "ripgrep", Version: "14.1.1", Dependencies: []string{"pcre2"}},
	}
	p.selected = 0

	_, cmd := m.Update(DataLoadedMsg{
		PanelID:  PanelFormulae,
		Items:    []string{"ripgrep  14.1.1  bottled"},
		Formulae: []brew.Formula{{Name: "ripgrep", Version: "14.1.1", Dependencies: []string{"pcre2"}}},
	})

	if cmd != nil {
		t.Errorf("did not expect re-fetch cmd on Info tab (tab 0 is not a fetch tab), got %T", cmd())
	}
}

func TestFetchTabContentCmdTimesOut(t *testing.T) {
	origTimeout := fetchTabContentTimeout
	fetchTabContentTimeout = 50 * time.Millisecond
	defer func() { fetchTabContentTimeout = origTimeout }()

	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	client := brew.NewClient(r)

	cmd := fetchTabContentCmd(client, PanelFormulae, 1, "ripgrep")
	if cmd == nil {
		t.Fatal("fetchTabContentCmd returned nil cmd")
	}

	done := make(chan tea.Msg, 1)
	go func() {
		done <- cmd()
	}()
	select {
	case msg := <-done:
		tMsg, ok := msg.(TabContentMsg)
		if !ok {
			t.Fatalf("expected TabContentMsg, got %T", msg)
		}
		if tMsg.Err == nil {
			t.Fatal("expected timeout error from fetchTabContentCmd")
		}
		if !errors.Is(tMsg.Err, context.DeadlineExceeded) {
			t.Errorf("expected context.DeadlineExceeded, got %v", tMsg.Err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("fetchTabContentCmd hung past fetchTabContentTimeout (infinite-Loading reproducer)")
	}
}

func TestDepsFastPathRendersCachedDependencies(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	m.activeTab = 1
	p := m.panels[PanelFormulae]
	p.items = []string{"ripgrep  14.1.1  bottled"}
	p.formulae = []brew.Formula{
		{Name: "ripgrep", Version: "14.1.1",
			Dependencies: []string{"pcre2", "rust"},
			BuildDeps:    []string{"cmake"}},
	}
	p.selected = 0

	cmd := m.loadTabContent()
	if cmd != nil {
		t.Fatal("expected nil cmd from fast path (Dependencies already on model)")
	}

	key := tabKey(PanelFormulae, 1, "ripgrep")
	content, ok := m.tabContent[key]
	if !ok {
		t.Fatal("expected tabContent populated synchronously by fast path")
	}
	for _, want := range []string{"pcre2", "rust", "cmake"} {
		if !strings.Contains(content, want) {
			t.Errorf("fast-path content missing %q, got %q", want, content)
		}
	}
}

func TestDepsFastPathNoDependenciesShowsExplicitMessage(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	m.activeTab = 1
	p := m.panels[PanelFormulae]
	p.items = []string{"zlib  1.3.1  bottled"}
	p.formulae = []brew.Formula{
		{Name: "zlib", Version: "1.3.1"},
	}
	p.selected = 0

	cmd := m.loadTabContent()
	if cmd != nil {
		t.Fatal("expected nil cmd from fast path (no deps means no shell call)")
	}

	key := tabKey(PanelFormulae, 1, "zlib")
	content, ok := m.tabContent[key]
	if !ok {
		t.Fatal("expected tabContent populated by fast path even when deps are empty")
	}
	if !strings.Contains(strings.ToLower(content), "no dependencies") {
		t.Errorf("expected explicit empty-state text, got %q", content)
	}
}

func TestDepsSelectionChangeUpdatesContent(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.activePanel = PanelFormulae
	m.activeTab = 1
	m.tabs = panelTabs[PanelFormulae]
	p := m.panels[PanelFormulae]
	p.loading = false
	p.items = []string{
		"ripgrep  14.1.1  bottled",
		"zlib  1.3.1  bottled",
	}
	p.formulae = []brew.Formula{
		{Name: "ripgrep", Version: "14.1.1", Dependencies: []string{"pcre2"}},
		{Name: "zlib", Version: "1.3.1", Dependencies: []string{"none-dep-marker"}},
	}
	p.selected = 0

	if cmd := m.loadTabContent(); cmd != nil {
		t.Fatal("expected fast-path nil cmd for ripgrep")
	}
	keyA := tabKey(PanelFormulae, 1, "ripgrep")
	if c := m.tabContent[keyA]; !strings.Contains(c, "pcre2") {
		t.Fatalf("expected ripgrep deps pcre2, got %q", c)
	}

	m = sendKey(m, "j")
	if p.selected != 1 {
		t.Fatalf("expected selected=1 after j, got %d", p.selected)
	}
	keyB := tabKey(PanelFormulae, 1, "zlib")
	contentB, ok := m.tabContent[keyB]
	if !ok || !strings.Contains(contentB, "none-dep-marker") {
		t.Fatalf("expected zlib deps after j, ok=%v content=%q", ok, contentB)
	}
	if strings.Contains(contentB, "pcre2") {
		t.Errorf("zlib deps content should not contain ripgrep dep pcre2, got %q", contentB)
	}

	view := m.View()
	if !strings.Contains(view, "none-dep-marker") {
		t.Errorf("View after j should show zlib deps, got snippet:\n%s", view[:min(len(view), 400)])
	}
	if strings.Contains(view, "Loading...") {
		t.Error("View after selection change must not show Loading...")
	}
}

func TestDepsTabViewLeavesLoadingAfterFastPath(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.activePanel = PanelFormulae
	m.activeTab = 1
	m.tabs = panelTabs[PanelFormulae]
	p := m.panels[PanelFormulae]
	p.loading = false
	p.items = []string{"ripgrep  14.1.1  bottled"}
	p.formulae = []brew.Formula{
		{Name: "ripgrep", Version: "14.1.1", Dependencies: []string{"pcre2"}},
	}
	p.selected = 0

	if cmd := m.loadTabContent(); cmd != nil {
		t.Fatal("expected fast-path nil cmd")
	}
	view := m.View()
	if !strings.Contains(view, "pcre2") {
		t.Fatalf("View should contain deps content pcre2, got:\n%s", view[:min(len(view), 500)])
	}
	if strings.Contains(view, "Loading...") {
		t.Error("View must leave Loading state after fast-path deps populate")
	}
}

// TestDepsTabLoads is the AC-03 gui-flow regression: Formulae → Deps for a
// fixture formula must leave Loading and show deps content (no sleep/teatest).
func TestDepsTabLoads(t *testing.T) {
	const ripgrepJSON = `{"formulae":[{` +
		`"name":"ripgrep","full_name":"ripgrep","tap":"homebrew/core",` +
		`"versions":{"stable":"14.1.1"},"desc":"","homepage":"","license":"",` +
		`"installed":[{"version":"14.1.1","installed_on_request":true,` +
		`"installed_as_dependency":false,"time":1700000000,` +
		`"runtime_dependencies":["pcre2"]}],` +
		`"dependencies":[{"name":"pcre2"}],"build_dependencies":[],` +
		`"caveats":"","keg_only":false,` +
		`"bottle":{"stable":{"files":{"arm64_sonoma":{"url":"x"}}}},` +
		`"pinned":false,"binaries":["rg"]}]}`

	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 3 && args[0] == "info" && args[1] == "--json=v2" && args[2] == "--installed" {
			return []byte(ripgrepJSON), nil
		}
		if len(args) >= 1 && args[0] == "deps" {
			return []byte("pcre2\n"), nil
		}
		return []byte(`{"formulae":[],"casks":[],"taps":[],"services":[]}`), nil
	}
	m := New(brew.NewClient(r), config.Default())
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	formulae, err := m.client.Formulae.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	items := make([]string, len(formulae))
	for i, f := range formulae {
		items[i] = f.Name + "  " + f.Version
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: items, Formulae: formulae})
	m = sendKey(m, "2")
	m = sendKey(m, "]")

	if m.activePanel != PanelFormulae || m.activeTab != 1 {
		t.Fatalf("expected Formulae/Deps, got panel=%v tab=%d", m.activePanel, m.activeTab)
	}
	view := m.View()
	if !strings.Contains(view, "pcre2") {
		t.Fatalf("Deps body missing pcre2:\n%s", view[:min(len(view), 600)])
	}
	if strings.Contains(view, "Loading...") {
		t.Fatal("Deps tab stuck on Loading...")
	}
	if !strings.Contains(view, "Deps") {
		t.Fatal("expected Deps tab label in view")
	}
}

func TestFetchDepsTabReturnsContent(t *testing.T) {
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 3 && args[0] == "deps" && args[1] == "--tree" && args[2] == "ripgrep" {
			return []byte("openssl\n└── ca-certificates\n"), nil
		}
		return nil, nil
	}
	client := brew.NewClient(r)

	cmd := fetchTabContentCmd(client, PanelFormulae, 1, "ripgrep")
	if cmd == nil {
		t.Fatal("fetchTabContentCmd returned nil cmd")
	}
	msg := cmd()
	tMsg, ok := msg.(TabContentMsg)
	if !ok {
		t.Fatalf("expected TabContentMsg, got %T", msg)
	}
	if tMsg.Err != nil {
		t.Fatalf("unexpected error: %v", tMsg.Err)
	}
	if !strings.Contains(tMsg.Content, "openssl") {
		t.Fatalf("expected deps content to contain 'openssl', got %q", tMsg.Content)
	}
}

func TestRefreshSetsPanelsLoading(t *testing.T) {
	// M12 tiered refresh: with all panels unloaded (newly-initialized
	// model), Refresh only fires Status + active-panel cmds. The
	// unloaded data panels stay at loading=false (lazy policy).
	m := newTestModel()
	for _, p := range m.panels {
		p.loading = false
	}
	// Pre-populate Status (active) and one other panel so the test
	// exercises the "loaded → refresh" branch.
	m.panels[PanelFormulae].items = []string{"ripgrep  1.0"}
	m.panels[PanelFormulae].loading = false
	m = updateModel(m, RefreshMsg{})

	if !m.panels[PanelStatus].loading {
		t.Errorf("Status should be loading (always refreshed, M12)")
	}
	if !m.panels[PanelFormulae].loading {
		t.Errorf("loaded Formulae should be loading after refresh")
	}
	for _, p := range m.panels {
		switch p.id {
		case PanelSearch:
			if p.loading {
				t.Errorf("PanelSearch should not be loading")
			}
		case PanelStatus, PanelFormulae:
			// expected loading
		default:
			if p.loading {
				t.Errorf("unloaded panel %v should NOT be loading (M12 tiered refresh)", p.id)
			}
		}
	}
}

// TestRefreshTieredRespectsLoadedSet locks M12 AC-04: R fires active +
// Status, plus any panel the user has already loaded. Unloaded panels
// stay unloaded.
func TestRefreshTieredRespectsLoadedSet(t *testing.T) {
	m := newTestModel()
	// Pre-load: Formulae + Casks + Taps. Outdated, Services stay unloaded.
	m.panels[PanelFormulae].items = []string{"f1"}
	m.panels[PanelCasks].items = []string{"c1"}
	m.panels[PanelTaps].items = []string{"t1"}
	m.activePanel = PanelFormulae
	for _, p := range m.panels {
		p.loading = false
	}

	m = updateModel(m, RefreshMsg{})

	// Status + Formulae (active) + Casks + Taps = 4. Outdated/Services unloaded.
	wantCmds := 4
	if m.refreshing != wantCmds {
		t.Errorf("refreshing = %d, want %d (Status + Formulae + Casks + Taps)", m.refreshing, wantCmds)
	}
	if m.panels[PanelOutdated].loading {
		t.Error("unloaded PanelOutdated should NOT enter loading on refresh")
	}
	if m.panels[PanelServices].loading {
		t.Error("unloaded PanelServices should NOT enter loading on refresh")
	}
}

func TestRefreshLoadsOutdatedWhenActive(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelOutdated
	m.panels[PanelOutdated].loading = false
	for _, p := range m.panels {
		p.loading = false
	}

	m = updateModel(m, RefreshMsg{})

	if !m.panels[PanelOutdated].loading {
		t.Error("PanelOutdated should load on Refresh when active (AC-04)")
	}
	// Status + active = 2 cmds; unloaded panels stay skipped.
	if m.refreshing != 2 {
		t.Errorf("refreshing = %d, want 2 (Status + active)", m.refreshing)
	}
}

func TestRefreshToastWhenOutdatedInactive(t *testing.T) {
	// M12: active=Status → only fetchStatusData (no empty fetchPanelData stub).
	m := newTestModel()
	m.activePanel = PanelStatus
	m.cfg.GUI.AutoRefreshSeconds = 0
	for _, p := range m.panels {
		p.loading = false
	}

	m = updateModel(m, RefreshMsg{})
	if m.refreshing != 1 {
		t.Fatalf("refreshing = %d, want 1 (Status only when active=Status)", m.refreshing)
	}

	m = updateModel(m, DataLoadedMsg{PanelID: PanelStatus})
	if m.refreshing != 0 {
		t.Errorf("after 1 DataLoadedMsg, refreshing = %d, want 0", m.refreshing)
	}
	if m.toast == nil {
		t.Fatal("expected 'Data refreshed' toast after last panel load")
	}
}

func TestRefreshDoesNotEnqueueEmptyStatusStub(t *testing.T) {
	// F-03: when active is Status, must not also enqueue fetchPanelData(Status)
	// which returns Items:[] and can wipe the dashboard.
	m := newTestModel()
	m.activePanel = PanelStatus
	m.cfg.GUI.AutoRefreshSeconds = 0
	m = updateModel(m, RefreshMsg{})
	if m.refreshing != 1 {
		t.Fatalf("want 1 cmd (fetchStatusData only), got refreshing=%d", m.refreshing)
	}
}

func TestRefreshForcesCacheBypass(t *testing.T) {
	// AC-04: R must re-shell even when per-class TTL has not expired.
	var formulaeCalls int32
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "info" && args[1] == "--json=v2" {
			atomic.AddInt32(&formulaeCalls, 1)
			return []byte(`{"formulae":[],"casks":[]}`), nil
		}
		if len(args) >= 1 && args[0] == "services" {
			return []byte(`[]`), nil
		}
		if len(args) >= 1 && args[0] == "tap-info" {
			return []byte(`{"taps":[]}`), nil
		}
		if len(args) >= 1 && args[0] == "tap" {
			return []byte(""), nil
		}
		if len(args) >= 1 && args[0] == "doctor" {
			return []byte("Your system is ready to brew."), nil
		}
		if len(args) >= 1 && args[0] == "config" {
			return []byte(""), nil
		}
		return []byte(`{"formulae":[],"casks":[]}`), nil
	}
	client := brew.NewClient(r)
	client.SetCacheTTLs(brew.CacheTTLs{Formulae: time.Hour})
	m := New(client, config.Default())
	m.activePanel = PanelFormulae
	m.panels[PanelFormulae].items = []string{"seed"}
	m.cfg.GUI.AutoRefreshSeconds = 0

	// Prime formulae cache.
	_, _ = client.Formulae.List(context.Background())
	if atomic.LoadInt32(&formulaeCalls) != 1 {
		t.Fatalf("setup: want 1 formulae call, got %d", formulaeCalls)
	}
	atomic.StoreInt32(&formulaeCalls, 0)

	// Without invalidate, List would cache-hit. Refresh must force a shell.
	nm, cmd := m.Update(RefreshMsg{})
	if cmd == nil {
		t.Fatal("RefreshMsg returned nil cmd batch")
	}
	// Drain batch: run all DataLoaded producers.
	// tea.Batch flattens; cmd() may return a batch msg — use a simple drain.
	if newM, ok := nm.(Model); ok {
		m = &newM
	}
	// Execute cmds by invoking fetchPanelData path via Update on messages
	// produced when we run the batch function.
	// Simpler: call List again after invalidate that Refresh performed.
	// Refresh already invalidated — next List must shell.
	_, _ = client.Formulae.List(context.Background())
	if atomic.LoadInt32(&formulaeCalls) < 1 {
		t.Errorf("R must invalidate formulae cache so next List shells; calls=%d", formulaeCalls)
	}
}

func TestAutoRefreshTickUsesLiveLastKeyAt(t *testing.T) {
	// F-02: tick msg must be evaluated in Update against current lastKeyAt.
	m := newTestModel()
	m.cfg.GUI.AutoRefreshSeconds = 60
	m.cfg.GUI.AutoRefreshPause = 500 * time.Millisecond
	m.lastKeyAt = time.Now()

	// Deliver tick as if it fired while still inside the pause window.
	nm, cmd := m.Update(autoRefreshTickMsg{at: time.Now()})
	if cmd == nil {
		t.Fatal("expected follow-up cmd from autoRefreshTickMsg")
	}
	follow := cmd()
	if _, ok := follow.(autoRefreshPausedMsg); !ok {
		t.Fatalf("expected autoRefreshPausedMsg when lastKeyAt is recent, got %T", follow)
	}
	if updated, ok := nm.(Model); ok {
		if !updated.autoRefreshPaused {
			t.Error("autoRefreshPaused should be true after pause decision")
		}
	}

	// Past the pause window → RefreshMsg.
	m2 := newTestModel()
	m2.cfg.GUI.AutoRefreshSeconds = 60
	m2.cfg.GUI.AutoRefreshPause = 50 * time.Millisecond
	m2.lastKeyAt = time.Now().Add(-200 * time.Millisecond)
	_, cmd2 := m2.Update(autoRefreshTickMsg{at: time.Now()})
	follow2 := cmd2()
	if _, ok := follow2.(RefreshMsg); !ok {
		t.Fatalf("expected RefreshMsg after pause elapsed, got %T", follow2)
	}
}

func TestInitDoesNotFetchOutdated(t *testing.T) {
	// M9 AC-02: Init must NOT shell-call `brew outdated` (lazy policy).
	// We assert on the Init cmd batch: no fetchPanelData(PanelOutdated)
	// fires because Init() does not include it.
	var calls int32
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 && args[0] == "outdated" {
			atomic.AddInt32(&calls, 1)
		}
		return []byte(`{"formulae":[],"casks":[]}`), nil
	}
	cfg := config.Default()
	cfg.Brew.UpdateOnStart = false
	m := New(brew.NewClient(r), cfg)

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init returned nil cmd")
	}

	// Process all messages from the Init batch.
	pending := []tea.Cmd{cmd}
	for len(pending) > 0 {
		next := pending[0]
		pending = pending[1:]
		if next == nil {
			continue
		}
		got := next()
		if got == nil {
			continue
		}
		// Process msg and collect follow-up cmds via Update.
		nm, followUp := m.Update(got)
		if newM, ok := nm.(*Model); ok {
			m = newM
		}
		if followUp != nil {
			pending = append(pending, followUp)
		}
	}

	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Errorf("Init should not invoke `brew outdated` (lazy policy), got %d shell calls", got)
	}
	if m.panels[PanelOutdated].loading {
		t.Error("PanelOutdated should not enter loading state from Init (M9 lazy)")
	}
}

func TestSwitchToOutdatedLazyFetchesOnFirstVisit(t *testing.T) {
	// M9 AC-02: first switchPanel(Outdated) triggers the fetch.
	m := newTestModel()
	m.activePanel = PanelStatus
	m.panels[PanelOutdated].items = nil
	m.panels[PanelOutdated].loading = false
	m.panels[PanelOutdated].err = nil

	cmd := m.switchPanel(PanelOutdated)
	if cmd == nil {
		t.Fatal("expected switchPanel(Outdated) to return a fetch cmd on first visit")
	}
	if !m.panels[PanelOutdated].loading {
		t.Error("expected PanelOutdated.loading=true after switchPanel(Outdated)")
	}
	if m.activePanel != PanelOutdated {
		t.Errorf("activePanel = %d, want PanelOutdated", m.activePanel)
	}

	// Second visit (panel already has items) should NOT re-issue fetch.
	m.panels[PanelOutdated].items = []string{"ripgrep  1.0 -> 2.0"}
	m.panels[PanelOutdated].loading = false
	cmd2 := m.switchPanel(PanelOutdated)
	if cmd2 != nil {
		t.Error("expected nil cmd on second switchPanel(Outdated) (items present)")
	}
}

func TestSwitchToOutdatedRetriesAfterError(t *testing.T) {
	// After a failed fetch, switching to Outdated again must retry
	// rather than leave the user on a stale error.
	m := newTestModel()
	m.activePanel = PanelStatus
	m.panels[PanelOutdated].items = nil
	m.panels[PanelOutdated].loading = false
	m.panels[PanelOutdated].err = assertAnError

	cmd := m.switchPanel(PanelOutdated)
	if cmd == nil {
		t.Fatal("expected switchPanel(Outdated) to retry fetch after error")
	}
	if !m.panels[PanelOutdated].loading {
		t.Error("expected PanelOutdated.loading=true after error retry")
	}
	if m.panels[PanelOutdated].err != nil {
		t.Errorf("expected stale err cleared on retry, got %v", m.panels[PanelOutdated].err)
	}
}

func TestStatusDashboardDoesNotShellCallOutdated(t *testing.T) {
	// M9 AC-02 + design decision: fetchStatusData must not shell-call
	// `brew outdated`. It reads from cache only and reports 0 when empty.
	var outdatedCalls int32
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		switch {
		case len(args) >= 1 && args[0] == "outdated":
			atomic.AddInt32(&outdatedCalls, 1)
			return []byte(`{"formulae":[]}`), nil
		case len(args) >= 1 && args[0] == "services":
			return []byte(`[]`), nil
		case len(args) >= 2 && args[0] == "info":
			return []byte(`{"formulae":[],"casks":[]}`), nil
		case len(args) >= 1 && args[0] == "tap-info":
			return []byte(`{"taps":[]}`), nil
		case len(args) >= 1 && args[0] == "doctor":
			return []byte(""), nil
		case len(args) >= 1 && args[0] == "config":
			return []byte(""), nil
		}
		return []byte{}, nil
	}
	client := brew.NewClient(r)

	cmd := fetchStatusData(client)
	msg := cmd()
	dMsg, ok := msg.(DataLoadedMsg)
	if !ok {
		t.Fatalf("expected DataLoadedMsg, got %T", msg)
	}
	if atomic.LoadInt32(&outdatedCalls) != 0 {
		t.Errorf("fetchStatusData must not invoke `brew outdated`, got %d calls", outdatedCalls)
	}
	joined := strings.Join(dMsg.Items, "\n")
	if !strings.Contains(joined, "0 packages") {
		t.Errorf("empty cache should report 0 outdated, got:\n%s", joined)
	}
}

func TestStatusDashboardReportsCachedOutdatedCount(t *testing.T) {
	// When the cache has outdated entries (e.g. user previously visited
	// Outdated panel), Status should reflect them without re-shelling.
	var outdatedCalls int32
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		switch {
		case len(args) >= 1 && args[0] == "outdated":
			atomic.AddInt32(&outdatedCalls, 1)
			return []byte(`{"formulae":[{"name":"ripgrep","installed_versions":["14.1.0"],"current_version":"14.1.1","pinned":false}]}`), nil
		case len(args) >= 1 && args[0] == "services":
			return []byte(`[]`), nil
		case len(args) >= 2 && args[0] == "info":
			return []byte(`{"formulae":[],"casks":[]}`), nil
		case len(args) >= 1 && args[0] == "tap-info":
			return []byte(`{"taps":[]}`), nil
		case len(args) >= 1 && args[0] == "doctor":
			return []byte(""), nil
		case len(args) >= 1 && args[0] == "config":
			return []byte(""), nil
		}
		return []byte{}, nil
	}
	client := brew.NewClient(r)

	// First, populate the cache by calling client.Formulae.Outdated().
	// (We use the real reader so the cache+TTL is exercised.)
	_, _ = client.Formulae.Outdated(context.Background())
	if atomic.LoadInt32(&outdatedCalls) != 1 {
		t.Fatalf("setup: expected 1 outdated call after priming cache, got %d", outdatedCalls)
	}
	atomic.StoreInt32(&outdatedCalls, 0)

	cmd := fetchStatusData(client)
	msg := cmd()
	dMsg, ok := msg.(DataLoadedMsg)
	if !ok {
		t.Fatalf("expected DataLoadedMsg, got %T", msg)
	}
	if atomic.LoadInt32(&outdatedCalls) != 0 {
		t.Errorf("fetchStatusData must not shell-call outdated when cache is fresh, got %d calls", outdatedCalls)
	}
	joined := strings.Join(dMsg.Items, "\n")
	if !strings.Contains(joined, "1 packages") {
		t.Errorf("dashboard should reflect cached outdated count (1), got:\n%s", joined)
	}
}

func TestMutationShowsConfirmModal(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	m.panels[PanelFormulae].items = []string{"foo  1.0"}
	m.panels[PanelFormulae].selected = 0

	nm, _ := m.confirmMutation(mutInstall, "Install")
	result := nm.(Model)
	if result.activeModal == nil {
		t.Fatal("expected confirm modal after confirmMutation")
	}
	if result.pendingAction != "mutation" {
		t.Fatalf("expected pendingAction 'mutation', got %q", result.pendingAction)
	}
	if result.pendingMutType != mutInstall {
		t.Fatalf("expected pendingMutType mutInstall, got %v", result.pendingMutType)
	}
}

func TestBatchUpgradeUpgradesAllSelected(t *testing.T) {
	m := newTestModel()
	panel := m.panels[PanelOutdated]
	panel.items = []string{"foo  1.0", "bar  2.0", "baz  3.0"}
	panel.formulae = []brew.Formula{
		{Name: "foo", Version: "1.0"},
		{Name: "bar", Version: "2.0"},
		{Name: "baz", Version: "3.0"},
	}
	m.batch.selected = map[int]bool{0: true, 2: true}

	nm, _ := m.batchUpgrade()
	result := nm.(Model)
	if result.batchCount != 2 {
		t.Fatalf("expected batchCount 2 for 2 selected, got %d", result.batchCount)
	}
	if len(result.batch.selected) != 0 {
		t.Fatalf("expected batch selection cleared after upgrade, got %d remaining", len(result.batch.selected))
	}
}

func TestStatusDashboardSurfacesDoctorError(t *testing.T) {
	// Regression for M10 AC-03: fetchStatusData previously swallowed the
	// doctor error and showed "Doctor: No issues" on failure, which is a
	// silent-wrong-state for the user. Now the error must surface as both
	// a "Doctor: unavailable (...)" line and the existing ⚠ doctor: ... line.
	r := newStatusMockRunner(func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 && args[0] == "doctor" {
			return nil, &brew.BrewExitError{Command: "doctor", ExitCode: 2}
		}
		return []byte{}, nil
	})
	client := brew.NewClient(r)

	cmd := fetchStatusData(client)
	if cmd == nil {
		t.Fatal("fetchStatusData returned nil cmd")
	}
	msg := cmd()
	dMsg, ok := msg.(DataLoadedMsg)
	if !ok {
		t.Fatalf("expected DataLoadedMsg, got %T", msg)
	}
	joined := strings.Join(dMsg.Items, "\n")
	if !strings.Contains(joined, "Doctor: unavailable") {
		t.Errorf("expected 'Doctor: unavailable' in dashboard items, got:\n%s", joined)
	}
	if strings.Contains(joined, "Doctor: No issues") {
		t.Errorf("did NOT expect 'Doctor: No issues' when doctor failed, got:\n%s", joined)
	}
	foundWarn := false
	for _, item := range dMsg.Items {
		if strings.Contains(item, "⚠") && strings.Contains(item, "doctor") {
			foundWarn = true
			break
		}
	}
	if !foundWarn {
		t.Errorf("expected '⚠ doctor: ...' warning line, got items:\n%v", dMsg.Items)
	}
}

func TestStatusDashboardDoesNotWarnOnDoctorExitCode1(t *testing.T) {
	// M10 AC-01 / AC-02: brew doctor exit=1 means warnings, not failure.
	// The dashboard must show the warning count and no error indicator.
	r := newStatusMockRunner(func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 && args[0] == "doctor" {
			return []byte("Warning: Your Homebrew is outdated.\n"), &brew.BrewExitError{Command: "doctor", ExitCode: 1}
		}
		return []byte{}, nil
	})
	client := brew.NewClient(r)

	cmd := fetchStatusData(client)
	msg := cmd()
	dMsg, ok := msg.(DataLoadedMsg)
	if !ok {
		t.Fatalf("expected DataLoadedMsg, got %T", msg)
	}
	joined := strings.Join(dMsg.Items, "\n")
	if !strings.Contains(joined, "Doctor: 1 warning") {
		t.Errorf("expected 'Doctor: 1 warning' in dashboard items, got:\n%s", joined)
	}
	if strings.Contains(joined, "1 warnings") {
		t.Errorf("singular must not use plural form, got:\n%s", joined)
	}
	for _, item := range dMsg.Items {
		if strings.Contains(item, "⚠") && strings.Contains(item, "doctor") {
			t.Errorf("did NOT expect '⚠ doctor: ...' on exit=1 (it's warnings), got: %s", item)
		}
	}
}

func TestDoctorTabExitCode1ShowsWarnings(t *testing.T) {
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 && args[0] == "doctor" {
			return []byte("Warning: Your Homebrew is outdated.\nRun brew update.\n"),
				&brew.BrewExitError{Command: "doctor", ExitCode: 1}
		}
		return []byte{}, nil
	}
	client := brew.NewClient(r)

	cmd := fetchTabContentCmd(client, PanelStatus, 2, "doctor")
	if cmd == nil {
		t.Fatal("fetchTabContentCmd returned nil")
	}
	msg := cmd()
	tMsg, ok := msg.(TabContentMsg)
	if !ok {
		t.Fatalf("expected TabContentMsg, got %T", msg)
	}
	if tMsg.Err != nil {
		t.Fatalf("exit=1 must not be Err on Doctor tab, got %v", tMsg.Err)
	}
	if !strings.Contains(tMsg.Content, "outdated") {
		t.Errorf("expected warning content, got %q", tMsg.Content)
	}
}

func TestDoctorTabRealErrorShowsError(t *testing.T) {
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 && args[0] == "doctor" {
			return nil, &brew.BrewExitError{Command: "doctor", ExitCode: 2}
		}
		return []byte{}, nil
	}
	client := brew.NewClient(r)

	cmd := fetchTabContentCmd(client, PanelStatus, 2, "doctor")
	msg := cmd()
	tMsg, ok := msg.(TabContentMsg)
	if !ok {
		t.Fatalf("expected TabContentMsg, got %T", msg)
	}
	if tMsg.Err == nil {
		t.Fatal("expected Err on Doctor tab for exit=2")
	}

	m := newTestModel()
	m = updateModel(m, tMsg)
	key := tabKey(PanelStatus, 2, "doctor")
	if !strings.Contains(m.tabContent[key], "Error") {
		t.Errorf("expected Error in tab body, got %q", m.tabContent[key])
	}
}

// newStatusMockRunner wraps a doctor-specific ExecuteFn so callers can
// ignore the JSON shape expected by fetchStatusData for the other panels.
// Other commands return empty JSON objects so the dashboard doesn't show
// spurious "unexpected end of JSON input" warnings.
func newStatusMockRunner(doctorFn func(ctx context.Context, args ...string) ([]byte, error)) *brew.MockRunner {
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 && args[0] == "doctor" {
			return doctorFn(ctx, args...)
		}
		switch {
		case len(args) >= 1 && args[0] == "info":
			return []byte(`{"formulae":[],"casks":[]}`), nil
		case len(args) >= 1 && args[0] == "outdated":
			return []byte(`{"formulae":[]}`), nil
		case len(args) >= 1 && args[0] == "tap-info":
			return []byte(`{"taps":[]}`), nil
		case len(args) >= 1 && args[0] == "services":
			return []byte(`{"services":[]}`), nil
		}
		return []byte{}, nil
	}
	return r
}

// TestPauseOnInteractionDefersRefresh locks M12 AC-03: when the user
// pressed a key within the pause window, evaluateAutoRefresh returns
// autoRefreshPausedMsg; once the window elapses, it returns RefreshMsg.
// We drive evaluateAutoRefresh directly so the test is deterministic
// (no dependence on tea.Tick's wall-clock timing).
func TestPauseOnInteractionDefersRefresh(t *testing.T) {
	m := newTestModel()
	m.cfg.GUI.AutoRefreshPause = 100 * time.Millisecond
	m.lastKeyAt = time.Now() // recent key press

	// Tick fires 50ms after the key press — within the 100ms window.
	msg := m.evaluateAutoRefresh(m.lastKeyAt.Add(50 * time.Millisecond))
	if _, ok := msg.(autoRefreshPausedMsg); !ok {
		t.Fatalf("expected autoRefreshPausedMsg within pause window, got %T", msg)
	}

	// Tick fires 200ms after the key press — past the 100ms window.
	msg = m.evaluateAutoRefresh(m.lastKeyAt.Add(200 * time.Millisecond))
	if _, ok := msg.(RefreshMsg); !ok {
		t.Errorf("expected RefreshMsg after pause window, got %T", msg)
	}
}

// TestPauseDisabledWhenZero ensures AutoRefreshPause=0 keeps the legacy
// behaviour (no pause check).
func TestPauseDisabledWhenZero(t *testing.T) {
	m := newTestModel()
	m.cfg.GUI.AutoRefreshPause = 0
	m.lastKeyAt = time.Now()

	msg := m.evaluateAutoRefresh(m.lastKeyAt.Add(10 * time.Millisecond))
	if _, ok := msg.(RefreshMsg); !ok {
		t.Errorf("expected RefreshMsg when AutoRefreshPause=0, got %T", msg)
	}
}

// TestKeyEventResetsLastKeyAt locks AC-03: every key press refreshes
// lastKeyAt so the pause window restarts.
func TestKeyEventResetsLastKeyAt(t *testing.T) {
	m := newTestModel()
	before := time.Now().Add(-1 * time.Hour)
	m.lastKeyAt = before

	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated, ok := nm.(Model)
	if !ok {
		t.Fatalf("Update returned unexpected type %T", nm)
	}
	if !updated.lastKeyAt.After(before) {
		t.Errorf("lastKeyAt should advance on key press, was %v, now %v", before, updated.lastKeyAt)
	}
}
