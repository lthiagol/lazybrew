package gui

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// sortFormulae orders outdated first, then pinned, then name A–Z (case-insensitive).
// Two separate ranks — outdated and pinned are not collapsed into one flag.
func sortFormulae(fs []brew.Formula) {
	sort.SliceStable(fs, func(i, j int) bool {
		oi, oj := 1, 1
		if fs[i].Outdated {
			oi = 0
		}
		if fs[j].Outdated {
			oj = 0
		}
		if oi != oj {
			return oi < oj
		}
		pi, pj := 1, 1
		if fs[i].Pinned {
			pi = 0
		}
		if fs[j].Pinned {
			pj = 0
		}
		if pi != pj {
			return pi < pj
		}
		return strings.ToLower(fs[i].Name) < strings.ToLower(fs[j].Name)
	})
}

func formulaVersionCell(f brew.Formula) string {
	if f.Outdated && f.NewVersion != "" {
		return f.Version + " -> " + f.NewVersion
	}
	return f.Version
}

func formulaStatusLabel(f brew.Formula) string {
	if f.Outdated {
		return "outdated"
	}
	if f.Pinned {
		return "pinned"
	}
	return "installed"
}

func formulaStatusStyled(f brew.Formula) string {
	label := formulaStatusLabel(f)
	switch label {
	case "outdated":
		return style.OutdatedBadge.Render(label)
	case "pinned":
		return style.PinnedBadge.Render(label)
	default:
		return style.InstalledBadge.Render(label)
	}
}

// shortenTap abbreviates taps for the List table (homebrew/core → hb/c).
func shortenTap(tap string) string {
	if tap == "" {
		return ""
	}
	parts := strings.Split(tap, "/")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		switch strings.ToLower(p) {
		case "homebrew":
			out = append(out, "hb")
		case "core":
			out = append(out, "c")
		default:
			r := []rune(p)
			if len(r) <= 4 {
				out = append(out, p)
			} else {
				out = append(out, string(r[:2]))
			}
		}
	}
	return strings.Join(out, "/")
}

func cellPad(s string, width int) string {
	if width <= 0 {
		return ""
	}
	n := utf8.RuneCountInString(s)
	if n >= width {
		r := []rune(s)
		if width <= 1 {
			return string(r[:width])
		}
		return string(r[:width-1]) + "…"
	}
	return s + strings.Repeat(" ", width-n)
}

// ansiPad pads a possibly styled string to a display width of `width` cells.
func ansiPad(s string, width int) string {
	vis := lipgloss.Width(s)
	if vis >= width {
		// Truncate plain text fallback when already too wide (styled cells rarely are).
		if vis > width {
			return lipgloss.NewStyle().MaxWidth(width).Render(s)
		}
		return s
	}
	return s + strings.Repeat(" ", width-vis)
}

func formulaeTableColWidths(total int) (nameW, verW, statusW, tapW int) {
	// Min widths; Name flexes with the remainder. gaps = 3 spaces between 4 cols.
	const gaps = 3
	if total <= gaps+4 {
		// Degenerate pane: give every column at least 1 cell.
		each := max(1, (total-gaps)/4)
		nameW, verW, statusW, tapW = each, each, each, each
		// Fix remainder on Name.
		nameW += total - (nameW + verW + statusW + tapW + gaps)
		if nameW < 1 {
			nameW = 1
		}
		return nameW, verW, statusW, tapW
	}
	verW, statusW, tapW = 16, 10, 8
	if total < verW+statusW+tapW+gaps+8 {
		verW, statusW, tapW = 12, 9, 6
	}
	nameW = total - (verW + statusW + tapW + gaps)
	if nameW < 4 {
		// Steal from fixed columns until Name has room or mins hit.
		need := 4 - nameW
		for need > 0 && (verW > 4 || statusW > 4 || tapW > 3) {
			if verW > 4 {
				verW--
				need--
				continue
			}
			if statusW > 4 {
				statusW--
				need--
				continue
			}
			if tapW > 3 {
				tapW--
				need--
			}
		}
		nameW = total - (verW + statusW + tapW + gaps)
	}
	if nameW < 1 {
		nameW = 1
		// Last resort: force exact sum by shrinking tap/status/ver.
		for nameW+verW+statusW+tapW+gaps > total {
			if tapW > 1 {
				tapW--
			} else if statusW > 1 {
				statusW--
			} else if verW > 1 {
				verW--
			} else {
				break
			}
		}
	}
	// Exact fill: leftover → Name.
	used := nameW + verW + statusW + tapW + gaps
	if used < total {
		nameW += total - used
	}
	return nameW, verW, statusW, tapW
}

// renderFormulaeTable draws a full-pane Name|Version|Status|Tap table.
// Selection follows panel.selected (sidebar is source of truth).
func (m Model) renderFormulaeTable(width, height int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	p := m.panels[PanelFormulae]
	if p == nil || len(p.formulae) == 0 {
		return renderFormulaeEmpty(width, height)
	}

	nameW, verW, statusW, tapW := formulaeTableColWidths(width)
	header := strings.Join([]string{
		style.SubtleText.Bold(true).Render(cellPad("Name", nameW)),
		style.SubtleText.Bold(true).Render(cellPad("Version", verW)),
		style.SubtleText.Bold(true).Render(cellPad("Status", statusW)),
		style.SubtleText.Bold(true).Render(cellPad("Tap", tapW)),
	}, " ")
	sep := style.SubtleText.Render(strings.Repeat("─", width))

	// Body rows available under sticky header + separator.
	bodyH := height - 2
	if bodyH < 1 {
		bodyH = 1
	}
	sel := p.selected
	if sel < 0 {
		sel = 0
	}
	if sel >= len(p.formulae) {
		sel = len(p.formulae) - 1
	}
	// Window so selected stays in view.
	offset := 0
	if sel >= bodyH {
		offset = sel - bodyH + 1
	}
	if offset < 0 {
		offset = 0
	}
	maxOff := len(p.formulae) - bodyH
	if maxOff < 0 {
		maxOff = 0
	}
	if offset > maxOff {
		offset = maxOff
	}

	var rows []string
	end := offset + bodyH
	if end > len(p.formulae) {
		end = len(p.formulae)
	}
	for i := offset; i < end; i++ {
		f := p.formulae[i]
		name := cellPad(f.Name, nameW)
		ver := cellPad(formulaVersionCell(f), verW)
		tap := cellPad(shortenTap(f.Tap), tapW)
		// Style name/version/tap for selection; status keeps badge colors (AC-11).
		rowStyle := style.NormalItem
		if i == sel {
			rowStyle = style.SelectedItem
		}
		st := ansiPad(formulaStatusStyled(f), statusW)
		line := rowStyle.Render(name) + " " + rowStyle.Render(ver) + " " + st + " " + rowStyle.Render(tap)
		rows = append(rows, line)
	}
	// Pad remaining body rows so the block fills height.
	for len(rows) < bodyH {
		rows = append(rows, strings.Repeat(" ", width))
	}

	body := lipgloss.JoinVertical(lipgloss.Top, rows...)
	out := lipgloss.JoinVertical(lipgloss.Top, header, sep, body)
	return lipgloss.NewStyle().Width(width).Height(height).MaxHeight(height).Render(out)
}

func renderFormulaeEmpty(width, height int) string {
	msg := style.SubtleText.Render("No formulae yet") + "\n" +
		style.SubtleText.Render("Press / to search & install")
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(msg)
}
