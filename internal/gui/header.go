package gui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/style"
	"github.com/lthiagol/lazybrew/internal/gui/textutil"
)

// headerRows is the height of the bar across the top of the screen.
const headerRows = 1

// headerGap separates the header's groups: the name, the operation, the badges.
const headerGap = "  "

// renderHeader draws the one-row bar at the top of the screen. The app name
// and the Homebrew version are on the left, the running operation after them,
// and the badges flush right. A value that is not known is left out, never
// shown as 0. The name always stays; badges drop from the left when the bar
// is short.
func (m Model) renderHeader() string {
	left := style.AccentText.Render(presentation.WithIcon(m.icons.App, "lazybrew"))
	if m.status.Version != "" {
		left += headerGap + style.SubtleText.Render("Homebrew "+m.status.Version)
	}

	// The operation outranks the badges: badges drop first, from the left, and
	// the operation is cut only once every badge is gone.
	op := m.headerOperation()
	badges := m.headerBadges()
	for len(badges) > 0 && lipgloss.Width(left)+opSpan(op)+badgesSpan(badges) > m.width {
		badges = badges[1:]
	}
	right := strings.Join(badges, headerGap)

	// The operation takes the room between the name and the badges, keeping
	// the gap before the badges, and is cut to fit.
	reserve := 0
	if right != "" {
		reserve = lipgloss.Width(headerGap)
	}
	room := m.width - lipgloss.Width(left) - lipgloss.Width(right) - reserve
	middle := ""
	if op != "" && room > lipgloss.Width(headerGap) {
		middle = headerGap + textutil.Clip(op, room-lipgloss.Width(headerGap))
	}
	spaces := max(0, m.width-lipgloss.Width(left)-lipgloss.Width(middle)-lipgloss.Width(right))
	line := left + middle + strings.Repeat(" ", spaces) + right
	return fitFrame(line, m.width, headerRows)
}

// headerOperation is what is running: the operation's title, or a refresh
// with no operation. It is empty when the program is idle.
func (m Model) headerOperation() string {
	switch {
	case m.opState != nil && m.opState.Running():
		return style.AccentText.Render(presentation.WithIcon(m.icons.Refresh, m.opState.Title))
	case m.refreshing > 0:
		return style.SubtleText.Render(presentation.WithIcon(m.icons.Refresh, "refreshing…"))
	}
	return ""
}

// headerBadges lists the badges the header shows in drop order: when the bar is
// short the leftmost go first. A count is shown only once it is known.
func (m Model) headerBadges() []string {
	var badges []string
	if n, ok := m.headerCount(PanelFormulae); ok {
		badges = append(badges, style.SubtleText.Render(fmt.Sprintf("%d formulae", n)))
	}
	if n, ok := m.headerCount(PanelCasks); ok {
		badges = append(badges, style.SubtleText.Render(fmt.Sprintf("%d casks", n)))
	}
	if n, ok := m.headerCount(PanelOutdated); ok && n > 0 {
		badges = append(badges, style.OutdatedBadge.Render(fmt.Sprintf("%d outdated", n)))
	}
	if status := m.updateStatusText(); status != "" {
		badges = append(badges, style.SubtleText.Render(status))
	}
	return badges
}

// headerCount is the count a badge shows: the panel's own once its list has
// loaded, else the status dashboard's, else nothing. An unknown count is left
// out, never shown as 0.
func (m Model) headerCount(id PanelID) (int, bool) {
	if p := m.panels[id]; p.loaded && p.err == nil {
		switch id {
		case PanelFormulae:
			return len(p.formulae), true
		case PanelCasks:
			return len(p.casks), true
		case PanelOutdated:
			return len(p.items), true
		}
	}
	switch id {
	case PanelFormulae:
		return m.status.Formulae.N, m.status.Formulae.Known
	case PanelCasks:
		return m.status.Casks.N, m.status.Casks.Known
	case PanelOutdated:
		return m.status.Outdated.N, m.status.Outdated.Known
	}
	return 0, false
}

// updateStatusText is the update badge: updating…, updated 2m ago, or never
// updated when update_on_start is true and no update has run yet. It is empty
// when there is nothing to say.
func (m Model) updateStatusText() string {
	switch {
	case m.isUpdating:
		return "updating…"
	case !m.lastUpdate.IsZero():
		return "updated " + agoText(time.Since(m.lastUpdate))
	case m.cfg.Brew.UpdateOnStart == config.UpdateAlways:
		return "never updated"
	}
	return ""
}

// agoText is a duration in the one unit that reads well at a glance.
func agoText(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	default:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
}

// joinedWidth is the width of badges laid end to end with headerGap between.
func joinedWidth(badges []string) int {
	return lipgloss.Width(strings.Join(badges, headerGap))
}

// opSpan is the width the operation takes in the header, its leading gap
// included; zero when nothing runs.
func opSpan(op string) int {
	if op == "" {
		return 0
	}
	return lipgloss.Width(headerGap) + lipgloss.Width(op)
}

// badgesSpan is the width the badges take, the gap before them included; zero
// when there are none.
func badgesSpan(badges []string) int {
	if len(badges) == 0 {
		return 0
	}
	return lipgloss.Width(headerGap) + joinedWidth(badges)
}
