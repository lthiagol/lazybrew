# Review brief — slice `v2-03-look-foundation`

> The slice contract is `work/briefs/v2-03-look-foundation.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-v2-03-look-foundation.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits that touch only
`work/questions.md`, `work/backlog.md`, `work/STATE.md` or `work/briefs/` are
not findings.

**A receipt is a claim, not evidence.** `CGO_ENABLED=1` for the race gate.
Never run brew write commands.

## Verify

- **V1 gate:** `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`.
- **V2 frames:** run the committed harness (`frame_dump_test.go.txt` →
  `internal/gui/dump/dump_test.go`, real Homebrew), diff against `after/`
  (must match), check every frame is exactly W×H cells (count cells, not
  bytes). Remove `internal/gui/dump` after.
- **V3 palette (AC-01):** official Catppuccin Mocha/Latte hex values (check a
  sample against the published palette you know: Mocha base `#1e1e2e`, text
  `#cdd6f4`, mauve `#cba6f7`; Latte base `#eff1f5`, text `#4c4f69`, mauve
  `#8839ef`); roles used instead of raw hex across `internal/gui/**`
  (`grep -rn '#[0-9a-fA-F]\{6\}' internal/gui --include=*.go | grep -v _test | grep -v style/`
  should be empty); the BgColor hack gone; View sets bg/fg.
- **V4 icons (AC-02):** config default/override/off; env overrides config;
  `show_icons: false` leaves no glyph and no stray leading space (look at a
  frame/test); PUA test is real; README documents both keys.
- **V5 panel primitive (AC-03):** one helper used by sidebar, main and Log;
  title in the top border; right label states (count / not loaded / loading /
  error) visible on collapsed boxes at 80×24 in `after/80x24`; no line
  overflows its box; mouse hit tests updated and still meaningful.
- **V6 header (AC-04):** unknown values omitted (never `0`); never wraps at
  80×24; running operation appears in the middle.
- **V7 footer (AC-05):** exactly one row; `? more` iff cut; modal/help
  footers; `panelHints` still single source of truth.
- **V8 every frame change explained (AC-06):** `diff before/ after/` per
  screen vs. `result.md`. Any content restructure (that is v2-03b) or
  modal/toast restyle beyond palette (v2-04) is a **major** finding.
- **V9 invariants:** test count ≥ start count; removed assertion lines in
  `git diff <brief-commit> -- '*_test.go'` each have a replacement of equal
  strength and are named in the receipt; no keybinding change, no
  `internal/brew/**` change, no new deps, one commit per AC, worktree clean.

Also look with your own eyes at `after/140x40/*.ansi` (`cat` them in your
terminal): flag anything that reads badly — clashing colours, unreadable
contrast, misaligned borders — as a **minor** finding with the frame name.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened, regression) ·
major (AC not met, scope breach, unexplained change) · minor (docs/process,
visual polish issues).

## Report

`work/inbox/reviewer-v2-03-look-foundation-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of every earlier finding. `ok` only with no
findings left. Commit once
(`git add <file> && git commit -m "inbox: reviewer reports v2-03-look-foundation rN"`),
worktree clean, reply with exactly one line:
`done: work/inbox/reviewer-v2-03-look-foundation-rN.md`. Do not touch other panes.
