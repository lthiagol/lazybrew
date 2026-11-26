package brew

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestSingleflightCoalescesConcurrent(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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
				val, err := sf.Do(context.Background(), "k", func(context.Context) (any, error) {
					atomic.AddInt32(&calls, 1)
					<-gate
					return "ok", nil
				})
				results[i] = val
				errs[i] = err
			}()
		}

		// Wait until the first caller is inside fn and the others are queued
		// behind it.
		synctest.Wait()
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
	})
}

func TestSingleflightSecondBatchStartsFresh(t *testing.T) {
	sf := newSingleflight()
	var calls int32

	// First batch.
	if _, err := sf.Do(context.Background(), "k", func(context.Context) (any, error) {
		atomic.AddInt32(&calls, 1)
		return "first", nil
	}); err != nil {
		t.Fatal(err)
	}

	// Second batch after first completes should trigger a fresh call.
	if _, err := sf.Do(context.Background(), "k", func(context.Context) (any, error) {
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
	synctest.Test(t, func(t *testing.T) {
		sf := newSingleflight()
		var a, b int32

		gate := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = sf.Do(context.Background(), "a", func(context.Context) (any, error) {
				atomic.AddInt32(&a, 1)
				<-gate
				return nil, nil
			})
		}()
		go func() {
			defer wg.Done()
			_, _ = sf.Do(context.Background(), "b", func(context.Context) (any, error) {
				atomic.AddInt32(&b, 1)
				<-gate
				return nil, nil
			})
		}()

		// Both keys must be inside fn before either gate opens.
		synctest.Wait()
		close(gate)
		wg.Wait()

		if atomic.LoadInt32(&a) != 1 || atomic.LoadInt32(&b) != 1 {
			t.Errorf("keys a and b should each run once independently, got a=%d b=%d", a, b)
		}
	})
}

func TestSingleflightPropagatesError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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

		// First caller: enters fn, signals firstFnRunning, blocks on gate.
		go func() {
			v, err := sf.Do(context.Background(), "k", func(context.Context) (any, error) {
				close(firstFnRunning)
				<-gate
				return nil, wantErr
			})
			results <- result{v, err}
		}()

		<-firstFnRunning

		// Second caller: enters Do. Should be queued, NOT invoke fn.
		go func() {
			v, err := sf.Do(context.Background(), "k", func(context.Context) (any, error) {
				t.Error("second caller must not invoke fn while first is in flight")
				return nil, nil
			})
			results <- result{v, err}
		}()

		// Release fn only once the second caller is durably blocked behind the
		// first. If it had not queued, it would run fn itself and fail the test.
		synctest.Wait()

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
	})
}

func TestSingleflightPanicUnblocksWaiters(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sf := newSingleflight()
		gate := make(chan struct{})
		started := make(chan struct{})

		leaderDone := make(chan struct{})
		leaderPanic := make(chan any, 1)
		go func() {
			defer close(leaderDone)
			defer func() { leaderPanic <- recover() }()
			_, _ = sf.Do(context.Background(), "k", func(context.Context) (any, error) {
				close(started)
				<-gate
				panic("boom")
			})
		}()

		<-started

		waiterErr := make(chan error, 1)
		go func() {
			// Must unblock when leader panics (defer wg.Done).
			_, err := sf.Do(context.Background(), "k", func(context.Context) (any, error) {
				t.Error("waiter must not invoke fn while sharing panicked call")
				return nil, nil
			})
			waiterErr <- err
		}()

		// The waiter must be queued behind the leader before the leader panics.
		synctest.Wait()
		close(gate)

		select {
		case err := <-waiterErr:
			if err == nil || !strings.Contains(err.Error(), "boom") {
				t.Errorf("waiter err = %v, want an error that mentions the panic value boom", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("waiter blocked after leader panic (missing defer cleanup)")
		}
		<-leaderDone
		if got := <-leaderPanic; got != "boom" {
			t.Errorf("leader recovered %v, want the panic value boom re-raised", got)
		}

		val, err := sf.Do(context.Background(), "k", func(context.Context) (any, error) {
			return "ok", nil
		})
		if err != nil || val != "ok" {
			t.Errorf("post-panic Do = (%v, %v), want (ok, nil)", val, err)
		}
	})
}

