package presentation

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

func FormatFormula(f brew.Formula, width int) string {
	var parts []string
	parts = append(parts, padRight(f.Name, 22))
	parts = append(parts, padRight(f.Version, 12))
	if f.Bottled {
		parts = append(parts, "bottled    ")
	} else {
		parts = append(parts, "           ")
	}
	if f.Pinned {
		parts = append(parts, "pinned")
	} else if f.Outdated && f.NewVersion != "" {
		parts = append(parts, fmt.Sprintf("outdated %s", f.NewVersion))
	} else if f.KegOnly {
		parts = append(parts, "keg-only")
	}
	return strings.Join(parts, " ")
}

func FormatCask(c brew.Cask, width int) string {
	var parts []string
	parts = append(parts, padRight(c.Name, 22))
	parts = append(parts, padRight(c.Version, 12))
	if c.Outdated && c.NewVersion != "" {
		parts = append(parts, fmt.Sprintf("outdated %s", c.NewVersion))
	} else if c.AutoUpdates {
		parts = append(parts, "auto-update")
	}
	return strings.Join(parts, " ")
}

func FormatTap(t brew.Tap, width int) string {
	var parts []string
	parts = append(parts, padRight(t.Name, 22))
	if t.IsOfficial {
		parts = append(parts, "official   ")
	}
	if t.Trusted || t.IsOfficial {
		parts = append(parts, "trusted    ")
	} else {
		parts = append(parts, "untrusted  ")
	}
	if t.IsAPI {
		parts = append(parts, "API")
	} else {
		parts = append(parts, "clone")
	}
	return strings.Join(parts, " ")
}

func FormatService(s brew.Service, width int) string {
	var parts []string
	parts = append(parts, padRight(s.Name, 22))
	switch s.Status {
	case brew.ServiceStarted:
		parts = append(parts, "started    ")
	case brew.ServiceStopped:
		parts = append(parts, "stopped    ")
	case brew.ServiceError:
		parts = append(parts, "error      ")
	default:
		parts = append(parts, "none       ")
	}
	if s.User != "" {
		parts = append(parts, s.User)
	}
	if s.ExitCode != 0 {
		parts = append(parts, fmt.Sprintf("exit: %d", s.ExitCode))
	}
	return strings.Join(parts, " ")
}

func FormatOutdatedFormula(f brew.Formula) string {
	return fmt.Sprintf("%s  %s -> %s", padRight(f.Name, 22), padRight(f.Version, 12), f.NewVersion)
}

func FormatOutdatedCask(c brew.Cask) string {
	return fmt.Sprintf("%s  %s -> %s  cask", padRight(c.Name, 22), padRight(c.Version, 12), c.NewVersion)
}

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
	return Info([]Field{
		{Label: "Name", Value: c.Name},
		{Label: "Version", Value: c.Version},
		{Label: "Tap", Value: c.Tap},
		{Label: "Status", Value: status, Role: role},
		{Label: "Homepage", Value: c.Homepage, Role: style.InfoColor},
		{Label: "Description", Value: c.Description, Prose: true, Right: true},
	}, width)
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

func padRight(s string, length int) string {
	runes := []rune(s)
	n := len(runes)
	if n >= length {
		return string(runes[:length])
	}
	return s + strings.Repeat(" ", length-n)
}
