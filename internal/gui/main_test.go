package gui

import (
	"os"
	"testing"

	"github.com/lthiagol/lazybrew/internal/config"
)

// TestMain runs every test with the Unicode icon set, so frames read as plain
// text, and with animations off, so every frame is the resting frame. A test
// that needs the Nerd set clears the variable with t.Setenv; a test of motion
// sets LAZYBREW_ANIMATIONS=1 with t.Setenv.
func TestMain(m *testing.M) {
	os.Setenv(config.IconsEnv, config.IconsUnicode)
	os.Setenv(config.AnimationsEnv, "0")
	os.Exit(m.Run())
}

// testConfig is the production default with the background preload off. The
// tests read each panel themselves and count the reads they expect; a preload
// running behind them would add reads to every count. The preload tests turn it
// on by hand, and TestProductionDefaultPreloads pins the production default.
func testConfig() *config.Config {
	cfg := config.Default()
	cfg.GUI.Preload = false
	return cfg
}
