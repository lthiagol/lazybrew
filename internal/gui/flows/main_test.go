package flows_test

import (
	"os"
	"testing"

	"github.com/lthiagol/lazybrew/internal/config"
)

// TestMain runs the flows with the Unicode icon set, so the frames they
// assert on read as plain text.
func TestMain(m *testing.M) {
	os.Setenv(config.IconsEnv, config.IconsUnicode)
	os.Exit(m.Run())
}
