# Fix brief — cr2-brew-cache, round r1

Source: `work/inbox/reviewer-cr2-brew-cache-r1.md`. Fix every finding, one
commit each, subject names the id. Brief `work/briefs/cr2-brew-cache.md` and
`work/briefs/cr-common.md` still apply. The eight AC commits are already on
`wip`; do not redo them.

- **F-r1-01 (major)** — `internal/brew/cache.go` `Cache.gen` and `cached`
  (~:182). The generation counter is cache-wide, so invalidating any key drops
  the result of every fetch in flight and stops new callers joining it. That
  runs `brew outdated` twice when an unrelated group is invalidated mid-fetch
  (e.g. a task-completion `RefreshMsg` invalidating "status" during a preload),
  which undoes the singleflight and the 30-minute TTL.
  Make the generation **per key**: bump it only for the keys an invalidation
  drops (`Invalidate`, `InvalidateFor`/groups); `InvalidateAll` bumps every key.
  `cached` captures the generation of its own key before the fetch, and
  `setIfCurrent` and the singleflight key compare that key's generation only.
  Tests, written first and failing on the current code:
  (a) start `cached(KeyOutdatedFormulae)` with the fetch blocked, invalidate an
  unrelated group (`taps`), start a second `cached(KeyOutdatedFormulae)`: the
  fetch runs **once**, and the first result is cached;
  (b) the same, but invalidating a group that **includes** the fetched key:
  stale result not stored, a new caller does not join the old fetch (the
  existing `TestCachedDropsFetchThatRacedAWrite` and `…DoesNotJoinOldFetch`
  stay green);
  (c) `InvalidateAll` mid-fetch drops every in-flight result.
- **F-r1-02 (minor)** — `internal/brew/singleflight.go:5-14`: the doc says it
  serves "the lazy Outdated fetch path only". Update it: it lives on `Cache`
  and serves every reader through `cached`.

Not a finding, do not touch: `staticcheck`/`deadcode` were not run (tool fetch
is denied here); `Brewfile bundle` streaming without a cache group.

Gate: `make test-fast PKG=./internal/brew/...` per commit, then `make ci`.
Receipt `work/inbox/runner-cr2-brew-cache-fix-r1.md` (≤ 30 lines, committed
once), worktree clean, reply with exactly one line:
`done: work/inbox/runner-cr2-brew-cache-fix-r1.md`.
