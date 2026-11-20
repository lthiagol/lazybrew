# Questions — open questions

Append to this file. Never rewrite it: the history of what was asked and when is
the point.

## Open questions

_None. No questions are open._

## Rules

- The orchestrator owns the open questions. A lane does not answer them itself.
- A lane that finds something ambiguous in its brief writes it here and keeps
  going on the criteria it does not affect.
- The reviewer reports each question as **answered**, **missed**, or **newly
  raised**.
- A question that is not written here is a lost question.

## Questions raised late — slice `detach-master-plan` (runner lane)

Both items below were found by the runner lane while executing
`detach-master-plan` and recorded only in its receipt. They are appended here
for the record. The "no questions are open" line above stays true: both are
resolved.

### Q-01 — the `.mp-write.lock` contradiction (raised late)

- **Question:** AC-04 said the smoke checklist was "tracked only inside
  `master-plan/`", but git also tracked `master-plan/.mp-write.lock` — so the
  checklist was not the only tracked file there. Which of the two readings
  wins?
- **Raised by:** runner lane (slice `detach-master-plan`).
- **Owner:** the orchestrator.
- **Resolution:** resolved — the checklist was copied out to
  `SMOKE-CHECKLIST.md` and the directory was untracked with no on-disk file
  modified; the extra tracked lock file was removed from the index with the
  rest of the directory.

### Q-02 — the LB-06 clause ambiguity (raised late)

- **Question:** AC-04 required LB-06 to be "closed-by-verification with
  LB-01", while also saying "every row's status starts as `active` except the
  dropped row". The two clauses conflict for LB-06 specifically. Which one
  wins?
- **Raised by:** runner lane (slice `detach-master-plan`).
- **Owner:** the orchestrator.
- **Resolution:** resolved — the more specific clause wins: LB-06's status is
  `closed-by-LB-01`.

## Questions raised late — slice `lb01-m13-deps-verify` (reviewer lane)

The item below was found by the reviewer lane while reviewing
`lb01-m13-deps-verify` and recorded only in its receipt. It is appended here
for the record. The "no questions are open" line above stays true: it is
resolved.

### Q-03 — the review brief's invariant→test mapping was wrong for I-2 and I-5 (raised late)

- **Observation:** the invariant→test table in
  `work/briefs/review-lb01-m13-deps-verify.md` mis-mapped two of the five
  invariants. For I-2 it named `TestDepsSelectionChangeUpdatesContent`, which
  never sends a `DataLoadedMsg`; for I-5 it named `TestTabKey`,
  `TestTabSwitching` and `TestDataLoadedOnNonFetchTabDoesNotReissueFetch`, none
  of which reach `internal/gui/gui.go:627-634`. The reviewer's own mutations
  confirmed it: under M2 and M5 those named tests still passed — they did not
  bite. The slice itself was not at fault.
- **Correct mapping:** the authoritative table is the one in
  `work/briefs/fix-lb01-m13-deps-verify.md`:

  | id | test that fails when the invariant is broken |
  |---|---|
  | I-1 | `TestFetchTabContentCmdTimesOut` |
  | I-2 | `TestDepsTabPopulatedAfterDataLoaded`, `TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply` |
  | I-3 | `TestDepsFastPathRendersCachedDependencies`, `TestDepsTabViewLeavesLoadingAfterFastPath`, `TestDepsTabLoads` |
  | I-4 | `TestParseFormulaBuildDepsFromBuildDependencies`, `TestFormulaeReaderList` |
  | I-5 | `TestDepsSelectionChangeUpdatesContent` |

- **Raised by:** reviewer lane (slice `lb01-m13-deps-verify`).
- **Owner:** the orchestrator.
- **Resolution:** resolved — the mapping is corrected in
  `work/briefs/fix-lb01-m13-deps-verify.md` and recorded, with the mutation
  evidence that proves each of the five tests bites, in
  `work/verification/lb01-m13-deps.md`. The stale rows in
  `work/briefs/review-lb01-m13-deps-verify.md` are the orchestrator's file and
  were left untouched by the fix lane.
