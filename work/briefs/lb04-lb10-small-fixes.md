# Slice: lb04-lb10-small-fixes — chunked tap-info; the confirm dialog's cut row

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

Two parked backlog items, ruling `Q-backlog-next` in `work/questions.md`:

- **LB-04.** `fetchTapInfoBatch` (`internal/brew/taps.go` ~:171) passes every
  tap name to one `brew tap-info --json name1 name2 …`. With enough taps the
  argument list can exceed the OS limit (ARG_MAX); the call then fails and the
  Taps panel silently falls back to the name-only `TrustUnknown` list.
- **LB-10.** When a confirm dialog's body is cut (`dialog` in
  `internal/gui/modal/frame.go` ~:82, more than `dialogMaxRows - 8` body
  rows — e.g. a long dependents list), the closing `…` row is centred on its
  own instead of sitting in the list's column, so it floats in the middle of
  the box under a left-aligned list.

## ⚠ Safety

Never run brew write commands on this machine (`brew install|uninstall|
upgrade|reinstall|tap|untap|trust|untrust|pin|unpin|cleanup|autoremove|update`,
`brew services start|stop|restart|run|cleanup|kill`). `moshi-hook` is a live
service. Tests use the mock runner only.

The frame harness may press `t`/`x` **only** to open a confirm or menu and
must then press `esc`; never `y` or `enter` on a confirm or menu, and never
submit the add-tap input.

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes `make test-fast PKG=<touched packages>`; the full gate (with `-race`)
runs once, before the receipt (ruling `Q-v2-builder-speed`).

### AC-01 — tap-info is chunked (LB-04)

1. `fetchTapInfoBatch` splits the names into chunks bounded by **both** a name
   count (a named constant, e.g. 100) and a total argument byte budget (a
   named constant well under the smallest real limit, e.g. 64 KiB), runs one
   `tap-info --json` per chunk in order under the same `ctx`, and returns the
   concatenated entries in input order. A typical install (well under the
   bounds) still makes exactly **one** call — same args as today.
2. Failure semantics unchanged: if any chunk fails, the whole batch returns
   the error, so `List` keeps its existing name-only `TrustUnknown`, not-cached
   fallback. A cancelled/timed-out ctx stops before the next chunk.
3. Unit tests with the mock runner: (a) few taps → one call, args identical to
   today; (b) more names than the count bound → several calls, each within
   both bounds, every name sent exactly once, result merged in order;
   (c) long names → split by the byte budget; (d) the second chunk fails →
   error, and `List` returns `TrustUnknown` and does not cache. Make the
   bounds overridable by tests (package variables, like `brewReadTimeout`),
   production values unchanged by tests.

### AC-02 — the cut row sits in the list column (LB-10)

1. When `dialog` cuts the body, the `…` row starts in the same column as the
   body line it replaces (it keeps that line's indent and is padded to its
   width, so per-row centring lands it under the list), not centred alone. An
   uncut body renders byte-identical to today.
2. Test in `internal/gui/modal`: a confirm with a dependents list long enough
   to be cut — assert on the rendered string that the `…` row's first
   non-space column equals the list rows' first non-space column, and the box
   still fits `dialogMaxRows` and is exactly its width. Existing
   `confirm_face_test.go` expectations unchanged (name any that must change
   and why).
3. Capture a frame (the existing frame harness or a test golden) of the
   uninstall confirm with a long dependents list at 80x24 and 140x40, saved
   under `work/verification/lb04-lb10-small-fixes/`; open it only to capture,
   then `esc`.

### AC-03 — bookkeeping

1. `work/verification/lb04-lb10-small-fixes/result.md`: the tests added, what
   each proves, the gate output summary.
2. `work/STATE.md` row
   `| lb04-lb10-small-fixes | — | implemented | awaiting reviewer |`.

## Invariants

- No other behaviour or frame change. No new dependencies. Coverage floors
  unchanged. Every frame exactly the terminal size.
- No test deleted, skipped or weakened; changed expectations named.
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.

## Gate

`make ci` (gofmt check, vet, `go test ./... -race -count=1` with
`CGO_ENABLED=1`, `make cover-check`, build). Also `GOOS=darwin go build ./...`.

## Report protocol

`work/inbox/runner-lb04-lb10-small-fixes.md` (≤ 40 lines), committed once;
worktree clean; reply with exactly one line:
`done: work/inbox/runner-lb04-lb10-small-fixes.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-lb04-lb10-small-fixes-rN.md` → fix every finding, one commit
each (subject names the id), gate, receipt
`work/inbox/runner-lb04-lb10-small-fixes-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
