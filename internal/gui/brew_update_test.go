package gui

import (
	"context"
	"github.com/lthiagol/lazybrew/internal/config"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// TestBrewUpdateEntryPointsUseDiagnosticsUpdate: the startup update and the R
// key both run `brew update` through DiagnosticsWrite.Update, so the update
// group's cache keys are dropped once it finishes.
func TestBrewUpdateEntryPointsUseDiagnosticsUpdate(t *testing.T) {
	entries := []struct {
		name  string
		start func(m *Model) tea.Cmd
	}{
		{"StartUpdateMsg", func(m *Model) tea.Cmd {
			_, cmd := m.Update(StartUpdateMsg{})
			return cmd
		}},
		{"R key", func(m *Model) tea.Cmd {
			_, cmd := m.Update(tea.KeyPressMsg{Code: 'R', Text: "R"})
			return cmd
		}},
	}
	updateKeys := []brew.CacheKey{brew.KeyFormulaeList, brew.KeyDoctorResult, brew.KeyTapsList}

	for _, e := range entries {
		t.Run(e.name, func(t *testing.T) {
			var ran []string
			runner := &brewtest.MockRunner{
				ExecuteStreamFn: func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
					ran = append(ran, strings.Join(args, " "))
					lines := make(chan string)
					close(lines)
					errs := make(chan error, 1)
					close(errs)
					return lines, errs
				},
			}
			client := brew.NewClient(runner)
			for _, key := range updateKeys {
				client.Cache.Set(key, "seeded")
			}

			m := New(client, testConfig())
			m.cfg.Brew.UpdateOnStart = config.UpdateAlways
			drainCmds(e.start(m))
			// The start command only announces the task. The task manager
			// reads the stream one chunk at a time, when the app asks for
			// the next one, so ask until the update completes.
			for cmd := m.tasks.RunNext(); cmd != nil; cmd = m.tasks.RunNext() {
				if _, done := cmd().(TaskCompletedMsg); done {
					break
				}
			}

			if len(ran) != 1 || ran[0] != "update" {
				t.Errorf("ran %q, want [update]", ran)
			}
			for _, key := range updateKeys {
				if _, ok := client.Cache.Get(key); ok {
					t.Errorf("%s still cached after brew update", key)
				}
			}
		})
	}
}