- **Update (appended, after the fix cycle):** the last sentence above was
  written when the fix lane reported and was already out of date by the time
  it was read. The orchestrator corrected the stale rows in
  `work/briefs/review-lb01-m13-deps-verify.md` in commit `fa32989`
  ("docs(brief): correct the invariant-test mapping flagged by the reviewer"),
  and the re-review confirmed the correction (see
  `work/inbox/reviewer-lb01-m13-deps-fix.md`, check S3). Q-03 is closed; this
  line records that the correction happened, not that it is pending.

## Questions raised late — slice `lb03-dashboard-outdated-count` (runner lane)

### Q-04 — AC-05.5's "exactly one `brew outdated` call" is unreachable on a cold cache

- **Question:** AC-05.5 asks that an explicit load produce "exactly one
  `brew outdated` call". Brew has two outdated readers with one cache key
  each (`KeyOutdatedFormulae`, `KeyOutdatedCasks`), each of which shells
  `brew outdated` with its own `--formula` / `--cask` filter, so a complete
  count on a cold cache costs two calls. One combined call would mean
  changing `internal/brew/**`, which the brief forbids. Which reading is
  the contract: one call per outdated class, or one call in total?
- **Raised by:** runner lane (slice `lb03-dashboard-outdated-count`).
- **Owner:** the orchestrator.
- **Resolution:** the runner implemented the per-class reading and covered
  both in `internal/gui/gui_test.go`:
  `TestStatusDashboardExplicitLoadShellsOutdatedOnce` (the casks class is
  already cached → the load makes exactly one `brew outdated` call and shows
  the real count) and
  `TestStatusDashboardExplicitLoadShellsOneCallPerOutdatedClass` (cold cache →
  exactly two calls, one per class). The brief's wording is met literally in
  the first case; the second documents the two-reader reality.
- **Orchestrator ruling (appended):** the **per-class** reading is the contract.
  AC-05.5's "exactly one `brew outdated` call" was imprecise wording on my
  part; the requirement it meant was that *only an explicit `R` may cause a
  `brew outdated` call at all*, and that the resulting count must be
  complete. One call per outdated class satisfies both: two calls on a cold
  cache (formulae and casks are separate readers with separate cache keys and
  separate `--formula` / `--cask` filters), one when the other class is
  already cached. Combining them into a single call would mean changing
  `internal/brew/**`, which the slice deliberately does not touch — and it
  would diverge from the Outdated panel's own path, which already makes two
  calls. The brief's AC-05.5 has been corrected to state the per-class
  wording; the contract is unchanged in every other respect. Q-04 is closed.

- **Question:** The Charm v2 modules require Go 1.26 (`go 1.26.0` now in `go.mod`), but `.github/workflows/wip-ci.yml` and `.github/workflows/stable-ci.yml` pin `go-version: '1.24'`. CI will not build the v2 port until the workflows are bumped. The brief does not name the CI files, so the slice leaves them unchanged. Which Go version should CI pin, and should the bump land with v2-01-charm-v2 or as its own change?
- **Raised by:** runner lane (slice `v2-01-charm-v2`).
- **Owner:** the orchestrator.
- **Orchestrator ruling (appended):** CI must build what `go.mod` declares. Both
  workflows switch from the pinned `go-version: '1.24'` to
  `go-version-file: go.mod`, so the next toolchain bump cannot drift again. It
  lands **inside `v2-01-charm-v2`**, as a fix-round commit: a slice whose CI
  cannot build it is not done. Closed.

## Q-v2-sidebar — what does the sidebar show in the overhaul?

- **Question:** the sidebar and the main area show the same thing today (the
  Status dashboard twice; the Formulae list in both). Option A: a compact nav
  list of panels with counts, content in the main area. Option B:
  lazygit-style — the sidebar keeps the stacked panel lists, the main area
  shows only the selected item's details/tabs, never the list again.
