# Review brief — slice `cr3-render-perf`

> The slice contract is `work/briefs/cr3-render-perf.md` (with
> `work/briefs/cr-common.md`); it wins where they differ. You review; you do
> **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-cr3-render-perf.md`),
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
- **V3 frames (all ACs):** every frame is byte-identical to before, except
  C11's error text. Goldens/teatest expectations are **unchanged** in the
  diff (a changed golden is a blocker unless it is the C11 text). Frames are
  exactly the terminal size at 80x24 and 140x40. Render with
  `./bin/lazybrew` read-only (never `s/S/r/f/c`, no brew writes) and compare
  with the pre-slice build (`git worktree` at the commit before AC-01).
- **V4 AC-01..03:** `renderSearchInfo` never calls `json.Unmarshal`; the
  error text is brew's own; `Operation` step state equals `parseSteps(all
  lines)`; the line cap drops old lines without changing progress;
  `TaskOutputMsg.Lines` consumers all updated; `readChunk` drains `outCh`
  before `TaskCompletedMsg` (stress test under `-race`, cancel path too); no
  raw goroutines or `program.Send`.
- **V5 AC-04..07:** caches are cleared on every input they depend on
  (tab content, selection, resize, theme) — look for a stale-frame path;
  `Len`/`Last` edge cases (empty, partial, wrapped); style sharing does not
  alias a mutable style; `sortFormulae` order identical, including case-only
  differences; `toastName` replacement keeps motion behaviour.
- **V6 benchmarks:** the receipt shows before/after for each AC that asks;
  re-run at least AC-02, AC-04 and AC-06 yourself (`go test -bench`).
- **V7 invariants:** no scope creep (cr4/cr5 items untouched: dedup, dead
  code); animations and motion unchanged; one commit per AC.

## Severity

blocker (gate fails, test deleted/weakened, regression, a Bubble Tea rule
broken) · major (AC not met, test passes without the fix, scope breach) ·
minor (docs/process, polish).

## Report

`work/inbox/reviewer-cr3-render-perf-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports cr3-render-perf rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-cr3-render-perf-rN.md`. Do not touch other panes.
