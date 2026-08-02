package brew

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSingleflightCoalescesConcurrent(t *testing.T) {
	sf := newSingleflight()
	var calls int32

	const goroutines = 16

	// Hold all callers inside fn until we release the gate, so we know
	// they really overlapped.
	gate := make(chan struct{})

	var wg sync.WaitGroup
	results := make([]any, goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := sf.Do("k", func() (any, error) {
				atomic.AddInt32(&calls, 1)
				<-gate
				return "ok", nil
			})
			results[i] = val
			errs[i] = err
		}()
	}

	// Give the first caller time to enter fn; others should be queued.
	time.Sleep(20 * time.Millisecond)
	close(gate)
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 fn invocation, got %d", got)
	}
	for i, r := range results {
		if r != "ok" {
			t.Errorf("results[%d] = %v, want 'ok'", i, r)
		}
		if errs[i] != nil {
			t.Errorf("errs[%d] = %v, want nil", i, errs[i])
		}
	}
}

func TestSingleflightSecondBatchStartsFresh(t *testing.T) {
	sf := newSingleflight()
	var calls int32

	// First batch.
	if _, err := sf.Do("k", func() (any, error) {
		atomic.AddInt32(&calls, 1)
		return "first", nil
	}); err != nil {
		t.Fatal(err)
	}

	// Second batch after first completes should trigger a fresh call.
	if _, err := sf.Do("k", func() (any, error) {
		atomic.AddInt32(&calls, 1)
		return "second", nil
	}); err != nil {
		t.Fatal(err)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 fn invocations across two sequential Do calls, got %d", got)
	}
}

func TestSingleflightKeysAreIndependent(t *testing.T) {
	sf := newSingleflight()
	var a, b int32

	gate := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = sf.Do("a", func() (any, error) {
			atomic.AddInt32(&a, 1)
			<-gate
			return nil, nil
		})
	}()
	go func() {
		defer wg.Done()
		_, _ = sf.Do("b", func() (any, error) {
			atomic.AddInt32(&b, 1)
			<-gate
			return nil, nil
		})
	}()

	time.Sleep(20 * time.Millisecond)
	close(gate)
	wg.Wait()

	if atomic.LoadInt32(&a) != 1 || atomic.LoadInt32(&b) != 1 {
		t.Errorf("keys a and b should each run once independently, got a=%d b=%d", a, b)
	}
}

func TestSingleflightPropagatesError(t *testing.T) {
	sf := newSingleflight()
	wantErr := context.DeadlineExceeded

	type result struct {
		val any
		err error
	}
	results := make(chan result, 2)

	// Gate for fn so we can synchronize the second caller.
	gate := make(chan struct{})

	// firstFnRunning signals the moment first caller's fn is actually
	// executing (and thus holding the inflight slot).
	firstFnRunning := make(chan struct{})
	// secondDoEntered signals that the second caller has called Do and is
	// now queued behind the first caller's inflight slot.
	secondDoEntered := make(chan struct{})

	// First caller: enters fn, signals firstFnRunning, blocks on gate.
	go func() {
		v, err := sf.Do("k", func() (any, error) {
			close(firstFnRunning)
			<-gate
			return nil, wantErr
		})
		results <- result{v, err}
	}()

	<-firstFnRunning

	// Second caller: enters Do. Should be queued, NOT invoke fn.
	go func() {
		v, err := sf.Do("k", func() (any, error) {
			t.Error("second caller must not invoke fn while first is in flight")
			return nil, nil
		})
		results <- result{v, err}
		close(secondDoEntered)
	}()

	// Wait for second caller to reach Do() and get queued.
	// Polling with a generous timeout is more reliable than a fixed sleep.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-secondDoEntered:
			deadline = time.Time{}
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}

	close(gate)

	for i := 0; i < 2; i++ {
		select {
		case r := <-results:
			if r.err != wantErr {
				t.Errorf("result %d err = %v, want %v", i, r.err, wantErr)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for result %d", i)
		}
	}
}

func TestSingleflightPanicUnblocksWaiters(t *testing.T) {
	sf := newSingleflight()
	gate := make(chan struct{})
	started := make(chan struct{})

	leaderDone := make(chan struct{})
	go func() {
		defer close(leaderDone)
		defer func() { _ = recover() }()
		_, _ = sf.Do("k", func() (any, error) {
			close(started)
			<-gate
			panic("boom")
		})
	}()

	<-started

	waiterDone := make(chan struct{})
	go func() {
		defer close(waiterDone)
		// Must unblock when leader panics (defer wg.Done).
		_, _ = sf.Do("k", func() (any, error) {
			t.Error("waiter must not invoke fn while sharing panicked call")
			return nil, nil
		})
	}()

	time.Sleep(20 * time.Millisecond)
	close(gate)

	select {
	case <-waiterDone:
	case <-time.After(2 * time.Second):
		t.Fatal("waiter blocked after leader panic (missing defer cleanup)")
	}
	<-leaderDone

	val, err := sf.Do("k", func() (any, error) {
		return "ok", nil
	})
	if err != nil || val != "ok" {
		t.Errorf("post-panic Do = (%v, %v), want (ok, nil)", val, err)
	}
}
