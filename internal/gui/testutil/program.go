package testutil

import (
	"testing"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/thiago/lazybrew/internal/brew"
	"github.com/thiago/lazybrew/internal/config"
	"github.com/thiago/lazybrew/internal/gui"
)

func NewTestModel(t *testing.T, opts ...teatest.TestOption) *teatest.TestModel {
	t.Helper()
	cfg := config.Default()
	client := brew.NewClient(brew.NewMockRunner())
	m := gui.New(client, cfg)
	tm := teatest.NewTestModel(t, m, opts...)
	return tm
}
