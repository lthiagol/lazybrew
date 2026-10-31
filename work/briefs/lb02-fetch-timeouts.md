# Slice: LB-02 — bound the remaining brew fetch paths with context timeouts

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

`LB-02` from `work/backlog.md` is: *apply context timeouts to the remaining
brew fetch paths (`fetchPanelData`, `fetchStatusData`, `fetchSelectedSearchInfo`,
`executeSearch`)* — residual from M13 findings F-02/F-07.

Today exactly **one** brew read path is bounded: the tab-content fetch, via
`fetchTabContentTimeout` (`internal/gui/commands.go:22`, used at `:833`). The
four paths below pass `context.Background()` straight to the brew client, so a
wedged or hung `brew` invocation leaves the panel spinning forever with no way
out. This slice bounds them.

## The four paths

| path | location | the unbounded call |
|---|---|---|
| `executeSearch` | `internal/gui/commands.go:294` (ctx at `:299`) | `m.client.Search.Search(context.Background(), query)` |
| `fetchSelectedSearchInfo` | `internal/gui/commands.go:322` (ctx at `:328`) | `m.client.Runner.Execute(ctx, "info", "--json=v2", r.Name)` |
| `fetchPanelData` | `internal/gui/commands.go:1166` (ctx at `:1168`) | `client.Formulae.List` / `Casks.List` / `Formulae.Outdated` / `Casks.Outdated` / `Taps.List` / `Services.List` |
| `fetchStatusData` | `internal/gui/commands.go:1243` (ctx at `:1245`) | five calls: `Formulae.List`, `Casks.List`, `Taps.List`, `Services.List`, `Diagnostics.Config` |

## The guarantees

This slice is specified by **behaviour**, not by code shape. Implement it your
way, as long as these four hold.

### G-1 — every individual brew read is capped

No single brew read inside the four paths may run longer than its cap:

- panel and status reads: **30 seconds**;
- search and search-info reads: **15 seconds**.

The caps must be **package-level `var`s**, so a test can override them the way
`TestFetchTabContentCmdTimesOut` overrides `fetchTabContentTimeout`
(`internal/gui/gui_test.go:1276-1279`). Name them so the intent is obvious.

### G-2 — the status dashboard resolves within a total budget

`fetchStatusData` aggregates **five independent sources** into one dashboard.
Bounding each call is not enough: if every call hangs, the dashboard must still
resolve. It must complete within **45 seconds** in total, even if all five
calls hang until their individual caps.

A per-call cap of 30s and a total budget of 45s means one slow source can
consume at most 30s of that budget, and the other four still get a fair share.
Keep the two bounds separate — do not let a per-call cap become the total.

### G-3 — a timeout is reported, and it is the error type the repo already has

On a deadline the path must surface an error to the user, never a silent empty
state or a spinner.

Use `brew.TimeoutError` — it already exists at `internal/brew/errors.go:42` and
is **currently used nowhere**. This slice is what gives it a purpose; do not
invent a new error type.

`TimeoutError.Command` is a **short semantic label** for the brew command —
`"list"`, `"outdated"`, `"search"`, `"info"`, `"taps"`, `"services"` — because
the GUI layer does not know the underlying args. Its `Error()` renders
`"brew <command> timed out after <duration>"`, which is what the user sees.

**Detection rule — this is the trap.** `exec.CommandContext` kills the process
when the deadline fires, so `cmd.Run()` returns an error like
`"signal: killed"`, **not** `context.DeadlineExceeded`. Detect the timeout from
the **context's own error** (`ctx.Err() == context.DeadlineExceeded`), never
from the type of the returned error. A timeout detected the wrong way silently
never fires.

The error must reach the user through the **existing** plumbing — no new UI,
no new toast:

| path | the message that carries it |
|---|---|
| `fetchPanelData` | `DataLoadedMsg.Err` |
| `fetchStatusData` | the aggregated `errs` list already collected there |
| `executeSearch` | `SearchDoneMsg.Err` |
| `fetchSelectedSearchInfo` | `SearchInfoLoadedMsg.Err` |

### G-4 — the tab-content path is untouched

`fetchTabContentTimeout` and `fetchTabContentCmd` are already bounded and
already tested. Do not change them, and do not change
`TestFetchTabContentCmdTimesOut`.

## Acceptance criteria

One commit per criterion, in order. Stage by name — never `git add -A`, never
`git add .`.

### AC-01 — the panel and status paths are bounded

`fetchPanelData` and `fetchStatusData` get their caps (G-1), the status path
gets its total budget (G-2), and a timeout in either surfaces as
`&brew.TimeoutError{...}` through the existing messages (G-3).

In `fetchPanelData`, the `PanelOutdated` branch makes **two** calls
(`Formulae.Outdated` then `Casks.Outdated`); each gets its own deadline. The
`default:` branch (`PanelStatus`/`PanelSearch`) is not a brew call — leave it.