- **Raised by:** orchestrator, 2026-11-19.
- **Owner's ruling (the user, 2026-11-19):** **Option B, lazygit-style.**
  Applies from slice `v2-03`; `v2-02` does not restructure the sidebar. Closed.

## Q-v2-look — icon set and accent colour for the overhaul

- **Owner's ruling (the user, 2026-11-19):** icons default to the **Nerd Font**
  set, with a plain-Unicode fallback selectable by config (`gui.icons:
  unicode`) or `LAZYBREW_ICONS=unicode`. The focus/selection accent stays
  **Catppuccin mauve** (`#cba6f7`). Applies from slice `v2-03`. Closed.

## Q-v2-run — autonomous run v2-02 → v2-06 (the user is away)

- **Owner's rulings (the user, 2026-11-19):**
  - Run `v2-02` through `v2-06` back to back; **one** checkpoint with the user
    after `v2-06` (before/after frames of every slice), then the release step.
  - Animations (`v2-06`) are **on by default**; `gui.animations: false` turns
    them off; tests always run with them off.
  - Push `wip` after each slice closes with reviewer `verdict: ok`. Never push
    to `stable`.
  - A slice with findings left after review round 5: close it as
    `open-findings`, record the findings in `work/STATE.md` and
    `work/backlog.md`, and continue only with slices that do not depend on
    the unfixed part; otherwise stop and wait for the user. Closed.

## Questions raised — slice `v2-02-layout-bugs` (runner lane)

Layout-only slice: none of these changes the layout contract. Each is recorded
here and not decided in code.

### Q-v2-hints — bottom-bar hints are wider than 80 columns

- **Question:** at 80 columns the panel hint line does not fit the bar. Status
  ends `… vulns m missing`; Formulae ends `… upgrade all p pin L leaves F
  fetch`. The bar is two rows, so the tail is cut. The v2-01 baseline wrapped
  the tail onto a third row past row 24, which a 24-row terminal never showed.
  Should the hints wrap onto more rows, be truncated, or be ordered so the most
  useful ones fit?
- **Raised by:** runner, v2-02-layout-bugs. Not decided: it belongs with the
  header and bottom-bar work in v2-03.
- **Orchestrator ruling:** v2-03. The footer becomes **one** context line: the
  panel's most useful keys first, then globals, cut at the frame width with a
  trailing `? more` that is always visible. Everything else lives in `?` help.
  Closed.

### Q-v2-help — help text is taller than an 80×24 terminal

- **Question:** the help text is 38 rows. At 80×24 the box shows 23 rows and
  cuts the rest (54 words of the lower sections). The baseline overflowed the
  terminal instead. Should help scroll, or be split into pages at small sizes?
- **Raised by:** runner, v2-02-layout-bugs.
- **Orchestrator ruling:** v2-05. Help scrolls (j/k, arrows, PgUp/PgDn, mouse
  wheel) with a scroll position hint; v2-02's cut-to-box stays until then. Closed.

### Q-v2-small — an 80×24 sidebar has no room for inactive panels' rows

- **Question:** at 24 rows the sidebar has 8 content rows for 7 boxes. The
  active box gets two rows and each inactive box gets its title only. So at
  80×24 the `not loaded`, `(empty)` and item counts are not visible for
  inactive panels. The baseline showed them only by overflowing the terminal
  (28 box rows in 24). This is part of the lazygit-style sidebar decision
  (Q-v2-sidebar, v2-03).
- **Raised by:** runner, v2-02-layout-bugs.
- **Orchestrator ruling:** v2-03. With titles drawn in the border, an
  inactive panel's count / `not loaded` state goes **in its border title**
  (e.g. `╭─ 3 Casks ── not loaded ╮`), so a collapsed box still shows it. Closed.

### Q-v2-cask-deps — the cask `depends_on` shape is not read (pre-existing)

