package gui

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// TestViewOwnsItsCanvas pins that View paints the theme's base and text as
// the terminal background and foreground, for both flavours.
func TestViewOwnsItsCanvas(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	for name, theme := range map[string]*style.Theme{
		"dark":  style.DarkTheme(),
		"light": style.LightTheme(),
	} {
		style.ApplyTheme(theme)
		m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 120, Height: 40})
		v := m.View()
		if v.BackgroundColor != theme.Background {
			t.Errorf("%s: BackgroundColor = %v, want the theme base %v", name, v.BackgroundColor, theme.Background)
		}
		if v.ForegroundColor != theme.Text {
			t.Errorf("%s: ForegroundColor = %v, want the theme text %v", name, v.ForegroundColor, theme.Text)
		}
	}
}

// plainLine is one rendered row with its styles removed.
func plainLine(row string) string {
	return string(plainRows(row)[0])
}

// TestPanelTitleAndLabelSitInTheBorder pins the panel primitive's layout: the
// title is inside the top border at the left, the label at the right, and
// the box is exactly w by h.
func TestPanelTitleAndLabelSitInTheBorder(t *testing.T) {
	got := panel(titled("2 Formulae"), borderLabel{full: "125", count: true}, "row one\nrow two", 24, 4, false)
	rows := strings.Split(got, "\n")
	if len(rows) != 4 {
		t.Fatalf("panel has %d rows, want 4", len(rows))
	}
	for i, row := range rows {
		if w := lipgloss.Width(row); w != 24 {
			t.Fatalf("row %d is %d cells wide, want 24: %q", i, w, plainLine(row))
		}
	}
	top := plainLine(rows[0])
	if !strings.HasPrefix(top, "╭─ 2 Formulae ") || !strings.HasSuffix(top, " 125 ─╮") {
		t.Fatalf("top border should read ╭─ title ─── label ─╮, got %q", top)
	}
	if got, want := plainLine(rows[1]), "│row one"+strings.Repeat(" ", 15)+"│"; got != want {
		t.Fatalf("first body row = %q, want %q", got, want)
	}
	if got := plainLine(rows[3]); !strings.HasPrefix(got, "╰") || !strings.HasSuffix(got, "╯") {
		t.Fatalf("bottom border = %q", got)
	}
}

// TestPanelCutsLinesNeverWraps: a long title or body line is cut with an
// ellipsis, and the box keeps its height.
func TestPanelCutsLinesNeverWraps(t *testing.T) {
	title := panel(titled("a title far too long for the box"), borderLabel{}, "", 20, 3, true)
	if top := plainLine(strings.Split(title, "\n")[0]); lipgloss.Width(top) != 20 || !strings.Contains(top, "…") {
		t.Fatalf("long title should be cut with an ellipsis at 20 cells, got %q", top)
	}

	body := panel(titled("t"), borderLabel{}, "this body line is far too long for the box", 20, 3, false)
	rows := strings.Split(body, "\n")
	if len(rows) != 3 {
		t.Fatalf("a long body line must not add rows: %d rows, want 3", len(rows))
	}
	if got := plainLine(rows[1]); !strings.Contains(got, "…") {
		t.Fatalf("long body line should end in an ellipsis, got %q", got)
	}

	// Body rows past the height are dropped, not drawn past the bottom border.
	many := panel(titled("t"), borderLabel{}, "a\nb\nc\nd\ne", 20, 4, false)
	if n := len(strings.Split(many, "\n")); n != 4 {
		t.Fatalf("five body lines in a 4-row box drew %d rows, want 4", n)
	}
}

// TestPanelFocusChangesBorder: focus is a visible difference, not only a
// title change.
func TestPanelFocusChangesBorder(t *testing.T) {
	if panel(titled("t"), borderLabel{}, "x", 12, 3, true) == panel(titled("t"), borderLabel{}, "x", 12, 3, false) {
		t.Fatal("a focused panel should draw differently from an unfocused one")
	}
}

