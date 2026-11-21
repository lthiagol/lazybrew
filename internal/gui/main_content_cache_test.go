package gui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// filesTabModel is a formula with a Files tab of n installed files, at 140x40.
func filesTabModel(t *testing.T, n int) *Model {
	t.Helper()
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: []string{"bash"},
		Formulae: []brew.Formula{{Name: "bash", Version: "5.2"}}})
	m = sendKey(m, "2")
	m.activeTab = formulaFilesTab
	m.tabContent[tabKey(PanelFormulae, formulaFilesTab, "bash")] = filePaths(n)
	return m
}

func filePaths(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "/home/linuxbrew/.linuxbrew/Cellar/bash/5.2/share/file%04d.txt\n", i)
	}
	return b.String()
}

// TestMainRowsMatchTheUncachedRender (cr3 AC-04): the cached rows of a Files tab
// are the rows a fresh render and wrap gives, at every point of a scroll.
func TestMainRowsMatchTheUncachedRender(t *testing.T) {
	m := filesTabModel(t, 300)
	for i := 0; i < 3; i++ {
		m.View()
		want := wrapContent(m.renderContent(contentWidthIn(m.screenLayout()), 30), contentWidthIn(m.screenLayout()))
		if got := m.mainRows(contentWidthIn(m.screenLayout()), 30); strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Fatalf("pass %d: cached rows differ from the uncached render", i)
		}
		m = wheelDown(m)
	}
}

// TestScrollingFilesTabRendersOnce (cr3 AC-04): the wheel and G over a
// Files tab of 5000 files render and wrap it once, not once per key.
func TestScrollingFilesTabRendersOnce(t *testing.T) {
	m := filesTabModel(t, 5000)
	m.View()
	renders := m.wrap.renders
	if renders == 0 {
		t.Fatal("the first frame should render the tab once")
	}
	for i := 0; i < 50; i++ {
		m = wheelDown(m)
		m.View()
	}
	m = sendKey(m, "G")
	m.View()
	m = sendKey(m, "g")
	m.View()
	if m.wrap.renders != renders {
		t.Fatalf("scrolling rendered the tab %d more times, want none", m.wrap.renders-renders)
	}
}

// TestFilesTabFollowsItsContent: a new tab content shows at once, not the one
// cached before it.
func TestFilesTabFollowsItsContent(t *testing.T) {
	m := filesTabModel(t, 300)
	if out := ansi.Strip(m.View().Content); !strings.Contains(out, "300 files") {
		t.Fatalf("the count should read 300 files, got:\n%s", out)
	}
	m = updateModel(m, TabContentMsg{PanelID: PanelFormulae, TabIndex: formulaFilesTab, ItemName: "bash",
		Content: filePaths(3)})
	if out := ansi.Strip(m.View().Content); !strings.Contains(out, "3 files") || strings.Contains(out, "300 files") {
		t.Fatalf("the Files tab should follow its new content, got:\n%s", out)
	}
}

// TestUsesTabFollowsTheInstalledList: a formula's Used By rows read the installed
// formulae, so a reload that removes one flips its row at once.
func TestUsesTabFollowsTheInstalledList(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: []string{"bash", "ncurses"},
		Formulae: []brew.Formula{{Name: "bash", Version: "5.2"}, {Name: "ncurses", Version: "6.5"}}})
	m = sendKey(m, "2")
	m.activeTab = formulaUsesTab
	m.tabContent[tabKey(PanelFormulae, formulaUsesTab, "bash")] = "ncurses\n"
	if out := ansi.Strip(m.View().Content); !hasRow(out, "✔", "ncurses") {
		t.Fatalf("ncurses is installed, its row should read installed, got:\n%s", out)
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: []string{"bash"},
		Formulae: []brew.Formula{{Name: "bash", Version: "5.2"}}})
	// The reload clears the tab's content; its fetch lands again.
	m.tabContent[tabKey(PanelFormulae, formulaUsesTab, "bash")] = "ncurses\n"
	if out := ansi.Strip(m.View().Content); !hasRow(out, "○", "ncurses") {
		t.Fatalf("ncurses is gone after the reload, its row should read not installed, got:\n%s", out)
	}
}

// TestInstalledListBuiltWhenDataLands (cr3 AC-04): the installed list of the
// formulae is built when the panel's data lands, not on each render.
func TestInstalledListBuiltWhenDataLands(t *testing.T) {
	m := filesTabModel(t, 10)
	gen := m.panels[PanelFormulae].installedGen
	if gen == 0 {
		t.Fatal("the data landing should build the installed list")
	}
	for i := 0; i < 20; i++ {
		m.View()
	}
	if got := m.panels[PanelFormulae].installedGen; got != gen {
		t.Fatalf("rendering rebuilt the installed list %d times", got-gen)
	}
}
