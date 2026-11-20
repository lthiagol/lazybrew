# Slice: v2-07-op-dismiss — a finished operation gets out of the way

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

User bug (ruling `Q-v2-next` in `work/questions.md`): after `restart
moshi-hook` the main area stayed on `✓ Completed in 0s` until Esc, with no
hint that Esc closes it. The finished Operation view is cleared only by Esc
(`internal/gui/gui.go` `case "esc":` block); `renderContent` shows it while
`m.opState != nil` (`internal/gui/render.go`).

## ⚠ Safety

Never run brew write commands on this machine (`brew install|uninstall|
upgrade|reinstall|tap|untap|trust|untrust|pin|unpin|cleanup|autoremove|update`,
`brew services start|stop|restart|run|cleanup|kill`). `moshi-hook` is a live
service. Tests use the mock runner only.

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes the tests; `TestEveryFrameIsTerminalSize` passes at every commit.

### AC-01 — a successful operation closes itself

1. When a task completes with success, the Operation view stays for **2 s**
   (a named constant) and then the main area returns to the panel's view.
   Use a `tea.Tick` returning a message that names **which** operation it is
   for (e.g. the `*Operation` pointer or a sequence number). The message
   clears `m.opState` only if it is still that same operation and it is done —
   a newer operation, a running one, or one already cleared is left alone.
2. A batch (batch upgrade) must not close between its tasks: while the batch
   still has tasks queued or running, no auto-close fires; the last task's
   success starts the 2 s.
3. No goroutine, no `program.Send`. The `brew-update` path is unchanged.
4. Tests (mock runner, fake time via the message — do not sleep 2 s): success
   → message → main area back to the panel view (assert on `View()`); a stale
   message after a second operation started does nothing; batch does not
   close mid-batch.

### AC-02 — a failed or cancelled operation says how to close it

1. Error and cancelled states stay on screen (no auto-close) and the status
   row ends with ` · Esc to close` (Subtle), like the running row's
   ` · Esc to cancel`. The success row also shows ` · Esc to close` during its
   2 s.
2. While a finished operation is shown, the footer's first hint is
   `esc close`.
3. Tests for the three rows and the footer hint.

### AC-03 — moving away closes a finished operation

1. Switching panel (keys `1`-`7`, Tab/S-Tab, mouse click on a sidebar box)
   while the operation is **done** clears it. A **running** operation is not
   cleared by moving (today's behaviour stays).
2. After a services task, the Services Info view the user lands on shows the
   reloaded state (the v2-05s reload already exists — check it is what is
   drawn once the operation closes).
3. Tests for both.

### AC-04 — evidence and bookkeeping

1. Copy the harness from `work/verification/v2-06-animations/` to
   `work/verification/v2-07-op-dismiss/`, run it (animations off), `after/`
   must equal `v2-06-animations/after/` (`diff -r` empty) — no resting frame
   changes. Add a test-side dump (no brew writes — mock runner) of the
   Operation view at 140×40 for success / error / cancelled into
   `work/verification/v2-07-op-dismiss/op/` (`.txt` + `.ansi`).
2. `result.md`; `work/STATE.md` row
   `| v2-07-op-dismiss | — | implemented | awaiting reviewer |`.

## Invariants

- No change outside the Operation view lifecycle, its status row and the
  footer hint. No restyle, no key change (Esc keeps cancel/close).
- Every frame exactly the terminal size. No test deleted, skipped or
  weakened; changed expectations named. Coverage floors unchanged. No new
  dependencies.
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.

## Gate

`make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.

## Report protocol

`work/inbox/runner-v2-07-op-dismiss.md` (≤ 40 lines), committed once;
worktree clean; reply with exactly one line:
`done: work/inbox/runner-v2-07-op-dismiss.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-v2-07-op-dismiss-rN.md` → fix every finding, one commit each
(subject names the id), gate, receipt
`work/inbox/runner-v2-07-op-dismiss-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
