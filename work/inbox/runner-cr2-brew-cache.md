# runner receipt: cr2-brew-cache

**Commits, one per AC:** AC-01 `48378ab` · AC-02 `7d79b31` · AC-03 `e4fe8ee` · AC-04 `7edc2f6` · AC-05 `02f3991` · AC-06 `6a4f7e2` · AC-07 `733ed25` · AC-08 `e0f08db`

## Evidence (failing run on the pre-change code, then pass)
- **AC-01 [C9]** `TestCacheExpiredGetKeepsConcurrentSet` (stress): before, `cache_test.go:285: iteration 26: fresh Set was lost (got <nil>, ok=false)`. After: passes with `-race -count=3`.
- **AC-02 [R15, P1]** `TestBrewReadsConcurrentCallsRunOnce`, before: eight readers `ran 20 times for 20 concurrent callers, want 1` (the two outdated readers already had singleflight). After: all ten pass. gui `TestRefreshReadsEachListOnce`, before: `ran 2 times for one refresh, want 1` for all four lists; after: pass.
- **AC-03 [C7, C8, R16]** `TestStreamingWriteInvalidatesAfterItFinishes`, before: `read after the install got [before-install], want after-install`. `TestAutoremoveClearsFormulaeList`, before: `autoremove left the formulae list cached`. `TestWritesInvalidateTheirGroup` (29 writers, each on success and failure): before, `services` and `cleanup` groups missing; after: all pass.
- **AC-04 [R3]** `TestBrewUpdateEntryPointsUseDiagnosticsUpdate`, before: `formulae:list / doctor:result / taps:list still cached after brew update` on both entry points; after: pass.
- **AC-05 [P2, P3, R14]** before: `an automatic refresh ran doctor 1 times and config 1 times, want 0 and 0`; `outdated panel took 601ms, want under 600ms` (services alike); `Status: doctor:result was dropped but is not this panel's`. After: all pass; at d=300ms both panels take 0.30s.
- **AC-06 [P14, P16, R20, D5]** `TestLineWriterSplitsLines`, before: `lines = ["one\rtwo\rthree"], want ["one" "two" "three"]` (CR, CRLF and progress cases). After: pass, and `TestLineWriterIndependentOfChunking` passes at every chunk size. `TestBrewCandidatesNamePrefixOnlyWhenSet`, before (old list shimmed in a worktree): `brewCandidates() = ["/bin/brew" ...]`. After: pass.
  - Benchmark, 1 MiB in 64-byte writes, `-benchtime=5x`:
  - no newlines: before `874410062 ns/op 8656705356 B/op 49153 allocs/op`, after `447665 ns/op 3145686 B/op 16 allocs/op`
  - newlines: before `1514996 ns/op 1572864 B/op 32768 allocs/op`, after `1275055 ns/op 1048640 B/op 16385 allocs/op`
- **AC-07 [P15]** `BenchmarkDecodeFormulaeList` (300 formulae, runtime deps as objects), `-benchtime=200x`: before `1843393 ns/op 1265442 B/op 19011 allocs/op`, after `1709910 ns/op 1008488 B/op 13560 allocs/op`. Four new `flexStrings` edge rows pass before and after. `FuzzParseFormula` and `FuzzParseCask`, 15 s each: no failures.
- **AC-08 [R17, R18, R19, R21, D5]** `TestParseCaskArtifactsKeepSortedOrder`, before: `artifacts = ["X.app" "a" "x.pkg" "bin/x"], want [... "bin/x" "a" "x.pkg"]`. After: pass over 20 parses. The other dedups are covered by existing tests, which pass unchanged.

## Gate
- `make ci` (after AC-08): fmt check, `go vet ./...`, `go test ./... -race -count=1`, all coverage floors met (brew 86.6% ≥ 78, config 85.0% ≥ 85, gui 86.9% ≥ 83), build. Pass.
- `GOOS=darwin go build ./...`: pass. `make test-fast` per commit: pass.
- **deadcode and staticcheck: NOT RUN.** The Bash call `go run …deadcode@latest` was denied by the auto-mode classifier (it fetches an unpinned remote tool). I did not retry it, and neither tool is installed here. Needs running before the review.

## Decisions beyond the brief
- AC-01: an expired entry is a miss and is not deleted.
- AC-02: singleflight moves onto `Cache`. `cached` takes a fetch that reports whether its result is storable (a TapList with TrustUnknown is not stored). A generation counter stops a read that began before a write from storing pre-write data (the "never stale after a write" rule).
- AC-03: `streamWrite(ctx, runner, cache, group, args...)`. `doctor` group deleted (no writer); `cleanup` drops formulae and casks lists; `services` group added.
- AC-05: panel cache groups live in brew's `invalidateGroups`; `RefreshMsg.ReloadDiagnostics` is set by R only.
- AC-08: readers no longer guard TTL <= 0; config is the one check. `Client.SetCacheTTLs` with 0 now resets a reader to the default.
- AC-06: `brewCandidates()` split out for testing. This machine has brew at `/home/linuxbrew/.linuxbrew` with `HOMEBREW_PREFIX` set, so the not-found path of `findBrewPath` cannot run here.

## Test changes
- Removed `TestTypedCacheWrongTypeReturnsMiss` (TypedCache removed, AC-02.3).
- `preloadRunner`'s call log now takes a mutex. AC-05 made panel reads concurrent and `-race` flagged the unlocked append. No expectation changed.

## Seen, not done
- `internal/config/config.go` `BrewConfig` doc says TTL `<= 0` keeps the 30s default; the loader sets the class default (pre-existing).
- `Diagnostics.Version` reads `KeyConfig` outside `cached`.
- The formulae and casks `fetchOutdated` bodies are still separate (different JSON shapes).

Nothing pushed. `work/STATE.md` row added.
