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
