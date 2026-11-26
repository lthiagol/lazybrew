package brew

import (
	"context"
	"fmt"
	"sync"
)

// singleflight coalesces concurrent calls to the same key into a single
// underlying invocation. The first caller starts fn; concurrent callers for
// the same key wait on the same call and receive its result. It lives on
// Cache and serves every reader through cached, so each brew read runs once
// per key however many callers ask at the same time. It is the M9 primitive
// that stops the GUI's `fetchPanelData(PanelOutdated)` and `fetchStatusData`
// from both running `brew outdated` when they race.
//
// fn runs on a context of its own, detached from every caller's: a caller is
// bound only by its own ctx. One whose ctx ends stops waiting and returns its
// ctx error; the others keep waiting. The flight's context is cancelled only
// when the last waiter has left, so a fetch nobody waits for is stopped.
//
// We intentionally avoid golang.org/x/sync/singleflight to keep the
// dependency surface small — the readers need only Do and shared/first-caller
// semantics.
type singleflight struct {
	mu       sync.Mutex
	inflight map[string]*sfCall
}

type sfCall struct {
	done    chan struct{} // closed once fn has returned
	cancel  context.CancelFunc
	waiters int // callers still waiting, guarded by singleflight.mu
	val     any
	err     error
	panic   any // fn's panic value, if it panicked
}

func newSingleflight() *singleflight {
	return &singleflight{inflight: make(map[string]*sfCall)}
}

// Do executes fn for key, coalescing concurrent callers. Exactly one fn
// runs per (sf, key) pair while callers are waiting; subsequent callers
// after fn completes start a fresh call. fn gets the flight's context, which
// carries the first caller's values but none of its deadline or cancel. If fn
// panics, the caller that started it re-panics and the others get an error.
func (sf *singleflight) Do(ctx context.Context, key string, fn func(context.Context) (any, error)) (any, error) {
	sf.mu.Lock()
	call, joined := sf.inflight[key]
	if !joined {
		flightCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
		call = &sfCall{done: make(chan struct{}), cancel: cancel}
		sf.inflight[key] = call
		go sf.run(flightCtx, key, call, fn)
	}
	call.waiters++
	sf.mu.Unlock()

	select {
	case <-call.done:
		if call.panic != nil {
			if !joined {
				panic(call.panic)
			}
			return nil, fmt.Errorf("singleflight %s: %v", key, call.panic)
		}
		return call.val, call.err
	case <-ctx.Done():
		sf.leave(key, call)
		return nil, ctx.Err()
	}
}

// run executes fn for call and wakes its waiters.
func (sf *singleflight) run(ctx context.Context, key string, call *sfCall, fn func(context.Context) (any, error)) {
	defer func() {
		if r := recover(); r != nil {
			call.panic = r
		}
		sf.mu.Lock()
		if sf.inflight[key] == call {
			delete(sf.inflight, key)
		}
		sf.mu.Unlock()
		call.cancel()
		close(call.done)
	}()
	call.val, call.err = fn(ctx)
}

// leave drops a waiter whose ctx ended. The last one to leave cancels the
// flight and drops its key, so a later caller starts afresh instead of
// joining a fetch that is being stopped.
func (sf *singleflight) leave(key string, call *sfCall) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	call.waiters--
	if call.waiters > 0 {
		return
	}
	if sf.inflight[key] == call {
		delete(sf.inflight, key)
	}
	call.cancel()
}
