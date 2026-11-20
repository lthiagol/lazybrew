package gui

import (
	"os"
	"testing"

	"github.com/lthiagol/lazybrew/internal/config"
)

// TestMain runs every test with the Unicode icon set, so frames read as plain
// text. A test that needs the Nerd set clears the variable with t.Setenv.
func TestMain(m *testing.M) {
	os.Setenv(config.IconsEnv, config.IconsUnicode)
	os.Exit(m.Run())
}
