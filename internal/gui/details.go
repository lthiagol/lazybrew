package gui

import (
	"fmt"
	"image/color"
	"maps"
	"slices"
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

// serviceInfoTab is the one tab of the Services panel: the selected service's
// Info.
const serviceInfoTab = 0

// trustUnknownText is what the Taps panel says when brew did not give the trust
// state of its taps: the rows, the border label and the Tap Info all use it.
const trustUnknownText = "trust unknown — press R"

// tapSummary is the one-line header of the Taps panel's Tap Info tab. With the
// trust state unread it counts the official taps and no trusted or untrusted.
func tapSummary(taps []brew.Tap, trustKnown bool) string {
	official, trusted := 0, 0
	for _, t := range taps {
		if t.IsOfficial {
			official++
		}
		if t.Trusted || t.IsOfficial {
			trusted++
		}
	}
	if !trustKnown {
		return fmt.Sprintf("%d taps · %d official · %s", len(taps), official, trustUnknownText)
	}
	return fmt.Sprintf("%d taps · %d official · %d trusted · %d untrusted",
		len(taps), official, trusted, len(taps)-trusted)
}

// tapDetails is the block describing one tap.
func tapDetails(t *brew.Tap, trustKnown bool, width int) string {
	trusted := "untrusted"
	if t.Trusted || t.IsOfficial {
		trusted = "trusted"
	}
	trustField := presentation.Field{Label: "Trusted", Value: trusted}
	if !trustKnown && !t.IsOfficial {
		trustField = presentation.Field{Label: "Trusted", Value: trustUnknownText, Role: style.WarningColor}
	}
	return presentation.Info([]presentation.Field{
		{Label: "Tap", Value: t.Name},
		trustField,
		{Label: "Official", Value: boolStr(t.IsOfficial)},
		{Label: "Remote", Value: t.Remote, Role: style.InfoColor, Right: true},
		{Label: "Formulae", Value: strconv.Itoa(t.FormulaCount), Right: true},
		{Label: "Casks", Value: strconv.Itoa(t.CaskCount), Right: true},
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
func tapTrustStory(t *brew.Tap, trustKnown bool, width int) string {
	var state, meaning string
	switch {
	case t.IsOfficial:
		state = "official"
		meaning = "Official taps (homebrew/*) are always trusted: their formulae and casks load without a trust step."
	case !trustKnown:
		state = "unknown"
		meaning = "Homebrew did not report this tap's trust state. Press R to read it again."
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
	if !trustKnown {
		story = strings.Join([]string{meaning, fmt.Sprintf("Its %d %s and %d %s follow that state.",
			t.FormulaCount, plural(t.FormulaCount, "formula", "formulae"),
			t.CaskCount, plural(t.CaskCount, "cask", "casks")), "", "Press R to read the trust state again."}, "\n")
	}
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
	summary := style.SubtleText
	if p.trustUnknown {
		summary = style.OutdatedBadge
	}
	header := summary.Render(tapSummary(p.taps, !p.trustUnknown))
	t := p.selectedTap()
	if t == nil {
		return header
	}
	keys := tapActionHints(*t, !p.trustUnknown)
	return header + "\n\n" + tapDetails(t, !p.trustUnknown, width) + "\n\n" + footerLine(keys, width)
}

// renderServiceInfo is the Info tab of Services: the selected service's detail
// from `brew services info`, then the actions that apply to its state. Until
// that read has loaded, it shows the list's fields, never a blank pane.
func renderServiceInfo(p *panelData, width int, icons style.Icons) string {
	s := p.selectedService()
	if s == nil {
		return style.SubtleText.Render(emptyMessage(PanelServices))
	}
	fields := serviceListDetail(*s, width, icons)
	if info := p.serviceInfoFor(s.Name); info != nil {
		fields = serviceInfoDetail(*info, width, icons)
	}
	keys := append(serviceActionHints(*s), keyHint{"c", "cleanup"})
	return fields + "\n\n" + footerLine(keys, width)
}

// serviceListDetail is the Info tab before `brew services info` has loaded: the
// fields the list already holds.
func serviceListDetail(s brew.Service, width int, icons style.Icons) string {
	status, role := serviceStatus(s.Status, icons)
	return presentation.Detail([]presentation.Field{
		{Label: "Name", Value: s.Name},
		{Label: "Status", Value: status, Role: role},
		{Label: "User", Value: s.User},
		{Label: "File", Value: s.File},
		{Label: "Exit code", Value: strconv.Itoa(s.ExitCode)},
	}, width)
}

// serviceInfoDetail is the Info tab from `brew services info`. A field brew
// reports as null is omitted.
func serviceInfoDetail(s brew.ServiceInfo, width int, icons style.Icons) string {
	status, role := serviceStatus(s.Status, icons)
	fields := []presentation.Field{
		{Label: "Name", Value: s.Name},
		{Label: "Status", Value: status, Role: role},
		{Label: "PID", Value: optionalInt(s.PID)},
		{Label: "User", Value: s.User},
		{Label: "Loaded", Value: boolStr(s.Loaded)},
		{Label: "Registered", Value: boolStr(s.Registered)},
		{Label: "Unit file", Value: s.File, Right: true},
		{Label: "Command", Value: s.Command, Right: true},
		{Label: "Working dir", Value: s.WorkingDir, Right: true},
		{Label: "Log", Value: s.LogPath, Right: true},
	}
	if s.ErrorLogPath != s.LogPath {
		fields = append(fields, presentation.Field{Label: "Error log", Value: s.ErrorLogPath, Right: true})
	}
	fields = append(fields,
		presentation.Field{Label: "Exit code", Value: optionalInt(s.ExitCode)},
		presentation.Field{Label: "Interval", Value: intervalString(s.Interval)},
		presentation.Field{Label: "Cron", Value: cronString(s.Cron)},
	)
	return presentation.Info(fields, width)
}

// serviceRunning is whether a service runs now: brew's started status, which
// is a running pid. Start is refused on it.
func serviceRunning(s brew.Service) bool {
	return s.Status == brew.ServiceStarted
}

// serviceStoppable is whether stop and restart apply: a running service, or a
// loaded one that is not running. brew derives the status in
// Library/Homebrew/services/formula_wrapper.rb (to_hash): error and scheduled
// are loaded units, and brew's stop acts on any loaded unit. brew also reports
// stopped as loaded; AC-03.1 of the brief keeps stop and restart refused there.
func serviceStoppable(s brew.Service) bool {
	switch s.Status {
	case brew.ServiceStarted, brew.ServiceError, brew.ServiceScheduled:
		return true
	}
	return false
}

// serviceStatus is the Status row of a service: its icon and word, in the role
// of its state. A state with no word of its own reads as brew names it.
func serviceStatus(status brew.ServiceStatus, icons style.Icons) (string, color.Color) {
	switch status {
	case brew.ServiceStarted:
		return withIcon(icons.Running, "running"), style.SuccessColor
	case brew.ServiceStopped:
		return withIcon(icons.Stopped, "stopped"), style.MutedColor
	case brew.ServiceError:
		return withIcon(icons.Error, "error"), style.ErrorColor
	case brew.ServiceScheduled:
		return withIcon(icons.Scheduled, "scheduled"), style.InfoColor
	case "":
		return "none", style.MutedColor
	default:
		return string(status), style.MutedColor
	}
}

// optionalInt is the number, or empty when brew reports null.
func optionalInt(n *int) string {
	if n == nil {
		return ""
	}
	return strconv.Itoa(*n)
}

// intervalString is a service's run interval in seconds, or empty.
func intervalString(n *int) string {
	if n == nil {
		return ""
	}
	return fmt.Sprintf("every %ds", *n)
}

// cronString is a cron schedule's fields in name order, such as
// "hour 3, minute 0", or empty.
func cronString(c map[string]any) string {
	keys := slices.Sorted(maps.Keys(c))
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %v", k, c[k]))
	}
	return strings.Join(parts, ", ")
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
