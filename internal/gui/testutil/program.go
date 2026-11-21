package testutil

import (
	"testing"

	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui"
)

// NewModel is the model the helpers drive: the production wiring over the
// runner, with the motion off whatever the environment says, so every frame
// it draws is a resting frame.
//
// The background preload and the launch `brew update` are off too: the flows
// read each panel as they open it, and a preload or an update behind them would
// add reads to what they count. The launch-check flows use NewModelWithUpdate.
func NewModel(runner brew.Runner) *gui.Model {
	return NewModelWithUpdate(runner, config.UpdateNever)
}

// NewModelWithUpdate is NewModel with brew.update_on_start set to mode.
func NewModelWithUpdate(runner brew.Runner, mode config.UpdateMode) *gui.Model {
	cfg := config.Default()
	cfg.GUI.Preload = false
	cfg.Brew.UpdateOnStart = mode
	return gui.New(brew.NewClient(runner), cfg).WithoutMotion()
}

func NewTestModel(t *testing.T, opts ...teatest.TestOption) *teatest.TestModel {
	t.Helper()
	return teatest.NewTestModel(t, NewModel(brewtest.NewMockRunner()), opts...)
}

func NewTestModelWithRunner(t *testing.T, runner brew.Runner, opts ...teatest.TestOption) *teatest.TestModel {
	t.Helper()
	return teatest.NewTestModel(t, NewModel(runner), opts...)
}

// NewTestModelWithUpdate is NewTestModelWithRunner with brew.update_on_start set to mode.
func NewTestModelWithUpdate(t *testing.T, runner brew.Runner, mode config.UpdateMode, opts ...teatest.TestOption) *teatest.TestModel {
	t.Helper()
	return teatest.NewTestModel(t, NewModelWithUpdate(runner, mode), opts...)
}
