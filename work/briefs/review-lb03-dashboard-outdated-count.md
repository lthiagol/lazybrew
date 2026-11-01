# Review brief — slice `lb03-dashboard-outdated-count`

> **Contract.** The slice contract is `work/briefs/lb03-dashboard-outdated-count.md`.
> This brief tells you *how to verify* it. The contract wins where the two
> differ.

## What the runner claims

The runner lane made the status dashboard's outdated count honest. Its
commits:

| AC | commit | subject |
|---|---|---|
| AC-01 | `8a7b191` | feat(gui): let the status dashboard express an unknown outdated count |
| AC-02 | `4ae32bd` | feat(gui): compute the dashboard outdated count as known or unknown |
| AC-03 | `99e54b7` | feat(gui): let the R key load the dashboard outdated count |
| AC-04 | `45fc3f8` | fix(gui): keep the outdated cache on a status refresh |
| AC-05 | `9fd23dd` | test(gui): cover the dashboard outdated count states and the R load path |
| AC-06 | `adb8418` | docs(work): close LB-03 and record the lb03 slice in the state record |
| — | `310256c` | docs(work): raise the AC-05.5 outdated-call count question |

Its receipt is `work/inbox/runner-lb03-dashboard-outdated-count.md`, committed
twice (`e8df8bc`, then `3683a8d`) while the runner brought it under the
brief's line limit. Note it as a process observation; history is not rewritten.

**Q-04 is resolved.** The runner raised a legitimate question about AC-05.5's
"exactly one `brew outdated` call": brew has **two** outdated readers
(formulae and casks), each with its own cache key and its own
`--formula` / `--cask` filter, so a cold load costs **two** calls. The
orchestrator ruled that the **per-class** reading is the contract — the
requirement meant was "only an explicit `R` may cause a `brew outdated` call",
not a literal call count — and corrected AC-05.5's wording in the brief. The
ruling is appended in `work/questions.md`. Verify against that ruling, not
against the old wording.

The runner claims all four gate commands passed. **A receipt is a claim, not
evidence.** Run the gate yourself.

## What this slice must guarantee

| id | guarantee |
|---|---|
| G-1 | the formatter can express "not loaded" (`presentation.OutdatedInfo{Count, Known}`), rendering `Outdated:  — not loaded (press R to load)` when unknown and `Outdated:  N packages` when known (including a genuine 0) |
| G-2 | the dashboard knows known from unknown: cache-only reports known **only** when both outdated cache keys hit; load mode reads them through the bounded reader; a failed or timed-out load falls back to **not known**, never to a misleading 0 |
| G-3 | **only** the `R` key may cause a `brew outdated` call — not Init, not the auto-refresh tick, not after an install, not after a task completes |
| G-4 | a status refresh no longer discards the outdated cache; the Outdated panel's own invalidation and the mutation invalidations are unchanged |
| G-5 | M9's invariant survives: `TestStatusDashboardDoesNotShellCallOutdated` keeps its no-shell-call assertion **byte-identical** |

## Verify by command

### V1 — G-1: the formatter

```bash
grep -n -B2 -A16 "type OutdatedInfo" internal/gui/presentation/formatters.go
grep -n -A8 "outdatedCount\|outdated OutdatedInfo" internal/gui/presentation/formatters.go | head -30
grep -rn "not loaded" internal/gui/presentation/formatters.go
```

The not-loaded rendering must contain the em dash (U+2014) and the hint
"press R to load". The known rendering must still produce `0 packages` for a
known zero. The two existing callers (`formatters_test.go`, `snapshot_test.go`)
must pass the new type, and the snapshot expectation must cover the not-loaded
rendering.

### V2 — G-2: known vs unknown

Read `statusDashboard` and `fetchStatusData` in `internal/gui/commands.go`:

1. cache-only mode: known **only** when **both** `KeyOutdatedFormulae` and
   `KeyOutdatedCasks` hit; a partial hit is **not** known; an empty cached
   slice **is** known (that is "checked, nothing outdated");
2. load mode: the outdated entries are read through the brew reader with
   `boundedReadIn(totalCtx, brewReadTimeout, "outdated", …)` — inside the
   existing concurrent fan-out, with **no new timeout var**;
3. a failed or timed-out outdated read in load mode adds a `⚠ outdated:` line
   and leaves the count **not known** — it must never fall back to `0`.

### V3 — G-3: only `R` may shell out

```bash
grep -rn "RefreshMsg{" internal/gui/gui.go
grep -rn "LoadOutdated" internal/gui/
```

The `R` key handler must set `LoadOutdated: true`. **Every other emitter must
stay plain `RefreshMsg{}`**: the auto-refresh tick, the install path, the
uninstall path, and the task-completion path. The Init call site must pass
`false`.

