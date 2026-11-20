package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
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
	got := panel("2 Formulae", "125", "row one\nrow two", 24, 4, false)
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
	title := panel("a title far too long for the box", "", "", 20, 3, true)
	if top := plainLine(strings.Split(title, "\n")[0]); lipgloss.Width(top) != 20 || !strings.Contains(top, "…") {
		t.Fatalf("long title should be cut with an ellipsis at 20 cells, got %q", top)
	}

	body := panel("t", "", "this body line is far too long for the box", 20, 3, false)
	rows := strings.Split(body, "\n")
	if len(rows) != 3 {
		t.Fatalf("a long body line must not add rows: %d rows, want 3", len(rows))
	}
	if got := plainLine(rows[1]); !strings.Contains(got, "…") {
		t.Fatalf("long body line should end in an ellipsis, got %q", got)
	}

	// Body rows past the height are dropped, not drawn past the bottom border.
	many := panel("t", "", "a\nb\nc\nd\ne", 20, 4, false)
	if n := len(strings.Split(many, "\n")); n != 4 {
		t.Fatalf("five body lines in a 4-row box drew %d rows, want 4", n)
	}
}

// TestPanelFocusChangesBorder: focus is a visible difference, not only a
// title change.
func TestPanelFocusChangesBorder(t *testing.T) {
	if panel("t", "", "x", 12, 3, true) == panel("t", "", "x", 12, 3, false) {
		t.Fatal("a focused panel should draw differently from an unfocused one")
	}
}

// TestSidebarLabelStates: the border label says what the panel's body cannot
// say when the box is collapsed.
func TestSidebarLabelStates(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	casks := m.panels[PanelCasks]

	if got := plainLine(m.sidebarLabel(casks)); got != "not loaded" {
		t.Fatalf("never-loaded label = %q, want not loaded", got)
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Items: []string{}})
	if got := plainLine(m.sidebarLabel(m.panels[PanelCasks])); got != "(empty)" {
		t.Fatalf("loaded-empty label = %q, want (empty)", got)
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Items: []string{"a", "b"}})
	if got := plainLine(m.sidebarLabel(m.panels[PanelCasks])); got != "2" {
		t.Fatalf("loaded label = %q, want the count 2", got)
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Err: assertAnError})
	if got := plainLine(m.sidebarLabel(m.panels[PanelCasks])); got != "!" {
		t.Fatalf("failed label = %q, want !", got)
	}
	m.panels[PanelCasks].loading = true
	if got := m.sidebarLabel(m.panels[PanelCasks]); got != m.spinner.View() {
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

	formulaeTop := heights[0] + 2
	nm, _ := m.Update(tea.MouseClickMsg{X: 3, Y: formulaeTop, Button: tea.MouseLeft})
	if got := nm.(Model).activePanel; got != PanelFormulae {
		t.Fatalf("click on Formulae's top border activated %v, want Formulae", got)
	}

	statusBottom := heights[0] + 1
	nm, _ = m.Update(tea.MouseClickMsg{X: 3, Y: statusBottom, Button: tea.MouseLeft})
	if got := nm.(Model).activePanel; got != PanelStatus {
		t.Fatalf("click on Status's bottom border activated %v, want Status", got)
	}
}
