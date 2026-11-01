# Re-review brief — fix cycle for slice `lb03-dashboard-outdated-count`

> **Contract.** The slice contract is `work/briefs/lb03-dashboard-outdated-count.md`;
> the fix contract is `work/briefs/fix-lb03-dashboard-outdated-count.md`. This
> brief tells you how to verify the fix. The contracts win where they differ.

## Context

The review of `lb03-dashboard-outdated-count` returned **`Verdict:
changes-needed`** with two major findings and one minor, in
`work/inbox/reviewer-lb03-dashboard-outdated-count.md` (commit `3cf36b7`):

- **F-01** — the non-`R` refresh emitters (install / uninstall / task
  completion) were unprotected: mutating them to `LoadOutdated: true` left the
  suite green.
- **F-02** — the cache-survival behaviour had no test.
- **F-03** — with `brew.update_on_start: true`, the first `R` did not load the
  count, because that branch emitted no `RefreshMsg`.

**The implementation was already correct on every guarantee.** These were
coverage gaps and one edge case.

The fix cycle produced:

| Finding | commit | subject |
|---|---|---|
| F-01 | `80065c4` | test(gui): pin the non-R refresh emitters to plain refreshes |
| F-02 | `b92ff9d` | test(gui): cover the outdated cache survival on a status refresh |
| F-03 | `93a036b` | fix(gui): load the outdated count when R also runs brew update |

Its receipt is `work/inbox/runner-lb03-fix.md` (commit `bd6f3e0`).

Your job: **re-derive the fix from the tree and prove it by command.** The
fix receipt is a claim, not evidence — reproduce the mutations.

## Verify by command

### S1 — F-01: the non-`R` emitters are protected

```bash
cat internal/gui/outdated_count_test.go
go test ./internal/gui/ -run TestNonRRefreshEmittersStayCacheOnly -count=1 -v
```

The test must assert on the **emitted message** (`LoadOutdated == false` for
every refresh the install, uninstall and task-completion paths emit), not on
the absence of a brew call — so that the reviewer's mutation (the emitters
changed to `LoadOutdated: true`) makes it fail.

### S2 — F-02: cache survival is covered

```bash
go test ./internal/gui/ -run TestStatusRefreshKeepsOutdatedCache -count=1 -v
```

It must prime **both** outdated cache keys, send a plain `RefreshMsg{}`, and
assert the cache still holds both keys, that no `brew outdated` call happened,
and that the dashboard renders the real count.

### S3 — F-03: `R` loads the count even when it runs `brew update`

```bash
grep -n -A28 'case "R":' internal/gui/gui.go | head -34
grep -n -A20 'msg.ID == "brew-update"' internal/gui/gui.go | head -26
```

The `R` handler's "no update running" branch must emit a **load-flagged**
refresh alongside the running update task. And the `brew-update` completion
handler must stay **plain** — it also serves the Init path, so a load-flagged
refresh there would load the count at startup and break M9.

Also confirm **no cache invalidation was added**: the `R` path runs
`brew update` through `Runner.ExecuteStream`, which never calls
`InvalidateFor`, so the loaded count sticks.

### S4 — scope

```bash
git diff --name-only 3cf36b7..93a036b
# must be internal/gui/gui.go and the new test file only

git diff --name-only 3cf36b7..93a036b -- work/ internal/brew/ internal/gui/presentation/ internal/gui/modal/ internal/gui/task/ go.mod go.sum .github/ Makefile scripts/
# must print nothing

grep -n "brewReadTimeout = \|brewSearchTimeout = \|brewStatusTotalTimeout = " internal/gui/commands.go
# still 30s / 15s / 45s
```

The formatter, the known/unknown logic, the `RefreshMsg` plumbing and
`invalidatePanelCache` must be exactly as the slice left them — this cycle
adds tests and one edge-case fix only.

### S5 — the gate, run by you

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Report the real output.

## Mutations to run and revert

Prove the new tests bite. **Revert every mutation**; `git status` must be
clean when you report.

- **M1 (F-01)** — change the install / uninstall / task-completion emitters to
  `RefreshMsg{LoadOutdated: true}`. Run
  `go test ./internal/gui/ -run TestNonRRefreshEmittersStayCacheOnly -count=1`.
  It must **fail**. Revert.
- **M2 (F-02)** — put the two outdated keys back into
  `invalidatePanelCache`'s `PanelStatus` case. Run
  `go test ./internal/gui/ -run TestStatusRefreshKeepsOutdatedCache -count=1`.
  It must **fail**. Revert.
- **M3 (F-03)** — put the `R` handler's "no update running" branch back to
  `return m, m.tasks.RunNext()`. Run
  `go test ./internal/gui/ -run TestRKeyLoadsOutdatedWhenUpdateOnStart -count=1`.
  It must **fail**. Revert.
- **M4 (F-03, the Init guard)** — set the `brew-update` completion handler to
  `RefreshMsg{LoadOutdated: true}`. Run
  `go test ./internal/gui/ -run TestBrewUpdateCompletionStaysPlain -count=1`.
  It must **fail**. Revert.

A mutation that does not make its test fail is a finding: the test is vacuous.

## Open questions

Read `work/questions.md`. Report whether Q-04's ruling holds and whether this
cycle raised anything new.

## Scope you do not own

- You do not fix anything.
- You mutate only to prove a check bites, and you revert it.
- You do not touch `work/**`, `internal/brew/**`, or the contents of
  `master-plan/`.
- You do not prompt the runner lane and you do not spawn agents.
- You commit only your own receipt file.

## Verdict format

Your receipt must end with one line:

```
Verdict: ok
```

or

```
Verdict: changes-needed
```

followed by findings, each as `F-NN | severity | what is wrong | the command
that proves it | the fix`. Severity is `blocker`, `major`, or `minor`.

## Report protocol

1. Write `work/inbox/reviewer-lb03-fix.md` — concise markdown, at most 30
   lines: the verdict, any findings, the evidence (command → result) for
   S1–S5, the outcome of each mutation M1–M4, and the state of the questions.
2. Commit it:
   `git add work/inbox/reviewer-lb03-fix.md && git commit -m "inbox: reviewer reports lb03-fix"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `reviewer finished the lb03 fix cycle; verdict in work/inbox/reviewer-lb03-fix.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.