- **Question:** real `brew info --json=v2 --installed --cask` returns
  `depends_on` as an object, for example `{"formula": ["python@3.14"]}` for
  gcloud-cli. `flexStrings` reads only arrays and objects that carry `name` or
  `full_name`, so `Cask.DependsOn` is empty for every real cask. AC-07 did not
  change this: cask parsing is identical before and after. Should the Casks
  Deps tab show the `formula` and `cask` lists?
- **Raised by:** runner, v2-02-layout-bugs.
- **Orchestrator ruling:** yes, in v2-05 (content): parse the object shape and
  show `Formulae:` / `Casks:` lists (and macOS/arch requirements when present).
  Closed.

## Q-v2-services — Services panel in the overhaul

- **Owner's ruling (the user, 2026-11-19):** besides an Info view, the user
  wants to **start / stop / restart services** from lazybrew as first-class
  actions. Orchestrator plan: a dedicated slice `v2-05s-services` after
  `v2-05`, before `v2-06` (animations):
  - Info view from `brew services info --json` (status, running, PID, loaded,
    user, file, command, working dir, exit code, log paths when present);
  - start / stop / restart (already wired as `s` / `S` / `r` through
    `task.Manager`) made obvious and safe: shown in the footer and the Info
    view, state-aware (start only when not running, stop/restart only when
    running — otherwise a clear toast), confirm before stop/restart, the list
    and Info view refresh after the action with the selection kept, and a
    clear error when brew needs `sudo` (system services).
  - A Logs tab is parked in the backlog for after the overhaul.
- **Safety:** lanes must **never** run `brew services start|stop|restart|run|cleanup`
  against this machine's real services (`moshi-hook` is live here). Tests use
  the mock runner only.

## Questions raised — slice `v2-03-look-foundation` (runner lane)

Each is a judgement call made while building the slice. None blocked a
criterion; each is decided in code as described, and the orchestrator can
redirect it.

### Q-v2-03-a — the Status sidebar's count is the dashboard's line count

- **Question:** the sidebar label shows the item count when a panel is loaded.
  The Status panel's items are dashboard lines, so its label reads `16`, which
  counts nothing the user recognises. Should Status show no count?
- **Runner's decision:** the count is shown, as AC-03.3 says ("the count when
  loaded"). Dropping it for Status is a one-line change if preferred.
- **Raised by:** runner lane (slice `v2-03-look-foundation`).

### Q-v2-03-b — header counts come from the panels, not the dashboard

- **Question:** the header shows `N formulae` / `N casks` only once the
  Formulae / Casks panels are loaded. The Status dashboard knows the counts,
  but only as text lines, so the header does not show them on first launch.
  Should the dashboard load carry the counts into the model?
- **Runner's decision:** the header shows a count only when the panel's own
  data is known (AC-04.2: unknown is omitted). Carrying counts from the
  dashboard needs a second return value from `statusDashboard`; it was not
  done here.
- **Raised by:** runner lane (slice `v2-03-look-foundation`).

### Q-v2-03-c — 80×24 sidebar labels are cut hard

- **Question:** at 80 columns the sidebar is 24 cells, so a box's border has
  about 5 cells for its label: `not loaded` shows as `no…`. The name keeps
  priority. Is that acceptable, or should narrow boxes use shorter states
  (for example a dot for `not loaded`)?
- **Runner's decision:** the name keeps priority and the state is cut. No
  shorter vocabulary was introduced, because the brief fixes the words.
- **Raised by:** runner lane (slice `v2-03-look-foundation`).

### Q-v2-03-d — a running operation is dropped at 80 columns

- **Question:** the header's middle shows the running operation, but at 80
  columns the badges take the room and the operation disappears. The Log box
  still names the task. Should the operation outrank the badges?
- **Runner's decision:** badges keep priority (AC-04.2 says badges drop when
  space is short, and does not say the operation survives). A refresh with no
  operation also shows in the middle; the brief did not list it, and the old
  bottom bar showed it, so it has a place to live.
