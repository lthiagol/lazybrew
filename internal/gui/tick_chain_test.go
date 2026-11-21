package gui

import (
	"github.com/lthiagol/lazybrew/internal/config"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// tickWait is how long a command may run before countPending calls it still
// pending. A read or a fetch against the mock runner finishes well inside it;
// a tea.Tick sleeps for its whole interval, so the auto-refresh and update
// ticks in these tests (1 s and 10 s) are still running when it expires.
const tickWait = 300 * time.Millisecond

// runCmds runs cmd and the commands of any batch it returns. It returns the
// messages of the commands that finish within tickWait, and how many are still
// running after it. Each command runs in its own goroutine, so a timer is not
// waited for.
func runCmds(cmd tea.Cmd) (msgs []tea.Msg, pending int) {
	if cmd == nil {
		return nil, 0
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		batch, ok := msg.(tea.BatchMsg)
		if !ok {
			return []tea.Msg{msg}, 0
		}
		for _, sub := range batch {
			subMsgs, subPending := runCmds(sub)
			msgs = append(msgs, subMsgs...)
			pending += subPending
		}
		return msgs, pending
	case <-time.After(tickWait):
		return nil, 1
	}
}

// countPending is how many commands of cmd are still running after tickWait.
func countPending(cmd tea.Cmd) int {
	_, pending := runCmds(cmd)
	return pending
}

// hasMsg reports whether msgs holds a message of type T.
func hasMsg[T any](msgs []tea.Msg) bool {
	for _, m := range msgs {
		if _, ok := m.(T); ok {
			return true
		}
	}
	return false
}

// finishedMsgs is the messages of the commands of cmd that finish within
// tickWait. A tick still waiting for its interval contributes none.
func finishedMsgs(cmd tea.Cmd) []tea.Msg {
	msgs, _ := runCmds(cmd)
	return msgs
}

// Auto-refresh keeps exactly one tick chain. Init starts it; every refresh
// after that (task completions, R, the ticks themselves) must leave the chain
// alone, not start another one that would multiply the full refresh.
func TestAutoRefreshKeepsOneTickChain(t *testing.T) {
	m := newTestModel()
	m.cfg.GUI.AutoRefreshSeconds = 1

	pending := countPending(m.Init())
	for i := 0; i < 5; i++ {
		nm, cmd := m.Update(RefreshMsg{})
		m = asModel(nm)
		pending += countPending(cmd)
	}
	if pending != 1 {
		t.Fatalf("auto-refresh: %d tick commands in flight after Init and 5 refreshes, want 1", pending)
	}
}

// With update_on_start, the update ticker is one chain too: R presses while a
// brew update runs must not each add a 10 s ticker.
func TestUpdateTickerKeepsOneChainAcrossRPresses(t *testing.T) {
	m := newTestModel()
	m.cfg.Brew.UpdateOnStart = config.UpdateAlways

	pending := countPending(m.Init())
	for i := 0; i < 3; i++ {
		nm, cmd := m.Update(tea.KeyPressMsg{Code: 'R', Text: "R"})
		m = asModel(nm)
		pending += countPending(cmd)
	}
	if pending != 1 {
		t.Fatalf("update ticker: %d tick commands in flight after Init and 3 R presses, want 1", pending)
	}
}

// A tick that finds the user typing arms the next tick and hands back a paused
// message; handling that message must not arm a second chain.
func TestPausedTickKeepsOneChain(t *testing.T) {
	m := newTestModel()
	m.cfg.GUI.AutoRefreshSeconds = 1
	m.cfg.GUI.AutoRefreshPause = time.Hour
	m.lastKeyAt = time.Now()

	nm, cmd := m.Update(autoRefreshTickMsg{at: time.Now()})
	m = asModel(nm)
	pending := countPending(cmd)
	msgs, _ := runCmds(cmd)
	for _, msg := range msgs {
		nm, cmd := m.Update(msg)
		m = asModel(nm)
		pending += countPending(cmd)
	}
	if pending != 1 {
		t.Fatalf("paused tick: %d tick commands in flight, want 1", pending)
	}
}