// TestSidebarLabelStates: the border label says what the panel's body cannot
// say when the box is collapsed.
func TestSidebarLabelStates(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	casks := m.panels[PanelCasks]

	if got := plainLine(m.sidebarLabel(casks).full); got != "not loaded" {
		t.Fatalf("never-loaded label = %q, want not loaded", got)
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Items: []string{}})
	if got := plainLine(m.sidebarLabel(m.panels[PanelCasks]).full); got != "0" {
		t.Fatalf("loaded-empty label = %q, want 0", got)
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Items: []string{"a", "b"}})
	if got := plainLine(m.sidebarLabel(m.panels[PanelCasks]).full); got != "2" {
		t.Fatalf("loaded label = %q, want the count 2", got)
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Err: assertAnError})
	if got := plainLine(m.sidebarLabel(m.panels[PanelCasks]).full); got != "!" {
		t.Fatalf("failed label = %q, want !", got)
	}
	m.panels[PanelCasks].loading = true
	if got := m.sidebarLabel(m.panels[PanelCasks]).full; got != m.spinner.View() {
		t.Fatalf("loading label = %q, want the spinner", got)
	}
}

// TestSidebarTitlesCarryIconOnlyWhenShown: the icon sits between the number
// and the name, and show_icons: false leaves no glyph and no double space.
func TestSidebarTitlesCarryIconOnlyWhenShown(t *testing.T) {
	// The title is on the top border, the first row of the sidebar.
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	if top := plainLine(strings.Split(m.renderSidebar(), "\n")[0]); !strings.HasPrefix(top, "╭─ 1 ◉ Status ") {
		t.Fatalf("sidebar title should carry the Unicode icon, got %q", top)
	}

	cfg := config.Default()
	cfg.GUI.ShowIcons = false
	off := updateModel(New(brew.NewClient(brew.NewMockRunner()), cfg), tea.WindowSizeMsg{Width: 140, Height: 40})
	if top := plainLine(strings.Split(off.renderSidebar(), "\n")[0]); !strings.HasPrefix(top, "╭─ 1 Status ") {
		t.Fatalf("show_icons off: title should be \"1 Status\" with no stray space, got %q", top)
	}
}

// TestMainBorderTitleIsBreadcrumb: the main panel's title is the breadcrumb.
func TestMainBorderTitleIsBreadcrumb(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "2")
	m = sendKey(m, "]")
	if top := plainLine(strings.Split(m.renderMainPanel(), "\n")[0]); !strings.Contains(top, "Formulae › Info") {
		t.Fatalf("main title should be the breadcrumb Formulae › Info, got %q", top)
	}
}

// TestLogBorderLabelNamesRunningTask: the Log box is titled Log, and while a
// task runs its label names the task.
func TestLogBorderLabelNamesRunningTask(t *testing.T) {
	m := newTestModel()
	if top := plainLine(strings.Split(m.renderCommandLogBox(40, 6), "\n")[0]); !strings.HasPrefix(top, "╭─ Log ") || strings.Contains(top, "upgrading") {
		t.Fatalf("idle Log title = %q", top)
	}
	m.opState = &Operation{Title: "upgrading curl", Status: opRunning}
	if top := plainLine(strings.Split(m.renderCommandLogBox(40, 6), "\n")[0]); !strings.Contains(top, "Log") || !strings.Contains(top, "upgrading curl") {
		t.Fatalf("busy Log title should name the task, got %q", top)
	}
}

// TestSidebarBorderClicks: a click on a sidebar box's top border activates
// that panel, and a click on its bottom border changes nothing.
func TestSidebarBorderClicks(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m.cfg.GUI.Mouse = true
	heights := m.computeContentHeights()

	formulaeTop := headerRows + heights[0] + 2
	nm, _ := m.Update(tea.MouseClickMsg{X: 3, Y: formulaeTop, Button: tea.MouseLeft})
	if got := nm.(Model).activePanel; got != PanelFormulae {
		t.Fatalf("click on Formulae's top border activated %v, want Formulae", got)
	}

	statusBottom := headerRows + heights[0] + 1
	nm, _ = m.Update(tea.MouseClickMsg{X: 3, Y: statusBottom, Button: tea.MouseLeft})
	if got := nm.(Model).activePanel; got != PanelStatus {
		t.Fatalf("click on Status's bottom border activated %v, want Status", got)
	}
}

