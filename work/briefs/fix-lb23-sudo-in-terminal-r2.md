# Fix brief — lb23-sudo-in-terminal, round r2

Findings: `work/inbox/reviewer-lb23-sudo-in-terminal-r1.md` (read in full). Tests
first, one commit per finding, subjects end with the finding id.

- **F-r1-01 (medium):** exclude `--zap` (ruling above). `terminalRetry` refuses
  any args containing `--zap`; replace the vacuous `{"zap","--cask","foo"}` case
  with the real `uninstall --zap --cask foo` (must fail with the exclusion
  reverted), and keep a positive case for plain `uninstall --cask foo`. Correct
  `Q-lb23-execprocess` in `work/questions.md` (append a correction line, do not
  rewrite) and the README sentence if it implies zap.
- **F-r1-02:** the fix-round receipt section stays ≤ 15 lines (append "Fix r2" to
  `work/inbox/runner-lb23-sudo-in-terminal.md`); do not edit the r1 text. Commit,
  do not edit afterwards.
- **F-r1-03:** pin the guard's status leg: a failed-looking op that is not
  `opError` (`opRunning`, `opCancelled`) carrying a `NeedsSudoError` → Enter runs
  nothing and the hint is absent; removing the status check must fail it.

Gate: `make test-fast PKG=./internal/gui/...`, then `make ci`. Never launch the
TUI. No brew write commands. Reply with one done line.