- **Raised by:** runner lane (slice `v2-03-look-foundation`).

### Q-v2-03-e — the config's first validation

- **Question:** AC-02.2 says an unknown `gui.icons` is "a config error like
  other invalid fields". No config field was validated before: `Load` only
  returned a YAML parse error. `Load` now returns an error for an unknown
  `gui.icons` and for an unknown `LAZYBREW_ICONS`. Is that the intended
  precedent, or should validation of other fields wait?
- **Runner's decision:** the error is returned from `Load`, so `app.go` fails at
  startup with the message, as it does for a YAML error.
- **Raised by:** runner lane (slice `v2-03-look-foundation`).

### Orchestrator rulings on Q-v2-03-a..e (2026-11-19)

- **Q-v2-03-a:** Status shows **no** count in its border label (dashboard line
  count means nothing to the user). Fix in the v2-03 fix round (F-orch-01).
- **Q-v2-03-b:** yes — the header should know the counts on first launch.
  Done in `v2-03b-sidebar-lists`, which already reshapes the Status data into a
  compact summary (carry typed counts from the dashboard load into the model).
  Not a v2-03 finding.
- **Q-v2-03-c:** when the full label does not fit, use a short form instead of
  cutting a word: `not loaded` → `–` (muted), loading → `…`, error → `!`;
  counts are always shown in full. Fix in the v2-03 fix round (F-orch-02).
- **Q-v2-03-d:** a **running operation outranks the badges**: when space is
  short, badges drop first, then the operation is cut with `…`; the app name
  always stays. Fix in the v2-03 fix round (F-orch-03).
- **Q-v2-03-e:** accepted — validating at `Load` and failing at startup is the
  precedent. Other fields are not validated in this overhaul.

## Questions raised — slice `v2-03b-sidebar-lists` (runner lane)

Each is a judgement call or a defect outside the brief's ACs. None blocked a
criterion; the code is as described and the orchestrator can redirect it.

### Q-v2-03b-a — the batch markers show on every sidebar list (pre-existing)

- **Question:** the batch selection is one map of indices into the Outdated
  list. The sidebar draws it on every panel, so after a batch on Outdated the
  Formulae, Casks, Taps and Services rows at the same indices show `●`. The old
  sidebar did the same; this slice did not introduce it. Fix it (pass the batch
  only when the active panel is Outdated, one line in `renderSidebar`), or
  leave it?
- **Raised by:** runner lane (slice `v2-03b-sidebar-lists`). Not fixed here.

### Q-v2-03b-b — an unknown Status count: left out, or a muted `–`?

- **Runner's decision:** left out. AC-03.4 says unknown values are omitted, so
  the Status sidebar has no `Outdated` row until the count is loaded; the
  Dashboard tab still says `not loaded (press R to load)`. A `–` row is a
  one-line change if preferred.
- **Raised by:** runner lane (slice `v2-03b-sidebar-lists`).

### Q-v2-03b-c — Tap Info and Trust show the same details block

- **Runner's decision:** both as the brief's table says. Tap Info is the
  summary line plus the selected tap's details; Trust keeps its details block.
  The two blocks now overlap. Should Trust show only the trust state?
- **Raised by:** runner lane (slice `v2-03b-sidebar-lists`).

### Q-v2-03b-d — the model's viewport is removed

- **Finding:** the viewport's content was set inside `render`, on a copy of the
  model, so the model's viewport never held content and a wheel over the main
  panel never scrolled. AC-04.1 needs it to scroll.
- **Runner's decision:** the viewport is removed; a scroll offset per tab item
  (`contentKey`: panel, tab and item) lives in the model, and the wheel moves it.
  The keys the viewport used to receive without content (PgUp and similar) now
  reach nothing. Confirm this reading.
- **Raised by:** runner lane (slice `v2-03b-sidebar-lists`).

### Orchestrator rulings on Q-v2-03b-a..d (2026-11-19)

