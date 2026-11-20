verdict: ok

# Reviewer — v2-11-cancel-and-test-speed — r2

New since r1 (`25f49f2`): orchestrator `6fdb3a9` (ruling Q-v2-11-b + fix brief r1),
Q-v2-11-b-1 `1872c96`, Q-v2-11-b-2 `5174d37`, receipt `d2173e8`. One commit per item, in order.
r1 had no findings; its open observation (Q-v2-11-b) is the subject of this round.

## Fix items

| item | verdict | evidence (run by the reviewer) |
|---|---|---|
| Q-v2-11-b-1 | met | `procgroup_unix.go`: `Setpgid` → `Setsid: true`; `Cancel` still `Kill(-pid, SIGKILL)` (session leader pid = pgid), `WaitDelay` kept. **Pty experiment repeated:** brew test binary run under `script` (controlling tty checked by opening `/dev/tty`). A fake brew doing `read x </dev/tty` through `NewClient(r).Runner.ExecuteStream` fails in **0.58 ms**, prints `NOTTY`, ctx not expired. Bite check with `Setpgid` restored: the same test ran into the 5 s timeout (`signal: killed`), and the runner's `TestBrewCannotReadTheTerminal` FAILs (5.00 s). **Grandchildren:** fake brew → `sh` → `sleep 30 &`, streaming `dl` every 10 ms, cancel mid-stream: `context canceled`, brew/mid/leaf pids all ESRCH. Existing grandchild, stray-holder and after-exit tests pass. `1872c96` alone: `go test ./internal/brew/` ok. Darwin build + vet ok. |
| Q-v2-11-b-2 | met | `NeedsSudoError` (`errors.go`) uses the same wording as `ServiceNeedsRootError`; `Unwrap` keeps the original error, so `errors.As(*BrewExitError)` and `errors.Is(context.Canceled)` still work. `sudoRunner` (`sudo.go`) wraps the runner in `NewClient`; converts only on the three brief phrases, in the error text or `BrewExitError.Stderr`. Hand test: stderr `sudo: a terminal is required to read the password` → error `needs sudo — run "brew install --cask x" in a terminal`. The services path still yields `ServiceNeedsRootError`: `rootMarkers` contains `sudo`, and `Unwrap` exposes the stderr. No production code type-asserts `Client.Runner` or asserts `*BrewExitError` without `errors.As` (grep). The stream wrapper forwards a single error and closes, and the mock's closed `errCh` passes through. Tests: `sudo_test.go` (3 phrases, stream, non-sudo identity), `gui/sudo_toast_test.go` (exact toast `install foo: needs sudo — run "brew install --cask foo" in a terminal`, other failure unchanged). |

## Gate (reviewer run, HEAD `d2173e8`)

`make build` ok · `make lint` ok · `CGO_ENABLED=1 go test ./... -race -count=1` ok, all 11
packages, wall **21.3 s** (receipt 23.2 s) · `make cover-check` ok (brew 79.6, presentation 98.5,
gui 85.3, modal 90.1; floors unchanged) · `GOOS=darwin go build ./...` + `go vet ./internal/brew/` ok.

## Invariants

No GUI production change and no workflow change in this round (diff `25f49f2..HEAD` on
`internal/gui` non-test and `.github`: empty). No test deleted, skipped or weakened. No new dependency.
No sudo and no real brew run: fake brew scripts only. Reviewer scratch tests removed; worktree clean.

## Findings

None.

## Notes (not findings)

- `TestBrewCannotReadTheTerminal` is linux only (`/dev/ptmx` ioctls), so darwin has no pty test.
  `Setsid` behaves the same there. The real sudo prompt was not exercised, because that needs a
  live brew write, which the human owns.
- The trade-off the ruling accepts: anything else brew runs that asks on the tty (for example a git
  credential prompt) now also fails at once instead of prompting over the TUI.
