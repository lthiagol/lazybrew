package gui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// globalHints are the keys every panel answers to, listed after the panel's
// own in the footer.
var globalHints = []keyHint{
	{"Tab", "next"}, {"S-Tab", "prev"}, {"1-7", "jump"}, {"/", "search"},
	{"?", "help"}, {"R", "refresh"}, {"C", "log"}, {"q", "quit"}, {"PgUp/Dn", "scroll"},
}

// footerHints is the key row the footer shows: the keys of the open modal,
// else the help screen's keys while help is open, else the active panel's
// keys followed by the global ones.
func (m Model) footerHints() []keyHint {
	if m.activeModal != nil {
		return modalHints(m.activeModal)
	}
	if m.showHelp {
		return []keyHint{{"?/esc", "close"}}
	}
	hints := append([]keyHint{}, panelHints(m.activePanel)...)
	return append(hints, globalHints...)
}

// modalHints lists the keys a modal answers to, as it handles them.
func modalHints(mo modal.Modal) []keyHint {
	switch mo.(type) {
	case *modal.ConfirmModal:
		return []keyHint{{"y", "yes"}, {"n", "no"}, {"esc", "cancel"}}
	case *modal.MenuModal:
		return []keyHint{{"↑↓", "move"}, {"enter", "choose"}, {"esc", "cancel"}}
	case *modal.InputModal:
		return []keyHint{{"enter", "submit"}, {"esc", "cancel"}}
	default:
		return []keyHint{{"esc", "close"}}
	}
}

// renderBottomBar draws the footer: one row of key hints across the bottom.
func (m Model) renderBottomBar() string {
	return footerLine(m.footerHints(), m.width)
}

// footerLine lays the hints out on one row, w cells wide. Each hint is its key
// in the accent and its description in the subtle text, and hints are
// separated by two spaces. When the row cannot hold them all, it keeps the
// hints that fit and ends with "? more", which is always shown: the ? help
// screen lists the rest.
func footerLine(hints []keyHint, w int) string {
	const gap = "  "
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, style.AccentText.Render(h.key)+" "+style.SubtleText.Render(h.desc))
	}

	line := " " + strings.Join(parts, gap)
	if lipgloss.Width(line) > w {
		more := style.AccentText.Render("?") + " " + style.SubtleText.Render("more")
		line = " " + more
		for i := range parts {
			candidate := " " + strings.Join(parts[:i+1], gap) + gap + more
			if lipgloss.Width(candidate) > w {
				break
			}
			line = candidate
		}
	}
	return fitFrame(line, w, bottomBarRows)
}
