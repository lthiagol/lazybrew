// Package anim holds the named values the GUI moves, each driven by a
// harmonica spring, and the frame tick that advances them. A value springs
// from where it is towards its target; once it is within half a cell of the
// target and no longer moving, it is snapped exactly onto it, so a settled
// frame is the frame an animation-free run would draw.
package anim

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/harmonica"
)

const (
	// Frequency and Damping shape the spring. The damping is critical, so a
	// value never overshoots its target, and the frequency settles a move of
	// about 35 cells in roughly 250 ms.
	Frequency = 30.0
	Damping   = 1.0

	// settleDistance is how close to its target a value must be to snap onto
	// it: half a cell, so the snap is never a visible jump.
	settleDistance = 0.5
	// settleSpeed is the speed, in cells a second, under which a value close
	// to its target counts as at rest. A faster value would jump on the snap.
	settleSpeed = 5.0

	// MaxStep caps the time one step may advance. A long pause between two
	// ticks must not fling a value across the screen.
	MaxStep = 50 * time.Millisecond
)

// FrameInterval is the time between two frames: 60 a second.
var FrameInterval = time.Duration(harmonica.FPS(60) * float64(time.Second))

// TickMsg is one animation frame. At is when the tick fired; the gap from the
// previous frame is the time a step advances.
type TickMsg struct {
	At time.Time
}

// Tick is the command that delivers the next frame after FrameInterval. A
// caller schedules it only while Set.Settled is false, so an idle screen
// never ticks.
func Tick() tea.Cmd {
	return tea.Tick(FrameInterval, func(at time.Time) tea.Msg {
		return TickMsg{At: at}
	})
}

// value is one animated number: where it is, how fast it moves, and where it
// is going.
type value struct {
	pos, vel, target float64
}

// Set is the named values one model animates. The zero value is not usable;
// build one with NewSet.
type Set struct {
	values map[string]*value
}

// NewSet returns an empty Set.
func NewSet() *Set {
	return &Set{values: make(map[string]*value)}
}

// Clone returns an independent copy of the set: changes to either are not seen
// by the other.
func (s *Set) Clone() *Set {
	c := NewSet()
	for name, v := range s.values {
		cv := *v
		c.values[name] = &cv
	}
	return c
}

// Target sets where the named value is going. A new name starts at target,
// with nothing to animate; a known name springs from where it is now.
func (s *Set) Target(name string, to float64) {
	v, ok := s.values[name]
	if !ok {
		s.Snap(name, to)
		return
	}
	v.target = to
}

// Snap puts the named value at to, at rest, with no animation.
func (s *Set) Snap(name string, to float64) {
	s.values[name] = &value{pos: to, target: to}
}

// Get is the current position of the named value, and whether it exists.
func (s *Set) Get(name string) (float64, bool) {
	v, ok := s.values[name]
	if !ok {
		return 0, false
	}
	return v.pos, true
}

// Names lists the names of every value in the set, in no particular order.
func (s *Set) Names() []string {
	names := make([]string, 0, len(s.values))
	for name := range s.values {
		names = append(names, name)
	}
	return names
}

// Remove forgets the named value.
func (s *Set) Remove(name string) {
	delete(s.values, name)
}

// Step advances every value by dt seconds. A value within half a cell of its
// target and moving slower than settleSpeed is snapped onto the target. A
// dt that is not positive does nothing.
func (s *Set) Step(dt float64) {
	if dt <= 0 {
		return
	}
	spring := harmonica.NewSpring(dt, Frequency, Damping)
	for _, v := range s.values {
		v.pos, v.vel = spring.Update(v.pos, v.vel, v.target)
		if math.Abs(v.pos-v.target) < settleDistance && math.Abs(v.vel) < settleSpeed {
			v.pos, v.vel = v.target, 0
		}
	}
}

// Settled reports whether every value is exactly at its target and at rest.
// A set with nothing in it is settled, so a screen with nothing to move
// schedules no tick.
func (s *Set) Settled() bool {
	for _, v := range s.values {
		if v.pos != v.target || v.vel != 0 {
			return false
		}
	}
	return true
}
