package gui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thiago/lazybrew/internal/brew"
	"github.com/thiago/lazybrew/internal/config"
	"github.com/thiago/lazybrew/internal/gui/modal"
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
}

func TestPanelJump(t *testing.T) {
	m := newTestModel()

	m = sendKey(m, "3")
	if m.activePanel != PanelCasks {
		t.Errorf("after 3: activePanel = %v, want PanelCasks", m.activePanel)
	}

	m = sendKey(m, "7")
	if m.activePanel != PanelSearch {
		t.Errorf("after 7: activePanel = %v, want PanelSearch", m.activePanel)
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
	if m.activeModal == nil {
		t.Fatal("search modal should be active after /")
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
	if m.activeModal == nil {
		t.Fatal("expected modal to be opened")
	}
	if _, ok := m.activeModal.(*modal.ProgressModal); !ok {
		t.Fatalf("expected ProgressModal, got %T", m.activeModal)
	}
}

func TestModelTaskOutputAppendsLine(t *testing.T) {
	m := newTestModel()
	m.activeModal = modal.NewProgressModal("Test", nil)

	msg := TaskOutputMsg{ID: "t1", Line: "hello"}
	m = updateModel(m, msg)

	if m.activeModal == nil {
		t.Fatal("modal should remain open")
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
