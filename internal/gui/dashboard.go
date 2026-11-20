package gui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// dashboardCardWidth is the outer width of one dashboard card. Two cards sit
// side by side when the main area is at least twice this wide; otherwise the
// cards stack.
const dashboardCardWidth = 36

// renderDashboard draws the Status dashboard as five titled cards: Homebrew,
// Packages, Taps, Services and Health. Every count comes from the typed status,
// so a count that never loaded reads as not loaded, never as 0.
func (m Model) renderDashboard(width int) string {
	inner := dashboardCardWidth - 2
	s := m.status
	// Health comes second, so that when the cards stack a problem is in view
	// without scrolling.
	cards := []string{
		dashCard("Homebrew", presentation.Detail([]presentation.Field{
			{Label: "Version", Value: valueOr(s.Version, "unknown")},
			{Label: "Prefix", Value: valueOr(s.Prefix, "unknown")},
		}, inner)),
		dashCard("Health", m.healthBody()),
		dashCard("Packages", m.packagesBody(inner)),
		dashCard("Taps", tapsBody(s, inner)),
		dashCard("Services", m.servicesBody(inner)),
	}
	if width < 2*dashboardCardWidth {
		return lipgloss.JoinVertical(lipgloss.Left, cards...)
	}
	var rows []string
	for i := 0; i < len(cards); i += 2 {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cards[i:min(i+2, len(cards))]...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// dashCard is one titled card around its body, sized to the body's rows.
func dashCard(title, body string) string {
	return panel(titled(title), borderLabel{}, body, dashboardCardWidth, lipgloss.Height(body)+2, false)
}

// packagesBody is the Packages card: the installed counts, and the outdated
// count with its gauge once that count is known.
func (m Model) packagesBody(inner int) string {
	s := m.status
	body := presentation.Detail([]presentation.Field{
		countField("Formulae", s.Formulae, "installed"),
		countField("Casks", s.Casks, "installed"),
		outdatedField(s.Outdated, m.outdatedPending()),
	}, inner)
	if !s.Outdated.Known {
		return body
	}
	installed := 0
	if s.Formulae.Known && s.Casks.Known {
		installed = s.Formulae.N + s.Casks.N
	}
	return body + "\n" + outdatedGauge(s.Outdated.N, installed, inner)
}

// tapsBody is the Taps card: the total, and how many are official and not.
func tapsBody(s statusSummary, inner int) string {
	if !s.Taps.Known {
		return presentation.Detail([]presentation.Field{countField("Taps", s.Taps, "")}, inner)
	}
	return presentation.Detail([]presentation.Field{
		{Label: "Total", Value: fmt.Sprint(s.Taps.N)},
		{Label: "Official", Value: fmt.Sprint(s.OfficialTaps)},
		{Label: "Third-party", Value: fmt.Sprint(s.Taps.N - s.OfficialTaps)},
	}, inner)
}

// servicesBody is the Services card: the running and stopped services, each
// with its icon.
func (m Model) servicesBody(inner int) string {
	s := m.status
	if !s.Services.Known {
		return presentation.Detail([]presentation.Field{countField("Services", s.Services, "")}, inner)
	}
	stopped := s.Services.N - s.ServicesStarted
	return presentation.Detail([]presentation.Field{
		{Label: "Running", Value: withIcon(m.icons.Running, fmt.Sprint(s.ServicesStarted)), Role: style.SuccessColor},
		{Label: "Stopped", Value: withIcon(m.icons.Stopped, fmt.Sprint(stopped)), Role: style.MutedColor},
	}, inner)
}

// healthBody is the Health card: the doctor summary when a run has answered,
// the reason when its run failed, the prompt to run it when none has, and the
// sources that failed. A line the card is too narrow for wraps, so a timeout
// or a failure is read whole.
func (m Model) healthBody() string {
	s := m.status
	var rows []string
	switch {
	case s.DoctorErr != "":
		rows = append(rows, warningLines(withIcon(m.icons.Warn, "doctor unavailable: "+s.DoctorErr))...)
	case !s.Doctor.Known:
		rows = append(rows, style.SubtleText.Render("press d to run doctor"))
	case s.Doctor.N == 0:
		rows = append(rows, lipgloss.NewStyle().Foreground(style.SuccessColor).
			Render(withIcon(m.icons.OK, "ready to brew")))
	default:
		rows = append(rows, lipgloss.NewStyle().Foreground(style.WarningColor).
			Render(withIcon(m.icons.Warn, fmt.Sprintf("%d %s", s.Doctor.N, plural(s.Doctor.N, "warning", "warnings")))))
	}
	for _, p := range s.Problems {
		rows = append(rows, warningLines(p)...)
	}
	return strings.Join(rows, "\n")
}

// warningLines is text in the Warning role, wrapped to the card's inner width
// and one row per wrapped line.
func warningLines(text string) []string {
	warning := lipgloss.NewStyle().Foreground(style.WarningColor)
	var out []string
	for _, line := range strings.Split(ansi.Wrap(text, dashboardCardWidth-2, ""), "\n") {
		out = append(out, warning.Render(line))
	}
	return out
}

// countField is a count row: the number and its noun, or not loaded.
func countField(label string, c knownCount, noun string) presentation.Field {
	if !c.Known {
		return presentation.Field{Label: label, Value: "not loaded", Role: style.SubtleColor}
	}
	if noun == "" {
		return presentation.Field{Label: label, Value: fmt.Sprint(c.N)}
	}
	return presentation.Field{Label: label, Value: fmt.Sprintf("%d %s", c.N, noun)}
}

// outdatedPending reports whether the Outdated count is on its way: the preload
// has that read queued or in flight. Without the preload, or once its read has
// failed, the count is not loaded and R is how to load it.
func (m Model) outdatedPending() bool {
	p := m.panels[PanelOutdated]
	return m.cfg.GUI.Preload && !p.loaded && p.err == nil
}

// outdatedField is the Outdated row. Unknown is said so, never shown as 0: as
// loading while its preload read is on its way, else as not loaded.
func outdatedField(c knownCount, pending bool) presentation.Field {
	switch {
	case !c.Known && pending:
		return presentation.Field{Label: "Outdated", Value: "loading…", Role: style.SubtleColor}
	case !c.Known:
		return presentation.Field{Label: "Outdated", Value: "not loaded — press R", Role: style.SubtleColor}
	case c.N > 0:
		return presentation.Field{Label: "Outdated", Value: fmt.Sprintf("%d %s", c.N, plural(c.N, "package", "packages")), Role: style.WarningColor}
	default:
		return presentation.Field{Label: "Outdated", Value: "0 packages", Role: style.SuccessColor}
	}
}

// outdatedGauge is the bar under Outdated: the current share of the installed
// packages as a solid run, then a dashed tail for the outdated ones. With
// nothing outdated it is all solid, in the Success role; otherwise the solid
// run is in the Warning role.
func outdatedGauge(outdated, installed, width int) string {
	if outdated == 0 {
		return lipgloss.NewStyle().Foreground(style.SuccessColor).Render(strings.Repeat("━", width))
	}
	current := 0
	if installed > 0 {
		current = min(width, max(0, width*(installed-outdated)/installed))
	}
	solid := lipgloss.NewStyle().Foreground(style.WarningColor).Render(strings.Repeat("━", current))
	dashed := lipgloss.NewStyle().Foreground(style.MutedColor).Render(strings.Repeat("╌", width-current))
	return solid + dashed
}

// valueOr is s, or the fallback when s is empty.
func valueOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// doctorView is the Doctor tab: each warning as its own block, a heading line
// with the warning icon and the details in Text, blocks separated by a blank
// line. A clean run reads as ready to brew.
func doctorView(warnings []brew.DoctorWarning, icons style.Icons) string {
	if len(warnings) == 0 {
		return lipgloss.NewStyle().Foreground(style.SuccessColor).
			Render(withIcon(icons.OK, "Your system is ready to brew."))
	}
	blocks := make([]string, 0, len(warnings))
	for _, w := range warnings {
		heading := lipgloss.NewStyle().Foreground(style.WarningColor).Render(withIcon(icons.Warn, w.Title))
		blocks = append(blocks, heading+"\n"+style.NormalItem.Render(strings.TrimRight(w.Details, "\n")))
	}
	return strings.Join(blocks, "\n\n")
}
