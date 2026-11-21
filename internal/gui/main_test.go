package gui

import (
	"os"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/config"
)

// TestMain runs every test with the Unicode icon set, so frames read as plain
// text, with animations off, so every frame is the resting frame, and with the
// toast tick and the operation's auto-close at a millisecond, so no test waits
// on the wall clock for them. A test
// that needs the Nerd set clears the variable with t.Setenv; a test of motion
// sets LAZYBREW_ANIMATIONS=1 with t.Setenv.
func TestMain(m *testing.M) {
	os.Setenv(config.IconsEnv, config.IconsUnicode)
	os.Setenv(config.AnimationsEnv, "0")
	toastTickInterval = time.Millisecond
	operationAutoClose = time.Millisecond
	os.Exit(m.Run())
}

// testConfig is the production default with the background preload and the
// launch update off. The tests read each panel themselves and count the reads
// they expect; a preload or a `brew update` running behind them would add reads
// to every count. The preload tests turn the preload on by hand, the launch
// tests set update_on_start themselves, and TestProductionDefaultPreloads and
// TestProductionDefaultUpdatesOnEveryLaunch pin the production defaults.
func testConfig() *config.Config {
	cfg := config.Default()
	cfg.GUI.Preload = false
	cfg.Brew.UpdateOnStart = config.UpdateNever
	return cfg
}
