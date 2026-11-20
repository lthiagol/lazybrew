package gui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// lineWith is the line of out that holds want, and the cell column want starts
// at within that line. It fails the test when want is not on any line.
func lineWith(t *testing.T, out, want string) (line string, col int) {
	t.Helper()
	for _, l := range strings.Split(out, "\n") {
		if at := strings.Index(l, want); at >= 0 {
			return l, ansi.StringWidth(l[:at])
		}
	}
	t.Fatalf("%q should be shown, got:\n%s", want, out)
	return "", 0
}

// TestTapInfoSplitsWhenItFits (v2-09 AC-03.1, Q-v2-09-a, F-r4-01): at the split
// width (82 cells of content) a short remote fits the right column whole, so Tap
// Info splits: the fields sit left and the right column (Remote, then the counts)
// sits beyond them.
func TestTapInfoSplitsWhenItFits(t *testing.T) {
	p := &panelData{taps: []brew.Tap{{
		Name: "lthiagol/tap", Remote: "https://x.io/t",
		FormulaCount: 3, CaskCount: 1,
	}}}
	wide := ansi.Strip(renderTapInfo(p, 82))
	_, col := lineWith(t, wide, "Formulae:")
	if col < 2*82/5 {
		t.Fatalf("at 82 cells Formulae should sit in the right column, at cell %d:\n%s", col, wide)
	}
	narrow := ansi.Strip(renderTapInfo(p, 81))
	for _, line := range strings.Split(narrow, "\n") {
		if strings.Contains(line, "Tap:") && strings.Contains(line, "Formulae:") {
			t.Fatalf("at 81 cells the view is one column, got %q", line)
		}
	}
}

// TestTapInfoKeepsARemoteWhole (F-r4-01): a remote too long for the right column
// keeps the single column at 87 cells, so the URL is never broken across rows.
func TestTapInfoKeepsARemoteWhole(t *testing.T) {
	const remote = "https://github.com/danielfoehrkn/homebrew-switch"
	p := &panelData{taps: []brew.Tap{{Name: "danielfoehrkn/switch", Remote: remote, FormulaCount: 0}}}
	out := ansi.Strip(renderTapInfo(p, 87))
	if !strings.Contains(out, remote) {
		t.Fatalf("at 87 cells the remote should be whole on one row, got:\n%s", out)
	}
	if line, _ := lineWith(t, out, "Remote:"); strings.Contains(line, "Tap:") {
		t.Fatalf("at 87 cells Tap Info should keep the single column, got %q", line)
	}
}

// TestServicesInfoKeepsPathsWhole (F-r4-01): with long paths, Services Info at 87
// cells keeps its single column and every path is whole on its row.
func TestServicesInfoKeepsPathsWhole(t *testing.T) {
	const (
		unit    = "/home/agent/.config/systemd/user/sh.brew.moshi-hook.service"
		command = "/home/linuxbrew/.linuxbrew/opt/moshi-hook/bin/moshi-hook serve"
		logPath = "/home/linuxbrew/.linuxbrew/var/log/moshi-hook.log"
	)
	m := selectService(loadServices(newTestModel(), servicesList(), servicesInfo()), "moshi-hook")
	p := m.panels[PanelServices]
	for i := range p.serviceInfo {
		if p.serviceInfo[i].Name == "moshi-hook" {
			p.serviceInfo[i].File = unit
			p.serviceInfo[i].Command = command
			p.serviceInfo[i].LogPath = logPath
		}
	}
	out := infoText(m, 87)
	for _, path := range []string{unit, command, logPath} {
		if !strings.Contains(out, path) {
			t.Fatalf("at 87 cells the path %q should be whole on one row, got:\n%s", path, out)
		}
	}
	if line, _ := lineWith(t, out, "Command:"); strings.Contains(line, "Name:") {
		t.Fatalf("at 87 cells Services Info should keep the single column, got %q", line)
	}
}

// TestInfoSplitsAt140x40 (Q-v2-09-a): the default 140x40 layout is wide enough
// for the Info views to split, so the description sits beside the name.
func TestInfoSplitsAt140x40(t *testing.T) {
	m := formulaWithDeps(t)
	p := m.panels[PanelFormulae]
	p.formulae[0].Description = "GNU Bourne-Again SHell"
	out := ansi.Strip(m.renderContent(m.mainContentWidth(), 20))
	first := strings.Split(out, "\n")[0]
	if !strings.Contains(first, "Name:") || !strings.Contains(first, "Description:") {
		t.Fatalf("at 140x40 the Info should split, first row %q in:\n%s", first, out)
	}
}