// TestHeaderIsOneFullWidthRow: row 0 is the header, exactly the frame's width,
// with the name at the left and nothing shown for values not yet known.
func TestHeaderIsOneFullWidthRow(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	head := m.renderHeader()
	if lipgloss.Height(head) != 1 || lipgloss.Width(head) != 140 {
		t.Fatalf("header is %d rows of %d cells, want 1 row of 140", lipgloss.Height(head), lipgloss.Width(head))
	}
	plain := plainLine(head)
	if !strings.HasPrefix(plain, "◆ lazybrew") {
		t.Fatalf("header should open with the app icon and name, got %q", plain)
	}
	for _, unknown := range []string{"formulae", "casks", "outdated", "Homebrew", "updated", "never updated", "0 "} {
		if strings.Contains(plain, unknown) {
			t.Errorf("unknown value shown in the header: %q in %q", unknown, plain)
		}
	}
	if top := strings.Split(m.View().Content, "\n")[0]; !strings.Contains(top, "lazybrew") {
		t.Fatalf("the frame's first row should be the header, got %q", top)
	}
}

// TestHeaderShowsKnownCountsAndBadges: counts appear once loaded, outdated
// only when there are some, and the version when the dashboard has it.
func TestHeaderShowsKnownCountsAndBadges(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelStatus, Items: []string{"Homebrew 4.5.2"}, BrewVersion: "4.5.2"})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Formulae: []brew.Formula{{Name: "a"}, {Name: "b"}}, Items: []string{"a", "b"}})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Casks: []brew.Cask{{Name: "c"}}, Items: []string{"c"}})

	plain := plainLine(m.renderHeader())
	for _, want := range []string{"Homebrew 4.5.2", "2 formulae", "1 casks"} {
		if !strings.Contains(plain, want) {
			t.Errorf("header missing %q: %q", want, plain)
		}
	}
	if strings.Contains(plain, "outdated") {
		t.Fatalf("no outdated badge without outdated packages: %q", plain)
	}

	m = updateModel(m, DataLoadedMsg{PanelID: PanelOutdated, Items: []string{"x 1 -> 2", "y 1 -> 2"}})
	if plain := plainLine(m.renderHeader()); !strings.Contains(plain, "2 outdated") {
		t.Fatalf("header should show the outdated count once known: %q", plain)
	}
	if m.brewVersion != "4.5.2" {
		t.Fatalf("the dashboard's version should reach the model, got %q", m.brewVersion)
	}
}

// TestHeaderUpdateBadge: the update status moved here from the bottom bar.
func TestHeaderUpdateBadge(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m.cfg.Brew.UpdateOnStart = true
	if plain := plainLine(m.renderHeader()); !strings.Contains(plain, "never updated") {
		t.Fatalf("update_on_start with no update yet should say never updated: %q", plain)
	}
	m.lastUpdate = time.Now().Add(-150 * time.Second)
	if plain := plainLine(m.renderHeader()); !strings.Contains(plain, "updated 2m ago") {
		t.Fatalf("header should say updated 2m ago: %q", plain)
	}
	m.isUpdating = true
	if plain := plainLine(m.renderHeader()); !strings.Contains(plain, "updating…") {
		t.Fatalf("header should say updating while brew update runs: %q", plain)
	}
	if plain := plainLine(m.renderBottomBar()); strings.Contains(plain, "updat") {
		t.Fatalf("the bottom bar should no longer carry the update status: %q", plain)
	}
}

// TestAgoTextUnits pins the one-unit form the update badge uses.
func TestAgoTextUnits(t *testing.T) {
	for d, want := range map[time.Duration]string{
		5 * time.Second:              "5s ago",
		150 * time.Second:            "2m ago",
		3*time.Hour + 20*time.Minute: "3h ago",
	} {
		if got := agoText(d); got != want {
			t.Errorf("agoText(%v) = %q, want %q", d, got, want)
		}
	}
}

