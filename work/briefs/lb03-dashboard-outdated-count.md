# Slice: LB-03 — the status dashboard must not report "0 outdated" when it means "not loaded"

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

`LB-03` from `work/backlog.md`: *the status dashboard's outdated count stays 0
until the first refresh.*

Today the dashboard renders a number that is sometimes a **lie**:

```
Outdated:  0 packages
```

That `0` means *"I have not checked yet"* whenever the Outdated panel has not
been visited — not *"you have nothing outdated"*. The formatter
(`internal/gui/presentation/formatters.go:112-116`) takes an `int` and cannot
tell the two states apart, and M9 made the Outdated fetch lazy on purpose.

This slice makes the dashboard honest, and gives the user a deliberate way to
get the real number.

**The maintainer has approved this approach explicitly.** M9's actual invariant
— *the dashboard must never shell out to `brew outdated` on its own* — stays
intact and stays enforced by its test.

## The target rendering

Before the count is known:

```
Formulae:  124 installed
Casks:     31 installed
Outdated:  — not loaded (press R to load)
```

After the count is known (including a genuine zero):

```
Outdated:  0 packages
Outdated:  7 packages
```

## The guarantees

### G-1 — the formatter can express "not loaded"

`FormatStatusDashboard` must be able to render both states. Add to
`internal/gui/presentation/formatters.go`:

```go
// OutdatedInfo carries the outdated-package count and whether the count is
// known yet. Known=false means the count has never been loaded — it must not
// be rendered as a number.
type OutdatedInfo struct {
	Count int
	Known bool
}
```

and replace the `outdatedCount int` parameter with `outdated OutdatedInfo`.

Rendering:

- `Known` → `Outdated:  %d packages` (so a genuine zero still reads
  `Outdated:  0 packages`);
- not `Known` → `Outdated:  — not loaded (press R to load)`.

The em dash is U+2014. The line keeps the existing `Outdated:  ` prefix so the
dashboard stays aligned.

Update the two existing callers of the formatter —
`internal/gui/presentation/formatters_test.go:94` and
`internal/gui/presentation/snapshot_test.go:98` — to pass the new type, and
update the snapshot expectation for the not-loaded rendering.

### G-2 — the dashboard knows known from unknown

`statusDashboard` (`internal/gui/commands.go:1382`) decides whether the count
is known:

- **Cache-only mode** (the default, and what Init uses): read the outdated
  entries from the client cache exactly as today. The count is **known only if
  both** `brew.KeyOutdatedFormulae` and `brew.KeyOutdatedCasks` are cache
  hits. A partial hit is **not** known — the count would be incomplete.
  A cache hit holding an *empty* slice **is** known: the reader caches an
  empty result when `brew outdated` exits 1, which means "checked, nothing
  outdated".
- **Load mode** (only on an explicit user request, see G-3): also read the
  outdated entries through the brew reader, bounded like every other read
  (`boundedReadIn` with `brewReadTimeout` and the command label `"outdated"`),
  as part of the same concurrent fan-out. The count is then known.

In load mode, if the outdated read **fails or times out**, the error joins the
dashboard's existing `⚠` lines like any other source, and the count falls back
to **not known** — never to a misleading `0`.

### G-3 — `R` is what loads it, and only `R`

`RefreshMsg` (`internal/gui/messages.go:45`) is an empty struct today and is
emitted from **several** places: the `R` key (`internal/gui/gui.go:591`), the
auto-refresh tick (`internal/gui/gui.go:181`), and after install / uninstall /
task completion (`internal/gui/gui.go:302`, `:345`, `:385`).

Those must be distinguished, because only the user pressing `R` may cause a
`brew outdated` call.

1. Give `RefreshMsg` a field: `type RefreshMsg struct { LoadOutdated bool }`.
2. The `R` key handler returns `RefreshMsg{LoadOutdated: true}`.
3. **Every other emitter stays `RefreshMsg{}`** — the auto-refresh tick, the
   install/uninstall paths, and the task-completion path. Their dashboard
   refresh remains cache-only and must not shell out to `brew outdated`.
