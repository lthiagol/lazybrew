package presentation

import (
	"fmt"
	"strings"

	"github.com/thiago/lazybrew/internal/brew"
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

func FormatStatusDashboard(
	formulaeCount int,
	casksCount int,
	outdatedCount int,
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
	lines = append(lines, fmt.Sprintf("Formulae: %d installed", formulaeCount))
	lines = append(lines, fmt.Sprintf("Casks: %d installed", casksCount))
	lines = append(lines, fmt.Sprintf("Outdated: %d packages", outdatedCount))
	lines = append(lines, fmt.Sprintf("Taps: %d (%d official, %d third-party)", tapsCount, officialTaps, thirdPartyTaps))
	lines = append(lines, fmt.Sprintf("Services: %d (%d running, %d stopped)", servicesCount, servicesStarted, servicesCount-servicesStarted))
	lines = append(lines, "")
	lines = append(lines, "R to refresh")
	return lines
}

func FormatDoctorStatus(warnings []brew.DoctorWarning) string {
	if len(warnings) == 0 {
		return "Doctor: No issues"
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
