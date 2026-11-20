verdict: ok

# Reviewer — v2-11-cancel-and-test-speed — r1

Range: `7e9b214..aa8edec` (`wip`). The v2-09 commits in the range were reviewed and closed
under v2-09 (r5 ok). Slice commits: AC-01 `d76b8de`, AC-02 `f11d2a6`, AC-02b `02e9cac`,
AC-03 `74c4438`, receipt `60543ae`; orchestrator `aa8edec` (brief paths + ruling only).
Known, not findings: Q-v2-11-a (brief paths), `d76b8de` has no Co-Authored-By trailer.

## AC table

| AC | verdict | evidence (run by the reviewer) |
|---|---|---|
| AC-01 | met | `procgroup_unix.go` (`//go:build unix`): `Setpgid`, `Cancel` = `Kill(-pid, SIGKILL)`, `WaitDelay` 2 s; both `Execute` and `ExecuteStream` call `setProcessGroup`. Go's `watchCtx` calls `Cancel` only before `Wait` reaps the leader, so `-pgid` always exists (no ESRCH path). Hand test with a fake brew: tree brew → sh → sh → sleep, plus a sh → `exec sleep`, writing `dl` every 10 ms; cancel mid-stream: stream closed in 0.45 ms, `context canceled`, all 4 pids ESRCH. Bite check: `setProcessGroup` without the group → both `*KillsGrandchild` tests FAIL (5 s) and the cancel test is back to 30 s. `GOOS=darwin go build ./...` ok. |
| AC-01.3 | met | Same fake brew (`printf 'a\n\nb\nlast'; echo oops >&2; exit 3`) run on the old `runner.go` and the new one: identical stream lines `["a" "" "b" "last"]`, identical errors `brew install foo: oops` / `... exited with code 3`. Log calls unchanged. Only change: the never-reachable `stdout pipe`/`stdout read` errors are gone with the pipes. |
| AC-02 | met | Old vs new read side by side; no assertion dropped. Toast test: first-tick survive + re-arm still asserted; expiry now on a pre-aged `NewToastAt` (existed at `c094035`). Singleflight: the old wait polled a channel closed only after `fn` returned (always 2 s); the new one waits on `waiters` under `sf.mu` — stronger. Cache TTL: fake clock, same hit/miss checks. Fetch budgets shortened, bound formulas unchanged. Flows: sleeps → waits on screen text and on the recorded call (stronger). Production values unchanged: toast tick 500 ms, auto-close 2 s, `killGroupDelay` 2 s, `Cache.now = time.Now`, `brewReadTimeout` untouched. |
| AC-02b | met | `02e9cac` diff: each workflow's `go test` step gains `env: CGO_ENABLED: "1"` and `-race`; nothing else. Missing EOF newline in `wip-ci.yml` existed before. The push and the `wip-ci` run are the orchestrator's (receipt says so). |
| AC-03 | met | `result.md`: 38.9 s → 20.8 s, slowest 10 before/after, every ≥ 100 ms test listed. STATE row is the brief's text. |

## Gate (reviewer run, HEAD `aa8edec`)

`make build` ok · `make lint` ok · `CGO_ENABLED=1 go test ./... -race -count=1` ok, wall
**20.2 s** (result.md: 20.8 s) · `make cover-check` ok (brew 79.2, presentation 98.5,
gui 85.4, modal 90.1; floors unchanged) · `GOOS=darwin go build ./...` ok.
Per commit: `d76b8de` and `f11d2a6` each build and pass `go test` for brew and gui/...

## Flake checks

Shortened-timing tests (`TestFetchStatusData*`, stray-holder, grandchild, cancel,
singleflight, toast tick, cache) with `-race -count=40` while every core ran a busy loop: all pass.
`TestInstallFlow|TestUninstallFlow -race -count=10` pass. `./internal/gui -race -count=5` pass.

## Invariants

No GUI behaviour or frame change: v2-11's GUI production changes are only the
`toastTickInterval` / `operationAutoClose` seams and `ToastTick(interval)`. No test deleted, skipped or
weakened. No new dependency. One commit per AC, in order.

## Findings

None.

## Observation (not a finding) — Q-v2-11-b

`Setpgid`, which the brief requires, takes brew out of the terminal's foreground group. Under a
pty, a `read </dev/tty` in its own group is stopped (state `T`), against `S` before. A `sudo`
password prompt (cask `.pkg`) would now stop silently until Esc. Appended as Q-v2-11-b for the orchestrator.
