# Review brief — slice `v2-02-layout-bugs`

> **Contract.** The slice contract is `work/briefs/v2-02-layout-bugs.md`. This
> brief tells you *how to verify* it. The contract wins where they differ.
> You review; you do **not** edit code, tests, or the runner's commits.

You will be asked to run review round **rN** (N = 1..5). Each round reviews
everything on `wip` after the commit that added this brief (find it with
`git log --oneline -- work/briefs/review-v2-02-layout-bugs.md`), including fix
commits from earlier rounds. In rounds ≥ 2, first confirm each earlier finding
is actually fixed, then look for new problems the fixes introduced.

**A receipt is a claim, not evidence.** Run every check yourself. On this
machine `go env CGO_ENABLED` is 0: run the race gate as
`CGO_ENABLED=1 go test ./... -race -count=1`.

## Verify

### V1 — the gate
`make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`

### V2 — frame size (AC-01)
Read the size test: does it really cover every state the contract lists, at
all three sizes, and assert **exact** height and **exact** per-row width?
Then reproduce on real frames: port/run the committed harness
(`work/verification/v2-02-layout-bugs/frame_dump_test.go.txt` →
`internal/gui/dump/dump_test.go`, `DUMP_DIR=$D go test ./internal/gui/dump/`)
and check every `.txt` frame: `awk 'END{print NR}'` = height and every line's
display width = width (use a small Go or Python check that counts cells, not
bytes). Diff your frames against `after/` — they must match.

### V3 — overlays (AC-02, AC-04)
Read the compositing code. Modal and toast must be layers over a full-size
base frame (Lip Gloss v2 compositor), not string joins. Confirm the tests
compare the base rows outside the overlay, and that a toast does not move any
row. Check the modal is centred both ways at 80×24 and 200×60.

### V4 — first-frame modals (AC-03)
No `Loading...` on first render for confirm, input and menu; resize while
open still works. Look for any remaining code path where a modal is created
without a size.

### V5 — not loaded vs empty (AC-05)
Explicit state, set on every fetch result (success **and** error). The
R-refresh lazy rule behaves identically — read the diff around
`// never loaded` in `internal/gui/gui.go`.

### V6 — Log sibling (AC-06)
One border around main, one around the Log; `C` collapse returns rows; mouse
hit-tests for tabs/list rows still land on the right element (read
`mouse.go` + tests; reason about the coordinates).

### V7 — Deps (AC-07)
`brew info --json=v2 bash` on this machine vs. the fixture shape. `bash` →
`ncurses`, `readline` in the test **and** in `after/140x40/11-deps-bash.txt`.
Not-installed fallback covered. Cask parsing unchanged.

### V8 — every frame change is explained (AC-08)
`diff` each v2-01 baseline frame against `after/`; every changed region must
be explained by B1–B7. An unexplained change, or any restyle (colour, border
style, glyph, header, sidebar structure), is a **major** finding.

### V9 — invariants
Test count ≥ 410 and none removed/skipped/weakened (read the removed assertion
lines in `git diff <brief-commit> -- '*_test.go'`); every changed expectation is
justified by a bug fix and named in the receipt. `internal/brew/**` changes only
in AC-07's scope. No new deps, no keybinding change, no `program.Send`, one
commit per AC in order, worktree clean, `internal/gui/dump` absent.

## Severity

- **blocker** — gate fails; a frame overflows/underflows; overlay is still a
  string join; a test deleted/weakened; behaviour regression.
- **major** — an AC not met; unexplained frame change; restyle; invariant broken.
- **minor** — docs/process issues that don't affect behaviour.

Only real, actionable findings, each with file:line and evidence. No style nits.

## Report

1. `work/inbox/reviewer-v2-02-layout-bugs-rN.md` (≤ 60 lines): first line
   `verdict: ok` or `verdict: changes-required`; AC table pass/fail with
   evidence; findings as `F-rN-01 (severity) — file:line — what — evidence —
   what the fix must achieve`; for rounds ≥ 2 the status of every earlier
   finding.
2. `verdict: ok` only with **no** findings of any severity left.
3. Commit it once: `git add work/inbox/reviewer-v2-02-layout-bugs-rN.md && git commit -m "inbox: reviewer reports v2-02-layout-bugs rN"`.
4. Worktree clean, `internal/gui/dump` removed. Reply with exactly one line:
   `done: work/inbox/reviewer-v2-02-layout-bugs-rN.md`. Do not touch other panes.
