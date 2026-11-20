package gui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

	mouse := msg.Mouse()
	_, isClick := msg.(tea.MouseClickMsg)
	leftClick := isClick && mouse.Button == tea.MouseLeft

	if m.showHelp {
		if leftClick {
			m.showHelp = false
		}
		return m, nil
	}
	if m.activeModal != nil {
		return m, nil
	}

	if _, isWheel := msg.(tea.MouseWheelMsg); isWheel {
		return m.handleMouseWheel(msg)
	}
	if leftClick {
		return m.handleMouseClick(mouse.X, mouse.Y)
	}
	return m, nil
}

func (m Model) handleMouseWheel(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	p := m.panels[m.activePanel]
	if p == nil {
		return m, nil
	}
	button := msg.Mouse().Button
	up := button == tea.MouseWheelUp
	// Prefer list navigation when the active panel has items.
	if len(p.items) > 0 {
		if up {
			if p.selected > 0 {
				p.selectIndex(p.selected - 1)
			}
		} else if button == tea.MouseWheelDown {
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
		m.viewport.ScrollUp(3)
	} else if button == tea.MouseWheelDown {
		m.viewport.ScrollDown(3)
	}
	return m, nil
}

func (m Model) handleMouseClick(x, y int) (tea.Model, tea.Cmd) {
	// The header and the bottom bar are not clickable. Below the header, y
	// counts from the top of the body, where the body is the screen layout's
	// rows.
	y -= headerRows
	if y < 0 || y >= m.screenLayout().bodyH {
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
		// Inside panel i (including borders). The title sits in the top
		// border, so innerY 0 is the first item row.
		pid := PanelID(i)
		innerY := y - yCursor - 1
		var cmd tea.Cmd
		if pid != m.activePanel {
			cmd = m.switchPanel(pid)
		}
		p := m.panels[pid]
		// Top border or empty: just activate panel.
		if innerY < 0 || p == nil || len(p.items) == 0 {
			return m, cmd
		}
		// Bottom border, below the body.
		if innerY >= h {
			return m, cmd
		}
		idx := p.offset + innerY
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
	// Main panel sits to the right of the sidebar. Its ActiveBorder (1 cell)
	// wraps tabs + content; the Log box is a separate box below.
	mx := x - sw
	my := y
	if mx < 1 || my < 1 {
		return m, nil
	}
	// Inner coordinates inside the border.
	ix, iy := mx-1, my-1

	l := m.screenLayout()
	if l.mainW < 3 || l.mainH < 3 {
		return m, nil
	}

	// Vertical stack inside the border: tab bar + content. The breadcrumb
	// is the border title now. The Log box sits below the main panel's
	// border, so it has no rows here.
	tabBarH := m.tabBarRows()

	// Tab bar click
	tabY0 := 0
	tabY1 := tabY0 + tabBarH
	if len(m.tabs) > 0 && iy >= tabY0 && iy < tabY1 {
		return m.handleTabBarClick(ix)
	}

	// Content region
	contentY0 := tabY1
	contentY1 := contentY0 + l.contentH
	if iy >= contentY0 && iy < contentY1 {
		cy := iy - contentY0
		return m.handleContentClick(ix, cy, l.mainW-2, l.contentH)
	}

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
