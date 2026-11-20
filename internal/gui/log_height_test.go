package gui

import (
	"context"
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// blockingTask is a task that starts and never finishes, so the manager reads as
// busy for the rest of the test.
func blockingTask() *task.Task {
	return &task.Task{
		ID: "blocker",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			return make(chan string), make(chan error), nil
		},
	}
}

// TestLogBoxHeightByState (v2-09 AC-04): the Log box is one body row while idle
// and empty, at most five body rows while idle with lines, and a fifth of the
// body while a task runs. Collapsed, it is gone. The main box and the Log fill
// the right column exactly in every state.
func TestLogBoxHeightByState(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {140, 40}, {200, 60}} {
		bodyH := sz[1] - headerRows - bottomBarRows
		full := bodyH / 5
		idle := func() *Model {
			return updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
		}
		cases := []struct {
			name string
			m    func() *Model
			want int
		}{
			{"idle and empty", idle, 3},
			{"idle with lines", func() *Model {
				m := idle()
				m.commandLog.Append("brew outdated")
				return m
			}, min(full, 7)},
			{"task running", func() *Model {
				m := idle()
				_, _ = m.tasks.Enqueue(blockingTask())
				m.tasks.RunNext()
				return m
			}, full},
			{"collapsed", func() *Model {
				m := idle()
				m.logCollapsed = true
				return m
			}, 0},
		}
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s/%dx%d", c.name, sz[0], sz[1]), func(t *testing.T) {
				l := c.m().screenLayout()
				if l.logH != c.want {
					t.Fatalf("Log box is %d rows, want %d", l.logH, c.want)
				}
				if l.mainH+l.logH != bodyH {
					t.Fatalf("main %d + Log %d should fill the %d body rows", l.mainH, l.logH, bodyH)
				}
			})
		}
	}
}

// TestLogBoxHeightChangesAtOnce: the height follows the state the moment it
// changes, with no animation in between.
func TestLogBoxHeightChangesAtOnce(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	if got := m.screenLayout().logH; got != 3 {
		t.Fatalf("idle and empty at 140x40: Log is %d rows, want 3", got)
	}
	m.commandLog.Append("brew outdated")
	if got := m.screenLayout().logH; got != 7 {
		t.Fatalf("idle with lines at 140x40: Log is %d rows, want 7", got)
	}
	m = updateModel(newTestModel(), tea.WindowSizeMsg{Width: 200, Height: 60})
	m.commandLog.Append("brew outdated")
	_, _ = m.tasks.Enqueue(blockingTask())
	m.tasks.RunNext()
	if got := m.screenLayout().logH; got != 11 {
		t.Fatalf("a task at 200x60 should give the Log a fifth of the body: 58/5 = 11, got %d", got)
	}
}
