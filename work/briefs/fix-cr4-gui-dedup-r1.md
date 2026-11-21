# Fix brief — cr4-gui-dedup, round r1

Source: `work/inbox/reviewer-cr4-gui-dedup-r1.md`. Fix every finding, one
commit each, subject names the id. Briefs `work/briefs/cr4-gui-dedup.md` and
`work/briefs/cr-common.md` still apply. The eight AC commits are already on
`wip`; do not redo them. Goldens, snapshots and teatest flows stay unchanged.

- **F-r1-01 (major)** — `internal/gui/task_build.go:84-88`. When ctx is
  cancelled, `linesTask` returns without sending on `errCh`, and
  `task/manager.go:186-188` (`rt.err = <-rt.errCh` in `readChunk`) then blocks
  forever: no `TaskCompletedMsg`, the overlay stays "running", the queue stalls.
  `runDoctor`/`runMissing` had this at the base commit; `runVulns` did not
  (its buffered send was unguarded and `errCh` was always nil) and now can,
  because the `select` may pick `ctx.Done`. `TestLinesTaskCancelMidStream`
  asserts "errCh delivers nothing", which encodes the hang.
  Fix: a cancelled `linesTask` always ends with an error on `errCh`
  (`ctx.Err()`), and every other exit path of every helper in `task_build.go`
  (`errTask`, `streamTask`, `linesTask`) also delivers exactly one value on
  `errCh`, so `readChunk` can never block on it. Check each exit path and say
  what you found. Tests, written first, failing now: change
  `TestLinesTaskCancelMidStream` to expect the error, and add a test through
  `task.Manager` for each of doctor, missing and vulns shaped like the real
  tasks: enqueue a 100-line `linesTask`, `RunNext`, read one chunk,
  `CancelCurrent` → a `TaskCompletedMsg` (cancelled) arrives within 2 s, under
  `-race`, repeated a few hundred times to catch the select race.
- **F-r1-02 (major)** — `internal/gui/keymap_test.go:80-104,129-158`. The
  keymap test does not fail when a handled key is missing from its footer:
  dropping `stDoctor` from `panelHints(PanelStatus)` still passes (`d` is
  dispatched on Status). The check is a package-wide union of strings, the
  footers are only checked as a subset of the table, and `keyTable` is listed by
  hand in the test. Fix: per context (each panel, global, each modal), every key
  dispatched there appears in that context's footer — or in help where the brief
  allows it (`cr4-gui-dedup.md` AC-05.2) — and every footer/help key is
  dispatched there. Derive the lists from the real tables and handlers where
  you can; do not hand-list them in the test. Mutations that must fail, shown
  in the receipt: drop `stDoctor` from the Status footer; drop a key from the
  Formulae footer; drop a key from a modal's `Hints()`; drop a key from help.
  Restore each afterwards.

Not a finding, do not touch: the remaining raw `ansi.Truncate` calls
(`panel.go:375` uses "..." and is out of R8; `presentation/detail.go:136`), the
read cmds that capture `m`, and `render.go:486-488`.

Gate: `make test-fast PKG=./internal/gui/...` per commit, then `make ci`.
Receipt `work/inbox/runner-cr4-gui-dedup-fix-r1.md` (at most 30 lines,
committed once), worktree clean, reply with exactly one line:
`done: work/inbox/runner-cr4-gui-dedup-fix-r1.md`.
