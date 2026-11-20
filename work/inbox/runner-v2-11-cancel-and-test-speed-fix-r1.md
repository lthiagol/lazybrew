# Receipt — runner — v2-11-cancel-and-test-speed, fix round r1

Commits, in item order: Q-v2-11-b-1 `1872c96`, Q-v2-11-b-2 `5174d37`. This receipt is the last commit. Worktree clean. Nothing pushed.

**Q-v2-11-b-1 — new session instead of process group.** `setProcessGroup` now sets `Setsid: true` (`internal/brew/procgroup_unix.go`). brew is its own process group and has no controlling terminal, and the cancel still kills `-pgid`. `WaitDelay` kept. Darwin build and vet ok.
- Test `TestBrewCannotReadTheTerminal` (`procgroup_linux_test.go`, linux only): a helper test process runs on a pty as its controlling terminal. The fake brew does `read x </dev/tty` and must print `no-terminal` at once. Bite check: with `Setpgid` the helper reports `signal: killed` after the 5 s timeout and the test fails.
- Grandchild, setsid-holder, stray-holder and after-exit tests all pass with `Setsid`.

**Q-v2-11-b-2 — needs-sudo toast.** New `brew.NeedsSudoError{Command, Err}`; the message is `needs sudo — run "brew <args>" in a terminal`, and `Unwrap` keeps the command's own error. `sudoRunner` (`internal/brew/sudo.go`) wraps the runner in `NewClient`, so every brew command gets it. It converts only when the text or stderr contains `a terminal is required`, `no tty present` or `a password is required`. The task toast is then `<title>: needs sudo — run "brew <args>" in a terminal`. Other failures pass through as the same error value and keep their toast.
- Tests with the mock runner: `sudo_test.go` (three sudo phrases, stream path, non-sudo passthrough by identity) and `internal/gui/sudo_toast_test.go` (sudo toast text, other-failure toast). Bite check: without the wrapper the toast test fails.

**Gate (final state, `5174d37`):**
- `make build`: pass
- `make lint`: pass
- `CGO_ENABLED=1 go test ./... -race -count=1`: pass, all 11 packages ok, 23.2 s wall
- `make cover-check`: pass (brew 79.6 ≥ 62, presentation 98.5 ≥ 91, gui 85.3 ≥ 36, modal 90.1 ≥ 41)
- `GOOS=darwin go build ./...`, `GOOS=darwin go vet ./internal/brew/`, `GOOS=windows go vet ./internal/brew/`: pass

**Notes for the next session:**
- `Client.Runner` is now the `sudoRunner` wrapper, not the runner passed to `NewClient`. Nothing in the repo compares it by identity.
- Services writes already had their own needs-sudo text. The wrapper runs first; `serviceWriteError` still turns a sudo refusal into the service-specific message.
- Not tested with real brew (the human owns live writes).
