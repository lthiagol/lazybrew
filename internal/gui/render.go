package gui

import (
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

func (m Model) renderSidebar() string {
	sw := sidebarWidth(m.cfg, m.width)
	heights := m.computeContentHeights()

	var boxes []string
	for i, p := range m.panels {
		active := i == int(m.activePanel)
		// The number is the jump key (1-7); the icon, when shown, is the
		// panel's own. The state or count sits in the border, so a box
		// too short for its body still shows it (Q-v2-small).
		title := borderTitle{
			full: strconv.Itoa(i+1) + " " + withIcon(m.panelIcon(p.id), p.title),
			bare: strconv.Itoa(i+1) + " " + p.title,
		}
		// visibleRows is derived in the Update loop (syncVisibleRows);
		// rendering must not write model state (r1 finding 29, r2
		// finding NEW-8). Fall back to the layout computation for models
		// that have not seen a WindowSizeMsg yet, without assigning back.
		rows := p.visibleRows
		if rows <= 0 {
			rows = heights[i]
		}
		// The batch is a set of Outdated rows. Only that list draws its marks.
		var batch map[int]bool
		if p.id == PanelOutdated {
			batch = m.batch.selected
		}
		body := p.renderSidebarContent(sw-2, rows, batch, m.icons, active)
		boxes = append(boxes, panel(title, m.sidebarLabel(p), body, sw, heights[i]+2, active))
	}

	return lipgloss.JoinVertical(lipgloss.Top, boxes...)
}

// sidebarLabel is the right-hand label of a sidebar box: the state while a
// fetch is pending or has failed, otherwise the item count (0 when empty).
// The Status box shows no count: its items are dashboard lines, not things
// that can be counted.
func (m Model) sidebarLabel(p *panelData) borderLabel {
	switch {
	case p.loading:
		return borderLabel{full: m.spinner.View(), short: "…"}
	case p.err != nil:
		badge := style.ErrorBadge.Render("!")
		return borderLabel{full: badge, short: badge}
	case !p.loaded:
		return borderLabel{
			full:  style.SubtleText.Render("not loaded"),
			short: lipgloss.NewStyle().Foreground(style.MutedColor).Render("–"),
		}
	case p.id == PanelStatus:
		return borderLabel{}
	case p.itemCount() == 0:
		return borderLabel{full: style.SubtleText.Render("0"), count: true}
	default:
		return borderLabel{full: style.SubtleText.Render(strconv.Itoa(p.itemCount())), count: true}
	}
}

// renderMainPanel draws the main column: the main panel box with the Log
// box stacked below it at the same width. The column is exactly
// l.mainW wide and l.bodyH rows tall.
func (m Model) renderMainPanel() string {
	l := m.screenLayout()
	innerW := l.mainW - 2

	// The title is the breadcrumb: the panel, then the tab.
	title := m.panels[m.activePanel].title
	if len(m.tabs) > 0 && m.activeTab < len(m.tabs) {
		title += " › " + m.tabs[m.activeTab].name
	}

	// The tab bar is cut to the box, not wrapped: a tab that does not fit is
	// cut, and the bar's trailing padding goes with it.
	tabBar := fitFrame(m.renderTabBar(innerW), innerW, m.tabBarRows())
	content := m.scrolledContent(m.renderContent(innerW, l.contentH), innerW, l.contentH)
	body := lipgloss.JoinVertical(lipgloss.Top, tabBar, content)

	box := panel(titled(title), borderLabel{}, body, l.mainW, l.mainH, true)
	if l.logH == 0 {
		return box
	}
	return lipgloss.JoinVertical(lipgloss.Top, box, m.renderCommandLogBox(l.mainW, l.logH))
}

// bottomBarRows is the height of the key-hint bar pinned under the body.
const bottomBarRows = 1

// screenLayout is the one geometry every frame is drawn to. Render, the
// C toggle and mouse hit-testing all read it, so a click lands on the row
// the frame drew. The body (sidebar and main column) fills every row above
// the bottom bar, and the main panel fills every column beside the sidebar.
type screenLayout struct {
	bodyH    int // rows above the bottom bar
	sideW    int // sidebar outer width
	mainW    int // main panel outer width; the Log box is the same width
	mainH    int // main panel outer height, borders included
	logH     int // Log box outer height; 0 when collapsed or too small to show
	contentH int // main content rows, below the tab bar
}

func (m Model) screenLayout() screenLayout {
	bodyH := max(0, m.height-headerRows-bottomBarRows)
	sideW := sidebarWidth(m.cfg, m.width)
	logH := bodyH / 5
	// The Log box needs a border top and bottom plus one body row; below
	// three rows it cannot show a body, so it stays hidden.
	if m.logCollapsed || logH < 3 {
		logH = 0
	}
	mainH := bodyH - logH
	return screenLayout{
		bodyH:    bodyH,
		sideW:    sideW,
		mainW:    m.width - sideW,
		mainH:    mainH,
		logH:     logH,
		contentH: max(0, mainH-2-m.tabBarRows()),
	}
}

// tabBarRows is the height of the tab bar. The active tab draws a thick
// underline, so a tabbed panel uses two rows. A panel without tabs still
// keeps one blank row, so its content starts on the same row as the others.
func (m Model) tabBarRows() int {
	if len(m.tabs) == 0 {
		return 1
	}
	return lipgloss.Height(m.renderTabBar(0))
}

// fitBox sizes s to exactly w cells by h rows: long lines wrap, short ones
// pad, and rows past h are dropped. Every box is fitted before its border
// is added, so an overlong child cannot cut the border off.
func fitBox(s string, w, h int) string {
	return lipgloss.NewStyle().Width(w).Height(h).MaxHeight(h).Render(s)
}

// fitFrame sizes a whole frame to exactly w cells by h rows. Unlike fitBox
// it cuts lines past w instead of wrapping them, so no row can add another
// row to the frame.
func fitFrame(s string, w, h int) string {
	return fitBox(lipgloss.NewStyle().MaxWidth(w).Render(s), w, h)
}

// compose draws overlays on top of the base frame and returns the result at
// exactly w cells by h rows. Each overlay is a layer with a higher z than
// the one before it, so the order of the arguments is the stacking order.
// Cells no overlay covers are the base frame's own.
func compose(base string, w, h int, overlays ...*lipgloss.Layer) string {
	layers := []*lipgloss.Layer{lipgloss.NewLayer(base).Z(0)}
	for i, o := range overlays {
		layers = append(layers, o.Z(i+1))
	}
	canvas := lipgloss.NewCanvas(w, h)
	canvas.Compose(lipgloss.NewCompositor(layers...))
	return fitFrame(canvas.Render(), w, h)
}

// dimBackdrop mutes the base frame behind an open modal: every row but the
// last (the footer, which shows the modal's keys) is drawn in the muted role.
// Only the style changes; the text of every cell stays as it was.
func dimBackdrop(base string) string {
	rows := strings.Split(base, "\n")
	muted := lipgloss.NewStyle().Foreground(style.MutedColor)
	for i := 0; i < len(rows)-1; i++ {
		rows[i] = muted.Render(ansi.Strip(rows[i]))
	}
	return strings.Join(rows, "\n")
}

// bottomRightOrigin is where content goes to sit flush right with its last
// row on the last body row, just above the bottom bar. A toast stack sits
// there, so it never moves the layout.
func bottomRightOrigin(content string, w, h int) (x, y int) {
	x = max(0, w-lipgloss.Width(content))
	y = max(0, h-bottomBarRows-lipgloss.Height(content))
	return x, y
}

// centred places content in the middle of a w x h frame.
func centred(content string, w, h int) *lipgloss.Layer {
	x := max(0, (w-lipgloss.Width(content))/2)
	y := max(0, (h-lipgloss.Height(content))/2)
	return lipgloss.NewLayer(content).X(x).Y(y)
}

// logBusy reports whether the log pane should use the Accent border.
func (m Model) logBusy() bool {
	return m.tasks.IsRunning() || m.opState != nil
}

// renderCommandLogBox draws the Log box. Its title is fixed; while a task
// runs, the label names the task.
func (m Model) renderCommandLogBox(width, height int) string {
	// height is the full vertical budget for the log section, border
	// included, so the body is two rows shorter than the box.
	bodyH := max(0, height-2)
	bodyW := max(1, width-2)
	body := m.commandLog.View(bodyW, bodyH)
	label := ""
	if m.logBusy() && m.opState != nil {
		label = style.SubtleText.Render(m.opState.Title)
	}
	return panel(titled("Log"), borderLabel{full: label}, body, width, height, m.logBusy())
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
	sidebarHeight := m.screenLayout().bodyH
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
	inactiveCount := n - 1
	inactiveRows := remaining / max(1, inactiveCount)
	if inactiveRows < minInactive {
		inactiveRows = minInactive
	}
	extra := availableRows - activeRows - inactiveRows*inactiveCount
	heights := make([]int, n)
	for i := range heights {
		if i == int(m.activePanel) {
			heights[i] = activeRows
		} else {
			heights[i] = inactiveRows
		}
		if extra > 0 {
			heights[i]++
			extra--
		}
	}
	return heights
}

func (m Model) renderContent(width, height int) string {
	if m.opState != nil {
		return m.renderOperation(width, height)
	}

	panel := m.panels[m.activePanel]
	// PanelSearch has its own " Searching..." path below; skip the generic
	// Loading... early-return so that label is reachable (M15 AC-03).
	if panel.loading && m.activePanel != PanelSearch {
		return m.spinner.View() + style.SubtleText.Render(" Loading...")
	}

	switch m.activePanel {
	case PanelFormulae:
		switch m.activeTab {
		case formulaInfoTab:
			f := panel.selectedFormula()
			if f == nil {
				return emptyPanel(width, height)
			}
			return presentation.FormatFormulaInfo(*f, width, m.icons)
		case formulaDepsTab, formulaUsesTab, formulaFilesTab:
			itemName := selectedItemName(panel)
			if itemName == "" {
				// No item to fetch for: without this explicit state the
				// tab showed a spinner forever because no fetch could be
				// scheduled (r1 finding 28).
				return style.SubtleText.Render("No package selected")
			}
			key := tabKey(m.activePanel, m.activeTab, itemName)
			if content, ok := m.tabContent[key]; ok {
				switch m.activeTab {
				case formulaUsesTab:
					return m.renderUses(content)
				case formulaFilesTab:
					return renderFiles(content)
				}
				return content
			}
			return m.spinner.View() + style.SubtleText.Render(" Loading...")
		case formulaCaveatsTab:
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
			return presentation.FormatCaskInfo(*c, width, m.icons)
		case 1:
			c := panel.selectedCask()
			if c == nil {
				return style.SubtleText.Render(noDependencies)
			}
			view := m.renderDepGroups([]depGroup{
				{heading: "Formulae", names: c.DependsOnFormulae, check: depFormula},
				{heading: "Casks", names: c.DependsOnCasks, check: depCask},
				{heading: "Requirements", names: c.Requirements, check: depNone},
			})
			if view == "" {
				return style.SubtleText.Render(noDependencies)
			}
			return view
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
			return result
		}

	case PanelStatus:
		switch m.activeTab {
		case statusDashboardTab:
			if !panel.loaded {
				return style.SubtleText.Render("Status dashboard not loaded yet — press R to refresh")
			}
			return m.renderDashboard(width)
		case statusConfigTab, statusDoctorTab:
			// Config/Doctor fetch from the loaded dashboard; before the
			// dashboard has loaded there is nothing to fetch, and the old
			// code left a permanent spinner (r1 finding 28).
			itemName := tabItemName(panel)
			if itemName == "" {
				return style.SubtleText.Render("Status dashboard not loaded yet — press R to refresh")
			}
			key := tabKey(m.activePanel, m.activeTab, itemName)
			if content, ok := m.tabContent[key]; ok {
				if m.activeTab == statusConfigTab {
					// The config rows are built uncut (formatConfig), so they
					// are cut to the main width here, with …
					return cutLines(content, width)
				}
				return content
			}
			return m.spinner.View() + style.SubtleText.Render(" Loading...")
		}

	case PanelTaps:
		switch m.activeTab {
		case tapInfoTab:
			return renderTapInfo(panel, width)
		case tapTrustTab:
			t := panel.selectedTap()
			if t == nil {
				return style.SubtleText.Render("No tap selected")
			}
			return tapTrustStory(t, width)
		case tapFormulaeTab:
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
		return renderServiceStatus(panel, width, m.icons)

	case PanelOutdated:
		switch m.activeTab {
		case outdatedInfoTab:
			// The outdated panel mixes formulae and casks in one item
			// list; cask rows are indexed past the formulae, so the cask
			// lookup must subtract len(formulae) — indexing selectedCask()
			// with the raw combined index showed the WRONG cask (r1
			// finding 11).
			if f := panel.selectedOutdatedFormula(); f != nil {
				return presentation.FormatFormulaInfo(*f, width, m.icons)
			}
			if c := panel.selectedOutdatedCask(); c != nil {
				return presentation.FormatCaskInfo(*c, width, m.icons)
			}
			return emptyPanel(width, height)
		case outdatedVersionsTab:
			return renderOutdatedVersions(panel, m.batch.selected, width)
		}

	case PanelSearch:
		searchBox := style.AccentText.Render("/ ") + m.searchInput.View()
		searchBox = style.NormalItem.Render(searchBox)
		help := style.SubtleText.Render("  Enter to search  |  Esc to clear")
		top := lipgloss.JoinVertical(lipgloss.Top, searchBox, help, "")

		if len(m.searchResults) == 0 && m.searchInfoContent == "" {
			if !m.panels[PanelSearch].loading {
				return top + "\n" + style.SubtleText.Render("Search for Homebrew packages...")
			}
			return top + "\n" + m.spinner.View() + style.SubtleText.Render(" Searching...")
		}
		if m.searchInfoContent != "" {
			return top + "\n" + m.renderSearchInfo(width, height-3)
		}
		return top + "\n" + style.SubtleText.Render("Select a result to see details")
	}

	return emptyPanel(width, height)
}

func (m Model) renderSearchInfo(width, height int) string {
	info, err := parsePackageInfo(m.searchInfoContent)
	if err != nil {
		return style.ErrorBadge.Render("Parse error: " + err.Error())
	}

	ver := info.Version
	if info.Bottled {
		ver += " (bottled)"
	}
	status, role := searchStatus(info, m.icons)
	result := presentation.Detail([]presentation.Field{
		{Label: "Name", Value: info.Name},
		{Label: "Version", Value: ver},
		{Label: "Type", Value: info.Type},
		{Label: "Status", Value: status, Role: role},
		{Label: "License", Value: info.License},
		{Label: "Homepage", Value: info.Homepage, Role: style.InfoColor},
		{Label: "Description", Value: info.Description, Prose: true},
	}, width)

	if view := m.renderDepGroups(info.Groups); view != "" {
		result += "\n\n" + style.AccentText.Render("Dependencies:") + "\n" + view
	}
	if info.Caveats != "" {
		result += "\n\n" + presentation.Detail([]presentation.Field{
			{Label: "Caveats", Value: info.Caveats, Prose: true},
		}, width)
	}
	return result
}

// searchStatus is the Status row of a search result: installed with its
// install path, or not installed.
func searchStatus(info *pkgInfo, icons style.Icons) (string, color.Color) {
	if !info.Installed {
		return "not installed", style.MutedColor
	}
	status := "installed"
	if info.InstallPath != "" {
		status += " (" + info.InstallPath + ")"
	}
	return withIcon(icons.Installed, status), style.SuccessColor
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