// TestSingleflightCancelledLeaderDoesNotFailJoiner: the caller that started a
// flight leaves when its ctx is cancelled; a joiner with a live ctx still gets
// the result, and the fetch did not see the leader's cancel.
func TestSingleflightCancelledLeaderDoesNotFailJoiner(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sf := newSingleflight()
		gate := make(chan struct{})
		leaderCtx, cancelLeader := context.WithCancel(context.Background())

		leaderErr := make(chan error, 1)
		go func() {
			_, err := sf.Do(leaderCtx, "k", func(ctx context.Context) (any, error) {
				<-gate
				return "ok", ctx.Err()
			})
			leaderErr <- err
		}()
		synctest.Wait()

		type result struct {
			val any
			err error
		}
		joiner := make(chan result, 1)
		go func() {
			v, err := sf.Do(context.Background(), "k", func(context.Context) (any, error) {
				t.Error("joiner must not run fn")
				return nil, nil
			})
			joiner <- result{v, err}
		}()
		synctest.Wait()

		cancelLeader()
		if err := <-leaderErr; !errors.Is(err, context.Canceled) {
			t.Errorf("leader err = %v, want context.Canceled", err)
		}
		close(gate)
		if r := <-joiner; r.err != nil || r.val != "ok" {
			t.Errorf("joiner = (%v, %v), want (ok, nil)", r.val, r.err)
		}
	})
}

// TestSingleflightJoinerReturnsAtItsDeadline: a joiner with a short deadline
// returns when it passes, without waiting out the leader's fetch.
func TestSingleflightJoinerReturnsAtItsDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sf := newSingleflight()
		gate := make(chan struct{})
		leaderDone := make(chan struct{})
		go func() {
			defer close(leaderDone)
			_, _ = sf.Do(context.Background(), "k", func(context.Context) (any, error) {
				<-gate
				return "ok", nil
			})
		}()
		synctest.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		start := time.Now()
		_, err := sf.Do(ctx, "k", func(context.Context) (any, error) {
			t.Error("joiner must not run fn")
			return nil, nil
		})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("joiner err = %v, want context.DeadlineExceeded", err)
		}
		if d := time.Since(start); d != time.Second {
			t.Errorf("joiner returned after %v, want 1s", d)
		}
		close(gate)
		<-leaderDone
	})
}

// TestSingleflightLastWaiterLeavingCancelsFlight: when every caller has left,
// the fetch's ctx is cancelled and the key is free for a fresh flight, even
// while the cancelled fetch is still running.
func TestSingleflightLastWaiterLeavingCancelsFlight(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sf := newSingleflight()
		ctx, cancel := context.WithCancel(context.Background())
		stopped := make(chan struct{})
		release := make(chan struct{})
		defer close(release)
		done := make(chan error, 1)
		go func() {
			_, err := sf.Do(ctx, "k", func(fctx context.Context) (any, error) {
				<-fctx.Done()
				close(stopped)
				// Hold the fetch open after its ctx ends: run drops the key only
				// when fn returns, so a key that leave did not drop stays taken.
				<-release
				return nil, fctx.Err()
			})
			done <- err
		}()
		synctest.Wait()
		cancel()
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want context.Canceled", err)
		}
		<-stopped

		type result struct {
			val any
			err error
		}
		fresh := make(chan result, 1)
		go func() {
			v, err := sf.Do(context.Background(), "k", func(context.Context) (any, error) {
				return "fresh", nil
			})
			fresh <- result{v, err}
		}()
		synctest.Wait()
		select {
		case r := <-fresh:
			if r.err != nil || r.val != "fresh" {
				t.Errorf("Do after abandon = (%v, %v), want (fresh, nil)", r.val, r.err)
			}
		default:
			t.Fatal("a caller after the last waiter left joined the cancelled fetch instead of starting its own")
		}
	})
}
