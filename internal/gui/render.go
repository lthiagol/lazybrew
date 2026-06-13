package gui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/gui/presentation"
	"github.com/thiago/lazybrew/internal/gui/style"
)

func (m Model) renderSidebar() string {
	sw := sidebarWidth(m.cfg, m.width)
	contentWidth := sw - 2
	heights := m.computeContentHeights()

	var boxes []string
	for i, p := range m.panels {
		title := p.title
		if p.loading {
			title += style.SubtleText.Render("  …")
		} else if count := p.itemCount(); count > 0 {
			title += style.SubtleText.Render("  " + strconv.Itoa(count))
		}
		titleLine := style.PanelTitle.Render(title)
		itemsMaxRows := max(0, heights[i]-1)
		itemsContent := p.renderSidebarContent(contentWidth, itemsMaxRows)
		fullContent := lipgloss.JoinVertical(lipgloss.Top, titleLine, itemsContent)
		box := renderBox(fullContent, contentWidth, heights[i], i == int(m.activePanel))
		boxes = append(boxes, box)
	}

	return lipgloss.JoinVertical(lipgloss.Top, boxes...)
}

func (m Model) renderMainPanel() string {
	sw := sidebarWidth(m.cfg, m.width)
	mw := m.width - sw - 4
	mh := m.height - 4

	panelName := m.panels[m.activePanel].title
	tabName := ""
	if len(m.tabs) > 0 && m.activeTab < len(m.tabs) {
		tabName = m.tabs[m.activeTab].name
	}
	breadcrumb := style.PanelTitle.Render(panelName)
	if tabName != "" {
		breadcrumb += style.SubtleText.Render(" › ") + style.AccentText.Render(tabName)
	}

	tabBar := m.renderTabBar(mw)
	content := m.renderContent(mw, mh-4)

	panel := lipgloss.JoinVertical(lipgloss.Top, breadcrumb, tabBar, content)

	return lipgloss.NewStyle().
		Width(mw + 2).
		Height(mh).
		Render(style.ActiveBorder.Render(panel))
}

