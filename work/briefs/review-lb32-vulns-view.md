# Review brief — slice `lb32-vulns-view`

> Contract: `work/briefs/lb32-vulns-view.md` (with `cr-common.md`). You review; you do **not**
> edit code, tests or commits. Review everything on `wip` after the commit that added the slice
> brief (`git log --oneline -- work/briefs/lb32-vulns-view.md`) up to HEAD. A receipt is a claim,
> not evidence. `CGO_ENABLED=1` for the race gate (`GOFLAGS=-p=1` if /tmp is tight). **Do not run
> `brew vulns`** (network): the real output is saved at `/home/agent/.cache/vulns-real.json` and
> in `internal/brew/testdata/vulns_v1.json`. No brew write commands; never launch the TUI against
> the real brew; scratch under /home/agent/.cache, deleted when done; do not touch other panes;
> ignore untracked Screenshot*.png.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 data:** parse the saved real JSON yourself (python) and compare every number the view
  shows: 24 findings / 7 packages / 90 checked / 37 skipped / 14 patched (libssh2), severity
  counts, per-package counts, the sort order (worst severity, then count, then name); fixed
  versions: SHA vs released version rule (try synthetic `1.2.3`, a 40-hex, a 7-hex short SHA,
  an empty list); CRITICAL/LOW synthetic; empty findings; exit 1 without JSON, malformed JSON,
  exit 0 with an empty report. Revert AC-01: tests fail.
- **V3 view as drawn (own harness, mock runner):** 80x24, 100x30, 140x40, both themes (and
  one dracula/nord frame): nothing wider than the terminal, summaries **wrapped** not cut, the
  chips and next steps coloured by theme role (check the SGR; HIGH/CRITICAL error, MEDIUM
  warning, LOW info, UNKNOWN muted; success for `upgrade →`), blocks readable, footer muted;
  the no-findings and fallback views; running/not-run/error states unchanged from lb27.
- **V4 next steps:** package in the loaded Outdated list → `upgrade → <version>` with the
  right version; Outdated not loaded → no claim; released fix → `fix in <highest>` (compare
  versions properly, e.g. 1.10 vs 1.9); SHAs only → `fixed upstream, not released`; none → `no
  fix yet`; UNKNOWN-only packages muted. `y` copy: full SHAs, skipped names, patched IDs, no ANSI.
- **V5 regressions:** `v` still runs once; opening the tab runs nothing; `R` never runs it;
  every other panel's frames byte-identical to the base (own diff); deadcode/staticcheck nothing
  new; scroll/reachability of a long report works (or the question is filed).
- **V6:** one commit per AC; floors not lowered; no new goroutines in handlers; no `program.Send`.

## Report
`work/inbox/reviewer-lb32-vulns-view-rN.md` (≤ 55 lines): first line `verdict: ok` or `verdict:
changes-required`; AC table; findings `F-rN-01 (severity) — file:line — what — evidence — fix
must achieve`. Commit once (`inbox: reviewer reports lb32-vulns-view rN`), worktree clean, reply
with one line: `done: work/inbox/reviewer-lb32-vulns-view-rN.md`.
