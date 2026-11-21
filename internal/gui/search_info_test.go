package gui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// searchFormulaInstalled is `brew info --json=v2 jq` for an installed formula,
// trimmed to the fields the search info reads.
const searchFormulaInstalled = `{
  "formulae": [{
    "name": "jq",
    "full_name": "jq",
    "tap": "homebrew/core",
    "desc": "Lightweight and flexible command-line JSON processor",
    "license": "MIT",
    "homepage": "https://jqlang.github.io/jq/",
    "versions": {"stable": "1.8.2", "head": "HEAD", "bottle": true},
    "bottle": {"stable": {"rebuild": 1, "root_url": "https://ghcr.io/v2/homebrew/core",
      "files": {"arm64_linux": {"cellar": ":any", "url": "https://ghcr.io/v2/homebrew/core/jq/blobs/sha256:d67a", "sha256": "d67a"}}}},
    "keg_only": false,
    "dependencies": ["oniguruma"],
    "build_dependencies": [],
    "caveats": null,
    "installed": [{
      "version": "1.8.2",
      "used_options": [],
      "built_as_bottle": true,
      "poured_from_bottle": true,
      "time": 1793793264,
      "runtime_dependencies": [{"full_name": "oniguruma", "version": "6.9.10"}],
      "installed_on_request": true
    }],
    "pinned": false,
    "outdated": false
  }],
  "casks": []
}`

// searchFormulaAvailable is the same shape for a formula that is not installed:
// no installed entries, so the stable version is the one to show.
const searchFormulaAvailable = `{
  "formulae": [{
    "name": "ripgrep",
    "full_name": "ripgrep",
    "tap": "homebrew/core",
    "desc": "Search tool like grep and The Silver Searcher",
    "license": "Unlicense OR MIT",
    "homepage": "https://github.com/BurntSushi/ripgrep",
    "versions": {"stable": "14.1.1", "head": "HEAD", "bottle": true},
    "bottle": {"stable": {"rebuild": 0, "root_url": "https://ghcr.io/v2/homebrew/core",
      "files": {"arm64_linux": {"cellar": ":any_skip_relocation", "url": "https://ghcr.io/v2/homebrew/core/ripgrep/blobs/sha256:aa", "sha256": "aa"}}}},
    "keg_only": false,
    "dependencies": [],
    "build_dependencies": [],
    "caveats": null,
    "installed": [],
    "pinned": false,
    "outdated": false
  }],
  "casks": []
}`

// searchCaskInstalled is `brew info --json=v2 --cask` for an installed cask,
// trimmed to the fields the search info reads. Brew writes the installed
// version as a string.
const searchCaskInstalled = `{
  "formulae": [],
  "casks": [{
    "token": "claude-code",
    "full_token": "claude-code",
    "tap": "homebrew/cask",
    "name": ["Claude Code"],
    "desc": "Terminal-based AI coding assistant",
    "homepage": "https://claude.com/product/claude-code",
    "version": "2.1.311",
    "installed": "2.1.311",
    "outdated": false,
    "pinned": false,
    "auto_updates": null,
    "artifacts": [{"binary": ["claude"], "target": "/home/linuxbrew/.linuxbrew/bin/claude"}],
    "depends_on": {}
  }]
}`

// searchInfoView is the search info pane for a payload, with ANSI stripped.
func searchInfoView(t *testing.T, payload string) string {
	t.Helper()
	m := updateModel(newTestModel(), SearchInfoLoadedMsg{Content: payload})
	return ansi.Strip(m.renderSearchInfo(100, 40))
}

// The formula's version and bottle come from the parsers, not from JSON tags
// that brew's payload does not have at the top level.
func TestSearchInfoFormulaInstalledShowsVersionBottleAndStatus(t *testing.T) {
	info, err := parsePackageInfo(searchFormulaInstalled)
	if err != nil {
		t.Fatalf("parsePackageInfo: %v", err)
	}
	if info.Version != "1.8.2" {
		t.Errorf("Version = %q, want 1.8.2", info.Version)
	}
	if !info.Bottled {
		t.Error("Bottled = false, want true (the stable bottle has files)")
	}
	if !info.Installed {
		t.Error("Installed = false, want true (brew lists an installed entry)")
	}

	view := searchInfoView(t, searchFormulaInstalled)
	for _, want := range []string{"1.8.2 (bottled)", "installed"} {
		if !strings.Contains(view, want) {
			t.Errorf("search info should show %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "not installed") {
		t.Errorf("an installed formula must not read as not installed:\n%s", view)
	}
}

func TestSearchInfoFormulaAvailableShowsStableVersionNotInstalled(t *testing.T) {
	info, err := parsePackageInfo(searchFormulaAvailable)
	if err != nil {
		t.Fatalf("parsePackageInfo: %v", err)
	}
	if info.Version != "14.1.1" {
		t.Errorf("Version = %q, want the stable version 14.1.1", info.Version)
	}
	if info.Installed {
		t.Error("Installed = true, want false (no installed entries)")
	}

	view := searchInfoView(t, searchFormulaAvailable)
	for _, want := range []string{"14.1.1 (bottled)", "not installed"} {
		if !strings.Contains(view, want) {
			t.Errorf("search info should show %q:\n%s", want, view)
		}
	}
}

// The cask's installed version is a string in brew's payload. The pane must
// read it as installed, with the version brew gives.
func TestSearchInfoCaskInstalledShowsVersionAndStatus(t *testing.T) {
	info, err := parsePackageInfo(searchCaskInstalled)
	if err != nil {
		t.Fatalf("parsePackageInfo: %v", err)
	}
	if info.Type != "cask" {
		t.Errorf("Type = %q, want cask", info.Type)
	}
	if info.Version != "2.1.311" {
		t.Errorf("Version = %q, want 2.1.311", info.Version)
	}
	if !info.Installed {
		t.Error("Installed = false, want true (brew writes an installed version)")
	}

	view := searchInfoView(t, searchCaskInstalled)
	for _, want := range []string{"2.1.311", "installed"} {
		if !strings.Contains(view, want) {
			t.Errorf("search info should show %q:\n%s", want, view)
		}
	}
}

// A failed `brew info` shows brew's own error line in the pane, not a parse
// error of the message (C11). brew's exit error carries only the exit code;
// the line that says what went wrong is in its stderr.
func TestSearchInfoBrewErrorIsShownAsBrewSaysIt(t *testing.T) {
	runner := brewtest.NewMockRunner()
	runner.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return nil, &brew.BrewExitError{Command: "info --json=v2 zz", ExitCode: 1,
			Stderr: "Error: No available formula with the name \"zz\".\n"}
	}
	m := New(brew.NewClient(runner), testConfig())
	m = updateModel(m, tea.WindowSizeMsg{Width: 100, Height: 40})
	m = sendKey(m, "7")
	m = updateModel(m, SearchDoneMsg{Results: []string{"zz"}, Raw: []brew.SearchResult{{Name: "zz"}}})
	m = updateModel(m, m.fetchSelectedSearchInfo()())

	view := ansi.Strip(m.renderSearchInfo(100, 40))
	if !strings.Contains(view, `Error: No available formula with the name "zz".`) {
		t.Fatalf("the pane should show brew's error line, got:\n%s", view)
	}
	if strings.Contains(view, "Parse error") {
		t.Fatalf("a brew error must not read as a parse error:\n%s", view)
	}
}
