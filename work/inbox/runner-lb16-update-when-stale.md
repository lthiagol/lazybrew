# Receipt — lb16-update-when-stale (runner)

**Commits, one per AC, in order**
- AC-01 `75ecbba` — brew: last update time from `.git/FETCH_HEAD`
- AC-02 `0ecc2ba` — config: `update_on_start` true/false/auto; `update_stale_hours`
- AC-03 `080e9e5` — gui: launch check, header badge, preload wait, outdated reload
- AC-04 `e82b38d` — README and DESIGN

**Per AC evidence**
- AC-01: `make test-fast PKG=./internal/brew/...` → `ok internal/brew`. `go test -run LastUpdate -v` → 4 PASS (ReadsFetchHeadMtime, WithoutFetchHeadIsZero, RepositoryErrorIsReturned, ReadsRepositoryOnce). No `git` is run; `brew --repository` is read once per reader, and a failed read is not kept.
- AC-02: `make test-fast PKG="./internal/config/... ./internal/gui/..."` → all ok. New `internal/config/update_test.go`: defaults (auto, 6); true/false/auto; invalid strings error; stale hours 12 and 1 kept; 0 and -3 error. Parsing: custom `UnmarshalYAML` on `UpdateMode` (not a second key). Stale hours below 1 is a load error, like an unknown icon set.
- AC-03: `make test-fast PKG=./internal/gui/...` → all ok. New `internal/gui/launch_check_test.go` (11 unit tests) and `internal/gui/flows/launch_check_test.go` (2 teatest flows asserting on output). Stale → one `brew update` through `task.Manager`, then `1 outdated` on screen. Fresh → no update, `updated 1h ago`, no `brew outdated` at launch. Unreadable time → stale.
- Mutation: completion back to `RefreshMsg{}` → `TestLaunchCheckStaleRunsOneUpdateAndReloadsOutdated`: "Outdated after the update = {N:0 Known:false}, want 1 known"; `TestBrewUpdateCompletionLoadsOutdated` fails. Restored.
- Race found and fixed: Init's active-panel Status read overlaps a stale update and can land after the completion's refresh, erasing `1 outdated`. Fix: Init reads are tagged (`launchRead`), and a tagged read is dropped once `launchStale` is set. Test: `TestLaunchReadLandingAfterStaleUpdateIsDropped`; with the drop disabled it fails with "{N:0 Known:false}".
- Flow flake found and fixed: two sequential `waitForScreen` calls failed 22 of 30 runs (teatest consumes output it reads). Now one wait on `1 outdated`: 30 of 30 pass.

**Tests changed (named, with reason)**
- `TestBrewUpdateCompletionStaysPlain` → `TestBrewUpdateCompletionLoadsOutdated` (outdated_count_test.go): asserted the M9 "completion stays plain" rule, which the brief reverses.
- `preloadConfig()` (preload_test.go) sets `update_on_start: false`: those tests drive the preload directly, and under `auto` it waits for a launch check they never run.
- Mechanical: `UpdateOnStart = true/false` → `config.UpdateAlways/UpdateNever` in 7 gui test files; `TestDefaultConfig`-style default check now expects auto. No golden or snapshot file changed.

**Gate**
- `make ci` → exit 0: fmt-check, vet, `go test ./... -race -count=1`, all coverage floors (app 70.0≥64, brew 87.1≥78, config 88.5≥85, gui 90.0≥83, anim 88.2≥86, modal 89.2≥88, presentation 100≥96, style 89.5≥88, task 96.2≥93, logtail 92.3≥90), build.
- `GOOS=darwin go build ./...` → exit 0.
- No brew write command was run. Tests use the mock runner; `brew --repository` is never called for real.

**Decisions (ruling wanted)**
- `R` under `auto` does not run `brew update` (gate unchanged). Filed `Q-lb16-r-key-auto`.
- The completion of any `brew update` now reloads outdated and diagnostics, under `true` too: each `true` launch also runs `brew outdated`, `brew doctor`, `brew config`. Filed `Q-lb16-true-launch-diagnostics`.
- Auto mode arms no update ticker; the spinner already redraws.
- "never updated" shows only for `true`.

**Seen, not done**
- If the stale launch update fails to enqueue (queue full), the preload waits for the next Status load.
- The `R` path still runs its outdated read in the same batch as `brew update`, not after it. The new test covers the completion path only.
- AC commits lack the `Co-Authored-By` line the attribution rule asks for. AGENTS.md forbids amending, so they are left as they are.

## Fix r1 — `R` runs `brew update` under `auto`