// TestHeaderOperationInTheMiddle: a running operation sits between the name
// and the badges, and an idle header has nothing there.
func TestHeaderOperationInTheMiddle(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	if plain := plainLine(m.renderHeader()); strings.Contains(plain, "⟳") {
		t.Fatalf("idle header should show no operation: %q", plain)
	}
	m.opState = &Operation{Title: "upgrading curl", Status: opRunning}
	if plain := plainLine(m.renderHeader()); !strings.Contains(plain, "⟳ upgrading curl") {
		t.Fatalf("running operation should show in the header: %q", plain)
	}
	if plain := plainLine(m.renderHeader()); lipgloss.Width(plain) != 140 {
		t.Fatalf("header with an operation is %d cells, want 140", lipgloss.Width(plain))
	}
}

// TestHeaderDropsBadgesNotTheName: at the narrowest frame the badges drop
// from the left and the app name stays, and the header keeps its width.
func TestHeaderDropsBadgesNotTheName(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 80, Height: 24})
	m.cfg.Brew.UpdateOnStart = true
	m = updateModel(m, DataLoadedMsg{PanelID: PanelStatus, BrewVersion: "4.5.2-123-gabcdef"})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Formulae: make([]brew.Formula, 12345), Items: make([]string, 12345)})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Casks: make([]brew.Cask, 16), Items: make([]string, 16)})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelOutdated, Items: []string{"a", "b", "c"}})
	m.lastUpdate = time.Now().Add(-5 * time.Minute)

	// The name, the version and the three rightmost badges need 75 cells;
	// the formulae badge is the one that does not fit.
	head := m.renderHeader()
	plain := plainLine(head)
	if !strings.HasPrefix(plain, "◆ lazybrew") {
		t.Fatalf("the name must stay at 80 columns, got %q", plain)
	}
	if lipgloss.Width(head) != 80 {
		t.Fatalf("header is %d cells, want 80", lipgloss.Width(head))
	}
	if strings.Contains(plain, "formulae") {
		t.Fatalf("the leftmost badge should drop first at 80 columns: %q", plain)
	}
	if !strings.Contains(plain, "16 casks") {
		t.Fatalf("the badges after the dropped one should stay: %q", plain)
	}
	if !strings.Contains(plain, "updated 5m ago") {
		t.Fatalf("the rightmost badge should survive at 80 columns: %q", plain)
	}
}

// TestFooterIsOneRowAtEveryWidth: the footer is one row, exactly the frame's
// width, at 80 and 140 columns for both Formulae and Status.
func TestFooterIsOneRowAtEveryWidth(t *testing.T) {
	for _, w := range []int{80, 140} {
		for _, key := range []string{"2", "1"} {
			m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: 40})
			m = sendKey(m, key)
			bar := m.renderBottomBar()
			if lipgloss.Height(bar) != 1 || lipgloss.Width(bar) != w {
				t.Fatalf("%d cols, panel %s: footer is %d rows of %d cells, want 1 of %d",
					w, key, lipgloss.Height(bar), lipgloss.Width(bar), w)
			}
		}
	}
}

// TestFooterLeadsWithThePanelsKeys: the active panel's keys come first, then
// the globals, and the row says "? more" when it has cut some.
func TestFooterLeadsWithThePanelsKeys(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 80, Height: 40})
	m = sendKey(m, "2")
	plain := strings.TrimRight(plainLine(m.renderBottomBar()), " ")
	if !strings.HasPrefix(plain, " j/k navigate") {
		t.Fatalf("Formulae footer should lead with its own keys, got %q", plain)
	}
	if !strings.HasSuffix(plain, "? more") {
		t.Fatalf("a cut Formulae footer at 80 columns should end with ? more, got %q", plain)
	}

	m = updateModel(newTestModel(), tea.WindowSizeMsg{Width: 200, Height: 40})
	plain = strings.TrimRight(plainLine(m.renderBottomBar()), " ")
	if strings.Contains(plain, "more") {
		t.Fatalf("a footer that fits must not say ? more, got %q", plain)
	}
	if !strings.HasPrefix(plain, " R refresh") {
		t.Fatalf("Status footer should lead with its own keys, got %q", plain)
	}
}

