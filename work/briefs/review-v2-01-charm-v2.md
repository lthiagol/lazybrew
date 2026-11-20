# Review brief — slice `v2-01-charm-v2`

> **Contract.** The slice contract is `work/briefs/v2-01-charm-v2.md`. This
> brief tells you *how to verify* it. The contract wins where they differ.
> You review; you do **not** edit code, tests, or the runner's commits.

You will be asked to run review round **rN** (N = 1..5). Each round reviews
everything on `wip` since `5f9c31e` (the last commit before the slice brief),
including fix commits from earlier rounds. In rounds ≥ 2, first confirm each
earlier finding is actually fixed, then look for new problems the fixes
introduced.

**A receipt is a claim, not evidence.** Run every check yourself.

## Verify by command

### V1 — the gate

```bash
make build && make lint && go test ./... -race -count=1 && make cover-check
```

### V2 — v1 is gone (AC-01.1–2)

```bash
grep -n "charmbracelet/bubbletea\|charmbracelet/lipgloss\|charmbracelet/bubbles\|x/exp/teatest " go.mod
grep -rn '"github.com/charmbracelet/\(bubbletea\|lipgloss\|bubbles\)' --include=*.go .
```

Both must print nothing (teatest **v2** is allowed).

### V3 — no test lost or weakened (AC-01.3)

```bash
grep -rho "^func Test\w*" --include=*_test.go . | wc -l    # must be >= 406
git diff 5f9c31e -- '*_test.go' | grep -n "^-.*\(t.Error\|t.Fatal\|Contains\|want\)" | head -80
grep -rn "t.Skip" --include=*_test.go .
```

Read the removed assertion lines: each must have a ported replacement of equal
strength. A deleted or loosened assertion is a **major** finding.

### V4 — behaviour preserved (AC-01.4)

Read the diff of `internal/gui/gui.go`, `mouse.go`, `commands.go`, `render.go`,
`modal/*`. Check in particular:
- the space bar still toggles batch selection (`"space"` vs `" "`);
- every key that used `msg.Type`/`msg.Runes` still maps to the same action
  (esc, enter, tab, shift+tab, arrows, ctrl+c, `[`/`]`, digits);
- mouse click-to-select, wheel scroll and the panel/tab hit-testing still use
  the same coordinates;
- the search textinput and the input modal still receive typing;
- the viewport still scrolls tab content.

### V5 — the View carries the terminal setup (AC-02)

`Model.View()` sets `AltScreen`, `WindowTitle: "lazybrew"`, and mouse mode
only when `cfg.GUI.Mouse`; `cmd/lazybrew/main.go` no longer passes those
options; a test covers mouse on and off.

### V6 — frame equivalence (AC-03) — reproduce it, do not trust result.md

```bash
D=$(mktemp -d); mkdir -p internal/gui/dump
cp work/verification/v2-01-charm-v2/frame_dump_v2_test.go.txt internal/gui/dump/dump_test.go
DUMP_DIR=$D go test ./internal/gui/dump/ -count=1
diff -r --exclude='*.ansi' work/verification/v2-01-charm-v2/baseline "$D"
rm -rf internal/gui/dump
```

The diff must be empty. Also diff the v2 harness against the v1 one
(`frame_dump_test.go.txt`) and confirm it drives the **same** screens with the
**same** keys at the **same** sizes — a harness that was changed to make the
diff pass is a **blocker**. Confirm `git log --stat 5f9c31e.. -- work/verification/v2-01-charm-v2/baseline` is empty
(the baseline was never edited). Remove `internal/gui/dump` when done.

### V7 — invariants

`git diff --stat 5f9c31e..` — no `internal/brew/**`, no `scripts/check-coverage.sh`,
no new non-Charm dependency in `go.mod`, no file moves/renames, no
`program.Send`, one commit per AC in order, worktree clean.

## Severity

- **blocker** — the gate fails, the frames differ, v1 still present, a test
  deleted/weakened, a behaviour regression, a doctored harness/baseline.
- **major** — an invariant broken, an AC not met, a likely runtime regression
  the tests miss.
- **minor** — process or docs issues that do not affect behaviour.

Report only real, actionable findings, each with file:line and the evidence
(command + output, or the code). No style nits.

## Report

1. Write `work/inbox/reviewer-v2-01-charm-v2-rN.md` (≤ 60 lines): verdict
   first line — `verdict: ok` or `verdict: changes-required`; then a table of
   ACs with pass/fail; then findings as `F-rN-01 (severity) — file:line —
   what is wrong — evidence — what the fix must achieve`; then, for rounds ≥ 2,
   the status of every earlier finding.
2. `verdict: ok` only when there are **no** findings of any severity left.
3. Commit it: `git add work/inbox/reviewer-v2-01-charm-v2-rN.md && git commit -m "inbox: reviewer reports v2-01-charm-v2 rN"`.
4. `git status` clean (remove `internal/gui/dump`). Then reply in your session
   with one line: `done: work/inbox/reviewer-v2-01-charm-v2-rN.md`. Do **not**
   send keys or text to any other pane.
