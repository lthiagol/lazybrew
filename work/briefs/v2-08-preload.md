# Slice: v2-08-preload — every panel loads at startup, in the background

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

User request (ruling `Q-v2-next` in `work/questions.md`): show everything
we can on startup instead of loading a panel only when the user first visits
it. Today `Init` (`internal/gui/gui.go`) fetches only the active panel and
`lazyLoadPanel` fetches the rest on first switch; boxes show `not loaded`
(`internal/gui/render.go`). Measured on this machine (read-only): Formulae
0.49 s, Casks 0.25 s, Outdated 0.37 s, Taps 0.87 s, Services list 0.39 s +
info 0.38 s — ~2.8 s serial.

## ⚠ Safety

Never run brew write commands on this machine (`brew install|uninstall|
upgrade|reinstall|tap|untap|trust|untrust|pin|unpin|cleanup|autoremove|update`,
`brew services start|stop|restart|run|cleanup|kill`). `moshi-hook` is a live
service. Tests use the mock runner only.

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes the tests; `TestEveryFrameIsTerminalSize` passes at every commit.

### AC-01 — a preload scheduler

1. Config `gui.preload` (bool, default `true`), env `LAZYBREW_PRELOAD=0|1`
   overrides it (same precedence as `LAZYBREW_ICONS`). README and DESIGN
   config table document it.
2. With preload on, after `Init` fetches Status (or runs `update_on_start` —
   preload starts once that update finishes, so it reads fresh data), the
   other panels load in this order: Formulae → Outdated → Casks → Taps →
   Services (list and info). Search is never preloaded.
3. **At most 2 reads in flight** at once, driven by messages in the Update
   loop (a queue on the model; each finished fetch starts the next). Every
   read goes through the existing `fetchPanelData` / `boundedRead` path with
   `brewReadTimeout`. No goroutine, no `program.Send`.
4. A panel the user opens while it is queued or in flight is fetched once,
   not twice (`lazyLoadPanel` and the scheduler share the `loading` flag). A
   panel already loaded is skipped. `R` refresh keeps working as today.
5. With preload off, behaviour is exactly today's lazy loading.
6. Test helpers and the frame harness set preload explicitly (state which and
   why in `result.md`); a test asserts the production default is on.
7. Unit tests: order, the 2-in-flight cap, dedupe with a user switch, a
   failed read, preload off, `update_on_start`.

### AC-02 — the app stays responsive

1. Keys and mouse work while panels preload (a test presses keys mid-preload
   and asserts on `View()`).
2. A failed read shows its error in that box only; the other panels still
   load.
3. The dashboard (Status) counts fill in as panels arrive (Outdated count no
   longer needs `R` once preloaded; keep `not loaded — press R` only when
   preload is off).

### AC-03 — the sidebar says what is happening

1. A box whose read is queued or in flight shows a loading state in the
   border's right label (spinner or `loading…`, Subtle) instead of
   `not loaded`; once loaded it shows its count as today.
2. Frames: `work/verification/v2-08-preload/` harness run (real brew,
   read-only) with preload **on** and the harness waiting until all panels
   have loaded, so the sidebar boxes show their rows in every screen; note in
   `result.md` every screen whose sidebar changed and why. `before/` =
   `v2-07-op-dismiss/after/`.
3. `result.md`; `work/STATE.md` row
   `| v2-08-preload | — | implemented | awaiting reviewer |`.

## Invariants

- No new brew command; only the reads each panel already uses.
- No main-area content change, no restyle, no key change.
- Every frame exactly the terminal size. No test deleted, skipped or
  weakened; changed expectations named. Coverage floors unchanged. No new
  dependencies.
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.

## Gate

`make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.

## Report protocol

`work/inbox/runner-v2-08-preload.md` (≤ 40 lines), committed once;
worktree clean; reply with exactly one line:
`done: work/inbox/runner-v2-08-preload.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-v2-08-preload-rN.md` → fix every finding, one commit each
(subject names the id), gate, receipt
`work/inbox/runner-v2-08-preload-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
