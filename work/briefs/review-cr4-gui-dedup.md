# Review brief — slice `cr4-gui-dedup`

> The slice contract is `work/briefs/cr4-gui-dedup.md` (with
> `work/briefs/cr-common.md`); it wins where they differ. You review; you do
> **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-cr4-gui-dedup.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands — do not start/stop/restart moshi-hook.

## Verify

- **V1 gate:** `make ci`; also `GOOS=darwin go build ./...`.
- **V2 per AC (AC-01 … AC-08):** each finding in the brief is really fixed in
  the code, not just in a test. For every regression test, revert the fix
  (`git stash`-free: check out the pre-fix file into a temp worktree) and
  confirm the test **fails** there; a test that passes without the fix is a
  finding.
- **V3 no behaviour change:** this slice is a refactor. Goldens, snapshots and
  teatest flows are **unchanged** in the diff (a changed expectation is a
  blocker). The only intended behaviour change is AC-06.3 (a cask uninstall
  confirm runs no `brew uses`). Build your **own frame harness** and diff
  `View()` at the commit before AC-01 and at HEAD, at 80x24 and 140x40, in
  both themes, across every panel/tab and every modal, the operation overlay,
  help and the footer. Also diff the brew args the `MockRunner` records for
  each write action (install, uninstall, reinstall, upgrade, pin, tap, untap,
  trust, services, cleanup, autoremove, update) before and after.
- **V4 AC-01..03:** every enqueue tail goes through `enqueue` (grep for a
  remaining `Enqueue(` outside it); queue-full and queued-behind toasts
  unchanged; `errTask`/`streamTask`/`linesTask` handle success, error, ctx
  cancel and a nil line channel; the `runVulns` buffering choice is justified
  in the receipt and safe; the writer interface in `doMutation` picks the same
  writer as before; closures capture only what they use; the `selectNamed`
  merge keeps each panel's selection behaviour.
- **V5 AC-04..07:** one clip/fit helper set with tests (ANSI, wide runes,
  width 0 and 1); help and footer generated from one keymap, with a test that
  fails when a handled key is missing from help or footer (check it really
  fails: remove a key from the table); `1`..`7` and `[`/`]` still behave the
  same; the label map and `pendingLabel` have no silent fallback; no bare tab
  indices left.
- **V6 AC-08:** moves only. `git diff -M --color-moved=dimmed-zebra` and
  `git diff --stat` show nothing but moves and imports; no file over ~800
  lines; the receipt maps old to new. `wc -l` of `commands.go` and `gui.go`
  before and after is in the receipt.
- **V7 dead code:** deadcode/staticcheck cannot be fetched here (the
  classifier denies it). Do not try. Instead grep that every function the
  slice adds is called from production code, and that no helper it replaces is
  left behind with no caller. Say that the tools were not run.
- **V8 invariants:** Bubble Tea rules held; no scope creep (cr5 deletes dead
  code, so `p.items`, `presentation.Format*` and `extractPackageName` stay);
  one commit per AC.

## Severity

blocker (gate fails, test deleted/weakened, regression, a Bubble Tea rule
broken) · major (AC not met, test passes without the fix, scope breach) ·
minor (docs/process, polish).

## Report

`work/inbox/reviewer-cr4-gui-dedup-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports cr4-gui-dedup rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-cr4-gui-dedup-rN.md`. Do not touch other panes.