4. The `RefreshMsg` handler (`internal/gui/gui.go:460`) passes the flag through
   to the status fetch.
5. `fetchStatusData` and `statusDashboard` take the flag. The Init call site
   (`internal/gui/gui.go:143`) passes **false**.

`R` from **any** panel loads the count — the hint on the dashboard says
"press R to load", so it must be true wherever you press it. The cost is
bounded by the reader's own TTL: `brew outdated` results are cached for 30
minutes (`DefaultOutdatedTTL`, `internal/config/config.go:52`), so `R` usually
reuses the cached result and only shells out when the data is stale. That is
the app's existing definition of "fresh" and `R` must not bypass it for this
class.

### G-4 — a status refresh must stop throwing the outdated cache away

`invalidatePanelCache(PanelStatus)` (`internal/gui/gui.go:873-880`) currently
invalidates `brew.KeyOutdatedFormulae` and `brew.KeyOutdatedCasks` along with
the other keys, and it runs on **every** refresh.

The status path reads the outdated entries from the cache only, so wiping
them there just discards a valid 30-minute result and forces an unnecessary
re-fetch later. **Remove the two outdated keys from the `PanelStatus` case**
and update its comment to say the dashboard reads outdated from the cache and
does not own that class.

Keep every other invalidation exactly as it is:

- `PanelOutdated` still invalidates both outdated keys, so `R` on the Outdated
  panel still force-refreshes that class;
- the mutation-level invalidations (`InvalidateFor("install")`, `"uninstall"`,
  `"upgrade"`, …) still drop the outdated entries. That is correct: after an
  install the outdated set may genuinely have changed, so the dashboard falls
  back to "not loaded" until the count is loaded again.

### G-5 — M9's invariant survives

The test `TestStatusDashboardDoesNotShellCallOutdated`
(`internal/gui/gui_test.go:1839`) has two assertions:

1. `fetchStatusData` must not invoke `brew outdated` — **this assertion must
   survive completely unchanged**;
2. the dashboard contains `"0 packages"` on an empty cache — **this is the
   lie this slice removes**, so it is the one assertion that changes.

Do not weaken assertion 1 into something weaker, and do not delete the test.

## Acceptance criteria

One commit per criterion, in order. Stage by name — never `git add -A`, never
`git add .`.

### AC-01 — the formatter expresses "not loaded"

Implement G-1: the `OutdatedInfo` type, the new `FormatStatusDashboard`
parameter, the two renderings, and the updates to the two existing callers
(`formatters_test.go`, `snapshot_test.go`) including the snapshot expectation.

### AC-02 — the dashboard computes known from unknown

Implement G-2: `statusDashboard` and `fetchStatusData` take the load flag; the
cache-only path reports "known" only on two cache hits; the load path reads
the outdated entries through the bounded reader; a failed or timed-out
outdated read in load mode falls back to "not known" and adds its `⚠` line.

### AC-03 — `R` loads it, and nothing else does

Implement G-3: the `RefreshMsg.LoadOutdated` field, the `R` key handler, the
`RefreshMsg` handler, and the Init call site. Every other `RefreshMsg`
emitter stays `RefreshMsg{}`.

### AC-04 — a status refresh stops discarding the outdated cache

Implement G-4: remove the two outdated keys from `invalidatePanelCache`'s
`PanelStatus` case and update the comment. Change nothing else in that
function.

### AC-05 — the tests

1. **Update** `TestStatusDashboardDoesNotShellCallOutdated`: keep the
   no-shell-call assertion exactly as it is; change only the rendering
   expectation to the not-loaded rendering. Add a comment saying why.
2. **A cache-only status fetch with an empty cache** renders the not-loaded
   rendering and contains no `"0 packages"`.
