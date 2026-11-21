package testutil

import (
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/config"
)

// TestHelpersBuildStillModels fails if a helper-built model would animate,
// even with LAZYBREW_ANIMATIONS=1 in the environment.
func TestHelpersBuildStillModels(t *testing.T) {
	t.Setenv(config.AnimationsEnv, "1")
	if NewModel(brewtest.NewMockRunner()).MotionOn() {
		t.Fatal("NewModel animates: the helpers must build still models")
	}
	if NewModel(brewtest.NewMockRunner()).MotionOn() {
		t.Fatal("NewModel with a runner animates")
	}
}