- **Q-v2-03b-a:** fix it — batch markers are drawn only on the Outdated list.
  v2-03b fix round, F-orch-01 (with a test: a batch on Outdated leaves the
  Formulae rows at the same indices unmarked).
- **Q-v2-03b-b:** accepted — unknown Status counts are left out.
- **Q-v2-03b-c:** `Trust` shows only the trust story for the selected tap:
  trusted / untrusted / official, what that means for its formulae and casks,
  and the `t` key hint. `Tap Info` keeps the details. Fix round, F-orch-02.
- **Q-v2-03b-d:** accepted — the viewport goes; the per-item scroll offset is
  the model state. Keys that reached the empty viewport must now do what a
  user expects: `PgUp`/`PgDn` scroll the main content by a page,
  `Home`/`End` (`g`/`G` only if not already bound) to top/bottom. Fix round,
  F-orch-03, with tests; help documents them.

### Q-v2-04-wrap — confirm body: wrap or cut? (from review F-r1-04)

- **Question:** AC-02.5 says every line of a confirm is cut to the inner
  width; the runner wraps the question and the consequence line instead.
- **Orchestrator ruling (2026-11-20):** **wrap the prose** (question,
  consequence, input prompt) — a cut question can hide what is being
  confirmed; **cut** the single-line parts (chip, key line, menu items) with
  `…`. The box must still fit the frame at 80×24 (height included). AC-02.5 is
  read this way from now on. Closed.

## Q-v2-services-toast — a needs-sudo toast is wider than a toast's cut

- **Question:** AC-04.1 wants the error toast to read
  `<name>: needs sudo — run "sudo brew services <action> <name>" in a terminal`,
  about 82 cells for moshi-hook. A toast message is cut at 60 cells
  (`toastMessageWidth`: half the frame, at most 60), so the command was cut
  mid-way at every terminal width.
- **Runner's decision:** a toast message now wraps to at most three rows at the
  same width cap (`ansi.Wrap` in `modal/toast.go`); only past three rows is the
  last row cut with `…`. This extends Q-v2-04-wrap's rule (prose wraps, one-line
  chrome is cut) to toast messages, because a cut message can hide what happened.
  A message that fits keeps its one-row draw, so the short-toast frames are
  unchanged (checked in v2-05s-services frames). The orchestrator can redirect:
  cut the toast again and let only the needs-sudo text wrap, or cut it as before.
- **Raised by:** runner lane (slice `v2-05s-services`).

## Q-v2-services-icons — the Info view's status glyphs and words

- **Question:** the slice brief writes the Info status as `● running`,
  `○ stopped`, `✗ error` and `⏲ scheduled`. The Unicode icon set has `▶` for
  running and `■` for stopped (the sidebar uses those, and the brief says not to
  restyle it). The Info words were `started` in the list and the brief says
  `running`.
- **Runner's decision:** the Info view draws its glyphs from the icon set
  (`gui.icons`), so `show_icons` and the Nerd set still apply: running `▶`,
  stopped `■`, error `✗`, and scheduled `⏲` (a new `Scheduled` icon, Unicode
  `⏲`, Nerd `clock-o`). The word is `running` in the Info view. The sidebar
  keeps `started`. Hard-coding the brief's `●` and `○` would have bypassed the
  icon setting. The orchestrator can redirect to the literal glyphs.
- **Raised by:** runner lane (slice `v2-05s-services`).

### Orchestrator rulings on the v2-05s runner decisions (2026-11-20)

- **Q-v2-services-toast:** accepted — a toast whose text exceeds its cut width
  wraps to at most **3** rows (then cut with `…`); short toasts unchanged. The
  stack still never covers the header or footer.
- **Q-v2-services-icons:** accepted — Info status glyphs come from the icon set
  with the word `running`/`stopped`/…; the sidebar keeps brew's status word.
- Cron shown as `hour 3, minute 0` and `c cleanup` always shown: accepted.

## Q-v2-next — the user's checkpoint feedback on LB-08 (2026-11-20)