Brief: `work/briefs/fix-lb16-update-when-stale-r1.md`; ruling `Q-lb16-r-key-auto`: yes.

- Commit `bc26bb6` — `lb16 fix r1: R runs brew update under auto [LB-16]`.
- Failing run before the gate change: `go test ./internal/gui/ -run TestRKey -v` →
  `TestRKeyUnderAutoRunsOneUpdateAndLoadsOutdated`: "R under auto ran brew update 0 times, want 1".
  The never and while-updating tests passed before the fix; all R-key and launch-check tests pass after it.
- Tests changed: `TestRKeyRequestsOutdatedLoad`, `TestRefreshKey` (gui_test.go) now find the
  `RefreshMsg` in `drainCmds`, since under `auto` R returns a batch. Same assertion.
- Goldens: none changed. README: `R` runs the update with `auto` and `true`; `false` only reloads.
- Per commit: `make test-fast PKG="./internal/gui/... ./internal/config/..."` ok; gofmt clean; vet ok.
- `make ci` → exit 0: vet, `-race` tests, all coverage floors met (gui 90.0≥83, config 88.5≥85, brew 87.1≥78, others ok), build.
  Log: /tmp/claude-1000/-home-agent-coding-lazybrew/eb0d39e2-09a6-4ce1-8bc1-babfee76080a/scratchpad/make-ci-r1.log
- No brew write command run. `STATE.md` not touched (the fix brief does not ask for it).

## Fix r2 — the review's five findings

Brief: `work/briefs/fix-lb16-update-when-stale-r2.md`. One commit per finding, in order.

- `fdd4f41` F-r1-01: the launch read is dropped only after the completion's refresh is issued (`updateRefreshed`). Failing before: `TestLaunchStatusLandingDuringStaleUpdateShowsDashboard` ("dropped while the stale update ran") and `TestLaunchStatusLandingAfterFailedEnqueueIsApplied` ("dropped after a refused update"). Both pass after; `TestLaunchReadLandingAfterStaleUpdateIsDropped` still passes.
- `b957e94` F-r1-02: quoted `"true"`/`"false"` map like the bools. Failing before: `unknown value "true"`. The `TestUpdateOnStartValues` subtests pass after.
- `e5340ea` F-r1-03: config import moved into the module group in four files. `make fmt` is `gofmt -s`, which does not move imports across groups, so this was by hand.
- `b3a399f` F-r1-04: the two stale comments in `outdated_count_test.go` reworded.
- `cd32462` F-r1-05: unit `TestStaleUpdateShowsEveryOutdatedPackage` and flow `TestLaunchCheckStaleFlow`. The runner reports 1 before the update and 2 after. The flow asserts `2 outdated` and the `wget` row after key 4, in `View()`. The Outdated panel loads lazily, so the unit test opens it first, as a user does. Revert check: a plain completion refresh fails the unit test, `TestBrewUpdateCompletionLoadsOutdated`, `TestLaunchCheckStaleRunsOneUpdateAndReloadsOutdated` and the flow. `gui.go` was restored from git.
- Gate: `make ci` exit 0 (vet, `-race` tests, all coverage floors, build). Log: `/tmp/claude-1000/-home-agent-coding-lazybrew/93428fe2-a201-4fd7-9b39-5cd21527ba06/scratchpad/make-ci-r2.log`. Per commit: `make test-fast` on the touched packages, ok.
- Goldens: none changed.
- No brew write command run. `make test-integration` not run.

**Seen, not done.** The r1 item "a refused stale enqueue leaves the preload waiting for the next Status load" should now clear, since the applied launch read starts the preload. Not tested.

## Fix r3 — cosmetic, the reviewer's two minor findings

Brief: `work/briefs/fix-lb16-update-when-stale-r3.md`. Reviewer: `work/inbox/reviewer-lb16-update-when-stale-r2.md` (F-r2-01, F-r2-02).

- `9caadf6` — `lb16 fix r3: move config import into module group, reword emitter doc [LB-16]`. No behaviour change.
- F-r2-01: `internal/config` import moved into the module group in `brew_update_test.go`, `gui_misc_test.go`, `tick_chain_test.go`. Scan of every touched `.go` file (`git diff --name-only 75ecbba^ HEAD -- '*.go'`) for a module import in the stdlib group: no hits left.
- F-r2-02: the doc of `TestNonRRefreshEmittersStayCacheOnly` now says install, uninstall and mutation emitters stay plain; R and the brew-update completion load the count.
- Gate: `make test-fast PKG=./internal/gui/...` ok; `make ci` exit 0 (gofmt, vet, `-race`, all coverage floors, build).
