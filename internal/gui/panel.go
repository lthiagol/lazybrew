package gui

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
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

// The Formulae tab indices, in the order panelTabs lists them. Code names a
// tab by these, never by a bare number, so a reordered bar changes one list.
// The Info tab is first, so a panel's default tab is its details.
const (
	formulaInfoTab = iota
	formulaDepsTab
	formulaUsesTab
	formulaCaveatsTab
	formulaFilesTab
)

var panelTabs = map[PanelID][]tabInfo{
	PanelStatus:   {{"Dashboard", statusDashboardTab}, {"Config", statusConfigTab}, {"Doctor", statusDoctorTab}},
	PanelFormulae: {{"Info", formulaInfoTab}, {"Deps", formulaDepsTab}, {"Used By", formulaUsesTab}, {"Caveats", formulaCaveatsTab}, {"Files", formulaFilesTab}},
	PanelCasks:    {{"Info", 0}, {"Deps", 1}, {"Caveats", 2}},
	PanelOutdated: {{"Info", outdatedInfoTab}, {"Versions", outdatedVersionsTab}},
	PanelTaps:     {{"Tap Info", tapInfoTab}, {"Trust", tapTrustTab}, {"Formulae", tapFormulaeTab}},
	PanelServices: {{"Info", serviceInfoTab}, {"Logs", servicesLogsTab}},
	PanelSearch:   {},
}

type panelData struct {
	id              PanelID
	title           string
	icon            string
	items           []string
	unfilteredItems []string
	// unfilteredFormulae holds the formulae the leaves filter narrows, so leaves
	// off can restore them. formulae is index-aligned with items.
	unfilteredFormulae []brew.Formula
	leavesActive       bool
	selected           int
	offset             int
	width              int
	height             int
	active             bool
	loading            bool
	err                error
	formulae           []brew.Formula
	casks              []brew.Cask
	// installed is the installed list the package tables read, built by
	// indexInstalled when the panel's data lands. installedGen counts its builds.
	installed    map[string]installedPkg
	installedGen int
	taps         []brew.Tap
	services     []brew.Service
	serviceInfo  []brew.ServiceInfo
	visibleRows  int
	// results is the Search panel's raw rows, index-aligned with items.
	results []brew.SearchResult
	// loaded is set once a fetch result has landed for this panel, error or
	// not. It tells "never fetched" (not loaded) from "fetched, nothing
	// there" (empty), which items alone cannot.
	loaded bool
	// trustUnknown is set when the taps were read without tap-info: the rows
	// and the Tap Info say so, and t and T refuse until a full read lands.
	trustUnknown bool
}

func (p *panelData) itemCount() int {
	return len(p.items)
}

func (p *panelData) visibleCount() int {
	if p.visibleRows > 0 {
		return p.visibleRows
	}
	return max(1, p.height-3)
}

func (p *panelData) up() {
	if p.selected > 0 {
		p.selected--
	} else if len(p.items) > 0 {
		p.selected = len(p.items) - 1
		p.offset = max(0, p.selected-p.visibleCount()+1)
		return
	}
	if p.selected < p.offset {
		p.offset = p.selected
	}
}

func (p *panelData) down() {
	if p.selected < len(p.items)-1 {
		p.selected++
	} else if len(p.items) > 0 {
		p.selected = 0
		p.offset = 0
		return
	}
	if p.selected >= p.offset+p.visibleCount() {
		p.offset = p.selected - p.visibleCount() + 1
	}
}

// selectIndex sets the cursor to i without wrap-around (used by mouse clicks).
func (p *panelData) selectIndex(i int) {
	if len(p.items) == 0 {
		return
	}
	if i < 0 {
		i = 0
	}
	if i >= len(p.items) {
		i = len(p.items) - 1
	}
	p.selected = i
	vc := p.visibleCount()
	if vc < 1 {
		vc = 1
	}
	if p.selected < p.offset {
		p.offset = p.selected
	}
	if p.selected >= p.offset+vc {
		p.offset = p.selected - vc + 1
	}
	if p.offset < 0 {
		p.offset = 0
	}
}

// normalize clamps selected and offset to the current items and viewport
// size. It is called from the Update loop (WindowSizeMsg, DataLoadedMsg,
// SearchDoneMsg) so rendering never mutates model state (r1 finding 29).
func (p *panelData) normalize() {
	p.normalizeToHeight(p.visibleCount())
}

// normalizeToHeight clamps selected and offset like normalize, but
// against an explicit viewport row count. The C (command-log) toggle
// changes the main list height without a WindowSizeMsg, so callers pass
// the effective render height (r3 finding R3-4).
func (p *panelData) normalizeToHeight(vc int) {
	count := len(p.items)
	if count == 0 {
		p.selected = 0
		p.offset = 0
		return
	}
	if p.selected >= count {
		p.selected = count - 1
	}
	if p.selected < 0 {
		p.selected = 0
	}
	if vc < 1 {
		vc = 1
	}
	if p.selected < p.offset {
		p.offset = p.selected
	}
	if p.selected >= p.offset+vc {
		p.offset = p.selected - vc + 1
	}
	if p.offset >= count {
		p.offset = max(0, count-vc)
	}
	if p.offset < 0 {
		p.offset = 0
	}
}

