# Slice: cr2-brew-cache — cache correctness, fewer brew calls, brew-package dedup

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `cr1-correctness` closed. AC-02 of cr1 changed
`LoggingRunner`; AC-03 here wraps stream errors the same way.

## Goal

The brew cache must never hold data that is stale after a write. A refresh
must not run the same `brew` command twice. The readers and writers in
`internal/brew` share one cached-read helper and one stream-write helper
instead of copy-pasted blocks. Small dedups in `internal/brew` and
`internal/config` close out the slice. Public behaviour (what the GUI shows,
which commands run for a write) is unchanged, except where an AC says
otherwise.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — cache `Get` race [C9]
- [ ] AC-02 — generic cached read with singleflight on every reader [R15, P1]
- [ ] AC-03 — writers invalidate after the command, through groups [C7, C8, R16]
- [ ] AC-04 — `brew update` goes through `DiagnosticsWrite.Update` [R3]
- [ ] AC-05 — refresh cost [P2, P3, R14]
- [ ] AC-06 — runner: line splitting and command setup [P14, P16, R20, runner sentinel from D5]
- [ ] AC-07 — JSON decoding cost [P15]
- [ ] AC-08 — small brew/config dedups [R17, R18, R19, R21, casks artifact loop from D5]

### AC-01 — cache `Get` race [C9]

1. `Cache.Get` (`internal/brew/cache.go`) no longer deletes an entry without
   re-checking it under the write lock. The simplest fix: an expired entry is
   a miss and is not deleted, because `Set` overwrites it. Alternatively,
   re-check the timestamp under the write lock. Choose one and say which.
2. Test (written first if you can make it fail deterministically, otherwise a
   `-race` stress test): a concurrent `Set` of a fresh value is never lost to
   an expired-entry delete.

### AC-02 — generic cached read with singleflight on every reader [R15, P1]

1. Add one generic helper in `internal/brew`, e.g.
   `cached[T any](c *Cache, sf *singleflight, key CacheKey, ttl time.Duration, fetch func() (T, error)) (T, error)`.
   It does: get from the cache, type-assert, run `fetch` inside singleflight,
   then set with the TTL.
2. Every reader goes through it: the `List`/`Outdated` methods of formulae and
   casks, taps, services `List`/`Info`, trust, and the doctor/config reads.
   Collapse the near-identical `fetchOutdated`/`Outdated` pairs.
3. Remove `TypedCache`/`NewTypedCache` and their tests (they are replaced; D1).
4. Tests: concurrent `List` calls with an empty cache run the mock runner's
   command **once**. Also, one `RefreshMsg` with Formulae, Casks, Taps and
   Services loaded runs each list command **once** in total, counted through
   the `MockRunner`. That second test is in `internal/gui`.

### AC-03 — writers invalidate after the command, through groups [C7, C8, R16]

1. Each write in `formulae.go`, `casks.go`, `taps.go`, `trust.go`,
   `services.go` and `doctor.go` invalidates its cache keys **after** the
   command finishes, whether it succeeds or fails. It may also invalidate
   before. For streaming writes, wrap the error channel: read it once,
   invalidate, then forward the error. Add one helper, e.g.
   `streamWrite(group string, args ...string)`, and use it in all streaming
   writers.
2. Every write invalidates through `InvalidateGroups` instead of
   hand-written `Invalidate(...)` lists. Add the groups that are missing
   (services). Delete the ones that are never used, or wire them up
   (`cleanup`, `doctor`). Unexport the map (`invalidateGroups`).
3. [C8] `Cleanup` and `Autoremove` invalidate their groups. Autoremove must at
   least clear `KeyFormulaeList`.
4. Tests (written first): (a) a streaming install. While it runs, a `List`
   call fills the cache with old data. After the stream completes, `List`
   returns fresh data from the runner. (b) Autoremove clears the formulae
   list. (c) A table test: each writer invalidates the keys of its group.

### AC-04 — `brew update` goes through `DiagnosticsWrite.Update` [R3]

1. Add one model helper, e.g. `m.startBrewUpdate() tea.Cmd`, used by both
   the `StartUpdateMsg` handler and the `R` key handler in
   `internal/gui/gui.go`. It enqueues a task that calls
   `client.DiagnosticsWrite.Update` (so the `"update"` invalidation group
   runs) instead of `Runner.ExecuteStream(ctx, "update")`. The reset of
   `isUpdating` when the enqueue fails stays in the helper.
