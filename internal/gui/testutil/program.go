package testutil

import (
	"testing"

	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui"
)

// NewModel is the model the helpers drive: the production wiring over the
// runner, with the motion off whatever the environment says, so every frame
// it draws is a resting frame.
//
// The background preload is off too: the flows read each panel as they open
// it, and a preload behind them would add reads to what they count.
func NewModel(runner brew.Runner) *gui.Model {
	cfg := config.Default()
	cfg.GUI.Preload = false
	return gui.New(brew.NewClient(runner), cfg).WithoutMotion()
}

func NewTestModel(t *testing.T, opts ...teatest.TestOption) *teatest.TestModel {
	t.Helper()
	return teatest.NewTestModel(t, NewModel(brew.NewMockRunner()), opts...)
}

func NewTestModelWithRunner(t *testing.T, runner brew.Runner, opts ...teatest.TestOption) *teatest.TestModel {
	t.Helper()
	return teatest.NewTestModel(t, NewModel(runner), opts...)
}
