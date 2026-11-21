package flows_test

import (
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

// TestFlowModelsAreStill fails if the model the flows drive would animate,
// even with LAZYBREW_ANIMATIONS=1 in the environment: a flow asserts on frames
// at rest.
func TestFlowModelsAreStill(t *testing.T) {
	t.Setenv(config.AnimationsEnv, "1")
	if testutil.NewModel(brewtest.NewMockRunner()).MotionOn() {
		t.Fatal("the flows' model animates: their frames would not be the resting frames")
	}
}
