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

// OutdatedInfo carries the outdated-package count and whether the count is
// known yet. Known=false means the count has never been loaded — it must not
// be rendered as a number.
type OutdatedInfo struct {
	Count int
	Known bool
}

func FormatStatusDashboard(
	formulaeCount int,
	casksCount int,
	outdated OutdatedInfo,
	tapsCount int,
	officialTaps int,
	thirdPartyTaps int,
	servicesCount int,
	servicesStarted int,
	brewVersion string,
	prefix string,
) []string {
	var lines []string
	if brewVersion != "" {
		lines = append(lines, fmt.Sprintf("Homebrew %s", brewVersion))
	}
	if prefix != "" {
		lines = append(lines, fmt.Sprintf("Prefix: %s", prefix))
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Formulae:  %d installed", formulaeCount))
	lines = append(lines, fmt.Sprintf("Casks:     %d installed", casksCount))
	if outdated.Known {
		lines = append(lines, fmt.Sprintf("Outdated:  %d packages", outdated.Count))
	} else {
		lines = append(lines, "Outdated:  — not loaded (press R to load)")
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Taps:      %d total", tapsCount))
	lines = append(lines, fmt.Sprintf("           %d official, %d third-party", officialTaps, thirdPartyTaps))
	lines = append(lines, "")
	if servicesCount > 0 {
		lines = append(lines, fmt.Sprintf("Services:  %d total", servicesCount))
		lines = append(lines, fmt.Sprintf("           %d running, %d stopped", servicesStarted, servicesCount-servicesStarted))
	} else {
		lines = append(lines, "Services:  0")
	}
	lines = append(lines, "")
	lines = append(lines, "R to refresh")
	return lines
}

// FormatFormulaInfo is the Info tab of a formula: its identity, status and
// links, as detail rows.
func FormatFormulaInfo(f brew.Formula, width int, icons style.Icons) string {
	status, role := formulaStatus(f, icons)
	return Detail([]Field{
		{Label: "Name", Value: f.Name},
		{Label: "Version", Value: versionDisplay(f)},
		{Label: "Tap", Value: f.Tap},
		{Label: "Status", Value: status, Role: role},
		{Label: "License", Value: f.License},
		{Label: "Homepage", Value: f.Homepage, Role: style.InfoColor},
		{Label: "Bottled", Value: yesNo(f.Bottled)},
		{Label: "Description", Value: f.Description, Prose: true},
	}, width)
}

// FormatCaskInfo is the Info tab of a cask, as detail rows.
func FormatCaskInfo(c brew.Cask, width int, icons style.Icons) string {
	status, role := caskStatus(c, icons)
	return Detail([]Field{
		{Label: "Name", Value: c.Name},
		{Label: "Version", Value: c.Version},
		{Label: "Tap", Value: c.Tap},
		{Label: "Status", Value: status, Role: role},
		{Label: "Homepage", Value: c.Homepage, Role: style.InfoColor},
		{Label: "Description", Value: c.Description, Prose: true},
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
		return withIcon(icons.Pinned, "pinned"), style.SecondaryColor
	case outdated:
		return withIcon(icons.Outdated, fmt.Sprintf("outdated %s → %s", version, newVersion)), style.WarningColor
	default:
		return withIcon(icons.Installed, "installed"), style.SuccessColor
	}
}

// withIcon prefixes text with its glyph and a space; an empty glyph leaves the
// text alone, as the icon set does when icons are off.
func withIcon(icon, text string) string {
	if icon == "" {
		return text
	}
	return icon + " " + text
}

func yesNo(b bool) string {
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

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

func FormatDoctorStatus(warnings []brew.DoctorWarning, err error) string {
	if err != nil {
		return fmt.Sprintf("Doctor: unavailable (%s)", err.Error())
	}
	if len(warnings) == 0 {
		return "Doctor: No issues"
	}
	if len(warnings) == 1 {
		return "Doctor: 1 warning"
	}
	return fmt.Sprintf("Doctor: %d warnings", len(warnings))
}

func padRight(s string, length int) string {
	runes := []rune(s)
	n := len(runes)
	if n >= length {
		return string(runes[:length])
	}
	return s + strings.Repeat(" ", length-n)
}
