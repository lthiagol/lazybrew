package gui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
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
func tapDetails(t *brew.Tap, width int) string {
	trusted := "untrusted"
	if t.Trusted || t.IsOfficial {
		trusted = "trusted"
	}
	return presentation.Detail([]presentation.Field{
		{Label: "Tap", Value: t.Name},
		{Label: "Trusted", Value: trusted},
		{Label: "Official", Value: boolStr(t.IsOfficial)},
		{Label: "Remote", Value: t.Remote, Role: style.InfoColor},
		{Label: "Formulae", Value: strconv.Itoa(t.FormulaCount)},
		{Label: "Casks", Value: strconv.Itoa(t.CaskCount)},
	}, width)
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
func tapTrustStory(t *brew.Tap, width int) string {
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
	story := strings.Join([]string{
		meaning,
		fmt.Sprintf("Its %d %s and %d %s follow that state.",
			t.FormulaCount, plural(t.FormulaCount, "formula", "formulae"),
			t.CaskCount, plural(t.CaskCount, "cask", "casks")),
		"",
		"Press t to trust or untrust the tap, or one of its formulae or casks.",
	}, "\n")
	return presentation.Detail([]presentation.Field{
		{Label: "Tap", Value: t.Name},
		{Label: "Trust", Value: state},
	}, width) + "\n\n" + style.NormalItem.Render(story) + "\n"
}

// renderTapInfo is the Tap Info tab: the summary of every tap, then the
// selected tap's details.
func renderTapInfo(p *panelData, width int) string {
	if len(p.taps) == 0 {
		return style.SubtleText.Render(emptyMessage(PanelTaps))
	}
	header := style.SubtleText.Render(tapSummary(p.taps))
	t := p.selectedTap()
	if t == nil {
		return header
	}
	return header + "\n\n" + tapDetails(t, width)
}

// renderServiceStatus is the Status tab of Services: the selected service's
// details, from the data the panel already holds.
func renderServiceStatus(p *panelData, width int, icons style.Icons) string {
	s := p.selectedService()
	if s == nil {
		return style.SubtleText.Render(emptyMessage(PanelServices))
	}
	status, role := serviceStatus(s.Status, icons)
	return presentation.Detail([]presentation.Field{
		{Label: "Name", Value: s.Name},
		{Label: "Status", Value: status, Role: role},
		{Label: "User", Value: s.User},
		{Label: "File", Value: s.File},
		{Label: "Exit code", Value: strconv.Itoa(s.ExitCode)},
	}, width)
}

// serviceStatus is the Status row of a service: its icon and word, in the role
// of its state. A service that has never run reads as none.
func serviceStatus(status brew.ServiceStatus, icons style.Icons) (string, color.Color) {
	switch status {
	case brew.ServiceStarted:
		return withIcon(icons.Running, "started"), style.SuccessColor
	case brew.ServiceStopped:
		return withIcon(icons.Stopped, "stopped"), style.MutedColor
	case brew.ServiceError:
		return withIcon(icons.Error, "error"), style.ErrorColor
	default:
		return "none", style.MutedColor
	}
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
func renderOutdatedVersions(p *panelData, batch map[int]bool, width int) string {
	var details string
	switch {
	case p.selectedOutdatedFormula() != nil:
		f := p.selectedOutdatedFormula()
		details = presentation.Detail([]presentation.Field{
			{Label: "Name", Value: f.Name},
			{Label: "Installed", Value: f.Version},
			{Label: "Latest", Value: f.NewVersion},
			{Label: "Pinned", Value: boolStr(f.Pinned)},
		}, width)
	case p.selectedOutdatedCask() != nil:
		c := p.selectedOutdatedCask()
		details = presentation.Detail([]presentation.Field{
			{Label: "Name", Value: c.Name},
			{Label: "Installed", Value: c.Version},
			{Label: "Latest", Value: c.NewVersion},
		}, width)
	default:
		return style.SubtleText.Render(emptyMessage(PanelOutdated))
	}
	selected := 0
	for _, on := range batch {
		if on {
			selected++
		}
	}
	return details + "\n\n" + batchSummary(selected)
}

// cutLines cuts every line of s to width cells with …. It is for text built
// without a width, cut where it is drawn.
func cutLines(s string, width int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "…")
	}
	return strings.Join(lines, "\n")
}
