package brew

import "sync"

// singleflight coalesces concurrent calls to the same key into a single
// underlying invocation. The first caller executes fn(); concurrent callers
// for the same key wait on the same call and receive its result. This is
// the M9 primitive that prevents a stampede when both the GUI's
// `fetchPanelData(PanelOutdated)` and `fetchStatusData` race to fetch
// outdated results.
//
// We intentionally avoid golang.org/x/sync/singleflight to keep the
// dependency surface small — the lazy Outdated fetch path only needs Do
// and shared/first-caller semantics.
type singleflight struct {
	mu       sync.Mutex
	inflight map[string]*sfCall
}

type sfCall struct {
	wg  sync.WaitGroup
	val any
	err error
}

func newSingleflight() *singleflight {
	return &singleflight{inflight: make(map[string]*sfCall)}
}

// Do executes fn for key, coalescing concurrent callers. Exactly one fn
// runs per (sf, key) pair while callers are waiting; subsequent callers
// after fn completes start a fresh call.
func (sf *singleflight) Do(key string, fn func() (any, error)) (any, error) {
	sf.mu.Lock()
	if call, ok := sf.inflight[key]; ok {
		sf.mu.Unlock()
		call.wg.Wait()
		return call.val, call.err
	}
	call := &sfCall{}
	call.wg.Add(1)
	sf.inflight[key] = call
	sf.mu.Unlock()

	// Defer cleanup so a panicking fn unblocks waiters and drops the key.
	defer func() {
		call.wg.Done()
		sf.mu.Lock()
		delete(sf.inflight, key)
		sf.mu.Unlock()
	}()

	call.val, call.err = fn()
	return call.val, call.err
}
