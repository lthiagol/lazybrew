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
	sh := m.height - 4

	var items []string
	for i := range m.panels {
		item := m.panels[i].renderSidebarItem(sw - 4)
		items = append(items, item)
	}

	list := lipgloss.JoinVertical(lipgloss.Top, items...)
	border := style.InactiveBorder

	return lipgloss.NewStyle().
		Width(sw).
		Height(sh).
		Render(border.Render(list))
}

func (m Model) renderMainPanel() string {
	sw := sidebarWidth(m.cfg, m.width)
	mw := m.width - sw - 4
	mh := m.height - 4

	tabBar := m.renderTabBar(mw)
	content := m.renderContent(mw, mh-3)

	panel := lipgloss.JoinVertical(lipgloss.Top, tabBar, content)

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
		return panel.renderList(width, height, nil)
	}

	return panel.renderList(width, height, nil)
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
