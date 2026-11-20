# Slice: v2-01-charm-v2 — move the TUI to the Charm v2 stack, no visual change

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

lazybrew is on Bubble Tea **v1** (`github.com/charmbracelet/bubbletea v1.3.10`,
`lipgloss v1.1.0`, `bubbles v1.0.0`, `x/exp/teatest` v1). The visual overhaul
that follows this slice (title-in-border panels, header bar, real overlay
modals and toasts via lipgloss v2 layers/compositing, harmonica animations)
is built on the **v2** stack. This slice moves the code to v2 and changes
**nothing the user can see**. It is a port, not a redesign.

Target modules (latest at the time of writing — use these exact versions):

| v1 | v2 |
|---|---|
| `github.com/charmbracelet/bubbletea` | `charm.land/bubbletea/v2 v2.1.2` |
| `github.com/charmbracelet/lipgloss` | `charm.land/lipgloss/v2 v2.1.1` |
| `github.com/charmbracelet/bubbles` | `charm.land/bubbles/v2 v2.2.1` |
| `github.com/charmbracelet/x/exp/teatest` | `github.com/charmbracelet/x/exp/teatest/v2` (latest pseudo-version, `go get ...@latest`) |

A working v2 TUI on this machine you may read for idioms (read-only, do not
modify): `/home/agent/coding/moises-training-troubleshoot/tools/change-gate-tui/`
(`internal/ui/view.go`, `internal/ui/model.go`). Prefer the module source in
`$(go env GOMODCACHE)/charm.land/...` and its `UPGRADE_GUIDE_V2.md` /
`README.md` over memory — the v2 API changed in many small ways.

## Known v2 differences you will hit (not exhaustive — verify each in the source)

- `View()` returns `tea.View`, not `string`. Alt-screen, mouse mode and window
  title move from `tea.NewProgram` options into the `tea.View` fields
  (`AltScreen`, `MouseMode`, `WindowTitle`). `tea.WithAltScreen()` /
  `tea.WithMouseCellMotion()` no longer exist.
- `tea.KeyMsg` is an interface; key presses are `tea.KeyPressMsg`.
  `tea.KeyRunes` / `msg.Runes` / `msg.Type` are gone (use `msg.Text`,
  `msg.Code`, `msg.String()`). **`msg.String()` for the space bar is
  `"space"`, not `" "`** — `internal/gui/gui.go:824` matches `" "` and must
  keep working (space toggles batch selection in Outdated).
- Mouse: `tea.MouseMsg` is an interface; clicks are `tea.MouseClickMsg`, wheel
  is `tea.MouseWheelMsg`, fields via `msg.Mouse()`. `internal/gui/mouse.go`.
- `lipgloss.Color` is a function returning `color.Color`; style/theme fields
  typed `lipgloss.Color` become `color.Color` (`image/color`).
  `lipgloss.WithWhitespaceBackground` and friends changed shape.
- Lip Gloss v2 measures `Width`/`Height` differently from v1 in some cases
  (borders/padding). The frame-equivalence check (AC-03) is how you find out —
  adjust sizes until the frames match, do not change the layout math's intent.
- bubbles v2: `viewport.New(...)` takes options, `Width`/`Height` are methods
  (`SetWidth`, `SetHeight`); `textinput` `Width` → `SetWidth`; spinner,
  key bindings and styles changed shape. `internal/gui/gui.go:121,488-489`,
  `internal/gui/render.go:205-206`, `internal/gui/modal/input.go:21`.
- Tests: every `tea.KeyMsg{Type: tea.KeyRunes, Runes: ...}` becomes a
  `tea.KeyPressMsg{...}`; `tea.KeyMsg{Type: tea.KeyTab}` etc. become
  `tea.KeyPressMsg{Code: tea.KeyTab}`. A small helper in each test package
  is fine. Tests that read `m.View()` as a string read the view's content.

## Acceptance criteria

One commit per criterion, in order. Stage by name — never `git add -A`, never
`git add .`. **Every commit must build and pass `go test ./...`** — that is why
AC-01 is one atomic port and not a per-package series.

### AC-01 — the code and its tests run on v2, and v1 is gone

1. `go.mod` requires the four v2 modules above; after `go mod tidy`, **no**
   `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`,
   `github.com/charmbracelet/bubbles` or `github.com/charmbracelet/x/exp/teatest`
   (v1) remain in `go.mod` — not even as `// indirect`.
2. `grep -rn '"github.com/charmbracelet/\(bubbletea\|lipgloss\|bubbles\)' --include=*.go .`
   returns nothing.
