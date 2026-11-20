package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// formulaWithDeps opens the Formulae panel with bash selected, whose runtime
// dependency ncurses is installed and whose build dependency cmake is not.
func formulaWithDeps(t *testing.T) *Model {
	t.Helper()
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "2")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: []string{"bash", "ncurses"},
		Formulae: []brew.Formula{
			{Name: "bash", Version: "5.2", Dependencies: []string{"ncurses", "readline"}, BuildDeps: []string{"cmake"}},
			{Name: "ncurses", Version: "6.5"},
		}})
	return m
}

// TestFormulaDepsShowInstalledState (v2-05 AC-02.1): each dependency row
// carries an installed or not-installed icon, under Runtime and Build headings.
func TestFormulaDepsShowInstalledState(t *testing.T) {
	m := formulaWithDeps(t)
	m = sendKey(m, "]")
	out := ansi.Strip(m.renderContent(70, 20))
	for _, want := range []string{"Runtime", "Build", "✓ ncurses", "○ readline", "○ cmake"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the Deps tab should show %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "✓ readline") || strings.Contains(out, "○ ncurses") {
		t.Fatalf("installed state is per package, got:\n%s", out)
	}
}

// TestDepIconsWaitForTheInstalledList: before the Formulae list has loaded,
// no dependency row claims to be installed or not.
func TestDepIconsWaitForTheInstalledList(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "2")
	// The selected formula is on screen, but the installed list has not
	// loaded: the switch left the panel loading, so clear that too.
	p := m.panels[PanelFormulae]
	p.loading, p.loaded = false, false
	p.items = []string{"bash"}
	p.formulae = []brew.Formula{{Name: "bash", Dependencies: []string{"ncurses"}}}
	m = sendKey(m, "]")
	out := ansi.Strip(m.renderContent(70, 20))
	if strings.Contains(out, "✓") || strings.Contains(out, "○") {
		t.Fatalf("no icon should show until the installed list loads, got:\n%s", out)
	}
	if !strings.Contains(out, "ncurses") {
		t.Fatalf("the names should still show, got:\n%s", out)
	}
}

// TestDepIconsWriteTheStateWhenIconsAreOff: with show_icons off the state is
// spelled out after each name.
func TestDepIconsWriteTheStateWhenIconsAreOff(t *testing.T) {
	m := formulaWithDeps(t)
	m.icons = style.Icons{}
	m = sendKey(m, "]")
	out := ansi.Strip(m.renderContent(70, 20))
	if !strings.Contains(out, "ncurses (installed)") || !strings.Contains(out, "cmake (not installed)") {
		t.Fatalf("with icons off the state should be written out, got:\n%s", out)
	}
}

// TestCaskDepsGroupFormulaeCasksAndRequirements (v2-05 AC-02.2): the Deps tab
// of a cask lists the formulae, the casks and the requirements, and only the
// first two carry an installed icon.
func TestCaskDepsGroupFormulaeCasksAndRequirements(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "3")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: []string{"git"},
		Formulae: []brew.Formula{{Name: "git", Version: "2.50"}}})
	m = sendKey(m, "3")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Items: []string{"gcloud-cli"},
		Casks: []brew.Cask{{
			Name:              "gcloud-cli",
			DependsOnFormulae: []string{"python@3.14"},
			DependsOnCasks:    []string{"docker"},
			Requirements:      []string{"macOS ≥ 13"},
		}}})
	m = sendKey(m, "]")
	out := ansi.Strip(m.renderContent(70, 20))
	for _, want := range []string{"Formulae", "○ python@3.14", "Casks", "○ docker", "Requirements", "  macOS ≥ 13"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the cask Deps tab should show %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "○ macOS") || strings.Contains(out, "✓ macOS") {
		t.Fatalf("a requirement is not a package and carries no icon, got:\n%s", out)
	}
}

func TestUsedByListsInstalledDependents(t *testing.T) {
	m := formulaWithDeps(t)
	m = sendKey(m, "]")
	m = sendKey(m, "]")
	m.tabContent[tabKey(PanelFormulae, formulaUsesTab, "bash")] = "ncurses\n"
	if out := ansi.Strip(m.renderContent(70, 20)); !strings.Contains(out, "✓ ncurses") {
		t.Fatalf("the Used By tab should list the installed dependent, got:\n%s", out)
	}
	m.tabContent[tabKey(PanelFormulae, formulaUsesTab, "bash")] = noDependents
	if out := ansi.Strip(m.renderContent(70, 20)); !strings.Contains(out, noDependents) {
		t.Fatalf("the terminal state should read %q, got:\n%s", noDependents, out)
	}
}

// TestFilesCountAndStyleThePaths (v2-05 AC-02.3): the Files tab leads with the
// count, and each path draws its file name after its directory.
func TestFilesCountAndStyleThePaths(t *testing.T) {
	content := "/opt/homebrew/Cellar/bash/5.2/bin/bash\n/opt/homebrew/Cellar/bash/5.2/share/man/man1/bash.1\n"
	out := renderFiles(content)
	plain := ansi.Strip(out)
	lines := strings.Split(plain, "\n")
	if lines[0] != "2 files" {
		t.Fatalf("the first line should count the files, got %q", lines[0])
	}
	if lines[1] != "/opt/homebrew/Cellar/bash/5.2/bin/bash" {
		t.Fatalf("the path should be kept whole, got %q", lines[1])
	}
	if !strings.Contains(out, "\x1b[") {
		t.Fatal("the directory and the file name should be styled apart")
	}
	if got := ansi.Strip(renderFiles(noFilesInstalled)); got != noFilesInstalled {
		t.Fatalf("an empty list should say %q, got %q", noFilesInstalled, got)
	}
}

// TestSearchResultDependenciesCheckInstalled (v2-05 AC-02.1): a search result's
// dependencies read as Runtime and Build, each with its installed state.
func TestSearchResultDependenciesCheckInstalled(t *testing.T) {
	m := formulaWithDeps(t)
	m = sendKey(m, "7")
	m.searchInfoContent = `{"formulae":[{"name":"wget","versions":{"stable":"1.25"},` +
		`"dependencies":["ncurses"],"build_dependencies":["cmake"]}],"casks":[]}`
	out := ansi.Strip(m.renderContent(70, 30))
	for _, want := range []string{"Dependencies:", "Runtime", "✓ ncurses", "Build", "○ cmake"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the search result should show %q, got:\n%s", want, out)
		}
	}
}
