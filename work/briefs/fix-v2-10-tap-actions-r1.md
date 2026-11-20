# Fix brief — v2-10-tap-actions, round r1

Source: `work/inbox/reviewer-v2-10-tap-actions-r1.md`. One commit (subject
names the id, Co-Authored-By trailer), receipt
`work/inbox/runner-v2-10-tap-actions-fix-r1.md` (≤ 30 lines, committed once).
Gate per `Q-v2-builder-speed` (+ amendment).

- **F-r1-01 (minor)** — `internal/gui/details.go:116`: the Trust tab still
  says `Press t to trust or untrust the tap, or one of its formulae or
  casks.` Make it match AC-02: `t` trusts/untrusts the tap, `T` trusts one of
  its formulae or casks; an official tap's Trust tab offers neither key. Add
  a `View()` assert. Regenerate frames if any change; name them.

Reply with exactly one line: `done: work/inbox/runner-v2-10-tap-actions-fix-r1.md`.