The user reviewed the checkpoint and liked it. They asked for four more things.
Rulings, all answered with the orchestrator's recommendation:

- **Plan:** four slices, run back to back with the same lane flow (Haiku builder,
  Opus reviewer, at most 5 rounds): `v2-07-op-dismiss` (bug) → `v2-08-preload`
  → `v2-09-main-layout` → `v2-10-tap-actions`. One user checkpoint at the end,
  then LB-05 smoke and the `wip` → `stable` PR.
- **Finished operation (bug):** after `restart moshi-hook` the main area stayed
  on `Completed in 0s` until Esc, with no hint. Ruling: a **successful**
  operation closes itself after ~2 s (the toast already reports it); a failed
  or cancelled one stays with an `Esc to close` hint; switching panels closes
  any finished operation.
- **Preload:** load every panel at startup, in the background, without
  blocking keys: Status first, then Formulae → Outdated → Casks → Taps →
  Services, at most 2 reads at once; the sidebar shows a loading state, not
  `not loaded`. `gui.preload` (default `true`) turns it off. Measured on this
  machine: every read is < 0.9 s, ~2.8 s serial.
- **Main panel layout:** full treatment — one column of inner padding,
  full-width tables for lists (Deps, Used By, tap Formulae, …), two-column
  Info views when the main area is wide, the Log box shrinks when idle.
- **Tap actions:** same standard as services — state-aware actions in Tap Info
  and the footer, `t` toggles trust (untrust confirmed), untap confirm states
  the consequence, selection kept by name after an action.

## Q-v2-07-pid — AC-04.1 cannot be a literal `diff -r` against v2-06 (runner lane)

- **Contradiction:** AC-04.1 says `v2-07-op-dismiss/after/` must equal
  `v2-06-animations/after/` with `diff -r` empty. The harness reads the live
  `moshi-hook` PID through `brew services info`: `1332457` now, `800` when v2-06
  was captured. The service was restarted at 2026-11-20 07:01:42 (before this
  slice), so frames 19 and 20 (six files, 140×40 and 80×24) differ on that PID
  row and its padding. The other 54 frames are byte-identical.
- **What the runner did:** kept the frames as captured, did not edit them or the
  harness, and reported the difference in
  `work/verification/v2-07-op-dismiss/result.md`.
- **Decision needed:** accept "identical except the live PID row" as the AC-04.1
  evidence, or pin the services frames to a fixed PID in the harness so the
  diff is empty.

### Orchestrator ruling on Q-v2-07-pid (2026-11-20)

Accepted. The user restarted `moshi-hook` from lazybrew before this slice
(that restart is the bug v2-07 fixes), so its PID changed from 800 to 1332457.
A frame that differs from `v2-06-animations/after/` only in that live PID row
satisfies AC-04.1. Later slices compare against `v2-07-op-dismiss/after/`.

## Questions raised — slice `v2-08-preload` (runner lane)

Neither blocks a criterion. Both are decided in code as described; the
orchestrator can redirect either.

### Q-v2-08-a — a failed preloaded panel retries when the user switches to it

- **Question:** AC-02.2 says a failed read shows its error in its box only. The
  existing `lazyLoadPanel` policy retries a panel that has no items when it is
  switched to, so a failed preloaded panel shows its `!` and error until the user
  switches to it; the switch then retries the read, and the error returns if it
  fails again. Keep the retry, or make a failed preloaded panel stay failed until
  `R`?
- **Runner's decision:** kept the existing retry (no change to the lazy policy).
  Test: `TestPreloadFailedReadShowsItsErrorOnItsPanelOnly`.
- **Raised by:** runner lane (slice `v2-08-preload`).

### Q-v2-08-b — a slow `brew tap-info` shows a degraded Taps panel with no error

