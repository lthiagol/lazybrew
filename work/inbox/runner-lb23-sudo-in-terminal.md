# Receipt — lb23-sudo-in-terminal (runner, LB-11)

**Commits, AC order:** AC-01 `a313fa2` · AC-02 `8e9ad25` · AC-03 `4fc15c8`.
Follow-up `a6dab0c` drops two unused assignments staticcheck flagged in AC-03's
tests (no amend).

**AC-01** — `NeedsSudoError.Args` (copy of the args), set by `sudoError`.
`go test ./internal/brew/` ok; args asserted in the Execute and stream sudo tests.

**AC-02** — Enter on a failed op whose error is a sudo refusal for install,
reinstall, upgrade or uninstall runs `exec.Command(BrewPath, args…)` with
`brew.CommandEnv()`, no SysProcAttr, through `m.execProcess` (set in `New`).
Guard `terminalRetry()`: failed op, `batchCount == 0`, new `Manager.Idle()`.
AGENTS.md rule 2 has the exception; `Q-lb23-execprocess` logged.
`go test ./internal/gui/ -run 'TestEnter|TestFailedOverlay' -v`:
- PASS TestEnterRunsTheFailedCommandInTheTerminal
- PASS TestEnterDoesNothingForOtherCommands (update, services, tap, trust, cleanup, autoremove, zap)
- PASS TestEnterRunsTheFailedBatchMemberWithItsArgs
- PASS TestEnterDoesNothingWhileTasksAreQueuedOrABatchRuns
- PASS TestFailedOverlayShowsTheRetryOnlyWhenItRuns (80 and 140 cols; Esc closes)

**AC-03** — the callback maps `*exec.ExitError` to `exited with status N`.
`sudoRetryDone` drops the command's cache group; success goes through
`finishedMutation` (extracted from the TaskCompletedMsg success tail, now shared);
failure toasts `install foo: exited with status 3` and leaves the op as it was.
- PASS TestRetrySuccessFinishesLikeATaskAndDropsItsCache
- PASS TestRetryFailureKeepsTheOperationAndToastsTheExit
- PASS TestRetryCallbackNamesTheExitStatus
- PASS TestSudoRetryRunsInTheTerminalUnderAPty (under `script -qec`, util-linux):
  the fake brew reads `/dev/tty`, writes `tty`, exits 3; asserts marker, toast, opError.
- Mutation: `SysProcAttr{Setsid:true}` in the run gives
  `terminal_retry_test.go:305: brew ran with "no-tty", want a controlling terminal (tty)`;
  reverted, PASS.
- Scope: the pty run calls `cmd.Run` in place of the tea.ExecProcess wrapper (no
  Program runs in a test); that wrapper only sets stdio and calls Run. Skips
  without script(1). The BSD `script` branch (macOS) is not run here.
- Docs: README paragraph under Keybindings. No help page for the overlay exists,
  so none changed. No golden changed.

**Gate** (on `a6dab0c`)
- `make ci`: exit 0. fmt-check, vet, `go test ./... -race -count=1` ok, coverage
  floors met (brew 88.8≥78, gui 90.9≥83, task 96.3≥93), build.
- `GOOS=darwin go build ./...`: ok.
- deadcode `./cmd/...`: no output.
- staticcheck `./...`: 5 findings, all pre-existing (keymap_test.go:39, :431;
  launch_check_test.go:180, :220, :453). None new.

**Choices to check**
- AC-03 says the failure "keeps the Operation closed". Read as: the failed op stays
  as it was (sudo text, Esc closes), per the Behaviour section.
- A failed retry does not refresh the panels, as a failed task does not; its cache
  group is still dropped.
- The hint is the brief's text on the status row only; the footer is unchanged.

**Seen, not done**
- A failed batch member that is not last: the next queued task starts at once and
  replaces the overlay, so Enter cannot reach it. Its toast still names the command.
- DESIGN.md rule 6 (line 83) and the decision row (line 129) still say a sudo step
  fails with a message. Update at LB-11 close.

**macOS manual test** (user, real `.pkg` cask; not run here)
1. `make run` on the Mac; Casks panel, `i` on a cask with a `.pkg` installer.
2. The failed overlay shows `Failed: needs sudo …` and `Enter  run in this terminal`.
3. Enter: the TUI suspends; brew's output and sudo's password prompt appear.
4. Success: toast `<name> completed`, overlay closes after 2 s, cask shows installed.
5. Wrong password or Ctrl-C: toast `…: exited with status N`; Esc closes the overlay.
