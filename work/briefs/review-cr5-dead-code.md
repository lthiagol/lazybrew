# Review brief — slice `cr5-dead-code`

> The slice contract is `work/briefs/cr5-dead-code.md` (with
> `work/briefs/cr-common.md`); it wins where they differ. You review; you do
> **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-cr5-dead-code.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands — do not start/stop/restart moshi-hook.

## Verify

- **V1 gate:** `make ci`; also `GOOS=darwin go build ./...`.
- **V2 per AC (AC-01 … AC-07):** each finding in the brief is really fixed in
  the code, not just in a test. For every regression test, revert the fix
  (`git stash`-free: check out the pre-fix file into a temp worktree) and
  confirm the test **fails** there; a test that passes without the fix is a
  finding.
- **V3 tools (the main check):** run both installed binaries yourself, never
  `go run`/`go install` (denied here):
  `T=/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin`
  then `$T/deadcode ./cmd/lazybrew`, `$T/deadcode -test ./...`,
  `$T/staticcheck ./...`. Compare with the receipt's "before" and "after".
  Anything still listed must be justified in the receipt; the one accepted
  leftover is `SA1019` (`parser.ParseDir`) in `keymap_test.go`.
- **V4 deletions are really dead:** for every deleted symbol, test and field,
  grep the pre-slice tree: no production caller, and no test that exercised
  something *other than* the deleted code. A deleted test must belong to code
  this slice deletes (otherwise: a finding). Coverage floors are not lowered
  (`scripts` / `make cover-check` in the diff).
- **V5 AC-01:** `go build ./cmd/...` no longer links `MockRunner`
  (`go build -o /tmp/x ./cmd/lazybrew && go tool nm /tmp/x | grep -c MockRunner`
  prints 0); no import cycle workaround left half-done; `ErrTest` replaces
  both `assertAnError`.
- **V6 AC-02/03/04:** `p.items` decision stated and applied; the
  `presentation.Format*` removals match the brief and `FormatFormulaInfo` /
  `FormatCaskInfo` stay; `CasksWrite.Pin`/`Unpin` removed and **no key binding
  offers pin on Casks** (try it in a test or by reading the key table); every
  removed domain field was checked for reads, and any UI line removed with it
  is named; the unreachable-branch removals each come with the proof the brief
  asks for (Deps tab shows its content without a fetch).
- **V7 AC-05/06/07:** removed styles are unused by production; the version
  sentinel: `./bin/lazybrew --version` prints sensibly with and without
  `-ldflags -X`, and the Makefile still targets the variable; the AC-07 test
  fixes add the assertion that was meant (check at least three SA4006 sites,
  and the SA4017/SA4000 pair in `hot_path_test.go` is no longer vacuous);
  the `git mv` renames are pure (`git diff -M --stat`, 100% similarity, no
  content edits in that commit) and the old to new mapping is in the receipt.
- **V8 frames and behaviour:** frames are unchanged except where AC-03.3 or
  AC-03.4 names a golden. Build your **own** frame harness and diff `View()`
  at the commit before AC-01 and at HEAD (80x24 and 140x40, both themes,
  every panel and tab, every modal, help, footer, the overlay; freeze any live
  log the Services Logs tab tails), and diff the brew args recorded for the
  write actions. Goldens/teatest expectations in the diff must be justified.
- **V9 invariants:** Bubble Tea rules held; no scope creep (no new features,
  no refactors beyond deletions); one commit per AC.

## Severity

blocker (gate fails, test deleted/weakened, regression, a Bubble Tea rule
broken) · major (AC not met, test passes without the fix, scope breach) ·
minor (docs/process, polish).

## Report

`work/inbox/reviewer-cr5-dead-code-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports cr5-dead-code rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-cr5-dead-code-rN.md`. Do not touch other panes.
