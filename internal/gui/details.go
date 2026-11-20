package gui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// The tab indices of the panels whose sidebar list is the list (v2-03b). Their
// main tabs show the selected item's details, never the list again.
const (
	statusDashboardTab = iota
	statusConfigTab
	statusDoctorTab
)

const (
	outdatedInfoTab = iota
	outdatedVersionsTab
)

const (
	tapInfoTab = iota
	tapTrustTab
	tapFormulaeTab
)

// serviceStatusTab is the one tab of the Services panel.
const serviceStatusTab = 0

// detailLine is a `Key:   value` row of a details block.
func detailLine(key, value string) string {
	return fmt.Sprintf("%-12s %s", key+":", value)
}

// dashboardText is the Dashboard tab of Status: the whole dashboard, as text.
// It is not a list, so no row is selected in it.
func dashboardText(p *panelData) string {
	if !p.loaded {
		return style.SubtleText.Render("Status dashboard not loaded yet — press R to refresh")
	}
	if len(p.dashboard) == 0 {
		return style.SubtleText.Render(emptyMessage(PanelStatus))
	}
	return style.NormalItem.Render(strings.Join(p.dashboard, "\n"))
}

// tapSummary is the one-line header of the Taps panel's Tap Info tab.
func tapSummary(taps []brew.Tap) string {
	official, trusted := 0, 0
	for _, t := range taps {
		if t.IsOfficial {
			official++
		}
		if t.Trusted || t.IsOfficial {
			trusted++
		}
	}
	return fmt.Sprintf("%d taps · %d official · %d trusted · %d untrusted",
		len(taps), official, trusted, len(taps)-trusted)
}

// tapDetails is the block describing one tap.
func tapDetails(t *brew.Tap) string {
	trusted := "untrusted"
	if t.Trusted || t.IsOfficial {
		trusted = "trusted"
	}
	return strings.Join([]string{
		detailLine("Tap", t.Name),
		detailLine("Trusted", trusted),
		detailLine("Official", boolStr(t.IsOfficial)),
		detailLine("Remote", t.Remote),
		detailLine("Formulae", strconv.Itoa(t.FormulaCount)),
		detailLine("Casks", strconv.Itoa(t.CaskCount)),
	}, "\n") + "\n"
}

// plural is the singular noun for one, else the plural.
func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// tapTrustStory is the Trust tab: what the selected tap's trust state means for
// its formulae and casks, and the key that changes it. Homebrew loads the
// formulae and casks of a non-official tap only once the tap is trusted.
func tapTrustStory(t *brew.Tap) string {
	var state, meaning string
	switch {
	case t.IsOfficial:
		state = "official"
		meaning = "Official taps (homebrew/*) are always trusted: their formulae and casks load without a trust step."
	case t.Trusted:
		state = "trusted"
		meaning = "Homebrew may load this tap's formulae and casks."
	default:
		state = "untrusted"
		meaning = "Homebrew may not load this tap's formulae and casks until the tap is trusted."
	}
	return strings.Join([]string{
		detailLine("Tap", t.Name),
		detailLine("Trust", state),
		"",
		meaning,
		fmt.Sprintf("Its %d %s and %d %s follow that state.",
			t.FormulaCount, plural(t.FormulaCount, "formula", "formulae"),
			t.CaskCount, plural(t.CaskCount, "cask", "casks")),
		"",
		"Press t to trust or untrust the tap, or one of its formulae or casks.",
	}, "\n") + "\n"
}

// renderTapInfo is the Tap Info tab: the summary of every tap, then the
// selected tap's details.
func renderTapInfo(p *panelData) string {
	if len(p.taps) == 0 {
		return style.SubtleText.Render(emptyMessage(PanelTaps))
	}
	header := style.SubtleText.Render(tapSummary(p.taps))
	t := p.selectedTap()
	if t == nil {
		return header
	}
	return header + "\n\n" + style.NormalItem.Render(tapDetails(t))
}

// renderServiceStatus is the Status tab of Services: the selected service's
// details, from the data the panel already holds.
func renderServiceStatus(p *panelData) string {
	s := p.selectedService()
	if s == nil {
		return style.SubtleText.Render(emptyMessage(PanelServices))
	}
	orDash := func(v string) string {
		if v == "" {
			return "—"
		}
		return v
	}
	return style.NormalItem.Render(strings.Join([]string{
		detailLine("Name", s.Name),
		detailLine("Status", string(s.Status)),
		detailLine("User", orDash(s.User)),
		detailLine("File", orDash(s.File)),
		detailLine("Exit code", strconv.Itoa(s.ExitCode)),
	}, "\n"))
}

// batchSummary says how many outdated packages are selected and what the
// batch keys do with them.
func batchSummary(selected int) string {
	if selected == 0 {
		return style.SubtleText.Render("none selected — Space to select, a for all")
	}
	return style.AccentText.Render(fmt.Sprintf("%d selected — u to upgrade, U all", selected))
}

// renderOutdatedVersions is the Versions tab of Outdated: the selected
// package's installed and latest versions, its pinned state, and the batch.
func renderOutdatedVersions(p *panelData, batch map[int]bool) string {
	var lines []string
	switch {
	case p.selectedOutdatedFormula() != nil:
		f := p.selectedOutdatedFormula()
		lines = []string{
			detailLine("Name", f.Name),
			detailLine("Installed", f.Version),
			detailLine("Latest", f.NewVersion),
			detailLine("Pinned", boolStr(f.Pinned)),
		}
	case p.selectedOutdatedCask() != nil:
		c := p.selectedOutdatedCask()
		lines = []string{
			detailLine("Name", c.Name),
			detailLine("Installed", c.Version),
			detailLine("Latest", c.NewVersion),
		}
	default:
		return style.SubtleText.Render(emptyMessage(PanelOutdated))
	}
	selected := 0
	for _, on := range batch {
		if on {
			selected++
		}
	}
	return style.NormalItem.Render(strings.Join(lines, "\n")) + "\n\n" + batchSummary(selected)
}
