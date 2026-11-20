package anim

import (
	"math"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

const dt = 1.0 / 60

// stepUntilSettled advances s one frame at a time and returns how many frames
// it took, failing the test if it never settles within limit frames.
func stepUntilSettled(t *testing.T, s *Set, limit int) int {
	t.Helper()
	for i := 1; i <= limit; i++ {
		s.Step(dt)
		if s.Settled() {
			return i
		}
	}
	t.Fatalf("not settled after %d frames", limit)
	return 0
}

func TestNewNameStartsAtTargetWithNothingToAnimate(t *testing.T) {
	s := NewSet()
	s.Target("a", 12)
	if got, ok := s.Get("a"); !ok || got != 12 {
		t.Fatalf("Get(a) = %v, %v; want 12, true", got, ok)
	}
	if !s.Settled() {
		t.Fatal("a value that starts at its target is not settled")
	}
}

func TestEmptySetIsSettled(t *testing.T) {
	if !NewSet().Settled() {
		t.Fatal("an empty set is not settled")
	}
}

func TestGetUnknownName(t *testing.T) {
	if got, ok := NewSet().Get("nope"); ok || got != 0 {
		t.Fatalf("Get(nope) = %v, %v; want 0, false", got, ok)
	}
}

func TestConvergesAndSnapsExactly(t *testing.T) {
	s := NewSet()
	s.Snap("a", 0)
	s.Target("a", 40)
	if s.Settled() {
		t.Fatal("a value 40 cells from its target reports settled")
	}
	stepUntilSettled(t, s, 600)
	got, _ := s.Get("a")
	if got != 40 {
		t.Fatalf("settled position = %v; want exactly 40", got)
	}
}

func TestSettlesWithinTheShortMotionBudget(t *testing.T) {
	// A move of about 35 cells (a long tab or a full sidebar) must settle in
	// 150 to 300 ms, the subtle-and-short budget of the animations.
	s := NewSet()
	s.Snap("a", 0)
	s.Target("a", 35)
	frames := stepUntilSettled(t, s, 600)
	if took := time.Duration(frames) * FrameInterval; took > 300*time.Millisecond {
		t.Fatalf("settled in %v; want at most 300ms", took)
	}
}

func TestNeverOvershoots(t *testing.T) {
	s := NewSet()
	s.Snap("a", 0)
	s.Target("a", 20)
	for i := 0; i < 600 && !s.Settled(); i++ {
		s.Step(dt)
		if got, _ := s.Get("a"); got > 20 {
			t.Fatalf("frame %d overshot to %v", i, got)
		}
	}
}

func TestRetargetMidFlight(t *testing.T) {
	s := NewSet()
	s.Snap("a", 0)
	s.Target("a", 40)
	for i := 0; i < 5; i++ {
		s.Step(dt)
	}
	mid, _ := s.Get("a")
	if mid <= 0 || mid >= 40 {
		t.Fatalf("mid-flight position = %v; want between 0 and 40", mid)
	}
	s.Target("a", 10)
	stepUntilSettled(t, s, 600)
	if got, _ := s.Get("a"); got != 10 {
		t.Fatalf("after retarget settled at %v; want exactly 10", got)
	}
}

func TestSnapIsInstant(t *testing.T) {
	s := NewSet()
	s.Snap("a", 0)
	s.Target("a", 40)
	s.Snap("a", 5)
	if got, _ := s.Get("a"); got != 5 {
		t.Fatalf("Snap left %v; want 5", got)
	}
	if !s.Settled() {
		t.Fatal("a snapped value is not settled")
	}
}

func TestStepIgnoresNonPositiveDt(t *testing.T) {
	s := NewSet()
	s.Snap("a", 0)
	s.Target("a", 40)
	s.Step(0)
	s.Step(-1)
	if got, _ := s.Get("a"); got != 0 {
		t.Fatalf("non-positive dt moved the value to %v", got)
	}
}

func TestRemove(t *testing.T) {
	s := NewSet()
	s.Snap("a", 1)
	s.Remove("a")
	if _, ok := s.Get("a"); ok {
		t.Fatal("Remove left the value behind")
	}
}

func TestTickDeliversATickMsg(t *testing.T) {
	msg := Tick()()
	if _, ok := msg.(TickMsg); !ok {
		t.Fatalf("Tick delivered %T; want TickMsg", msg)
	}
	if FrameInterval <= 0 || math.Abs(FrameInterval.Seconds()-1.0/60) > 1e-6 {
		t.Fatalf("FrameInterval = %v; want 1/60 s", FrameInterval)
	}
}

// Tick must be a command, not a goroutine the caller manages.
var _ tea.Cmd = Tick()
