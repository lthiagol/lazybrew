package flows_test

import (
	"os"
	"testing"

	"github.com/lthiagol/lazybrew/internal/config"
)

// TestMain runs the flows with the Unicode icon set, so the frames they
// assert on read as plain text, and with animations off, so the frames are
// the resting frames.
func TestMain(m *testing.M) {
	os.Setenv(config.IconsEnv, config.IconsUnicode)
	os.Setenv(config.AnimationsEnv, "0")
	os.Exit(m.Run())
}
