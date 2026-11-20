package task

import (
	"context"
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// BenchmarkTaskOutput is a task that prints 10 000 lines: the messages the
// reader takes to deliver them (cr3 AC-03). It is not part of the gate: run it
// with -bench TaskOutput.
func BenchmarkTaskOutput(b *testing.B) {
	const lines = 10000
	b.ReportAllocs()
	var msgs int
	for i := 0; i < b.N; i++ {
		m := NewManager(10)
		m.Enqueue(&Task{
			ID:    "bench",
			Title: "bench",
			Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
				out := make(chan string, 256)
				errCh := make(chan error, 1)
				go func() {
					for j := 0; j < lines; j++ {
						out <- fmt.Sprint(j)
					}
					errCh <- nil
					close(out)
				}()
				return out, errCh, nil
			},
		})
		m.RunNext()()
		for done := false; !done; {
			var msg tea.Msg = m.RunNext()()
			switch msg.(type) {
			case TaskOutputMsg:
				msgs++
			case TaskCompletedMsg:
				done = true
			}
		}
	}
	b.ReportMetric(float64(msgs)/float64(b.N), "msgs/run")
}
