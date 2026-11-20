verdict: ok

# reviewer: cr2-brew-cache r2 (wip @ db96acd, fix commits c86d029, 492ab80)

## Gate (V1)
- `make ci`: exit 0. That covers fmt, vet, `-race -count=1` (all ok), every coverage floor and the build.
- `GOOS=darwin go build ./...`: exit 0.
- staticcheck/deadcode: still not run. The auto-mode classifier denies the remote `go run …@version` (see r1). This needs the user to allow it or run it themselves. It is not a builder finding.

## Earlier findings
| Finding | Status | Evidence |
|---|---|---|
| F-r1-01 (major) | **fixed** | See the four checks below. |
| F-r1-02 (minor) | **fixed** | The `singleflight.go` doc now says it lives on `Cache` and serves every reader through `cached`. Comment only. |

F-r1-01 checks:
- **The fix:** `Cache` keeps `seq`, `keyGen[key]` and `allGen`. A key's generation is `max(keyGen[key], allGen)`. `cached` puts it in the flight key, and `setIfCurrent` re-reads it under the write lock.
- **Revert check:** with `c86d029^`'s cache.go in a temp worktree, `TestCachedUnrelatedInvalidationKeepsFetch` fails: `cached_test.go:212: outdated ran 2 times, want 1`. With the fix it passes.
- **Kept behaviour:** `TestCachedGroupInvalidationDropsItsFetch`, `TestCachedInvalidateAllDropsEveryFetch`, `TestCachedDropsFetchThatRacedAWrite` and `…DoesNotJoinOldFetch` all pass.
- **Race stress:** `-race -count=20` over `TestCached|TestBrewReads|TestCacheExpired` plus my probe: ok, no race reported.

## Per-key generation: analysis and probe (probe test removed afterwards)
- **Monotonic.** `seq` only grows, and every stamp is a fresh `seq`. Any drop of a key (`Invalidate` with it, or `InvalidateAll`) therefore moves `max(keyGen, allGen)` strictly upward. A fetch that started before the drop always fails `setIfCurrent` and never shares a flight key with a later caller. No stale store is possible.
- **No wrong drop.** A key's generation moves only when that key, or every key, is dropped. These leave a running fetch's result stored:
  - `InvalidateFor` of another group (probed);
  - an unknown group (probed);
  - an empty `Invalidate()` (probed).
- **Key invalidated before any fetch.** `keyGen` is already set when the fetch notes it, so the result stores (probed).
- **InvalidateAll cases:**
  - a fetch after `InvalidateAll` stores (probed);
  - an `InvalidateAll` during a fetch drops it, even when the key's `keyGen` is older (probed);
  - an `Invalidate(key)` during a fetch that began after an `InvalidateAll` drops it (probed).
- **Locking.** `generationLocked` is called only under `c.mu`, from `generation` (read lock) or `setIfCurrent` (write lock). `keyGen` is written only under the write lock.
- **Window between the miss and the generation read.** A fetch storing in between is caught by the inner `cachedValue` check. An invalidation in between only gives a newer generation. Both are safe.
- **Unchecked Set.** `Set`/`SetWithTTL` skip the generation check, but no production caller exists outside `cache.go` (grep). Every reader stores through `setIfCurrent`.
- **Map size.** `keyGen` holds at most one entry per `CacheKey` constant.

## New problems
None. The fix range (`a4f7ba6..HEAD`) touches only `cache.go`, `cached_test.go`, `singleflight.go`, the fix brief and the receipt. No exported name changed and no other behaviour moved.

Note, not counted as a finding: the fix receipt `runner-cr2-brew-cache-fix-r1.md` is 43 lines, over the 30-line limit in cr-common. It is process only; the orchestrator can trim it or let it stand.

## AC status
AC-01 … AC-08 stay as accepted in r1. Nothing in the fix range touches them, and the full gate is green.
