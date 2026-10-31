# State — the resume record

One row per slice. This file is the resume record: a cold session reads it to
learn what is done, what is in flight, and where to pick up.

| slice | commits | state | verdict |
|---|---|---|---|
| detach-master-plan | 3a770a7, b95af56, 7428739, 3431b9d, 180d398, 17b1b8d, dce1468, 26f44cf, 88dbed9 | closed | ok |
| lb01-m13-deps-verify | 789d506, a7d5498, 196575b, b7dec88, 544e0ed | closed | ok |
| lb02-fetch-timeouts | 0027596, f7bc287, 00f73f2, 16c24f4, 8647165, 0cfb213 | closed | ok |

**detach-master-plan, closed.** The five slice commits implement the detach
(`AGENTS.md`, `README.md`, `DESIGN.md`, the `work/` lane-flow structure,
`SMOKE-CHECKLIST.md`, and the untracking of `master-plan/`). The four fix
commits answer the first review's findings F-01..F-04. The re-review returned
`ok`: the gate passes (`make build`, `make lint`, `go test ./... -race
-count=1`, `make cover-check`), the archive is byte-identical on disk, and no
live `mp` instruction survives. Receipts: `work/inbox/runner-detach-master-plan.md`,
`work/inbox/reviewer-detach-master-plan.md`,
`work/inbox/runner-fix-detach-master-plan.md`,
`work/inbox/reviewer-fix-detach-master-plan.md`.

**Accepted, not fixed:** reviewer finding F-05 (minor) — the fix receipt
explained a command's output wrongly; its recorded output was right and its
explanation was not. No acceptance criterion was affected, so the receipt was
left as the historical record and the wrong sentence was corrected only in
`work/briefs/review-fix-detach-master-plan.md`.

**lb01-m13-deps-verify, closed.** `LB-01` is closed and `LB-06` with it: the
M13 Deps-tab fix (`ac5bac6`) is verified in all five of its parts, and
`work/verification/lb01-m13-deps.md` carries the mutation evidence — for each
of the five invariants, the exact mutation, the command, and the verbatim
failing output — proving that **nine** tests across the five invariants fail
when the invariant they cover is broken. The first review found no blocker;
its two remaining findings (F-01, F-05) were fixed by `b7dec88` and
`544e0ed`, and the re-review returned `ok` after reproducing every mutation
byte-for-byte. Receipts: `work/inbox/runner-lb01-m13-deps-verify.md`,
`work/inbox/reviewer-lb01-m13-deps-verify.md`,
`work/inbox/runner-lb01-m13-deps-fix.md`,
`work/inbox/reviewer-lb01-m13-deps-fix.md`.

**Accepted, not fixed:** F-03 (minor) — the runner receipt was committed twice
(`5841cc9`, then `7cc8084` tightening it to the brief's line limit). The
second commit was content-harmless and history is not rewritten.

**lb02-fetch-timeouts, closed.** `LB-02` is closed. The four brew read paths
that used `context.Background()` — `fetchPanelData`, `fetchStatusData`,
`fetchSelectedSearchInfo`, `executeSearch` — are now bounded: 30s per
panel/status read, 15s per search read, and a separate 45s **total** budget for
the status dashboard so it resolves even when every source hangs. A timeout
surfaces as `*brew.TimeoutError` — the error type that already existed at
`internal/brew/errors.go:42` and had no user until now — detected from the
context's own error, because `exec.CommandContext` kills the brew process and
returns `"signal: killed"` rather than `context.DeadlineExceeded`. The
dashboard's six reads now run concurrently under the total budget; the reviewer
judged that sound (race-clean over five runs, deterministic ⚠ order, the shared
brew cache is mutex-guarded) and **required** by the total-budget guarantee.
The tab-content path was not touched. Review verdict `ok` after five mutations
proved the new tests bite, then one fix cycle for a negative duration the
reviewer found in the clamp (`brew list timed out after -1s` → now `0s`, and a
spent parent no longer starts the read). Receipts:
`work/inbox/runner-lb02-fetch-timeouts.md`,
`work/inbox/reviewer-lb02-fetch-timeouts.md`,
`work/inbox/runner-lb02-timeout-budget-fix.md`,
`work/inbox/reviewer-lb02-timeout-budget-fix.md`.

**Accepted, not changed:** reviewer finding F-02 (minor) — the dashboard's
fan-out puts both heavy `brew info --json=v2` loads and `brew doctor` in the
same instant, a higher peak than the sequential shape would have. The reviewer
judged it acceptable (Init and manual `R` only; auto-refresh is off by
default) and it is a shape note, not a defect. Reconsider only if the dashboard
feels slow on a cold cache.

**Next slice:** `LB-04` in `work/backlog.md` — chunk `brew tap-info --json`
batches when the tap count risks ARG_MAX (`internal/brew/taps.go`). `LB-03`
(status dashboard outdated count) is **blocked on a product decision**: the
dashboard reports 0 outdated until the user first visits the Outdated panel,
because M9 made that fetch lazy — either label it, or pay a `brew outdated`
call at startup. `LB-07` (the two brew reads still left unbounded: the
uninstall dependency pre-check and `toggleLeaves`) was found by this slice and
is parked. `LB-05` is a human gate for the stable tag.

**How a cold session resumes work.** Read `work/briefs/<slice>.md` for the
contract of the slice in the row above, then read the receipts in `work/inbox/`
for that slice — the receipt is the record of what was actually done and what the
evidence was. The `commits` column tells you which commits to inspect, in AC
order. If a lane reported a blocker, it is named in the receipt and the
corresponding question (if any) is in `work/questions.md`. If the state is
`implementing` but the receipt is missing, the lane died mid-slice: re-read the
brief, check `git log`, and redo only the criteria without a commit. Do not
re-derive the plan from anywhere else — the brief is the contract, and
`master-plan/` is a frozen, untracked, local-only archive that is not an input to
any decision.
