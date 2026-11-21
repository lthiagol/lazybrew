# Review brief — slice `lb23-sudo-in-terminal`

> Contract: `work/briefs/lb23-sudo-in-terminal.md` (with `cr-common.md`). You
> review; you do **not** edit code, tests or commits. Review everything on `wip`
> after the commit that added the slice brief
> (`git log --oneline -- work/briefs/lb23-sudo-in-terminal.md`) up to HEAD. A
> receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate. No brew
> write commands; never launch the TUI against the real brew; scratch under
> /home/agent/.cache; do not touch other panes.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 args and guard:** the `NeedsSudoError` carries args for install, reinstall,
  upgrade, uninstall (formula, `--cask`, batch member); every other command
  (update, services, tap, trust, cleanup, autoremove) gets no Enter action. The
  guard: Enter does nothing while a task runs or is queued or an Operation is in
  progress. Revert the allow-list or the guard: a test must fail.
- **V3 exec command:** path is the runner's `BrewPath()`; env is exactly the
  runner's (`HOMEBREW_NO_AUTO_UPDATE=1`, `HOMEBREW_NO_ASK=1`, inherited);
  **no** `SysProcAttr` (no Setsid/Setpgid), so a child can read the tty.
  Check a command with spaces or odd characters in a name is passed as args,
  never through a shell.
- **V4 real terminal (own harness):** a fake `brew` script in a temp dir that
  prints sudo's "a terminal is required" under the normal runner (reproducing
  the failed Operation), and reads `/dev/tty` when run through the Enter path.
  Under a pty (`script(1)` or Go `os/exec` + a pty if available without a new
  dependency): the Enter path gives the script the tty and the typed answer
  reaches it; the TUI resumes after; exit 0 → success toast + refresh; exit 3 →
  failure toast. Also confirm the normal runner path is unchanged (still
  Setsid, still fails fast).
- **V5 UI:** overlay hint line shows `Enter run in this terminal` only for the
  allowed sudo failure, at 80x24 and 140x40, both themes; every other frame is
  unchanged (own `View()` diff base vs HEAD for all panels and a failed
  non-sudo Operation).
- **V6 rules:** AGENTS.md rule 2 and `Q-lb23-execprocess` are written and say the
  true thing; no `program.Send`; no raw goroutines in handlers; one commit per
  AC; floors not lowered.

## Report
`work/inbox/reviewer-lb23-sudo-in-terminal-rN.md` (≤ 55 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table; findings `F-rN-01
(severity) — file:line — what — evidence — fix must achieve`. Commit once
(`inbox: reviewer reports lb23-sudo-in-terminal rN`), worktree clean, reply with
one line: `done: work/inbox/reviewer-lb23-sudo-in-terminal-rN.md`.
