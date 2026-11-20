package gui

import (
	"slices"
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
		// A toast is drawn over the panels, so a click on one dismisses it
		// and goes no further.
		if i := m.toastAt(mouse.X, mouse.Y); i >= 0 {
			m.toasts = slices.Delete(slices.Clone(m.toasts), i, i+1)
			return m, nil
		}
		return m.handleMouseClick(mouse.X, mouse.Y)
	}
	return m, nil
}

// handleMouseWheel routes a wheel notch by where the pointer is. Over a sidebar
// box it moves that box's list, focusing the box as a click does. Over the main
// panel it scrolls the tab content, three rows a notch.
func (m Model) handleMouseWheel(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	mouse := msg.Mouse()
	delta := 0
	switch mouse.Button {
	case tea.MouseWheelUp:
		delta = -1
	case tea.MouseWheelDown:
		delta = 1
	default:
		return m, nil
	}

	y := mouse.Y - headerRows
	l := m.screenLayout()
	if mouse.X >= sidebarWidth(m.cfg, m.width) {
		// The Log box is below the main panel; a notch there scrolls nothing.
		if y < 0 || y >= l.mainH {
			return m, nil
		}
		return m.scrollContent(delta * 3), nil
	}

	if y < 0 || y >= l.bodyH {
		return m, nil
	}
	pid, innerY, h, ok := m.sidebarBoxAt(y)
	if !ok {
		return m, nil
	}
	var cmd tea.Cmd
	if pid != m.activePanel {
		cmd = m.switchPanel(pid)
	}
	// A border row has no list row under the pointer.
	if innerY < 0 || innerY >= h {
		return m, cmd
	}
	p := m.panels[pid]
	if next := p.selected + delta; next >= 0 && next < len(p.items) {
		p.selectIndex(next)
	}
	return m, tea.Batch(cmd, m.loadTabContent())
}

// sidebarBoxAt returns the sidebar box under body row y, and the row within its
// body: innerY is below zero on the top border and at least h on the bottom
// one. ok is false when no box is under y.
func (m Model) sidebarBoxAt(y int) (pid PanelID, innerY, h int, ok bool) {
	heights := m.computeContentHeights()
	yCursor := 0
	for i, boxBody := range heights {
		boxH := boxBody + 2 // rounded border top and bottom
		if y >= yCursor && y < yCursor+boxH {
			return PanelID(i), y - yCursor - 1, boxBody, true
		}
		yCursor += boxH
	}
	return 0, 0, 0, false
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
		return m.handleSidebarClick(y)
	}
	return m.handleMainClick(x, y, sw)
}

func (m Model) handleSidebarClick(y int) (tea.Model, tea.Cmd) {
	pid, innerY, h, ok := m.sidebarBoxAt(y)
	if !ok {
		return m, nil
	}
	var cmd tea.Cmd
	if pid != m.activePanel {
		cmd = m.switchPanel(pid)
	}
	p := m.panels[pid]
	// Top border or empty: just activate panel.
	if innerY < 0 || len(p.items) == 0 {
		return m, cmd
	}
	// Bottom border, below the body.
	if innerY >= h {
		return m, cmd
	}
	idx := p.offset + innerY
	if idx >= 0 && idx < len(p.items) {
		p.selectIndex(idx)
		return m, tea.Batch(cmd, m.loadTabContent())
	}
	return m, cmd
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

	// A click on the content selects nothing: the sidebar is the list.
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
