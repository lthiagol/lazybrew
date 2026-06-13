package gui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/brew"
	"github.com/thiago/lazybrew/internal/gui/style"
)

type PanelID int

const (
	PanelStatus PanelID = iota
	PanelFormulae
	PanelCasks
	PanelOutdated
	PanelTaps
	PanelServices
	PanelSearch
)

func (p PanelID) String() string {
	switch p {
	case PanelStatus:
		return "Status"
	case PanelFormulae:
		return "Formulae"
	case PanelCasks:
		return "Casks"
	case PanelOutdated:
		return "Outdated"
	case PanelTaps:
		return "Taps"
	case PanelServices:
		return "Services"
	case PanelSearch:
		return "Search"
	default:
		return "?"
	}
}

type tabInfo struct {
	name string
	id   int
}

var panelTabs = map[PanelID][]tabInfo{
	PanelStatus:   {{"Dashboard", 0}, {"Config", 1}, {"Doctor", 2}},
	PanelFormulae: {{"Info", 0}, {"Deps", 1}, {"Used By", 2}, {"Caveats", 3}, {"Files", 4}},
	PanelCasks:    {{"Info", 0}, {"Deps", 1}, {"Caveats", 2}},
	PanelOutdated: {{"Info", 0}, {"Versions", 1}},
	PanelTaps:     {{"Tap Info", 0}, {"Trust", 1}, {"Formulae", 2}},
	PanelServices: {{"Status", 0}},
	PanelSearch:   {{"Info", 0}},
}

type panelData struct {
	id              PanelID
	title           string
	icon            string
	items           []string
	unfilteredItems []string
	leavesActive    bool
	selected        int
	offset          int
	width           int
	height          int
	active          bool
	loading         bool
	err             error
	formulae        []brew.Formula
	casks           []brew.Cask
	taps            []brew.Tap
	services        []brew.Service
}

func (p *panelData) itemCount() int {
	return len(p.items)
}

func (p *panelData) visibleCount() int {
	return max(1, p.height-3)
}

func (p *panelData) up() {
	if p.selected > 0 {
		p.selected--
	}
	if p.selected < p.offset {
		p.offset = p.selected
	}
}

func (p *panelData) down() {
	if p.selected < len(p.items)-1 {
		p.selected++
	}
	if p.selected >= p.offset+p.visibleCount() {
		p.offset = p.selected - p.visibleCount() + 1
	}
}

func (p *panelData) selectedItem() string {
	if p.selected >= 0 && p.selected < len(p.items) {
		return p.items[p.selected]
	}
	return ""
}

func (p *panelData) selectedFormula() *brew.Formula {
	if p.selected >= 0 && p.selected < len(p.formulae) {
		return &p.formulae[p.selected]
	}
	return nil
}

func (p *panelData) selectedCask() *brew.Cask {
	if p.selected >= 0 && p.selected < len(p.casks) {
		return &p.casks[p.selected]
	}
	return nil
}

func (p *panelData) selectedTap() *brew.Tap {
	if p.selected >= 0 && p.selected < len(p.taps) {
		return &p.taps[p.selected]
	}
	return nil
}

func (p *panelData) selectedService() *brew.Service {
	if p.selected >= 0 && p.selected < len(p.services) {
		return &p.services[p.selected]
	}
	return nil
}

func (p *panelData) renderList(width, height int, batch map[int]bool) string {
	if p.loading {
		return lipgloss.NewStyle().Width(width).Height(height).Render(style.SubtleText.Render("Loading..."))
	}

	if p.err != nil {
		return lipgloss.NewStyle().Width(width).Height(height).Render(style.ErrorBadge.Render("Error: " + p.err.Error()))
	}

	visible := min(p.visibleCount(), height)
	if visible < 1 {
		visible = 1
	}
	end := min(p.offset+visible, len(p.items))
	if p.offset >= end && p.offset > 0 {
		p.offset = max(0, end-visible)
		end = min(p.offset+visible, len(p.items))
	}
	items := p.items[p.offset:end]

	lines := make([]string, 0, len(items))
	for i, item := range items {
		idx := p.offset + i
		prefix := "  "
		if batch != nil && batch[idx] {
			prefix = "● "
		}
		itemStyle := style.NormalItem
		if idx == p.selected {
			prefix = "▸ "
			itemStyle = style.SelectedItem
		}
		if batch != nil && batch[idx] && idx == p.selected {
			prefix = "▸●"
			itemStyle = style.SelectedItem
		}
		rendered := lipgloss.NewStyle().Width(width).MaxWidth(width).Render(prefix + item)
		lines = append(lines, itemStyle.Render(rendered))
	}

	if len(lines) == 0 {
		return lipgloss.NewStyle().Width(width).Height(height).Render(style.SubtleText.Render(emptyMessage(p.id)))
	}

	return lipgloss.JoinVertical(lipgloss.Top, lines...)
}

func emptyPanel(width, height int) string {
	return lipgloss.NewStyle().Width(width).Height(height).Render(
		style.SubtleText.Render("No selection"),
	)
}

func (p *panelData) renderSidebarContent(width, maxRows int) string {
	if p.loading {
		return lipgloss.NewStyle().Width(width).Render(style.SubtleText.Render("..."))
	}
	if p.err != nil {
		return lipgloss.NewStyle().Width(width).Render(style.ErrorBadge.Render("!"))
	}
	count := len(p.items)
	if count == 0 {
		return style.SubtleText.Render("(empty)")
	}
	if p.selected >= count {
		p.selected = max(0, count-1)
	}
	if p.offset >= count {
		p.offset = max(0, count-maxRows)
	}
	visible := min(maxRows, count-p.offset)
	if visible <= 0 {
		return ""
	}
	end := p.offset + visible
	slice := p.items[p.offset:end]
	lines := make([]string, 0, visible)
	for i, item := range slice {
		idx := p.offset + i
		prefix := "  "
		itemStyle := style.NormalItem
		if idx == p.selected {
			prefix = "▸ "
			itemStyle = style.SelectedItem
		}
		text := truncateWithEllipsis(prefix+item, width)
		lines = append(lines, itemStyle.Render(text))
	}
	return lipgloss.JoinVertical(lipgloss.Top, lines...)
}

func truncateWithEllipsis(s string, maxWidth int) string {
	runes := []rune(s)
	if len(runes) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return string(runes[:maxWidth])
	}
	return string(runes[:maxWidth-3]) + "..."
}

func emptyMessage(id PanelID) string {
	switch id {
	case PanelFormulae:
		return "No formulae installed"
	case PanelCasks:
		return "No casks installed"
	case PanelOutdated:
		return "Everything up to date!"
	case PanelTaps:
		return "No custom taps"
	case PanelServices:
		return "No services configured"
	case PanelSearch:
		return "No results"
	default:
		return "No data"
	}
}

func initPanels() []*panelData {
	panels := make([]*panelData, 7)
	panelDefs := []struct {
		id    PanelID
		title string
	}{
		{PanelStatus, "Status"},
		{PanelFormulae, "Formulae"},
		{PanelCasks, "Casks"},
		{PanelOutdated, "Outdated"},
		{PanelTaps, "Taps"},
		{PanelServices, "Services"},
		{PanelSearch, "Search"},
	}
	for i, def := range panelDefs {
		panels[i] = &panelData{
			id:      def.id,
			title:   def.title,
			loading: true,
		}
	}
	panels[0].active = true
	return panels
}
