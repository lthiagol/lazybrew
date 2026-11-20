package gui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// minSidebarName is the least room a name keeps before its right-hand text
// gives way: a version is dropped before a name is cut to a stub.
const minSidebarName = 6

// sidebarCells is one sidebar row before it is laid out: a status glyph, the
// name, right-aligned text (a version, a status word) and a trailing glyph.
// Each part carries its own role, so a highlighted row can take the row
// background on every part.
type sidebarCells struct {
	icon, name, right, tail string
	iconRole, rightRole     lipgloss.Style
	tailRole                lipgloss.Style
}

// sidebarRow draws row idx of the panel's list at width w. A row the typed
// data covers is drawn from it; any other row is its item text. The highlight
// is for the focused box only: an unfocused box shows the marker alone.
func (p *panelData) sidebarRow(idx, w int, batch map[int]bool, ic style.Icons, focused bool) string {
	selected := idx == p.selected
	marked := batch != nil && batch[idx]

	marker, markerRole := "  ", style.NormalItem
	switch {
	case selected && marked:
		marker, markerRole = "▸●", style.AccentText
	case selected:
		marker, markerRole = "▸ ", style.AccentText
	case marked:
		marker, markerRole = "● ", style.OutdatedBadge
	}

	cells, ok := p.typedRow(idx, ic)
	if !ok {
		cells = sidebarCells{name: p.items[idx]}
	}
	hl := selected && focused
	nameRole := style.NormalItem
	if hl {
		nameRole = style.SelectedItem
	}
	return layoutSidebarRow(cells, sidebarRowRoles{marker: marker, markerRole: markerRole, nameRole: nameRole}, w, hl)
}

// sidebarRowRoles are the parts of a row that do not come from its data.
type sidebarRowRoles struct {
	marker     string
	markerRole lipgloss.Style
	nameRole   lipgloss.Style
}

// layoutSidebarRow lays the row out at exactly w cells: the marker, the glyph,
// the name, then the right-hand text flush right and the trailing glyph. A long
// name is cut with an ellipsis; a right-hand text gives way before the name
// does. When hl is set every part is drawn on the selected background.
func layoutSidebarRow(c sidebarCells, r sidebarRowRoles, w int, hl bool) string {
	cell := func(s string, st lipgloss.Style) string {
		if hl {
			st = st.Background(style.SurfaceColor)
		}
		return st.Render(s)
	}

	lead := ""
	if c.icon != "" {
		lead = c.icon + " "
	}
	tail := ""
	if c.tail != "" {
		tail = " " + c.tail
	}
	right := ""
	if c.right != "" {
		right = " " + c.right
	}

	fixed := lipgloss.Width(r.marker) + lipgloss.Width(lead) + lipgloss.Width(tail)
	room := w - fixed - lipgloss.Width(right)
	if right != "" && room < minSidebarName {
		right = ""
		room = w - fixed
	}
	room = max(0, room)
	name := clipLine(c.name, room)
	name += strings.Repeat(" ", max(0, room-lipgloss.Width(name)))

	line := cell(r.marker, r.markerRole)
	if lead != "" {
		line += cell(c.icon, c.iconRole) + cell(" ", lipgloss.NewStyle())
	}
	line += cell(name, r.nameRole)
	if right != "" {
		line += cell(" ", lipgloss.NewStyle()) + cell(c.right, c.rightRole)
	}
	if tail != "" {
		line += cell(" ", lipgloss.NewStyle()) + cell(c.tail, c.tailRole)
	}
	return fitLine(line, w)
}

// typedRow is row idx of the panel's list, built from the panel's typed data.
// ok is false when the panel holds no typed row at idx.
func (p *panelData) typedRow(idx int, ic style.Icons) (sidebarCells, bool) {
	switch p.id {
	case PanelFormulae:
		if idx < len(p.formulae) {
			return formulaCells(p.formulae[idx], ic), true
		}
	case PanelCasks:
		if idx < len(p.casks) {
			return caskCells(p.casks[idx], ic), true
		}
	case PanelOutdated:
		// The formulae come first in the list, then the casks.
		if idx < len(p.formulae) {
			f := p.formulae[idx]
			return outdatedCells(f.Name, f.Version, f.NewVersion, ic), true
		}
		if i := idx - len(p.formulae); i >= 0 && i < len(p.casks) {
			c := p.casks[i]
			return outdatedCells(c.Name, c.Version, c.NewVersion, ic), true
		}
	case PanelTaps:
		if idx < len(p.taps) {
			return tapCells(p.taps[idx], ic), true
		}
	case PanelServices:
		if idx < len(p.services) {
			return serviceCells(p.services[idx], ic), true
		}
	}
	return sidebarCells{}, false
}

// versionCell is the version text of a row: the installed version, or
// `installed → latest` while an upgrade is waiting.
func versionCell(version, newVersion string, outdated bool) (string, lipgloss.Style) {
	if outdated && newVersion != "" {
		return version + " → " + newVersion, style.OutdatedBadge
	}
	return version, style.SubtleText
}

func formulaCells(f brew.Formula, ic style.Icons) sidebarCells {
	c := installedCells(f.Name, f.Version, f.NewVersion, f.Outdated, ic)
	if f.Pinned {
		c.tail, c.tailRole = ic.Pinned, style.PinnedBadge
	}
	return c
}

func caskCells(cask brew.Cask, ic style.Icons) sidebarCells {
	return installedCells(cask.Name, cask.Version, cask.NewVersion, cask.Outdated, ic)
}

// installedCells is a formula or cask on the Formulae or Casks list: an
// installed glyph, or an outdated one with the version it would change to.
func installedCells(name, version, newVersion string, outdated bool, ic style.Icons) sidebarCells {
	c := sidebarCells{name: name, icon: ic.Installed, iconRole: style.InstalledBadge}
	if outdated {
		c.icon, c.iconRole = ic.Upgradable, style.OutdatedBadge
	}
	c.right, c.rightRole = versionCell(version, newVersion, outdated)
	return c
}

// outdatedCells is a row on the Outdated list: every row is waiting for an
// upgrade, so it always shows `installed → latest`.
func outdatedCells(name, version, newVersion string, ic style.Icons) sidebarCells {
	c := sidebarCells{name: name, icon: ic.Upgradable, iconRole: style.OutdatedBadge}
	c.right, c.rightRole = versionCell(version, newVersion, true)
	return c
}

func tapCells(t brew.Tap, ic style.Icons) sidebarCells {
	c := sidebarCells{name: t.Name, icon: ic.OK, iconRole: style.InstalledBadge, right: "trusted", rightRole: style.SubtleText}
	switch {
	case t.IsOfficial:
		c.right = "official"
	case !t.Trusted:
		c.icon, c.iconRole, c.right = ic.Warn, style.OutdatedBadge, "untrusted"
	}
	return c
}

func serviceCells(s brew.Service, ic style.Icons) sidebarCells {
	c := sidebarCells{name: s.Name, right: string(s.Status), rightRole: style.SubtleText}
	switch s.Status {
	case brew.ServiceStarted:
		c.icon, c.iconRole = ic.Running, style.InstalledBadge
	case brew.ServiceStopped:
		c.icon, c.iconRole = ic.Stopped, style.SubtleText
	case brew.ServiceError:
		c.icon, c.iconRole = ic.Error, style.ErrorBadge
	default:
		c.right = string(brew.ServiceNone)
	}
	return c
}
