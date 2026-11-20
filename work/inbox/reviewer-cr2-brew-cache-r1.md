verdict: changes-required

# reviewer: cr2-brew-cache r1 (wip @ a5e02ab, commits 48378ab..e0f08db)

## Gate (V1)
`make ci`: pass (fmt, vet, `-race -count=1` all ok, every coverage floor met, build). `GOOS=darwin go build ./...`: pass.
staticcheck/deadcode: **not run**. `go run honnef.co/go/tools/cmd/staticcheck@v0.6.1` was denied by the auto-mode classifier here too (the runner hit the same denial). Needs the user to allow it or run it; it is not counted as a builder finding.

## AC table (V2: each fix reverted in a temp worktree, the tests re-run)
| AC | Verdict | Evidence |
|---|---|---|
| 01 | ok | Get no longer deletes. Revert cache.go: `TestCacheExpiredGetKeepsConcurrentSet` fails under `-race` (`iteration 189: fresh Set was lost`), and in 2 of 30 runs without race. That is a stress test, which the brief allows. |
| 02 | ok, F-r1-01 | Revert readers: `TestBrewReadsConcurrentCallsRunOnce` fails (`ran 20 times … want 1` × 8), gui `TestRefreshReadsEachListOnce` fails (`ran 2 times` × 4). Drop the `c.gen != gen` check: `TestCachedDropsFetchThatRacedAWrite` and `…DoesNotJoinOldFetch` fail. TypedCache is gone. |
| 03 | ok | Revert writers: streaming, autoremove, cleanup and table tests all fail. Every writer is either `streamWrite` or a `defer InvalidateFor(group)`. Args unchanged (diff). |
| 04 | ok | Revert gui.go/commands.go: `TestBrewUpdateEntryPointsUseDiagnosticsUpdate` fails on both entry points. One helper, `startBrewUpdate`. |
| 05 | ok | Revert: doctor/config, outdated 601ms, services 602ms, and `TestInvalidatePanelCacheDropsItsKeys` all fail. Reads overlap inside one tea.Cmd. |
| 06 | ok | Revert runner.go: `TestLineWriterSplitsLines` fails on the CR/CRLF/progress rows. `newCmd` is shared, `cmdLine` is computed once, and the `/bin/brew` sentinel is gone. |
| 07 | ok | Benches below. Parser and deps tests green. |
| 08 | ok | Revert casks.go: `TestParseCaskArtifactsKeepSortedOrder` fails. `decodeEach`, `errorMentions`, `nonEmptyLines`, `boolEnv`, TTL loop all present. `TTLSetter` removed. |

## Benchmarks re-run (V5), arm64, `-benchmem`, count=3
- LineWriter no newlines, 5x: before 0.88–1.24 s/op, 8656705 KB, 49152 allocs. After 0.37–0.65 ms/op, 3145664 B, 16 allocs. Matches the receipt.
- LineWriter newlines, 5x: before ~1.48–1.55 ms, 1572864 B, 32768 allocs. After ~1.44–1.54 ms, 1048640 B, 16385 allocs. The allocs match. The receipt's time gain (1.51 → 1.28 ms) is within noise here: not a finding.
- DecodeFormulaeList, 200x: before ~1.86 ms, 1265 KB, 19010 allocs. After ~1.70 ms, 1008 KB, 13560 allocs. Matches.

## Focus points
- **(1) gen/storable.** I found no path that stores pre-write data. `setIfCurrent` checks the generation under the write lock, and a caller that starts after a write gets a new flight key. A not-storable TapList goes to joiners exactly as it goes to the leader. Joiners share the leader's ctx, so they also get its error. Every reader's deadline is the same `brewReadTimeout` (or the cap under the status total), so this changes nothing in practice. **But** the generation is cache-wide → F-r1-01.
- **(2) SetCacheTTLs(0).** This is not a break. Readers start at 0, which means the cache default. The only production caller (`app.go`) passes config values that `Default()`/`loadFile` force to > 0. The tests set fields once. Only a second call with 0 behaves differently, and no caller makes one.
- **(3) streamWrite.** `DefaultRunner.ExecuteStream` closes errChan only after `cmd.Wait` returns, and that includes the kill on cancel. So the invalidation runs after the process is gone, on success, failure and cancel alike. The error is forwarded after the invalidation, so the completion refresh sees the dropped keys. The streaming writers no longer invalidate before the command, which the brief allows.

## Findings
**F-r1-01 (major) — internal/brew/cache.go `Cache.gen` / `cached` (~:182).**
- **What:** the generation counter is cache-wide. Any invalidation, of any key, does two things to every fetch already running:
  - it drops that fetch's result, even when the fetch's own key was not touched;
  - it stops new callers of that key from joining the running fetch.
- **Why it matters:** this undoes the M9 singleflight on `brew outdated` and its 30-minute TTL. `panelCacheGroups` leaves outdated out of "status" for exactly this reason (gui.go:1285).
- **Evidence:** probe test, removed afterwards. Start `cached(KeyOutdatedFormulae)` with the fetch blocked, call `InvalidateFor("taps")`, then a second `cached(KeyOutdatedFormulae)` → `fetch runs=2` (before AC-02: 1). With no second caller, the first result is not cached at all. In the GUI this is reachable: an auto-refresh tick or a task-completion `RefreshMsg` (gui.go:719, it invalidates "status") arriving while the preload or `LoadOutdated` `brew outdated` is in flight forces a second `brew outdated`.
- **Fix must:** keep a per-key generation, bumped only for the keys an invalidation drops; `InvalidateAll` bumps them all.
- **Tests must show:**
  - invalidating an unrelated group while a fetch runs neither drops its result nor runs the command a second time for a concurrent caller (fails today);
  - the existing raced-a-write tests stay green.

**F-r1-02 (minor) — internal/brew/singleflight.go:5-14.** The doc still says the primitive serves "the lazy Outdated fetch path only". After AC-02 it lives on `Cache` and serves every reader through `cached`. Fix: update the comment.

## Invariants (V6)
- Read timeouts unchanged (`brewReadTimeout` 30s and the rest are untouched in the diff).
- No golden or testdata changes. `preload_test.go` only adds a mutex.
- Exported brew changes: `TypedCache`, `TTLSetter` and `InvalidateGroups` (now unexported) are gone, and `SetCacheTTLs` is on `TapsReader`. Each is covered by its AC.
- The new goroutines run inside `tea.Cmd` closures, and the update goes through `task.Manager`.
- One commit per AC, in order. No cr3/cr4/cr5 files touched.
- Seen, not a finding (pre-existing, out of scope): `Brewfile bundle` (commands.go:415) streams through `Runner` directly, with no cache group.
