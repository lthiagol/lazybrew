package gui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// handleMouse routes mouse input when gui.mouse is enabled.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.cfg == nil || !m.cfg.GUI.Mouse {
		return m, nil
	}
	if !m.ready || m.terminalTooSmall {
		return m, nil
	}

	// Count as interaction for auto-refresh pause (same as keys).
	m.lastKeyAt = time.Now()
	m.autoRefreshPaused = false

	if m.showHelp {
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			m.showHelp = false
		}
		return m, nil
	}
	if m.activeModal != nil {
		return m, nil
	}

	if tea.MouseEvent(msg).IsWheel() {
		return m.handleMouseWheel(msg)
	}
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
		return m.handleMouseClick(msg.X, msg.Y)
	}
	return m, nil
}

func (m Model) handleMouseWheel(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	p := m.panels[m.activePanel]
	if p == nil {
		return m, nil
	}
	up := msg.Button == tea.MouseButtonWheelUp
	// Prefer list navigation when the active panel has items.
	if len(p.items) > 0 {
		if up {
			if p.selected > 0 {
				p.selectIndex(p.selected - 1)
			}
		} else if msg.Button == tea.MouseButtonWheelDown {
			if p.selected < len(p.items)-1 {
				p.selectIndex(p.selected + 1)
			}
		}
		if needsTabFetch(m.activePanel, m.activeTab) {
			return m, m.loadTabContent()
		}
		return m, nil
	}
	// Otherwise scroll the content viewport.
	if up {
		m.viewport.LineUp(3)
	} else if msg.Button == tea.MouseButtonWheelDown {
		m.viewport.LineDown(3)
	}
	return m, nil
}

func (m Model) handleMouseClick(x, y int) (tea.Model, tea.Cmd) {
	// Bottom bar occupies the last 2 rows of the view.
	if y >= m.height-2 {
		return m, nil
	}

	sw := sidebarWidth(m.cfg, m.width)
	if x < sw {
		return m.handleSidebarClick(x, y, sw)
	}
	return m.handleMainClick(x, y, sw)
}

func (m Model) handleSidebarClick(x, y, sw int) (tea.Model, tea.Cmd) {
	heights := m.computeContentHeights()
	contentW := sw - 2
	_ = contentW
	yCursor := 0
	for i, h := range heights {
		boxH := h + 2 // rounded border top+bottom
		if y < yCursor || y >= yCursor+boxH {
			yCursor += boxH
			continue
		}
		// Inside panel i (including borders).
		pid := PanelID(i)
		innerY := y - yCursor - 1 // 0 = title row inside content
		var cmd tea.Cmd
		if pid != m.activePanel {
			cmd = m.switchPanel(pid)
		}
		p := m.panels[pid]
		// Title row or empty: just activate panel.
		if innerY <= 0 || p == nil || len(p.items) == 0 {
			return m, cmd
		}
		// Item rows start after the title line.
		rowInWindow := innerY - 1
		if rowInWindow < 0 || rowInWindow >= max(0, h-1) {
			return m, cmd
		}
		idx := p.offset + rowInWindow
		if idx >= 0 && idx < len(p.items) {
			p.selectIndex(idx)
			if needsTabFetch(m.activePanel, m.activeTab) {
				fetch := m.loadTabContent()
				if cmd != nil && fetch != nil {
					return m, tea.Batch(cmd, fetch)
				}
				if fetch != nil {
					return m, fetch
				}
			}
		}
		return m, cmd
	}
	return m, nil
}

func (m Model) handleMainClick(x, y, sw int) (tea.Model, tea.Cmd) {
	// Main panel sits to the right of the sidebar. Outer wrapper uses
	// ActiveBorder (~1 cell) around breadcrumb + tabs + content + log.
	mx := x - sw
	my := y
	if mx < 1 || my < 1 {
		return m, nil
	}
	// Inner coordinates inside the border.
	ix, iy := mx-1, my-1

	mw := m.width - sw - 4
	mh := m.height - 4
	if mw < 1 || mh < 1 {
		return m, nil
	}

	// Vertical stack: breadcrumb (1) + tabBar + content + optional log.
	tabBarH := 0
	if len(m.tabs) > 0 {
		tabBarH = 1
		// Active tab has a thick bottom border → often 2 rows tall.
		if lipgloss.Height(style.TabActive.Render(" x ")) > 1 {
			tabBarH = lipgloss.Height(style.TabActive.Render(" x "))
		}
	}
	breadcrumbH := 1
	cmdLogH, contentH := m.logLayoutHeights(mh)

	// Tab bar click
	tabY0 := breadcrumbH
	tabY1 := tabY0 + tabBarH
	if len(m.tabs) > 0 && iy >= tabY0 && iy < tabY1 {
		return m.handleTabBarClick(ix)
	}

	// Content region
	contentY0 := tabY1
	contentY1 := contentY0 + contentH
	if iy >= contentY0 && iy < contentY1 {
		cy := iy - contentY0
		return m.handleContentClick(ix, cy, mw, contentH)
	}

	_ = cmdLogH
	return m, nil
}

func (m Model) handleTabBarClick(ix int) (tea.Model, tea.Cmd) {
	x := 0
	for i, tab := range m.tabs {
		label := " " + tab.name + " "
		var w int
		if i == m.activeTab {
			w = lipgloss.Width(style.TabActive.Render(label))
		} else {
			w = lipgloss.Width(style.TabInactive.Render(label))
		}
		if ix >= x && ix < x+w {
			if i == m.activeTab {
				return m, nil
			}
			m.activeTab = i
			return m, m.loadTabContent()
		}
		x += w
	}
	return m, nil
}

func (m Model) handleContentClick(ix, cy, contentW, contentH int) (tea.Model, tea.Cmd) {
	// Formulae List table: sticky header + separator, then body rows.
	if m.activePanel == PanelFormulae && m.activeTab == 0 {
		p := m.panels[PanelFormulae]
		if p == nil || len(p.formulae) == 0 {
			return m, nil
		}
		// header + sep = 2 rows
		if cy < 2 {
			return m, nil
		}
		bodyH := contentH - 2
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
		offset := 0
		if sel >= bodyH {
			offset = sel - bodyH + 1
		}
		maxOff := len(p.formulae) - bodyH
		if maxOff < 0 {
			maxOff = 0
		}
		if offset > maxOff {
			offset = maxOff
		}
		row := offset + (cy - 2)
		if row < 0 || row >= len(p.formulae) {
			return m, nil
		}
		// Keep items/formulae selection in sync.
		if len(p.items) == len(p.formulae) {
			p.selectIndex(row)
		} else {
			p.selected = row
		}
		return m, nil
	}

	// Other list-backed panels: clicking main content does not change
	// selection (sidebar is the list). No-op.
	_ = ix
	_ = contentW
	return m, nil
}