In `fetchStatusData`, `Diagnostics.Config`'s error is currently **ignored**
(`cfg, _ := client.Diagnostics.Config(ctx)`). Keep that behaviour — the
dashboard has always tolerated a missing config — but still give the call a
deadline so it cannot hang. Do not start surfacing that error.

### AC-02 — the search paths are bounded

`executeSearch` and `fetchSelectedSearchInfo` get the 15s cap and surface a
timeout as `&brew.TimeoutError{...}` through `SearchDoneMsg.Err` and
`SearchInfoLoadedMsg.Err`.

### AC-03 — tests prove the timeout is an error, not a hang

Add tests covering **all four paths**. The pattern already exists: use
`brew.NewMockRunner()` with an `ExecuteFn` that blocks on `ctx.Done()`, and
override the cap var to something small (50ms is the value the existing test
uses). See `TestFetchTabContentCmdTimesOut`
(`internal/gui/gui_test.go:1276`) for the exact shape.

Each test must assert that:

1. the call **returns** (it does not hang) — a test that can hang is worse
   than no test, so bound the wait itself;
2. the error is a `*brew.TimeoutError` (use `errors.As`), not a raw context
   error;
3. the message the user sees names the command and the duration.

Also add one test that the **status path's total budget** holds: with tiny
caps and a mock runner that hangs, `fetchStatusData` returns within the total
budget rather than after the sum of the per-call caps.

### AC-04 — the slice closes its own bookkeeping

1. In `work/backlog.md`: set `LB-02`'s `status` cell to `closed`, and add a
   one-line note under the table naming this slice's evidence.
2. In `work/backlog.md`: add a **new** row `LB-07` for the two brew reads
   that are still unbounded and that this slice found but did not touch —
   the uninstall dependency pre-check (`internal/gui/commands.go:951`,
   `Formulae.Uses`) and `toggleLeaves` (`internal/gui/commands.go:1045`,
   `Formulae.Leaves`). Priority low, status `active`, source "found by
   lb02-fetch-timeouts". **Do not change `LB-01`, `LB-03`, `LB-04`, `LB-05`,
   `LB-06` or `LB-DROP-01`.**
3. In `work/STATE.md`: add a row for this slice —
   `| lb02-fetch-timeouts | — | implemented | awaiting reviewer |`. The `—`
   is correct for now; the orchestrator fills it in when the slice closes.

## Invariants

- **The tab-content path is untouched** (G-4).
- **No behaviour change on the success path.** A fetch that completes in time
  must produce exactly what it produces today.
- **No new dependencies**, no toolchain changes, no changes to `go.mod` /
  `go.sum`.
- **No new error type.** `brew.TimeoutError` is reused as-is.
- No changes outside `internal/gui/commands.go`, the test files you add to
  under `internal/gui/`, and the three `work/` files named in AC-04.
- `internal/brew/**` is **not** modified — the error type already exists; you
  only use it.
- One commit per acceptance criterion, in AC order, staged by name.
- Do not amend, rebase, or revert existing commits.
- `master-plan/` is a frozen local archive — do not read or modify it.
- The worktree must be clean when you report.

## Scope — do NOT touch

- `internal/brew/**` (unless a test there is strictly required — prefer
  `internal/gui/` tests)
- `internal/gui/gui.go`, `internal/gui/presentation/**`, `internal/gui/modal/**`,
  `internal/gui/task/**`
- `Makefile`, `.github/workflows/**`, `scripts/**`
- `work/briefs/**`, `work/inbox/**`, `work/verification/**`

## Evidence to produce

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Plus, for the new behaviour:

```bash
# the caps exist and are overridable
grep -n "Timeout = " internal/gui/commands.go
# the timeout error type is used, not reinvented
grep -rn "TimeoutError" internal/gui/
# no bare context.Background() is left in the four paths
grep -n "context.Background()" internal/gui/commands.go
# every hit must be either the tab-content path's own WithTimeout or a
# path this slice has already bounded
```

Record the real output of each in your receipt. If a check fails, report it —
do not paper over it.

## Report protocol

1. Write `work/inbox/runner-lb02-fetch-timeouts.md` — concise markdown, at
   most 40 lines: the four commit ids in order, each AC answered with the
   evidence that proves it, the gate result, the cap values you chose and why,
   anything the next session must know.
2. Commit it:
   `git add work/inbox/runner-lb02-fetch-timeouts.md && git commit -m "inbox: runner reports lb02-fetch-timeouts"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `runner finished lb02-fetch-timeouts; receipt at work/inbox/runner-lb02-fetch-timeouts.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.

## Questions

If something in this brief is wrong or ambiguous, append it to
`work/questions.md` (append, do not rewrite) and keep going on the criteria
that are unaffected. The orchestrator owns the open questions; do not answer
them yourself.
