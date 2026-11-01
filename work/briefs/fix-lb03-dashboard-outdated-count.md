# Fix brief — slice `lb03-dashboard-outdated-count`, reviewer findings F-01, F-02, F-03

> **Contract.** The slice contract is `work/briefs/lb03-dashboard-outdated-count.md`.
> This brief names the findings it must fix. Where the two differ, this brief
> wins for the fix cycle only.

## Context

The runner lane made the status dashboard's outdated count honest (commits
`8a7b191`, `4ae32bd`, `99e54b7`, `45fc3f8`, `9fd23dd`, `adb8418`, `310256c`;
receipt `e8df8bc`/`3683a8d`). The reviewer lane returned
**`Verdict: changes-needed`** in
`work/inbox/reviewer-lb03-dashboard-outdated-count.md` (commit `3cf36b7`).

**The implementation is correct on every guarantee.** The reviewer verified
G-1..G-5 by command, the gate passes, and five mutations confirmed the new
tests bite. The findings are **two unprotected behaviours and one narrow edge
case**:

- **F-01 (major)** — the install / uninstall / task-completion emitters are
  unprotected. Mutating `internal/gui/gui.go:302`, `:345` or `:385` to
  `RefreshMsg{LoadOutdated: true}` leaves the whole suite green. Only the
  auto-refresh tick has a structural test. The exact regression this slice
  forbids would ship unnoticed.
- **F-02 (major)** — the cache-survival behaviour (AC-04: a status refresh
  keeps the outdated cache) has **no test**. The reviewer wrote a probe that
  passes on current code and fails when the change is reverted, then deleted
  it. The test is viable; it is simply missing.
- **F-03 (minor)** — with `brew.update_on_start: true` (non-default; the
  default is `false`, `internal/config/config.go:70`), the first `R` takes
  `internal/gui/gui.go:570-586`, which enqueues `brew update` and returns
  **no** `RefreshMsg`; the follow-up refresh comes from the task-completion
  path (a plain `RefreshMsg{}`), so the count stays "not loaded" until `R` is
  pressed again.

**F-04** (the receipt committed twice) is a process note — accepted, no
action, history is not rewritten.

**Q-04 is resolved** — the per-class reading is the contract. The ruling is
appended in `work/questions.md`.

## The findings to fix

One commit per finding, in order. Stage by name — never `git add -A`, never
`git add .`.

### AC-01 (F-01) — the non-`R` emitters are protected by a test

Add a test that proves the install / uninstall / task-completion paths cannot
silently start shelling out to `brew outdated`.

Mirror the shape of the existing `TestAutoRefreshTickStaysCacheOnly`: drive
the model with `TaskCompletedMsg` and with `MutationResultMsg` (the install
and uninstall completions), and assert that the emitted refresh message is
**plain** — `LoadOutdated == false`.

The assertion must be on the *emitted message*, not on the absence of a brew
call, so that the test bites exactly as the reviewer's mutation showed: when
the emitter is changed to `RefreshMsg{LoadOutdated: true}`, this test must
fail.

### AC-02 (F-02) — the cache-survival behaviour is covered

Add the test the reviewer probed and deleted:

1. prime **both** outdated cache keys (`brew.KeyOutdatedFormulae` and
   `brew.KeyOutdatedCasks`) with a known count;
2. send a plain `RefreshMsg{}` — the path used after an install and on the
   auto-refresh tick;
3. assert the cache **still holds** both keys afterwards;
4. assert the dashboard still renders the real count (`Outdated:  1 packages`
   for one cached package), not the not-loaded rendering.

This test must fail if the two outdated keys are put back into
`invalidatePanelCache`'s `PanelStatus` case.

### AC-03 (F-03) — `R` loads the count even when it also runs `brew update`

With `brew.update_on_start: true`, pressing `R` must still load the outdated
count.

The `R` key handler (`internal/gui/gui.go:569-586`) has two branches:

- **no update running** — it enqueues the `brew-update` task and returns
  `m.tasks.RunNext()`, emitting **no** `RefreshMsg`;
- **an update already running** — it returns a batch with
  `RefreshMsg{LoadOutdated: true}` already.

Fix the first branch so it also emits a load-flagged refresh. The simplest
shape is a batch of the running task plus
`func() tea.Msg { return RefreshMsg{LoadOutdated: true} }`.

**Two constraints on the fix:**

1. **Do not** change the `TaskCompletedMsg` handler for `msg.ID == "brew-update"`
   (`internal/gui/gui.go:288-304`) to set the flag. That handler serves **two**
   callers — the `R` key path *and* the Init path (`StartUpdateMsg`,
   `internal/gui/gui.go:262-268`) — so setting the flag there would load the
   count at **startup**, which is exactly what M9 forbids. The flag belongs
   on the `R` path only.
2. The `R`-path update runs `brew update` through
   `m.client.Runner.ExecuteStream(ctx, "update")`, which does **not** go
   through `diagnosticsWriter.Update` — so it does **not** call
   `InvalidateFor("update")` and does not wipe the outdated cache. The
   load-flagged refresh therefore populates the cache and the count sticks.
   Do not add an invalidation, and do not change `internal/brew/**`.

Add a test that drives the `R` key with `brew.update_on_start: true` and
asserts that a load-flagged refresh is emitted. Bound every wait.

## Invariants

- **Exactly three commits**, one per finding, in F-01 → F-02 → F-03 order.
- **The dashboard still never shells out to `brew outdated` on its own** —
  not on Init, not on the auto-refresh tick, not after an install, not after
  an uninstall, not after a task completes. Only an explicit `R` may.
- **No production behaviour changes beyond F-03's fix.** The formatter, the
  known/unknown logic, the `RefreshMsg` plumbing and
  `invalidatePanelCache` stay exactly as they are.
- **No change to `internal/brew/**`**, to the timeout values (30s / 15s /
  45s), to `boundedRead` / `boundedReadIn` / `timeoutError`, to the
  detection rule, to the tab-content path, or to the Outdated panel.
- The existing `TestStatusDashboardDoesNotShellCallOutdated` keeps its
  no-shell-call assertion **byte-identical**.
- The only files changed: `internal/gui/gui.go` and the test files under
  `internal/gui/`.
- **No `work/` file is touched by this cycle** — the slice's state row is
  the orchestrator's to close.
- Do not amend, rebase, or revert existing commits.
- `master-plan/` is a frozen local archive — do not read or modify it.
- The worktree must be clean when you report.

## Evidence to produce

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Plus, for the fix:

```bash
# the R path now emits a load-flagged refresh in both branches
grep -n -A28 'case "R":' internal/gui/gui.go | head -34

# the brew-update completion handler is untouched (it serves Init too)
grep -n -A4 'msg.ID == "brew-update"' internal/gui/gui.go

# the new tests exist and pass
go test ./internal/gui/ -run 'Outdated|StatusDashboard|TaskCompleted|MutationResult|RKey' -count=1 -v
```

Record the real output of each in your receipt.

## Report protocol

1. Write `work/inbox/runner-lb03-fix.md` — concise markdown, at most 30 lines:
   the three commit ids in order, each finding answered with the evidence
   that proves it, the gate result, and the exact shape of the F-03 fix.
2. Commit it:
   `git add work/inbox/runner-lb03-fix.md && git commit -m "inbox: runner reports lb03-fix"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `runner finished the lb03 fix cycle; receipt at work/inbox/runner-lb03-fix.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.

## Questions

If something in this brief is wrong or ambiguous, append it to
`work/questions.md` (append, do not rewrite) and keep going on the criteria
that are unaffected. The orchestrator owns the open questions; do not answer
them yourself.
