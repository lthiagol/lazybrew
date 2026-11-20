# Review brief — slice `cr2-brew-cache`

> The slice contract is `work/briefs/cr2-brew-cache.md` (with
> `work/briefs/cr-common.md`); it wins where they differ. You review; you do
> **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-cr2-brew-cache.md`),
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
- **V3 cache (AC-01..03):** `Cache.Get` never loses a concurrent fresh `Set`;
  every reader goes through the generic cached helper with singleflight;
  `TypedCache` gone; every writer (streaming and plain, success and failure)
  invalidates its group **after** the command; the set of brew commands a
  write runs and their args are unchanged (MockRunner).
- **V4 refresh (AC-04..05):** `update` has one helper used by both entry
  points and goes through `DiagnosticsWrite.Update`; an automatic refresh no
  longer runs doctor/config, `R` does; concurrent loads really overlap.
- **V5 runner/JSON/dedup (AC-06..08):** benchmarks before/after are in the
  receipt and re-runnable (`go test -bench`); `\r`/`\r\n` splitting tested;
  `findBrewPath` sentinel gone; JSON parser/fuzz tests green; casks artifact
  order deterministic.
- **V6 invariants:** read timeouts unchanged; exported `internal/brew` names
  used by `internal/gui` changed only where an AC says; footer/help/goldens
  unchanged; no scope creep (cr3/cr4/cr5 items untouched); Bubble Tea rules
  held; one commit per AC.

## Severity

blocker (gate fails, test deleted/weakened, regression, a Bubble Tea rule
broken) · major (AC not met, test passes without the fix, scope breach) ·
minor (docs/process, polish).

## Report

`work/inbox/reviewer-cr2-brew-cache-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports cr2-brew-cache rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-cr2-brew-cache-rN.md`. Do not touch other panes.