// TestFooterShowsTheOpenModalsKeys: while a modal is open the footer lists
// the keys that modal answers to.
func TestFooterShowsTheOpenModalsKeys(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m.activeModal = modal.NewConfirmModal("Uninstall bash", "Remove bash?")
	if plain := plainLine(m.renderBottomBar()); !strings.HasPrefix(plain, " y yes  n no  esc cancel") {
		t.Fatalf("confirm footer = %q, want the confirm keys", plain)
	}

	m.activeModal = modal.NewMenuModal("Trust", []string{"Trust tap", "Cancel"})
	if plain := plainLine(m.renderBottomBar()); !strings.Contains(plain, "enter choose") {
		t.Fatalf("menu footer = %q, want the menu keys", plain)
	}
}

// TestFooterShowsHelpsKeysWhileHelpIsOpen: the help screen's last row is the
// footer, carrying help's own close keys.
func TestFooterShowsHelpsKeysWhileHelpIsOpen(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "?")
	rows := plainRows(m.View().Content)
	if last := string(rows[len(rows)-1]); !strings.Contains(last, "?/esc close") {
		t.Fatalf("help's footer should list its close keys, got %q", last)
	}
}

// TestSearchInputDrawsFromRoles: the search input's prompt, text, placeholder
// and cursor come from the theme roles, on both flavours.
func TestSearchInputDrawsFromRoles(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	for name, theme := range map[string]*style.Theme{
		"dark":  style.DarkTheme(),
		"light": style.LightTheme(),
	} {
		style.ApplyTheme(theme)
		s := newTestModel().searchInput.Styles()
		if got := s.Focused.Prompt.GetForeground(); got != theme.Accent {
			t.Errorf("%s: prompt = %v, want the accent %v", name, got, theme.Accent)
		}
		if got := s.Focused.Text.GetForeground(); got != theme.Text {
			t.Errorf("%s: text = %v, want the text role %v", name, got, theme.Text)
		}
		if got := s.Focused.Placeholder.GetForeground(); got != theme.Muted {
			t.Errorf("%s: placeholder = %v, want the muted role %v", name, got, theme.Muted)
		}
		if s.Cursor.Color != theme.Accent {
			t.Errorf("%s: cursor = %v, want the accent %v", name, s.Cursor.Color, theme.Accent)
		}
	}
}

// TestOperationAndLogGlyphsFollowIcons: the operation title and the Log's
// status marks come from the icon set, and show_icons: false draws none.
func TestOperationAndLogGlyphsFollowIcons(t *testing.T) {
	cases := []struct {
		name  string
		icons string
		show  bool
		want  style.Icons
	}{
		{"nerd", config.IconsNerd, true, style.NerdIcons},
		{"unicode", config.IconsUnicode, true, style.UnicodeIcons},
		{"off", config.IconsNerd, false, style.Icons{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(config.IconsEnv, "")
			cfg := config.Default()
			cfg.GUI.Icons = tc.icons
			cfg.GUI.ShowIcons = tc.show
			m := New(brew.NewClient(brew.NewMockRunner()), cfg)

			m.opState = &Operation{Title: "upgrading curl", Status: opRunning}
			title := strings.Split(m.renderOperation(60, 10), "\n")[0]
			if got, want := strings.TrimRight(plainLine(title), " "), withIcon(tc.want.Refresh, "upgrading curl"); got != want {
				t.Errorf("operation title = %q, want %q", got, want)
			}

			id := m.commandLog.Append("install bash")
			m.commandLog.SetStatus(id, CommandSuccess)
			if got, want := plainLine(m.commandLog.View(40, 4)), withIcon(tc.want.OK, "brew install bash"); !strings.HasPrefix(got, want) {
				t.Errorf("log line = %q, want it to start %q", got, want)
			}
		})
	}
}