func (p *panelData) selectedItem() string {
	if p.selected >= 0 && p.selected < len(p.items) {
		return p.items[p.selected]
	}
	return ""
}

// nameAt is the real name of the row at index i, read from the typed slice the
// panel's rows were built from. It never parses the display string, which cuts
// a name longer than its column. "" when there is no such row.
func (p *panelData) nameAt(i int) string {
	if i < 0 {
		return ""
	}
	switch p.id {
	case PanelFormulae:
		if i < len(p.formulae) {
			return p.formulae[i].Name
		}
	case PanelCasks:
		if i < len(p.casks) {
			return p.casks[i].Name
		}
	case PanelOutdated:
		// Outdated lists the formulae, then the casks.
		if i < len(p.formulae) {
			return p.formulae[i].Name
		}
		if j := i - len(p.formulae); j < len(p.casks) {
			return p.casks[j].Name
		}
	case PanelTaps:
		if i < len(p.taps) {
			return p.taps[i].Name
		}
	case PanelServices:
		if i < len(p.services) {
			return p.services[i].Name
		}
	case PanelSearch:
		if i < len(p.results) {
			return p.results[i].Name
		}
	}
	return ""
}

// selectedName is the real name of the selected row; see nameAt.
func (p *panelData) selectedName() string {
	return p.nameAt(p.selected)
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

// selectedOutdatedFormula returns the formula for the cursor when the
// panel mixes formulae and casks (the Outdated panel): the raw item index
// is valid for formulae as long as it lands inside the formulae slice.
func (p *panelData) selectedOutdatedFormula() *brew.Formula {
	if p.selected >= 0 && p.selected < len(p.formulae) {
		return &p.formulae[p.selected]
	}
	return nil
}

// selectedOutdatedCask returns the cask for the cursor when the panel
// mixes formulae and casks (the Outdated panel): cask rows live past the
// formulae, so the cask index is selected - len(formulae).
func (p *panelData) selectedOutdatedCask() *brew.Cask {
	idx := p.selected - len(p.formulae)
	if idx >= 0 && idx < len(p.casks) {
		return &p.casks[idx]
	}
	return nil
}

func (p *panelData) selectedService() *brew.Service {
	if p.selected >= 0 && p.selected < len(p.services) {
		return &p.services[p.selected]
	}
	return nil
}

// serviceInfoFor is the Info view's entry for the named service, nil when
// `brew services info` has not loaded it.
func (p *panelData) serviceInfoFor(name string) *brew.ServiceInfo {
	for i := range p.serviceInfo {
		if p.serviceInfo[i].Name == name {
			return &p.serviceInfo[i]
		}
	}
	return nil
}

// selectServiceNamed moves the cursor to the named service when the list still
// has it. An empty or unknown name leaves the cursor where it is.
func (p *panelData) selectServiceNamed(name string) {
	for i, s := range p.services {
		if name != "" && s.Name == name {
			p.selected = i
			return
		}
	}
}

// selectTapNamed moves the cursor to the named tap when the list still has it.
// Otherwise the cursor stays on its index, which after an untap is the tap that
// followed the removed one (v2-10 AC-04.1).
func (p *panelData) selectTapNamed(name string) {
	for i, t := range p.taps {
		if name != "" && t.Name == name {
			p.selected = i
			return
		}
	}
}

func emptyPanel(width, height int) string {
	return lipgloss.NewStyle().Width(width).Height(height).Render(
		style.SubtleText.Render("No selection"),
	)
}

// renderSidebarContent is the body of a sidebar box. A pending fetch, a panel
// that was never loaded and an empty one draw no body: their state is the
// box's border label (sidebarLabel).
func (p *panelData) renderSidebarContent(width, maxRows int, batch map[int]bool, ic style.Icons, focused bool) string {
	if p.loading {
		return ""
	}
	if p.err != nil {
		msg := p.err.Error()
		if width > 10 {
			// Cut by display cells: a byte cut can split a multibyte rune.
			msg = ansi.Truncate(msg, width-2, "...")
		}
		return lipgloss.NewStyle().Width(width).Render(style.ErrorBadge.Render(msg))
	}
	count := len(p.items)
	if count == 0 {
		return ""
	}
	// Cursor/offset clamping happens in normalize() from the Update loop;
	// rendering must not mutate model state (r1 finding 29).
	visible := min(maxRows, count-p.offset)
	if visible <= 0 {
		return ""
	}
	lines := make([]string, 0, visible)
	for i := range visible {
		lines = append(lines, p.sidebarRow(p.offset+i, width, batch, ic, focused))
	}
	return lipgloss.JoinVertical(lipgloss.Top, lines...)
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
		// M12: every panel starts with loading=false. Init only fetches
		// the active panel (Status by default); everything else lazy-
		// loads on first switchPanel via lazyLoadPanel(). RefreshMsg
		// sets loading=true on the panels it re-fetches.
		loading := false
		panels[i] = &panelData{
			id:      def.id,
			title:   def.title,
			loading: loading,
		}
	}
	panels[0].active = true
	return panels
}