- **Observation:** on this machine `brew tap-info --json <3 taps>` takes ~87 s
  (rc 0), past the 30 s `brewReadTimeout`. `tapsReader.List` then returns the
  name-only list and does not cache it, so the Taps panel shows `0 trusted`, no
  remote, and no error. The lazy path does this too (checked with preload off).
  Preload makes it visible at startup for every user with a slow brew. The
  frames in `work/verification/v2-08-preload/` (06, 14) carry the degraded read.
- **Question:** should a timed-out tap-info surface as an error on the Taps panel
  (as the other reads do), or stay a silent fallback? Changing it touches
  `internal/brew/**`, which this slice does not.
- **Raised by:** runner lane (slice `v2-08-preload`).

### Orchestrator rulings on the v2-08 runner questions (2026-11-20)

- **Q-v2-08-a:** accepted — a failed preloaded panel keeps the existing retry
  on switch.
- **Q-v2-08-b:** not a silent fallback any more, but not in v2-08. The slow
  read came from a network outage this morning; brew is back to 0.9 s. It moves
  to `v2-10-tap-actions` (new AC-00 there), because the trust toggle must not
  act on a trust state that was never read: a name-only Taps list must say so
  in the panel and refuse `t`/`T` until a full read lands.

## Q-v2-09-a — the Info split threshold never fires at 140x40 (runner lane)

- **Asked:** AC-03.1 of `work/briefs/v2-09-main-layout.md` splits the Info views
  when the main area's inner width is ≥ 100 columns.
- **Found:** at 140x40 with the default `sidebar_width` 35 the sidebar is 49
  cells, so the main box's inner width is 89. No Info view splits at the 140x40
  size the brief names for the frames. The split shows at 200x60 (inner 128),
  captured as an extra screen set in `work/verification/v2-09-main-layout/after/200x60/`.
- **Done meanwhile:** the rule is implemented as written (content width ≥ 98,
  which is box inner ≥ 100). The 140x40 frames show today's single column.
- **Decision needed:** keep 100 on the box's inner width (the 140x40 frames never
  split), or lower the threshold, or change the sidebar default so 140 columns
  splits.
- **Raised by:** runner lane (slice `v2-09-main-layout`).

## Q-v2-test-speed — slow gate and the cancel bug (2026-11-20)

The user asked why tests take so long. Measured: race run ~39 s, of which
`TestExecuteStreamCancelReportsCanceled` is 30 s because cancel kills only
brew's own process and its children keep the pipe open — a real bug (Esc on a
running install waits for curl/git to finish). Ruling (user, recommended
option): a new slice `v2-11-cancel-and-test-speed` **after** v2-10, before the
final checkpoint — kill the process group on cancel, and drive the gui tests'
real waits with fake time.

### Q-v2-builder-speed — the user's ruling on builder speed (2026-11-20)

Measured: the v2-09 builder spent 12.6 of its 25 min running tests (36 runs,
17 full suites at ~40 s each) and ~11.5 min generating; slice size (~800
lines, half tests) is normal. Rulings (user, both recommended):

- **Order:** `v2-11-cancel-and-test-speed` runs **next**, before `v2-10`
  (supersedes the "after v2-10" order in Q-v2-test-speed).
- **Lighter per-commit gate**, for v2-11 onward and all fix rounds: each
  commit must build and pass `go test` (with `-race`) for the packages it
  touches; the **full gate** (`make build`, `make lint`,
  `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`) runs once,
  before the receipt. Where a brief says "every commit passes the tests", this
  ruling defines it. CI still runs the full suite on every push.
- Builder stays on high effort.

### Orchestrator ruling on Q-v2-09-a (2026-11-20)

Lower the threshold. The user's complaint ("tables squeezed on the left, we're
not using the space") is about everyday terminal sizes like 140x40, where the
main box's inner width is 89 — a threshold that never fires there misses the
point. New rule: Info views split when the main box's **inner width ≥ 84**
(140x40 splits; 120 columns and 80x24 stay single-column). The sidebar default
stays. Also (the r3 note): the Log box gets the same one-column padding as the
main box, for a consistent right column.
