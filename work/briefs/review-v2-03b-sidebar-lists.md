# Review brief — slice `v2-03b-sidebar-lists`

> The slice contract is `work/briefs/v2-03b-sidebar-lists.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-v2-03b-sidebar-lists.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands.

## Verify

- **V1 gate:** `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`.
- **V2 frames:** run the committed harness against real Homebrew, diff vs
  `after/` (must match), every frame exactly W×H cells. Remove
  `internal/gui/dump` after.
- **V3 no duplication (the point of the slice):** for every panel, the main
  area never draws the list the sidebar draws. Check the frames and
  `grep -n renderList internal/gui/*.go` (no main-area caller). Status
  sidebar is a compact summary, not the full dashboard.
- **V4 Formulae tabs:** `List` gone; named constants for tab indices; grep for
  remaining hard-coded Formulae tab numbers (`activeTab == `, `case 0:` … in
  Formulae branches, `needsTabFetch`, mouse); tab content fetches still hit the
  right tab (Deps / Used By / Files) — read `needsTabFetch`/`loadTabContent`
  and the tests; sorting and the leaves filter (`L`) still work on the sidebar.
- **V5 Outdated batch:** `Space`/`a`/`u`/`U` from the sidebar; `Versions` shows
  the selected item + batch summary; batch marks land on the right rows after
  scrolling.
- **V6 sidebar width & rows:** default 35 %, cap 60, floor 20, README updated;
  rows built from typed data with icons; versions right-aligned; no `...`
  remnants; full-row highlight only in the focused box.
- **V7 scrolling/state:** one source of truth per box's window; nothing written
  at render time (look for assignments to panel fields inside render funcs).
- **V8 mouse/keys/help:** click row, wheel per box, click tab, wheel in main;
  help/hints no longer mention a `List` tab.
- **V9 every frame change explained (`result.md`)**; restyle of v2-03 helpers,
  modal/toast restyle, new data fetches = **major**.
- **V10 invariants:** test count rule from the contract; each removed
  assertion has a named replacement covering surviving behaviour; no
  keybinding change; `internal/brew/**` untouched; one commit per AC.

Also `cat after/140x40/*.ansi` in your terminal and flag anything that reads
badly as a **minor** finding with the frame name.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression) · major (AC not met, scope breach, unexplained
change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-v2-03b-sidebar-lists-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports v2-03b-sidebar-lists rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-v2-03b-sidebar-lists-rN.md`. Do not touch other panes.
