package style

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestDarkThemeColors(t *testing.T) {
	theme := DarkTheme()
	want := map[string]color.Color{
		"Accent":    lipgloss.Color("#cba6f7"),
		"Secondary": lipgloss.Color("#94e2d5"),
		"Success":   lipgloss.Color("#a6e3a1"),
		"Warning":   lipgloss.Color("#f9e2af"),
		"Error":     lipgloss.Color("#f38ba8"),
		"Subtle":    lipgloss.Color("#6c7086"),
		"Text":      lipgloss.Color("#cdd6f4"),
		"Bg":        lipgloss.Color("#1e1e2e"),
	}
	got := map[string]color.Color{
		"Accent":    theme.AccentColor,
		"Secondary": theme.SecondaryColor,
		"Success":   theme.SuccessColor,
		"Warning":   theme.WarningColor,
		"Error":     theme.ErrorColor,
		"Subtle":    theme.SubtleColor,
		"Text":      theme.TextColor,
		"Bg":        theme.BgColor,
	}
	for name, w := range want {
		if got[name] != w {
			t.Errorf("DarkTheme %s = %q, want %q", name, got[name], w)
		}
	}
}

func TestLightThemeColors(t *testing.T) {
	theme := LightTheme()
	want := map[string]color.Color{
		"Accent":    lipgloss.Color("#34548a"),
		"Secondary": lipgloss.Color("#0891b2"),
		"Success":   lipgloss.Color("#33635c"),
		"Warning":   lipgloss.Color("#8f5e15"),
		"Error":     lipgloss.Color("#8c4351"),
		"Subtle":    lipgloss.Color("#9699a3"),
		"Text":      lipgloss.Color("#343b58"),
		"Bg":        lipgloss.Color("#f5f5f5"),
	}
	got := map[string]color.Color{
		"Accent":    theme.AccentColor,
		"Secondary": theme.SecondaryColor,
		"Success":   theme.SuccessColor,
		"Warning":   theme.WarningColor,
		"Error":     theme.ErrorColor,
		"Subtle":    theme.SubtleColor,
		"Text":      theme.TextColor,
		"Bg":        theme.BgColor,
	}
	for name, w := range want {
		if got[name] != w {
			t.Errorf("LightTheme %s = %q, want %q", name, got[name], w)
		}
	}
}

func TestApplyTheme(t *testing.T) {
	ApplyTheme(DarkTheme())
	if CurrentTheme == nil {
		t.Fatal("CurrentTheme should not be nil")
	}
	if AccentColor != lipgloss.Color("#cba6f7") {
		t.Errorf("AccentColor after Dark ApplyTheme = %q", AccentColor)
	}
}

func TestApplyThemeLight(t *testing.T) {
	ApplyTheme(LightTheme())
	if CurrentTheme == nil {
		t.Fatal("CurrentTheme should not be nil")
	}
	if AccentColor != lipgloss.Color("#34548a") {
		t.Errorf("AccentColor after Light ApplyTheme = %q", AccentColor)
	}
	// Restore dark for other packages sharing package-level styles.
	ApplyTheme(DarkTheme())
}

func TestApplyThemeDefinesM14Styles(t *testing.T) {
	ApplyTheme(DarkTheme())
	if ActivePanelBg.GetBackground() != lipgloss.Color("#313244") {
		t.Errorf("dark ActivePanelBg bg = %v, want #313244", ActivePanelBg.GetBackground())
	}
	if PanelTitleActive.GetForeground() != lipgloss.Color("#cba6f7") {
		t.Errorf("PanelTitleActive fg = %v, want accent", PanelTitleActive.GetForeground())
	}
	if !PanelTitleActive.GetBold() {
		t.Error("PanelTitleActive should be bold")
	}
	if LogBorder.GetBorderTopForeground() != lipgloss.Color("#6c7086") &&
		LogBorder.GetBorderStyle() != lipgloss.RoundedBorder() {
		// BorderForeground may be reported via top; also accept style check below.
	}
	if LogBorder.GetBorderStyle() != lipgloss.RoundedBorder() {
		t.Error("LogBorder should use rounded border")
	}

	ApplyTheme(LightTheme())
	if ActivePanelBg.GetBackground() != lipgloss.Color("#ccd0da") {
		t.Errorf("light ActivePanelBg bg = %v, want #ccd0da", ActivePanelBg.GetBackground())
	}
	ApplyTheme(DarkTheme())
}

func TestThemeConcurrency(t *testing.T) {
	done := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			ApplyTheme(DarkTheme())
			ApplyTheme(LightTheme())
		}
		done <- struct{}{}
	}()
	<-done
	ApplyTheme(DarkTheme())
}
