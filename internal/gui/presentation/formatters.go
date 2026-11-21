package presentation

import (
	"fmt"
	"image/color"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// FormatFormulaInfo is the Info tab of a formula: its identity, status and
// links, as detail rows.
func FormatFormulaInfo(f brew.Formula, width int, icons style.Icons) string {
	status, role := formulaStatus(f, icons)
	return Info([]Field{
		{Label: "Name", Value: f.Name},
		{Label: "Version", Value: versionDisplay(f)},
		{Label: "Tap", Value: f.Tap},
		{Label: "Status", Value: status, Role: role},
		{Label: "License", Value: f.License},
		{Label: "Homepage", Value: f.Homepage, Role: style.InfoColor},
		{Label: "Bottled", Value: YesNo(f.Bottled)},
		{Label: "Description", Value: f.Description, Prose: true, Right: true},
	}, width)
}

// FormatCaskInfo is the Info tab of a cask, as detail rows.
func FormatCaskInfo(c brew.Cask, width int, icons style.Icons) string {
	status, role := caskStatus(c, icons)
	fields := []Field{{Label: "Name", Value: c.Name}}
	// The token is what brew acts on; the display name is shown beside it
	// only when it reads differently (LB-12).
	if c.DisplayName != "" && c.DisplayName != c.Name {
		fields = append(fields, Field{Label: "Display", Value: c.DisplayName})
	}
	fields = append(fields,
		Field{Label: "Version", Value: c.Version},
		Field{Label: "Tap", Value: c.Tap},
		Field{Label: "Status", Value: status, Role: role},
		Field{Label: "Homepage", Value: c.Homepage, Role: style.InfoColor},
		Field{Label: "Description", Value: c.Description, Prose: true, Right: true},
	)
	return Info(fields, width)
}

// formulaStatus is the Status row of a formula: its icon and word, in the
// role that says how it stands. Keg-only is noted after the status.
func formulaStatus(f brew.Formula, icons style.Icons) (string, color.Color) {
	status, role := installedStatus(icons, f.Pinned, f.Outdated && f.NewVersion != "", f.Version, f.NewVersion)
	if f.KegOnly {
		status += " · keg-only"
	}
	return status, role
}

// caskStatus is the Status row of a cask: as formulaStatus, with auto-update
// noted after the status.
func caskStatus(c brew.Cask, icons style.Icons) (string, color.Color) {
	status, role := installedStatus(icons, c.Pinned, c.Outdated && c.NewVersion != "", c.Version, c.NewVersion)
	if c.AutoUpdates {
		status += " · auto-update"
	}
	return status, role
}

// installedStatus is the icon, word and role shared by formulae and casks.
// Pinned outranks outdated, as it did before the detail rows.
func installedStatus(icons style.Icons, pinned, outdated bool, version, newVersion string) (string, color.Color) {
	switch {
	case pinned:
		return WithIcon(icons.Pinned, "pinned"), style.SecondaryColor
	case outdated:
		return WithIcon(icons.Outdated, fmt.Sprintf("outdated %s → %s", version, newVersion)), style.WarningColor
	default:
		return WithIcon(icons.Installed, "installed"), style.SuccessColor
	}
}

// WithIcon prefixes text with its glyph and a space; an empty glyph leaves the
// text alone, as the icon set does when icons are off.
func WithIcon(icon, text string) string {
	if icon == "" {
		return text
	}
	return icon + " " + text
}

// YesNo is the text of a yes/no field.
func YesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func versionDisplay(f brew.Formula) string {
	v := f.Version
	for _, lv := range f.ListVersions {
		if lv != v {
			v += " (" + lv + ")"
			break
		}
	}
	return v
}
