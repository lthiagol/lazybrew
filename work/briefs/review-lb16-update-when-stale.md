# Review brief — slice `lb16-update-when-stale`

> The slice contract is `work/briefs/lb16-update-when-stale.md`, plus
> `work/briefs/fix-lb16-update-when-stale-r1.md` and the rulings at the bottom
> of `work/questions.md` (with `work/briefs/cr-common.md`); they win where they
> differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-lb16-update-when-stale.md`),
including fix commits. Rounds ≥ 2: confirm each earlier finding is fixed, then
look for new problems. Orchestrator commits touching only `work/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate.
Never run brew write commands (no `brew update`, no installs); read-only
`brew --repository`, `brew outdated`, `brew info` are fine. Do not touch
moshi-hook. Scratch under /home/agent/.cache, not /tmp.

## Verify

- **V1 gate:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 regression tests:** revert each fix (temp worktree, pre-fix file) and
  confirm its test fails: the completion `RefreshMsg{LoadOutdated:true,...}`,
  the dropped stale `launchRead`, the `R` gate under `auto`, the unreadable
  time → stale rule.
- **V3 real brew:** run the real `LastUpdate` against this machine
  (`brew --repository`, then `.git/FETCH_HEAD` mtime). It must return that mtime
  and not the zero time. Also check a repository whose `FETCH_HEAD` is missing
  (zero → stale) and that no `git` process is spawned.
- **V4 launch behaviour (the user's bug):** with a mock runner, stale
  metadata: launch shows the active panel immediately, one `brew update` is
  enqueued through `task.Manager`, and when it finishes the Outdated count and
  the Outdated panel list reload and include packages that the pre-update
  `brew outdated` omitted. Fresh metadata: no `brew update`, header shows
  `updated Nh ago`. Mode `false`: nothing automatic, as before. Mode `true`:
  as before plus the ruled reloads. Hunt races: an Init read landing after the
  completion, a launch while `isUpdating`, a window resize/tick during the
  check, `R` pressed during the launch update, quitting during the update.
  Run the new tests with `-race -count=20`.
- **V5 config:** `update_on_start` accepts `true`/`false`/`auto` (bool and
  string forms in YAML), default `auto`; `update_stale_hours` default 6,
  `< 1` is an error; an invalid string errors with a clear message; the README
  and DESIGN describe it; an old config with `update_on_start: false` keeps
  its meaning.
- **V6 frames:** goldens/teatest changes in the diff are justified; build your
  own `View()` diff at the commit before AC-01 and at HEAD for the default
  config (80x24 and 140x40): only header update-badge text may differ.
- **V7 invariants:** Bubble Tea rules held (no `program.Send`, no raw
  goroutines, writes through `task.Manager`); one commit per AC; floors not
  lowered; no test deleted or weakened (the receipt names `TestBrewUpdate-
  CompletionStaysPlain` → `…LoadsOutdated` and the `preloadConfig()` change:
  judge both).

## Severity

blocker (gate fails, regression, race, a write outside `task.Manager`) · major
(AC not met, test passes without the fix, the user's bug not fixed) · minor.

## Report

`work/inbox/reviewer-lb16-update-when-stale-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must achieve`.
Commit once (`inbox: reviewer reports lb16-update-when-stale rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-lb16-update-when-stale-rN.md`. Do not touch other panes.