2. Test: both entry points run `update` through the mock runner and
   invalidate the update group's keys.

### AC-05 — refresh cost [P2, P3, R14]

1. [P2] An automatic refresh (task completion, mutation result, auto-refresh
   tick) does not clear `KeyDoctorResult`/`KeyConfig`. Only an explicit `R`
   press does. Put a field on `RefreshMsg` (like the existing `LoadOutdated`),
   and don't use a second message type. Test: after a task completes, the
   refresh doesn't run `brew doctor`; after `R`, it does.
2. [P3] The Outdated panel fetches the formulae and casks outdated lists
   concurrently, and Services fetches `List` and `Info` concurrently, inside
   one `tea.Cmd`. Use the pattern from `statusDashboard`. The timeouts per
   read stay the same. Test: with a mock runner that blocks each call for d,
   the panel loads in < 2d.
3. [R14] `invalidatePanelCache` (`internal/gui/gui.go`) gets its keys from
   the brew package (an exported `brew.PanelKeys(...)`, or by group name)
   instead of repeating which keys belong to which panel.

### AC-06 — runner: line splitting and command setup [P14, P16, R20, runner sentinel from D5]

1. [P14] `lineWriter.Write` (`runner.go`) keeps a `bytes.Buffer` and pulls
   out complete lines with `bytes.IndexByte`. It no longer re-joins
   `partial + string(p)` and splits on every write. `\r` also ends a line, so
   brew's progress bars become separate lines rather than one growing line.
   `\r\n` counts as a single break. Tests: lines split across writes, `\r`,
   `\r\n`, and a final partial line flushed on close. Benchmark: 1 MiB
   written in 64-byte chunks with no newline, then the same with newlines.
   Show before and after.
2. [R20, P16] `Execute` and `ExecuteStream` share a `newCmd(ctx, args)` for
   env, stdin and process group. The command line string is computed once per
   call.
3. [D5 sentinel] `findBrewPath` adds the `$HOMEBREW_PREFIX/bin/brew`
   candidate only when the variable is set. The `"/bin/brew"` sentinel and
   its check are removed, and no longer appear in
   `BrewNotFoundError.Searched`. Test with the env var unset and with it set.

### AC-07 — JSON decoding cost [P15]

1. `flexStrings.UnmarshalJSON` decodes into `[]json.RawMessage` and branches
   on the first byte: a string, or an object decoded into
   `struct{ Name, FullName string }`. It no longer goes through
   `[]interface{}`.
2. `bottleJSON` decodes `files` as `map[string]json.RawMessage` (or a type
   that only counts entries). The behaviour is unchanged.
3. Benchmark: parse a fixture of about 300 formulae with runtime
   dependencies. Build it in the test, or generate it from one real entry.
   Show allocations before and after. Existing parser tests and fuzz tests
   stay green.

### AC-08 — small brew/config dedups [R17, R18, R19, R21, casks artifact loop from D5]

1. [R17] Add `decodeEach[T]` for the services `List`/`Info` decode loop.
2. [R18] Merge `needsTerminal` and `refusedForRoot` into one
   `errorMentions(err, markers)` (each keeps its own list of markers).
3. [R19] Add `nonEmptyLines([]byte) []string` and use it in `Leaves`, `Uses`
   and `parseMissingOutput`.
4. [R21] In `config.go`: add `boolEnv(name string, dst *bool) error` for the
   Animations/Preload env switches, and turn the TTL default blocks into a
   loop. Check TTLs `<= 0` in **one** place (config). Add `SetCacheTTLs` to
   `TapsReader`, so `Client.SetCacheTTLs` calls each reader directly with no
   `TTLSetter` type assertion. Remove `TTLSetter` if nothing else uses it.
5. [D5 casks] In the casks artifacts loop: drop the leftover `_ = key` and
   iterate in sorted key order, so the order of artifacts is deterministic.
   Test: the same input gives the same order across 20 runs.

## Invariants

The common invariants apply. In addition: the set of brew commands a write
runs, and their arguments, are unchanged (assert through the `MockRunner`).
The read timeouts (`brewReadTimeout` etc.) are unchanged. The exported names
of `internal/brew` used by `internal/gui` may change only where an AC says
so.