3. All production code under `cmd/` and `internal/` is ported; all tests are
   ported. **No test is deleted, skipped (`t.Skip`), or weakened** (an
   assertion removed, a wanted string loosened). The repo has **406**
   `func Test...` today; it must have ≥ 406 after this slice:
   `grep -rho "^func Test\w*" --include=*_test.go . | wc -l`.
   If a test genuinely cannot survive the port (it tests a v1-only API), keep
   an equivalent test of the same behaviour and name the substitution in the
   receipt.
4. Behaviour is unchanged: every key binding, mouse action (click to select,
   wheel to scroll, click on panels/tabs), modal, toast, task/operation flow
   and refresh path works exactly as before.

### AC-02 — the program's terminal setup lives in the View

1. `Model.View()` returns a `tea.View` with `AltScreen: true`, `WindowTitle:
   "lazybrew"`, and `MouseMode` set to cell motion **only when**
   `cfg.GUI.Mouse` is true (today's `tea.WithMouseCellMotion()` condition in
   `cmd/lazybrew/main.go:43-45`); mouse off otherwise.
2. `cmd/lazybrew/main.go` no longer passes alt-screen or mouse program
   options.
3. A unit test in `internal/gui` asserts all three fields, mouse on **and**
   mouse off.

(If you found it impossible to do AC-01 without AC-02's View change — the
View type changes in the same port — put the `tea.View` plumbing in AC-01 and
use AC-02 for the `WindowTitle`, the config-driven mouse mode, the `main.go`
cleanup and the test. Say so in the receipt.)

### AC-03 — frame equivalence: the user sees the same screen

Follow `work/verification/v2-01-charm-v2/README.md`:

1. port the harness to v2 (same screens, same key sequence, same sizes);
2. run it and `diff -r` against `baseline/` (ANSI-stripped `.txt` frames);
3. the diff **must be empty**. If it is not, fix the code, not the baseline.
   The baseline is never edited.
4. commit the ported harness as
   `work/verification/v2-01-charm-v2/frame_dump_v2_test.go.txt` (`.txt`, never
   compiled) and a `work/verification/v2-01-charm-v2/result.md` holding the
   exact commands you ran and their real output (the empty diff, the test
   count, the go.mod grep).
5. `internal/gui/dump/` must **not** be committed and must not exist when you
   report.

Note: the baseline intentionally contains today's bugs (the confirm modal
renders `Loading...` below the screen; panels never opened say `(empty)`).
**Do not fix them in this slice** — the next slice does. Preserve them.

### AC-04 — docs and bookkeeping

1. `DESIGN.md`: the stack line names the v2 modules; add a decision-log row
   dated 2026-11-19: *Moved to the Charm v2 stack (charm.land bubbletea/lipgloss/bubbles v2)
   — prerequisite for overlay compositing and animation in the visual overhaul.*
2. `AGENTS.md`: any v1-specific statement (e.g. `program.Send`, teatest import
   path) updated if it names a v1 API; leave the rules themselves alone.
3. `work/STATE.md`: add a row
   `| v2-01-charm-v2 | — | implemented | awaiting reviewer |`.

## Invariants

- **No visual change.** AC-03's diff is the proof.
- No change to `internal/brew/**` (it does not import Charm; if it does, stop
  and report).
- No change to any coverage floor in `scripts/check-coverage.sh`; `make
  cover-check` must pass.
- Bubble Tea rules in `AGENTS.md` hold: no `program.Send`, all writes through
  `task.Manager`, no raw goroutines in handlers.
- No new dependencies beyond the four v2 modules and their transitive deps.
- No refactors beyond what the port requires. Do not rename, move, or
  restructure files. Do not fix the known visual bugs.
- One commit per acceptance criterion, in AC order, staged by name. Do not
  amend, rebase, revert, or push.
- `master-plan/` is a frozen local archive — do not read or modify it.
- The worktree must be clean when you report.

## Gate

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass, with real output recorded in the receipt.

## Report protocol

1. Write `work/inbox/runner-v2-01-charm-v2.md` — at most 40 lines: the commit
   ids in AC order, each AC answered with its evidence (command + output), the
   gate result, the test count before/after, any substitution or contradiction.
2. Commit it: `git add work/inbox/runner-v2-01-charm-v2.md && git commit -m "inbox: runner reports v2-01-charm-v2"`.
3. `git status` must be clean. Then stop and reply in your session with one
   line: `done: work/inbox/runner-v2-01-charm-v2.md`. The orchestrator is
   watching your session; **do not** send keys or text to any other pane.

## Fix rounds

The reviewer may return findings. A fix round has its own brief,
`work/briefs/fix-v2-01-charm-v2-rN.md`. **Every** finding in it must be fixed
(one commit per finding, `fix(...)` subject naming the finding id), with the
gate passing again and a receipt at `work/inbox/runner-v2-01-charm-v2-fix-rN.md`.

## Questions

If something here is wrong or ambiguous, append it to `work/questions.md`
(append, never rewrite) and keep going on the criteria it does not affect.
