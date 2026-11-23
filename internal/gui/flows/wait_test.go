package flows_test

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest/v2"
)

// waitForScreen waits until the rendered output shows text, so a flow presses
// its next key once the screen is ready for it, not after a fixed sleep.
func waitForScreen(t *testing.T, tm *teatest.TestModel, text string) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), text)
	}, teatest.WithDuration(3*time.Second), teatest.WithCheckInterval(10*time.Millisecond))
}

// waitForScreenAll is waitForScreen for texts drawn in one frame. Each wait reads
// only the output after the last one, so a second waitForScreen for text already
// drawn with the first would never match.
func waitForScreenAll(t *testing.T, tm *teatest.TestModel, texts ...string) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		for _, text := range texts {
			if !strings.Contains(string(bts), text) {
				return false
			}
		}
		return true
	}, teatest.WithDuration(3*time.Second), teatest.WithCheckInterval(10*time.Millisecond))
}

// waitForCall waits until the runner has recorded a stream call for verb and
// name, so a flow quits only after the action it started has reached brew.
func waitForCall(t *testing.T, calls func() [][]string, verb, name string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		for _, call := range calls() {
			if len(call) >= 2 && call[0] == verb && call[1] == name {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("no %s %s call, got %v", verb, name, calls())
		}
		time.Sleep(5 * time.Millisecond)
	}
}
