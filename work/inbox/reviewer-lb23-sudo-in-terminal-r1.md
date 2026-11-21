verdict: changes-required

# Review: lb23-sudo-in-terminal, r1 (a313fa2..a24c084)

| AC | Commit | Verdict |
|---|---|---|
| AC-01 Args on NeedsSudoError | a313fa2 | ok (`slices.Clone(args)`, tests assert args) |
| AC-02 Enter → execProcess, guard, rule 2 | 8e9ad25 | changes-required (F-r1-01, F-r1-03) |
| AC-03 result, refresh, docs, receipt | 4fc15c8, a6dab0c, a24c084 | ok except the receipt cap (F-r1-02) |

**V1:** `make ci` exit 0, all floors met (brew 88.8, gui 90.9, task 96.3); `GOOS=darwin go build ./...` ok.
**V2:** mutations in my own worktree. Dropping the allow-list check makes TestEnterDoesNothingForOtherCommands and
TestFailedOverlayShows… fail. Dropping `!m.tasks.Idle()` or `m.batchCount > 0` makes TestEnterDoesNothingWhile… fail. Zap: F-r1-01.
**V3:** `exec.Command(Runner.BrewPath(), Args...)`, `Env = brew.CommandEnv()` (the same func `newCmd` uses),
no SysProcAttr. The odd-name harness (`"my app;rm -rf x" "$(id)"`) gets argv as-is, with no shell.
**V4:** own harness under `script -qec`. A fake brew without a tty prints sudo's words. Through the client's
runner it fails fast as `*brew.NeedsSudoError` (~0.5 ms; Setsid still in `setProcessGroup`), and the hint shows.
Enter path, real cmd.Run in the pty: the answer typed on stdin reached the script through /dev/tty
(`secret0`/`secret3`). Exit 0 → `install foo completed`, opSuccess, 1 RefreshMsg. Exit 3 →
`install foo: exited with status 3`, opError, no refresh. Not covered: tea.ExecProcess's own release/restore
of the terminal (no Program in a test), so "TUI resumes" is left to the macOS manual test.
**V5:** 80 cols `✗ Failed: needs sudo — run "brew ins… · Enter  run in this terminal · Esc  close`; 140 cols shows it whole.
Other frames: the only render change is the `opError` branch gated by `terminalRetry()`, and every
golden/teatest passes unchanged. I did not run a per-panel base-vs-HEAD View diff in both themes. The hint uses
`style.SubtleText`, like the close hint.
**V6:** rule 2 and Q-lb23-execprocess are written; no `program.Send` and no `go func` in the diff; Makefile and
scripts are untouched, so floors are unchanged; one commit per AC plus a staticcheck follow-up (a6dab0c).
Q-lb23-execprocess is false about zap (F-r1-01).
**macOS manual test section:** present in the receipt (5 lines, ≤ 8). ok.

## Findings

F-r1-01 (medium) — internal/gui/terminal_retry.go:17, terminal_retry_test.go:97, work/questions.md
(Q-lb23-execprocess) — a cask zap runs in the terminal, though the test and the decision log say it never does.
Evidence: `CaskWriter` zap runs `uninstall --zap --cask <name>` (internal/brew/casks.go:232), so `Args[0]` is
`uninstall`. Harness `sudoFailedModel("uninstall","--zap","--cask","foo")` + Enter → 1 command run. The test
case `{"zap","--cask","foo"}` is args brew never gets, so it passes vacuously. Fix must achieve: decide zap on
purpose. Recommended: exclude `--zap`, because the brief lists four commands and zap deletes user data. Test it
with the real args `uninstall --zap --cask foo`; a revert must fail that test. Q-lb23-execprocess and the
receipt must say what the code does. (If zap is allowed instead: drop the fake case, add a positive test with the
real args, and correct Q-lb23-execprocess.)

F-r1-02 (low) — work/inbox/runner-lb23-sudo-in-terminal.md — the receipt is 66 lines; the cap is 40
(brief "Done means", cr-common). Evidence: `wc -l` → 66. Fix must achieve: the fix-round receipt keeps to its
cap (30). The r1 receipt stays as committed.

F-r1-03 (low) — internal/gui/terminal_retry.go:24 — the guard's "Operation still in progress" leg is not pinned.
Evidence: deleting `|| m.opState.Status != opError` leaves `go test ./internal/gui/` green. Fix must achieve: a
test where the op is not opError (e.g. opRunning or opCancelled while it carries a NeedsSudoError) and Enter
runs nothing and the hint is absent; removing the status check must make it fail.

Note (no fix asked): the in-repo pty test covers exit 3 only. The brief's "exits 0 or 3" was checked by my own harness.
