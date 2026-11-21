package gui

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// sgrColor is the SGR parameters lipgloss writes for a colour: 38;2;r;g;b
// for the foreground, 48;2;r;g;b for the background. A frame carries them
// only where the colour is drawn.
func sgrColor(kind string, c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("%s;2;%d;%d;%d", kind, r>>8, g>>8, b>>8)
}

// themeScreen is one screen a theme must draw at 80x24. setup readies the
// model; want names the colour tokens its frame must carry for the theme.
type themeScreen struct {
	name  string
	setup func(m *Model) *Model
	want  func(th *style.Theme) []string
}

var themeScreens = []themeScreen{
	{
		// The focused sidebar's selected row: accent text on the surface row.
		name: "formulae selected",
		setup: func(m *Model) *Model {
			m.activePanel = PanelFormulae
			m.panels[PanelFormulae].items = []string{"foo  1.0"}
			m.panels[PanelFormulae].formulae = []brew.Formula{{Name: "foo"}}
			m.panels[PanelFormulae].selected = 0
			return m
		},
		want: func(th *style.Theme) []string {
			return []string{sgrColor("38", th.Accent), sgrColor("48", th.Surface)}
		},
	},
	{
		// The operation overlay: the accent title and the subtle status line.
		name: "operation",
		setup: func(m *Model) *Model {
			m.opState = &Operation{Title: "install bash", Status: opRunning}
			return m
		},
		want: func(th *style.Theme) []string {
			return []string{sgrColor("38", th.Accent), sgrColor("38", th.Subtle)}
		},
	},
	{
		// A modal: every row behind it is drawn in the muted role, and the
		// dialog's keys in the accent.
		name: "confirm modal",
		setup: func(m *Model) *Model {
			m.activeModal = modal.NewConfirmModal(modal.ActionDanger, "Uninstall bash", "Remove bash?")
			return m
		},
		want: func(th *style.Theme) []string {
			return []string{sgrColor("38", th.Muted), sgrColor("38", th.Accent)}
		},
	},
}

// TestEveryThemeDrawsItsScreensFromItsRoles: on each theme, each screen at
// 80x24 has the view's canvas in the theme's base and text, carries the
// colours of the theme's roles, and carries no other theme's accent.
func TestEveryThemeDrawsItsScreensFromItsRoles(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	var accents []style.Theme
	for _, name := range style.ThemeNames() {
		th, _ := style.ThemeByName(name)
		accents = append(accents, *th)
	}
	for _, name := range style.ThemeNames() {
		th, _ := style.ThemeByName(name)
		style.ApplyTheme(th)
		for _, sc := range themeScreens {
			t.Run(name+"/"+sc.name, func(t *testing.T) {
				m := sc.setup(updateModel(newTestModel(), tea.WindowSizeMsg{Width: 80, Height: 24}))
				v := m.View()
				if v.BackgroundColor != th.Background {
					t.Errorf("canvas background = %v, want the theme base %v", v.BackgroundColor, th.Background)
				}
				if v.ForegroundColor != th.Text {
					t.Errorf("canvas text = %v, want the theme text %v", v.ForegroundColor, th.Text)
				}
				for _, tok := range sc.want(th) {
					if !strings.Contains(v.Content, tok) {
						t.Errorf("frame does not draw with %q", tok)
					}
				}
				for _, other := range accents {
					if other.Accent != th.Accent && strings.Contains(v.Content, sgrColor("38", other.Accent)) {
						t.Errorf("frame draws with another theme's accent %v", other.Accent)
					}
				}
			})
		}
	}
}
