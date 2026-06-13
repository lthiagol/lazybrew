package task

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

var assertAnError = errors.New("test error")

func afterTask() (<-chan string, <-chan error, error) {
	errCh := make(chan error, 1)
	errCh <- nil
	return nil, errCh, nil
}

func TestManager_Sequential(t *testing.T) {
	m := NewManager(10)

	runCh := make(chan int, 2)

	t1 := &Task{
		ID:    "1",
		Title: "task1",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			runCh <- 1
			return afterTask()
		},
	}
	t2 := &Task{
		ID:    "2",
		Title: "task2",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			runCh <- 2
			return afterTask()
		},
	}

	_, err := m.Enqueue(t1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Enqueue(t2)
	if err != nil {
		t.Fatal(err)
	}

	if m.IsRunning() {
		t.Fatal("expected manager to be idle before RunNext")
	}

	msg := m.RunNext()()
	if _, ok := msg.(TaskStartedMsg); !ok {
		t.Fatalf("expected TaskStartedMsg, got %T", msg)
	}

	if !m.IsRunning() {
		t.Fatal("expected manager to be running after start")
	}

	msg = m.RunNext()()
	completed, ok := msg.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}
	if completed.ID != "1" {
		t.Errorf("expected task 1 to complete first, got %s", completed.ID)
	}

	msg = m.RunNext()()
	if _, ok := msg.(TaskStartedMsg); !ok {
		t.Fatalf("expected TaskStartedMsg for task 2, got %T", msg)
	}

	msg = m.RunNext()()
	completed, ok = msg.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}
	if completed.ID != "2" {
		t.Errorf("expected task 2 to complete, got %s", completed.ID)
	}

	close(runCh)
	var order []int
	for v := range runCh {
		order = append(order, v)
	}
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Errorf("expected order [1 2], got %v", order)
	}

	if m.RunNext() != nil {
		t.Fatal("expected nil when queue empty")
	}
	if m.IsRunning() {
		t.Fatal("expected manager to be idle")
	}
}

func TestManager_QueueFull(t *testing.T) {
	m := NewManager(3)

	_, err := m.Enqueue(&Task{ID: "a", Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
		return afterTask()
	}})
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Enqueue(&Task{ID: "b", Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
		return afterTask()
	}})
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Enqueue(&Task{ID: "c", Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
		return afterTask()
	}})
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Enqueue(&Task{ID: "d", Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
		return nil, nil, nil
	}})
	if err != ErrQueueFull {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}

	if m.IsRunning() {
		t.Fatal("expected idle before RunNext")
	}
}

func TestManager_RunNextIdle(t *testing.T) {
	m := NewManager(10)
	if cmd := m.RunNext(); cmd != nil {
		t.Fatal("expected nil when no tasks queued")
	}
}

func TestManager_EnqueueAfterCompletion(t *testing.T) {
	m := NewManager(10)
	var ran atomic.Int32

	_, _ = m.Enqueue(&Task{
		ID: "t1",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			ran.Add(1)
			return afterTask()
		},
	})

	msg := m.RunNext()()
	if _, ok := msg.(TaskStartedMsg); !ok {
		t.Fatalf("expected TaskStartedMsg, got %T", msg)
	}

	msg = m.RunNext()()
	if _, ok := msg.(TaskCompletedMsg); !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}

	_, _ = m.Enqueue(&Task{
		ID: "t2",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			ran.Add(1)
			return afterTask()
		},
	})

	msg = m.RunNext()()
	if _, ok := msg.(TaskStartedMsg); !ok {
		t.Fatalf("expected TaskStartedMsg, got %T", msg)
	}

	msg = m.RunNext()()
	if _, ok := msg.(TaskCompletedMsg); !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}

	if n := ran.Load(); n != 2 {
		t.Errorf("expected both tasks to run, got %d", n)
	}
}

func TestManager_StreamingOutput(t *testing.T) {
	m := NewManager(10)
	expected := []string{"a", "b", "c"}

	task := &Task{
		ID:    "s1",
		Title: "Stream",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			out := make(chan string)
			errCh := make(chan error, 1)
			go func() {
				for _, s := range expected {
					out <- s
				}
				errCh <- nil
				close(out)
			}()
			return out, errCh, nil
		},
	}

	m.Enqueue(task)

	msg := m.RunNext()()
	if _, ok := msg.(TaskStartedMsg); !ok {
		t.Fatalf("expected TaskStartedMsg, got %T", msg)
	}

	var got []string
	for {
		msg := m.RunNext()()
		switch v := msg.(type) {
		case TaskOutputMsg:
			got = append(got, v.Line)
		case TaskCompletedMsg:
			goto done
		default:
			t.Fatalf("unexpected msg type %T", msg)
		}
	}
done:

	if len(got) != len(expected) {
		t.Fatalf("got %d lines, want %d: %v", len(got), len(expected), got)
	}
	for i, s := range expected {
		if got[i] != s {
			t.Errorf("line %d: got %q, want %q", i, got[i], s)
		}
	}

	if task.Status != StatusSuccess {
		t.Errorf("expected StatusSuccess, got %v", task.Status)
	}
}

func TestManager_RunImmediateError(t *testing.T) {
	m := NewManager(10)

	task := &Task{
		ID: "fail",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			return nil, nil, assertAnError
		},
	}

	m.Enqueue(task)
	msg := m.RunNext()()
	comp, ok := msg.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}
	if comp.Err != assertAnError {
		t.Errorf("expected error, got %v", comp.Err)
	}
	if task.Status != StatusFailed {
		t.Errorf("expected StatusFailed, got %v", task.Status)
	}
}

func TestManager_EnqueueStartsWhenIdle(t *testing.T) {
	m := NewManager(10)

	task := &Task{
		ID: "t1",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			return afterTask()
		},
	}

	started, err := m.Enqueue(task)
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if !started {
		t.Fatal("expected started=true when manager is idle")
	}
}

func TestManager_EnqueuePendingWhenRunning(t *testing.T) {
	m := NewManager(10)

	t1 := &Task{
		ID: "t1",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			return make(chan string), make(chan error, 1), nil
		},
	}

	m.Enqueue(t1)
	m.RunNext()()

	t2 := &Task{
		ID: "t2",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			return afterTask()
		},
	}

	started, err := m.Enqueue(t2)
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if started {
		t.Fatal("expected started=false when a task is already running")
	}
}

func TestManager_CancelCurrent(t *testing.T) {
	m := NewManager(10)
	cancelled := make(chan struct{})

	task := &Task{
		ID: "cancelme",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			out := make(chan string)
			errCh := make(chan error, 1)
			go func() {
				<-ctx.Done()
				close(cancelled)
				errCh <- ctx.Err()
			}()
			return out, errCh, nil
		},
	}

	m.Enqueue(task)
	m.RunNext()()

	m.CancelCurrent()

	<-cancelled

	msg := m.RunNext()()
	comp, ok := msg.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}
	if comp.ID != "cancelme" {
		t.Errorf("completed ID = %q, want cancelme", comp.ID)
	}
	if comp.Err == nil {
		t.Fatal("expected non-nil error after cancel")
	}
}
