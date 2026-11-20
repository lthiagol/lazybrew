package gui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
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
