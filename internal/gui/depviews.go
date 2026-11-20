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

// installedPkg is what a dependency table knows of an installed package: its
// installed version and description, both from the loaded lists.
type installedPkg struct {
	version, desc string
}

// installedList is the installed packages a check reads, by name: the list the
// panel built when its data landed. It returns nil when that list has not
// loaded, so no row claims an icon it cannot back up.
func (m Model) installedList(c depCheck) map[string]installedPkg {
	var p *panelData
	switch c {
	case depFormula:
		p = m.panels[PanelFormulae]
	case depCask:
		p = m.panels[PanelCasks]
	default:
		return nil
	}
	if !p.loaded {
		return nil
	}
	return p.installed
}

// indexInstalled builds the installed list of a formulae or casks panel from its
// packages. It runs when the panel's data lands, and when the leaves filter
// narrows the formulae, so a render reads the list rather than building it.
func (p *panelData) indexInstalled() {
	switch p.id {
	case PanelFormulae:
		p.installed = make(map[string]installedPkg, len(p.formulae))
		for _, f := range p.formulae {
			p.installed[f.Name] = installedPkg{version: f.Version, desc: f.Description}
		}
	case PanelCasks:
		p.installed = make(map[string]installedPkg, len(p.casks))
		for _, c := range p.casks {
			p.installed[c.Name] = installedPkg{version: c.Version, desc: c.Description}
		}
	default:
		return
	}
	p.installedGen++
}

// renderDepGroups draws the groups as one table at most width cells wide: a
// header row, then each non-empty group's heading and its rows. It returns ""
// when every group is empty.
func (m Model) renderDepGroups(groups []depGroup, width int) string {
	var rows []tableRow
	for _, g := range groups {
		if len(g.names) == 0 {
			continue
		}
		if len(rows) > 0 {
			rows = append(rows, tableRow{})
		}
		if g.heading != "" {
			rows = append(rows, tableRow{heading: g.heading})
		}
		rows = append(rows, m.pkgRows(g.names, g.check)...)
	}
	if len(rows) == 0 {
		return ""
	}
	return renderTable(m.pkgColumns(), rows, width)
}

// pkgColumns is the column set of a package table: the installed icon when
// icons are on, the name, then the installed version and the description. The
// description is the first to go when the table is narrow, then the version.
func (m Model) pkgColumns() []tableCol {
	cols := []tableCol{{head: ""}, {head: "Name"}, {head: "Version", drop: 1}, {head: "Description", drop: 2}}
	if m.iconsOff() {
		return cols[1:]
	}
	return cols
}

// iconsOff is whether the installed state is written out after each name, not
// drawn as an icon.
func (m Model) iconsOff() bool {
	return m.icons.Absent == ""
}

// pkgRows is one row per package name in a package table. The installed list
// is read once for the whole table, not once per row (v2-09 F-r1-03).
func (m Model) pkgRows(names []string, check depCheck) []tableRow {
	list := m.installedList(check)
	rows := make([]tableRow, 0, len(names))
	for _, name := range names {
		rows = append(rows, m.pkgRow(name, list))
	}
	return rows
}

// pkgRow is one package in a package table, its cells in the order pkgColumns
// gives. list is the installed list the table reads (nil when it has not
// loaded). The state is an installed or not-installed icon, or the word after
// the name when icons are off. Version and description are empty when unknown.
func (m Model) pkgRow(name string, list map[string]installedPkg) tableRow {
	pkg, installed := list[name]
	var cells []string
	if m.iconsOff() {
		if list != nil {
			word := "not installed"
			if installed {
				word = "installed"
			}
			name += " " + style.MutedText.Render("("+word+")")
		}
		cells = []string{name}
	} else {
		icon := ""
		if list != nil {
			mark, role := m.icons.Absent, style.MutedColor
			if installed {
				mark, role = m.icons.Installed, style.SuccessColor
			}
			icon = lipgloss.NewStyle().Foreground(role).Render(mark)
		}
		cells = []string{icon, name}
	}
	return tableRow{cells: append(cells, pkg.version, pkg.desc)}
}

// formulaDepsView is the Deps tab of a formula: its runtime and build
// dependencies, each with its installed state.
func (m Model) formulaDepsView(f *brew.Formula, width int) string {
	view := m.renderDepGroups([]depGroup{
		{heading: "Runtime", names: f.Dependencies, check: depFormula},
		{heading: "Build", names: f.BuildDeps, check: depFormula},
	}, width)
	if view == "" {
		return style.SubtleText.Render(noDependencies)
	}
	return view
}

// renderUses is the Used By tab: the installed formulae that depend on the
// selected one. `brew uses --installed` lists only installed packages, so
// each row is checked against the installed formulae like any other.
func (m Model) renderUses(content string, width int) string {
	if content == noDependents {
		return style.SubtleText.Render(content)
	}
	return m.renderDepGroups([]depGroup{{names: lines(content), check: depFormula}}, width)
}

// renderFiles is the Files tab: a count of the files `brew list` printed, then
// a table of their paths, each with its directory in the Muted role and its
// file name in Text.
func renderFiles(content string, width int) string {
	paths := lines(content)
	if content == noFilesInstalled || len(paths) == 0 {
		return style.SubtleText.Render(noFilesInstalled)
	}
	muted := style.MutedText
	text := style.TextStyle
	rows := make([]tableRow, 0, len(paths))
	for _, p := range paths {
		dir, file := path.Split(p)
		rows = append(rows, tableRow{cells: []string{muted.Render(dir) + text.Render(file)}})
	}
	count := style.SubtleText.Render(fmt.Sprintf("%d %s", len(paths), plural(len(paths), "file", "files")))
	return count + "\n" + renderTable([]tableCol{{head: "Path"}}, rows, width)
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
