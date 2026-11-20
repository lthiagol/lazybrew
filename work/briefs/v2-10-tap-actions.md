# Slice: v2-10-tap-actions — tap, untap, trust and untrust as clear, safe actions

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

User request (ruling `Q-v2-next` in `work/questions.md`): give taps what
v2-05s gave services. The actions already exist (`internal/gui/gui.go` key
block: `a` add → input, `x` untap → confirm, `r` repair → confirm, `t` →
trust menu; `executeTrustAction` / `executeTrustMenuAction` in
`internal/gui/commands.go`; `TrustWriter.TrustTap/UntrustTap`,
`TapsWriter.Tap/TapWithURL/Untap/Repair` in `internal/brew`), but untrust is
hidden in a menu and nothing shows which actions apply. Read
`work/briefs/v2-05s-services.md` and its code — follow the same patterns
(state-aware action line, footer, guard toasts, confirm kinds, reload by
name).

## ⚠ Safety

Never run brew write commands on this machine (`brew install|uninstall|
upgrade|reinstall|tap|untap|trust|untrust|pin|unpin|cleanup|autoremove|update`,
`brew services start|stop|restart|run|cleanup|kill`). `moshi-hook` is a live
service. Tests use the mock runner only.

The frame harness may press `t`/`x` **only** to open a confirm or menu and
must then press `esc`; never `y` or `enter` on a tap confirm or menu, and
never submit the add-tap input.

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes the tests; `TestEveryFrameIsTerminalSize` passes at every commit.

### AC-01 — Tap Info shows the actions that apply now

1. Below the Tap Info fields, a key line of the actions that apply to the
   selected tap: untrusted → `t trust  x untap  r repair`; trusted (not
   official) → `t untrust  x untap  r repair`; official → `x untap  r repair`
   (no trust action — official taps do not need it). `a add tap` is always
   in the footer for the panel. The same state-aware set leads the footer.

### AC-02 — `t` toggles trust

1. `t` on an untrusted tap trusts it at once (no confirm — like services
   `s`). `t` on a trusted tap opens a confirm: upgrade kind, chip
   `UNTRUST`, consequence `Homebrew will refuse to load its formulae and
   casks until you trust it again`. `t` on an official tap shows an info
   toast (`homebrew/core is official — no trust needed`) and enqueues
   nothing.
2. Trusting a single formula or cask from a tap stays reachable: `T` opens
   the existing menu (formula / cask submenus). Update help and
   `keybindings.go`; this is the one key change allowed.

### AC-03 — untap says what it will do

1. The untap confirm (danger kind, chip `UNTAP`) states the consequence:
   `the tap's formulae and casks are no longer available`; if installed
   packages come from that tap (from the Formulae/Casks lists already loaded:
   full name prefix `user/repo/`), it lists them and says brew will refuse
   to untap while they are installed. No new brew read.
2. Errors from untap/trust/untrust keep the existing error toast, readable
   (brew's first meaningful line).

### AC-04 — after the action

1. After any tap task completes (success or error), the Taps list and trust
   state reload, and the selection stays on the **same tap by name** (after
   an untap: the next tap in the list).
2. Tests with the mock runner for every state × action, the refused
   official-trust, the confirms, reload-by-name.

### AC-05 — evidence and bookkeeping

1. Harness copied to `work/verification/v2-10-tap-actions/`; same screens
   plus `21-tap-info-untrusted`, `22-tap-info-trusted`,
   `23-tap-untrust-confirm` (press `t` on a trusted tap, capture, **`esc`**),
   `24-tap-untap-confirm` (press `x`, capture, **`esc`**). `before/` =
   `v2-09-main-layout/after/`; `.ansi` at 140×40.
2. `result.md`; `work/STATE.md` row
   `| v2-10-tap-actions | — | implemented | awaiting reviewer |`.

## Invariants

- No change to how other panels work. Only key change: `T` for the per-item
  trust menu (AC-02.2).
- Every write through `task.Manager`; no new brew command beyond those
  already wired.
- Every frame exactly the terminal size. No test deleted, skipped or
  weakened; changed expectations named. Coverage floors unchanged. No new
  dependencies.
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.

## Gate

`make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.

## Report protocol

`work/inbox/runner-v2-10-tap-actions.md` (≤ 40 lines), committed once;
worktree clean; reply with exactly one line:
`done: work/inbox/runner-v2-10-tap-actions.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-v2-10-tap-actions-rN.md` → fix every finding, one commit each
(subject names the id), gate, receipt
`work/inbox/runner-v2-10-tap-actions-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