// TestHeaderKeepsAGapBeforeBadges sweeps every width from 40 to 140: the
// version and the first badge never run together (F-r1-04), the bar is exactly
// the frame's width, and the name always stays.
func TestHeaderKeepsAGapBeforeBadges(t *testing.T) {
	version := regexp.MustCompile(`4\.5\.2\d`)
	for w := 40; w <= 140; w++ {
		m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
		m.width = w
		m = updateModel(m, DataLoadedMsg{PanelID: PanelStatus, BrewVersion: "4.5.2"})
		m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Formulae: make([]brew.Formula, 125), Items: make([]string, 125)})
		m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Casks: make([]brew.Cask, 16), Items: make([]string, 16)})
		m = updateModel(m, DataLoadedMsg{PanelID: PanelOutdated, Items: []string{"a", "b"}})
		m.cfg.Brew.UpdateOnStart = true
		head := m.renderHeader()
		plain := plainLine(head)
		if lipgloss.Width(head) != w {
			t.Fatalf("width %d: header is %d cells", w, lipgloss.Width(head))
		}
		if !strings.HasPrefix(plain, "◆ lazybrew") {
			t.Fatalf("width %d: the name must stay, got %q", w, plain)
		}
		if version.MatchString(plain) {
			t.Fatalf("width %d: version and badge run together: %q", w, plain)
		}
	}
}

// TestHeaderOperationOutranksBadges (F-orch-03): at 80 columns with an
// operation running, the badges drop first, from the left, and the operation
// stays whole. Only when no badge is left is the operation cut with an ellipsis.
func TestHeaderOperationOutranksBadges(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 80, Height: 24})
	m.cfg.Brew.UpdateOnStart = true
	m = updateModel(m, DataLoadedMsg{PanelID: PanelStatus, BrewVersion: "4.5.2"})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Formulae: make([]brew.Formula, 125), Items: make([]string, 125)})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Casks: make([]brew.Cask, 16), Items: make([]string, 16)})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelOutdated, Items: []string{"a", "b"}})
	m.lastUpdate = time.Now().Add(-2 * time.Minute)
	m.opState = &Operation{Title: "upgrading curl", Status: opRunning}

	head := m.renderHeader()
	plain := plainLine(head)
	if lipgloss.Width(head) != 80 || !strings.HasPrefix(plain, "◆ lazybrew") {
		t.Fatalf("header must be 80 cells with the name first, got %d cells: %q", lipgloss.Width(head), plain)
	}
	if !strings.Contains(plain, "⟳ upgrading curl") {
		t.Fatalf("the operation should stay whole at 80 columns: %q", plain)
	}
	if strings.Contains(plain, "formulae") || strings.Contains(plain, "casks") {
		t.Fatalf("the leftmost badges should drop first: %q", plain)
	}
	if !strings.Contains(plain, "updated 2m ago") {
		t.Fatalf("the rightmost badge should survive while the operation fits: %q", plain)
	}

	// An operation too long for the bar takes every badge with it, and is cut.
	m.opState.Title = strings.Repeat("reinstalling ", 10)
	head = m.renderHeader()
	plain = plainLine(head)
	if lipgloss.Width(head) != 80 || !strings.HasPrefix(plain, "◆ lazybrew") {
		t.Fatalf("a long operation must keep the bar at 80 cells with the name: %q", plain)
	}
	if !strings.Contains(plain, "…") || strings.Contains(plain, "updated") {
		t.Fatalf("a long operation should be cut once the badges are gone: %q", plain)
	}
}

// TestStatusBoxLabelShowsNoCount (F-orch-01): the Status box's border label
// names no count once loaded; the loading, error and not-loaded states still
// show.
func TestStatusBoxLabelShowsNoCount(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	if got := plainLine(m.sidebarLabel(m.panels[PanelStatus]).full); got != "not loaded" {
		t.Fatalf("unloaded Status label = %q, want not loaded", got)
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelStatus, Items: []string{"Homebrew 4.5.2", "Prefix: /opt", "Formulae: 125 installed"}})
	if got := m.sidebarLabel(m.panels[PanelStatus]).full; got != "" {
		t.Fatalf("loaded Status label = %q, want no count", got)
	}
	if top := plainLine(strings.Split(m.renderSidebar(), "\n")[0]); strings.Contains(top, "3") {
		t.Fatalf("Status border must not show the dashboard's line count: %q", top)
	}
}

