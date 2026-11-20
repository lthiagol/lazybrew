package gui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
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
	left := style.AccentText.Render(withIcon(m.icons.App, "lazybrew"))
	if m.brewVersion != "" {
		left += headerGap + style.SubtleText.Render("Homebrew "+m.brewVersion)
	}

	badges := m.headerBadges()
	for len(badges) > 0 && lipgloss.Width(left)+joinedWidth(badges) > m.width {
		badges = badges[1:]
	}
	right := strings.Join(badges, headerGap)

	// The operation takes the room between the name and the badges, keeping
	// a gap on each side, and is cut to fit.
	room := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	middle := ""
	if op := m.headerOperation(); op != "" && room-2*len(headerGap) >= 1 {
		middle = headerGap + clipLine(op, room-2*len(headerGap))
	}
	line := left + middle + strings.Repeat(" ", max(0, room-lipgloss.Width(middle))) + right
	return fitFrame(line, m.width, headerRows)
}

// headerOperation is what is running: the operation's title, or a refresh
// with no operation. It is empty when the program is idle.
func (m Model) headerOperation() string {
	switch {
	case m.opState != nil && m.opState.Running():
		return style.AccentText.Render(withIcon(m.icons.Refresh, m.opState.Title))
	case m.refreshing > 0:
		return style.SubtleText.Render(withIcon(m.icons.Refresh, "refreshing…"))
	}
	return ""
}

// headerBadges lists the badges the header shows, most important first, so
// the ones dropped when the bar is short are the least useful. A count is
// shown only once it is known.
func (m Model) headerBadges() []string {
	var badges []string
	if p := m.panels[PanelFormulae]; p.loaded && p.err == nil {
		badges = append(badges, style.SubtleText.Render(fmt.Sprintf("%d formulae", len(p.formulae))))
	}
	if p := m.panels[PanelCasks]; p.loaded && p.err == nil {
		badges = append(badges, style.SubtleText.Render(fmt.Sprintf("%d casks", len(p.casks))))
	}
	if p := m.panels[PanelOutdated]; p.loaded && p.err == nil && len(p.items) > 0 {
		badges = append(badges, style.OutdatedBadge.Render(fmt.Sprintf("%d outdated", len(p.items))))
	}
	if status := m.updateStatusText(); status != "" {
		badges = append(badges, style.SubtleText.Render(status))
	}
	return badges
}

// updateStatusText is the update badge: updating…, updated 2m ago, or never
// updated when update_on_start is set and no update has run yet. It is empty
// when there is nothing to say.
func (m Model) updateStatusText() string {
	switch {
	case m.isUpdating:
		return "updating…"
	case !m.lastUpdate.IsZero():
		return "updated " + agoText(time.Since(m.lastUpdate))
	case m.cfg.Brew.UpdateOnStart:
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
