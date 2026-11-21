# Review brief — slice `lb19-keys-and-search`

> The slice contract is `work/briefs/lb19-keys-and-search.md` (with
> `work/briefs/cr-common.md`); it wins where they differ. You review; you do
> **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-lb19-keys-and-search.md`),
including fix commits. Orchestrator commits touching only `work/` and
`SMOKE-CHECKLIST.md` wording are not findings. A receipt is a claim, not
evidence. `CGO_ENABLED=1` for the race gate. Never run brew write commands
(no install/uninstall/update); read-only `brew search/info/list` are fine.
Never launch the TUI against the real brew. Scratch under /home/agent/.cache,
not /tmp. Do not touch other panes.

## Verify

- **V1 gate:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 regression tests (revert each fix in a temp worktree, the test must
  fail):** AC-01 tab from a focused search; AC-02 headerless rows get types
  and `installed`; AC-04 confirm/toast/re-mark. The receipt says AC-03 has no
  red run: show that the removed keys really do nothing by mutation (restore
  `[` or `pgdown` in the key table: `TestRetiredTabAndPageKeysDoNothing` must
  fail), and the AC-01 flow assertion is weak (receipt): is the Search-wrap
  case actually pinned by a `View()` assertion? If not, a finding.
- **V3 real search shape:** run `brew search --formula claude` and `brew search
  --cask claude-code` for real (read-only) and feed that exact output through
  the parser/service: formula rows say `formula`, cask rows say `cask`, none
  say `unknown`; `claude-code` (an installed cask on this machine) shows
  `installed`; a tap-qualified name (`user/tap/name`) matches its last part;
  a name that is both a formula and a cask is handled. Check the exit-1
  "no results" case for each call, one call failing while the other succeeds,
  and the search timeout.
- **V4 Search actions:** by a harness with the mock runner: `x` on an installed
  formula/cask row sends the right args (`uninstall <f>`, `uninstall --cask
  <c>`), `X` zap on a cask, `x` on a non-installed row toasts and runs
  nothing, `i` on an installed row toasts and runs nothing; the uninstall goes
  through `task.Manager`; after completion the mark clears with no new
  search, and stale results (a search that lands after a newer one) cannot
  re-mark the wrong rows. Install from Search still sends the right args.
- **V5 keys:** the single keymap, footers and help agree (`keymap_test.go`),
  `h`/`l` switch tabs on every panel with tabs, type text in the search input
  and any other text input, and collide with no binding; `[`, `]`, `pgup`,
  `pgdown` do nothing in lists, the help screen, and the Files/Logs tabs;
  the mouse wheel still scrolls; nothing else still mentions the old keys
  (`grep` README, help, footers, comments, SMOKE-CHECKLIST, DESIGN).
- **V6 frames:** build your own harness and diff `View()` at the commit before
  AC-01 and at HEAD (80x24 and 140x40, both themes) for every panel and tab:
  only footers, help and Search rows may differ; any other change is a finding.
  Goldens changed in the diff must be justified.
- **V7 invariants:** Bubble Tea rules held; one commit per AC; floors not
  lowered; no test deleted without its code (the receipt renames the
  PgDn tests and `BenchmarkFilesTabPageDown`: judge them).

## Severity

blocker (gate fails, regression, wrong brew args, a write outside
`task.Manager`) · major (AC not met, test passes without the fix, wrong row
marked installed) · minor.

## Report

`work/inbox/reviewer-lb19-keys-and-search-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must achieve`.
Commit once (`inbox: reviewer reports lb19-keys-and-search rN`), worktree
clean, reply with exactly one line: `done: work/inbox/reviewer-lb19-keys-and-search-rN.md`.
