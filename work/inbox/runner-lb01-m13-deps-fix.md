# Receipt — runner: lb01-m13-deps-fix

Verdict: done. No blocker — all five mutations bit.

Commits in order: `b7dec88` (F-01 → `work/verification/lb01-m13-deps.md`), `544e0ed` (F-05 → `work/questions.md`).

## F-01 — mutation evidence added, Conclusion rewritten

Per invariant the record now carries the exact diff, the command and the verbatim failing output; the presence-check section is unchanged.

| mut | inv | mutation | test that failed | failure message |
|---|---|---|---|---|
| M1 | I-1 | `internal/gui/commands.go:833` deadline removed | `TestFetchTabContentCmdTimesOut` | `gui_test.go:1310: fetchTabContentCmd hung past fetchTabContentTimeout (infinite-Loading reproducer)` |
| M2 | I-2 | `internal/gui/gui.go:455-456` re-kick deleted | `TestDepsTabPopulatedAfterDataLoaded` / `TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply` | `gui_test.go:1187: expected Deps tabContent populated after DataLoadedMsg, got cmd=nil content="" ok=false` / `gui_test.go:1220: expected loadTabContent cmd re-issued for Used By (no fast path) after DataLoadedMsg` |
| M3 | I-3 | `internal/gui/commands.go:795` fast path never applies | `TestDepsFastPathRendersCachedDependencies` / `TestDepsTabViewLeavesLoadingAfterFastPath` / `TestDepsTabLoads` | `gui_test.go:1329: expected nil cmd from fast path (Dependencies already on model)` / `gui_test.go:1433: expected fast-path nil cmd` / `gui_test.go:1489: Deps body missing pcre2` |
| M4 | I-4 | `internal/brew/formulae.go:318` build deps from `dependencies` | `TestFormulaeReaderList` / `TestParseFormulaBuildDepsFromBuildDependencies` | `formulae_test.go:155: python BuildDeps = [openssl xz zlib], want [pkg-config]` / `formulae_test.go:195: BuildDeps = [runtime-a], want [build-only]` |
| M5 | I-5 | `internal/gui/gui.go:625-630` fetch cmd dropped | `TestDepsSelectionChangeUpdatesContent` | `gui_test.go:1403: expected zlib deps after j, ok=false content=""` |

Every mutation exited 1, was reverted with `git checkout --` right after its run, and edited no test file. `git status` clean.

## F-05 — question appended

Q-03 in `work/questions.md` (raiser: reviewer lane; owner: the orchestrator; resolved): the review brief's invariant→test table was wrong for I-2 and I-5; the corrected mapping lives in `work/briefs/fix-lb01-m13-deps-verify.md` and is recorded with mutation evidence in `work/verification/lb01-m13-deps.md`. The "_None. No questions are open._" line is untouched.

## Gate — on the clean tree after all reverts

`make build` → 0 · `make lint` → 0 · `go test ./... -race -count=1` → 0 · `make cover-check` → 0 ("All coverage floors met")

For the orchestrator: `work/briefs/review-lb01-m13-deps-verify.md` still holds the wrong I-2/I-5 rows; it is the orchestrator's file and this lane did not touch it.
