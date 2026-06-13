package task

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestManager_Sequential(t *testing.T) {
	m := NewManager(10)

	runCh := make(chan int, 2)

	t1 := &Task{
		ID:    "1",
		Title: "task1",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			runCh <- 1
			errCh := make(chan error, 1)
			errCh <- nil
			close(errCh)
			return nil, errCh, nil
		},
	}
	t2 := &Task{
		ID:    "2",
		Title: "task2",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			runCh <- 2
			errCh := make(chan error, 1)
			errCh <- nil
			close(errCh)
			return nil, errCh, nil
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

	cmd := m.RunNext()
	if cmd == nil {
		t.Fatal("expected RunNext to return a command")
	}

	msg := cmd()
	completed, ok := msg.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}
	if completed.ID != "1" {
		t.Errorf("expected task 1 to complete first, got %s", completed.ID)
	}

	cmd = m.RunNext()
	if cmd == nil {
		t.Fatal("expected RunNext to return a command for task 2")
	}

	msg = cmd()
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

	_, err := m.Enqueue(&Task{ID: "a",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			errCh := make(chan error, 1); errCh <- nil; close(errCh)
			return nil, errCh, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Enqueue(&Task{ID: "b",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			errCh := make(chan error, 1); errCh <- nil; close(errCh)
			return nil, errCh, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Enqueue(&Task{ID: "c",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			errCh := make(chan error, 1); errCh <- nil; close(errCh)
			return nil, errCh, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Enqueue(&Task{ID: "d",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			return nil, nil, nil
		},
	})
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
			errCh := make(chan error, 1)
			errCh <- nil
			close(errCh)
			return nil, errCh, nil
		},
	})

	msg := m.RunNext()()
	if _, ok := msg.(TaskCompletedMsg); !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}

	_, _ = m.Enqueue(&Task{
		ID: "t2",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			ran.Add(1)
			errCh := make(chan error, 1)
			errCh <- nil
			close(errCh)
			return nil, errCh, nil
		},
	})

	msg = m.RunNext()()
	if _, ok := msg.(TaskCompletedMsg); !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}

	if n := ran.Load(); n != 2 {
		t.Errorf("expected both tasks to run, got %d", n)
	}
}