3. **A cache holding an empty outdated slice** (the "checked, nothing
   outdated" case) renders `Outdated:  0 packages` — known zero, not unknown.
4. **A populated cache** renders the real count.
5. **An explicit load** (`RefreshMsg{LoadOutdated: true}`) results in exactly
   one `brew outdated` call and a dashboard showing the real count.
6. **A plain `RefreshMsg{}`** — the path used after an install — results in
   **zero** `brew outdated` calls. This is the assertion that protects M9.
7. The snapshot test covers the not-loaded rendering.

Every wait in a test must be bounded. Use the mock-runner pattern already in
`internal/gui/gui_test.go` (`brew.NewMockRunner()` with an `ExecuteFn` that
counts `outdated` calls, as the existing test does).

### AC-06 — the slice closes its own bookkeeping

1. In `work/backlog.md`: set `LB-03`'s `status` cell to `closed`, and add a
   one-line note under the table naming this slice. **Do not change any other
   row.**
2. In `work/STATE.md`: add a row for this slice —
   `| lb03-dashboard-outdated-count | — | implemented | awaiting reviewer |`.
   The `—` is correct for now; the orchestrator fills it in when the slice
   closes.

## Invariants

- **The dashboard never shells out to `brew outdated` on its own** — not on
  Init, not on the auto-refresh tick, not after an install, not after a task
  completes. Only an explicit `R` may.
- **No new timeout var**; the outdated read reuses `brewReadTimeout` and the
  existing `boundedReadIn` helper, and it joins the existing concurrent
  fan-out under the existing `brewStatusTotalTimeout`.
- **No change to the detection rule**, to `boundedRead`, `boundedReadIn`, or
  `timeoutError`.
- **No change to the timeout values** (30s / 15s / 45s).
- **No change to the tab-content path**, to `fetchPanelData`, or to any other
  panel's fetch.
- **No change to the Outdated panel's own behaviour** — it stays lazy, and `R`
  on it still force-refreshes the outdated class.
- **No change to the install/uninstall mutation invalidations.**
- No changes to `internal/brew/**`, `internal/gui/presentation/**` beyond the
  formatter and its tests, `internal/gui/modal/**`, `internal/gui/task/**`,
  `Makefile`, `.github/workflows/**`, `scripts/**`, `go.mod`, `go.sum`.
- The only files changed: `internal/gui/presentation/formatters.go`,
  `internal/gui/commands.go`, `internal/gui/gui.go`, `internal/gui/messages.go`,
  the test files under `internal/gui/presentation/` and `internal/gui/`, and
  the two `work/` files named in AC-06.
- One commit per acceptance criterion, in AC order, staged by name.
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

All four must pass. Plus:

```bash
# the not-loaded rendering exists and the number is still reachable
grep -rn "not loaded" internal/gui/presentation/formatters.go
grep -rn "OutdatedInfo" internal/gui/presentation/formatters.go internal/gui/commands.go

# only the R path can ask for the outdated read
grep -rn "RefreshMsg{" internal/gui/gui.go
grep -rn "LoadOutdated" internal/gui/

# the status refresh no longer wipes the outdated cache
grep -n -A10 "case PanelStatus:" internal/gui/gui.go | head -14
```

Record the real output of each in your receipt. If a check fails, report it —
do not paper over it.

## Report protocol

1. Write `work/inbox/runner-lb03-dashboard-outdated-count.md` — concise
   markdown, at most 40 lines: the six commit ids in order, each AC answered
   with the evidence that proves it, the gate result, the exact wording of the
   two renderings, and anything the next session must know.
2. Commit it:
   `git add work/inbox/runner-lb03-dashboard-outdated-count.md && git commit -m "inbox: runner reports lb03-dashboard-outdated-count"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `runner finished lb03-dashboard-outdated-count; receipt at work/inbox/runner-lb03-dashboard-outdated-count.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.

## Questions

If something in this brief is wrong or ambiguous, append it to
`work/questions.md` (append, do not rewrite) and keep going on the criteria
that are unaffected. The orchestrator owns the open questions; do not answer
them yourself.
