package gui

import (
	"fmt"

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

// Title is the name the panel's sidebar box and its help page show. It is String
// except where the panel does more than its identifier says.
func (p PanelID) Title() string {
	if p == PanelTaps {
		return "Taps & Trust"
	}
	return p.String()
}

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
	PanelStatus:   {{"Dashboard", statusDashboardTab}, {"Config", statusConfigTab}, {"Doctor", statusDoctorTab}, {"Missing", statusMissingTab}, {"Vulns", statusVulnsTab}, {"Cleanup", statusCleanupTab}, {"Unused", statusUnusedTab}},
	PanelFormulae: {{"Info", formulaInfoTab}, {"Deps", formulaDepsTab}, {"Used By", formulaUsesTab}, {"Caveats", formulaCaveatsTab}, {"Files", formulaFilesTab}},
	PanelCasks:    {{"Info", caskInfoTab}, {"Deps", caskDepsTab}, {"Caveats", caskCaveatsTab}},
	PanelOutdated: {{"Info", outdatedInfoTab}, {"Versions", outdatedVersionsTab}},
	PanelTaps:     {{"Tap Info", tapInfoTab}, {"Trust", tapTrustTab}, {"Formulae", tapFormulaeTab}},
	PanelServices: {{"Info", serviceInfoTab}, {"Logs", servicesLogsTab}},
	PanelSearch:   {},
}

type panelData struct {
	id    PanelID
	title string
	// items holds one entry per row. On the package, tap, service and Search
	// panels it is the bare name; Status has no typed rows, so it is its drawn
	// text.
	items           []string
	unfilteredItems []string
	// unfiltered* hold the full list while a filter narrows it (the leaves
	// filter or the `/` query); items, formulae and casks are then the rows that
	// pass. formulae and casks are index-aligned with items. See filter.go.
	unfilteredFormulae []brew.Formula
	unfilteredCasks    []brew.Cask
	leavesActive       bool
	// leaves is the set of names the leaves filter keeps.
	leaves map[string]bool
	// query is the `/` filter's text: a case-insensitive substring of the name.
	query    string
	selected int
	offset   int
	height   int
	loading  bool
	err      error
	formulae []brew.Formula
	casks    []brew.Cask
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
	// trusted is the Taps list's trusted formulae, casks and commands, in
	// trustedRows order. tapRows is the Taps list in display order, built from
	// taps and trusted by setTapRows; see trusted.go.
	trusted []brew.TrustEntry
	tapRows []tapRow
	// trustFailed is set when the last trust read failed: the failure is toasted
	// once, not again on every refresh that repeats it.
	trustFailed bool
}

// itemCount is the count the sidebar border shows. On the Taps list it is the
// taps, not the rows under them.
func (p *panelData) itemCount() int {
	if p.id == PanelTaps {
		return len(p.taps)
	}
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
		if r, ok := p.tapRowAt(i); ok {
			if r.tap >= 0 {
				return p.taps[r.tap].Name
			}
			return p.trusted[r.item].Name
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

// writeNameAt is the name brew acts on for the row at i. A formula keeps its
// full name (user/tap/tool), since the short name would resolve to the
// homebrew/core formula of the same name. Casks write by token, so they keep
// nameAt.
func (p *panelData) writeNameAt(i int) string {
	if p.id == PanelFormulae || p.id == PanelOutdated {
		if i >= 0 && i < len(p.formulae) && p.formulae[i].FullName != "" {
			return p.formulae[i].FullName
		}
	}
	return p.nameAt(i)
}

// selectedName is the real name of the selected row; see nameAt.
func (p *panelData) selectedName() string {
	return p.nameAt(p.selected)
}

// selectedTrusted is the trusted item under the cursor on the Taps list, nil on a
// tap row.
func (p *panelData) selectedTrusted() *brew.TrustEntry {
	if r, ok := p.tapRowAt(p.selected); ok && r.item >= 0 {
		return &p.trusted[r.item]
	}
	return nil
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
	if r, ok := p.tapRowAt(p.selected); ok && r.tap >= 0 {
		return &p.taps[r.tap]
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

// selectPackage moves the cursor of a Formulae, Casks or Outdated list to the
// package called name, a cask row when cask is set (Outdated lists formulae,
// then casks, and both can share a name). An unknown name leaves the cursor.
func (p *panelData) selectPackage(name string, cask bool) {
	if name == "" {
		return
	}
	for i := range p.items {
		if p.nameAt(i) == name && (i >= len(p.formulae)) == cask {
			p.selected = i
			return
		}
	}
}

// selectNamed moves the cursor of p to the row of rows called name, when the
// list still has it. An empty or unknown name leaves the cursor where it is:
// after an untap, that is the tap that followed the removed one (v2-10
// AC-04.1).
func selectNamed[T any](p *panelData, rows []T, nameOf func(T) string, name string) {
	if name == "" {
		return
	}
	for i, row := range rows {
		if nameOf(row) == name {
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

// renderSidebarContent is the body of a sidebar box. A panel that was never
// loaded and an empty one draw no body: their state is the box's border label
// (sidebarLabel). A pending fetch does the same, unless the panel was loaded
// before: its rows stay until the fresh ones land. Search is not kept, since
// its rows belong to the last query.
func (p *panelData) renderSidebarContent(width, maxRows int, batch map[int]bool, ic style.Icons, focused bool) string {
	if p.loading && (!p.loaded || p.id == PanelSearch) {
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
		if p.query != "" {
			msg := fmt.Sprintf(`No matches for "%s"`, p.query)
			return style.SubtleText.Render(ansi.Truncate(msg, width, "..."))
		}
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

// panelCount is how many panels the sidebar lists: every PanelID, in order.
const panelCount = int(PanelSearch) + 1

func initPanels() []*panelData {
	panels := make([]*panelData, panelCount)
	for i := range panels {
		// M12: every panel starts with loading=false. Init only fetches
		// the active panel (Status by default); everything else lazy-
		// loads on first switchPanel via lazyLoadPanel(). RefreshMsg
		// sets loading=true on the panels it re-fetches.
		id := PanelID(i)
		panels[i] = &panelData{id: id, title: id.Title()}
	}
	return panels
}
