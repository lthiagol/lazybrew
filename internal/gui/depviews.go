package gui

import (
	"fmt"
	"path"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// The terminal states of the dependency and file tabs.
const (
	noDependencies   = "No dependencies"
	noDependents     = "No dependents"
	noFilesInstalled = "No files installed"
)

// depCheck is the installed list a dependency row is checked against.
type depCheck int

const (
	// depNone checks nothing: a requirement is not a package, so its row has
	// no installed icon.
	depNone depCheck = iota
	// depFormula checks the installed formulae.
	depFormula
	// depCask checks the installed casks.
	depCask
)

// depGroup is one heading and its names in a dependency list.
type depGroup struct {
	heading string
	names   []string
	check   depCheck
}

// listedCheck returns whether a name is in the installed list the check reads.
// It returns nil when that list has not loaded, so no row claims an icon it
// cannot back up.
func (m Model) listedCheck(c depCheck) func(string) bool {
	listed := make(map[string]bool)
	switch c {
	case depFormula:
		p := m.panels[PanelFormulae]
		if !p.loaded {
			return nil
		}
		for _, f := range p.formulae {
			listed[f.Name] = true
		}
	case depCask:
		p := m.panels[PanelCasks]
		if !p.loaded {
			return nil
		}
		for _, c := range p.casks {
			listed[c.Name] = true
		}
	default:
		return nil
	}
	return func(name string) bool { return listed[name] }
}

// renderDepGroups draws each non-empty group: its heading, then one row per
// name. It returns "" when every group is empty.
func (m Model) renderDepGroups(groups []depGroup) string {
	var blocks []string
	for _, g := range groups {
		if len(g.names) == 0 {
			continue
		}
		listed := m.listedCheck(g.check)
		var rows []string
		if g.heading != "" {
			rows = append(rows, style.SubtleText.Render(g.heading))
		}
		for _, name := range g.names {
			rows = append(rows, depRow(name, listed, m.icons))
		}
		blocks = append(blocks, strings.Join(rows, "\n"))
	}
	return strings.Join(blocks, "\n\n")
}

// depRow is one dependency: an installed or not-installed icon, then the name.
// With icons off the state is written out after the name instead.
func depRow(name string, listed func(string) bool, icons style.Icons) string {
	if listed == nil {
		return "  " + name
	}
	mark, role, word := icons.Absent, style.MutedColor, "not installed"
	if listed(name) {
		mark, role, word = icons.Installed, style.SuccessColor, "installed"
	}
	if mark == "" {
		return "  " + name + " " + lipgloss.NewStyle().Foreground(role).Render("("+word+")")
	}
	return "  " + lipgloss.NewStyle().Foreground(role).Render(mark) + " " + name
}

// formulaDepsView is the Deps tab of a formula: its runtime and build
// dependencies, each with its installed state.
func (m Model) formulaDepsView(f *brew.Formula) string {
	view := m.renderDepGroups([]depGroup{
		{heading: "Runtime", names: f.Dependencies, check: depFormula},
		{heading: "Build", names: f.BuildDeps, check: depFormula},
	})
	if view == "" {
		return style.SubtleText.Render(noDependencies)
	}
	return view
}

// renderUses is the Used By tab: the installed formulae that depend on the
// selected one. `brew uses --installed` lists only installed packages, so
// each row is checked against the installed formulae like any other.
func (m Model) renderUses(content string) string {
	if content == noDependents {
		return style.SubtleText.Render(content)
	}
	return m.renderDepGroups([]depGroup{{names: lines(content), check: depFormula}})
}

// renderFiles is the Files tab: a count of the files `brew list` printed, then
// each path with its directory in the Muted role and its file name in Text.
func renderFiles(content string) string {
	paths := lines(content)
	if content == noFilesInstalled || len(paths) == 0 {
		return style.SubtleText.Render(noFilesInstalled)
	}
	rows := []string{style.SubtleText.Render(fmt.Sprintf("%d %s", len(paths), plural(len(paths), "file", "files")))}
	muted := lipgloss.NewStyle().Foreground(style.MutedColor)
	text := lipgloss.NewStyle().Foreground(style.TextColor)
	for _, p := range paths {
		dir, file := path.Split(p)
		rows = append(rows, muted.Render(dir)+text.Render(file))
	}
	return strings.Join(rows, "\n")
}

// lines is the non-empty, trimmed lines of a fetched list.
func lines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}
