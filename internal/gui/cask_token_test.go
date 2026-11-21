package gui

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// casksFromBrewPayload reads the casks of brew's own `info --json=v2 --cask
// --installed` output, the same bytes the Casks list is built from.
func casksFromBrewPayload(t *testing.T) []brew.Cask {
	t.Helper()
	raw, err := os.ReadFile("../brew/testdata/casks_installed_token.json")
	if err != nil {
		t.Fatal(err)
	}
	_, casks, err := brew.ParseInfo(raw)
	if err != nil {
		t.Fatal(err)
	}
	return casks
}

// TestCaskUpgradeSendsTokenFromBrewPayload (LB-12, Q-cr1-cask-name): the Casks
// row built from brew's payload carries the token, so upgrade asks brew for
// claude-code. Before the fix it sent "Claude Code", which brew rejects.
func TestCaskUpgradeSendsTokenFromBrewPayload(t *testing.T) {
	var gotArgs [][]string
	r := brewtest.NewMockRunner()
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		gotArgs = append(gotArgs, args)
		ch := make(chan string)
		close(ch)
		errCh := make(chan error, 1)
		close(errCh)
		return ch, errCh
	}
	m := New(brew.NewClient(r), testConfig())
	m.activePanel = PanelCasks
	m.panels[PanelCasks].casks = casksFromBrewPayload(t)
	m.panels[PanelCasks].items = []string{"claude-code"}
	m.panels[PanelCasks].selected = 0

	nm, _ := m.confirmMutation(mutUpgrade, "Upgrade")
	mm := nm.(Model)
	if mm.activeModal == nil {
		t.Fatal("upgrade must open its confirm")
	}
	if view := mm.activeModal.View().Content; !strings.Contains(view, "Upgrade claude-code?") {
		t.Fatalf("confirm must name the token, got view:\n%s", view)
	}
	mm.activeModal = nil
	res, cmd := mm.handleModalResult(&modal.ConfirmResult{Confirmed: true}, nil)
	nm2 := res.(Model)
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatal("expected the upgrade to be enqueued")
	}
	if msg := nm2.tasks.RunNext()(); msg.(TaskCompletedMsg).Err != nil {
		t.Fatal("unexpected task error")
	}
	want := []string{"upgrade", "--cask", "claude-code"}
	for _, args := range gotArgs {
		if reflect.DeepEqual(args, want) {
			return
		}
	}
	t.Fatalf("expected brew args %v, got %v", want, gotArgs)
}

// TestCaskInstalledMapKeysByToken (LB-12): a formula's or cask's depends_on
// names a cask by its token, so the Casks installed map must key on it too.
func TestCaskInstalledMapKeysByToken(t *testing.T) {
	p := &panelData{id: PanelCasks, casks: casksFromBrewPayload(t)}
	p.indexInstalled()
	if _, ok := p.installed["claude-code"]; !ok {
		t.Fatalf("installed map must key on the token, keys = %v", p.installed)
	}
	if _, ok := p.installed["Claude Code"]; ok {
		t.Fatal("installed map must not key on the display name")
	}
}
