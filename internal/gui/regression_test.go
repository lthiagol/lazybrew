package gui

import (
	"testing"

	"github.com/thiago/lazybrew/internal/brew"
	"github.com/thiago/lazybrew/internal/config"
)

func TestTabContentChangesWithSelection_KeyDiffers(t *testing.T) {
	k1 := tabKey(PanelFormulae, 1, "ripgrep")
	k2 := tabKey(PanelFormulae, 1, "neovim")
	if k1 == k2 {
		t.Error("tab keys for different items must differ")
	}
	if tabKey(PanelFormulae, 1, "ripgrep") != k1 {
		t.Error("tab key must be deterministic")
	}
}

func TestTabContentChangesWithSelection_Refetch(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	p := m.panels[PanelFormulae]
	p.items = []string{"ripgrep  14.1.1  bottled", "neovim  0.10.4  bottled"}
	p.selected = 0

	m.activeTab = 1

	cmd := m.loadTabContent()
	if cmd == nil {
		t.Fatal("expected fetch cmd for unvisited tab")
	}
	msg := cmd()
	tc, ok := msg.(TabContentMsg)
	if !ok {
		t.Fatalf("expected TabContentMsg, got %T", msg)
	}
	if tc.ItemName != "ripgrep" {
		t.Errorf("expected item name ripgrep, got %q", tc.ItemName)
	}

	m = updateModel(m, tc)

	p.selected = 1
	cmd = m.loadTabContent()
	if cmd == nil {
		t.Fatal("expected fetch cmd for different item (stale cache)")
	}
}

func TestBatchUpgradeCallsBrewWithSelectedNames(t *testing.T) {
	cfg := config.Default()
	client := brew.NewClient(brew.NewMockRunner())
	m := New(client, cfg)
	m.activePanel = PanelOutdated
	p := m.panels[PanelOutdated]
	p.items = []string{"ripgrep  1.0 -> 2.0", "neovim  0.9 -> 0.10"}
	p.formulae = []brew.Formula{
		{Name: "ripgrep", Version: "1.0", NewVersion: "2.0", Outdated: true},
		{Name: "neovim", Version: "0.9", NewVersion: "0.10", Outdated: true},
	}
	p.selected = 0

	m.batch.selected = map[int]bool{0: true, 1: true}

	tm2, cmd := m.batchUpgrade()
	if cmd == nil {
		t.Fatal("expected cmd from batchUpgrade")
	}
	rm := tm2.(Model)
	if rm.batchCount != 2 {
		t.Errorf("expected batchCount=2, got %d", rm.batchCount)
	}
}

func TestPinRespectsPinnedFlag(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	p := m.panels[PanelFormulae]
	p.items = []string{"ripgrep  14.1.1  bottled"}
	p.formulae = []brew.Formula{
		{Name: "ripgrep", Version: "14.1.1", Pinned: true},
	}
	p.selected = 0

	_, cmd := m.togglePin(mutInstall)
	if cmd == nil {
		t.Fatal("expected cmd from togglePin")
	}

	msg := cmd()
	if _, ok := msg.(TaskStartedMsg); !ok {
		t.Fatalf("expected TaskStartedMsg, got %T", msg)
	}
}