This is the most important check in the review: a single extra
`LoadOutdated: true` on any non-`R` path would silently bring back a
`brew outdated` call on every refresh, which is exactly what M9 forbids.

### V4 — G-4: the cache survives a status refresh

```bash
grep -n -A10 "case PanelStatus:" internal/gui/gui.go | head -14
grep -n -A3 "case PanelOutdated:" internal/gui/gui.go | head -6
```

The `PanelStatus` case must **not** invalidate `KeyOutdatedFormulae` /
`KeyOutdatedCasks`. The `PanelOutdated` case must still invalidate them, and
the mutation-level invalidations in `internal/brew/cache.go` must be
untouched (`InvalidateFor("install")` / `"uninstall"` / `"upgrade"` still drop
the outdated keys).

**Also judge:** does any test cover AC-04? If the cache-survival behaviour has
no test, say so — it is a behaviour change with no coverage, which is a
finding even if the change is right.

### V5 — AC-05: the tests

```bash
go test ./internal/gui/ -run 'StatusDashboard|Outdated' -count=1 -v
```

Judge the assertions, not their presence:

1. `TestStatusDashboardDoesNotShellCallOutdated` keeps its no-shell-call
   assertion **byte-identical**; only the rendering expectation changed, and
   it must now assert the not-loaded rendering and the **absence** of
   `"0 packages"`;
2. a known zero from an empty cached slice renders `0 packages`;
3. a partial cache is **not** known;
4. an explicit load produces the real count, with **one `brew outdated` call
   per outdated class** (two on a cold cache, one when the other class is
   cached) — per the Q-04 ruling;
5. a plain `RefreshMsg{}` produces **zero** `brew outdated` calls — the
   assertion that protects M9;
6. the auto-refresh tick does **not** set the flag;
7. every wait is bounded.

### V6 — scope

```bash
git diff --name-only 7af8166..adb8418
git diff --name-only 7af8166..adb8418 -- internal/brew/ internal/gui/modal/ internal/gui/task/ internal/gui/tab/ go.mod go.sum .github/ Makefile scripts/
# must print nothing

grep -n "brewReadTimeout = \|brewSearchTimeout = \|brewStatusTotalTimeout = " internal/gui/commands.go
# still 30s / 15s / 45s
```

The tab-content path, `fetchPanelData`, the Outdated panel's own behaviour and
the mutation invalidations must be untouched.

### V7 — the gate, run by you

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Report the real output.

### V8 — AC-06: the bookkeeping

```bash
git show adb8418 -- work/backlog.md work/STATE.md
```

`LB-03` must be `closed` with a note; no other backlog row changed; the state
row must match the contract.

## Mutations to run and revert

Prove the new tests bite. **Revert every mutation**; `git status` must be
clean when you report.

- **N1 (G-3)** — make the `R` key return plain `RefreshMsg{}`. Run the
  "R loads the count" test. It must **fail**. Revert.
- **N2 (G-3, the M9 guard)** — make the auto-refresh tick (or the install
  path) return `RefreshMsg{LoadOutdated: true}`. Run the "plain refresh makes
  zero `brew outdated` calls" test. It must **fail**. Revert.
- **N3 (G-2, the lie)** — make the cache-only path treat a cache miss as a
  known zero. Run the not-loaded rendering test. It must **fail**. Revert.
- **N4 (G-2, the load)** — make the load mode ignore the flag and read only
  from the cache. Run the "explicit load shows the real count" test. It must
  **fail**. Revert.
- **N5 (G-4)** — restore the two outdated keys to `invalidatePanelCache`'s
  `PanelStatus` case. Run the dashboard tests. Report whether any test
  notices — if none does, that is the coverage gap asked about in V4.

A mutation that does not make its test fail is a finding: the test is vacuous.

## Open questions

Read `work/questions.md`. Report whether Q-04 is answered by the recorded
ruling, and whether this slice raised anything else.

## Scope you do not own

- You do not fix anything — including the brief's wording, which the
  orchestrator has already corrected.
- You mutate only to prove a check bites, and you revert it.
- You do not rewrite the implementation, the tests, or the receipts.
- You do not touch `work/STATE.md` beyond confirming it.
- You do not touch the contents of `master-plan/`.
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

1. Write `work/inbox/reviewer-lb03-dashboard-outdated-count.md` — concise
   markdown, at most 45 lines: the verdict, the findings, the evidence
   (command → result) for V1–V8, the outcome of each mutation N1–N5, and the
   state of the questions.
2. Commit it:
   `git add work/inbox/reviewer-lb03-dashboard-outdated-count.md && git commit -m "inbox: reviewer reports lb03-dashboard-outdated-count"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `reviewer finished lb03-dashboard-outdated-count; verdict in work/inbox/reviewer-lb03-dashboard-outdated-count.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.
