# Fix brief — v2-11-cancel-and-test-speed, round r1 (orchestrator ruling)

Source: ruling on `Q-v2-11-b` in `work/questions.md`. One commit per item
(subject names the id), receipt
`work/inbox/runner-v2-11-cancel-and-test-speed-fix-r1.md` (≤ 30 lines,
committed once). Gate per `Q-v2-builder-speed` (+ amendment). Never run real
brew; fake brew scripts only. Never run sudo.

- **Q-v2-11-b-1** — both runner paths start brew with `Setsid: true` instead
  of `Setpgid` (a new session: no controlling terminal, still its own process
  group — the cancel keeps killing `-pgid`). Keep `WaitDelay`. Keep the darwin
  build. Tests: a fake brew that does `read x </dev/tty` (or `sudo -n`-like
  behaviour via `tty` check) **fails fast** instead of stopping — run it under
  a pty if the test needs one, as the reviewer did — and the existing
  grandchild / setsid-holder cancel tests still pass.
- **Q-v2-11-b-2** — when a brew task fails because it needed a terminal for
  sudo (brew/sudo error text such as `a terminal is required`,
  `no tty present`, `sudo: a password is required`), the error toast says:
  `<title>: needs sudo — run "brew <args>" in a terminal` (reuse the style of
  the services needs-sudo text in `internal/brew/errors.go`). Other failures
  unchanged. Tests with the mock runner.

Reply with exactly one line:
`done: work/inbox/runner-v2-11-cancel-and-test-speed-fix-r1.md`.
