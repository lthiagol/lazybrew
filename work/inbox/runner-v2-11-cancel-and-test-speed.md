# Receipt — runner — v2-11-cancel-and-test-speed

Commits, in AC order: AC-01 `d76b8de`, AC-02 `f11d2a6`, AC-02b `02e9cac`, AC-03 `74c4438`. This receipt is the last commit. Worktree clean. Nothing pushed.

**AC-01 — cancel kills the process group.** Both runner paths set `Setpgid` and a cancel sends SIGKILL to `-pgid`, with `WaitDelay` 2 s (`internal/brew/procgroup_unix.go`; no-op in `procgroup_other.go`). `ExecuteStream` now uses cmd writers instead of pipes, because with pipes `Wait` could not run while a stray holder kept the pipe open, so `WaitDelay` never applied. Output and error text are unchanged. Evidence: `TestExecuteStreamCancelReportsCanceled` 30.00 s → 0.01 s. New tests kill a grandchild (stream and `Execute`) and cut off a `setsid` holder; each fails when its fix is removed. `GOOS=darwin go build ./...` ok.

**AC-02 — gui tests stop waiting on the wall clock.** Seams: `toastTickInterval`, `operationAutoClose`, `killGroupDelay`, `Cache.now`, and a waiter count in `singleflight`. Production durations are unchanged. The 0.5 s batch tests, `TestNonRRefreshEmittersStayCacheOnly` (2 s), `TestToastTickMsgEventuallyClearsToast` (3.3 s), `TestSingleflightPropagatesError` (2 s), `TestCacheTTL` (0.1 s) and the flows' sleeps are gone. Each changed test, and what it still asserts, is in `work/verification/v2-11-cancel-and-test-speed/result.md`. Remaining waits are listed there too: the bubbles progress frame tick in three bar tests (about 0.37 s each), and `-race` render cost.

**AC-02b — CI race.** `wip-ci.yml` and `stable-ci.yml`: `go test ./... -race -count=1`, `CGO_ENABLED: "1"`. YAML checked with `yaml.v3`. **The `wip-ci` run after the push must be checked by the orchestrator.** I did not push.

**AC-03 — evidence.** `result.md` has the race run before and after, and the slowest 10 tests before and after. Every test at or above 100 ms is listed with both times.

**Gate (final state, `02e9cac`):**
- `make build`: pass
- `make lint`: pass (`go vet ./...`, no output)
- `CGO_ENABLED=1 go test ./... -race -count=1`: pass, 0 failing. Wall 38.9 s before, 20.8 s after.
- `make cover-check`: pass (brew 79.2 ≥ 62, presentation 98.5 ≥ 91, gui 85.3 ≥ 36, modal 90.1 ≥ 41)

**Contradictions and notes for the next session:**
- The brief's Report protocol and Fix-rounds name the v2-10 receipt path. I wrote `work/inbox/runner-v2-11-cancel-and-test-speed.md` and named it in the reply. Question `Q-v2-11-a` is in `work/questions.md`.
- `d76b8de` (AC-01) has no `Co-Authored-By` trailer. I had not yet read the attribution rule. The brief forbids amending, so it stays. Later commits have it.
- `STATE.md` row is the brief's literal text, with `—` in the commits column.
