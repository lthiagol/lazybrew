package style

import "testing"

func TestDarkThemeColors(t *testing.T) {
	theme := DarkTheme()
	if theme.AccentColor == "" {
		t.Error("dark accent color should not be empty")
	}
}

func TestLightThemeColors(t *testing.T) {
	theme := LightTheme()
	if theme.AccentColor == DarkTheme().AccentColor {
		t.Error("light theme should differ from dark")
	}
}

func TestApplyTheme(t *testing.T) {
	ApplyTheme(DarkTheme())
	if CurrentTheme == nil {
		t.Fatal("CurrentTheme should not be nil")
	}
}

func TestApplyThemeLight(t *testing.T) {
	ApplyTheme(LightTheme())
	if CurrentTheme == nil {
		t.Fatal("CurrentTheme should not be nil")
	}
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