func (m Model) renderTabBar(width int) string {
	if len(m.tabs) == 0 {
		return ""
	}

	var tabs []string
	for i, tab := range m.tabs {
		if i == m.activeTab {
			tabs = append(tabs, style.TabActive.Render(" "+tab.name+" "))
		} else {
			tabs = append(tabs, style.TabInactive.Render(" "+tab.name+" "))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m Model) computeContentHeights() []int {
	n := len(m.panels)
	if n == 0 {
		return nil
	}
	sidebarHeight := m.height - 4
	borderOverhead := n * 2
	availableRows := sidebarHeight - borderOverhead
	if availableRows < n {
		availableRows = n
	}
	minActive := 4
	minInactive := 2
	needed := minActive + (n-1)*minInactive
	if availableRows < needed {
		heights := make([]int, n)
		for i := range heights {
			if i == int(m.activePanel) {
				heights[i] = max(1, availableRows-(n-1))
			} else {
				heights[i] = 1
			}
		}
		return heights
	}
	activeRows := availableRows * 40 / 100
	if activeRows < minActive {
		activeRows = minActive
	}
	remaining := availableRows - activeRows
	inactiveRows := remaining / max(1, n-1)
	if inactiveRows < minInactive {
		inactiveRows = minInactive
	}
	heights := make([]int, n)
	for i := range heights {
		if i == int(m.activePanel) {
			heights[i] = activeRows
		} else {
			heights[i] = inactiveRows
		}
	}
	return heights
}

func (m Model) renderContentInViewport(content string) string {
	m.viewport.SetContent(content)
	m.viewport.Width = max(10, m.width-4)
	m.viewport.Height = max(10, m.height-6)
	return m.viewport.View()
}

func (m Model) renderContent(width, height int) string {
	panel := m.panels[m.activePanel]
	if panel.loading {
		return style.SubtleText.Render("Loading...")
	}

	switch m.activePanel {
	case PanelFormulae:
		switch m.activeTab {
		case 0:
			f := panel.selectedFormula()
			if f == nil {
				return emptyPanel(width, height)
			}
			info := presentation.FormatFormulaInfo(*f, width)
			return style.NormalItem.Render(info)
		case 1, 2, 4:
			itemName := selectedItemName(panel)
			key := tabKey(m.activePanel, m.activeTab, itemName)
			if content, ok := m.tabContent[key]; ok {
				return m.renderContentInViewport(content)
			}
			return style.SubtleText.Render("Loading...")
		case 3:
			f := panel.selectedFormula()
			if f == nil || f.Caveats == "" {
				return style.SubtleText.Render("No caveats")
			}
			return style.NormalItem.Render(f.Caveats)
		}

	case PanelCasks:
		switch m.activeTab {
		case 0:
			c := panel.selectedCask()
			if c == nil {
				return emptyPanel(width, height)
			}
			info := presentation.FormatCaskInfo(*c, width)
			return style.NormalItem.Render(info)
		case 1:
			c := panel.selectedCask()
			if c == nil || len(c.DependsOn) == 0 {
				return style.SubtleText.Render("No dependencies")
			}
			result := ""
			for _, d := range c.DependsOn {
				result += d + "\n"
			}
			return m.renderContentInViewport(result)
		case 2:
			c := panel.selectedCask()
			if c == nil {
				return style.SubtleText.Render("No cask selected")
			}
			result := "Description: " + c.Description + "\n"
			if len(c.Artifacts) > 0 {
				result += "\nArtifacts:\n"
				for _, a := range c.Artifacts {
					result += "  " + a + "\n"
				}
			}
			return m.renderContentInViewport(result)
		}

	case PanelStatus:
		switch m.activeTab {
		case 0:
			return panel.renderList(width, height, nil)
		case 1:
			itemName := selectedItemName(panel)
			key := tabKey(m.activePanel, m.activeTab, itemName)
			if content, ok := m.tabContent[key]; ok {
				return m.renderContentInViewport(content)
			}
			return style.SubtleText.Render("Loading...")
		case 2:
			itemName := selectedItemName(panel)
			key := tabKey(m.activePanel, m.activeTab, itemName)
			if content, ok := m.tabContent[key]; ok {
				return m.renderContentInViewport(content)
			}
			return style.SubtleText.Render("Loading...")
		}

	case PanelTaps:
		switch m.activeTab {
		case 0:
			return panel.renderList(width, height, nil)
		case 1:
			t := panel.selectedTap()
			if t == nil {
				return style.SubtleText.Render("No tap selected")
			}
			trusted := "untrusted"
			if t.Trusted || t.IsOfficial {
				trusted = "trusted"
			}
			return style.NormalItem.Render("Tap: " + t.Name + "\n" +
				"Trusted: " + trusted + "\n" +
				"Official: " + boolStr(t.IsOfficial) + "\n" +
				"Remote: " + t.Remote + "\n" +
				"Formulae: " + strconv.Itoa(t.FormulaCount) + "\n" +
				"Casks: " + strconv.Itoa(t.CaskCount) + "\n")
		case 2:
			t := panel.selectedTap()
			if t == nil || len(t.FormulaNames) == 0 {
				return style.SubtleText.Render("No formulae in this tap")
			}
			result := ""
			for _, fn := range t.FormulaNames {
				result += fn + "\n"
			}
			return style.NormalItem.Render(result)
		}

	case PanelServices:
		return panel.renderList(width, height, nil)

	case PanelOutdated:
		switch m.activeTab {
		case 0:
			f := panel.selectedFormula()
			if f != nil {
				info := presentation.FormatFormulaInfo(*f, width)
				return style.NormalItem.Render(info)
			}
			c := panel.selectedCask()
			if c != nil {
				info := presentation.FormatCaskInfo(*c, width)
				return style.NormalItem.Render(info)
			}
			return emptyPanel(width, height)
		default:
			return panel.renderList(width, height, m.batch.selected)
		}

	case PanelSearch:
		if m.searchInfoContent == "" {
			return style.SubtleText.Render("No package selected")
		}
		return m.renderSearchInfo(width, height)
	}

	return panel.renderList(width, height, nil)
}

func (m Model) renderSearchInfo(width, height int) string {
	info, err := parsePackageInfo(m.searchInfoContent)
	if err != nil {
		return style.ErrorBadge.Render("Parse error: " + err.Error())
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Name:     %s", info.Name))

	ver := info.Version
	if info.Bottled {
		ver += " (bottled)"
	}
	lines = append(lines, fmt.Sprintf("Version:  %s", ver))
	lines = append(lines, fmt.Sprintf("Type:     %s", info.Type))

	status := "not installed"
	if info.Installed {
		status = fmt.Sprintf("installed (%s)", info.InstallPath)
	}
	lines = append(lines, fmt.Sprintf("Status:   %s", status))
	if info.License != "" {
		lines = append(lines, fmt.Sprintf("License:  %s", info.License))
	}
	if info.Homepage != "" {
		lines = append(lines, fmt.Sprintf("Homepage: %s", truncateWithEllipsis(info.Homepage, width-12)))
	}

	all := lipgloss.JoinVertical(lipgloss.Top, lines...)
	result := all

	if info.Description != "" {
		desc := "\n\n" + style.AccentText.Render("Description:") + "\n" + info.Description
		result += desc
	}
	if len(info.Dependencies) > 0 {
		deps := "\n\n" + style.AccentText.Render("Dependencies:") + "\n"
		for _, d := range info.Dependencies {
			deps += "  " + d + "\n"
		}
		result += deps
	}
	if info.Caveats != "" {
		cav := "\n\n" + style.AccentText.Render("Caveats:") + "\n" + info.Caveats
		result += cav
	}

	return style.NormalItem.Render(result)
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func (m Model) renderBottomBar() string {
	hints := panelHints(m.activePanel)
	globalHints := []keyHint{{"Tab", "next"}, {"S-Tab", "prev"}, {"1-7", "jump"}, {"/", "search"}, {"?", "help"}, {"R", "refresh"}, {"q", "quit"}}

	var parts []string
	for _, h := range hints {
		parts = append(parts, style.HintKey.Render(h.key)+" "+style.HintDesc.Render(h.desc))
	}
	for _, h := range globalHints {
		parts = append(parts, style.HintKey.Render(h.key)+" "+style.HintDesc.Render(h.desc))
	}

	updateStatus := m.updateStatusText()
	if updateStatus != "" {
		parts = append(parts, style.SubtleText.Render("│"), updateStatus)
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	return lipgloss.NewStyle().
		Width(m.width - 2).
		Padding(0, 1).
		Render(bar)
}

func (m Model) updateStatusText() string {
	if m.isUpdating {
		return style.SubtleText.Render("⟳ Updating...")
	}
	if !m.lastUpdate.IsZero() {
		ago := time.Since(m.lastUpdate).Round(time.Second)
		return style.SubtleText.Render(fmt.Sprintf("⟳ Updated %s ago", ago))
	}
	if m.cfg.Brew.UpdateOnStart {
		return style.SubtleText.Render("⟳ Never updated")
	}
	return ""
}
