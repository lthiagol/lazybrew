# Fix receipt: cr2-brew-cache r1 (runner)

Source: `work/inbox/reviewer-cr2-brew-cache-r1.md`. Brief: `work/briefs/fix-cr2-brew-cache-r1.md`.

## Commits
- `c86d029` cr2 fix-r1 F-r1-01: per-key invalidation generation (tests and fix)
- `492ab80` cr2 fix-r1 F-r1-02: singleflight doc names every cached reader

## F-r1-01: per-key generation (`internal/brew/cache.go`, `cached_test.go`)
Test first, run on the r1 head with the fix not yet applied:
```
--- FAIL: TestCachedUnrelatedInvalidationKeepsFetch (0.02s)
    cached_test.go:212: outdated ran 2 times, want 1: the unrelated write made a second caller fetch
FAIL	github.com/lthiagol/lazybrew/internal/brew
```
- (a) `TestCachedUnrelatedInvalidationKeepsFetch`: fails on the old code (above).
- (b) `TestCachedGroupInvalidationDropsItsFetch`: passes on the old code. It
  guards the stale-store and no-join behaviour, which must stay.
- (c) `TestCachedInvalidateAllDropsEveryFetch`: passes on the old code. It
  guards that `InvalidateAll` drops every in-flight result.
- Existing `TestCachedDropsFetchThatRacedAWrite` and
  `TestCachedCallAfterWriteDoesNotJoinOldFetch` stay green.
- Fix: one counter `seq`. `Invalidate` stamps each dropped key in `keyGen`.
  `InvalidateAll` stamps `allGen`. A key's generation is
  `max(keyGen[key], allGen)`. `cached` captures it for the key's flight key, and
  `setIfCurrent` compares it under the write lock.
- Race: `CGO_ENABLED=1 go test ./internal/brew/ -race -count=20 -run 'TestCached|TestBrewReads|TestCacheExpired|TestCacheInvalidate'` → `ok 9.588s`.

## F-r1-02: singleflight doc (`internal/brew/singleflight.go`)
Comment rewritten to say the primitive lives on `Cache` and serves every reader
through `cached`. No code change.

## Gate
- Per commit: `make test-fast PKG=./internal/brew/...` → `ok` for `c86d029` and `492ab80`.
- `make ci` → exit 0: gofmt check, `go vet`, `go test ./... -race -count=1` (all ok), every coverage floor OK (brew 86.6% ≥ 78%), build. Full output in the scratchpad at `/tmp/claude-1000/-home-agent-coding-lazybrew/f992d232-553a-4b2b-9fd1-1eccfec61203/scratchpad/ci.txt`.
- `GOOS=darwin go build ./...` → pass.
- staticcheck and deadcode: not run, as in r1. Not a builder finding.

## Next session must know
- `work/STATE.md` row for cr2-brew-cache still reads `implemented | awaiting reviewer`. This round did not change it.
- The generation map gets one entry per key ever invalidated. The keys are the fixed `CacheKey` constants, so it stays small.
- No question filed in `work/questions.md`.
- Seen, not done: `Brewfile bundle` streams through `Runner` with no cache group (pre-existing, out of scope).