// TestSidebarLabelsAtEightyColumns (F-orch-02): at 80x24 the sidebar box is 24
// cells, so a label may not fit whole. Not loaded shows as the short dash, a
// count (0 included) is never cut, a label with no room is left out, and the loading
// spinner and the error mark stay.
func TestSidebarLabelsAtEightyColumns(t *testing.T) {
	// These labels are drawn for a 24-cell sidebar, which is 80 columns at the
	// old 30% width. The 35% default gives 28 cells, where every label fits whole.
	m := newTestModel()
	m.cfg.GUI.SidebarWidth = 30
	m = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelTaps, Items: []string{"a", "b", "c"}})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelServices, Items: []string{}})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelSearch, Err: assertAnError})
	m.panels[PanelOutdated].loading = true

	rows := strings.Split(m.renderSidebar(), "\n")
	top := func(title string) string {
		t.Helper()
		for _, r := range rows {
			if p := plainLine(r); strings.Contains(p, title) {
				return p
			}
		}
		t.Fatalf("no sidebar border shows %q", title)
		return ""
	}
	if got := top("3 ▣ Casks"); !strings.HasSuffix(got, " – ─╮") {
		t.Errorf("unloaded Casks at 80 columns should show the short dash, got %q", got)
	}
	if got := top("5 ⊕ Taps"); !strings.HasSuffix(got, " 3 ─╮") {
		t.Errorf("the Taps count must show in full, got %q", got)
	}
	if got := top("6 ◎ Services"); !strings.HasSuffix(got, " 0 ─╮") {
		t.Errorf("an empty panel shows 0 in full at 80 columns, got %q", got)
	}
	if got := top("4 ↑ Outdated"); !strings.HasSuffix(got, " "+plainLine(m.spinner.View())+" ─╮") {
		t.Errorf("the loading spinner should stay at 80 columns, got %q", got)
	}
	if got := top("7 ⌕ Search"); !strings.HasSuffix(got, " ! ─╮") {
		t.Errorf("the error mark should stay at 80 columns, got %q", got)
	}
}

// TestCountsStayWholeAt80Columns (F-r2-01): a count in a sidebar border always
// shows whole. At 80 columns with sidebar_width 30 the box is 24 cells, so the
// title yields first its icon and then its name; the row stays exactly the
// box's width.
func TestCountsStayWholeAt80Columns(t *testing.T) {
	for _, n := range []int{1234, 12345} {
		m := newTestModel()
		m.cfg.GUI.SidebarWidth = 30
		m = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})
		m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: make([]string, n)})
		m = updateModel(m, DataLoadedMsg{PanelID: PanelOutdated, Items: make([]string, n)})
		m = updateModel(m, DataLoadedMsg{PanelID: PanelServices, Items: make([]string, n)})
		rows := strings.Split(m.renderSidebar(), "\n")
		count := strconv.Itoa(n)
		for _, name := range []string{"Formulae", "Outdated", "Services"} {
			var row string
			for _, r := range rows {
				if strings.Contains(plainLine(r), name) {
					row = plainLine(r)
				}
			}
			if !strings.HasSuffix(row, " "+count+" ─╮") {
				t.Errorf("%d items: %s border must end with the whole count, got %q", n, name, row)
			}
			if w := lipgloss.Width(row); w != 24 {
				t.Errorf("%d items: %s border is %d cells, want 24: %q", n, name, w, row)
			}
		}
	}
}

// TestBorderTitleYieldsToCount: when even the bare name cannot fit beside a
// count, the name is cut and the count is kept whole.
func TestBorderTitleYieldsToCount(t *testing.T) {
	name, tail := borderHead(
		borderTitle{full: "2 ∘ Formulae", bare: "2 Formulae"},
		borderLabel{full: "1234567", count: true}, 24)
	if tail != " 1234567 " {
		t.Fatalf("the count must be whole, got tail %q", tail)
	}
	if !strings.HasSuffix(name, "…") || lipgloss.Width(name) > 24-7-lipgloss.Width(tail) {
		t.Fatalf("the name should be cut to make room for the count, got %q", name)
	}
}
