package gui

import (
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/anim"
)

// probeMsg is a message no handler knows: Update passes it through unchanged,
// so it runs the motion settle alone.
type probeMsg struct{}

// frameAt is a frame tick n frames after start, one 60 Hz frame apart.
func frameAt(start time.Time, n int) anim.TickMsg {
	return anim.TickMsg{At: start.Add(time.Duration(n) * anim.FrameInterval)}
}

func TestTestsRunWithAnimationsOff(t *testing.T) {
	if animationsOn(config.Default()) {
		t.Fatal("the test run animates: the gui tests must run with LAZYBREW_ANIMATIONS=0")
	}
}

func TestAnimationsEnvOverridesTheConfig(t *testing.T) {
	on, off := config.Default(), config.Default()
	off.GUI.Animations = false
	cases := []struct {
		env  string
		cfg  *config.Config
		want bool
	}{
		{"1", off, true},
		{"0", on, false},
		{"", on, true},
		{"", off, false},
		{"yes", on, true}, // an unknown value is ignored, as LAZYBREW_ICONS is
	}
	for _, c := range cases {
		t.Setenv(config.AnimationsEnv, c.env)
		if got := animationsOn(c.cfg); got != c.want {
			t.Errorf("env %q, gui.animations %v: animationsOn = %v, want %v",
				c.env, c.cfg.GUI.Animations, got, c.want)
		}
	}
}

func TestAnimationsOffNeverTicks(t *testing.T) {
	m := newTestModel()
	if m.motionOn {
		t.Fatal("model built under LAZYBREW_ANIMATIONS=0 has motion on")
	}
	m.motion.Snap("probe", 0)
	m.motion.Target("probe", 10)
	if _, cmd := m.Update(probeMsg{}); cmd != nil {
		t.Fatal("a moving value scheduled a frame with animations off")
	}
}

func TestIdleScreenSchedulesNoTick(t *testing.T) {
	t.Setenv(config.AnimationsEnv, "1")
	m := newTestModel()
	if !m.motionOn {
		t.Fatal("LAZYBREW_ANIMATIONS=1 did not turn motion on")
	}
	if _, cmd := m.Update(probeMsg{}); cmd != nil {
		t.Fatal("an idle screen scheduled a frame")
	}
}

func TestMovingValueTicksUntilItSettles(t *testing.T) {
	t.Setenv(config.AnimationsEnv, "1")
	m := newTestModel()
	m.motion.Snap("probe", 0)
	m.motion.Target("probe", 10)

	nm, cmd := m.Update(probeMsg{})
	if cmd == nil {
		t.Fatal("a moving value did not schedule a frame")
	}
	start := time.Now()
	frames := 0
	for ; cmd != nil; frames++ {
		if frames > 60 {
			t.Fatal("the frame tick did not stop")
		}
		nm, cmd = nm.Update(frameAt(start, frames+1))
	}
	if got, _ := m.motion.Get("probe"); got != 10 {
		t.Fatalf("settled at %v; want exactly 10", got)
	}
	if took := time.Duration(frames) * anim.FrameInterval; took > 300*time.Millisecond {
		t.Fatalf("settled after %v; want within 300ms", took)
	}
	if _, cmd := nm.Update(probeMsg{}); cmd != nil {
		t.Fatal("a settled screen scheduled a frame")
	}
}

func TestStrayFrameAfterSettleSchedulesNothing(t *testing.T) {
	t.Setenv(config.AnimationsEnv, "1")
	m := newTestModel()
	if _, cmd := m.Update(frameAt(time.Now(), 1)); cmd != nil {
		t.Fatal("a frame with nothing moving scheduled another frame")
	}
}
